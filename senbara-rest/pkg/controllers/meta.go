package controllers

import (
	"context"
	"errors"
	"io"

	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
	"gopkg.in/yaml.v2"
)

func (c *Controller) GetOpenAPISpec(ctx context.Context, request api.GetOpenAPISpecRequestObject) (api.GetOpenAPISpecResponseObject, error) {
	c.log.Debug("Handling getting OpenAPI Spec")

	s := *c.spec
	s.Info.Description = `REST API for a simple personal ERP web application built with the Go standard library, OpenID Connect authentication and PostgreSQL data storage. Designed as a reference for modern REST API development with Go.

Imprint: ` + c.imprintURL
	s.Info.TermsOfService = c.privacyURL
	s.Info.Contact.Name = c.contactName
	s.Info.Contact.URL = c.contactURL
	s.Info.Contact.Email = c.contactEmail
	s.Servers[0].URL = c.serverURL
	s.Servers[0].Description = c.serverDescription
	s.Components.SecuritySchemes["oidc"].Value.OpenIdConnectUrl = c.oidcDiscoveryURL

	reader, writer := io.Pipe()
	enc := yaml.NewEncoder(writer)
	go func() {
		defer writer.Close()
		defer enc.Close()

		if err := enc.Encode(s); err != nil {
			c.log.Warn("Could not encode OpenAPI response", "err", errors.Join(errCouldNotEncodeResponse, err))

			writer.CloseWithError(errCouldNotEncodeResponse)

			return
		}
	}()

	return api.GetOpenAPISpec200ApplicationyamlResponse{
		Body: reader,
	}, nil
}
