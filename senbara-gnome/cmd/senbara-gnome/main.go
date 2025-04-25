package main

import (
	"context"
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
	"github.com/pojntfx/senbara/senbara-common/pkg/authn"
	"github.com/pojntfx/senbara/senbara-gnome/assets/resources"
	"github.com/pojntfx/senbara/senbara-gnome/config/locales"
	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
	"github.com/rymdport/portal/openuri"
	"github.com/zalando/go-keyring"
	"gopkg.in/yaml.v2"
)

var (
	errCouldNotLogin = errors.New("could not login")
)

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
		authner *authn.Authner
	)

	authorize := func(
		ctx context.Context,

		nv *adw.NavigationView,

		privacyPolicyConsentGiven bool,

		loginIfSignedOut bool,
	) (
		redirected bool,

		u userData,
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

				return false, userData{}, http.StatusUnauthorized, errors.Join(errCouldNotLogin, err)
			}
		} else {
			refreshToken = &rt
		}

		it, err := keyring.Get(resources.AppID, resources.SecretIDTokenKey)
		if err != nil {
			if !errors.Is(err, keyring.ErrNotFound) {
				log.Debug("Failed to read ID token cookie", "error", err)

				return false, userData{}, http.StatusUnauthorized, errors.Join(errCouldNotLogin, err)
			}
		} else {
			idToken = &it
		}

		nextURL, requirePrivacyConsent, email, logoutURL, err := authner.Authorize(
			ctx,

			loginIfSignedOut,

			nv.VisiblePage().Tag(),
			nv.VisiblePage().Tag(),

			privacyPolicyConsentGiven,

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
				return false, userData{}, http.StatusUnauthorized, err
			}

			return false, userData{}, http.StatusInternalServerError, err
		}

		redirected = nextURL != ""
		u = userData{
			Email:     email,
			LogoutURL: logoutURL,
		}

		if redirected {
			nv.PushByTag("exchange")

			if err := openuri.OpenURI("", nextURL, nil); err != nil {
				panic(err)
			}

			return redirected, u, http.StatusTemporaryRedirect, nil
		}

		if requirePrivacyConsent {
			nv.PushByTag("privacy-policy")

			log.Debug("Refresh token cookie is missing, but can't reauthenticate with auth provider since privacy policy consent is not yet given. Redirecting to privacy policy consent page")

			return true, u, http.StatusTemporaryRedirect, nil
		}

		return redirected, u, http.StatusOK, nil
	}

	a.ConnectActivate(func() {
		gtk.StyleContextAddProviderForDisplay(
			gdk.DisplayGetDefault(),
			c,
			gtk.STYLE_PROVIDER_PRIORITY_APPLICATION,
		)

		b := gtk.NewBuilderFromResource(resources.ResourceWindowUIPath)

		var (
			w  = b.GetObject("main-window").Cast().(*adw.Window)
			nv = b.GetObject("main-navigation").Cast().(*adw.NavigationView)

			ppckb = b.GetObject("privacy-policy-checkbutton").Cast().(*gtk.CheckButton)

			lb = b.GetObject("login-button").Cast().(*gtk.Button)

			ssui  = b.GetObject("select-server-url-input").Cast().(*adw.EntryRow)
			ssuoi = b.GetObject("select-server-oidc-issuer-input").Cast().(*adw.EntryRow)
			ssuoc = b.GetObject("select-server-oidc-client-id-input").Cast().(*adw.EntryRow)
			sscb  = b.GetObject("select-server-continue-button").Cast().(*gtk.Button)
			sscs  = b.GetObject("select-server-continue-spinner").Cast().(*gtk.Widget)

			ppl = b.GetObject("privacy-policy-link").Cast().(*gtk.Label)
		)

		lb.ConnectClicked(func() {
			nv.PushByTag("select-server")
		})

		settings.Bind(resources.SettingServerURLKey, ssui.Object, "text", gio.SettingsBindDefault)
		settings.Bind(resources.SettingOIDCIssuerKey, ssuoi.Object, "text", gio.SettingsBindDefault)
		settings.Bind(resources.SettingOIDCClientIDKey, ssuoc.Object, "text", gio.SettingsBindDefault)

		updateSelectServerContinueButtonSensitive := func() {
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
				updateSelectServerContinueButtonSensitive()
			}
		})

		checkConfiguration := func() error {
			var (
				serverURL    = settings.String(resources.SettingServerURLKey)
				oidcIssuer   = settings.String(resources.SettingOIDCIssuerKey)
				oidcClientID = settings.String(resources.SettingOIDCClientIDKey)
			)

			authner = authn.NewAuthner(
				slog.New(log.Handler().WithGroup("authner")),

				oidcIssuer,
				oidcClientID,
				"senbara:///authorize",
			)

			if err := authner.Init(ctx); err != nil {
				return err
			}

			client, err := api.NewClientWithResponses(serverURL)
			if err != nil {
				return err
			}

			res, err := client.GetOpenAPISpec(ctx)
			if err != nil {
				return err
			}

			var spec *openapi3.T
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

		sscb.ConnectClicked(func() {
			sscb.SetSensitive(false)
			sscs.SetVisible(true)

			go func() {
				defer sscs.SetVisible(false)

				if err := checkConfiguration(); err != nil {
					panic(err)
				}

				nv.PushByTag("privacy-policy")
			}()
		})

		handleNavigation := func() {
			switch nv.VisiblePage().Tag() {
			case "loading-config":
				log.Info("Handling loading-config")

				if err := checkConfiguration(); err != nil {
					log.Info("Could not check configuration, redirecting to login", "err", err)

					nv.PushByTag("login")

					return
				}

				_, userData, _, err := authorize(
					ctx,

					nv,

					ppckb.Active(),

					false,
				)
				if err != nil {
					log.Warn("Could not authorize user for index page", "err", err)

					panic(err)
				}

				if strings.TrimSpace(userData.Email) != "" {
					nv.PushByTag("home")

					return
				}

				nv.PushByTag("login")
			case "select-server":
				log.Info("Handling select-server")

				updateSelectServerContinueButtonSensitive()
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
			_, err := url.Parse(r.URI())
			if err != nil {
				panic(err)
			}
		}
	})

	if code := a.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}
