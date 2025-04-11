package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"time"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/pojntfx/senbara/senbara-common/pkg/authn"
	"github.com/pojntfx/senbara/senbara-gnome/assets/resources"
	"github.com/rymdport/portal/openuri"
	"github.com/zalando/go-keyring"
)

const (
	idTokenKey      = "id_token"
	refreshTokenKey = "refresh_token"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := &slog.HandlerOptions{}
	// TODO: Get verbosity level from GSettings
	if true {
		opts.Level = slog.LevelDebug
	}
	log := slog.New(slog.NewJSONHandler(os.Stderr, opts))

	authner := authn.NewAuthner(
		slog.New(log.Handler().WithGroup("authner")),

		// TODO: Read from GSettings
		"https://dev-4op4cmts68nqcenb.us.auth0.com/",
		"SiNcjPaYVYCOzeVvQAYl4mqhglsSWNY4",
		"senbara:///authorize",
	)

	if err := authner.Init(ctx); err != nil {
		panic(err)
	}

	a := adw.NewApplication("com.pojtinger.felicitas.SenbaraHandler", gio.ApplicationHandlesOpen)

	var (
		w *adw.Window
		l *gtk.Label
	)
	a.ConnectActivate(func() {
		w = adw.NewWindow()
		w.SetVisible(true)

		b := gtk.NewBox(gtk.OrientationVertical, 4)

		l = gtk.NewLabel("Home")
		b.Append(l)

		w.SetContent(b)

		a.AddWindow(&w.Window)

		var (
			refreshToken,
			idToken *string
		)
		rt, err := keyring.Get(resources.AppID, refreshTokenKey)
		if err != nil {
			if !errors.Is(err, keyring.ErrNotFound) {
				panic(err)
			}
		} else {
			refreshToken = &rt
		}

		it, err := keyring.Get(resources.AppID, idTokenKey)
		if err != nil {
			if !errors.Is(err, keyring.ErrNotFound) {
				panic(err)
			}
		} else {
			idToken = &it
		}

		nextURL, requirePrivacyConsent, _, logoutURL, err := authner.Authorize(
			ctx,

			false,
			"/",
			"/",

			false,

			refreshToken,
			idToken,

			func(s string, t time.Time) error {
				// TODO: Handle expiry time
				return keyring.Set(resources.AppID, refreshTokenKey, s)
			},
			func(s string, t time.Time) error {
				// TODO: Handle expiry time
				return keyring.Set(resources.AppID, idTokenKey, s)
			},
		)
		if err != nil {
			panic(err)
		}

		redirected := nextURL != ""
		if redirected {
			if err := openuri.OpenURI("", nextURL, nil); err != nil {
				panic(err)
			}

			return
		}

		if requirePrivacyConsent {
			// TODO: Implement privacy consent page
		}

		if logoutURL != "" {
			rt, err := keyring.Get(resources.AppID, refreshTokenKey)
			if err != nil {
				panic(err)
			}

			it, err := keyring.Get(resources.AppID, idTokenKey)
			if err != nil {
				panic(err)
			}

			l.SetText(fmt.Sprintf(`Refresh token: %v, ID token: %v`, rt, it))

			// TODO: Add logout button

			// TODO: Navigate to internal "next page"/nextURL
		} else {
			bt := gtk.NewButton()
			bt.SetLabel("Login")
			bt.ConnectClicked(func() {
				// TODO: Implement behavior from https://github.com/pojntfx/donna/blob/20af86f20378b5810258395652bbf57d71e2d184/senbara-forms/pkg/controllers/authn.go#L190-L207

				if err := openuri.OpenURI("http://localhost:1337/login", nextURL, nil); err != nil {
					panic(err)
				}
			})

			b.Append(bt)
		}
	})

	a.ConnectOpen(func(files []gio.Filer, hint string) {
		if w == nil || l == nil {
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

			_, signedOut, err := authner.Exchange(
				ctx,

				authCode,
				state,

				func(s string, t time.Time) error {
					// TODO: Handle expiry time
					return keyring.Set(resources.AppID, refreshTokenKey, s)
				},
				func(s string, t time.Time) error {
					// TODO: Handle expiry time
					return keyring.Set(resources.AppID, idTokenKey, s)
				},

				func() error {
					return keyring.Delete(resources.AppID, refreshTokenKey)
				},
				func() error {
					return keyring.Delete(resources.AppID, idTokenKey)
				},
			)
			if err != nil {
				panic(err)
			}

			if signedOut {
				l.SetText("Home")

				// TODO: Navigate to internal "start page"/nextURL

				return
			}

			rt, err := keyring.Get(resources.AppID, refreshTokenKey)
			if err != nil {
				panic(err)
			}

			it, err := keyring.Get(resources.AppID, idTokenKey)
			if err != nil {
				panic(err)
			}

			l.SetText(fmt.Sprintf(`Refresh token: %v, ID token: %v`, rt, it))

			// TODO: Navigate to internal "next page"/nextURL
		}
	})

	if code := a.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}
