package controllers

import (
	"errors"
	"log"
	"net/http"
	"net/url"
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

// authorize authorizes a user based on a request and returns their data if it's available.
// If `w` is not nil and a user is not authorized yet, authorize will redirect the user to
// the sign in URL.
func (b *Controller) authorize(w http.ResponseWriter, r *http.Request) (bool, userData, int, error) {
	if w == nil {
		locale, err := b.localize(r)
		if err != nil {
			return false, userData{}, http.StatusInternalServerError, errors.Join(errCouldNotLocalize, err)
		}

		if _, err := r.Cookie(refreshTokenKey); err != nil {
			return false, userData{
				Locale: locale,
			}, http.StatusOK, nil
		}

		it, err := r.Cookie(idTokenKey)
		if err != nil {
			return false, userData{
				Locale: locale,
			}, http.StatusOK, nil
		}

		id, err := b.verifier.Verify(r.Context(), it.Value)
		if err != nil {
			return false, userData{
				Locale: locale,
			}, http.StatusOK, nil
		}

		var claims struct {
			Email         string `json:"email"`
			EmailVerified bool   `json:"email_verified"`
		}
		if err := id.Claims(&claims); err != nil {
			return false, userData{}, http.StatusUnauthorized, errors.Join(errCouldNotLogin, err)
		}

		if !claims.EmailVerified {
			return false, userData{
				Locale: locale,
			}, http.StatusOK, nil
		}

		logoutURL, err := url.Parse(b.oidcIssuer)
		if err != nil {
			return false, userData{}, http.StatusUnauthorized, errors.Join(errCouldNotLogin, err)
		}

		q := logoutURL.Query()
		q.Set("id_token_hint", it.Value)
		q.Set("post_logout_redirect_uri", b.oidcRedirectURL)
		logoutURL.RawQuery = q.Encode()

		logoutURL = logoutURL.JoinPath("oidc", "logout")

		return false, userData{
			Email:     claims.Email,
			LogoutURL: logoutURL.String(),

			Locale: locale,
		}, http.StatusOK, nil
	}

	locale, err := b.localize(r)
	if err != nil {
		return false, userData{}, http.StatusInternalServerError, errors.Join(errCouldNotLocalize, err)
	}

	rt, err := r.Cookie(refreshTokenKey)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			http.Redirect(w, r, b.config.AuthCodeURL(b.oidcRedirectURL), http.StatusFound)

			return true, userData{}, http.StatusTemporaryRedirect, nil
		}

		return false, userData{}, http.StatusUnauthorized, errors.Join(errCouldNotLogin, err)
	}
	refreshToken := rt.Value

	it, err := r.Cookie(idTokenKey)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			http.Redirect(w, r, b.config.AuthCodeURL(b.oidcRedirectURL), http.StatusFound)

			return true, userData{}, http.StatusTemporaryRedirect, nil
		}

		return false, userData{}, http.StatusUnauthorized, errors.Join(errCouldNotLogin, err)
	}
	idToken := it.Value

	id, err := b.verifier.Verify(r.Context(), idToken)
	if err != nil {
		oauth2Token, err := b.config.TokenSource(r.Context(), &oauth2.Token{
			RefreshToken: refreshToken,
		}).Token()
		if err != nil {
			http.Redirect(w, r, b.config.AuthCodeURL(b.oidcRedirectURL), http.StatusFound)

			return true, userData{}, http.StatusOK, nil
		}

		var ok bool
		idToken, ok = oauth2Token.Extra("id_token").(string)
		if !ok {
			http.Redirect(w, r, b.config.AuthCodeURL(b.oidcRedirectURL), http.StatusFound)

			return true, userData{}, http.StatusOK, nil
		}

		id, err = b.verifier.Verify(r.Context(), idToken)
		if err != nil {
			http.Redirect(w, r, b.config.AuthCodeURL(b.oidcRedirectURL), http.StatusFound)

			return true, userData{}, http.StatusOK, nil
		}

		if refreshToken = oauth2Token.RefreshToken; refreshToken != "" {
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
		return false, userData{}, http.StatusUnauthorized, errors.Join(errCouldNotLogin, err)
	}

	if !claims.EmailVerified {
		return false, userData{}, http.StatusUnauthorized, errors.Join(errCouldNotLogin, errEmailNotVerified)
	}

	logoutURL, err := url.Parse(b.oidcIssuer)
	if err != nil {
		return false, userData{}, http.StatusUnauthorized, errors.Join(errCouldNotLogin, err)
	}

	q := logoutURL.Query()
	q.Set("id_token_hint", idToken)
	q.Set("post_logout_redirect_uri", b.oidcRedirectURL)
	logoutURL.RawQuery = q.Encode()

	logoutURL = logoutURL.JoinPath("oidc", "logout")

	return false, userData{
		Email:     claims.Email,
		LogoutURL: logoutURL.String(),

		Locale: locale,
	}, http.StatusOK, nil
}

type redirectData struct {
	pageData
	Href string
}

func (b *Controller) HandleLogin(w http.ResponseWriter, r *http.Request) {
	redirected, _, status, err := b.authorize(w, r)
	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusFound)
}

func (b *Controller) HandleAuthorize(w http.ResponseWriter, r *http.Request) {
	locale, err := b.localize(r)
	if err != nil {
		log.Println(errCouldNotLocalize, err)

		http.Error(w, errCouldNotLocalize.Error(), http.StatusInternalServerError)

		return
	}

	authCode := r.URL.Query().Get("code")

	// Sign out
	if authCode == "" {
		http.SetCookie(w, &http.Cookie{
			Name:   refreshTokenKey,
			Value:  "",
			MaxAge: -1,
		})

		http.SetCookie(w, &http.Cookie{
			Name:   idTokenKey,
			Value:  "",
			MaxAge: -1,
		})

		if err := b.tpl.ExecuteTemplate(w, "redirect.html", redirectData{
			pageData: pageData{
				userData: userData{
					Locale: locale,
				},

				Page:       "Signing You Out ...",
				PrivacyURL: b.privacyURL,
				ImprintURL: b.imprintURL,
			},
			Href: "/",
		}); err != nil {
			log.Println(errCouldNotRenderTemplate, err)

			http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

			return
		}

		return
	}

	// Sign in
	oauth2Token, err := b.config.Exchange(r.Context(), authCode)
	if err != nil {
		log.Println(errCouldNotLogin, err)

		http.Error(w, errCouldNotLogin.Error(), http.StatusUnauthorized)

		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenKey,
		Value:    oauth2Token.RefreshToken,
		Expires:  time.Now().Add(time.Hour * 24 * 365),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	idToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		log.Println(errCouldNotLogin, err)

		http.Error(w, errCouldNotLogin.Error(), http.StatusUnauthorized)

		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     idTokenKey,
		Value:    idToken,
		Expires:  oauth2Token.Expiry,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	if err := b.tpl.ExecuteTemplate(w, "redirect.html", redirectData{
		pageData: pageData{
			userData: userData{
				Locale: locale,
			},

			Page:       "Signing You In ...",
			PrivacyURL: b.privacyURL,
			ImprintURL: b.imprintURL,
		},
		Href: "/",
	}); err != nil {
		log.Println(errCouldNotRenderTemplate, err)

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}
