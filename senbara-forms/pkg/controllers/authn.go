package controllers

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/leonelquinteros/gotext"
	"golang.org/x/oauth2"
)

type pageData struct {
	userData

	Page       string
	PrivacyURL string
	ImprintURL string

	BackURL string
}

type userData struct {
	Email     string
	LogoutURL string

	Locale *gotext.Locale
}

// authorize authorizes a user based on a request and returns their data if it's available. If a user has been
// previously signed in, but their session has expired, authorize refresh their session. If `loginIfSignedOut` is set
// and a user has not signed in, authorize will redirect the user to the sign in URL instead - else, authorize will return
// only the data is has available on the user, without signing them in.
func (c *Controller) authorize(w http.ResponseWriter, r *http.Request, loginIfSignedOut bool) (bool, userData, int, error) {
	returnURL := r.Header.Get("Referer")
	if strings.TrimSpace(returnURL) == "" {
		returnURL = "/"
	}

	c.log.Debug("Starting auth flow",
		"loginIfSignedOut", loginIfSignedOut,
		"method", r.Method,
		"path", r.URL.Path,
		"returnURL", returnURL,
	)

	locale, err := c.localize(r)
	if err != nil {
		return false, userData{}, http.StatusInternalServerError, errors.Join(errCouldNotLocalize, err)
	}

	privacyPolicyConsent := r.FormValue("consent") == "on"

	c.log.Debug("Checking auth state", "privacyPolicyConsent", privacyPolicyConsent)

	var refreshToken, idToken string
	if loginIfSignedOut || privacyPolicyConsent {
		rt, err := r.Cookie(refreshTokenKey)
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				if privacyPolicyConsent {
					c.log.Debug("Refresh token cookie is missing and privacy policy consent is given, reauthenticating with auth provider")

					http.Redirect(w, r, c.config.AuthCodeURL(url.QueryEscape(returnURL)), http.StatusFound)

					return true, userData{
						Locale: locale,
					}, http.StatusTemporaryRedirect, nil
				}

				if err := c.tpl.ExecuteTemplate(w, "redirect.html", redirectData{
					pageData: pageData{
						userData: userData{
							Locale: locale,
						},

						Page:       locale.Get("Privacy policy consent"),
						PrivacyURL: c.privacyURL,
						ImprintURL: c.imprintURL,
					},

					RequiresPrivacyPolicyConsent: true,
				}); err != nil {
					return false, userData{
						Locale: locale,
					}, http.StatusInternalServerError, errors.Join(errCouldNotRenderTemplate, err)
				}

				c.log.Debug("Refresh token cookie is missing, but can't reauthenticate with auth provider since privacy policy consent is not yet given. Redirecting to privacy policy consent page")

				return true, userData{
					Locale: locale,
				}, http.StatusTemporaryRedirect, nil
			}

			return false, userData{
				Locale: locale,
			}, http.StatusUnauthorized, errors.Join(errCouldNotLogin, err)
		}
		refreshToken = rt.Value

		it, err := r.Cookie(idTokenKey)
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				// Here, the user has still got a refresh token, so they've accepted the privacy policy already,
				// meaning we can re-authorize them immediately without redirecting them back to the consent page.
				// For updating privacy policies this is not an issue since we can simply invalidate the refresh
				// tokens in Auth0, which requires users to re-read and re-accept the privacy policy.
				// Here, we don't use the HTTP Referer header, but instead the current URL, since we don't redirect
				// with "redirect.html"
				returnURL := r.URL.String()

				c.log.Debug("ID token cookie is missing and privacy policy consent is given since a valid refresh token exists, reauthenticating with auth provider")

				http.Redirect(w, r, c.config.AuthCodeURL(url.QueryEscape(returnURL)), http.StatusFound)

				return true, userData{
					Locale: locale,
				}, http.StatusTemporaryRedirect, nil
			}

			return false, userData{
				Locale: locale,
			}, http.StatusUnauthorized, errors.Join(errCouldNotLogin, err)
		}
		idToken = it.Value
	} else {
		rt, err := r.Cookie(refreshTokenKey)
		if err != nil {
			c.log.Debug("Refresh token cookie is missing, but logging in the user if the they are signed out is not requested, continuing without auth")

			return false, userData{
				Locale: locale,
			}, http.StatusOK, nil
		}
		refreshToken = rt.Value

		it, err := r.Cookie(idTokenKey)
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				// Here, the user has still got a refresh token, so they've accepted the privacy policy already,
				// meaning we can re-authorize them immediately without redirecting them back to the consent page.
				// For updating privacy policies this is not an issue since we can simply invalidate the refresh
				// tokens in Auth0, which requires users to re-read and re-accept the privacy policy.
				// Here, we don't use the HTTP Referer header, but instead the current URL, since we don't redirect
				// with "redirect.html"
				returnURL := r.URL.String()

				c.log.Debug("ID token cookie is missing and privacy policy consent is given since a refresh token exists, reauthenticating with auth provider")

				http.Redirect(w, r, c.config.AuthCodeURL(url.QueryEscape(returnURL)), http.StatusFound)

				return true, userData{
					Locale: locale,
				}, http.StatusTemporaryRedirect, nil
			}

			c.log.Debug("ID token cookie is missing, but logging in the user if the they are signed out is not requested, continuing without auth")

			return false, userData{
				Locale: locale,
			}, http.StatusOK, nil
		}
		idToken = it.Value
	}

	c.log.Debug("Verifying tokens")

	id, err := c.verifier.Verify(r.Context(), idToken)
	if err != nil {
		c.log.Debug("ID token verification failed, attempting refresh", "error", err)

		oauth2Token, err := c.config.TokenSource(r.Context(), &oauth2.Token{
			RefreshToken: refreshToken,
		}).Token()
		if err != nil {
			c.log.Debug("Token refresh failed, reauthenticating with auth provider", "error", err)

			http.Redirect(w, r, c.config.AuthCodeURL(url.QueryEscape(returnURL)), http.StatusFound)

			return true, userData{
				Locale: locale,
			}, http.StatusOK, nil
		}

		var ok bool
		idToken, ok = oauth2Token.Extra("id_token").(string)
		if !ok {
			c.log.Debug("ID token missing from refreshed refresh token, reauthenticating with auth provider")

			http.Redirect(w, r, c.config.AuthCodeURL(url.QueryEscape(returnURL)), http.StatusFound)

			return true, userData{
				Locale: locale,
			}, http.StatusOK, nil
		}

		id, err = c.verifier.Verify(r.Context(), idToken)
		if err != nil {
			c.log.Debug("Refresh token verification failed, attempting refresh", "error", err)

			http.Redirect(w, r, c.config.AuthCodeURL(url.QueryEscape(returnURL)), http.StatusFound)

			return true, userData{
				Locale: locale,
			}, http.StatusOK, nil
		}

		if refreshToken = oauth2Token.RefreshToken; refreshToken != "" {
			c.log.Debug("Setting new refresh token cookie, expires in one year")

			http.SetCookie(w, &http.Cookie{
				Name:     refreshTokenKey,
				Value:    refreshToken,
				Expires:  time.Now().Add(time.Hour * 24 * 365),
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
				Path:     "/",
			})
		}

		c.log.Debug("Setting new ID token cookie", "expiry", oauth2Token.Expiry)

		http.SetCookie(w, &http.Cookie{
			Name:     idTokenKey,
			Value:    idToken,
			Expires:  oauth2Token.Expiry,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
		})
	}

	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err := id.Claims(&claims); err != nil {
		c.log.Debug("Failed to parse ID token claims", "error", err)

		return false, userData{
			Locale: locale,
		}, http.StatusUnauthorized, errors.Join(errCouldNotLogin, err)
	}

	if !claims.EmailVerified {
		c.log.Debug("Email from ID token claims not verified, user is unauthorized", "email", claims.Email)

		return false, userData{
			Locale: locale,
		}, http.StatusUnauthorized, errors.Join(errCouldNotLogin, errEmailNotVerified)
	}

	logoutURL, err := url.Parse(c.oidcIssuer)
	if err != nil {
		return false, userData{
			Locale: locale,
		}, http.StatusUnauthorized, errors.Join(errCouldNotLogin, err)
	}

	q := logoutURL.Query()
	q.Set("id_token_hint", idToken)
	q.Set("post_logout_redirect_uri", c.oidcRedirectURL)
	logoutURL.RawQuery = q.Encode()

	logoutURL = logoutURL.JoinPath("oidc", "logout")

	c.log.Debug("Auth successful", "email", claims.Email)

	return false, userData{
		Email:     claims.Email,
		LogoutURL: logoutURL.String(),

		Locale: locale,
	}, http.StatusOK, nil
}

type redirectData struct {
	pageData

	Href                         string
	RequiresPrivacyPolicyConsent bool
}

func (c *Controller) HandleLogin(w http.ResponseWriter, r *http.Request) {
	returnURL := r.Header.Get("Referer")

	log := c.log.With("returnURL", returnURL)

	log.Debug("Logging in user")

	redirected, _, status, err := c.authorize(w, r, true)
	if err != nil {
		log.Warn("Could not authorize user for login", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	http.Redirect(w, r, returnURL, http.StatusFound)
}

// exchange exchanges the OIDC auth code and state for a user's refresh token and ID token.
// If authCode is an empty string, it will sign out the user.
func (c *Controller) exchange(
	ctx context.Context,

	authCode,
	state string,

	setRefreshToken,
	setIDToken func(string, time.Time) error,

	clearRefreshToken,
	clearIDToken func() error,
) (
	signedOut bool,
	returnURL string,

	err error,
) {
	log := c.log.With(
		"authCode", authCode != "",
		"state", state,
	)

	c.log.Debug("Handling auth code exchange")

	returnURL, err = url.QueryUnescape(state)
	if err != nil || strings.TrimSpace(returnURL) == "" {
		returnURL = "/"
	}

	log = log.With("returnURL", returnURL)

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

		return true, returnURL, nil
	}

	ru, err := url.Parse(returnURL)
	if err != nil {
		log.Warn("Could not parse return URL", "err", errors.Join(errCouldNotLogin, err))

		return false, "", errCouldNotLogin
	}

	// If the return URL points to login or authorize endpoints, redirect to root instead
	if apiHandlerPath := ru.Query().Get("path"); ru.Path == "/login" || apiHandlerPath == "/login" || ru.Path == "/authorize" || apiHandlerPath == "/authorize" {
		returnURL = "/"
	}

	// Sign in
	log.Debug("Exchanging auth code for tokens")

	oauth2Token, err := c.config.Exchange(ctx, authCode)
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

	return false, returnURL, nil
}

func (c *Controller) HandleAuthorize(w http.ResponseWriter, r *http.Request) {
	authCode := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	log := c.log.With(
		"authCode", authCode != "",
		"state", state,
	)

	c.log.Debug("Handling user auth")

	locale, err := c.localize(r)
	if err != nil {
		log.Warn("Could not localize auth page", "err", errors.Join(errCouldNotLocalize, err))

		http.Error(w, errCouldNotLocalize.Error(), http.StatusInternalServerError)

		return
	}

	signedOut, returnURL, err := c.exchange(
		r.Context(),

		authCode,
		state,

		func(s string, t time.Time) error {
			http.SetCookie(w, &http.Cookie{
				Name:     refreshTokenKey,
				Value:    s,
				Expires:  time.Now().Add(time.Hour * 24 * 365),
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
				Path:     "/",
			})

			return nil
		},
		func(s string, t time.Time) error {
			http.SetCookie(w, &http.Cookie{
				Name:     idTokenKey,
				Value:    s,
				Expires:  t,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
				Path:     "/",
			})

			return nil
		},

		func() error {
			http.SetCookie(w, &http.Cookie{
				Name:     refreshTokenKey,
				Value:    "",
				MaxAge:   -1,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
				Path:     "/",
			})

			return nil
		},
		func() error {
			http.SetCookie(w, &http.Cookie{
				Name:     idTokenKey,
				Value:    "",
				MaxAge:   -1,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
				Path:     "/",
			})

			return nil
		},
	)
	if err != nil {
		log.Warn("Could not exchange the OIDC auth code and state for refresh and ID token", "err", errors.Join(errCouldNotExchange, err))

		http.Error(w, errors.Join(errCouldNotExchange, err).Error(), http.StatusInternalServerError) // All errors returned by `exchange()` are already sanitized

		return
	}

	if signedOut {
		if err := c.tpl.ExecuteTemplate(w, "redirect.html", redirectData{
			pageData: pageData{
				userData: userData{
					Locale: locale,
				},

				Page:       locale.Get("Signing you out ..."),
				PrivacyURL: c.privacyURL,
				ImprintURL: c.imprintURL,
			},

			Href: returnURL,
		}); err != nil {
			log.Warn("Could not render sign out template", "err", errors.Join(errCouldNotRenderTemplate, err))

			http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

			return
		}

		return
	}

	if err := c.tpl.ExecuteTemplate(w, "redirect.html", redirectData{
		pageData: pageData{
			userData: userData{
				Locale: locale,
			},

			Page:       locale.Get("Signing you in ..."),
			PrivacyURL: c.privacyURL,
			ImprintURL: c.imprintURL,
		},

		Href: returnURL,
	}); err != nil {
		log.Warn("Could not render sign in template", "err", errors.Join(errCouldNotRenderTemplate, err))

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}
