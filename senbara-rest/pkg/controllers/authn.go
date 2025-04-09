package controllers

import (
	"net/http"

	"github.com/oapi-codegen/runtime/strictmiddleware/nethttp"
)

type contextKey int

const (
	ContextKeyNamespace contextKey = iota
)

func (c *Controller) Authenticate(r *http.Request) (string, error) {
	return c.authner.AuthenticateRequest(r)
}

func (c *Controller) Authorize(f nethttp.StrictHTTPHandlerFunc, operationID string) nethttp.StrictHTTPHandlerFunc {
	return c.authner.AuthorizeRequest(f, operationID)
}
