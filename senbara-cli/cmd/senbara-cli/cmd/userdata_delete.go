package cmd

import (
	"context"
	"errors"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var userDataDeleteCommand = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"del", "d"},
	Short:   "Delete all user data",
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

		log.Debug("Deleting user data")

		res, err := c.DeleteUserDataWithResponse(ctx)
		if err != nil {
			return err
		}

		log.Debug("Deleted user data", "status", res.StatusCode())

		if res.StatusCode() != http.StatusOK {
			return errors.New(res.Status())
		}

		return nil
	},
}

func init() {
	addAuthFlags(userDataDeleteCommand.PersistentFlags())

	viper.AutomaticEnv()

	userDataCommand.AddCommand(userDataDeleteCommand)
}
