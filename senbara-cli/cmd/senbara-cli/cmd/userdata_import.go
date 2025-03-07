package cmd

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var userDataImportCommand = &cobra.Command{
	Use:     "import",
	Aliases: []string{"imp", "i"},
	Short:   "Import user data",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		c, err := createClient(true)
		if err != nil {
			return err
		}

		log.Debug("Importing user data, reading from stdin and streaming to API")

		reader, writer := io.Pipe()
		enc := multipart.NewWriter(writer)
		go func() {
			defer writer.Close()

			if err := func() error {
				file, err := enc.CreateFormFile("userData", "")
				if err != nil {
					return err
				}

				if _, err := io.Copy(file, os.Stdin); err != nil {
					return err
				}

				if err := enc.Close(); err != nil {
					return err
				}

				return nil
			}(); err != nil {
				log.Warn("Could not stream user data to API", "err", err)

				writer.CloseWithError(err)

				return
			}
		}()

		res, err := c.ImportUserDataWithBodyWithResponse(ctx, enc.FormDataContentType(), reader)
		if err != nil {
			return err
		}

		log.Debug("Imported user data", "status", res.StatusCode())

		if res.StatusCode() != http.StatusOK {
			return errors.New(res.Status())
		}

		return nil
	},
}

func init() {
	addAuthFlags(userDataImportCommand.PersistentFlags())

	viper.AutomaticEnv()

	userDataCommand.AddCommand(userDataImportCommand)
}
