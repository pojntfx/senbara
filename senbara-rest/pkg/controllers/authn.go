package controllers

import (
	"errors"
	"net/http"
)

func (b *Controller) authorize(r *http.Request) (string, error) {
	it, err := r.Cookie(idTokenKey)
	if err != nil {
		return "", errors.Join(errCouldNotLogin, err)
	}
	idToken := it.Value

	id, err := b.verifier.Verify(r.Context(), idToken)
	if err != nil {
		return "", errors.Join(errCouldNotLogin, err)
	}

	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err := id.Claims(&claims); err != nil {
		return "", errors.Join(errCouldNotLogin, err)
	}

	if !claims.EmailVerified {
		return "", errors.Join(errCouldNotLogin, errEmailNotVerified)
	}

	return claims.Email, nil
}
