package cmd

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"gopkg.in/yaml.v2"
)

var debtSettleCommand = &cobra.Command{
	Use:     "settle <id>",
	Aliases: []string{"set", "s"},
	Short:   "Settle a debt",
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

		log.Debug("Settling debt", "id", id)

		res, err := c.SettleDebtWithResponse(ctx, int64(id))
		if err != nil {
			return err
		}

		log.Debug("Settled debt", "status", res.StatusCode())

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
	addAuthFlags(debtSettleCommand.PersistentFlags())

	debtSettleCommand.PersistentFlags().Int64(idKey, 0, "ID of the debt")

	viper.AutomaticEnv()

	debtCommand.AddCommand(debtSettleCommand)
}
