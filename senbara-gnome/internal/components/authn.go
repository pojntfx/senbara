package components

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
	"github.com/pojntfx/senbara/senbara-common/pkg/authn"
	"github.com/pojntfx/senbara/senbara-gnome/assets/resources"
	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
	"github.com/zalando/go-keyring"
)

type userData struct {
	Email     string
	LogoutURL string
}

type authorizer struct {
	log *slog.Logger

	authner  *authn.Authner
	settings *gio.Settings

	getVisiblePageTag func() string
	replaceWithTags   func(tags []string, position int)
	setUserData       func(u userData)
	getServerURL      func() string
}

func newAuthorizer(
	log *slog.Logger,

	authner *authn.Authner,
	settings *gio.Settings,

	getVisiblePageTag func() string,
	replaceWithTags func(tags []string, position int),
	setUserData func(u userData),
	getServerURL func() string,
) *authorizer {
	return &authorizer{
		log: log,

		authner:  authner,
		settings: settings,

		getVisiblePageTag: getVisiblePageTag,
		replaceWithTags:   replaceWithTags,
		setUserData:       setUserData,
		getServerURL:      getServerURL,
	}
}

func (a *authorizer) authorize(
	ctx context.Context,

	loginIfSignedOut bool,
) (
	redirected bool,

	client *api.ClientWithResponses,
	status int,

	err error,
) {
	path := a.getVisiblePageTag()

	log := a.log.With(
		"loginIfSignedOut", loginIfSignedOut,
		"path", path,
	)

	log.Debug("Handling user auth")

	var (
		refreshToken,
		idToken *string
	)
	rt, err := keyring.Get(resources.AppID, resources.SecretRefreshTokenKey)
	if err != nil {
		if !errors.Is(err, keyring.ErrNotFound) {
			log.Debug("Failed to read refresh token cookie", "error", err)

			return false, nil, http.StatusUnauthorized, errors.Join(errCouldNotLogin, err)
		}
	} else {
		refreshToken = &rt
	}

	it, err := keyring.Get(resources.AppID, resources.SecretIDTokenKey)
	if err != nil {
		if !errors.Is(err, keyring.ErrNotFound) {
			log.Debug("Failed to read ID token cookie", "error", err)

			return false, nil, http.StatusUnauthorized, errors.Join(errCouldNotLogin, err)
		}
	} else {
		idToken = &it
	}

	nextURL, email, logoutURL, err := a.authner.Authorize(
		ctx,

		loginIfSignedOut,

		path,
		path,

		refreshToken,
		idToken,

		func(s string, t time.Time) error {
			// TODO: Handle expiry time
			return keyring.Set(resources.AppID, resources.SecretRefreshTokenKey, s)
		},
		func(s string, t time.Time) error {
			// TODO: Handle expiry time
			return keyring.Set(resources.AppID, resources.SecretIDTokenKey, s)
		},

		func(s string) error {
			return keyring.Set(resources.AppID, resources.SecretStateNonceKey, s)
		},
		func(s string) error {
			return keyring.Set(resources.AppID, resources.SecretPKCECodeVerifierKey, s)
		},
		func(s string) error {
			return keyring.Set(resources.AppID, resources.SecretOIDCNonceKey, s)
		},
	)
	if err != nil {
		if errors.Is(err, authn.ErrCouldNotLogin) {
			return false, nil, http.StatusUnauthorized, err
		}

		return false, nil, http.StatusInternalServerError, err
	}

	redirected = nextURL != ""
	u := userData{
		Email:     email,
		LogoutURL: logoutURL,
	}
	a.setUserData(u)

	if redirected {
		a.replaceWithTags([]string{resources.PageExchangeLogin}, 1)

		if _, err := gio.AppInfoLaunchDefaultForUri(nextURL, nil); err != nil {
			log.Debug("Could not open nextURL", "error", err)

			return false, nil, http.StatusUnauthorized, errors.Join(errCouldNotLogin, err)
		}

		return redirected, nil, http.StatusTemporaryRedirect, nil
	}

	opts := []api.ClientOption{}
	if strings.TrimSpace(u.Email) != "" {
		log.Debug("Creating authenticated client")

		it, err = keyring.Get(resources.AppID, resources.SecretIDTokenKey)
		if err != nil {
			log.Debug("Failed to read ID token cookie", "error", err)

			return false, nil, http.StatusUnauthorized, errors.Join(errCouldNotLogin, err)
		}

		sp, err := securityprovider.NewSecurityProviderBearerToken(it)
		if err != nil {
			log.Debug("Could not create bearer token security provider", "error", err)

			return false, nil, http.StatusUnauthorized, errors.Join(errCouldNotLogin, err)
		}

		opts = append(opts, api.WithRequestEditorFn(sp.Intercept))
	} else {
		log.Debug("Creating unauthenticated client")
	}

	client, err = api.NewClientWithResponses(
		a.getServerURL(),
		opts...,
	)
	if err != nil {
		log.Debug("Could not create authenticated API client", "error", err)

		return false, nil, http.StatusUnauthorized, errors.Join(errCouldNotLogin, err)
	}

	return redirected, client, http.StatusOK, nil
}
