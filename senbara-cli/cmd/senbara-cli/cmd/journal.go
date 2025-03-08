package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var journalCommand = &cobra.Command{
	Use:     "journal",
	Aliases: []string{"jour", "j"},
	Short:   "Journal entry operations",
}

func init() {
	viper.AutomaticEnv()

	indexCommand.AddCommand(journalCommand)
}
