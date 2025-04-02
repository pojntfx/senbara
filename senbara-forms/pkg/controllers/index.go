package controllers

import (
	"bytes"
	"context"
	"errors"
	"html/template"
	"log/slog"
	"math"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/pojntfx/senbara/senbara-common/pkg/persisters"
	"github.com/pojntfx/senbara/senbara-forms/web/templates"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"golang.org/x/oauth2"
)

var (
	errCouldNotRenderTemplate    = errors.New("could not render template")
	errCouldNotFetchFromDB       = errors.New("could not fetch from DB")
	errCouldNotParseForm         = errors.New("could not parse form")
	errInvalidForm               = errors.New("could not use invalid form")
	errCouldNotInsertIntoDB      = errors.New("could not insert into DB")
	errCouldNotDeleteFromDB      = errors.New("could not delete from DB")
	errCouldNotUpdateInDB        = errors.New("could not update in DB")
	errInvalidQueryParam         = errors.New("could not use invalid query parameter")
	errCouldNotLogin             = errors.New("could not login")
	errEmailNotVerified          = errors.New("email not verified")
	errCouldNotLocalize          = errors.New("could not localize")
	errCouldNotWriteResponse     = errors.New("could not write response")
	errCouldNotEncodeResponse    = errors.New("could not encode response")
	errCouldNotReadRequest       = errors.New("could not read request")
	errUnknownEntityName         = errors.New("unknown entity name")
	errCouldNotStartTransaction  = errors.New("could not start transaction")
	errCouldNotSetRefreshToken   = errors.New("could not set refresh token")
	errCouldNotSetIDToken        = errors.New("could not set ID token")
	errCouldNotClearRefreshToken = errors.New("could not clear refresh token")
	errCouldNotClearIDToken      = errors.New("could not clear ID token")
	errCouldNotExchange          = errors.New("could not exchange the OIDC auth code and state for refresh and ID token")
)

const (
	idTokenKey      = "id_token"
	refreshTokenKey = "refresh_token"
)

type Controller struct {
	log       *slog.Logger
	tpl       *template.Template
	persister *persisters.Persister

	oidcIssuer      string
	oidcClientID    string
	oidcRedirectURL string

	privacyURL string
	imprintURL string

	code []byte

	config   *oauth2.Config
	verifier *oidc.IDTokenVerifier
}

func NewController(
	log *slog.Logger,

	persister *persisters.Persister,

	oidcIssuer,
	oidcClientID,
	oidcRedirectURL,

	privacyURL,
	imprintURL string,

	code []byte,
) *Controller {
	return &Controller{
		log: log,

		persister: persister,

		oidcIssuer:      oidcIssuer,
		oidcClientID:    oidcClientID,
		oidcRedirectURL: oidcRedirectURL,

		privacyURL: privacyURL,
		imprintURL: imprintURL,

		code: code,
	}
}

func (c *Controller) Init(ctx context.Context) error {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
	)

	tpl, err := template.New("").Funcs(template.FuncMap{
		"TruncateText": func(text string, length int) string {
			if len(text) <= length {
				return text
			}

			return text[:length] + "…"
		},
		"RenderMarkdown": func(text string) template.HTML {
			var buf bytes.Buffer
			if err := md.Convert([]byte(text), &buf); err != nil {
				panic(err)
			}

			return template.HTML(buf.String())
		},
		"Abs": func(number float64) float64 {
			return math.Abs(number)
		},
	}).ParseFS(templates.FS, "*.html")
	if err != nil {
		return err
	}

	c.tpl = tpl

	c.log.Info("Connecting to OIDC issuer", "oidcIssuer", c.oidcIssuer)

	provider, err := oidc.NewProvider(ctx, c.oidcIssuer)
	if err != nil {
		return err
	}

	c.config = &oauth2.Config{
		ClientID:    c.oidcClientID,
		RedirectURL: c.oidcRedirectURL,
		Endpoint:    provider.Endpoint(),
		Scopes:      []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess, "email", "email_verified"},
	}

	c.verifier = provider.Verifier(&oidc.Config{
		ClientID: c.oidcClientID,
	})

	return nil
}
