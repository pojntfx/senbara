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

	"gopkg.in/yaml.v2"
)

const (
	amountKey      = "amount"
	currencyKey    = "currency"
	descriptionKey = "description"
	youOweKey      = "you-owe"
)

var debtCreateCommand = &cobra.Command{
	Use:     "create <contact-id>",
	Aliases: []string{"cre", "c"},
	Short:   "Create a new debt",
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

		req := api.CreateDebtJSONRequestBody{
			Amount:      float32(viper.GetFloat64(amountKey)),
			ContactId:   int64(contactID),
			Currency:    viper.GetString(currencyKey),
			Description: description,
			YouOwe:      viper.GetBool(youOweKey),
		}

		log.Debug("Creating debt", "request", req)

		res, err := c.CreateDebtWithResponse(ctx, req)
		if err != nil {
			return err
		}

		log.Debug("Created debt", "status", res.StatusCode())

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
	addAuthFlags(debtCreateCommand.PersistentFlags())

	debtCreateCommand.PersistentFlags().Float64(amountKey, 0.0, "Amount of the debt")
	debtCreateCommand.PersistentFlags().String(currencyKey, "", "Currency for the debt")
	debtCreateCommand.PersistentFlags().String(descriptionKey, "", "Description of the debt")
	debtCreateCommand.PersistentFlags().Bool(youOweKey, false, "Whether you owe the debt")

	viper.AutomaticEnv()

	debtCommand.AddCommand(debtCreateCommand)
}
