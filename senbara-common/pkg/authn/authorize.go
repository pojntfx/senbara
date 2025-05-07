package authn

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type oidcState struct {
	Nonce   string `json:"nonce"`
	NextURL string `json:"nextURL"`
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

	privacyPolicyConsentGiven bool,

	refreshToken,
	idToken *string,

	setRefreshToken,
	setIDToken func(string, time.Time) error,

	setStateNonce,
	setPKCECodeVerifier,
	setOIDCNonce func(string) error,
) (
	nextURL string,

	requirePrivacyConsent bool,

	email,
	logoutURL string,

	err error,
) {
	log := a.log.With("loginIfSignedOut", loginIfSignedOut)

	if strings.TrimSpace(returnURL) == "" {
		returnURL = "/"
	}

	log.Debug("Starting auth flow")

	log.Debug("Checking auth state", "privacyPolicyConsentGiven", privacyPolicyConsentGiven)

	getAuthCodeURL := func(returnURL string) (string, error) {
		var (
			stateNonce       = oauth2.GenerateVerifier()
			pkceCodeVerifier = oauth2.GenerateVerifier()
			oidcNonce        = oauth2.GenerateVerifier()
		)

		if err := setStateNonce(stateNonce); err != nil {
			log.Warn("Could not set state nonce", "err", errors.Join(errCouldNotSetStateNonce, err))

			return "", errCouldNotSetStateNonce
		}

		if err := setPKCECodeVerifier(pkceCodeVerifier); err != nil {
			log.Warn("Could not set PKCE code verifier", "err", errors.Join(errCouldNotSetPKCECodeVerifier, err))

			return "", errCouldNotSetPKCECodeVerifier
		}

		if err := setOIDCNonce(oidcNonce); err != nil {
			log.Warn("Could not set OIDC nonce", "err", errors.Join(errCouldNotSetOIDCNonce, err))

			return "", errCouldNotSetOIDCNonce
		}

		rawOIDCState := oidcState{
			Nonce:   stateNonce,
			NextURL: returnURL,
		}

		jsonOIDCState, err := json.Marshal(rawOIDCState)
		if err != nil {
			log.Debug("Failed to marshal OIDC state", "error", errors.Join(ErrCouldNotLogin, err))

			return "", ErrCouldNotLogin
		}

		state := url.QueryEscape(string(jsonOIDCState))

		return a.config.AuthCodeURL(state, oidc.Nonce(oidcNonce), oauth2.S256ChallengeOption(pkceCodeVerifier)), nil
	}

	if loginIfSignedOut {
		if refreshToken == nil {
			if privacyPolicyConsentGiven {
				log.Debug("Refresh token cookie is missing and privacy policy consent is given, reauthenticating with auth provider")

				authCodeURL, err := getAuthCodeURL(returnURL)
				if err != nil {
					log.Warn("Could not get auth code URL", "err", errors.Join(errCouldNotGetAuthCodeURL, err))

					return "", false, "", "", errCouldNotGetAuthCodeURL
				}

				return authCodeURL, false, "", "", nil
			}

			log.Debug("Refresh token cookie is missing, but can't reauthenticate with auth provider since privacy policy consent is not yet given. Redirecting to privacy policy consent page")

			return "", true, "", "", nil
		}

		if idToken == nil {
			// Here, the user has still got a refresh token, so they've accepted the privacy policy already,
			// meaning we can re-authorize them immediately without redirecting them back to the consent page.
			// For updating privacy policies this is not an issue since we can simply invalidate the refresh
			// tokens in Auth0, which requires users to re-read and re-accept the privacy policy.
			// Here, we don't use the HTTP Referer header, but instead the current URL, since we don't redirect
			// with "redirect.html"

			log.Debug("ID token cookie is missing and privacy policy consent is given since a valid refresh token exists, reauthenticating with auth provider")

			authCodeURL, err := getAuthCodeURL(currentURL)
			if err != nil {
				log.Warn("Could not get auth code URL", "err", errors.Join(errCouldNotGetAuthCodeURL, err))

				return "", false, "", "", errCouldNotGetAuthCodeURL
			}

			return authCodeURL, false, "", "", nil
		}
	} else {
		if refreshToken == nil {
			log.Debug("Refresh token cookie is missing, but logging in the user if the they are signed out is not requested, continuing without auth")

			return "", false, "", "", nil
		}

		if idToken == nil {
			// Here, the user has still got a refresh token, so they've accepted the privacy policy already,
			// meaning we can re-authorize them immediately without redirecting them back to the consent page.
			// For updating privacy policies this is not an issue since we can simply invalidate the refresh
			// tokens in Auth0, which requires users to re-read and re-accept the privacy policy.
			// Here, we don't use the HTTP Referer header, but instead the current URL, since we don't redirect
			// with "redirect.html"

			log.Debug("ID token cookie is missing and privacy policy consent is given since a refresh token exists, reauthenticating with auth provider")

			authCodeURL, err := getAuthCodeURL(currentURL)
			if err != nil {
				log.Warn("Could not get auth code URL", "err", errors.Join(errCouldNotGetAuthCodeURL, err))

				return "", false, "", "", errCouldNotGetAuthCodeURL
			}

			return authCodeURL, false, "", "", nil
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

			authCodeURL, err := getAuthCodeURL(returnURL)
			if err != nil {
				log.Warn("Could not get auth code URL", "err", errors.Join(errCouldNotGetAuthCodeURL, err))

				return "", false, "", "", errCouldNotGetAuthCodeURL
			}

			return authCodeURL, false, "", "", nil
		}

		var ok bool
		*idToken, ok = oauth2Token.Extra("id_token").(string)
		if !ok {
			log.Debug("ID token missing from refreshed refresh token, reauthenticating with auth provider")

			authCodeURL, err := getAuthCodeURL(returnURL)
			if err != nil {
				log.Warn("Could not get auth code URL", "err", errors.Join(errCouldNotGetAuthCodeURL, err))

				return "", false, "", "", errCouldNotGetAuthCodeURL
			}

			return authCodeURL, false, "", "", nil
		}

		id, err = a.verifier.Verify(ctx, *idToken)
		if err != nil {
			log.Debug("Refresh token verification failed, attempting refresh", "error", err)

			authCodeURL, err := getAuthCodeURL(returnURL)
			if err != nil {
				log.Warn("Could not get auth code URL", "err", errors.Join(errCouldNotGetAuthCodeURL, err))

				return "", false, "", "", errCouldNotGetAuthCodeURL
			}

			return authCodeURL, false, "", "", nil
		}

		if *refreshToken = oauth2Token.RefreshToken; *refreshToken != "" {
			log.Debug("Setting new refresh token cookie, expires in one year")

			if err := setRefreshToken(*refreshToken, time.Now().Add(time.Hour*24*365)); err != nil {
				log.Warn("Could not set refresh token", "err", errors.Join(errCouldNotSetRefreshToken, err))

				return "", false, "", "", errCouldNotSetRefreshToken
			}
		}

		log.Debug("Setting new ID token cookie", "expiry", oauth2Token.Expiry)

		if err := setIDToken(*idToken, oauth2Token.Expiry); err != nil {
			log.Warn("Could not set ID token", "err", errors.Join(errCouldNotSetIDToken, err))

			return "", false, "", "", errCouldNotSetIDToken
		}
	}

	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err := id.Claims(&claims); err != nil {
		log.Debug("Failed to parse ID token claims", "error", errors.Join(ErrCouldNotLogin, err))

		return "", false, "", "", ErrCouldNotLogin
	}

	if !claims.EmailVerified {
		log.Debug("Email from ID token claims not verified, user is unauthorized", "email", claims.Email, "error", errors.Join(ErrCouldNotLogin, errEmailNotVerified))

		return "", false, "", "", errors.Join(ErrCouldNotLogin, errEmailNotVerified)
	}

	lu, err := url.Parse(a.oidcIssuer)
	if err != nil {
		log.Debug("Could not parse OIDC issuer URL", "error", errors.Join(ErrCouldNotLogin, err))

		return "", false, "", "", ErrCouldNotLogin
	}

	q := lu.Query()
	q.Set("id_token_hint", *idToken)
	q.Set("post_logout_redirect_uri", a.oidcRedirectURL)
	lu.RawQuery = q.Encode()

	lu = lu.JoinPath("oidc", "logout")

	log.Debug("Auth successful", "email", claims.Email)

	return "", false, claims.Email, lu.String(), nil
}
