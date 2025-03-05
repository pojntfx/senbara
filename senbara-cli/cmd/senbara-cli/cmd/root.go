package cmd

import (
	"errors"
	"log/slog"
	"os"
	"strings"

	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	verboseKey = "verbose"
	configKey  = "config"
	raddrKey   = "laddr"
)

var (
	log     *slog.Logger
	rootCmd = &cobra.Command{
		Use:   "senbara-cli",
		Short: "CLI to interact with the Senbara REST API",
		Long: `CLI to interact with the Senbara REST API. Designed as a reference for modern CLI development with Go.

For more information, please visit https://github.com/pojntfx/senbara.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			opts := &slog.HandlerOptions{}
			if viper.GetBool(verboseKey) {
				opts.Level = slog.LevelDebug
			}
			log = slog.New(slog.NewJSONHandler(os.Stderr, opts))

			if viper.IsSet(configKey) {
				viper.SetConfigFile(viper.GetString(configKey))
				if err := viper.ReadInConfig(); err != nil {
					return err
				}
			} else {
				viper.SetConfigName(cmd.Root().Use)
				viper.AddConfigPath(xdg.ConfigHome)
				if err := viper.ReadInConfig(); err != nil && !errors.As(err, &viper.ConfigFileNotFoundError{}) {
					return err
				}
			}

			return nil
		},
	}
)

func Execute() error {
	rootCmd.PersistentFlags().BoolP(verboseKey, "v", false, "Whether to enable verbose logging")
	rootCmd.PersistentFlags().StringP(configKey, "c", "", "Config file to use (by default "+rootCmd.Use+".yaml in the XDG config directory is read if it exists)")
	rootCmd.PersistentFlags().StringP(raddrKey, "r", "http://localhost:1337/", "Remote address")

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		return err
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	return rootCmd.Execute()
}
