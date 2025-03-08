package cmd

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"gopkg.in/yaml.v2"
)

var journalGetCommand = &cobra.Command{
	Use:     "get",
	Aliases: []string{"g"},
	Short:   "List all journal entries",
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

		log.Debug("Geting journal entries")

		res, err := c.GetJournalEntriesWithResponse(ctx)
		if err != nil {
			return err
		}

		log.Debug("Got journal entries", "status", res.StatusCode())

		if res.StatusCode() != http.StatusOK {
			return errors.New(res.Status())
		}

		log.Debug("Writing journal entries to stdout")

		if err := yaml.NewEncoder(os.Stdout).Encode(res.JSON200); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	addAuthFlags(journalGetCommand.PersistentFlags())

	viper.AutomaticEnv()

	journalCommand.AddCommand(journalGetCommand)
}
