package main

import (
	"log"
	"os"
	"unsafe"

	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gobject"
	"github.com/jwijenbergh/puregotk/v4/gobject/types"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/pojntfx/senbara/senbara-gnome-puregotk/assets/resources"
)

var (
	gTypeSenbaraPureGoTKMainWindow gobject.Type

	buttonTestClickedSignalID uint

	currentSenbaraPureGoTKMainWindow *senbaraPureGoTKMainWindow
)

type senbaraPureGoTKMainWindow struct {
	*gtk.ApplicationWindow
}

func newSenbaraPureGoTKMainWindow() *senbaraPureGoTKMainWindow {
	obj := gobject.NewObject(gTypeSenbaraPureGoTKMainWindow, "application")

	parent := (*gtk.ApplicationWindow)(unsafe.Pointer(obj))
	parent.InitTemplate()

	return &senbaraPureGoTKMainWindow{
		parent,
	}
}

func (w *senbaraPureGoTKMainWindow) emitButtonTestClicked() {
	obj := (*gobject.Object)(unsafe.Pointer(w.ApplicationWindow))
	gobject.SignalEmit(obj, buttonTestClickedSignalID, 0)
}

func (w *senbaraPureGoTKMainWindow) ConnectButtonTestClicked(cb func()) {
	w.ConnectSignal("button-test-clicked", &cb)
}

func init() {
	resource, err := gio.NewResourceFromData(glib.NewBytes(resources.ResourceContents, uint(len(resources.ResourceContents))))
	if err != nil {
		panic(err)
	}
	gio.ResourcesRegister(resource)

	var classInit gobject.ClassInitFunc = func(tc *gobject.TypeClass, u uintptr) {
		var (
			callbackFunc gobject.Callback      = func() {}
			destroyData  gobject.ClosureNotify = func(u uintptr, c *gobject.Closure) {}
		)

		buttonTestClickedSignalID = gobject.SignalNewv(
			"button-test-clicked",
			gTypeSenbaraPureGoTKMainWindow,
			gobject.GSignalRunFirstValue,
			gobject.CclosureNew(&callbackFunc, 0, &destroyData),
			nil,
			0,
			nil,
			types.GType(gobject.TypeNoneVal),
			0,
			nil,
		)

		typeClass := (*gtk.WidgetClass)(unsafe.Pointer(tc))
		typeClass.SetTemplateFromResource(resources.ResourceWindowUIPath)

		var callbackSymbol gobject.Callback = func() {
			log.Println("Callback on_button_test_clicked called")

			currentSenbaraPureGoTKMainWindow.emitButtonTestClicked()
		}

		typeClass.BindTemplateCallbackFull("on_button_test_clicked", &callbackSymbol)
	}

	var instanceInit gobject.InstanceInitFunc = func(ti *gobject.TypeInstance, tc *gobject.TypeClass) {}

	gTypeSenbaraPureGoTKMainWindow = gobject.TypeRegisterStaticSimple(
		gtk.ApplicationWindowGLibType(),
		"SenbaraPureGoTKMainWindow",
		1024,
		&classInit,
		1024,
		&instanceInit,
		gobject.TypeNoneVal,
	)
}

func main() {
	app := gtk.NewApplication(resources.AppID, gio.GApplicationFlagsNoneValue)
	defer app.Unref()

	activate := func(g gio.Application) {
		a := (*gtk.Application)(unsafe.Pointer(&g))

		window := newSenbaraPureGoTKMainWindow()
		currentSenbaraPureGoTKMainWindow = window

		window.ConnectButtonTestClicked(func() {
			log.Println("Signal button-test-clicked received")
		})

		window.SetApplication(a)
		window.Present()
	}

	app.ConnectActivate(&activate)

	if code := app.Run(len(os.Args), os.Args); code > 0 {
		os.Exit(code)
	}
}
