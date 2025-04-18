package main

import (
	"context"
	"errors"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	gcore "github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pojntfx/senbara/senbara-common/pkg/authn"
	"github.com/pojntfx/senbara/senbara-gnome/assets/resources"
	"github.com/pojntfx/senbara/senbara-gnome/config/locales"
	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
	"github.com/rymdport/portal/openuri"
	"github.com/zalando/go-keyring"
)

type linkTemplateData struct {
	Href  string
	Label string
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

	var w *adw.Window
	a.ConnectActivate(func() {
		b := gtk.NewBuilderFromResource(resources.ResourceWindowUIPath)

		w = b.GetObject("main-window").Cast().(*adw.Window)

		nv := b.GetObject("main-navigation").Cast().(*adw.NavigationView)

		lb := b.GetObject("login-button").Cast().(*gtk.Button)
		lb.ConnectClicked(func() {
			nv.PushByTag("select-server")
		})

		var (
			ssui  = b.GetObject("select-server-url-input").Cast().(*adw.EntryRow)
			ssuoi = b.GetObject("select-server-oidc-issuer-input").Cast().(*adw.EntryRow)
			ssuoc = b.GetObject("select-server-oidc-client-id-input").Cast().(*adw.EntryRow)
		)

		settings.Bind(resources.SettingServerURLKey, ssui.Object, "text", gio.SettingsBindDefault)
		settings.Bind(resources.SettingOIDCIssuerKey, ssuoi.Object, "text", gio.SettingsBindDefault)
		settings.Bind(resources.SettingOIDCClientIDKey, ssuoc.Object, "text", gio.SettingsBindDefault)

		sscb := b.GetObject("select-server-continue-button").Cast().(*gtk.Button)
		sscs := b.GetObject("select-server-continue-spinner").Cast().(*gtk.Widget)

		checkCanContinueSelectServer := func() {
			if len(settings.String(resources.SettingServerURLKey)) > 0 &&
				len(settings.String(resources.SettingOIDCIssuerKey)) > 0 &&
				len(settings.String(resources.SettingOIDCClientIDKey)) > 0 {
				sscb.SetSensitive(true)
			} else {
				sscb.SetSensitive(false)
			}
		}

		settings.ConnectChanged(func(key string) {
			if key == resources.SettingServerURLKey ||
				key == resources.SettingOIDCIssuerKey ||
				key == resources.SettingOIDCClientIDKey {
				checkCanContinueSelectServer()
			}
		})

		nv.ConnectPushed(func() {
			if nv.VisiblePage().Tag() == "select-server" {
				checkCanContinueSelectServer()
			}
		})

		nv.ConnectPopped(func(page *adw.NavigationPage) {
			if nv.VisiblePage().Tag() == "select-server" {
				checkCanContinueSelectServer()
			}
		})

		ppl := b.GetObject("privacy-policy-link").Cast().(*gtk.Label)

		sscb.ConnectClicked(func() {
			sscb.SetSensitive(false)
			sscs.SetVisible(true)

			go func() {
				defer sscs.SetVisible(false)

				var (
					serverURL    = settings.String(resources.SettingServerURLKey)
					oidcIssuer   = settings.String(resources.SettingOIDCIssuerKey)
					oidcClientID = settings.String(resources.SettingOIDCClientIDKey)
				)

				authner := authn.NewAuthner(
					slog.New(log.Handler().WithGroup("authner")),

					oidcIssuer,
					oidcClientID,
					"senbara:///authorize",
				)

				if err := authner.Init(ctx); err != nil {
					// TODO: Display error by marking entry fields as errored and with a toast
					panic(err)
				}

				client, err := api.NewClientWithResponses(serverURL)
				if err != nil {
					// TODO: Display error by marking entry fields as errored and with a toast
					panic(err)
				}

				res, err := client.GetOpenAPISpec(ctx)
				if err != nil {
					// TODO: Display error by marking entry fields as errored and with a toast
					panic(err)
				}

				var spec *openapi3.T
				if err := yaml.NewDecoder(res.Body).Decode(&spec); err != nil {
					// TODO: Display error by marking entry fields as errored and with a toast
					panic(err)
				}

				var ltsb strings.Builder
				if err := lt.Execute(&ltsb, linkTemplateData{
					Href:  spec.Info.TermsOfService,
					Label: gcore.Local("privacy policy"),
				}); err != nil {
					// TODO: Display error by marking entry fields as errored and with a toast
					panic(err)
				}

				ppl.SetLabel(ltsb.String())

				nv.PushByTag("privacy-policy")
			}()
		})

		ppckb := b.GetObject("privacy-policy-checkbutton").Cast().(*gtk.CheckButton)
		ppcb := b.GetObject("privacy-policy-continue-button").Cast().(*gtk.Button)

		ppckb.ConnectToggled(func() {
			ppcb.SetSensitive(ppckb.Active())
		})

		nv.ConnectPopped(func(page *adw.NavigationPage) {
			if page.Tag() == "privacy-policy" {
				ppckb.SetActive(false)
			}
		})

		ppcb.ConnectClicked(func() {
			nv.PushByTag("exchange")
		})

		nv.ConnectPushed(func() {
			if nv.VisiblePage().Tag() == "exchange" {
				if err := openuri.OpenURI("", "https://example.com/", nil); err != nil {
					panic(err)
				}

				time.AfterFunc(time.Second*2, func() {
					if err := keyring.Set(resources.AppID, resources.SecretRefreshTokenKey, "testvalue"); err != nil {
						panic(err)
					}

					if nv.VisiblePage().Tag() == "exchange" {
						nv.PushByTag("home")
					}
				})
			}
		})

		nv.ConnectPushed(func() {
			if nv.VisiblePage().Tag() == "exchange-logout" {
				if err := openuri.OpenURI("", "https://example.com/", nil); err != nil {
					panic(err)
				}

				time.AfterFunc(time.Second*2, func() {
					if err := keyring.Delete(resources.AppID, resources.SecretRefreshTokenKey); err != nil {
						panic(err)
					}

					if nv.VisiblePage().Tag() == "exchange-logout" {
						nv.PopToTag("loading-config")
					}
				})
			}
		})

		logoutAction := gio.NewSimpleAction("logout", nil)
		logoutAction.ConnectActivate(func(parameter *glib.Variant) {
			nv.PushByTag("exchange-logout")
		})
		a.AddAction(logoutAction)

		aboutAction := gio.NewSimpleAction("about", nil)
		aboutAction.ConnectActivate(func(parameter *glib.Variant) {
			log.Info("Showing about screen")
		})
		a.AddAction(aboutAction)

		gtk.StyleContextAddProviderForDisplay(
			gdk.DisplayGetDefault(),
			c,
			gtk.STYLE_PROVIDER_PRIORITY_APPLICATION,
		)

		hydrateFromConfig := func() {
			// TODO: Check if existing auth configuration works here, if not try to renew
			// ID token or sign in again. If no configuration is set or it is not recoverable,
			// clear configuration (although the library should do that automatically) and
			// continue to login page for setup.
			time.AfterFunc(time.Millisecond*500, func() {
				if _, err := keyring.Get(resources.AppID, resources.SecretRefreshTokenKey); err != nil {
					if errors.Is(err, keyring.ErrNotFound) {
						nv.PushByTag("login")
					} else {
						panic(err)
					}
				} else {
					nv.PushByTag("home")
				}
			})
		}

		nv.ConnectPushed(func() {
			if nv.VisiblePage().Tag() == "loading-config" {
				hydrateFromConfig()
			}
		})

		nv.ConnectPopped(func(page *adw.NavigationPage) {
			if nv.VisiblePage().Tag() == "loading-config" {
				hydrateFromConfig()
			}
		})

		hydrateFromConfig()

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

			log.Info("Handling URI", "uri", u)
		}
	})

	if code := a.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}
