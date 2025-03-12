package cmd

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/oapi-codegen/runtime/types"
	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"gopkg.in/yaml.v2"
)

const (
	firstNameKey = "first-name"
	lastNameKey  = "last-name"
	emailKey     = "email"
	pronounsKey  = "pronouns"
	nicknameKey  = "nickname"
)

var contactCreateCommand = &cobra.Command{
	Use:     "create",
	Aliases: []string{"cre", "c"},
	Short:   "Create a new contact",
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

		var nickname *string
		if viper.IsSet(nicknameKey) {
			v := viper.GetString(nicknameKey)

			nickname = &v
		}

		req := api.CreateContactJSONRequestBody{
			Email:     (types.Email)(viper.GetString(emailKey)),
			FirstName: viper.GetString(firstNameKey),
			LastName:  viper.GetString(lastNameKey),
			Nickname:  nickname,
			Pronouns:  viper.GetString(pronounsKey),
		}

		log.Debug("Creating contact", "request", req)

		res, err := c.CreateContactWithResponse(ctx, req)
		if err != nil {
			return err
		}

		log.Debug("Created contact", "status", res.StatusCode())

		if res.StatusCode() != http.StatusOK {
			return errors.New(res.Status())
		}

		log.Debug("Writing contact to stdout")

		if err := yaml.NewEncoder(os.Stdout).Encode(res.JSON200); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	addAuthFlags(contactCreateCommand.PersistentFlags())

	contactCreateCommand.PersistentFlags().String(emailKey, "", "Email address for the contact")
	contactCreateCommand.PersistentFlags().String(firstNameKey, "", "First name for the contact")
	contactCreateCommand.PersistentFlags().String(lastNameKey, "", "Last name for the contact")
	contactCreateCommand.PersistentFlags().String(nicknameKey, "", "Nickname for the contact (optional)")
	contactCreateCommand.PersistentFlags().String(pronounsKey, "", "Pronouns for the contact")

	viper.AutomaticEnv()

	contactCommand.AddCommand(contactCreateCommand)
}
