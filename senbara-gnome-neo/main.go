package main

import (
	"fmt"
	"os"
	"runtime"
	"time"
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
	window *senbaragtk.MainApplicationWindow
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

		applicationClass := (*gio.ApplicationClass)(unsafe.Pointer(tc))

		applicationClass.OverrideActivate(func(a *gio.Application) {
			exampleApp := (*exampleApplication)(unsafe.Pointer(a.GetData(dataKeyGoInstance)))

			var app gtk.Application
			a.Cast(&app)

			obj := gobject.NewObject(senbaragtk.MainApplicationWindowGLibType(),
				"application", app,
			)

			var window senbaragtk.MainApplicationWindow
			obj.Cast(&window)

			exampleApp.window = &window

			var cb func(senbaragtk.MainApplicationWindow) = func(w senbaragtk.MainApplicationWindow) {
				fmt.Println("Test button clicked")

				exampleApp.window.ShowToast("Button was clicked!")

				var v gobject.Value
				v.Init(gobject.TypeBooleanVal)
				v.SetBoolean(false)
				exampleApp.window.SetProperty("test-button-sensitive", &v)

				time.AfterFunc(time.Second*3, func() {
					exampleApp.window.ShowToast("Button re-enabled after 3 seconds")

					var v gobject.Value
					v.Init(gobject.TypeBooleanVal)
					v.SetBoolean(true)
					exampleApp.window.SetProperty("test-button-sensitive", &v)
				})
			}
			exampleApp.window.ConnectButtonTestClicked(&cb)

			exampleApp.window.Present()
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
		parentQuery.InstanceSize+uint(unsafe.Sizeof(exampleApplication{}))+uint(unsafe.Sizeof(&exampleApplication{}))+uint(unsafe.Sizeof(&senbaragtk.MainApplicationWindow{})),
		&instanceInit,
		0,
	)
}

func main() {
	obj := gobject.NewObject(gTypeExampleApplication,
		"application_id", "com.pojtinger.felicitas.SenbaraGnomeNeo",
		"flags", gio.GApplicationFlagsNoneValue,
	)

	var app exampleApplication
	obj.Cast(&app)

	os.Exit(app.Run(len(os.Args), os.Args))
}
