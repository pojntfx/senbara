package cmd

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/oapi-codegen/runtime/types"
	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"gopkg.in/yaml.v3"
)

const (
	nameKey = "name"
	dateKey = "date"
)

var activityCreateCommand = &cobra.Command{
	Use:     "create <contact-id>",
	Aliases: []string{"cre", "c"},
	Short:   "Create a new activity",
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

		contactID, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		var description *string
		if viper.IsSet(descriptionKey) {
			v := viper.GetString(descriptionKey)

			description = &v
		}

		req := api.CreateActivityJSONRequestBody{
			ContactId: int64(contactID),
			Date: types.Date{
				Time: viper.GetTime(dateKey),
			},
			Description: description,
			Name:        viper.GetString(nameKey),
		}

		log.Debug("Creating activity", "request", req)

		res, err := c.CreateActivityWithResponse(ctx, req)
		if err != nil {
			return err
		}

		log.Debug("Created activity", "status", res.StatusCode())

		if res.StatusCode() != http.StatusOK {
			return errors.New(res.Status())
		}

		log.Debug("Writing activity to stdout")

		if err := yaml.NewEncoder(os.Stdout).Encode(res.JSON200); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	addAuthFlags(activityCreateCommand.PersistentFlags())

	activityCreateCommand.PersistentFlags().String(nameKey, "", "Name of the activity")
	activityCreateCommand.PersistentFlags().String(dateKey, "", "Date of the activity (format: YYYY-MM-DD)")
	activityCreateCommand.PersistentFlags().String(descriptionKey, "", "Description of the activity (optional)")

	viper.AutomaticEnv()

	activityCommand.AddCommand(activityCreateCommand)
}
