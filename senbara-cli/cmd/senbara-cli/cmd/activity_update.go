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

	"gopkg.in/yaml.v2"
)

var activityUpdateCommand = &cobra.Command{
	Use:     "update <id>",
	Aliases: []string{"upd", "up", "u"},
	Short:   "Update an activity",
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

		var description *string
		if viper.IsSet(descriptionKey) {
			v := viper.GetString(descriptionKey)

			description = &v
		}

		req := api.UpdateActivityJSONRequestBody{
			Date: types.Date{
				Time: viper.GetTime(dateKey),
			},
			Description: description,
			Name:        viper.GetString(nameKey),
		}

		log.Debug("Updating activity", "id", id, "request", req)

		res, err := c.UpdateActivityWithResponse(ctx, int64(id), req)
		if err != nil {
			return err
		}

		log.Debug("Updated activity", "status", res.StatusCode())

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
	addAuthFlags(activityUpdateCommand.PersistentFlags())

	activityUpdateCommand.PersistentFlags().String(nameKey, "", "Name of the activity")
	activityUpdateCommand.PersistentFlags().String(dateKey, "", "Date of the activity (format: YYYY-MM-DD)")
	activityUpdateCommand.PersistentFlags().String(descriptionKey, "", "Description of the activity (optional)")

	viper.AutomaticEnv()

	activityCommand.AddCommand(activityUpdateCommand)
}
