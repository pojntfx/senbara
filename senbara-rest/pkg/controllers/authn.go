package controllers

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strings"
)

type contextKey int

const (
	ContextKeyNamespace contextKey = iota
)

func (c *Controller) Authorize(next http.Handler, pathsThatDontRequireAuth []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if slices.Contains(pathsThatDontRequireAuth, r.URL.Path) {
			c.log.Debug("Auth skipped since path doesn't require auth", "path", r.URL.Path, "pathsThatDontRequireAuth", pathsThatDontRequireAuth)

			next.ServeHTTP(w, r)

			return
		}

		idToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

		c.log.Debug("Starting auth",
			"method", r.Method,
			"path", r.URL.Path,
		)

		id, err := c.verifier.Verify(r.Context(), idToken)
		if err != nil {
			c.log.Debug("ID token verification failed", "error", errors.Join(errCouldNotLogin, err))

			http.Error(w, errCouldNotLogin.Error(), http.StatusUnauthorized)

			return
		}

		var claims struct {
			Email         string `json:"email"`
			EmailVerified bool   `json:"email_verified"`
		}
		if err := id.Claims(&claims); err != nil {
			c.log.Debug("Failed to parse ID token claims", "error", errors.Join(errCouldNotLogin, err))

			http.Error(w, errCouldNotLogin.Error(), http.StatusUnauthorized)

			return
		}

		if !claims.EmailVerified {
			c.log.Debug("Email from ID token claims not verified, user is unauthorized", "email", claims.Email)

			http.Error(w, errCouldNotLogin.Error(), http.StatusUnauthorized)

			return
		}

		c.log.Debug("Auth successful", "email", claims.Email)

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ContextKeyNamespace, claims.Email)))
	})
}
