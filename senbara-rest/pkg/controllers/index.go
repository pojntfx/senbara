package controllers

import (
	"context"
	"errors"
	"log/slog"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pojntfx/senbara/senbara-common/pkg/persisters"
	"golang.org/x/oauth2"
)

var (
	errCouldNotFetchFromDB    = errors.New("could not fetch from DB")
	errCouldNotLogin          = errors.New("could not login")
	errCouldNotEncodeResponse = errors.New("could not encode response")
	errCouldNotWriteResponse  = errors.New("could not write response")
)

type Controller struct {
	log       *slog.Logger
	persister *persisters.Persister
	spec      *openapi3.T

	oidcIssuer       string
	oidcClientID     string
	oidcRedirectURL  string
	oidcDiscoveryURL string

	privacyURL string
	imprintURL string

	code []byte

	config   *oauth2.Config
	verifier *oidc.IDTokenVerifier
}

func NewController(
	log *slog.Logger,

	persister *persisters.Persister,

	spec *openapi3.T,

	oidcIssuer,
	oidcClientID,
	oidcRedirectURL,
	oidcDiscoveryURL,

	privacyURL,
	imprintURL string,

	code []byte,
) *Controller {
	return &Controller{
		log: log,

		persister: persister,

		spec: spec,

		oidcIssuer:       oidcIssuer,
		oidcClientID:     oidcClientID,
		oidcRedirectURL:  oidcRedirectURL,
		oidcDiscoveryURL: oidcDiscoveryURL,

		privacyURL: privacyURL,
		imprintURL: imprintURL,

		code: code,
	}
}

func (c *Controller) Init(ctx context.Context) error {
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
