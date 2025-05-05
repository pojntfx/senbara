package main

import (
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
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
	"gopkg.in/yaml.v2"
)

var (
	errCouldNotLogin = errors.New("could not login")
)

type oidcConfig struct {
	Issuer string `json:"issuer"`
}

type oidcSpec struct {
	Info       openapi3.Info `yaml:"info"`
	Components struct {
		SecuritySchemes map[string]openapi3.SecurityScheme `yaml:"securitySchemes"`
	} `yaml:"components"`
}

type linkTemplateData struct {
	Href  string
	Label string
}

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

	lt, err := template.New("").Parse(`<a href="{{ .Href }}">{{ .Label }}</a>`)
	if err != nil {
		panic(err)
	}

	var (
		w       *adw.Window
		nv      *adw.NavigationView
		authner *authn.Authner
		ppckb   *gtk.CheckButton
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

		nextURL, requirePrivacyConsent, email, logoutURL, err := authner.Authorize(
			ctx,

			loginIfSignedOut,

			nv.VisiblePage().Tag(),
			nv.VisiblePage().Tag(),

			ppckb.Active(),

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

		if requirePrivacyConsent {
			nv.PushByTag("privacy-policy")

			log.Debug("Refresh token cookie is missing, but can't reauthenticate with auth provider since privacy policy consent is not yet given. Redirecting to privacy policy consent page")

			return true, nil, http.StatusTemporaryRedirect, nil
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

		ppckb = b.GetObject("privacy-policy-checkbutton").Cast().(*gtk.CheckButton)

		var (
			sb = b.GetObject("setup-button").Cast().(*gtk.Button)

			sssui = b.GetObject("select-senbara-server-url-input").Cast().(*adw.EntryRow)
			ssscb = b.GetObject("select-senbara-server-continue-button").Cast().(*gtk.Button)
			ssscs = b.GetObject("select-senbara-server-continue-spinner").Cast().(*gtk.Widget)

			ppl = b.GetObject("privacy-policy-link").Cast().(*gtk.Label)

			spec oidcSpec

			plb = b.GetObject("preview-login-button").Cast().(*gtk.Button)
			pls = b.GetObject("preview-login-spinner").Cast().(*gtk.Widget)

			ppcb = b.GetObject("privacy-policy-continue-button").Cast().(*gtk.Button)

			sasoc = b.GetObject("setup-authn-server-oidc-client-id-input").Cast().(*adw.EntryRow)
			sascb = b.GetObject("setup-authn-server-continue-button").Cast().(*gtk.Button)
			sascs = b.GetObject("setup-authn-server-continue-spinner").Cast().(*gtk.Widget)
		)

		sb.ConnectClicked(func() {
			nv.PushByTag("select-senbara-server")
		})

		settings.Bind(resources.SettingServerURLKey, sssui.Object, "text", gio.SettingsBindDefault)
		settings.Bind(resources.SettingOIDCClientIDKey, sasoc.Object, "text", gio.SettingsBindDefault)

		updateSelectSenbaraServerContinueButtonSensitive := func() {
			if len(settings.String(resources.SettingServerURLKey)) > 0 {
				ssscb.SetSensitive(true)
			} else {
				ssscb.SetSensitive(false)
			}
		}

		settings.ConnectChanged(func(key string) {
			if key == resources.SettingServerURLKey {
				updateSelectSenbaraServerContinueButtonSensitive()
			}
		})

		updateSelectAuthnServerContinueButtonSensitive := func() {
			if len(settings.String(resources.SettingOIDCClientIDKey)) > 0 {
				sascb.SetSensitive(true)
			} else {
				sascb.SetSensitive(false)
			}
		}

		settings.ConnectChanged(func(key string) {
			if key == resources.SettingOIDCClientIDKey {
				updateSelectAuthnServerContinueButtonSensitive()
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

			// We can't just use `*openapi3.T` here because the security schemes
			// can't be parsed with YAML, only with JSON (due to the go-jsonpointer requirement),
			// and  we can't switch to JSON since it can't be streaming encoded by the server
			if err := yaml.NewDecoder(res.Body).Decode(&spec); err != nil {
				return err
			}

			var ltsb strings.Builder
			if err := lt.Execute(&ltsb, linkTemplateData{
				Href:  spec.Info.TermsOfService,
				Label: gcore.Local("privacy policy"),
			}); err != nil {
				return err
			}

			ppl.SetLabel(ltsb.String())

			return nil
		}

		checkAuthnServerConfiguration := func() error {
			var (
				oidcClientID = settings.String(resources.SettingOIDCClientIDKey)
			)

			res, err := http.Get(spec.Components.SecuritySchemes["oidc"].OpenIdConnectUrl)
			if err != nil {
				return err
			}

			if res.StatusCode != http.StatusOK {
				return errors.New(res.Status)
			}

			var p oidcConfig
			if err := json.NewDecoder(res.Body).Decode(&p); err != nil {
				return err
			}

			authner = authn.NewAuthner(
				slog.New(log.Handler().WithGroup("authner")),

				p.Issuer,
				oidcClientID,
				"senbara:///authorize",
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

				nv.PushByTag("privacy-policy")
			}()
		})

		ppckb.ConnectToggled(func() {
			ppcb.SetSensitive(ppckb.Active())
		})

		ppcb.ConnectClicked(func() {
			nv.PushByTag("setup-authn-server")
		})

		sascb.ConnectClicked(func() {
			sascb.SetSensitive(false)
			sascs.SetVisible(true)

			go func() {
				defer sascs.SetVisible(false)

				if err := checkAuthnServerConfiguration(); err != nil {
					panic(err)
				}

				nv.PushByTag("home")
			}()
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

					nv.PushByTag("welcome")

					return
				}

				if err := checkAuthnServerConfiguration(); err != nil {
					log.Info("Could not check authn server configuration, redirecting to login", "err", err)

					nv.PushByTag("welcome")

					return
				}

				_, _, _, err := authorize(
					ctx,

					false,
				)
				if err != nil {
					log.Warn("Could not authorize user for index page", "err", err)

					panic(err)
				}

				if strings.TrimSpace(u.Email) != "" {
					nv.PushByTag("home")

					return
				}

				nv.PushByTag("welcome")

			case "select-senbara-server":
				log.Info("Handling select-senbara-server")

				ppckb.SetActive(false)

				updateSelectSenbaraServerContinueButtonSensitive()

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

				log.Debug("Getting OpenAPI spec")

				res, err := c.GetOpenAPISpec(ctx)
				if err != nil {
					panic(err)
				}

				log.Debug("Got OpenAPI spec", "status", res.StatusCode)

				if res.StatusCode != http.StatusOK {
					panic(errors.New(res.Status))
				}

				log.Debug("Writing OpenAPI spec to stdout")

				if _, err := io.Copy(os.Stdout, res.Body); err != nil {
					panic(err)
				}

			case "setup-authn-server":
				log.Info("Handling setup-authn-server")

				updateSelectAuthnServerContinueButtonSensitive()

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

			nextURL, signedOut, err := authner.Exchange(
				ctx,

				authCode,
				state,

				func(s string, t time.Time) error {
					// TODO: Handle expiry time
					return keyring.Set(resources.AppID, resources.SecretRefreshTokenKey, s)
				},
				func(s string, t time.Time) error {
					// TODO: Handle expiry time
					return keyring.Set(resources.AppID, resources.SecretIDTokenKey, s)
				},

				func() error {
					return keyring.Delete(resources.AppID, resources.SecretRefreshTokenKey)
				},
				func() error {
					return keyring.Delete(resources.AppID, resources.SecretIDTokenKey)
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
