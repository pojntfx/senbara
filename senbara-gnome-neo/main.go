package main

import (
	"os"
	"runtime"
	"unsafe"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gobject"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/pojntfx/senbara/senbara-gtk-go/senbaragtk"
)

var (
	gTypeExampleApplication gobject.Type
)

const (
	dataKeyGoInstance = "go_instance"
)

type exampleApplication struct {
	adw.Application
}

func init() {
	var classInit gobject.ClassInitFunc = func(tc *gobject.TypeClass, u uintptr) {
		objClass := (*gobject.ObjectClass)(unsafe.Pointer(tc))

		objClass.OverrideConstructed(func(o *gobject.Object) {
			parentObjClass := (*gobject.ObjectClass)(unsafe.Pointer(tc.PeekParent()))

			parentObjClass.GetConstructed()(o)

			var parent adw.Application
			o.Cast(&parent)

			w := &exampleApplication{
				Application: parent,
			}

			var pinner runtime.Pinner
			pinner.Pin(w)

			var cleanupCallback glib.DestroyNotify = func(data uintptr) {
				pinner.Unpin()
			}
			o.SetDataFull(dataKeyGoInstance, uintptr(unsafe.Pointer(w)), &cleanupCallback)
		})

		adwApplicationClass := (*gio.ApplicationClass)(unsafe.Pointer(tc))

		adwApplicationClass.OverrideActivate(func(a *gio.Application) {
			var app gtk.Application
			a.Cast(&app)

			obj := gobject.NewObject(senbaragtk.MainApplicationWindowGLibType(),
				"application", app,
			)

			var window senbaragtk.MainApplicationWindow
			obj.Cast(&window)

			window.SetApplication(&app)

			window.Present()
		})
	}

	var instanceInit gobject.InstanceInitFunc = func(ti *gobject.TypeInstance, tc *gobject.TypeClass) {}

	var parentQuery gobject.TypeQuery
	gobject.NewTypeQuery(adw.ApplicationGLibType(), &parentQuery)

	gTypeExampleApplication = gobject.TypeRegisterStaticSimple(
		parentQuery.Type,
		"ExampleApplication",
		parentQuery.ClassSize,
		&classInit,
		1024, // TODO: Calculate correct size here
		&instanceInit,
		0,
	)
}

func main() {
	obj := gobject.NewObject(gTypeExampleApplication,
		"application_id", "com.pojtinger.felicitas.SenbaraGnomeNeo", // TODO: Do this by overwriting the constructor above instead
		"flags", gio.GApplicationFlagsNoneValue,
	)

	var app exampleApplication
	obj.Cast(&app)

	os.Exit(app.Run(len(os.Args), os.Args))
}
