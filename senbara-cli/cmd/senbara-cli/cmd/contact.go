package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var contactCommand = &cobra.Command{
	Use:     "contact",
	Aliases: []string{"con", "c"},
	Short:   "Contact operations",
}

func init() {
	viper.AutomaticEnv()

	indexCommand.AddCommand(contactCommand)
}
