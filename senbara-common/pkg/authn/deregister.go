package authn

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
)

func DeregisterOIDCClient(
	ctx context.Context,

	log *slog.Logger,

	registrationAccessToken,
	registrationClientURI string,
) error {
	l := log.With(
		"registrationAccessToken", registrationAccessToken != "",
		"registrationClientURI", registrationClientURI,
	)

	l.Debug("Starting OIDC client deregistration")

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, registrationClientURI, nil)
	if err != nil {
		l.Debug("Could not create OIDC client deregistration request", "error", err)

		return err
	}

	req.Header.Set("Authorization", "Bearer "+registrationAccessToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		l.Debug("Could not send OIDC client deregistration request", "error", err)

		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		l.Debug("OIDC client deregistration request returned an unexpected status", "statusCode", res.StatusCode)

		return errors.New(res.Status)
	}

	return nil
}
