package controllers

import (
	"errors"
	"net/http"
	"time"

	"github.com/leonelquinteros/gotext"
	"github.com/pojntfx/senbara/senbara-common/pkg/authn"
)

type pageData struct {
	userData

	Page       string
	PrivacyURL string
	TosURL     string
	ImprintURL string

	BackURL string
}

type userData struct {
	Email     string
	LogoutURL string

	Locale *gotext.Locale
}

func (c *Controller) authorize(
	w http.ResponseWriter, r *http.Request,

	loginIfSignedOut bool,
) (
	redirected bool,

	u userData,
	status int,

	err error,
) {
	log := c.log.With(
		"loginIfSignedOut", loginIfSignedOut,
		"method", r.Method,
		"path", r.URL.Path,
	)

	log.Debug("Handling user auth")

	locale, err := c.localize(r)
	if err != nil {
		log.Warn("Could not localize auth page", "err", errors.Join(errCouldNotLocalize, err))

		http.Error(w, errCouldNotLocalize.Error(), http.StatusInternalServerError)

		return
	}

	var (
		refreshToken,
		idToken *string
	)
	rt, err := r.Cookie(refreshTokenKey)
	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			log.Debug("Failed to read refresh token cookie", "error", err)

			return false, userData{
				Locale: locale,
			}, http.StatusUnauthorized, errors.Join(errCouldNotLogin, err)
		}
	} else {
		refreshToken = &rt.Value
	}

	it, err := r.Cookie(idTokenKey)
	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			log.Debug("Failed to read ID token cookie", "error", err)

			return false, userData{
				Locale: locale,
			}, http.StatusUnauthorized, errors.Join(errCouldNotLogin, err)
		}
	} else {
		idToken = &it.Value
	}

	nextURL, email, logoutURL, err := c.authner.Authorize(
		r.Context(),

		loginIfSignedOut,

		r.Header.Get("Referer"),
		r.URL.String(),

		refreshToken,
		idToken,

		func(s string, t time.Time) error {
			http.SetCookie(w, &http.Cookie{
				Name:     refreshTokenKey,
				Value:    s,
				Expires:  t,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
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
				SameSite: http.SameSiteLaxMode,
				Path:     "/",
			})

			return nil
		},

		func(s string) error {
			http.SetCookie(w, &http.Cookie{
				Name:     stateNonceKey,
				Value:    s,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
				Path:     "/",
			})

			return nil
		},
		func(s string) error {
			http.SetCookie(w, &http.Cookie{
				Name:     pkceCodeVerifierKey,
				Value:    s,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
				Path:     "/",
			})

			return nil
		},
		func(s string) error {
			http.SetCookie(w, &http.Cookie{
				Name:     oidcNonceKey,
				Value:    s,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
				Path:     "/",
			})

			return nil
		},
	)
	if err != nil {
		if errors.Is(err, authn.ErrCouldNotLogin) {
			return false, userData{
				Locale: locale,
			}, http.StatusUnauthorized, err
		}

		return false, userData{
			Locale: locale,
		}, http.StatusInternalServerError, err
	}

	redirected = nextURL != ""
	u = userData{
		Email:     email,
		LogoutURL: logoutURL,

		Locale: locale,
	}

	if redirected {
		http.Redirect(w, r, nextURL, http.StatusFound)

		return redirected, u, http.StatusTemporaryRedirect, nil
	}

	return redirected, u, http.StatusOK, nil
}

func (c *Controller) HandleLogin(w http.ResponseWriter, r *http.Request) {
	nextURL := r.Header.Get("Referer")

	c.log.Debug("Logging in user")

	redirected, _, status, err := c.authorize(w, r, true)
	if err != nil {
		c.log.Warn("Could not authorize user for login", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	http.Redirect(w, r, nextURL, http.StatusFound)
}

func (c *Controller) HandleAuthorize(w http.ResponseWriter, r *http.Request) {
	authCode := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	log := c.log.With(
		"authCode", authCode != "",
		"state", state,
	)

	log.Debug("Handling user auth exchange")

	var (
		stateNonce,
		pkceCodeVerifier,
		oidcNonce string
	)
	sn, err := r.Cookie(stateNonceKey)
	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			log.Warn("Failed to read state nonce cookie", "err", errors.Join(errCouldNotExchange, err))

			http.Error(w, errCouldNotExchange.Error(), http.StatusInternalServerError)

			return
		}
	} else {
		stateNonce = sn.Value
	}

	pcv, err := r.Cookie(pkceCodeVerifierKey)
	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			log.Warn("Failed to read PKCE code verifier cookie", "err", errors.Join(errCouldNotExchange, err))

			http.Error(w, errCouldNotExchange.Error(), http.StatusInternalServerError)

			return
		}
	} else {
		pkceCodeVerifier = pcv.Value
	}

	on, err := r.Cookie(oidcNonceKey)
	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			log.Warn("Failed to read OIDC nonce cookie", "err", errors.Join(errCouldNotExchange, err))

			http.Error(w, errCouldNotExchange.Error(), http.StatusInternalServerError)

			return
		}
	} else {
		oidcNonce = on.Value
	}

	nextURL, _, err := c.authner.Exchange(
		r.Context(),

		authCode,
		state,

		stateNonce,
		pkceCodeVerifier,
		oidcNonce,

		func(s string, t time.Time) error {
			http.SetCookie(w, &http.Cookie{
				Name:     refreshTokenKey,
				Value:    s,
				Expires:  t,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
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
				SameSite: http.SameSiteLaxMode,
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
				SameSite: http.SameSiteLaxMode,
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
				SameSite: http.SameSiteLaxMode,
				Path:     "/",
			})

			return nil
		},

		func() error {
			http.SetCookie(w, &http.Cookie{
				Name:     stateNonceKey,
				Value:    "",
				MaxAge:   -1,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
				Path:     "/",
			})

			return nil
		},
		func() error {
			http.SetCookie(w, &http.Cookie{
				Name:     pkceCodeVerifierKey,
				Value:    "",
				MaxAge:   -1,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
				Path:     "/",
			})

			return nil
		},
		func() error {
			http.SetCookie(w, &http.Cookie{
				Name:     oidcNonceKey,
				Value:    "",
				MaxAge:   -1,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
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

	http.Redirect(w, r, nextURL, http.StatusFound)
}
