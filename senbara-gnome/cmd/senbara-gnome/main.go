package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"math"
	"mime/multipart"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4-webkitgtk/pkg/webkit/v6"
	gcore "github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
	"github.com/oapi-codegen/runtime/types"
	"github.com/pojntfx/senbara/senbara-common/pkg/authn"
	"github.com/pojntfx/senbara/senbara-gnome/assets/resources"
	"github.com/pojntfx/senbara/senbara-gnome/config/locales"
	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/zalando/go-keyring"
)

var (
	errCouldNotLogin            = errors.New("could not login")
	errCouldNotWriteSettingsKey = errors.New("could not write settings key")
	errMissingPrivacyURL        = errors.New("missing privacy policy URL")
	errMissingContactID         = errors.New("missing contact ID")
	errInvalidContactID         = errors.New("invalid contact ID")
	errMissingActivityID        = errors.New("missing activity ID")
	errInvalidActivityID        = errors.New("invalid activity ID")
)

const (
	redirectURL = "senbara:///authorize"
)

type userData struct {
	Email     string
	LogoutURL string
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := &slog.HandlerOptions{}
	// TODO: Get verbosity level from GSettings
	if true {
		opts.Level = slog.LevelDebug
	}
	log := slog.New(slog.NewJSONHandler(os.Stderr, opts))

	i18t, err := os.MkdirTemp("", "")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(i18t)

	if err := fs.WalkDir(locales.FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if err := os.MkdirAll(filepath.Join(i18t, path), os.ModePerm); err != nil {
				return err
			}

			return nil
		}

		src, err := locales.FS.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		dst, err := os.Create(filepath.Join(i18t, path))
		if err != nil {
			return err
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return err
		}

		return nil
	}); err != nil {
		panic(err)
	}

	gcore.InitI18n("default", i18t)

	r, err := gio.NewResourceFromData(glib.NewBytesWithGo(resources.ResourceContents))
	if err != nil {
		panic(err)
	}
	gio.ResourcesRegister(r)

	st, err := os.MkdirTemp("", "")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(st)

	sc, err := r.LookupData(resources.ResourceGSchemasCompiledPath, gio.ResourceLookupFlagsNone)
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile(filepath.Join(st, path.Base(resources.ResourceGSchemasCompiledPath)), sc.Data(), os.ModePerm); err != nil {
		panic(err)
	}

	if err := os.Setenv("GSETTINGS_SCHEMA_DIR", st); err != nil {
		panic(err)
	}

	settings := gio.NewSettings(resources.AppID)

	c := gtk.NewCSSProvider()
	c.LoadFromResource(resources.ResourceIndexCSSPath)

	a := adw.NewApplication(resources.AppID, gio.ApplicationHandlesOpen)

	var (
		w       *adw.Window
		nv      *adw.NavigationView
		mto     *adw.ToastOverlay
		authner *authn.Authner
		u       userData
	)

	authorize := func(
		ctx context.Context,

		loginIfSignedOut bool,
	) (
		redirected bool,

		client *api.ClientWithResponses,
		status int,

		err error,
	) {
		log := log.With(
			"loginIfSignedOut", loginIfSignedOut,
			"path", nv.VisiblePage().Tag(),
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

		nextURL, email, logoutURL, err := authner.Authorize(
			ctx,

			loginIfSignedOut,

			nv.VisiblePage().Tag(),
			nv.VisiblePage().Tag(),

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
		u = userData{
			Email:     email,
			LogoutURL: logoutURL,
		}

		if redirected {
			nv.ReplaceWithTags([]string{resources.PageExchangeLogin})

			var (
				fl = gtk.NewURILauncher(nextURL)
				cc = make(chan error)
			)
			fl.Launch(ctx, &w.Window, func(res gio.AsyncResulter) {
				if err := fl.LaunchFinish(res); err != nil {
					cc <- err

					return
				}

				cc <- nil
			})

			if err := <-cc; err != nil {
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

			a, err := securityprovider.NewSecurityProviderBearerToken(it)
			if err != nil {
				log.Debug("Could not create bearer token security provider", "error", err)

				return false, nil, http.StatusUnauthorized, errors.Join(errCouldNotLogin, err)
			}

			opts = append(opts, api.WithRequestEditorFn(a.Intercept))
		} else {
			log.Debug("Creating unauthenticated client")
		}

		client, err = api.NewClientWithResponses(
			settings.String(resources.SettingServerURLKey),
			opts...,
		)
		if err != nil {
			log.Debug("Could not create authenticated API client", "error", err)

			return false, nil, http.StatusUnauthorized, errors.Join(errCouldNotLogin, err)
		}

		return redirected, client, http.StatusOK, nil
	}

	var rawError string
	handlePanic := func(err error) {
		rawError = err.Error()
		i18nErr := gcore.Local(rawError)

		log.Error(
			"An unexpected error occured, showing error message to user",
			"rawError", rawError,
			"i18nErr", i18nErr,
		)

		toast := adw.NewToast(i18nErr)
		toast.SetButtonLabel(gcore.Local("Copy details"))
		toast.SetActionName("app.copyErrorToClipboard")

		mto.AddToast(toast)
	}

	a.ConnectActivate(func() {
		gtk.StyleContextAddProviderForDisplay(
			gdk.DisplayGetDefault(),
			c,
			gtk.STYLE_PROVIDER_PRIORITY_APPLICATION,
		)

		aboutDialog := adw.NewAboutDialogFromAppdata(resources.ResourceMetainfoPath, resources.AppVersion)
		aboutDialog.SetDevelopers([]string{"Felicitas Pojtinger"})
		aboutDialog.SetArtists([]string{"Felicitas Pojtinger"})
		aboutDialog.SetCopyright("© 2025 Felicitas Pojtinger")

		b := gtk.NewBuilderFromResource(resources.ResourceWindowUIPath)

		contactsCreateDialogBuilder := gtk.NewBuilderFromResource(resources.ResourceContactsCreateDialogUIPath)
		debtsCreateDialogBuilder := gtk.NewBuilderFromResource(resources.ResourceDebtsCreateDialogUIPath)

		w = b.GetObject("main-window").Cast().(*adw.Window)

		nv = b.GetObject("main-navigation").Cast().(*adw.NavigationView)

		mto = b.GetObject("main-toasts-overlay").Cast().(*adw.ToastOverlay)

		var (
			welcomeGetStartedButton  = b.GetObject("welcome-get-started-button").Cast().(*gtk.Button)
			welcomeGetStartedSpinner = b.GetObject("welcome-get-started-spinner").Cast().(*adw.Spinner)

			configServerURLInput           = b.GetObject("config-server-url-input").Cast().(*adw.EntryRow)
			configServerURLContinueButton  = b.GetObject("config-server-url-continue-button").Cast().(*gtk.Button)
			configServerURLContinueSpinner = b.GetObject("config-server-url-continue-spinner").Cast().(*adw.Spinner)

			spec openapi3.T

			previewLoginButton  = b.GetObject("preview-login-button").Cast().(*gtk.Button)
			previewLoginSpinner = b.GetObject("preview-login-spinner").Cast().(*adw.Spinner)

			oidcDcrInitialAccessTokenPortalUrl string

			registerRegisterButton = b.GetObject("register-register-button").Cast().(*gtk.Button)

			configInitialAccessTokenInput        = b.GetObject("config-initial-access-token-input").Cast().(*adw.PasswordEntryRow)
			configInitialAccessTokenLoginButton  = b.GetObject("config-initial-access-token-login-button").Cast().(*gtk.Button)
			configInitialAccessTokenLoginSpinner = b.GetObject("config-initial-access-token-login-spinner").Cast().(*adw.Spinner)

			exchangeLoginCancelButton  = b.GetObject("exchange-login-cancel-button").Cast().(*gtk.Button)
			exchangeLogoutCancelButton = b.GetObject("exchange-logout-cancel-button").Cast().(*gtk.Button)

			homeSplitView      = b.GetObject("home-split-view").Cast().(*adw.NavigationSplitView)
			homeNavigation     = b.GetObject("home-navigation").Cast().(*adw.NavigationView)
			homeSidebarListbox = b.GetObject("home-sidebar-listbox").Cast().(*gtk.ListBox)
			homeContentPage    = b.GetObject("home-content-page").Cast().(*adw.NavigationPage)

			homeUserMenuButton  = b.GetObject("home-user-menu-button").Cast().(*gtk.MenuButton)
			homeUserMenuAvatar  = b.GetObject("home-user-menu-avatar").Cast().(*adw.Avatar)
			homeUserMenuSpinner = b.GetObject("home-user-menu-spinner").Cast().(*adw.Spinner)

			homeHamburgerMenuButton  = b.GetObject("home-hamburger-menu-button").Cast().(*gtk.MenuButton)
			homeHamburgerMenuIcon    = b.GetObject("home-hamburger-menu-icon").Cast().(*gtk.Image)
			homeHamburgerMenuSpinner = b.GetObject("home-hamburger-menu-spinner").Cast().(*adw.Spinner)

			homeSidebarContactsCountLabel   = b.GetObject("home-sidebar-contacts-count-label").Cast().(*gtk.Label)
			homeSidebarContactsCountSpinner = b.GetObject("home-sidebar-contacts-count-spinner").Cast().(*adw.Spinner)

			homeSidebarJournalEntriesCountLabel   = b.GetObject("home-sidebar-journal-entries-count-label").Cast().(*gtk.Label)
			homeSidebarJournalEntriesCountSpinner = b.GetObject("home-sidebar-journal-entries-count-spinner").Cast().(*adw.Spinner)

			contactsStack       = b.GetObject("contacts-stack").Cast().(*gtk.Stack)
			contactsListBox     = b.GetObject("contacts-list").Cast().(*gtk.ListBox)
			contactsSearchEntry = b.GetObject("contacts-searchentry").Cast().(*gtk.SearchEntry)

			contactsAddButton    = b.GetObject("contacts-add-button").Cast().(*gtk.Button)
			contactsSearchButton = b.GetObject("contacts-search-button").Cast().(*gtk.ToggleButton)

			contactsEmptyAddButton = b.GetObject("contacts-empty-add-button").Cast().(*gtk.Button)

			contactsCreateDialog = contactsCreateDialogBuilder.GetObject("contacts-create-dialog").Cast().(*adw.Dialog)

			contactsCreateDialogAddButton  = contactsCreateDialogBuilder.GetObject("contacts-create-dialog-add-button").Cast().(*gtk.Button)
			contactsCreateDialogAddSpinner = contactsCreateDialogBuilder.GetObject("contacts-create-dialog-add-spinner").Cast().(*adw.Spinner)

			contactsCreateDialogFirstNameInput = contactsCreateDialogBuilder.GetObject("contacts-create-dialog-first-name-input").Cast().(*adw.EntryRow)
			contactsCreateDialogLastNameInput  = contactsCreateDialogBuilder.GetObject("contacts-create-dialog-last-name-input").Cast().(*adw.EntryRow)
			contactsCreateDialogNicknameInput  = contactsCreateDialogBuilder.GetObject("contacts-create-dialog-nickname-input").Cast().(*adw.EntryRow)
			contactsCreateDialogEmailInput     = contactsCreateDialogBuilder.GetObject("contacts-create-dialog-email-input").Cast().(*adw.EntryRow)
			contactsCreateDialogPronounsInput  = contactsCreateDialogBuilder.GetObject("contacts-create-dialog-pronouns-input").Cast().(*adw.EntryRow)

			contactsCreateDialogEmailWarningButton = contactsCreateDialogBuilder.GetObject("contacts-create-dialog-email-warning-button").Cast().(*gtk.MenuButton)

			contactsErrorStatusPage        = b.GetObject("contacts-error-status-page").Cast().(*adw.StatusPage)
			contactsErrorRefreshButton     = b.GetObject("contacts-error-refresh-button").Cast().(*gtk.Button)
			contactsErrorCopyDetailsButton = b.GetObject("contacts-error-copy-details").Cast().(*gtk.Button)

			contactsViewPageTitle              = b.GetObject("contacts-view-page-title").Cast().(*adw.WindowTitle)
			contactsViewStack                  = b.GetObject("contacts-view-stack").Cast().(*gtk.Stack)
			contactsViewErrorStatusPage        = b.GetObject("contacts-view-error-status-page").Cast().(*adw.StatusPage)
			contactsViewErrorRefreshButton     = b.GetObject("contacts-view-error-refresh-button").Cast().(*gtk.Button)
			contactsViewErrorCopyDetailsButton = b.GetObject("contacts-view-error-copy-details").Cast().(*gtk.Button)

			contactsViewOptionalFieldsPreferencesGroup = b.GetObject("contacts-view-optional-fields").Cast().(*adw.PreferencesGroup)

			contactsViewBirthdayRow = b.GetObject("contacts-view-birthday").Cast().(*adw.ActionRow)
			contactsViewAddressRow  = b.GetObject("contacts-view-address").Cast().(*adw.ActionRow)
			contactsViewNotesRow    = b.GetObject("contacts-view-notes").Cast().(*adw.ActionRow)

			contactsViewDebtsListBox      = b.GetObject("contacts-view-debts").Cast().(*gtk.ListBox)
			contactsViewActivitiesListBox = b.GetObject("contacts-view-activities").Cast().(*gtk.ListBox)

			activitiesViewPageTitle              = b.GetObject("activities-view-page-title").Cast().(*adw.WindowTitle)
			activitiesViewStack                  = b.GetObject("activities-view-stack").Cast().(*gtk.Stack)
			activitiesViewErrorStatusPage        = b.GetObject("activities-view-error-status-page").Cast().(*adw.StatusPage)
			activitiesViewErrorRefreshButton     = b.GetObject("activities-view-error-refresh-button").Cast().(*gtk.Button)
			activitiesViewErrorCopyDetailsButton = b.GetObject("activities-view-error-copy-details").Cast().(*gtk.Button)

			activitiesViewPageBodyWebView = b.GetObject("activities-view-body").Cast().(*webkit.WebView)

			debtsCreateDialog = debtsCreateDialogBuilder.GetObject("debts-create-dialog").Cast().(*adw.Dialog)

			debtsCreateDialogAddButton  = debtsCreateDialogBuilder.GetObject("debts-create-dialog-add-button").Cast().(*gtk.Button)
			debtsCreateDialogAddSpinner = debtsCreateDialogBuilder.GetObject("debts-create-dialog-add-spinner").Cast().(*adw.Spinner)

			debtsCreateDialogTitle = debtsCreateDialogBuilder.GetObject("debts-create-dialog-title").Cast().(*adw.WindowTitle)

			debtsCreateDialogYouOweRadio         = debtsCreateDialogBuilder.GetObject("debts-create-dialog-you-owe-radio").Cast().(*gtk.CheckButton)
			debtsCreateDialogAmountInput         = debtsCreateDialogBuilder.GetObject("debts-create-dialog-amount-input").Cast().(*adw.SpinRow)
			debtsCreateDialogCurrencyInput       = debtsCreateDialogBuilder.GetObject("debts-create-dialog-currency-input").Cast().(*adw.EntryRow)
			debtsCreateDialogDescriptionExpander = debtsCreateDialogBuilder.GetObject("debts-create-dialog-description-expander").Cast().(*adw.ExpanderRow)
			debtsCreateDialogDescriptionInput    = debtsCreateDialogBuilder.GetObject("debts-create-dialog-description-input").Cast().(*gtk.TextView)

			debtsCreateDialogYouOweActionRow  = debtsCreateDialogBuilder.GetObject("debts-create-dialog-debt-type-you-owe-row").Cast().(*adw.ActionRow)
			debtsCreateDialogTheyOweActionRow = debtsCreateDialogBuilder.GetObject("debts-create-dialog-debt-type-they-owe-row").Cast().(*adw.ActionRow)
		)

		welcomeGetStartedButton.ConnectClicked(func() {
			nv.PushByTag(resources.PageConfigServerURL)
		})

		settings.Bind(resources.SettingServerURLKey, configServerURLInput.Object, "text", gio.SettingsBindDefault)

		updateConfigServerURLContinueButtonSensitive := func() {
			if len(settings.String(resources.SettingServerURLKey)) > 0 {
				configServerURLContinueButton.SetSensitive(true)
			} else {
				configServerURLContinueButton.SetSensitive(false)
			}
		}

		var deregistrationLock sync.Mutex
		deregisterOIDCClient := func() error {
			deregistrationLock.Lock()
			defer deregistrationLock.Unlock()

			if registrationClientURI := settings.String(resources.SettingRegistrationClientURIKey); registrationClientURI != "" {
				registrationAccessToken, err := keyring.Get(resources.AppID, resources.SecretRegistrationAccessToken)
				if err != nil {
					return err
				}

				if err := authn.DeregisterOIDCClient(
					ctx,

					slog.New(log.Handler().WithGroup("oidcDeregistration")),

					registrationAccessToken,
					registrationClientURI,
				); err != nil {
					return err
				}

				if ok := settings.SetString(resources.SettingRegistrationClientURIKey, ""); !ok {
					return errCouldNotWriteSettingsKey
				}

				if err := keyring.Delete(resources.AppID, resources.SecretRegistrationAccessToken); err != nil && !errors.Is(err, keyring.ErrNotFound) {
					return err
				}
			}

			// We indiscriminately clear the client ID, even if the client was never registered
			// via OIDC dynamic client registration so that we can switch Senbara servers (which
			// configure different OIDC endpoints and thus expect different OIDC client IDs) properly
			if ok := settings.SetString(resources.SettingOIDCClientIDKey, ""); !ok {
				return errCouldNotWriteSettingsKey
			}

			return nil
		}

		deregisterClientAction := gio.NewSimpleAction("deregisterClient", nil)

		updateDeregisterClientActionEnabled := func() {
			deregisterClientAction.SetEnabled(settings.String(resources.SettingOIDCClientIDKey) != "")
		}

		deregisterClientAction.ConnectActivate(func(parameter *glib.Variant) {
			configServerURLContinueButton.SetSensitive(false)
			welcomeGetStartedButton.SetSensitive(false)
			configServerURLContinueSpinner.SetVisible(true)
			welcomeGetStartedSpinner.SetVisible(true)

			go func() {
				defer welcomeGetStartedButton.SetSensitive(true)
				defer configServerURLContinueSpinner.SetVisible(false)
				defer welcomeGetStartedSpinner.SetVisible(false)

				if err := deregisterOIDCClient(); err != nil {
					handlePanic(err)

					return
				}

				updateDeregisterClientActionEnabled()
				updateConfigServerURLContinueButtonSensitive()
			}()
		})
		a.AddAction(deregisterClientAction)

		settings.ConnectChanged(func(key string) {
			if key == resources.SettingServerURLKey {
				configServerURLContinueButton.SetSensitive(false)
				configServerURLContinueSpinner.SetVisible(true)

				go func() {
					defer configServerURLContinueSpinner.SetVisible(false)

					if err := deregisterOIDCClient(); err != nil {
						handlePanic(err)

						return
					}

					updateDeregisterClientActionEnabled()
					updateConfigServerURLContinueButtonSensitive()
				}()
			}
		})

		checkSenbaraServerConfiguration := func() error {
			var (
				serverURL = settings.String(resources.SettingServerURLKey)
			)

			client, err := api.NewClientWithResponses(serverURL)
			if err != nil {
				return err
			}

			log.Debug("Getting OpenAPI spec")

			res, err := client.GetOpenAPISpec(ctx)
			if err != nil {
				return err
			}

			log.Debug("Got OpenAPI spec", "status", res.StatusCode)

			if res.StatusCode != http.StatusOK {
				return errors.New(res.Status)
			}

			if err := json.NewDecoder(res.Body).Decode(&spec); err != nil {
				return err
			}

			return nil
		}

		setupAuthn := func(registerClient bool) error {
			o, err := authn.DiscoverOIDCProviderConfiguration(
				ctx,

				slog.New(log.Handler().WithGroup("oidcDiscovery")),

				spec.Components.SecuritySchemes["oidc"].Value.OpenIdConnectUrl,
			)
			if err != nil {
				return err
			}

			oidcClientID := settings.String(resources.SettingOIDCClientIDKey)
			if oidcClientID == "" && registerClient {
				c, err := authn.RegisterOIDCClient(
					ctx,

					slog.New(log.Handler().WithGroup("oidcRegistration")),

					o,

					"Senbara GNOME",
					redirectURL,

					configInitialAccessTokenInput.Text(),
				)
				if err != nil {
					return err
				}

				if ok := settings.SetString(resources.SettingOIDCClientIDKey, c.ClientID); !ok {
					return errCouldNotWriteSettingsKey
				}

				if ok := settings.SetString(resources.SettingRegistrationClientURIKey, c.RegistrationClientURI); !ok {
					return errCouldNotWriteSettingsKey
				}

				if err := keyring.Set(resources.AppID, resources.SecretRegistrationAccessToken, c.RegistrationAccessToken); err != nil {
					return err
				}

				oidcClientID = c.ClientID
			}

			updateDeregisterClientActionEnabled()

			authner = authn.NewAuthner(
				slog.New(log.Handler().WithGroup("authner")),

				o.Issuer,
				o.EndSessionEndpoint,

				oidcClientID,
				redirectURL,
			)

			if err := authner.Init(ctx); err != nil {
				return err
			}

			return nil
		}

		configServerURLContinueButton.ConnectClicked(func() {
			configServerURLContinueButton.SetSensitive(false)
			configServerURLContinueSpinner.SetVisible(true)

			go func() {
				defer configServerURLContinueSpinner.SetVisible(false)

				if err := checkSenbaraServerConfiguration(); err != nil {
					handlePanic(err)

					return
				}

				nv.PushByTag(resources.PagePreview)
			}()
		})

		previewLoginButton.ConnectClicked(func() {
			previewLoginButton.SetSensitive(false)
			previewLoginSpinner.SetVisible(true)

			go func() {
				defer previewLoginButton.SetSensitive(true)
				defer previewLoginSpinner.SetVisible(false)

				if err := checkSenbaraServerConfiguration(); err != nil {
					handlePanic(err)

					return
				}

				if err := setupAuthn(true); err != nil {
					handlePanic(err)

					return
				}

				if v := spec.Components.SecuritySchemes["oidc"].Value.Extensions[api.OidcDcrInitialAccessTokenPortalUrlExtensionKey]; v != nil {
					vv, ok := v.(string)
					if ok {
						oidcDcrInitialAccessTokenPortalUrl = vv

						nv.PushByTag(resources.PageRegister)

						return
					}
				}

				nv.PushByTag(resources.PageHome)
			}()
		})

		registerRegisterButton.ConnectClicked(func() {
			go func() {
				var (
					fl = gtk.NewURILauncher(oidcDcrInitialAccessTokenPortalUrl)
					cc = make(chan error)
				)
				fl.Launch(ctx, &w.Window, func(res gio.AsyncResulter) {
					if err := fl.LaunchFinish(res); err != nil {
						cc <- err

						return
					}

					cc <- nil
				})

				if err := <-cc; err != nil {
					handlePanic(err)

					return
				}

				nv.PushByTag(resources.PageConfigInitialAccessToken)
			}()
		})

		configInitialAccessTokenInput.ConnectChanged(func() {
			if configInitialAccessTokenInput.TextLength() > 0 {
				configInitialAccessTokenLoginButton.SetSensitive(true)
			} else {
				configInitialAccessTokenLoginButton.SetSensitive(false)
			}
		})

		configInitialAccessTokenLoginButton.ConnectClicked(func() {
			configInitialAccessTokenLoginButton.SetSensitive(false)
			configInitialAccessTokenLoginSpinner.SetVisible(true)

			go func() {
				defer configInitialAccessTokenLoginButton.SetSensitive(true)
				defer configInitialAccessTokenLoginSpinner.SetVisible(false)

				if err := checkSenbaraServerConfiguration(); err != nil {
					handlePanic(err)

					return
				}

				if err := setupAuthn(true); err != nil {
					handlePanic(err)

					return
				}

				nv.PushByTag(resources.PageHome)
			}()
		})

		selectDifferentServerAction := gio.NewSimpleAction("selectDifferentServer", nil)
		selectDifferentServerAction.ConnectActivate(func(parameter *glib.Variant) {
			nv.ReplaceWithTags([]string{resources.PageWelcome})
		})
		a.AddAction(selectDifferentServerAction)

		exchangeLoginCancelButton.ConnectClicked(func() {
			nv.ReplaceWithTags([]string{resources.PageWelcome})
		})

		exchangeLogoutCancelButton.ConnectClicked(func() {
			nv.ReplaceWithTags([]string{resources.PageHome})
		})

		enableHomeUserMenuLoading := func() {
			homeUserMenuButton.SetSensitive(false)
			homeUserMenuAvatar.SetVisible(false)
			homeUserMenuSpinner.SetVisible(true)
		}

		disableHomeUserMenuLoading := func() {
			homeUserMenuSpinner.SetVisible(false)
			homeUserMenuAvatar.SetVisible(true)
			homeUserMenuButton.SetSensitive(true)
		}

		enableHomeHamburgerMenuLoading := func() {
			homeHamburgerMenuButton.SetSensitive(false)
			homeHamburgerMenuIcon.SetVisible(false)
			homeHamburgerMenuSpinner.SetVisible(true)
		}

		disableHomeHamburgerMenuLoading := func() {
			homeHamburgerMenuSpinner.SetVisible(false)
			homeHamburgerMenuIcon.SetVisible(true)
			homeHamburgerMenuButton.SetSensitive(true)
		}

		enableHomeSidebarLoading := func() {
			homeSidebarContactsCountLabel.SetVisible(false)
			homeSidebarContactsCountSpinner.SetVisible(true)

			homeSidebarJournalEntriesCountLabel.SetVisible(false)
			homeSidebarJournalEntriesCountSpinner.SetVisible(true)
		}

		disableHomeSidebarLoading := func() {
			homeSidebarJournalEntriesCountSpinner.SetVisible(false)
			homeSidebarJournalEntriesCountLabel.SetVisible(true)

			homeSidebarContactsCountSpinner.SetVisible(false)
			homeSidebarContactsCountLabel.SetVisible(true)
		}

		var (
			contactsCount        = 0
			visibleContactsCount = 0
		)

		contactsSearchEntry.ConnectSearchChanged(func() {
			go func() {
				contactsStack.SetVisibleChildName(resources.PageContactsLoading)

				visibleContactsCount = 0

				contactsListBox.InvalidateFilter()

				if visibleContactsCount > 0 {
					contactsStack.SetVisibleChildName(resources.PageContactsList)
				} else {
					contactsStack.SetVisibleChildName(resources.PageContactsNoResults)
				}
			}()
		})

		contactsListBox.SetFilterFunc(func(row *gtk.ListBoxRow) (ok bool) {
			var (
				r = row.Cast().(*adw.ActionRow)
				f = strings.ToLower(contactsSearchEntry.Text())

				rt = strings.ToLower(r.Title())
				rs = strings.ToLower(r.Subtitle())
			)

			log.Debug(
				"Filtering contact",
				"filter", f,
				"title", rt,
				"subtitle", rs,
			)

			if strings.Contains(rt, f) || strings.Contains(rs, f) {
				visibleContactsCount++

				return true
			}

			return false
		})

		contactsAddButton.ConnectClicked(func() {
			contactsCreateDialog.Present(w)
		})

		contactsEmptyAddButton.ConnectClicked(func() {
			contactsCreateDialog.Present(w)
		})

		validateContactsCreateDialogForm := func() {
			if email := contactsCreateDialogEmailInput.Text(); email != "" {
				if _, err := mail.ParseAddress(email); err != nil {
					contactsCreateDialogEmailInput.AddCSSClass("error")
					contactsCreateDialogEmailWarningButton.SetVisible(true)

					contactsCreateDialogAddButton.SetSensitive(false)

					return
				}
			}

			contactsCreateDialogEmailInput.RemoveCSSClass("error")
			contactsCreateDialogEmailWarningButton.SetVisible(false)

			if contactsCreateDialogFirstNameInput.Text() != "" &&
				contactsCreateDialogLastNameInput.Text() != "" &&
				contactsCreateDialogEmailInput.Text() != "" &&
				contactsCreateDialogPronounsInput.Text() != "" {
				contactsCreateDialogAddButton.SetSensitive(true)
			} else {
				contactsCreateDialogAddButton.SetSensitive(false)
			}
		}

		contactsCreateDialogFirstNameInput.ConnectChanged(validateContactsCreateDialogForm)
		contactsCreateDialogLastNameInput.ConnectChanged(validateContactsCreateDialogForm)
		contactsCreateDialogNicknameInput.ConnectChanged(validateContactsCreateDialogForm)
		contactsCreateDialogEmailInput.ConnectChanged(validateContactsCreateDialogForm)
		contactsCreateDialogPronounsInput.ConnectChanged(validateContactsCreateDialogForm)

		contactsCreateDialog.ConnectClosed(func() {
			contactsCreateDialogFirstNameInput.SetText("")
			contactsCreateDialogLastNameInput.SetText("")
			contactsCreateDialogNicknameInput.SetText("")
			contactsCreateDialogEmailInput.SetText("")
			contactsCreateDialogPronounsInput.SetText("")

			contactsCreateDialogEmailWarningButton.SetVisible(false)
			contactsCreateDialogEmailInput.RemoveCSSClass("error")
		})

		validateDebtsCreateDialogForm := func() {
			if debtsCreateDialogCurrencyInput.Text() != "" {
				debtsCreateDialogAddButton.SetSensitive(true)
			} else {
				debtsCreateDialogAddButton.SetSensitive(false)
			}
		}

		debtsCreateDialogCurrencyInput.ConnectChanged(validateDebtsCreateDialogForm)

		debtsCreateDialog.ConnectClosed(func() {
			debtsCreateDialogYouOweRadio.SetActive(true)
			debtsCreateDialogAmountInput.SetValue(0)
			debtsCreateDialogCurrencyInput.SetText("")
			debtsCreateDialogDescriptionExpander.SetExpanded(false)
			debtsCreateDialogDescriptionInput.Buffer().SetText("")
		})

		createErrAndLoadingHandlers := func(
			errorStatusPage *adw.StatusPage,
			errorRefreshButton *gtk.Button,
			errorCopyDetailsButton *gtk.Button,

			handleRefresh func(),

			handleEnableLoading func(),
			handleDisableLoading func(err string),
		) (
			handleError func(error),
			enableLoading func(),
			disableLoading func(),
			clearError func(),
		) {
			var rawErr string
			handleError = func(err error) {
				rawErr = err.Error()
				i18nErr := gcore.Local(rawErr)

				log.Error(
					"An unexpected error occured, showing error message to user",
					"rawError", rawErr,
					"i18nErr", i18nErr,
				)

				errorStatusPage.SetDescription(i18nErr)
			}

			errorRefreshButton.ConnectClicked(handleRefresh)

			errorCopyDetailsButton.ConnectClicked(func() {
				w.Clipboard().SetText(rawErr)
			})

			enableLoading = handleEnableLoading

			disableLoading = func() {
				handleDisableLoading(rawErr)
			}

			return handleError,
				enableLoading,
				disableLoading,
				func() {
					rawErr = ""
				}
		}

		handleContactsError,
			enableContactsLoading,
			disableContactsLoading,
			clearContactsError := createErrAndLoadingHandlers(
			contactsErrorStatusPage,
			contactsErrorRefreshButton,
			contactsErrorCopyDetailsButton,

			func() {
				homeNavigation.ReplaceWithTags([]string{resources.PageContacts})
			},

			func() {
				homeSidebarContactsCountLabel.SetVisible(false)
				homeSidebarContactsCountSpinner.SetVisible(true)

				contactsStack.SetVisibleChildName(resources.PageContactsLoading)
			},
			func(err string) {
				homeSidebarContactsCountSpinner.SetVisible(false)
				homeSidebarContactsCountLabel.SetVisible(true)

				homeSidebarContactsCountLabel.SetText(fmt.Sprintf("%v", contactsCount))

				if err == "" {
					if contactsCount > 0 {
						contactsStack.SetVisibleChildName(resources.PageContactsList)
					} else {
						contactsStack.SetVisibleChildName(resources.PageContactsEmpty)
					}
				} else {
					contactsStack.SetVisibleChildName(resources.PageContactsError)
				}
			},
		)

		handleContactsViewError,
			enableContactsViewLoading,
			disableContactsViewLoading,
			clearContactsViewError := createErrAndLoadingHandlers(
			contactsViewErrorStatusPage,
			contactsViewErrorRefreshButton,
			contactsViewErrorCopyDetailsButton,

			func() {
				homeNavigation.ReplaceWithTags([]string{resources.PageContacts, resources.PageContactsView})
			},

			func() {
				contactsViewStack.SetVisibleChildName(resources.PageContactsViewLoading)
			},
			func(err string) {
				if err == "" {
					contactsViewStack.SetVisibleChildName(resources.PageContactsViewData)
				} else {
					contactsViewStack.SetVisibleChildName(resources.PageContactsViewError)
				}
			},
		)

		handleActivitiesViewError,
			enableActivitiesViewLoading,
			disableActivitiesViewLoading,
			clearActivitiesViewError := createErrAndLoadingHandlers(
			activitiesViewErrorStatusPage,
			activitiesViewErrorRefreshButton,
			activitiesViewErrorCopyDetailsButton,

			func() {
				homeNavigation.ReplaceWithTags([]string{resources.PageContacts, resources.PageContactsView, resources.PageActivitiesView})
			},

			func() {
				activitiesViewStack.SetVisibleChildName(resources.PageActivitiesViewLoading)
			},
			func(err string) {
				if err == "" {
					activitiesViewStack.SetVisibleChildName(resources.PageActivitiesViewData)
				} else {
					activitiesViewStack.SetVisibleChildName(resources.PageActivitiesViewError)
				}
			},
		)

		contactsCreateDialogAddButton.ConnectClicked(func() {
			log.Info("Handling contact creation")

			contactsCreateDialogAddButton.SetSensitive(false)
			contactsCreateDialogAddSpinner.SetVisible(true)

			go func() {
				defer contactsCreateDialogAddSpinner.SetVisible(false)
				defer contactsCreateDialogAddButton.SetSensitive(true)

				redirected, c, _, err := authorize(
					ctx,

					false,
				)
				if err != nil {
					log.Warn("Could not authorize user for create contact action", "err", err)

					handlePanic(err)

					return
				} else if redirected {
					return
				}

				var nickname *string
				if v := contactsCreateDialogNicknameInput.Text(); v != "" {
					nickname = &v
				}

				req := api.CreateContactJSONRequestBody{
					Email:     (types.Email)(contactsCreateDialogEmailInput.Text()),
					FirstName: contactsCreateDialogFirstNameInput.Text(),
					LastName:  contactsCreateDialogLastNameInput.Text(),
					Nickname:  nickname,
					Pronouns:  contactsCreateDialogPronounsInput.Text(),
				}

				log.Debug("Creating contact", "request", req)

				res, err := c.CreateContactWithResponse(ctx, req)
				if err != nil {
					handlePanic(err)

					return
				}

				log.Debug("Created contact", "status", res.StatusCode())

				if res.StatusCode() != http.StatusOK {
					handlePanic(errors.New(res.Status()))

					return
				}

				mto.AddToast(adw.NewToast(gcore.Local("Created contact")))

				contactsCreateDialog.Close()

				homeNavigation.ReplaceWithTags([]string{resources.PageContacts})
			}()
		})

		debtsCreateDialogAddButton.ConnectClicked(func() {
			id := debtsCreateDialogAddButton.ActionTargetValue().Int64()

			log := log.With(
				"id", id,
			)

			log.Info("Handling debt creation")

			debtsCreateDialogAddButton.SetSensitive(false)
			debtsCreateDialogAddSpinner.SetVisible(true)

			go func() {
				defer debtsCreateDialogAddSpinner.SetVisible(false)
				defer debtsCreateDialogAddButton.SetSensitive(true)

				redirected, c, _, err := authorize(
					ctx,

					false,
				)
				if err != nil {
					log.Warn("Could not authorize user for create debt action", "err", err)

					handlePanic(err)

					return
				} else if redirected {
					return
				}

				var description *string
				if v := debtsCreateDialogDescriptionInput.Buffer().Text(
					debtsCreateDialogDescriptionInput.Buffer().StartIter(),
					debtsCreateDialogDescriptionInput.Buffer().EndIter(),
					true,
				); v != "" {
					description = &v
				}

				req := api.CreateDebtJSONRequestBody{
					Amount:      float32(debtsCreateDialogAmountInput.Value()),
					ContactId:   id,
					Currency:    debtsCreateDialogCurrencyInput.Text(),
					Description: description,
					YouOwe:      debtsCreateDialogYouOweRadio.Active(),
				}

				log.Debug("Creating debt", "request", req)

				res, err := c.CreateDebtWithResponse(ctx, req)
				if err != nil {
					handlePanic(err)

					return
				}

				log.Debug("Created debt", "status", res.StatusCode())

				if res.StatusCode() != http.StatusOK {
					handlePanic(errors.New(res.Status()))

					return
				}

				mto.AddToast(adw.NewToast(gcore.Local("Created debt")))

				debtsCreateDialog.Close()

				homeNavigation.ReplaceWithTags([]string{resources.PageContacts, resources.PageContactsView})
			}()
		})

		logoutAction := gio.NewSimpleAction("logout", nil)
		logoutAction.ConnectActivate(func(parameter *glib.Variant) {
			nv.ReplaceWithTags([]string{resources.PageExchangeLogout})

			go func() {
				var (
					fl = gtk.NewURILauncher(u.LogoutURL)
					cc = make(chan error)
				)
				fl.Launch(ctx, &w.Window, func(res gio.AsyncResulter) {
					if err := fl.LaunchFinish(res); err != nil {
						cc <- err

						return
					}

					cc <- nil
				})

				if err := <-cc; err != nil {
					handlePanic(err)

					return
				}
			}()
		})
		a.AddAction(logoutAction)

		licenseAction := gio.NewSimpleAction("license", nil)
		licenseAction.ConnectActivate(func(parameter *glib.Variant) {
			log.Info("Handling getting license action", "url", spec.Info.License.URL)

			go func() {
				var (
					fl = gtk.NewURILauncher(spec.Info.License.URL)
					cc = make(chan error)
				)
				fl.Launch(ctx, &w.Window, func(res gio.AsyncResulter) {
					if err := fl.LaunchFinish(res); err != nil {
						cc <- err

						return
					}

					cc <- nil
				})

				if err := <-cc; err != nil {
					handlePanic(err)

					return
				}
			}()
		})
		a.AddAction(licenseAction)

		privacyAction := gio.NewSimpleAction("privacy", nil)
		privacyAction.ConnectActivate(func(parameter *glib.Variant) {
			var privacyURL string
			if v := spec.Info.Extensions[api.PrivacyPolicyExtensionKey]; v != nil {
				vv, ok := v.(string)
				if ok {
					privacyURL = vv
				} else {
					handlePanic(errMissingPrivacyURL)

					return
				}
			}

			log.Info("Handling getting privacy action", "url", privacyURL)

			go func() {
				var (
					fl = gtk.NewURILauncher(privacyURL)
					cc = make(chan error)
				)
				fl.Launch(ctx, &w.Window, func(res gio.AsyncResulter) {
					if err := fl.LaunchFinish(res); err != nil {
						cc <- err

						return
					}

					cc <- nil
				})

				if err := <-cc; err != nil {
					handlePanic(err)

					return
				}
			}()
		})
		a.AddAction(privacyAction)

		tosAction := gio.NewSimpleAction("tos", nil)
		tosAction.ConnectActivate(func(parameter *glib.Variant) {
			log.Info("Handling getting terms of service action", "url", spec.Info.TermsOfService)

			go func() {
				var (
					fl = gtk.NewURILauncher(spec.Info.TermsOfService)
					cc = make(chan error)
				)
				fl.Launch(ctx, &w.Window, func(res gio.AsyncResulter) {
					if err := fl.LaunchFinish(res); err != nil {
						cc <- err

						return
					}

					cc <- nil
				})

				if err := <-cc; err != nil {
					handlePanic(err)

					return
				}
			}()
		})
		a.AddAction(tosAction)

		imprintAction := gio.NewSimpleAction("imprint", nil)
		imprintAction.ConnectActivate(func(parameter *glib.Variant) {
			log.Info("Handling getting imprint action", "url", spec.Info.Contact.URL)

			go func() {
				var (
					fl = gtk.NewURILauncher(spec.Info.Contact.URL)
					cc = make(chan error)
				)
				fl.Launch(ctx, &w.Window, func(res gio.AsyncResulter) {
					if err := fl.LaunchFinish(res); err != nil {
						cc <- err

						return
					}

					cc <- nil
				})

				if err := <-cc; err != nil {
					handlePanic(err)

					return
				}
			}()
		})
		a.AddAction(imprintAction)

		codeAction := gio.NewSimpleAction("code", nil)
		codeAction.ConnectActivate(func(parameter *glib.Variant) {
			log.Info("Handling getting code action")

			enableHomeHamburgerMenuLoading()

			redirected, c, _, err := authorize(
				ctx,

				false,
			)
			if err != nil {
				disableHomeHamburgerMenuLoading()

				log.Warn("Could not authorize user for getting code action", "err", err)

				handlePanic(err)

				return
			} else if redirected {
				disableHomeHamburgerMenuLoading()

				return
			}

			log.Debug("Getting code")

			res, err := c.GetSourceCode(ctx)
			if err != nil {
				disableHomeHamburgerMenuLoading()

				handlePanic(err)

				return
			}

			log.Debug("Received code", "status", res.StatusCode)

			if res.StatusCode != http.StatusOK {
				_ = res.Body.Close()

				disableHomeHamburgerMenuLoading()

				handlePanic(errors.New(res.Status))

				return
			}

			log.Debug("Writing code to file")

			fd := gtk.NewFileDialog()
			fd.SetTitle(gcore.Local("Senbara REST source code"))
			fd.SetInitialName("code.tar.gz")
			fd.Save(ctx, &w.Window, func(r gio.AsyncResulter) {
				go func() {
					defer disableHomeHamburgerMenuLoading()
					defer res.Body.Close()

					fp, err := fd.SaveFinish(r)
					if err != nil {
						handlePanic(err)

						return
					}

					log.Debug("Writing code to file", "path", fp.Path())

					f, err := os.OpenFile(fp.Path(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
					if err != nil {
						handlePanic(err)

						return
					}
					defer f.Close()

					if _, err := io.Copy(f, res.Body); err != nil {
						handlePanic(err)

						return
					}

					log.Debug("Downloaded code", "status", res.StatusCode)

					mto.AddToast(adw.NewToast(gcore.Local("Downloaded code")))
				}()
			})
		})
		a.AddAction(codeAction)

		exportUserDataAction := gio.NewSimpleAction("exportUserData", nil)
		exportUserDataAction.ConnectActivate(func(parameter *glib.Variant) {
			log.Info("Handling export user data action")

			enableHomeUserMenuLoading()

			redirected, c, _, err := authorize(
				ctx,

				false,
			)
			if err != nil {
				disableHomeUserMenuLoading()

				log.Warn("Could not authorize user for export user data action", "err", err)

				handlePanic(err)

				return
			} else if redirected {
				disableHomeUserMenuLoading()

				return
			}

			log.Debug("Exporting user data")

			res, err := c.ExportUserData(ctx)
			if err != nil {
				disableHomeUserMenuLoading()

				handlePanic(err)

				return
			}

			log.Debug("Exported user data", "status", res.StatusCode)

			if res.StatusCode != http.StatusOK {
				_ = res.Body.Close()

				disableHomeUserMenuLoading()

				handlePanic(errors.New(res.Status))

				return
			}

			log.Debug("Writing user data to file")

			fd := gtk.NewFileDialog()
			fd.SetTitle(gcore.Local("Senbara Forms userdata"))
			fd.SetInitialName("userdata.jsonl")
			fd.Save(ctx, &w.Window, func(r gio.AsyncResulter) {
				go func() {
					defer disableHomeUserMenuLoading()
					defer res.Body.Close()

					fp, err := fd.SaveFinish(r)
					if err != nil {
						handlePanic(err)

						return
					}

					log.Debug("Writing user data to file", "path", fp.Path())

					f, err := os.OpenFile(fp.Path(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
					if err != nil {
						handlePanic(err)

						return
					}
					defer f.Close()

					if _, err := io.Copy(f, res.Body); err != nil {
						handlePanic(err)

						return
					}

					log.Debug("Exported user data", "status", res.StatusCode)

					mto.AddToast(adw.NewToast(gcore.Local("Exported user data")))
				}()
			})
		})
		a.AddAction(exportUserDataAction)

		importUserDataAction := gio.NewSimpleAction("importUserData", nil)
		importUserDataAction.ConnectActivate(func(parameter *glib.Variant) {
			log.Info("Handling import user data action")

			fd := gtk.NewFileDialog()
			fd.SetTitle(gcore.Local("Senbara Forms userdata"))

			ls := gio.NewListStore(glib.TypeObject)

			{
				fi := gtk.NewFileFilter()
				fi.SetName(gcore.Local("Senbara Forms userdata files"))
				fi.AddPattern("*.jsonl")
				ls.Append(fi.Object)
			}

			{
				fi := gtk.NewFileFilter()
				fi.SetName(gcore.Local("All files"))
				fi.AddPattern("*")
				ls.Append(fi.Object)
			}

			fd.SetFilters(ls)

			fd.Open(ctx, &w.Window, func(r gio.AsyncResulter) {
				fp, err := fd.OpenFinish(r)
				if err != nil {
					handlePanic(err)

					return
				}

				confirm := adw.NewAlertDialog(
					gcore.Local("Importing user data"),
					gcore.Local("Are you sure you want to import this user data into your account?"),
				)
				confirm.AddResponse("cancel", gcore.Local("Cancel"))
				confirm.AddResponse("import", gcore.Local("Import"))
				confirm.SetResponseAppearance("import", adw.ResponseSuggested)
				confirm.ConnectResponse(func(response string) {
					if response == "import" {
						go func() {
							enableHomeUserMenuLoading()
							defer disableHomeUserMenuLoading()

							redirected, c, _, err := authorize(
								ctx,

								false,
							)
							if err != nil {
								disableHomeUserMenuLoading()

								log.Warn("Could not authorize user for import user data action", "err", err)

								handlePanic(err)

								return
							} else if redirected {
								disableHomeUserMenuLoading()

								return
							}

							log.Debug("Reading user data from file", "path", fp.Path())

							f, err := os.OpenFile(fp.Path(), os.O_RDONLY, os.ModePerm)
							if err != nil {
								handlePanic(err)

								return
							}
							defer f.Close()

							log.Debug("Importing user data, reading from file and streaming to API")

							reader, writer := io.Pipe()
							enc := multipart.NewWriter(writer)
							go func() {
								defer writer.Close()

								if err := func() error {
									file, err := enc.CreateFormFile("userData", "")
									if err != nil {
										return err
									}

									if _, err := io.Copy(file, f); err != nil {
										return err
									}

									if err := enc.Close(); err != nil {
										return err
									}

									return nil
								}(); err != nil {
									log.Warn("Could not stream user data to API", "err", err)

									writer.CloseWithError(err)

									return
								}
							}()

							res, err := c.ImportUserDataWithBodyWithResponse(ctx, enc.FormDataContentType(), reader)
							if err != nil {
								handlePanic(err)

								return
							}

							log.Debug("Imported user data", "status", res.StatusCode())

							if res.StatusCode() != http.StatusOK {
								handlePanic(errors.New(res.Status()))

								return
							}

							mto.AddToast(adw.NewToast(gcore.Local("Imported user data")))
						}()
					}
				})

				confirm.Present(w)
			})
		})
		a.AddAction(importUserDataAction)

		deleteUserDataAction := gio.NewSimpleAction("deleteUserData", nil)
		deleteUserDataAction.ConnectActivate(func(parameter *glib.Variant) {
			log.Info("Handling delete user data action")

			confirm := adw.NewAlertDialog(
				gcore.Local("Deleting your data"),
				gcore.Local("Are you sure you want to delete your data and your account?"),
			)
			confirm.AddResponse("cancel", gcore.Local("Cancel"))
			confirm.AddResponse("delete", gcore.Local("Delete"))
			confirm.SetResponseAppearance("delete", adw.ResponseDestructive)
			confirm.ConnectResponse(func(response string) {
				if response == "delete" {
					redirected, c, _, err := authorize(
						ctx,

						false,
					)
					if err != nil {
						log.Warn("Could not authorize user for delete user data action", "err", err)

						handlePanic(err)

						return
					} else if redirected {
						return
					}

					log.Debug("Deleting user data")

					res, err := c.DeleteUserDataWithResponse(ctx)
					if err != nil {
						handlePanic(err)

						return
					}

					log.Debug("Deleted user data", "status", res.StatusCode())

					if res.StatusCode() != http.StatusOK {
						handlePanic(errors.New(res.Status()))

						return
					}

					mto.AddToast(adw.NewToast(gcore.Local("Deleted user data")))
				}
			})

			confirm.Present(w)
		})
		a.AddAction(deleteUserDataAction)

		aboutAction := gio.NewSimpleAction("about", nil)
		aboutAction.ConnectActivate(func(parameter *glib.Variant) {
			aboutDialog.Present(&w.Window)
		})
		a.AddAction(aboutAction)

		copyErrorToClipboardAction := gio.NewSimpleAction("copyErrorToClipboard", nil)
		copyErrorToClipboardAction.ConnectActivate(func(parameter *glib.Variant) {
			w.Clipboard().SetText(rawError)
		})
		a.AddAction(copyErrorToClipboardAction)

		deleteContactAction := gio.NewSimpleAction("deleteContact", glib.NewVariantType("x"))
		deleteContactAction.ConnectActivate(func(parameter *glib.Variant) {
			id := parameter.Int64()

			log := log.With(
				"id", id,
			)

			log.Info("Handling delete contact action")

			confirm := adw.NewAlertDialog(
				gcore.Local("Deleting a contact"),
				gcore.Local("Are you sure you want to delete this contact?"),
			)
			confirm.AddResponse("cancel", gcore.Local("Cancel"))
			confirm.AddResponse("delete", gcore.Local("Delete"))
			confirm.SetResponseAppearance("delete", adw.ResponseDestructive)
			confirm.ConnectResponse(func(response string) {
				if response == "delete" {
					redirected, c, _, err := authorize(
						ctx,

						false,
					)
					if err != nil {
						log.Warn("Could not authorize user for delete contact action", "err", err)

						handlePanic(err)

						return
					} else if redirected {
						return
					}

					log.Debug("Deleting contact")

					res, err := c.DeleteContactWithResponse(ctx, int64(id))
					if err != nil {
						handlePanic(err)

						return
					}

					log.Debug("Deleted contact", "status", res.StatusCode())

					if res.StatusCode() != http.StatusOK {
						handlePanic(errors.New(res.Status()))

						return
					}

					mto.AddToast(adw.NewToast(gcore.Local("Contact deleted")))

					homeNavigation.ReplaceWithTags([]string{resources.PageContacts})
				}
			})

			confirm.Present(w)
		})
		a.AddAction(deleteContactAction)

		settleDebtAction := gio.NewSimpleAction("settleDebt", glib.NewVariantType("x"))
		settleDebtAction.ConnectActivate(func(parameter *glib.Variant) {
			id := parameter.Int64()

			log := log.With(
				"id", id,
			)

			log.Info("Handling settle debt action")

			confirm := adw.NewAlertDialog(
				gcore.Local("Settling a debt"),
				gcore.Local("Are you sure you want to settle this debt?"),
			)
			confirm.AddResponse("cancel", gcore.Local("Cancel"))
			confirm.AddResponse("delete", gcore.Local("Delete"))
			confirm.SetResponseAppearance("delete", adw.ResponseDestructive)
			confirm.ConnectResponse(func(response string) {
				if response == "delete" {
					redirected, c, _, err := authorize(
						ctx,

						false,
					)
					if err != nil {
						log.Warn("Could not authorize user for settle debt action", "err", err)

						handlePanic(err)

						return
					} else if redirected {
						return
					}

					log.Debug("Settling debt")

					res, err := c.SettleDebtWithResponse(ctx, int64(id))
					if err != nil {
						handlePanic(err)

						return
					}

					log.Debug("Settled debt", "status", res.StatusCode())

					if res.StatusCode() != http.StatusOK {
						handlePanic(errors.New(res.Status()))

						return
					}

					mto.AddToast(adw.NewToast(gcore.Local("Settled debt")))

					homeNavigation.ReplaceWithTags([]string{resources.PageContacts, resources.PageContactsView})
				}
			})

			confirm.Present(w)
		})
		a.AddAction(settleDebtAction)

		deleteActivityAction := gio.NewSimpleAction("deleteActivity", glib.NewVariantType("x"))
		deleteActivityAction.ConnectActivate(func(parameter *glib.Variant) {
			id := parameter.Int64()

			log := log.With(
				"id", id,
			)

			log.Info("Handling delete activity action")

			confirm := adw.NewAlertDialog(
				gcore.Local("Deleting an activity"),
				gcore.Local("Are you sure you want to delete this activity?"),
			)
			confirm.AddResponse("cancel", gcore.Local("Cancel"))
			confirm.AddResponse("delete", gcore.Local("Delete"))
			confirm.SetResponseAppearance("delete", adw.ResponseDestructive)
			confirm.ConnectResponse(func(response string) {
				if response == "delete" {
					redirected, c, _, err := authorize(
						ctx,

						false,
					)
					if err != nil {
						log.Warn("Could not authorize user for delete activity action", "err", err)

						handlePanic(err)

						return
					} else if redirected {
						return
					}

					log.Debug("Deleting activity")

					res, err := c.DeleteActivityWithResponse(ctx, int64(id))
					if err != nil {
						handlePanic(err)

						return
					}

					log.Debug("Deleted activity", "status", res.StatusCode())

					if res.StatusCode() != http.StatusOK {
						handlePanic(errors.New(res.Status()))

						return
					}

					mto.AddToast(adw.NewToast(gcore.Local("Activity deleted")))

					homeNavigation.ReplaceWithTags([]string{resources.PageContacts, resources.PageContactsView})
				}
			})

			confirm.Present(w)
		})
		a.AddAction(deleteActivityAction)

		md := goldmark.New(
			goldmark.WithExtensions(extension.GFM),
		)

		var (
			selectedContactID  = -1
			selectedActivityID = -1
		)

		handleHomeNavigation := func() {
			var (
				tag = homeNavigation.VisiblePage().Tag()
				log = log.With("tag", tag)
			)

			log.Info("Handling page")

			homeContentPage.SetTitle(homeNavigation.VisiblePage().Title())

			homeSplitView.SetShowContent(true)

			switch tag {
			case resources.PageContacts:
				go func() {
					enableContactsLoading()
					defer disableContactsLoading()

					redirected, c, _, err := authorize(
						ctx,

						true,
					)
					if err != nil {
						log.Warn("Could not authorize user for contacts page", "err", err)

						handleContactsError(err)

						return
					} else if redirected {
						return
					}

					log.Debug("Listing contacts")

					res, err := c.GetContactsWithResponse(ctx)
					if err != nil {
						handleContactsError(err)

						return
					}

					log.Debug("Got contacts", "status", res.StatusCode())

					if res.StatusCode() != http.StatusOK {
						handleContactsError(errors.New(res.Status()))

						return
					}

					defer clearContactsError()

					validateContactsCreateDialogForm()

					contactsListBox.RemoveAll()

					contactsCount = len(*res.JSON200)
					if contactsCount > 0 {
						contactsAddButton.SetVisible(true)
						contactsSearchButton.SetVisible(true)

						for _, contact := range *res.JSON200 {
							r := adw.NewActionRow()

							title := *contact.FirstName + " " + *contact.LastName
							if *contact.Nickname != "" {
								title += " (" + *contact.Nickname + ")"
							}
							r.SetTitle(title)

							subtitle := ""
							if *contact.Email != "" {
								subtitle = string(*contact.Email)
							}
							if string(*contact.Email) != "" && string(*contact.Pronouns) != "" {
								subtitle += " | " + *contact.Pronouns
							} else if string(*contact.Pronouns) != "" {
								subtitle = *contact.Pronouns
							}
							if subtitle != "" {
								r.SetSubtitle(subtitle)
							}

							r.SetName("/contacts/view?id=" + strconv.Itoa(int(*contact.Id)))

							menuButton := gtk.NewMenuButton()
							menuButton.SetVAlign(gtk.AlignCenter)
							menuButton.SetIconName("view-more-symbolic")
							menuButton.AddCSSClass("flat")

							menu := gio.NewMenu()

							deleteContactMenuItem := gio.NewMenuItem(gcore.Local("Delete contact"), "app.deleteContact")
							deleteContactMenuItem.SetActionAndTargetValue("app.deleteContact", glib.NewVariantInt64(*contact.Id))
							menu.AppendItem(deleteContactMenuItem)

							editContactMenuItem := gio.NewMenuItem(gcore.Local("Edit contact"), "app.editContact")
							editContactMenuItem.SetActionAndTargetValue("app.editContact", glib.NewVariantInt64(*contact.Id))
							menu.AppendItem(editContactMenuItem)

							menuButton.SetMenuModel(menu)

							r.AddSuffix(menuButton)

							r.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))

							contactsListBox.Append(r)
						}
					} else {
						contactsAddButton.SetVisible(false)
						contactsSearchButton.SetVisible(false)
					}
				}()

			case resources.PageContactsView:
				go func() {
					enableContactsViewLoading()
					defer disableContactsViewLoading()

					redirected, c, _, err := authorize(
						ctx,

						true,
					)
					if err != nil {
						log.Warn("Could not authorize user for contacts view page", "err", err)

						handleContactsViewError(err)

						return
					} else if redirected {
						return
					}

					log.Debug("Getting contact", "id", selectedContactID)

					res, err := c.GetContactWithResponse(ctx, int64(selectedContactID))
					if err != nil {
						handleContactsViewError(err)

						return
					}

					log.Debug("Got contact", "status", res.StatusCode())

					if res.StatusCode() != http.StatusOK {
						handleContactsViewError(errors.New(res.Status()))

						return
					}

					defer clearContactsViewError()

					title := *res.JSON200.Entry.FirstName + " " + *res.JSON200.Entry.LastName
					if *res.JSON200.Entry.Nickname != "" {
						title += " (" + *res.JSON200.Entry.Nickname + ")"
					}
					contactsViewPageTitle.SetTitle(*res.JSON200.Entry.FirstName + " " + *res.JSON200.Entry.LastName)

					subtitle := ""
					if *res.JSON200.Entry.Email != "" {
						subtitle = string(*res.JSON200.Entry.Email)
					}
					if string(*res.JSON200.Entry.Email) != "" && string(*res.JSON200.Entry.Pronouns) != "" {
						subtitle += " | " + *res.JSON200.Entry.Pronouns
					} else if string(*res.JSON200.Entry.Pronouns) != "" {
						subtitle = *res.JSON200.Entry.Pronouns
					}
					if subtitle != "" {
						contactsViewPageTitle.SetSubtitle(subtitle)
					}

					var (
						birthday = res.JSON200.Entry.Birthday
						address  = res.JSON200.Entry.Address
						notes    = res.JSON200.Entry.Notes
					)
					if (birthday != nil) || (*address != "") || (*notes != "") {
						if birthday != nil {
							contactsViewBirthdayRow.SetVisible(true)
							contactsViewBirthdayRow.SetSubtitle(glib.NewDateTimeFromGo(birthday.Time).Format("%x"))
						} else {
							contactsViewBirthdayRow.SetVisible(false)
						}

						if *address != "" {
							contactsViewAddressRow.SetVisible(true)
							contactsViewAddressRow.SetSubtitle(*address)
						} else {
							contactsViewAddressRow.SetVisible(false)
						}

						if *notes != "" {
							contactsViewNotesRow.SetVisible(true)
							contactsViewNotesRow.SetSubtitle(*notes)
						} else {
							contactsViewNotesRow.SetVisible(false)
						}

						contactsViewOptionalFieldsPreferencesGroup.SetVisible(true)
					} else {
						contactsViewOptionalFieldsPreferencesGroup.SetVisible(false)
					}

					validateDebtsCreateDialogForm()

					debtsCreateDialogAddButton.SetActionTargetValue(glib.NewVariantInt64(*res.JSON200.Entry.Id))

					debtsCreateDialogTitle.SetSubtitle(*res.JSON200.Entry.FirstName + " " + *res.JSON200.Entry.LastName)

					debtsCreateDialogYouOweActionRow.SetTitle(gcore.Local(fmt.Sprintf("You owe %v", *res.JSON200.Entry.FirstName)))
					debtsCreateDialogTheyOweActionRow.SetTitle(gcore.Local(fmt.Sprintf("%v owes you", *res.JSON200.Entry.FirstName)))

					contactsViewDebtsListBox.RemoveAll()

					for _, debt := range *res.JSON200.Debts {
						r := adw.NewActionRow()

						subtitle := ""
						if *debt.Amount <= 0.0 {
							subtitle = gcore.Local(fmt.Sprintf("You owe %v %v %v", *res.JSON200.Entry.FirstName, math.Abs(float64(*debt.Amount)), *debt.Currency))
						} else {
							subtitle = gcore.Local(fmt.Sprintf("%v owes you %v %v", *res.JSON200.Entry.FirstName, math.Abs(float64(*debt.Amount)), *debt.Currency))
						}

						if *debt.Description != "" {
							subtitle += ": " + *debt.Description
						}

						r.SetTitle(subtitle)

						menuButton := gtk.NewMenuButton()
						menuButton.SetVAlign(gtk.AlignCenter)
						menuButton.SetIconName("view-more-symbolic")
						menuButton.AddCSSClass("flat")

						menu := gio.NewMenu()

						settleDebtMenuItem := gio.NewMenuItem(gcore.Local("Settle debt"), "app.settleDebt")
						settleDebtMenuItem.SetActionAndTargetValue("app.settleDebt", glib.NewVariantInt64(*debt.Id))

						menu.AppendItem(settleDebtMenuItem)

						editDebtMenuItem := gio.NewMenuItem(gcore.Local("Edit debt"), "app.editDebt")
						editDebtMenuItem.SetActionAndTargetValue("app.editDebt", glib.NewVariantInt64(*debt.Id))

						menu.AppendItem(editDebtMenuItem)

						menuButton.SetMenuModel(menu)

						r.AddSuffix(menuButton)

						contactsViewDebtsListBox.Append(r)
					}

					addDebtButton := adw.NewButtonRow()
					addDebtButton.SetStartIconName("list-add-symbolic")
					addDebtButton.SetTitle(gcore.Local("Add a debt"))

					addDebtButton.ConnectActivated(func() {
						debtsCreateDialog.Present(w)
					})

					contactsViewDebtsListBox.Append(addDebtButton)

					contactsViewActivitiesListBox.RemoveAll()

					for _, activity := range *res.JSON200.Activities {
						r := adw.NewActionRow()

						r.SetTitle(*activity.Name)
						r.SetSubtitle(glib.NewDateTimeFromGo(activity.Date.Time).Format("%x"))

						r.SetName("/activities/view?id=" + strconv.Itoa(int(*activity.Id)))

						menuButton := gtk.NewMenuButton()
						menuButton.SetVAlign(gtk.AlignCenter)
						menuButton.SetIconName("view-more-symbolic")
						menuButton.AddCSSClass("flat")

						menu := gio.NewMenu()

						deleteActivityMenuItem := gio.NewMenuItem(gcore.Local("Delete activity"), "app.deleteActivity")
						deleteActivityMenuItem.SetActionAndTargetValue("app.deleteActivity", glib.NewVariantInt64(*activity.Id))
						menu.AppendItem(deleteActivityMenuItem)

						editActivityMenuItem := gio.NewMenuItem(gcore.Local("Edit activity"), "app.editActivity")
						editActivityMenuItem.SetActionAndTargetValue("app.editActivity", glib.NewVariantInt64(*activity.Id))
						menu.AppendItem(editActivityMenuItem)

						menuButton.SetMenuModel(menu)

						r.AddSuffix(menuButton)

						r.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))

						contactsViewActivitiesListBox.Append(r)
					}

					addActivityButton := adw.NewButtonRow()
					addActivityButton.SetStartIconName("list-add-symbolic")
					addActivityButton.SetTitle(gcore.Local("Add an activity"))
					addActivityButton.SetSensitive(false)

					contactsViewActivitiesListBox.Append(addActivityButton)
				}()

			case resources.PageActivitiesView:
				go func() {
					enableActivitiesViewLoading()
					defer disableActivitiesViewLoading()

					redirected, c, _, err := authorize(
						ctx,

						true,
					)
					if err != nil {
						log.Warn("Could not authorize user for activities view page", "err", err)

						handleActivitiesViewError(err)

						return
					} else if redirected {
						return
					}

					log.Debug("Getting activity", "id", selectedActivityID)

					res, err := c.GetActivityWithResponse(ctx, int64(selectedActivityID))
					if err != nil {
						handleActivitiesViewError(err)

						return
					}

					log.Debug("Got activity", "status", res.StatusCode())

					if res.StatusCode() != http.StatusOK {
						handleActivitiesViewError(errors.New(res.Status()))

						return
					}

					activitiesViewPageTitle.SetTitle(*res.JSON200.Name)
					activitiesViewPageTitle.SetSubtitle(glib.NewDateTimeFromGo(res.JSON200.Date.Time).Format("%x"))

					var buf bytes.Buffer
					if err := md.Convert([]byte(*res.JSON200.Description), &buf); err != nil {
						log.Warn("Could not render Markdown for activities view page", "err", err)

						handleActivitiesViewError(err)

						return
					}

					bg := gdk.NewRGBA(0, 0, 0, 0)
					activitiesViewPageBodyWebView.SetBackgroundColor(&bg)

					activitiesViewPageBodyWebView.ConnectDecidePolicy(func(decision webkit.PolicyDecisioner, decisionType webkit.PolicyDecisionType) (ok bool) {
						if decisionType == webkit.PolicyDecisionTypeNavigationAction {
							u, err := url.Parse(decision.(*webkit.NavigationPolicyDecision).NavigationAction().Request().URI())
							if err != nil {
								log.Warn("Could not parse activity view WebView", "err", err)

								handleActivitiesViewError(err)

								return true
							}

							openExternally := u.Scheme != "about"

							log.Debug("Handling navigation in activity view WebView", "openExternally", openExternally, "url", u.String())

							if openExternally {
								go func() {
									var (
										fl = gtk.NewURILauncher(u.String())
										cc = make(chan error)
									)
									fl.Launch(ctx, &w.Window, func(res gio.AsyncResulter) {
										if err := fl.LaunchFinish(res); err != nil {
											cc <- err

											return
										}

										cc <- nil
									})

									if err := <-cc; err != nil {
										handlePanic(err)

										return
									}
								}()

								return true
							}

							return false
						}

						return false
					})

					glib.IdleAdd(func() {
						activitiesViewPageBodyWebView.LoadHtml(`<meta name="color-scheme" content="light dark" />`+buf.String(), "about:blank")
					})

					defer clearActivitiesViewError()
				}()
			}
		}

		contactsListBox.ConnectRowSelected(func(row *gtk.ListBoxRow) {
			if row != nil {
				u, err := url.Parse(row.Cast().(*adw.ActionRow).Name())
				if err != nil {
					log.Warn("Could not parse contact row URL", "err", err)

					handlePanic(err)

					return
				}

				rid := u.Query().Get("id")
				if strings.TrimSpace(rid) == "" {
					log.Warn("Could not get ID from contact row URL", "err", errMissingContactID)

					handlePanic(errMissingContactID)

					return
				}

				id, err := strconv.Atoi(rid)
				if err != nil {
					log.Warn("Could not parse ID from contact row URL", "err", errInvalidContactID)

					handlePanic(errInvalidContactID)

					return
				}

				selectedContactID = id

				homeNavigation.PushByTag(resources.PageContactsView)
			}
		})

		contactsViewActivitiesListBox.ConnectRowSelected(func(row *gtk.ListBoxRow) {
			if row != nil {
				u, err := url.Parse(row.Cast().(*adw.ActionRow).Name())
				if err != nil {
					log.Warn("Could not parse activity row URL", "err", err)

					handlePanic(err)

					return
				}

				rid := u.Query().Get("id")
				if strings.TrimSpace(rid) == "" {
					log.Warn("Could not get ID from activity row URL", "err", errMissingActivityID)

					handlePanic(errMissingActivityID)

					return
				}

				id, err := strconv.Atoi(rid)
				if err != nil {
					log.Warn("Could not parse ID from activity row URL", "err", errInvalidActivityID)

					handlePanic(errInvalidActivityID)

					return
				}

				selectedActivityID = id

				homeNavigation.PushByTag(resources.PageActivitiesView)
			}
		})

		homeNavigation.ConnectPopped(func(page *adw.NavigationPage) {
			handleHomeNavigation()

			var (
				tag = page.Tag()
				log = log.With("tag", tag)
			)

			log.Info("Handling popped page")

			switch tag {
			case resources.PageContactsView:
				contactsViewPageTitle.SetTitle("")
				contactsViewPageTitle.SetSubtitle("")

			case resources.PageActivitiesView:
				activitiesViewPageTitle.SetTitle("")
				activitiesViewPageTitle.SetSubtitle("")
			}
		})
		homeNavigation.ConnectPushed(handleHomeNavigation)
		homeNavigation.ConnectReplaced(handleHomeNavigation)

		homeSidebarListbox.ConnectRowSelected(func(row *gtk.ListBoxRow) {
			homeNavigation.ReplaceWithTags([]string{row.Cast().(*adw.ActionRow).Name()})
		})

		handleNavigation := func() {
			var (
				tag = nv.VisiblePage().Tag()
				log = log.With("tag", tag)
			)

			log.Info("Handling page")

			switch tag {
			case resources.PageIndex:
				go func() {
					if err := checkSenbaraServerConfiguration(); err != nil {
						log.Info("Could not check Senbara server configuration, redirecting to login", "err", err)

						updateDeregisterClientActionEnabled()

						nv.PushByTag(resources.PageWelcome)

						return
					}

					if err := setupAuthn(false); err != nil {
						log.Info("Could not setup authn, redirecting to login", "err", err)

						nv.PushByTag(resources.PageWelcome)

						return
					}

					redirected, _, _, err := authorize(
						ctx,

						false,
					)
					if err != nil {
						log.Warn("Could not authorize user for index page", "err", err)

						handlePanic(err)

						return
					} else if redirected {
						return
					}

					if settings.Boolean(resources.SettingAnonymousMode) {
						nv.PushByTag(resources.PagePreview)

						return
					}

					if strings.TrimSpace(u.Email) != "" {
						nv.PushByTag(resources.PageHome)

						return
					}

					nv.PushByTag(resources.PageWelcome)
				}()

			case resources.PageConfigServerURL:
				configServerURLContinueButton.SetSensitive(false)
				configServerURLContinueSpinner.SetVisible(true)

				go func() {
					defer configServerURLContinueSpinner.SetVisible(false)

					if err := deregisterOIDCClient(); err != nil {
						handlePanic(err)

						return
					}

					updateDeregisterClientActionEnabled()
					updateConfigServerURLContinueButtonSensitive()
				}()

			case resources.PagePreview:
				go func() {
					redirected, c, _, err := authorize(
						ctx,

						false,
					)
					if err != nil {
						log.Warn("Could not authorize user for home page", "err", err)

						handlePanic(err)

						return
					} else if redirected {
						return
					}

					settings.SetBoolean(resources.SettingAnonymousMode, true)

					log.Debug("Getting OpenAPI spec")

					res, err := c.GetOpenAPISpec(ctx)
					if err != nil {
						handlePanic(err)

						return
					}
					defer res.Body.Close()

					log.Debug("Got OpenAPI spec", "status", res.StatusCode)

					if res.StatusCode != http.StatusOK {
						handlePanic(errors.New(res.Status))

						return
					}

					log.Debug("Writing OpenAPI spec to stdout")

					if _, err := io.Copy(os.Stdout, res.Body); err != nil {
						handlePanic(err)

						return
					}
				}()

			case resources.PageRegister:
				configInitialAccessTokenInput.SetText("")

			case resources.PageHome:
				go func() {
					enableHomeSidebarLoading()
					defer disableHomeSidebarLoading()

					redirected, c, _, err := authorize(
						ctx,

						true,
					)
					if err != nil {
						log.Warn("Could not authorize user for home page", "err", err)

						handlePanic(err)

						return
					} else if redirected {
						return
					}

					settings.SetBoolean(resources.SettingAnonymousMode, false)

					homeSidebarListbox.SelectRow(homeSidebarListbox.RowAtIndex(0))

					log.Debug("Getting summary")

					res, err := c.GetIndexWithResponse(ctx)
					if err != nil {
						handlePanic(err)

						return
					}

					log.Debug("Got summary", "status", res.StatusCode())

					if res.StatusCode() != http.StatusOK {
						handlePanic(errors.New(res.Status()))

						return
					}

					homeSidebarContactsCountLabel.SetText(fmt.Sprintf("%v", *res.JSON200.ContactsCount))
					homeSidebarJournalEntriesCountLabel.SetText(fmt.Sprintf("%v", *res.JSON200.JournalEntriesCount))
				}()
			}
		}

		nv.ConnectPopped(func(page *adw.NavigationPage) {
			handleNavigation()
		})
		nv.ConnectPushed(handleNavigation)
		nv.ConnectReplaced(handleNavigation)

		handleNavigation()

		a.AddWindow(&w.Window)
	})

	a.ConnectOpen(func(files []gio.Filer, hint string) {
		if w == nil {
			a.Activate()
		} else {
			w.Present()
		}

		for _, r := range files {
			u, err := url.Parse(r.URI())
			if err != nil {
				handlePanic(err)

				return
			}

			authCode := u.Query().Get("code")
			state := u.Query().Get("state")

			log := log.With(
				"authCode", authCode != "",
				"state", state,
			)

			log.Debug("Handling user auth exchange")

			var (
				stateNonce,
				pkceCodeVerifier,
				oidcNonce string
			)
			sn, err := keyring.Get(resources.AppID, resources.SecretStateNonceKey)
			if err != nil {
				if !errors.Is(err, keyring.ErrNotFound) {
					log.Debug("Failed to read state nonce cookie", "error", err)

					handlePanic(errors.Join(errCouldNotLogin, err))

					return
				}
			} else {
				stateNonce = sn
			}

			pcv, err := keyring.Get(resources.AppID, resources.SecretPKCECodeVerifierKey)
			if err != nil {
				if !errors.Is(err, keyring.ErrNotFound) {
					log.Debug("Failed to read PKCE code verifier cookie", "error", err)

					handlePanic(errors.Join(errCouldNotLogin, err))

					return
				}
			} else {
				pkceCodeVerifier = pcv
			}

			on, err := keyring.Get(resources.AppID, resources.SecretOIDCNonceKey)
			if err != nil {
				if !errors.Is(err, keyring.ErrNotFound) {
					log.Debug("Failed to read OIDC nonce cookie", "error", err)

					handlePanic(errors.Join(errCouldNotLogin, err))

					return
				}
			} else {
				oidcNonce = on
			}

			nextURL, signedOut, err := authner.Exchange(
				ctx,

				authCode,
				state,

				stateNonce,
				pkceCodeVerifier,
				oidcNonce,

				func(s string, t time.Time) error {
					// TODO: Handle expiry time
					return keyring.Set(resources.AppID, resources.SecretRefreshTokenKey, s)
				},
				func(s string, t time.Time) error {
					// TODO: Handle expiry time
					return keyring.Set(resources.AppID, resources.SecretIDTokenKey, s)
				},

				func() error {
					if err := keyring.Delete(resources.AppID, resources.SecretRefreshTokenKey); err != nil && !errors.Is(err, keyring.ErrNotFound) {
						return err
					}

					return nil
				},
				func() error {
					if err := keyring.Delete(resources.AppID, resources.SecretIDTokenKey); err != nil && !errors.Is(err, keyring.ErrNotFound) {
						return err
					}

					return nil
				},

				func() error {
					if err := keyring.Delete(resources.AppID, resources.SecretStateNonceKey); err != nil && !errors.Is(err, keyring.ErrNotFound) {
						return err
					}

					return nil
				},
				func() error {
					if err := keyring.Delete(resources.AppID, resources.SecretPKCECodeVerifierKey); err != nil && !errors.Is(err, keyring.ErrNotFound) {
						return err
					}

					return nil
				},
				func() error {
					if err := keyring.Delete(resources.AppID, resources.SecretOIDCNonceKey); err != nil && !errors.Is(err, keyring.ErrNotFound) {
						return err
					}

					return nil
				},
			)
			if err != nil {
				handlePanic(err)

				return
			}

			// In the web version, redirecting to the home page after signing out is possible without
			// authn. In the GNOME version, that is not the case since the unauthenticated
			// page is a separate page from home, so we need to rewrite the path to distinguish
			// between the two manually
			if signedOut && nextURL == resources.PageHome {
				nextURL = resources.PageIndex
			}

			nv.ReplaceWithTags([]string{nextURL})
		}
	})

	if code := a.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}
