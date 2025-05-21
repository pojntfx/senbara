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
	birthdayKey = "birthday"
	addressKey  = "address"
	notesKey    = "notes"
)

var contactUpdateCommand = &cobra.Command{
	Use:     "update <id>",
	Aliases: []string{"upd", "up", "u"},
	Short:   "Update a contact",
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

		var nickname *string
		if viper.IsSet(nicknameKey) {
			v := viper.GetString(nicknameKey)

			nickname = &v
		}

		var birthday *types.Date
		if viper.IsSet(birthdayKey) {
			birthday = &types.Date{
				Time: viper.GetTime(birthdayKey),
			}
		}

		var address *string
		if viper.IsSet(addressKey) {
			v := viper.GetString(addressKey)

			address = &v
		}

		var notes *string
		if viper.IsSet(notesKey) {
			v := viper.GetString(notesKey)

			notes = &v
		}

		req := api.UpdateContactJSONRequestBody{
			Address:   address,
			Birthday:  birthday,
			Email:     (types.Email)(viper.GetString(emailKey)),
			FirstName: viper.GetString(firstNameKey),
			LastName:  viper.GetString(lastNameKey),
			Nickname:  nickname,
			Notes:     notes,
			Pronouns:  viper.GetString(pronounsKey),
		}

		log.Debug("Updating contact", "id", id, "request", req)

		res, err := c.UpdateContactWithResponse(ctx, int64(id), req)
		if err != nil {
			return err
		}

		log.Debug("Updated contact", "status", res.StatusCode())

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
	addAuthFlags(contactUpdateCommand.PersistentFlags())

	contactUpdateCommand.PersistentFlags().String(addressKey, "", "Address for the contact (optional)")
	contactUpdateCommand.PersistentFlags().String(birthdayKey, "", "Birthday for the contact (optional, format: YYYY-MM-DD)")
	contactUpdateCommand.PersistentFlags().String(emailKey, "", "Email address for the contact")
	contactUpdateCommand.PersistentFlags().String(firstNameKey, "", "First name for the contact")
	contactUpdateCommand.PersistentFlags().String(lastNameKey, "", "Last name for the contact")
	contactUpdateCommand.PersistentFlags().String(nicknameKey, "", "Nickname for the contact (optional)")
	contactUpdateCommand.PersistentFlags().String(notesKey, "", "Notes for the contact (optional)")
	contactUpdateCommand.PersistentFlags().String(pronounsKey, "", "Pronouns for the contact")

	viper.AutomaticEnv()

	contactCommand.AddCommand(contactUpdateCommand)
}
