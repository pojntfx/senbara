package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/adrg/xdg"
	"github.com/pojntfx/senbara/senbara-common/pkg/persisters"
	senbaraForms "github.com/pojntfx/senbara/senbara-forms/api/senbara-forms"
	"github.com/pojntfx/senbara/senbara-forms/pkg/controllers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	errMissingOIDCIssuer      = errors.New("missing OIDC issuer")
	errMissingOIDCClientID    = errors.New("missing OIDC client ID")
	errMissingOIDCRedirectURL = errors.New("missing OIDC redirect URL")
	errMissingPrivacyURL      = errors.New("missing privacy policy URL")
	errMissingImprintURL      = errors.New("missing imprint URL")
)

const (
	verboseKey         = "verbose"
	configKey          = "config"
	laddrKey           = "laddr"
	pgaddrKey          = "pgaddr"
	oidcIssuerKey      = "oidc-issuer"
	oidcClientIDKey    = "oidc-client-id"
	oidcRedirectURLKey = "oidc-redirect-url"
	privacyURLKey      = "privacy-url"
	imprintURLKey      = "imprint-url"
)

func main() {
	cmd := &cobra.Command{
		Use:   "senbara-forms",
		Short: "Personal ERP web app using HTML forms, OIDC and PostgreSQL",
		Long: `Simple personal ERP web application built with HTML forms, OpenID Connect authentication and PostgreSQL data storage. Designed as a reference for modern JS-free "Web 2.0" development with Go.

For more information, please visit https://github.com/pojntfx/senbara.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			opts := &slog.HandlerOptions{}
			if viper.GetBool(verboseKey) {
				opts.Level = slog.LevelDebug
			}
			log := slog.New(slog.NewJSONHandler(os.Stderr, opts))

			if viper.IsSet(configKey) {
				viper.SetConfigFile(viper.GetString(configKey))
				if err := viper.ReadInConfig(); err != nil {
					return err
				}
			} else {
				viper.SetConfigName(cmd.Use)
				viper.AddConfigPath(xdg.ConfigHome)
				if err := viper.ReadInConfig(); err != nil && !errors.As(err, &viper.ConfigFileNotFoundError{}) {
					return err
				}
			}

			if v := os.Getenv("PORT"); v != "" {
				log.Info("Using port from PORT env variable")

				la, err := net.ResolveTCPAddr("tcp", viper.GetString(laddrKey))
				if err != nil {
					return err
				}

				p, err := strconv.Atoi(v)
				if err != nil {
					return err
				}

				la.Port = p

				viper.Set(laddrKey, la.String())
			}

			if !viper.IsSet(oidcIssuerKey) {
				return errMissingOIDCIssuer
			}

			if !viper.IsSet(oidcClientIDKey) {
				return errMissingOIDCClientID
			}

			if !viper.IsSet(oidcRedirectURLKey) {
				return errMissingOIDCRedirectURL
			}

			if !viper.IsSet(privacyURLKey) {
				return errMissingPrivacyURL
			}

			if !viper.IsSet(imprintURLKey) {
				return errMissingImprintURL
			}

			p := persisters.NewPersister(log, viper.GetString(pgaddrKey))

			if err := p.Init(ctx); err != nil {
				return err
			}

			c := controllers.NewController(
				p,

				viper.GetString(oidcIssuerKey),
				viper.GetString(oidcClientIDKey),
				viper.GetString(oidcRedirectURLKey),

				viper.GetString(privacyURLKey),
				viper.GetString(imprintURLKey),
			)

			if err := c.Init(ctx); err != nil {
				return err
			}

			log.Info("Listening", "laddr", viper.GetString(laddrKey))

			panic(http.ListenAndServe(viper.GetString(laddrKey), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				senbaraForms.SenbaraFormsHandler(w, r, c)
			})))
		},
	}

	cmd.PersistentFlags().BoolP(verboseKey, "v", false, "Whether to enable verbose logging")
	cmd.PersistentFlags().StringP(configKey, "c", "", "Config file to use (by default "+cmd.Use+".yaml in the XDG config directory is read if it exists)")
	cmd.PersistentFlags().StringP(laddrKey, "l", ":1337", "Listen address (port can also be set with `PORT` env variable)")
	cmd.PersistentFlags().StringP(pgaddrKey, "p", "postgresql://postgres@localhost:5432/senbara_forms?sslmode=disable", "Database address")
	cmd.PersistentFlags().String(oidcIssuerKey, "", "OIDC Issuer (i.e. https://pojntfx.eu.auth0.com/)")
	cmd.PersistentFlags().String(oidcClientIDKey, "", "OIDC Client ID (i.e. myoidcclientid))")
	cmd.PersistentFlags().String(oidcRedirectURLKey, "http://localhost:1337/authorize", "OIDC redirect URL")
	cmd.PersistentFlags().String(privacyURLKey, "", "Privacy policy URL")
	cmd.PersistentFlags().String(imprintURLKey, "", "Imprint URL")

	if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
		panic(err)
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
