package main

import (
	"unsafe"

	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gobject"
	"github.com/jwijenbergh/puregotk/v4/gobject/types"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/pojntfx/senbara/libsenbara-gtk/assets/resources"
)

import "C"

var gTypeSenbaraGtkMainApplicationWindow gobject.Type

type senbaraGtkMainApplicationWindow struct {
	*gtk.ApplicationWindow
}

func newSenbaraGtkMainApplicationWindow() *senbaraGtkMainApplicationWindow {
	obj := gobject.NewObject(gTypeSenbaraGtkMainApplicationWindow, "application")

	parent := (*gtk.ApplicationWindow)(unsafe.Pointer(obj))
	parent.InitTemplate()

	return &senbaraGtkMainApplicationWindow{
		parent,
	}
}

//export senbara_gtk_main_application_window_get_type
func senbara_gtk_main_application_window_get_type() C.ulong {
	if gTypeSenbaraGtkMainApplicationWindow == 0 {
		senbara_gtk_init_types()
	}

	return C.ulong(gTypeSenbaraGtkMainApplicationWindow)
}

//export senbara_gtk_main_application_window_new
func senbara_gtk_main_application_window_new() unsafe.Pointer {
	window := newSenbaraGtkMainApplicationWindow()

	return unsafe.Pointer(window.ApplicationWindow.Ptr)
}

//export senbara_gtk_init_types
func senbara_gtk_init_types() {
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

		gobject.SignalNewv(
			"button-test-clicked",
			gTypeSenbaraGtkMainApplicationWindow,
			gobject.GSignalRunFirstValue,
			gobject.CclosureNew(&callbackFunc, 0, &destroyData),
			nil,
			0,
			nil,
			types.GType(gobject.TypeNoneVal),
			0,
			nil,
		)

	}

	var instanceInit gobject.InstanceInitFunc = func(ti *gobject.TypeInstance, tc *gobject.TypeClass) {
		typeClass := (*gtk.WidgetClass)(unsafe.Pointer(tc))
		typeClass.SetTemplateFromResource(resources.ResourceWindowUIPath)

		var callbackSymbol gobject.Callback = func() {
			gobject.SignalEmit(
				(*gobject.Object)(unsafe.Pointer(ti)),
				gobject.SignalLookup("button-test-clicked", gTypeSenbaraGtkMainApplicationWindow),
				0,
			)
		}

		typeClass.BindTemplateCallbackFull("on_button_test_clicked", &callbackSymbol)
	}

	gTypeSenbaraGtkMainApplicationWindow = gobject.TypeRegisterStaticSimple(
		gtk.ApplicationWindowGLibType(),
		"SenbaraGtkMainApplicationWindow",
		1024,
		&classInit,
		1024,
		&instanceInit,
		gobject.TypeNoneVal,
	)
}

func main() {}
