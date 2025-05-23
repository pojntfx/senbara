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

	errEmailNotVerified        = errors.New("email not verified")
	errCouldNotSetRefreshToken = errors.New("could not set refresh token")
	errCouldNotSetIDToken      = errors.New("could not set ID token")

	errCouldNotSetStateNonce       = errors.New("could not set state nonce")
	errCouldNotSetPKCECodeVerifier = errors.New("could not set PKCE code verifier")
	errCouldNotSetOIDCNonce        = errors.New("could not set OIDC nonce")

	errCouldNotClearRefreshToken = errors.New("could not clear refresh token")
	errCouldNotClearIDToken      = errors.New("could not clear ID token")

	errCouldNotClearStateNonce       = errors.New("could not clear state nonce")
	errCouldNotClearPKCECodeVerifier = errors.New("could not clear PKCE code verifier")
	errCouldNotClearOIDCNonce        = errors.New("could not clear OIDC nonce")

	errCouldNotGetAuthCodeURL = errors.New("could not get auth code URL")
)

type Authner struct {
	log *slog.Logger

	oidcIssuer,
	oidcEndSessionEndpoint,

	oidcClientID,
	oidcRedirectURL string

	config   *oauth2.Config
	verifier *oidc.IDTokenVerifier
}

func NewAuthner(
	log *slog.Logger,

	oidcIssuer,
	oidcEndSessionEndpoint,

	oidcClientID,
	oidcRedirectURL string,
) *Authner {
	return &Authner{
		log: log,

		oidcIssuer:             oidcIssuer,
		oidcEndSessionEndpoint: oidcEndSessionEndpoint,

		oidcClientID:    oidcClientID,
		oidcRedirectURL: oidcRedirectURL,
	}
}

func (a *Authner) Init(ctx context.Context) error {
	log := a.log.With("oidcIssuer", a.oidcIssuer)

	log.Info("Connecting to OIDC issuer")

	provider, err := oidc.NewProvider(ctx, a.oidcIssuer)
	if err != nil {
		log.Debug("Could not create OIDC provider", "error", err)

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
