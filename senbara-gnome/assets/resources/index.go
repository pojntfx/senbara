package resources

import (
	_ "embed"
	"path"
)

const (
	AppID = "com.pojtinger.felicitas.Senbara"

	appPath = "/com/pojtinger/felicitas/Senbara/"

	SettingServerURLKey    = "server-url"
	SettingOIDCIssuerKey   = "oidc-issuer"
	SettingOIDCClientIDKey = "oidc-client-id"

	SecretRefreshTokenKey = "refresh-token"
	SecretIDTokenKey      = "id-token"
)

//go:generate sh -c "blueprint-compiler batch-compile . . *.blp && sass .:. && glib-compile-schemas . && glib-compile-resources *.gresource.xml"
//go:embed index.gresource
var ResourceContents []byte

var (
	ResourceWindowUIPath         = path.Join(appPath, "window.ui")
	ResourceIndexCSSPath         = path.Join(appPath, "index.css")
	ResourceGSchemasCompiledPath = path.Join(appPath, "gschemas.compiled")
)
