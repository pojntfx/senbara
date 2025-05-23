package authn

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"path"
)

var (
	// See https://openid.net/specs/openid-connect-discovery-1_0-errata2.html#ProviderConfig
	OIDCWellKnownURLSuffix = path.Join("/", ".well-known", "openid-configuration")
)

type OIDCProviderConfiguration struct {
	Issuer               string `json:"issuer"`
	EndSessionEndpoint   string `json:"end_session_endpoint"`
	RegistrationEndpoint string `json:"registration_endpoint"`
}

func DiscoverOIDCProviderConfiguration(
	ctx context.Context,

	log *slog.Logger,

	wellKnownURL string,
) (*OIDCProviderConfiguration, error) {
	l := log.With("wellKnownURL", wellKnownURL)

	l.Debug("Starting OIDC provider configuration discovery")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, wellKnownURL, nil)
	if err != nil {
		l.Debug("Could not create OIDC provider configuration discovery request", "error", err)

		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		l.Debug("Could not send OIDC provider configuration discovery request", "error", err)

		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		l.Debug("OIDC provider configuration discovery request returned an unexpected status", "statusCode", res.StatusCode)

		return nil, errors.New(res.Status)
	}

	var p OIDCProviderConfiguration
	if err := json.NewDecoder(res.Body).Decode(&p); err != nil {
		l.Debug("Could not decode discovered OIDC provider configuration", "error", err)

		return nil, err
	}

	return &p, nil
}
