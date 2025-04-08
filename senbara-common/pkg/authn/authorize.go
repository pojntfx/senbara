package authn

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

type UserData struct {
	Email     string
	LogoutURL string
}

// Authorize authorizes a user based on a request and returns their data if it's available. If a user has been
// previously signed in, but their session has expired, authorize refresh their session. If `loginIfSignedOut` is set
// and a user has not signed in, authorize will redirect the user to the sign in URL instead - else, authorize will return
// only the data is has available on the user, without signing them in.
func (a *Authner) Authorize(
	ctx context.Context,

	loginIfSignedOut bool,

	returnURL string,
	currentURL string,

	privacyPolicyConsent bool,

	refreshToken,
	idToken *string,

	setRefreshToken,
	setIDToken func(string, time.Time) error,
) (
	nextURL string,
	requirePrivacyConsent bool,

	u UserData,
	status int,

	err error,
) {
	log := a.log.With("loginIfSignedOut", loginIfSignedOut)

	if strings.TrimSpace(returnURL) == "" {
		returnURL = "/"
	}

	log.Debug("Starting auth flow")

	log.Debug("Checking auth state", "privacyPolicyConsent", privacyPolicyConsent)

	if loginIfSignedOut || privacyPolicyConsent {
		if refreshToken == nil {
			if privacyPolicyConsent {
				log.Debug("Refresh token cookie is missing and privacy policy consent is given, reauthenticating with auth provider")

				return a.config.AuthCodeURL(url.QueryEscape(returnURL)), false, UserData{}, http.StatusTemporaryRedirect, nil
			}

			log.Debug("Refresh token cookie is missing, but can't reauthenticate with auth provider since privacy policy consent is not yet given. Redirecting to privacy policy consent page")

			return "", true, UserData{}, http.StatusTemporaryRedirect, nil
		}

		if idToken == nil {
			// Here, the user has still got a refresh token, so they've accepted the privacy policy already,
			// meaning we can re-authorize them immediately without redirecting them back to the consent page.
			// For updating privacy policies this is not an issue since we can simply invalidate the refresh
			// tokens in Auth0, which requires users to re-read and re-accept the privacy policy.
			// Here, we don't use the HTTP Referer header, but instead the current URL, since we don't redirect
			// with "redirect.html"
			returnURL := currentURL

			log.Debug("ID token cookie is missing and privacy policy consent is given since a valid refresh token exists, reauthenticating with auth provider")

			return a.config.AuthCodeURL(url.QueryEscape(returnURL)), false, UserData{}, http.StatusTemporaryRedirect, nil
		}
	} else {
		if refreshToken == nil {
			log.Debug("Refresh token cookie is missing, but logging in the user if the they are signed out is not requested, continuing without auth")

			return "", false, UserData{}, http.StatusOK, nil
		}

		if idToken == nil {
			// Here, the user has still got a refresh token, so they've accepted the privacy policy already,
			// meaning we can re-authorize them immediately without redirecting them back to the consent page.
			// For updating privacy policies this is not an issue since we can simply invalidate the refresh
			// tokens in Auth0, which requires users to re-read and re-accept the privacy policy.
			// Here, we don't use the HTTP Referer header, but instead the current URL, since we don't redirect
			// with "redirect.html"
			returnURL := currentURL

			log.Debug("ID token cookie is missing and privacy policy consent is given since a refresh token exists, reauthenticating with auth provider")

			return a.config.AuthCodeURL(url.QueryEscape(returnURL)), false, UserData{}, http.StatusTemporaryRedirect, nil
		}
	}

	log.Debug("Verifying tokens")

	id, err := a.verifier.Verify(ctx, *idToken)
	if err != nil {
		log.Debug("ID token verification failed, attempting refresh", "error", err)

		oauth2Token, err := a.config.TokenSource(ctx, &oauth2.Token{
			RefreshToken: *refreshToken,
		}).Token()
		if err != nil {
			log.Debug("Token refresh failed, reauthenticating with auth provider", "error", err)

			return a.config.AuthCodeURL(url.QueryEscape(returnURL)), false, UserData{}, http.StatusOK, nil
		}

		var ok bool
		*idToken, ok = oauth2Token.Extra("id_token").(string)
		if !ok {
			log.Debug("ID token missing from refreshed refresh token, reauthenticating with auth provider")

			return a.config.AuthCodeURL(url.QueryEscape(returnURL)), false, UserData{}, http.StatusOK, nil
		}

		id, err = a.verifier.Verify(ctx, *idToken)
		if err != nil {
			log.Debug("Refresh token verification failed, attempting refresh", "error", err)

			return a.config.AuthCodeURL(url.QueryEscape(returnURL)), false, UserData{}, http.StatusOK, nil
		}

		if *refreshToken = oauth2Token.RefreshToken; *refreshToken != "" {
			log.Debug("Setting new refresh token cookie, expires in one year")

			if err := setRefreshToken(*refreshToken, time.Now().Add(time.Hour*24*365)); err != nil {
				log.Warn("Could not set refresh token", "err", errors.Join(errCouldNotSetRefreshToken, err))

				return "", false, UserData{}, http.StatusInternalServerError, errCouldNotSetRefreshToken
			}
		}

		log.Debug("Setting new ID token cookie", "expiry", oauth2Token.Expiry)

		if err := setIDToken(*idToken, oauth2Token.Expiry); err != nil {
			log.Warn("Could not set ID token", "err", errors.Join(errCouldNotSetIDToken, err))

			return "", false, UserData{}, http.StatusInternalServerError, errCouldNotSetIDToken
		}
	}

	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err := id.Claims(&claims); err != nil {
		log.Debug("Failed to parse ID token claims", "error", err)

		return "", false, UserData{}, http.StatusUnauthorized, errors.Join(errCouldNotLogin, err)
	}

	if !claims.EmailVerified {
		log.Debug("Email from ID token claims not verified, user is unauthorized", "email", claims.Email)

		return "", false, UserData{}, http.StatusUnauthorized, errors.Join(errCouldNotLogin, errEmailNotVerified)
	}

	logoutURL, err := url.Parse(a.oidcIssuer)
	if err != nil {
		return "", false, UserData{}, http.StatusUnauthorized, errors.Join(errCouldNotLogin, err)
	}

	q := logoutURL.Query()
	q.Set("id_token_hint", *idToken)
	q.Set("post_logout_redirect_uri", a.oidcRedirectURL)
	logoutURL.RawQuery = q.Encode()

	logoutURL = logoutURL.JoinPath("oidc", "logout")

	log.Debug("Auth successful", "email", claims.Email)

	return "", false, UserData{
		Email:     claims.Email,
		LogoutURL: logoutURL.String(),
	}, http.StatusOK, nil
}
