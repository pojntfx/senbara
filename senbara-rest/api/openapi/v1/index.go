package v1

//go:generate tar czf code.tar.gz --exclude .git --exclude */api/openapi/v1/code.tar.gz -C ../../../ .

import (
	"context"
	_ "embed"
	"log/slog"
	"net/http"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"github.com/rs/cors"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/pojntfx/senbara/senbara-common/pkg/authn"
	"github.com/pojntfx/senbara/senbara-common/pkg/persisters"
	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
	"github.com/pojntfx/senbara/senbara-rest/pkg/controllers"
)

//go:embed code.tar.gz
var Code []byte

var (
	p *persisters.Persister
	a *authn.Authner
	c *controllers.Controller
	s *openapi3.T
)

func SenbaraRESTHandler(
	w http.ResponseWriter,
	r *http.Request,

	ctx context.Context,
	log *slog.Logger,
	o []string,
	c *controllers.Controller,
	s *openapi3.T,
) {
	mux := http.NewServeMux()

	mux.Handle(
		"/",
		middleware.OapiRequestValidatorWithOptions(
			s,
			&middleware.Options{
				SilenceServersWarning: true,
				Options: openapi3filter.Options{
					AuthenticationFunc: func(ctx context.Context, ai *openapi3filter.AuthenticationInput) error {
						_, err := c.Authenticate(r)

						return err
					},
				},
			},
		)(
			api.Handler(
				api.NewStrictHandler(c, []api.StrictMiddlewareFunc{
					c.Authorize,
				}),
			),
		),
	)

	if len(o) <= 0 {
		mux.ServeHTTP(w, r)
	} else {
		cors.New(cors.Options{
			AllowedOrigins:   o,
			AllowCredentials: true,
			AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPut},
			AllowedHeaders:   []string{"authorization"},
			Debug:            log.Enabled(ctx, slog.LevelDebug),
			Logger:           slog.NewLogLogger(log.Handler(), slog.LevelDebug),
		}).Handler(mux).ServeHTTP(w, r)
	}

}

func Handler(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = r.URL.Query().Get("path")

	opts := &slog.HandlerOptions{}
	if os.Getenv("VERBOSE") == "true" {
		opts.Level = slog.LevelDebug
	}
	log := slog.New(slog.NewJSONHandler(os.Stderr, opts))

	if p == nil {
		p = persisters.NewPersister(slog.New(log.Handler().WithGroup("persister")), os.Getenv("POSTGRES_URL"))

		if err := p.Init(r.Context()); err != nil {
			panic(err)
		}
	}

	if a == nil {
		a = authn.NewAuthner(
			slog.New(log.Handler().WithGroup("authner")),

			os.Getenv("OIDC_ISSUER"),
			os.Getenv("OIDC_CLIENT_ID"),
			os.Getenv("OIDC_REDIRECT_URL"),
		)

		if err := a.Init(r.Context()); err != nil {
			panic(err)
		}
	}

	if s == nil {
		var err error
		s, err = api.GetSwagger()
		if err != nil {
			panic(err)
		}
	}

	if c == nil {
		var (
			contactName  = os.Getenv("CONTACT_NAME")
			contactURL   = os.Getenv("CONTACT_URL")
			contactEmail = os.Getenv("CONTACT_EMAIL")

			serverURL         = os.Getenv("SERVER_URL")
			serverDescription = os.Getenv("SERVER_DESCRIPTION")
		)
		if contactName == "" {
			contactName = "Felicitas Pojtinger"
		}
		if contactURL == "" {
			contactURL = "https://matrix.to/#/@pojntfx:matrix.org"
		}
		if contactEmail == "" {
			contactEmail = "felicitas@pojtinger.com"
		}

		if serverURL == "" {
			serverURL = "http://localhost:1337/"
		}
		if serverDescription == "" {
			serverDescription = "Local development server"
		}

		c = controllers.NewController(
			slog.New(log.Handler().WithGroup("controller")),

			p,
			a,

			s,

			os.Getenv("OIDC_ISSUER"),

			os.Getenv("PRIVACY_URL"),
			os.Getenv("IMPRINT_URL"),

			contactName,
			contactURL,
			contactEmail,

			serverURL,
			serverDescription,

			Code,
		)
	}

	o := []string{}
	if v := os.Getenv("CORS_ORIGINS"); v != "" {
		o = strings.Split(v, ",")
	}

	SenbaraRESTHandler(
		w,
		r,

		r.Context(),
		slog.New(log.Handler().WithGroup("handler")),
		o,
		c,
		s,
	)
}
