package authn

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/oapi-codegen/runtime/strictmiddleware/nethttp"
	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
)

type contextKey int

const (
	ContextKeyNamespace contextKey = iota
)

// AuthenticateRequest reads the OIDC token from the request headers, verifies it, and returns the user's verified email
func (c *Authner) AuthenticateRequest(r *http.Request) (string, error) {
	idToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

	c.log.Debug("Starting authentication",
		"method", r.Method,
		"path", r.URL.Path,
	)

	id, err := c.verifier.Verify(r.Context(), idToken)
	if err != nil {
		c.log.Debug("ID token verification failed", "error", errors.Join(ErrCouldNotLogin, err))

		return "", ErrCouldNotLogin
	}

	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err := id.Claims(&claims); err != nil {
		c.log.Debug("Failed to parse ID token claims", "error", errors.Join(ErrCouldNotLogin, err))

		return "", ErrCouldNotLogin
	}

	if !claims.EmailVerified {
		c.log.Debug("Email from ID token claims not verified, user is unauthenticated", "email", claims.Email, "error", errors.Join(ErrCouldNotLogin, errEmailNotVerified))

		return "", errors.Join(ErrCouldNotLogin, errEmailNotVerified)
	}

	c.log.Debug("Authentication successful", "email", claims.Email)

	return claims.Email, nil
}

// AuthorizeRequest checks if a route requires authentication, authenticates the user, and adds the user's verified email to the context
func (c *Authner) AuthorizeRequest(f nethttp.StrictHTTPHandlerFunc, operationID string) nethttp.StrictHTTPHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (response interface{}, err error) {
		if _, ok := r.Context().Value(api.OidcScopes).([]string); ok {
			c.log.Debug("Starting authorization",
				"method", r.Method,
				"path", r.URL.Path,
			)

			namespace, err := c.AuthenticateRequest(r)
			if err != nil {
				c.log.Debug("Could not re-authenticate to extract namespace", "error", errors.Join(ErrCouldNotLogin, err))

				return struct{}{}, ErrCouldNotLogin
			}

			ctx = context.WithValue(r.Context(), ContextKeyNamespace, namespace)

			c.log.Debug("Authorization successful", "email", namespace)
		} else {
			c.log.Debug("Authorization skipped since route doesn't require it",
				"method", r.Method,
				"path", r.URL.Path,
			)
		}

		return f(ctx, w, r, request)
	}
}
