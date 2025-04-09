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

	oidcDiscoveryURL string

	privacyURL string
	imprintURL string

	contactName  string
	contactURL   string
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

	privacyURL,
	imprintURL,

	contactName,
	contactURL,
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

		oidcDiscoveryURL: strings.TrimSuffix(oidcIssuer, "/") + spec.Components.SecuritySchemes["oidc"].Value.OpenIdConnectUrl,

		privacyURL: privacyURL,
		imprintURL: imprintURL,

		contactName:  contactName,
		contactURL:   contactURL,
		contactEmail: contactEmail,

		serverURL:         serverURL,
		serverDescription: serverDescription,

		code: code,
	}
}
