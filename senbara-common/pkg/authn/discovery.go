package authn

import (
	"context"
	"encoding/json"
	"errors"
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

	wellKnownURL string,
) (*OIDCProviderConfiguration, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, wellKnownURL, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, errors.New(res.Status)
	}

	var p OIDCProviderConfiguration
	if err := json.NewDecoder(res.Body).Decode(&p); err != nil {
		return nil, err
	}

	return &p, nil
}
