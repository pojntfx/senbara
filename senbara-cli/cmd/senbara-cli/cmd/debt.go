package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var debtCommand = &cobra.Command{
	Use:     "debt",
	Aliases: []string{"deb", "d"},
	Short:   "Debt operations",
}

func init() {
	viper.AutomaticEnv()

	indexCommand.AddCommand(debtCommand)
}
