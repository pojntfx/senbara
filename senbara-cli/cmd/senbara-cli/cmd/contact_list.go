package cmd

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"gopkg.in/yaml.v3"
)

var contactListCommand = &cobra.Command{
	Use:     "list",
	Aliases: []string{"lis", "ls", "l"},
	Short:   "List all contacts",
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

		log.Debug("Listing contacts")

		res, err := c.GetContactsWithResponse(ctx)
		if err != nil {
			return err
		}

		log.Debug("Got contacts", "status", res.StatusCode())

		if res.StatusCode() != http.StatusOK {
			return errors.New(res.Status())
		}

		log.Debug("Writing contacts to stdout")

		if err := yaml.NewEncoder(os.Stdout).Encode(res.JSON200); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	addAuthFlags(contactListCommand.PersistentFlags())

	viper.AutomaticEnv()

	contactCommand.AddCommand(contactListCommand)
}
