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
