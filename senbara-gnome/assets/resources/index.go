package resources

import (
	_ "embed"
	"path"
)

const (
	AppID      = "com.pojtinger.felicitas.Senbara"
	AppVersion = "0.1.0"

	appPath = "/com/pojtinger/felicitas/Senbara/"

	SettingServerURLKey             = "server-url"
	SettingRegistrationClientURIKey = "registration-client-uri"
	SettingOIDCClientIDKey          = "oidc-client-id"
	SettingAnonymousMode            = "anonymous-mode"

	SecretRegistrationAccessToken = "registration-access-token"

	SecretRefreshTokenKey = "refresh-token"
	SecretIDTokenKey      = "id-token"

	SecretStateNonceKey       = "state-nonce"
	SecretPKCECodeVerifierKey = "pkce-code_verifier"
	SecretOIDCNonceKey        = "oidc-nonce"

	PageIndex = "/"

	PageWelcome  = "/welcome"
	PagePreview  = "/preview"
	PageRegister = "/register"
	PageHome     = "/home"

	PageConfigServerURL          = "/config/server-url"
	PageConfigInitialAccessToken = "/config/initial-access-token"

	PageExchangeLogin  = "/exchange/login"
	PageExchangeLogout = "/exchange/logout"

	PageContacts = "/contacts"
	PageJournal  = "/journal"

	PageContactsLoading   = "/contacts/loading"
	PageContactsList      = "/contacts/list"
	PageContactsNoResults = "/contacts/no-results"
	PageContactsEmpty     = "/contacts/empty"
	PageContactsError     = "/contacts/error"

	PageContactsView        = "/contacts/view"
	PageContactsViewLoading = "/contacts/view/loading"
	PageContactsViewData    = "/contacts/view/data"
	PageContactsViewError   = "/contacts/view/error"
)

//go:generate sh -c "blueprint-compiler batch-compile . . *.blp && sass .:. && glib-compile-schemas . && glib-compile-resources *.gresource.xml"
//go:embed index.gresource
var ResourceContents []byte

var (
	ResourceWindowUIPath               = path.Join(appPath, "window.ui")
	ResourceContactsCreateDialogUIPath = path.Join(appPath, "contacts-create-dialog.ui")
	ResourceIndexCSSPath               = path.Join(appPath, "index.css")
	ResourceGSchemasCompiledPath       = path.Join(appPath, "gschemas.compiled")
	ResourceMetainfoPath               = path.Join(appPath, "metainfo.xml")
)
