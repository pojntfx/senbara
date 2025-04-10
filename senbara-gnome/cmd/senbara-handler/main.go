package main

import (
	"context"
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

		// TODO: Show "login" button if not already signed in/authenticated and redirect to browser if requested, else show "logout" button
		l = gtk.NewLabel("Unauthenticated")
		w.SetContent(l)

		a.AddWindow(&w.Window)
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
				l.SetText("Unauthenticated")

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
