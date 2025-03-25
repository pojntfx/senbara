package controllers

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pojntfx/senbara/senbara-common/pkg/persisters"
	"golang.org/x/oauth2"
)

var (
	errCouldNotFetchFromDB      = errors.New("could not fetch from DB")
	errCouldNotLogin            = errors.New("could not login")
	errCouldNotEncodeResponse   = errors.New("could not encode response")
	errCouldNotReadRequest      = errors.New("could not read request")
	errCouldNotWriteResponse    = errors.New("could not write response")
	errCouldNotInsertIntoDB     = errors.New("could not insert into DB")
	errCouldNotDeleteFromDB     = errors.New("could not delete from DB")
	errCouldNotUpdateInDB       = errors.New("could not update in DB")
	errCouldNotStartTransaction = errors.New("could not start transaction")
	errUnknownEntityName        = errors.New("unknown entity name")
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

	contactName  string
	contactURL   string
	contactEmail string

	serverURL         string
	serverDescription string

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

	privacyURL,
	imprintURL,

	contactName,
	contactURL,
	contactEmail,

	serverURL,
	serverDescription string,

	code []byte,
) *Controller {
	return &Controller{
		log: log,

		persister: persister,

		spec: spec,

		oidcIssuer:       oidcIssuer,
		oidcClientID:     oidcClientID,
		oidcRedirectURL:  oidcRedirectURL,
		oidcDiscoveryURL: strings.TrimSuffix(oidcIssuer, "/") + "/.well-known/openid-configuration",

		privacyURL: privacyURL,
		imprintURL: imprintURL,

		contactName:  contactName,
		contactURL:   contactURL,
		contactEmail: contactEmail,

		serverURL:         serverURL,
		serverDescription: serverDescription,

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
