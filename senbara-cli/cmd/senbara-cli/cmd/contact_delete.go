package cmd

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"gopkg.in/yaml.v3"
)

var contactDeleteCommand = &cobra.Command{
	Use:     "delete <id>",
	Aliases: []string{"del", "rm", "d"},
	Short:   "Delete a contact",
	Args:    cobra.ExactArgs(1),
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

		id, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		log.Debug("Deleting contact", "id", id)

		res, err := c.DeleteContactWithResponse(ctx, int64(id))
		if err != nil {
			return err
		}

		log.Debug("Deleted contact", "status", res.StatusCode())

		if res.StatusCode() != http.StatusOK {
			return errors.New(res.Status())
		}

		log.Debug("Writing contact to stdout")

		if err := yaml.NewEncoder(os.Stdout).Encode(res.JSON200); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	addAuthFlags(contactDeleteCommand.PersistentFlags())

	contactDeleteCommand.PersistentFlags().Int64(idKey, 0, "ID of the contact")

	viper.AutomaticEnv()

	contactCommand.AddCommand(contactDeleteCommand)
}
