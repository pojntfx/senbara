package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
)

func (c *Controller) GetOpenAPISpec(ctx context.Context, request api.GetOpenAPISpecRequestObject) (api.GetOpenAPISpecResponseObject, error) {
	c.log.Debug("Handling getting OpenAPI Spec")

	s := *c.spec
	s.Info.Extensions[api.PrivacyPolicyExtensionKey] = c.privacyURL
	s.Info.TermsOfService = c.tosURL
	s.Info.Contact.Name = c.contactName
	s.Info.Contact.Email = c.contactEmail
	s.Servers[0].URL = c.serverURL
	s.Servers[0].Description = c.serverDescription
	s.Components.SecuritySchemes["oidc"].Value.OpenIdConnectUrl = c.oidcDiscoveryURL

	if c.oidcDcrInitialAccessTokenPortalUrl != "" {
		s.Components.SecuritySchemes["oidc"].Value.Extensions[api.OidcDcrInitialAccessTokenPortalUrlExtensionKey] = c.oidcDcrInitialAccessTokenPortalUrl
	}

	reader, writer := io.Pipe()
	enc := json.NewEncoder(writer)
	go func() {
		defer writer.Close()

		if err := enc.Encode(s); err != nil {
			c.log.Warn("Could not encode OpenAPI response", "err", errors.Join(errCouldNotEncodeResponse, err))

			writer.CloseWithError(errCouldNotEncodeResponse)

			return
		}
	}()

	return api.GetOpenAPISpec200ApplicationoctetStreamResponse{
		Body: reader,
		Headers: api.GetOpenAPISpec200ResponseHeaders{
			ContentType: "application/json",
		},
	}, nil
}
