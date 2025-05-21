package cmd

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"gopkg.in/yaml.v3"
)

const (
	titleKey  = "title"
	bodyKey   = "body"
	ratingKey = "rating"
)

var journalCreateCommand = &cobra.Command{
	Use:     "create",
	Aliases: []string{"cre", "c"},
	Short:   "Create a new journal entry",
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

		req := api.CreateJournalEntryJSONRequestBody{
			Body:   viper.GetString(bodyKey),
			Rating: viper.GetInt32(ratingKey),
			Title:  viper.GetString(titleKey),
		}

		log.Debug("Creating journal entry", "request", req)

		res, err := c.CreateJournalEntryWithResponse(ctx, req)
		if err != nil {
			return err
		}

		log.Debug("Created journal entry", "status", res.StatusCode())

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
	addAuthFlags(journalCreateCommand.PersistentFlags())

	journalCreateCommand.PersistentFlags().String(titleKey, "", "Title for the journal entry")
	journalCreateCommand.PersistentFlags().String(bodyKey, "", "Body for the journal entry")
	journalCreateCommand.PersistentFlags().Int32(ratingKey, 0, "Rating for the journal entry (between 1 and 3)")

	viper.AutomaticEnv()

	journalCommand.AddCommand(journalCreateCommand)
}
