package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var codeCmd = &cobra.Command{
	Use:     "code",
	Aliases: []string{"cod", "c"},
	Short:   "Code operations",
}

func init() {
	viper.AutomaticEnv()

	rootCmd.AddCommand(codeCmd)
}
