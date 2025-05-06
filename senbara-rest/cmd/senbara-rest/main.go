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
	"github.com/pojntfx/senbara/senbara-common/pkg/authn"
	"github.com/pojntfx/senbara/senbara-common/pkg/persisters"
	v1 "github.com/pojntfx/senbara/senbara-rest/api/openapi/v1"
	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
	"github.com/pojntfx/senbara/senbara-rest/pkg/controllers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	errMissingOIDCIssuer = errors.New("missing OIDC issuer")
	errMissingPrivacyURL = errors.New("missing privacy policy URL")
	errMissingImprintURL = errors.New("missing imprint URL")
)

const (
	verboseKey           = "verbose"
	configKey            = "config"
	laddrKey             = "laddr"
	pgaddrKey            = "pgaddr"
	oidcIssuerKey        = "oidc-issuer"
	corsOriginsKey       = "cors-origins"
	privacyURLKey        = "privacy-url"
	imprintURLKey        = "imprint-url"
	contactNameKey       = "contact-name"
	contactEmailKey      = "contact-email"
	serverURLKey         = "server-url"
	serverDescriptionKey = "server-description"
)

func main() {
	cmd := &cobra.Command{
		Use:   "senbara-rest",
		Short: "Personal ERP REST API using the Go stdlib, OIDC and PostgreSQL",
		Long: `REST API for a simple personal ERP web application built with the Go standard library, OpenID Connect authentication and PostgreSQL data storage. Designed as a reference for modern REST API development with Go.

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

			if !viper.IsSet(privacyURLKey) {
				return errMissingPrivacyURL
			}

			if !viper.IsSet(imprintURLKey) {
				return errMissingImprintURL
			}

			p := persisters.NewPersister(slog.New(log.Handler().WithGroup("persister")), viper.GetString(pgaddrKey))

			if err := p.Init(ctx); err != nil {
				return err
			}

			a := authn.NewAuthner(
				slog.New(log.Handler().WithGroup("authner")),

				viper.GetString(oidcIssuerKey),
				"",
				"",
			)

			if err := a.Init(ctx); err != nil {
				return err
			}

			s, err := api.GetSwagger()
			if err != nil {
				return err
			}

			c := controllers.NewController(
				slog.New(log.Handler().WithGroup("controller")),

				p,
				a,

				s,

				viper.GetString(oidcIssuerKey),

				viper.GetString(privacyURLKey),
				viper.GetString(imprintURLKey),

				viper.GetString(contactNameKey),
				viper.GetString(contactEmailKey),

				viper.GetString(serverURLKey),
				viper.GetString(serverDescriptionKey),

				v1.Code,
			)

			log.Info("Listening", "laddr", viper.GetString(laddrKey))

			panic(http.ListenAndServe(viper.GetString(laddrKey), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				v1.SenbaraRESTHandler(
					w,
					r,

					r.Context(),
					slog.New(log.Handler().WithGroup("handler")),
					viper.GetStringSlice(corsOriginsKey),
					c,
					s,
				)
			})))
		},
	}

	cmd.PersistentFlags().BoolP(verboseKey, "v", false, "Whether to enable verbose logging")
	cmd.PersistentFlags().StringP(configKey, "c", "", "Config file to use (by default "+cmd.Use+".yaml in the XDG config directory is read if it exists)")
	cmd.PersistentFlags().StringP(laddrKey, "l", ":1337", "Listen address (port can also be set with `PORT` env variable)")
	cmd.PersistentFlags().StringP(pgaddrKey, "p", "postgresql://postgres@localhost:5432/senbara?sslmode=disable", "Database address")
	cmd.PersistentFlags().String(oidcIssuerKey, "", "OIDC Issuer (e.g. https://pojntfx.eu.auth0.com/)")
	cmd.PersistentFlags().StringArray(corsOriginsKey, []string{}, "CORS origins to allow")
	cmd.PersistentFlags().String(privacyURLKey, "", "Privacy policy URL")
	cmd.PersistentFlags().String(imprintURLKey, "", "Imprint URL")
	cmd.PersistentFlags().String(contactNameKey, "Felicitas Pojtinger", "Contact name")
	cmd.PersistentFlags().String(contactEmailKey, "felicitas@pojtinger.com", "Contact email")
	cmd.PersistentFlags().String(serverURLKey, "http://localhost:1337/", "Server URL")
	cmd.PersistentFlags().String(serverDescriptionKey, "Local development server", "Server description")

	if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
		panic(err)
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
