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

var userDataGetCommand = &cobra.Command{
	Use:     "get",
	Aliases: []string{"g"},
	Short:   "Export all user data",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		log.Debug("Getting user data")

		c, err := createClient(true)
		if err != nil {
			return err
		}

		res, err := c.ExportUserData(ctx)
		if err != nil {
			return err
		}

		log.Debug("Received user data", "status", res.StatusCode)

		if res.StatusCode != http.StatusOK {
			return errors.New(res.Status)
		}

		log.Debug("Writing user data to stdout")

		if _, err := io.Copy(os.Stdout, res.Body); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	addAuthFlags(userDataGetCommand.PersistentFlags())

	viper.AutomaticEnv()

	userDataCommand.AddCommand(userDataGetCommand)
}
