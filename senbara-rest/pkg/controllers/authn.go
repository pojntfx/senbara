package controllers

import (
	"errors"
	"net/http"
)

type userData struct {
	Email string
}

func (b *Controller) authorize(r *http.Request) (bool, userData, int, error) {
	it, err := r.Cookie(idTokenKey)
	if err != nil {
		return false, userData{}, http.StatusUnauthorized, errors.Join(errCouldNotLogin, err)
	}
	idToken := it.Value

	id, err := b.verifier.Verify(r.Context(), idToken)
	if err != nil {
		return false, userData{}, http.StatusUnauthorized, errors.Join(errCouldNotLogin, err)
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

	return false, userData{
		Email: claims.Email,
	}, http.StatusOK, nil
}
