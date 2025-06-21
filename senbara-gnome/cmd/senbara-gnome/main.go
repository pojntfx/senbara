package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	gcore "github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
	"github.com/pojntfx/senbara/senbara-common/pkg/authn"
	"github.com/pojntfx/senbara/senbara-gnome/assets/resources"
	"github.com/pojntfx/senbara/senbara-gnome/config/locales"
	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
	"github.com/zalando/go-keyring"
)

var (
	errCouldNotLogin            = errors.New("could not login")
	errCouldNotWriteSettingsKey = errors.New("could not write settings key")
	errMissingPrivacyURL        = errors.New("missing privacy policy URL")
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

		w = b.GetObject("main-window").Cast().(*adw.Window)

		nv = b.GetObject("main-navigation").Cast().(*adw.NavigationView)

		mto = b.GetObject("main-toasts-overlay").Cast().(*adw.ToastOverlay)

		var (
			welcomeGetStartedButton  = b.GetObject("welcome-get-started-button").Cast().(*gtk.Button)
			welcomeGetStartedSpinner = b.GetObject("welcome-get-started-spinner").Cast().(*gtk.Widget)

			configServerURLInput           = b.GetObject("config-server-url-input").Cast().(*adw.EntryRow)
			configServerURLContinueButton  = b.GetObject("config-server-url-continue-button").Cast().(*gtk.Button)
			configServerURLContinueSpinner = b.GetObject("config-server-url-continue-spinner").Cast().(*gtk.Widget)

			spec openapi3.T

			previewLoginButton  = b.GetObject("preview-login-button").Cast().(*gtk.Button)
			previewLoginSpinner = b.GetObject("preview-login-spinner").Cast().(*gtk.Widget)

			oidcDcrInitialAccessTokenPortalUrl string

			registerRegisterButton = b.GetObject("register-register-button").Cast().(*gtk.Button)

			configInitialAccessTokenInput        = b.GetObject("config-initial-access-token-input").Cast().(*adw.PasswordEntryRow)
			configInitialAccessTokenLoginButton  = b.GetObject("config-initial-access-token-login-button").Cast().(*gtk.Button)
			configInitialAccessTokenLoginSpinner = b.GetObject("config-initial-access-token-login-spinner").Cast().(*gtk.Widget)

			exchangeLoginCancelButton  = b.GetObject("exchange-login-cancel-button").Cast().(*gtk.Button)
			exchangeLogoutCancelButton = b.GetObject("exchange-logout-cancel-button").Cast().(*gtk.Button)

			homeSplitView      = b.GetObject("home-split-view").Cast().(*adw.NavigationSplitView)
			homeNavigation     = b.GetObject("home-navigation").Cast().(*adw.NavigationView)
			homeSidebarListbox = b.GetObject("home-sidebar-listbox").Cast().(*gtk.ListBox)
			homeContentPage    = b.GetObject("home-content-page").Cast().(*adw.NavigationPage)

			homeUserMenuButton  = b.GetObject("home-user-menu-button").Cast().(*gtk.MenuButton)
			homeUserMenuAvatar  = b.GetObject("home-user-menu-avatar").Cast().(*adw.Avatar)
			homeUserMenuSpinner = b.GetObject("home-user-menu-spinner").Cast().(*gtk.Widget)

			homeHamburgerMenuButton  = b.GetObject("home-hamburger-menu-button").Cast().(*gtk.MenuButton)
			homeHamburgerMenuIcon    = b.GetObject("home-hamburger-menu-icon").Cast().(*gtk.Image)
			homeHamburgerMenuSpinner = b.GetObject("home-hamburger-menu-spinner").Cast().(*gtk.Widget)

			homeSidebarContactsCountLabel   = b.GetObject("home-sidebar-contacts-count-label").Cast().(*gtk.Label)
			homeSidebarContactsCountSpinner = b.GetObject("home-sidebar-contacts-count-spinner").Cast().(*gtk.Widget)

			homeSidebarJournalEntriesCountLabel   = b.GetObject("home-sidebar-journal-entries-count-label").Cast().(*gtk.Label)
			homeSidebarJournalEntriesCountSpinner = b.GetObject("home-sidebar-journal-entries-count-spinner").Cast().(*gtk.Widget)
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

		handleHomeNavigation := func() {
			var (
				tag = homeNavigation.VisiblePage().Tag()
				log = log.With("tag", tag)
			)

			log.Info("Handling")

			homeContentPage.SetTitle(homeNavigation.VisiblePage().Title())

			homeSplitView.SetShowContent(true)

			switch tag {
			case resources.PageContacts:
				// TODO: Load the contacts and then disable the loading state
			}
		}

		homeNavigation.ConnectPopped(func(page *adw.NavigationPage) {
			handleHomeNavigation()
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

			log.Info("Handling")

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
