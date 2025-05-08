package resources

import (
	_ "embed"
	"path"
)

const (
	AppID = "com.pojtinger.felicitas.Senbara"

	appPath = "/com/pojtinger/felicitas/Senbara/"

	SettingServerURLKey    = "server-url"
	SettingOIDCClientIDKey = "oidc-client-id"
	SettingAnonymousMode   = "anonymous-mode"

	SecretRefreshTokenKey = "refresh-token"
	SecretIDTokenKey      = "id-token"

	SecretStateNonceKey       = "state-nonce"
	SecretPKCECodeVerifierKey = "pkce-code_verifier"
	SecretOIDCNonceKey        = "oidc-nonce"
)

//go:generate sh -c "blueprint-compiler batch-compile . . *.blp && sass .:. && glib-compile-schemas . && glib-compile-resources *.gresource.xml"
//go:embed index.gresource
var ResourceContents []byte

var (
	ResourceWindowUIPath         = path.Join(appPath, "window.ui")
	ResourceIndexCSSPath         = path.Join(appPath, "index.css")
	ResourceGSchemasCompiledPath = path.Join(appPath, "gschemas.compiled")
)
