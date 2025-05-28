package controllers

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pojntfx/senbara/senbara-common/pkg/authn"
	"github.com/pojntfx/senbara/senbara-common/pkg/persisters"
)

var (
	errCouldNotFetchFromDB      = errors.New("could not fetch from DB")
	errCouldNotEncodeResponse   = errors.New("could not encode response")
	errCouldNotReadRequest      = errors.New("could not read request")
	errCouldNotWriteResponse    = errors.New("could not write response")
	errCouldNotInsertIntoDB     = errors.New("could not insert into DB")
	errCouldNotDeleteFromDB     = errors.New("could not delete from DB")
	errCouldNotUpdateInDB       = errors.New("could not update in DB")
	errCouldNotStartTransaction = errors.New("could not start transaction")
	errUnknownEntityName        = errors.New("unknown entity name")
)

type Controller struct {
	log *slog.Logger

	persister *persisters.Persister
	authner   *authn.Authner

	spec *openapi3.T

	oidcDiscoveryURL                   string
	oidcDcrInitialAccessTokenPortalUrl string

	privacyURL string
	tosURL     string
	imprintURL string

	contactName  string
	contactEmail string

	serverURL         string
	serverDescription string

	code []byte
}

func NewController(
	log *slog.Logger,

	persister *persisters.Persister,
	authner *authn.Authner,

	spec *openapi3.T,

	oidcIssuer,
	oidcDcrInitialAccessTokenPortalUrl,

	privacyURL,
	tosURL,
	imprintURL,

	contactName,
	contactEmail,

	serverURL,
	serverDescription string,

	code []byte,
) *Controller {
	return &Controller{
		log: log,

		persister: persister,
		authner:   authner,

		spec: spec,

		oidcDiscoveryURL:                   strings.TrimSuffix(oidcIssuer, "/") + authn.OIDCWellKnownURLSuffix,
		oidcDcrInitialAccessTokenPortalUrl: oidcDcrInitialAccessTokenPortalUrl,

		privacyURL: privacyURL,
		tosURL:     tosURL,
		imprintURL: imprintURL,

		contactName:  contactName,
		contactEmail: contactEmail,

		serverURL:         serverURL,
		serverDescription: serverDescription,

		code: code,
	}
}
