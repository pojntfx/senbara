package resources

import (
	_ "embed"
	"path"
	"strings"
)

//go:generate sh -c "find ../../po -name '*.po' | sed 's|^\\../../po/||; s|\\.po$||' > ../../po/LINGUAS"
//go:generate sh -c "msgfmt --desktop --template ../../assets/meta/com.pojtinger.felicitas.Senbara.desktop.in -d ../../po -o - -f | sed 's|/LC_MESSAGES/default||g' > ../../assets/meta/com.pojtinger.felicitas.Senbara.desktop"
//go:generate sh -c "msgfmt --xml -L metainfo --template ../../assets/resources/metainfo.xml.in -d ../../po -o - -f | sed 's|/LC[-_]MESSAGES/default||g' > ../../assets/resources/metainfo.xml"

const (
	AppID      = "com.pojtinger.felicitas.Senbara"
	AppVersion = "0.1.0"

	SettingVerboseKey               = "verbose"
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

	PageContacts       = "/contacts"
	PageJournalEntries = "/journal"

	PageContactsLoading   = "/contacts/loading"
	PageContactsList      = "/contacts/list"
	PageContactsNoResults = "/contacts/no-results"
	PageContactsEmpty     = "/contacts/empty"
	PageContactsError     = "/contacts/error"

	PageContactsView        = "/contacts/view"
	PageContactsViewLoading = "/contacts/view/loading"
	PageContactsViewData    = "/contacts/view/data"
	PageContactsViewError   = "/contacts/view/error"

	PageActivitiesView        = "/activities/view"
	PageActivitiesViewLoading = "/activities/view/loading"
	PageActivitiesViewData    = "/activities/view/data"
	PageActivitiesViewError   = "/activities/view/error"

	PageActivitiesEdit        = "/activities/edit"
	PageActivitiesEditLoading = "/activities/edit/loading"
	PageActivitiesEditData    = "/activities/edit/data"
	PageActivitiesEditError   = "/activities/edit/error"

	PageDebtsEdit        = "/debts/edit"
	PageDebtsEditLoading = "/debts/edit/loading"
	PageDebtsEditData    = "/debts/edit/data"
	PageDebtsEditError   = "/debts/edit/error"

	PageContactsEdit        = "/contacts/edit"
	PageContactsEditLoading = "/contacts/edit/loading"
	PageContactsEditData    = "/contacts/edit/data"
	PageContactsEditError   = "/contacts/edit/error"

	PageJournalEntriesLoading   = "/journal/loading"
	PageJournalEntriesList      = "/journal/list"
	PageJournalEntriesNoResults = "/journal/no-results"
	PageJournalEntriesEmpty     = "/journal/empty"
	PageJournalEntriesError     = "/journal/error"

	PageJournalEntriesView        = "/journal/view"
	PageJournalEntriesViewLoading = "/journal/view/loading"
	PageJournalEntriesViewData    = "/journal/view/data"
	PageJournalEntriesViewError   = "/journal/view/error"

	PageJournalEntriesEdit        = "/journal/edit"
	PageJournalEntriesEditLoading = "/journal/edit/loading"
	PageJournalEntriesEditData    = "/journal/edit/data"
	PageJournalEntriesEditError   = "/journal/edit/error"
)

//go:generate sh -c "blueprint-compiler batch-compile . . *.blp && sass .:. && glib-compile-schemas . && glib-compile-resources *.gresource.xml"
//go:embed index.gresource
var ResourceContents []byte

var (
	AppPath = path.Join("/com", "pojtinger", "felicitas", "Senbara")

	AppDevelopers = []string{"Felicitas Pojtinger"}
	AppArtists    = AppDevelopers
	AppCopyright  = "© 2025 " + strings.Join(AppDevelopers, ", ")

	ResourceWindowUIPath               = path.Join(AppPath, "window.ui")
	ResourcePreferencesDialogUIPath    = path.Join(AppPath, "preferences-dialog.ui")
	ResourceHelpOverviewUIPath         = path.Join(AppPath, "help-overlay.ui")
	ResourceContactsCreateDialogUIPath = path.Join(AppPath, "contacts-create-dialog.ui")
	ResourceDebtsCreateDialogUIPath    = path.Join(AppPath, "debts-create-dialog.ui")
	ActivitiesDebtsCreateDialogUIPath  = path.Join(AppPath, "activities-create-dialog.ui")
	JournalEntriesCreateDialogUIPath   = path.Join(AppPath, "journal-entries-create-dialog.ui")
	ResourceGSchemasCompiledPath       = path.Join(AppPath, "gschemas.compiled")
	ResourceMetainfoPath               = path.Join(AppPath, "metainfo.xml")
)
