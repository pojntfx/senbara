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
	senbaraREST "github.com/pojntfx/senbara/senbara-rest/api/senbara-rest"
	"github.com/pojntfx/senbara/senbara-rest/pkg/controllers"
)

var (
	errMissingOIDCIssuer      = errors.New("missing OIDC issuer")
	errMissingOIDCClientID    = errors.New("missing OIDC client ID")
	errMissingOIDCRedirectURL = errors.New("missing OIDC redirect URL")
)

func main() {
	laddr := flag.String("laddr", ":1337", "Listen address (port can also be set with `PORT` env variable)")
	pgaddr := flag.String("pgaddr", "postgresql://postgres@localhost:5432/senbara_rest?sslmode=disable", "Database address (can also be set using `POSTGRES_URL` env variable)")
	oidcIssuer := flag.String("oidc-issuer", "", "OIDC Issuer (i.e. https://pojntfx.eu.auth0.com/) (can also be set using the OIDC_ISSUER env variable)")
	oidcClientID := flag.String("oidc-client-id", "", "OIDC Client ID (i.e. myoidcclientid) (can also be set using the OIDC_CLIENT_ID env variable)")
	oidcRedirectURL := flag.String("oidc-redirect-url", "http://localhost:1337/authorize", "OIDC redirect URL (can also be set using the OIDC_REDIRECT_URL env variable)")

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

	if strings.TrimSpace(*oidcIssuer) == "" {
		panic(errMissingOIDCIssuer)
	}

	if strings.TrimSpace(*oidcClientID) == "" {
		panic(errMissingOIDCClientID)
	}

	if strings.TrimSpace(*oidcRedirectURL) == "" {
		panic(errMissingOIDCRedirectURL)
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
	)

	if err := c.Init(ctx); err != nil {
		panic(err)
	}

	log.Println("Listening on", *laddr)

	panic(http.ListenAndServe(*laddr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		senbaraREST.SenbaraRESTHandler(w, r, c)
	})))
}
