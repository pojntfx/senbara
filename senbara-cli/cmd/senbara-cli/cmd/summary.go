package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var summaryCommand = &cobra.Command{
	Use:     "summary",
	Aliases: []string{"sum", "s"},
	Short:   "Summary operations",
}

func init() {
	viper.AutomaticEnv()

	indexCommand.AddCommand(summaryCommand)
}
