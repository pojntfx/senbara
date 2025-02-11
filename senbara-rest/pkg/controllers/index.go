package controllers

import (
	"context"
	"errors"
	"log/slog"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/pojntfx/senbara/senbara-common/pkg/persisters"
	"golang.org/x/oauth2"
)

var (
	errCouldNotFetchFromDB = errors.New("could not fetch from DB")
	errCouldNotLogin       = errors.New("could not login")
	errEmailNotVerified    = errors.New("email not verified")
)

type Controller struct {
	log       *slog.Logger
	persister *persisters.Persister

	oidcIssuer      string
	oidcClientID    string
	oidcRedirectURL string

	privacyURL string
	imprintURL string

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
) *Controller {
	return &Controller{
		log: log,

		persister: persister,

		oidcIssuer:      oidcIssuer,
		oidcClientID:    oidcClientID,
		oidcRedirectURL: oidcRedirectURL,

		privacyURL: privacyURL,
		imprintURL: imprintURL,
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
