package controllers

import (
	"context"
	"errors"
	"io"

	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
)

func (c *Controller) GetSourceCode(ctx context.Context, request api.GetSourceCodeRequestObject) (api.GetSourceCodeResponseObject, error) {
	c.log.Debug("Handling getting source code")

	reader, writer := io.Pipe()
	go func() {
		defer writer.Close()

		if _, err := writer.Write(c.code); err != nil {
			c.log.Warn("Could not write source code", "err", errors.Join(errCouldNotWriteResponse, err))

			writer.CloseWithError(errCouldNotWriteResponse)

			return
		}
	}()

	return api.GetSourceCode200ApplicationgzipResponse{
		Body: reader,
		Headers: api.GetSourceCode200ResponseHeaders{
			ContentDisposition: `attachment; filename="code.tar.gz"`,
		},
	}, nil
}
