package cmd

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var openapiGetCommand = &cobra.Command{
	Use:     "get",
	Aliases: []string{"g"},
	Short:   "Get the OpenAPI spec",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		c, err := createClient(false)
		if err != nil {
			return err
		}

		log.Debug("Getting OpenAPI spec")

		res, err := c.GetOpenAPISpec(ctx)
		if err != nil {
			return err
		}

		log.Debug("Received OpenAPI spec", "status", res.StatusCode)

		if res.StatusCode != http.StatusOK {
			return errors.New(res.Status)
		}

		log.Debug("Writing OpenAPI spec to stdout")

		if _, err := io.Copy(os.Stdout, res.Body); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	viper.AutomaticEnv()

	openapiCommand.AddCommand(openapiGetCommand)
}
