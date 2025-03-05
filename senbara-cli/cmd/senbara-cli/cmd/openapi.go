package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var openapiCmd = &cobra.Command{
	Use:     "openapi",
	Aliases: []string{"oai", "o"},
	Short:   "OpenAPI operations",
}

func init() {
	viper.AutomaticEnv()

	rootCmd.AddCommand(openapiCmd)
}
