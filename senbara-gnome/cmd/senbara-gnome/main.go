package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"log/slog"
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
	"gopkg.in/yaml.v3"
)

var (
	errCouldNotLogin            = errors.New("could not login")
	errCouldNotWriteSettingsKey = errors.New("could not write settings key")
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
			nv.ReplaceWithTags([]string{"exchange"})

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

	a.ConnectActivate(func() {
		gtk.StyleContextAddProviderForDisplay(
			gdk.DisplayGetDefault(),
			c,
			gtk.STYLE_PROVIDER_PRIORITY_APPLICATION,
		)

		b := gtk.NewBuilderFromResource(resources.ResourceWindowUIPath)

		w = b.GetObject("main-window").Cast().(*adw.Window)

		nv = b.GetObject("main-navigation").Cast().(*adw.NavigationView)

		var (
			sb = b.GetObject("setup-button").Cast().(*gtk.Button)
			ss = b.GetObject("setup-spinner").Cast().(*gtk.Widget)

			sssui = b.GetObject("select-senbara-server-url-input").Cast().(*adw.EntryRow)
			ssscb = b.GetObject("select-senbara-server-continue-button").Cast().(*gtk.Button)
			ssscs = b.GetObject("select-senbara-server-continue-spinner").Cast().(*gtk.Widget)

			spec openapi3.T

			plb = b.GetObject("preview-login-button").Cast().(*gtk.Button)
			pls = b.GetObject("preview-login-spinner").Cast().(*gtk.Widget)

			oidcDcrInitialAccessTokenPortalUrl string

			rb = b.GetObject("register-button").Cast().(*gtk.Button)

			siati  = b.GetObject("set-initial-access-token-input").Cast().(*adw.PasswordEntryRow)
			siatlb = b.GetObject("set-initial-access-token-login-button").Cast().(*gtk.Button)
			siatls = b.GetObject("set-initial-access-token-login-spinner").Cast().(*gtk.Widget)

			ecb  = b.GetObject("exchange-cancel-button").Cast().(*gtk.Button)
			elcb = b.GetObject("exchange-logout-cancel-button").Cast().(*gtk.Button)
		)

		sb.ConnectClicked(func() {
			nv.PushByTag("select-senbara-server")
		})

		settings.Bind(resources.SettingServerURLKey, sssui.Object, "text", gio.SettingsBindDefault)

		updateSelectSenbaraServerContinueButtonSensitive := func() {
			if len(settings.String(resources.SettingServerURLKey)) > 0 {
				ssscb.SetSensitive(true)
			} else {
				ssscb.SetSensitive(false)
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
			ssscb.SetSensitive(false)
			sb.SetSensitive(false)
			ssscs.SetVisible(true)
			ss.SetVisible(true)

			go func() {
				defer sb.SetSensitive(true)
				defer ssscs.SetVisible(false)
				defer ss.SetVisible(false)

				if err := deregisterOIDCClient(); err != nil {
					panic(err)
				}

				updateDeregisterClientActionEnabled()
				updateSelectSenbaraServerContinueButtonSensitive()
			}()
		})
		a.AddAction(deregisterClientAction)

		settings.ConnectChanged(func(key string) {
			if key == resources.SettingServerURLKey {
				ssscb.SetSensitive(false)
				ssscs.SetVisible(true)

				go func() {
					defer ssscs.SetVisible(false)

					if err := deregisterOIDCClient(); err != nil {
						panic(err)
					}

					updateDeregisterClientActionEnabled()
					updateSelectSenbaraServerContinueButtonSensitive()
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

				spec.Components.SecuritySchemes["oidc"].Value.OpenIdConnectUrl,
			)
			if err != nil {
				return err
			}

			oidcClientID := settings.String(resources.SettingOIDCClientIDKey)
			if oidcClientID == "" && registerClient {
				c, err := authn.RegisterOIDCClient(ctx, o, "Senbara GNOME", redirectURL)
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

		ssscb.ConnectClicked(func() {
			ssscb.SetSensitive(false)
			ssscs.SetVisible(true)

			go func() {
				defer ssscs.SetVisible(false)

				if err := checkSenbaraServerConfiguration(); err != nil {
					panic(err)
				}

				nv.PushByTag("preview")
			}()
		})

		plb.ConnectClicked(func() {
			plb.SetSensitive(false)
			pls.SetVisible(true)

			go func() {
				defer plb.SetSensitive(true)
				defer pls.SetVisible(false)

				if err := checkSenbaraServerConfiguration(); err != nil {
					panic(err)
				}

				if err := setupAuthn(true); err != nil {
					panic(err)
				}

				if v := spec.Components.SecuritySchemes["oidc"].Value.Extensions[api.OidcDcrInitialAccessTokenPortalUrlExtensionKey]; v != nil {
					vv, ok := v.(string)
					if ok {
						oidcDcrInitialAccessTokenPortalUrl = vv

						nv.PushByTag("register")

						return
					}
				}

				nv.PushByTag("home")
			}()
		})

		rb.ConnectClicked(func() {
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
					panic(err)
				}

				nv.PushByTag("set-initial-access-token")
			}()
		})

		siati.ConnectChanged(func() {
			if siati.TextLength() > 0 {
				siatlb.SetSensitive(true)
			} else {
				siatlb.SetSensitive(false)
			}
		})

		siatlb.ConnectClicked(func() {
			siatlb.SetSensitive(false)
			siatls.SetVisible(true)

			go func() {
				defer siatlb.SetSensitive(true)
				defer siatls.SetVisible(false)

				if err := checkSenbaraServerConfiguration(); err != nil {
					panic(err)
				}

				// TODO: Pass `siati.Text()` as the initial access token to the client registration endpoint
				if err := setupAuthn(true); err != nil {
					panic(err)
				}

				nv.PushByTag("home")
			}()
		})

		selectDifferentServerAction := gio.NewSimpleAction("selectDifferentServer", nil)
		selectDifferentServerAction.ConnectActivate(func(parameter *glib.Variant) {
			nv.ReplaceWithTags([]string{"welcome"})
		})
		a.AddAction(selectDifferentServerAction)

		ecb.ConnectClicked(func() {
			nv.ReplaceWithTags([]string{"welcome"})
		})

		elcb.ConnectClicked(func() {
			nv.ReplaceWithTags([]string{"home"})
		})

		logoutAction := gio.NewSimpleAction("logout", nil)
		logoutAction.ConnectActivate(func(parameter *glib.Variant) {
			nv.ReplaceWithTags([]string{"exchange-logout"})

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
					panic(err)
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
					panic(err)
				}
			}()
		})
		a.AddAction(licenseAction)

		privacyAction := gio.NewSimpleAction("privacy", nil)
		privacyAction.ConnectActivate(func(parameter *glib.Variant) {
			log.Info("Handling getting privacy action", "url", spec.Info.TermsOfService)

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
					panic(err)
				}
			}()
		})
		a.AddAction(privacyAction)

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
					panic(err)
				}
			}()
		})
		a.AddAction(imprintAction)

		codeAction := gio.NewSimpleAction("code", nil)
		codeAction.ConnectActivate(func(parameter *glib.Variant) {
			log.Info("Handling getting code action")

			redirected, c, _, err := authorize(
				ctx,

				false,
			)
			if err != nil {
				log.Warn("Could not authorize user for getting code action", "err", err)

				panic(err)
			} else if redirected {
				return
			}

			log.Debug("Getting code")

			res, err := c.GetSourceCode(ctx)
			if err != nil {
				panic(err)
			}
			defer res.Body.Close()

			log.Debug("Received code", "status", res.StatusCode)

			if res.StatusCode != http.StatusOK {
				panic(errors.New(res.Status))
			}

			log.Debug("Writing code to file")

			fd := gtk.NewFileDialog()
			fd.SetTitle(gcore.Local("Senbara REST source code"))
			fd.SetInitialName("code.tar.gz")
			fd.Save(ctx, &w.Window, func(r gio.AsyncResulter) {
				fp, err := fd.SaveFinish(r)
				if err != nil {
					panic(err)
				}

				log.Debug("Writing code to file", "path", fp.Path())

				f, err := os.OpenFile(fp.Path(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
				if err != nil {
					panic(err)
				}
				defer f.Close()

				if _, err := io.Copy(f, res.Body); err != nil {
					panic(err)
				}
			})
		})
		a.AddAction(codeAction)

		aboutAction := gio.NewSimpleAction("about", nil)
		aboutAction.ConnectActivate(func(parameter *glib.Variant) {
			log.Info("Opening about menu")
		})
		a.AddAction(aboutAction)

		handleNavigation := func() {
			switch nv.VisiblePage().Tag() {
			case "/":
				log.Info("Handling loading-config")

				if err := checkSenbaraServerConfiguration(); err != nil {
					log.Info("Could not check Senbara server configuration, redirecting to login", "err", err)

					updateDeregisterClientActionEnabled()

					nv.PushByTag("welcome")

					return
				}

				if err := setupAuthn(false); err != nil {
					log.Info("Could not setup authn, redirecting to login", "err", err)

					nv.PushByTag("welcome")

					return
				}

				redirected, _, _, err := authorize(
					ctx,

					false,
				)
				if err != nil {
					log.Warn("Could not authorize user for index page", "err", err)

					panic(err)
				} else if redirected {
					return
				}

				if settings.Boolean(resources.SettingAnonymousMode) {
					nv.PushByTag("preview")

					return
				}

				if strings.TrimSpace(u.Email) != "" {
					nv.PushByTag("home")

					return
				}

				nv.PushByTag("welcome")

			case "select-senbara-server":
				log.Info("Handling select-senbara-server")

				ssscb.SetSensitive(false)
				ssscs.SetVisible(true)

				go func() {
					defer ssscs.SetVisible(false)

					if err := deregisterOIDCClient(); err != nil {
						panic(err)
					}

					updateDeregisterClientActionEnabled()
					updateSelectSenbaraServerContinueButtonSensitive()
				}()

			case "preview":
				log.Info("Handling preview")

				redirected, c, _, err := authorize(
					ctx,

					false,
				)
				if err != nil {
					log.Warn("Could not authorize user for home page", "err", err)

					panic(err)
				} else if redirected {
					return
				}

				settings.SetBoolean(resources.SettingAnonymousMode, true)

				log.Debug("Getting OpenAPI spec")

				res, err := c.GetOpenAPISpec(ctx)
				if err != nil {
					panic(err)
				}
				defer res.Body.Close()

				log.Debug("Got OpenAPI spec", "status", res.StatusCode)

				if res.StatusCode != http.StatusOK {
					panic(errors.New(res.Status))
				}

				log.Debug("Writing OpenAPI spec to stdout")

				if _, err := io.Copy(os.Stdout, res.Body); err != nil {
					panic(err)
				}

			case "home":
				log.Info("Handling home")

				redirected, c, _, err := authorize(
					ctx,

					true,
				)
				if err != nil {
					log.Warn("Could not authorize user for home page", "err", err)

					panic(err)
				} else if redirected {
					return
				}

				settings.SetBoolean(resources.SettingAnonymousMode, false)

				log.Debug("Getting summary")

				res, err := c.GetIndexWithResponse(ctx)
				if err != nil {
					panic(err)
				}

				log.Debug("Got summary", "status", res.StatusCode())

				if res.StatusCode() != http.StatusOK {
					panic(errors.New(res.Status()))
				}

				log.Debug("Writing summary to stdout")

				if err := yaml.NewEncoder(os.Stdout).Encode(res.JSON200); err != nil {
					panic(err)
				}
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
				panic(err)
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

					panic(errors.Join(errCouldNotLogin, err))
				}
			} else {
				stateNonce = sn
			}

			pcv, err := keyring.Get(resources.AppID, resources.SecretPKCECodeVerifierKey)
			if err != nil {
				if !errors.Is(err, keyring.ErrNotFound) {
					log.Debug("Failed to read PKCE code verifier cookie", "error", err)

					panic(errors.Join(errCouldNotLogin, err))
				}
			} else {
				pkceCodeVerifier = pcv
			}

			on, err := keyring.Get(resources.AppID, resources.SecretOIDCNonceKey)
			if err != nil {
				if !errors.Is(err, keyring.ErrNotFound) {
					log.Debug("Failed to read OIDC nonce cookie", "error", err)

					panic(errors.Join(errCouldNotLogin, err))
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
				panic(err)
			}

			// In the web version, redirecting to the home page after signing out is possible without
			// authn. In the GNOME version, that is not the case since the unauthenticated
			// page is a separate page from home, so we need to rewrite the path to distinguish
			// between the two manually
			if signedOut && nextURL == "home" {
				nextURL = "/"
			}

			nv.ReplaceWithTags([]string{nextURL})
		}
	})

	if code := a.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}
