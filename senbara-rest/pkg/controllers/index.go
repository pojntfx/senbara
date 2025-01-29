package controllers

import (
	"context"
	"errors"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/pojntfx/senbara/senbara-common/pkg/persisters"
	"golang.org/x/oauth2"
)

var (
	errCouldNotFetchFromDB      = errors.New("could not fetch from DB")
	errCouldNotParseForm        = errors.New("could not parse form")
	errInvalidForm              = errors.New("could not use invalid form")
	errCouldNotInsertIntoDB     = errors.New("could not insert into DB")
	errCouldNotDeleteFromDB     = errors.New("could not delete from DB")
	errCouldNotUpdateInDB       = errors.New("could not update in DB")
	errInvalidQueryParam        = errors.New("could not use invalid query parameter")
	errCouldNotLogin            = errors.New("could not login")
	errEmailNotVerified         = errors.New("email not verified")
	errCouldNotWriteResponse    = errors.New("could not write response")
	errCouldNotReadRequest      = errors.New("could not read request")
	errUnknownEntityName        = errors.New("unknown entity name")
	errCouldNotStartTransaction = errors.New("could not start transaction")
)

const (
	idTokenKey = "id_token"
)

type Controller struct {
	persister *persisters.Persister

	oidcIssuer      string
	oidcClientID    string
	oidcRedirectURL string

	config   *oauth2.Config
	verifier *oidc.IDTokenVerifier
}

func NewController(
	persister *persisters.Persister,

	oidcIssuer,
	oidcClientID,
	oidcRedirectURL string,
) *Controller {
	return &Controller{
		persister: persister,

		oidcIssuer:      oidcIssuer,
		oidcClientID:    oidcClientID,
		oidcRedirectURL: oidcRedirectURL,
	}
}

func (b *Controller) Init(ctx context.Context) error {
	provider, err := oidc.NewProvider(ctx, b.oidcIssuer)
	if err != nil {
		return err
	}

	b.config = &oauth2.Config{
		ClientID:    b.oidcClientID,
		RedirectURL: b.oidcRedirectURL,
		Endpoint:    provider.Endpoint(),
		Scopes:      []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess, "email", "email_verified"},
	}

	b.verifier = provider.Verifier(&oidc.Config{
		ClientID: b.oidcClientID,
	})

	return nil
}
