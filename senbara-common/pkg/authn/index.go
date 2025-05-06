package authn

import (
	"context"
	"errors"
	"log/slog"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

var (
	ErrCouldNotLogin = errors.New("could not login")

	errEmailNotVerified          = errors.New("email not verified")
	errCouldNotSetRefreshToken   = errors.New("could not set refresh token")
	errCouldNotSetIDToken        = errors.New("could not set ID token")
	errCouldNotClearRefreshToken = errors.New("could not clear refresh token")
	errCouldNotClearIDToken      = errors.New("could not clear ID token")
)

type Authner struct {
	log *slog.Logger

	oidcIssuer,
	oidcClientID,
	oidcRedirectURL string

	config   *oauth2.Config
	verifier *oidc.IDTokenVerifier
}

func NewAuthner(
	log *slog.Logger,

	oidcIssuer,
	oidcClientID,
	oidcRedirectURL string,
) *Authner {
	return &Authner{
		log: log,

		oidcIssuer:      oidcIssuer,
		oidcClientID:    oidcClientID,
		oidcRedirectURL: oidcRedirectURL,
	}
}

func (a *Authner) Init(ctx context.Context) error {
	a.log.Info("Connecting to OIDC issuer", "oidcIssuer", a.oidcIssuer)

	provider, err := oidc.NewProvider(ctx, a.oidcIssuer)
	if err != nil {
		return err
	}

	a.config = &oauth2.Config{
		ClientID:    a.oidcClientID,
		RedirectURL: a.oidcRedirectURL,
		Endpoint:    provider.Endpoint(),
		Scopes:      []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess, "email", "email_verified"},
	}

	a.verifier = provider.Verifier(&oidc.Config{
		ClientID:          a.oidcClientID,
		SkipClientIDCheck: a.oidcClientID == "",
	})

	return nil
}
