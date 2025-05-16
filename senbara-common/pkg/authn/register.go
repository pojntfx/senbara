package authn

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

type OIDCClientRegistrationResponse struct {
	ClientID                string `json:"client_id"`
	RegistrationAccessToken string `json:"registration_access_token"`
	RegistrationClientURI   string `json:"registration_client_uri"`
}

type oidcClientRegistrationRequest struct {
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	ClientName              string   `json:"client_name"`
	RedirectURIs            []string `json:"redirect_uris"`
	PostLogoutRedirectURIs  []string `json:"post_logout_redirect_uris"`
	GrantTypes              []string `json:"grant_types"`
	Scopes                  []string `json:"scopes"`
}

func RegisterOIDCClient(
	ctx context.Context,

	providerConfiguration *OIDCProviderConfiguration,

	clientName,
	redirectURL string,
) (*OIDCClientRegistrationResponse, error) {
	b, err := json.Marshal(oidcClientRegistrationRequest{
		TokenEndpointAuthMethod: "none",
		ClientName:              clientName,
		RedirectURIs:            []string{redirectURL},
		PostLogoutRedirectURIs:  []string{redirectURL},
		GrantTypes:              []string{"authorization_code", "implicit", "refresh_token"},
		Scopes:                  []string{"offline_access", "offline", "openid", "email", "email_verified"},
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, providerConfiguration.RegistrationEndpoint, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		return nil, errors.New(res.Status)
	}

	var r OIDCClientRegistrationResponse
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, err
	}

	return &r, nil
}
