package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var statisticsCommand = &cobra.Command{
	Use:     "statistics",
	Aliases: []string{"stats", "stat", "sta", "st"},
	Short:   "Statistics operations",
}

func init() {
	viper.AutomaticEnv()

	indexCommand.AddCommand(statisticsCommand)
}
