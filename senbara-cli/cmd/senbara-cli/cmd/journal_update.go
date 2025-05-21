package cmd

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"gopkg.in/yaml.v3"
)

var journalUpdateCommand = &cobra.Command{
	Use:     "update <id>",
	Aliases: []string{"upd", "up", "u"},
	Short:   "Update a journal entry",
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

		req := api.UpdateJournalEntryJSONRequestBody{
			Body:   viper.GetString(bodyKey),
			Rating: viper.GetInt32(ratingKey),
			Title:  viper.GetString(titleKey),
		}

		log.Debug("Updating journal entry", "id", id, "request", req)

		res, err := c.UpdateJournalEntryWithResponse(ctx, int64(id), req)
		if err != nil {
			return err
		}

		log.Debug("Updated journal entry", "status", res.StatusCode())

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
	addAuthFlags(journalUpdateCommand.PersistentFlags())

	journalUpdateCommand.PersistentFlags().String(titleKey, "", "Title for the journal entry")
	journalUpdateCommand.PersistentFlags().String(bodyKey, "", "Body for the journal entry")
	journalUpdateCommand.PersistentFlags().Int32(ratingKey, 0, "Rating for the journal entry (between 1 and 3)")

	viper.AutomaticEnv()

	journalCommand.AddCommand(journalUpdateCommand)
}
