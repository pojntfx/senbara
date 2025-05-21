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

const (
	idKey = "id"
)

var journalGetCommand = &cobra.Command{
	Use:     "get",
	Aliases: []string{"g"},
	Short:   "Get a specific journal entry",
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

		log.Debug("Getting journal entry", "id", id)

		res, err := c.GetJournalEntryWithResponse(ctx, int64(id))
		if err != nil {
			return err
		}

		log.Debug("Got journal entry", "status", res.StatusCode())

		if res.StatusCode() != http.StatusOK {
			return errors.New(res.Status())
		}

		log.Debug("Writing journal entry to stdout")

		if err := yaml.NewEncoder(os.Stdout).Encode(res.JSON200); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	addAuthFlags(journalGetCommand.PersistentFlags())

	journalGetCommand.PersistentFlags().Int64(idKey, 0, "ID of the journal entry")

	viper.AutomaticEnv()

	journalCommand.AddCommand(journalGetCommand)
}
