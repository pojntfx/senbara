package main

import (
	"os"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func main() {
	a := adw.NewApplication("com.pojtinger.senbara.gnome", gio.ApplicationNonUnique)

	a.ConnectActivate(func() {
		w := adw.NewApplicationWindow(&a.Application)
		w.SetDefaultSize(960, 540)
		w.SetTitle("Senbara Forms")

		v := adw.NewToolbarView()

		h := adw.NewHeaderBar()
		h.SetShowTitle(false)

		b := gtk.NewMenuButton()
		b.SetIconName("open-menu-symbolic")
		b.SetPrimary(true)

		h.PackEnd(b)

		v.AddTopBar(h)

		p := adw.NewStatusPage()
		p.SetTitle("Senbara Forms")
		p.SetDescription("Simple personal ERP web application built with HTML forms, OpenID Connect authentication and PostgreSQL data storage. Designed as a reference for modern JS-free Web 2.0 development with Go.")
		p.SetIconName("open-book-symbolic")

		v.SetContent(p)

		w.SetContent(v)

		w.SetVisible(true)
	})

	if code := a.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}
