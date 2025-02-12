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

	s, err := api.GetSwagger()
	if err != nil {
		c.log.Warn("Could not get OpenAPI spec", "err", errors.Join(errCouldNotGetOpenAPISpec, err))

		return api.GetOpenAPISpec500TextResponse(errCouldNotGetOpenAPISpec.Error()), nil
	}

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
