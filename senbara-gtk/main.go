package main

import (
	"log"
	"os"
	"time"
	"unsafe"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gobject"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/pojntfx/senbara/libsenbara-gtk-go/v4/senbaragtk"
	"github.com/pojntfx/senbara/senbara-gtk/assets/resources"
)

func init() {
	resource, err := gio.NewResourceFromData(glib.NewBytes(resources.ResourceContents, uint(len(resources.ResourceContents))))
	if err != nil {
		panic(err)
	}
	gio.ResourcesRegister(resource)
}

type ExampleApplication struct {
	*adw.Application
	window *senbaragtk.MainApplicationWindow
}

func NewExampleApplication() *ExampleApplication {
	app := adw.NewApplication("com.pojtinger.felicitas.senbaragtk.Example", gio.GApplicationFlagsNoneValue)

	exampleApp := &ExampleApplication{
		Application: app,
	}

	activateCallback := func(gio.Application) {
		exampleApp.onActivate()
	}
	app.ConnectActivate(&activateCallback)

	return exampleApp
}

func (app *ExampleApplication) onActivate() {
	b := gtk.NewBuilderFromResource(resources.ResourceWindowUIPath)

	app.window = (*senbaragtk.MainApplicationWindow)(unsafe.Pointer(b.GetObject("main_window")))
	app.window.SetApplication(&app.Application.Application)

	cb := func(senbaragtk.MainApplicationWindow) {
		log.Println("Test button clicked")

		app.window.ShowToast("Button was clicked!")

		var value gobject.Value
		value.Init(gobject.TypeBooleanVal)
		value.SetBoolean(false)
		app.window.SetProperty("test-button-sensitive", &value)
		value.Unset()

		go func() {
			time.Sleep(3 * time.Second)

			var idleFunc glib.SourceFunc = func(uintptr) bool {
				app.window.ShowToast("Button re-enabled after 3 seconds")

				var value gobject.Value
				value.Init(gobject.TypeBooleanVal)
				value.SetBoolean(true)
				app.window.SetProperty("test-button-sensitive", &value)
				value.Unset()

				return false
			}
			glib.IdleAdd(&idleFunc, 0)
		}()
	}

	app.window.ConnectButtonTestClicked(&cb)

	app.window.Present()
}

func main() {
	os.Exit(NewExampleApplication().Run(len(os.Args), os.Args))
}
