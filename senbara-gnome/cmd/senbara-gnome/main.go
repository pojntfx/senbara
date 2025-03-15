package main

import (
	"os"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/pojntfx/senbara/senbara-gnome/pkg/resources"
)

func main() {
	r, err := gio.NewResourceFromData(glib.NewBytesWithGo(resources.ResourceContents))
	if err != nil {
		panic(err)
	}
	gio.ResourcesRegister(r)

	a := adw.NewApplication(resources.AppID, gio.ApplicationNonUnique)
	a.ConnectActivate(func() {
		b := gtk.NewBuilderFromResource(resources.ResourceWindowPath)

		w := b.GetObject("main-window").Cast().(*adw.Window)

		a.AddWindow(&w.Window)
	})

	if code := a.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}
