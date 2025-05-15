package authn

import (
	"context"
	"errors"
	"net/http"
)

func DeregisterOIDCClient(
	ctx context.Context,

	registrationAccessToken,
	registrationClientURI string,
) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, registrationClientURI, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+registrationAccessToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		return errors.New(res.Status)
	}

	return nil
}
