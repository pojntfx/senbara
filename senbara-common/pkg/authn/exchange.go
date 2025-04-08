package authn

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"
)

// Exchange exchanges the OIDC auth code and state for a user's refresh token and ID token.
// If authCode is an empty string, it will sign out the user.
func (a *Authner) Exchange(
	ctx context.Context,

	authCode,
	state string,

	setRefreshToken,
	setIDToken func(string, time.Time) error,

	clearRefreshToken,
	clearIDToken func() error,
) (
	signedOut bool,
	nextURL string,

	err error,
) {
	log := a.log.With(
		"authCode", authCode != "",
		"state", state,
	)

	a.log.Debug("Handling auth code exchange")

	nextURL, err = url.QueryUnescape(state)
	if err != nil || strings.TrimSpace(nextURL) == "" {
		nextURL = "/"
	}

	// Sign out
	if strings.TrimSpace(authCode) == "" {
		log.Debug("Signing out user")

		if err := clearRefreshToken(); err != nil {
			log.Warn("Could not clear refresh token", "err", errors.Join(errCouldNotClearRefreshToken, err))

			return false, "", errCouldNotClearRefreshToken
		}

		if err := clearIDToken(); err != nil {
			log.Warn("Could not clear ID token", "err", errors.Join(errCouldNotClearIDToken, err))

			return false, "", errCouldNotClearIDToken
		}

		return true, nextURL, nil
	}

	ru, err := url.Parse(nextURL)
	if err != nil {
		log.Warn("Could not parse return URL", "err", errors.Join(errCouldNotLogin, err))

		return false, "", errCouldNotLogin
	}

	// If the return URL points to login or authorize endpoints, redirect to root instead
	if apiHandlerPath := ru.Query().Get("path"); ru.Path == "/login" || apiHandlerPath == "/login" || ru.Path == "/authorize" || apiHandlerPath == "/authorize" {
		nextURL = "/"
	}

	// Sign in
	log.Debug("Exchanging auth code for tokens")

	oauth2Token, err := a.config.Exchange(ctx, authCode)
	if err != nil {
		log.Warn("Could not exchange auth code", "err", errors.Join(errCouldNotLogin, err))

		return false, "", errCouldNotLogin
	}

	log.Debug("Setting refresh token, expires in one year")

	if err := setRefreshToken(oauth2Token.RefreshToken, time.Now().Add(time.Hour*24*365)); err != nil {
		log.Warn("Could not set refresh token", "err", errors.Join(errCouldNotSetRefreshToken, err))

		return false, "", errCouldNotSetRefreshToken
	}

	idToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		log.Warn("Could not extract ID token", "err", errors.Join(errCouldNotLogin, err))

		return false, "", errCouldNotLogin
	}

	log.Debug("Setting ID token", "expiry", oauth2Token.Expiry)

	if err := setIDToken(idToken, oauth2Token.Expiry); err != nil {
		log.Warn("Could not set ID token", "err", errors.Join(errCouldNotSetIDToken, err))

		return false, "", errCouldNotSetIDToken
	}

	return false, nextURL, nil
}
