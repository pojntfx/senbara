package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var activityCommand = &cobra.Command{
	Use:     "activity",
	Aliases: []string{"act", "a"},
	Short:   "Activity operations",
}

func init() {
	viper.AutomaticEnv()

	indexCommand.AddCommand(activityCommand)
}
