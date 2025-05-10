package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"net/http"
	"os"
	"strings"

	"github.com/pojntfx/senbara/senbara-common/pkg/authn"
	"gopkg.in/yaml.v3"
)

type oidcConfig struct {
	RegistrationEndpoint string `json:"registration_endpoint"`
}

type oidcClientRegistrationRequest struct {
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	ClientName              string   `json:"client_name"`
	RedirectURIs            []string `json:"redirect_uris"`
	PostLogoutRedirectURIs  []string `json:"post_logout_redirect_uris"`
	GrantTypes              []string `json:"grant_types"`
	Scopes                  []string `json:"scopes"`
}

type oidcClientRegistrationResponse struct {
	ClientID                string `json:"client_id"`
	RegistrationAccessToken string `json:"registration_access_token"`
	RegistrationClientURI   string `json:"registration_client_uri"`
}

func main() {
	oidcIssuer := flag.String("oidc-issuer", "https://heuristic-rhodes-wqkaaxzmwj.projects.oryapis.com/", "OIDC issuer")
	oidcClientName := flag.String("oidc-client-name", "Senbara Forms Web (3rd-party test app)", "OIDC client name")
	oidcRedirectURL := flag.String("oidc-redirect-url", "http://localhost:1337/authorize", "OIDC redirect URL")

	flag.Parse()

	var p oidcConfig
	{
		res, err := http.Get(strings.TrimSuffix(*oidcIssuer, "/") + authn.OIDCDiscoverySuffix)
		if err != nil {
			panic(err)
		}

		if res.StatusCode != http.StatusOK {
			panic(errors.New(res.Status))
		}

		if err := json.NewDecoder(res.Body).Decode(&p); err != nil {
			panic(err)
		}
	}

	var r oidcClientRegistrationResponse
	{
		b, err := json.Marshal(oidcClientRegistrationRequest{
			TokenEndpointAuthMethod: "none",
			ClientName:              *oidcClientName,
			RedirectURIs:            []string{*oidcRedirectURL},
			PostLogoutRedirectURIs:  []string{*oidcRedirectURL},
			GrantTypes:              []string{"authorization_code", "implicit", "refresh_token"},
			Scopes:                  []string{"offline_access", "offline", "openid", "email", "email_verified"},
		})
		if err != nil {
			panic(err)
		}

		res, err := http.Post(p.RegistrationEndpoint, "application/json", bytes.NewBuffer(b))
		if err != nil {
			panic(err)
		}

		if res.StatusCode != http.StatusCreated {
			panic(errors.New(res.Status))
		}

		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			panic(err)
		}
	}

	if err := yaml.NewEncoder(os.Stdout).Encode(r); err != nil {
		panic(err)
	}
}
