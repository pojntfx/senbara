package main

import (
	"context"
	"flag"
	"os"
	"strings"

	"github.com/pojntfx/senbara/senbara-common/pkg/authn"
	"gopkg.in/yaml.v3"
)

func main() {
	oidcIssuer := flag.String("oidc-issuer", "https://heuristic-rhodes-wqkaaxzmwj.projects.oryapis.com/", "OIDC issuer")
	oidcClientName := flag.String("oidc-client-name", "Senbara Forms Web (3rd-party test app)", "OIDC client name")
	oidcRedirectURL := flag.String("oidc-redirect-url", "http://localhost:1337/authorize", "OIDC redirect URL")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	o, err := authn.DiscoverOIDCProviderConfiguration(
		ctx,

		strings.TrimSuffix(*oidcIssuer, "/")+authn.OIDCWellKnownURLSuffix,
	)
	if err != nil {
		panic(err)
	}

	c, err := authn.RegisterOIDCClient(
		ctx,

		o,

		*oidcClientName,
		*oidcRedirectURL,
	)
	if err != nil {
		panic(err)
	}

	if err := yaml.NewEncoder(os.Stdout).Encode(c); err != nil {
		panic(err)
	}
}
