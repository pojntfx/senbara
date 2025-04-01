package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func main() {
	a := adw.NewApplication("com.pojtinger.felicitas.SenbaraHandler", gio.ApplicationHandlesOpen)

	var (
		w *adw.Window
		l *gtk.Label
	)
	a.ConnectActivate(func() {
		w = adw.NewWindow()
		w.SetVisible(true)

		l = gtk.NewLabel("No auth code or state set")
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

			l.SetText(fmt.Sprintf(`Auth code: %v, state: %v`, authCode, state))
		}
	})

	if code := a.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}
