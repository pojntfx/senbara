package cmd

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	errMissingToken = errors.New("missing token")
)

const (
	verboseKey = "verbose"
	configKey  = "config"
	raddrKey   = "laddr"
	tokenKey   = "token"
)

var (
	log          *slog.Logger
	indexCommand = &cobra.Command{
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

				log.Debug("Config key set, reading from file", "path", viper.GetViper().ConfigFileUsed())

				if err := viper.ReadInConfig(); err != nil {
					return err
				}
			} else {
				configBase := xdg.ConfigHome
				configName := cmd.Root().Use

				viper.SetConfigName(configName)
				viper.AddConfigPath(configBase)

				log.Debug("Config key not set, reading from default location", "path", filepath.Join(configBase, configName))

				if err := viper.ReadInConfig(); err != nil && !errors.As(err, &viper.ConfigFileNotFoundError{}) {
					return err
				}
			}

			return nil
		},
	}
)

func Execute() error {
	indexCommand.PersistentFlags().BoolP(verboseKey, "v", false, "Whether to enable verbose logging")
	indexCommand.PersistentFlags().StringP(configKey, "c", "", "Config file to use (by default "+indexCommand.Use+".yaml in the XDG config directory is read if it exists)")
	indexCommand.PersistentFlags().StringP(raddrKey, "r", "http://localhost:1337/", "Remote address")

	if err := viper.BindPFlags(indexCommand.PersistentFlags()); err != nil {
		return err
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	return indexCommand.Execute()
}

func addAuthFlags(f *pflag.FlagSet) {
	f.String(tokenKey, "", "Bearer token to authenticate with")
}

func createClient(auth bool) (*api.ClientWithResponses, error) {
	opts := []api.ClientOption{}
	if auth {
		log.Debug("Creating authenticated client")

		if !viper.IsSet(tokenKey) {
			log.Debug("Missing token")

			return nil, errMissingToken
		}

		a, err := securityprovider.NewSecurityProviderBearerToken(viper.GetString(tokenKey))
		if err != nil {
			log.Debug("Could not set up security provider for token")

			return nil, err
		}

		opts = append(opts, api.WithRequestEditorFn(a.Intercept))
	} else {
		log.Debug("Creating unauthenticated client")
	}

	return api.NewClientWithResponses(
		viper.GetString(raddrKey),
		opts...,
	)
}
