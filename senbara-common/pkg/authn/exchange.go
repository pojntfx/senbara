package authn

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// Exchange exchanges the OIDC auth code and state for a user's refresh token and ID token.
// If authCode is an empty string, it will sign out the user.
func (a *Authner) Exchange(
	ctx context.Context,

	authCode,
	state string,

	stateNonce,
	pkceCodeVerifier,
	oidcNonce string,

	setRefreshToken,
	setIDToken func(string, time.Time) error,

	clearRefreshToken,
	clearIDToken func() error,

	clearStateNonce,
	clearPKCECodeVerifier,
	clearOIDCNonce func() error,
) (
	nextURL string,

	signedOut bool,

	err error,
) {
	log := a.log.With(
		"authCode", authCode != "",
		"state", state,
	)

	a.log.Debug("Starting auth code exchange")

	// Sign out
	if strings.TrimSpace(authCode) == "" {
		log.Debug("Signing out user")

		nextURL = state
		if strings.TrimSpace(nextURL) == "" {
			nextURL = "/"
		}

		if err := clearRefreshToken(); err != nil {
			log.Warn("Could not clear refresh token", "err", errors.Join(errCouldNotClearRefreshToken, err))

			return "", false, errCouldNotClearRefreshToken
		}

		if err := clearIDToken(); err != nil {
			log.Warn("Could not clear ID token", "err", errors.Join(errCouldNotClearIDToken, err))

			return "", false, errCouldNotClearIDToken
		}

		return nextURL, true, nil
	}

	jsonOIDCState, err := url.QueryUnescape(state)
	if err != nil {
		log.Warn("Could not parse OIDC state", "err", errors.Join(ErrCouldNotLogin, err))

		return "", false, ErrCouldNotLogin
	}

	var rawOIDCState oidcState
	if err := json.Unmarshal([]byte(jsonOIDCState), &rawOIDCState); err != nil {
		log.Debug("Failed to unmarshal OIDC state", "error", errors.Join(ErrCouldNotLogin, err))

		return "", false, ErrCouldNotLogin
	}

	if rawOIDCState.Nonce != stateNonce {
		log.Debug("State nonce not valid, user is unauthorized")

		return "", false, ErrCouldNotLogin
	}

	if err := clearStateNonce(); err != nil {
		log.Warn("Could not clear state nonce", "err", errors.Join(errCouldNotClearStateNonce, err))

		return "", false, errCouldNotClearStateNonce
	}

	nextURL = rawOIDCState.NextURL
	if strings.TrimSpace(nextURL) == "" {
		nextURL = "/"
	}

	ru, err := url.Parse(nextURL)
	if err != nil {
		log.Warn("Could not parse return URL", "err", errors.Join(ErrCouldNotLogin, err))

		return "", false, ErrCouldNotLogin
	}

	// If the return URL points to login or authorize endpoints, redirect to root instead
	if apiHandlerPath := ru.Query().Get("path"); ru.Path == "/login" || apiHandlerPath == "/login" || ru.Path == "/authorize" || apiHandlerPath == "/authorize" {
		nextURL = "/"
	}

	// Sign in
	log.Debug("Exchanging auth code for tokens")

	oauth2Token, err := a.config.Exchange(ctx, authCode, oauth2.VerifierOption(pkceCodeVerifier))
	if err != nil {
		log.Warn("Could not exchange auth code", "err", errors.Join(ErrCouldNotLogin, err))

		return "", false, ErrCouldNotLogin
	}

	if err := clearPKCECodeVerifier(); err != nil {
		log.Warn("Could not clear PKCE code verifier", "err", errors.Join(errCouldNotClearPKCECodeVerifier, err))

		return "", false, errCouldNotClearPKCECodeVerifier
	}

	log.Debug("Setting refresh token, expires in one year")

	if err := setRefreshToken(oauth2Token.RefreshToken, time.Now().Add(time.Hour*24*365)); err != nil {
		log.Warn("Could not set refresh token", "err", errors.Join(errCouldNotSetRefreshToken, err))

		return "", false, errCouldNotSetRefreshToken
	}

	idToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		log.Warn("Could not extract ID token", "err", errors.Join(ErrCouldNotLogin, err))

		return "", false, ErrCouldNotLogin
	}

	log.Debug("Verifying tokens")

	id, err := a.verifier.Verify(ctx, idToken)
	if err != nil {
		log.Warn("Could not parse verify token", "err", errors.Join(ErrCouldNotLogin, err))

		return "", false, ErrCouldNotLogin
	}

	if id.Nonce != oidcNonce {
		log.Debug("OIDC nonce not valid, user is unauthorized")

		return "", false, ErrCouldNotLogin
	}

	if err := clearOIDCNonce(); err != nil {
		log.Warn("Could not clear OIDC nonce", "err", errors.Join(errCouldNotClearOIDCNonce, err))

		return "", false, errCouldNotClearOIDCNonce
	}

	log.Debug("Setting ID token", "expiry", oauth2Token.Expiry)

	if err := setIDToken(idToken, oauth2Token.Expiry); err != nil {
		log.Warn("Could not set ID token", "err", errors.Join(errCouldNotSetIDToken, err))

		return "", false, errCouldNotSetIDToken
	}

	return nextURL, false, nil
}
