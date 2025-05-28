package controllers

import (
	"bytes"
	"context"
	"errors"
	"html/template"
	"log/slog"
	"math"

	"github.com/pojntfx/senbara/senbara-common/pkg/authn"
	"github.com/pojntfx/senbara/senbara-common/pkg/persisters"
	"github.com/pojntfx/senbara/senbara-forms/web/templates"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

var (
	errCouldNotRenderTemplate   = errors.New("could not render template")
	errCouldNotFetchFromDB      = errors.New("could not fetch from DB")
	errCouldNotParseForm        = errors.New("could not parse form")
	errInvalidForm              = errors.New("could not use invalid form")
	errCouldNotInsertIntoDB     = errors.New("could not insert into DB")
	errCouldNotDeleteFromDB     = errors.New("could not delete from DB")
	errCouldNotUpdateInDB       = errors.New("could not update in DB")
	errInvalidQueryParam        = errors.New("could not use invalid query parameter")
	errCouldNotLogin            = errors.New("could not login")
	errCouldNotLocalize         = errors.New("could not localize")
	errCouldNotWriteResponse    = errors.New("could not write response")
	errCouldNotEncodeResponse   = errors.New("could not encode response")
	errCouldNotReadRequest      = errors.New("could not read request")
	errUnknownEntityName        = errors.New("unknown entity name")
	errCouldNotStartTransaction = errors.New("could not start transaction")
	errCouldNotExchange         = errors.New("could not exchange the OIDC auth code and state for refresh and ID token")
)

const (
	idTokenKey      = "id_token"
	refreshTokenKey = "refresh_token"

	stateNonceKey       = "state_nonce"
	pkceCodeVerifierKey = "pkce_code_verifier"
	oidcNonceKey        = "oidc_nonce"
)

type Controller struct {
	log *slog.Logger
	tpl *template.Template

	persister *persisters.Persister
	authner   *authn.Authner

	privacyURL string
	tosURL     string
	imprintURL string

	code []byte
}

func NewController(
	log *slog.Logger,

	persister *persisters.Persister,
	authner *authn.Authner,

	privacyURL,
	tosURL,
	imprintURL string,

	code []byte,
) *Controller {
	return &Controller{
		log: log,

		persister: persister,
		authner:   authner,

		privacyURL: privacyURL,
		tosURL:     tosURL,
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

	return nil
}
