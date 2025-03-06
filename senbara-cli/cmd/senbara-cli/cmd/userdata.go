package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var userDataCommand = &cobra.Command{
	Use:     "userdata",
	Aliases: []string{"user", "usr", "use", "u"},
	Short:   "User data operations",
}

func init() {
	viper.AutomaticEnv()

	indexCommand.AddCommand(userDataCommand)
}
