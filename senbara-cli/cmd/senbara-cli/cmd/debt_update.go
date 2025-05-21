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

var debtUpdateCommand = &cobra.Command{
	Use:     "update <id>",
	Aliases: []string{"upd", "up", "u"},
	Short:   "Update a debt",
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

		req := api.UpdateDebtJSONRequestBody{
			Amount:      float32(viper.GetFloat64(amountKey)),
			Currency:    viper.GetString(currencyKey),
			Description: description,
			YouOwe:      viper.GetBool(youOweKey),
		}

		log.Debug("Updating debt", "id", id, "request", req)

		res, err := c.UpdateDebtWithResponse(ctx, int64(id), req)
		if err != nil {
			return err
		}

		log.Debug("Updated debt", "status", res.StatusCode())

		if res.StatusCode() != http.StatusOK {
			return errors.New(res.Status())
		}

		log.Debug("Writing debt to stdout")

		if err := yaml.NewEncoder(os.Stdout).Encode(res.JSON200); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	addAuthFlags(debtUpdateCommand.PersistentFlags())

	debtUpdateCommand.PersistentFlags().Float64(amountKey, 0.0, "Amount of the debt")
	debtUpdateCommand.PersistentFlags().String(currencyKey, "", "Currency for the debt")
	debtUpdateCommand.PersistentFlags().String(descriptionKey, "", "Description of the debt")
	debtUpdateCommand.PersistentFlags().Bool(youOweKey, false, "Whether you owe the debt")

	viper.AutomaticEnv()

	debtCommand.AddCommand(debtUpdateCommand)
}
