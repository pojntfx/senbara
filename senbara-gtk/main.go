package main

import (
	"log"
	"os"
	"time"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gobject"
	"github.com/pojntfx/senbara/libsenbara-gtk-go/v4/senbaragtk"
)

func init() {
	senbaragtk.InitTypes()
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
	app.window = senbaragtk.NewMainApplicationWindow()
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
