package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/pojntfx/senbara/senbara-common/pkg/persisters"
	senbaraForms "github.com/pojntfx/senbara/senbara-forms/api/senbara-forms"
	"github.com/pojntfx/senbara/senbara-forms/pkg/controllers"
)

var (
	errMissingOIDCIssuer      = errors.New("missing OIDC issuer")
	errMissingOIDCClientID    = errors.New("missing OIDC client ID")
	errMissingOIDCRedirectURL = errors.New("missing OIDC redirect URL")
	errMissingPrivacyURL      = errors.New("missing privacy policy URL")
	errMissingImprintURL      = errors.New("missing imprint URL")
)

func main() {
	laddr := flag.String("laddr", ":1337", "Listen address (port can also be set with `PORT` env variable)")
	pgaddr := flag.String("pgaddr", "postgresql://postgres@localhost:5432/senbara_forms?sslmode=disable", "Database address (can also be set using `POSTGRES_URL` env variable)")
	oidcIssuer := flag.String("oidc-issuer", "", "OIDC Issuer (i.e. https://pojntfx.eu.auth0.com/) (can also be set using the OIDC_ISSUER env variable)")
	oidcClientID := flag.String("oidc-client-id", "", "OIDC Client ID (i.e. myoidcclientid) (can also be set using the OIDC_CLIENT_ID env variable)")
	oidcRedirectURL := flag.String("oidc-redirect-url", "http://localhost:1337/authorize", "OIDC redirect URL (can also be set using the OIDC_REDIRECT_URL env variable)")
	privacyURL := flag.String("privacy-url", "", "Privacy policy URL (can also be set using the PRIVACY_URL env variable)")
	imprintURL := flag.String("imprint-url", "", "Imprint URL (can also be set using the IMPRINT_URL env variable)")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if v := os.Getenv("PORT"); v != "" {
		log.Println("Using port from PORT env variable")

		la, err := net.ResolveTCPAddr("tcp", *laddr)
		if err != nil {
			panic(err)
		}

		p, err := strconv.Atoi(v)
		if err != nil {
			panic(err)
		}

		la.Port = p
		*laddr = la.String()
	}

	if v := os.Getenv("POSTGRES_URL"); v != "" {
		log.Println("Using database address from POSTGRES_URL env variable")

		*pgaddr = v
	}

	if v := os.Getenv("OIDC_ISSUER"); v != "" {
		log.Println("Using OIDC issuer from OIDC_ISSUER env variable")

		*oidcIssuer = v
	}

	if v := os.Getenv("OIDC_CLIENT_ID"); v != "" {
		log.Println("Using OIDC client ID from OIDC_CLIENT_ID env variable")

		*oidcClientID = v
	}

	if v := os.Getenv("OIDC_REDIRECT_URL"); v != "" {
		log.Println("Using OIDC redirect URL from OIDC_REDIRECT_URL env variable")

		*oidcRedirectURL = v
	}

	if v := os.Getenv("PRIVACY_URL"); v != "" {
		log.Println("Using privacy policy URL from PRIVACY_URL env variable")

		*privacyURL = v
	}

	if v := os.Getenv("IMPRINT_URL"); v != "" {
		log.Println("Using imprint URL from IMPRINT_URL env variable")

		*imprintURL = v
	}

	if strings.TrimSpace(*oidcIssuer) == "" {
		panic(errMissingOIDCIssuer)
	}

	if strings.TrimSpace(*oidcClientID) == "" {
		panic(errMissingOIDCClientID)
	}

	if strings.TrimSpace(*oidcRedirectURL) == "" {
		panic(errMissingOIDCRedirectURL)
	}

	if strings.TrimSpace(*privacyURL) == "" {
		panic(errMissingPrivacyURL)
	}

	if strings.TrimSpace(*imprintURL) == "" {
		panic(errMissingImprintURL)
	}

	p := persisters.NewPersister(*pgaddr)

	if err := p.Init(); err != nil {
		panic(err)
	}

	c := controllers.NewController(
		p,

		*oidcIssuer,
		*oidcClientID,
		*oidcRedirectURL,

		*privacyURL,
		*imprintURL,
	)

	if err := c.Init(ctx); err != nil {
		panic(err)
	}

	log.Println("Listening on", *laddr)

	panic(http.ListenAndServe(*laddr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		senbaraForms.SenbaraFormsHandler(w, r, c)
	})))
}
