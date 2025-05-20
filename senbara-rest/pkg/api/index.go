package api

const (
	OidcDcrInitialAccessTokenPortalUrlExtensionKey = `x-oidc-dcr-initial-access-token-portal-url`
)

//go:generate go tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=../../openapi-codegen.yaml ../../api/openapi/v1/openapi.yaml
