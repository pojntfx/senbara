package main

import (
	"unsafe"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gobject"
	"github.com/jwijenbergh/puregotk/v4/gobject/types"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/pojntfx/senbara/libsenbara-gtk/assets/resources"
)

import "C"

var gTypeSenbaraGtkMainApplicationWindow gobject.Type

//export senbara_gtk_main_application_window_get_type
func senbara_gtk_main_application_window_get_type() C.ulong {
	if gTypeSenbaraGtkMainApplicationWindow == 0 {
		senbara_gtk_init_types()
	}

	return C.ulong(gTypeSenbaraGtkMainApplicationWindow)
}

type senbaraGtkMainApplicationWindow struct {
	*adw.ApplicationWindow

	buttonTest *gtk.Button
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

		typeClass.BindTemplateChildFull("button_test", false, 0)
	}

	gTypeSenbaraGtkMainApplicationWindow = gobject.TypeRegisterStaticSimple(
		adw.ApplicationWindowGLibType(),
		"SenbaraGtkMainApplicationWindow",
		1024,
		&classInit,
		1024,
		&instanceInit,
		0,
	)
}

func newSenbaraGtkMainApplicationWindow() *senbaraGtkMainApplicationWindow {
	obj := gobject.NewObject(gTypeSenbaraGtkMainApplicationWindow, "application")

	parent := (*adw.ApplicationWindow)(unsafe.Pointer(obj))
	parent.InitTemplate()

	rawButtonTest := parent.Widget.GetTemplateChild(
		gTypeSenbaraGtkMainApplicationWindow,
		"button_test",
	)
	buttonTest := (*gtk.Button)(unsafe.Pointer(rawButtonTest))

	w := &senbaraGtkMainApplicationWindow{
		ApplicationWindow: parent,

		buttonTest: buttonTest,
	}

	cb := func(gtk.Button) {
		gobject.SignalEmit(
			obj,
			gobject.SignalLookup("button-test-clicked", gTypeSenbaraGtkMainApplicationWindow),
			0,
		)
	}

	buttonTest.ConnectClicked(&cb)

	var cleanupCallback glib.DestroyNotify = func(data uintptr) {
		obj.Unref()
	}
	obj.SetDataFull("go_instance", uintptr(unsafe.Pointer(w)), &cleanupCallback)

	return w
}

//export senbara_gtk_main_application_window_new
func senbara_gtk_main_application_window_new() unsafe.Pointer {
	window := newSenbaraGtkMainApplicationWindow()

	window.Object.Ref()

	return unsafe.Pointer(window.Object.Ptr)
}

func (w *senbaraGtkMainApplicationWindow) setTestButtonSensitive(sensitive bool) {
	w.buttonTest.SetSensitive(sensitive)
}

//export senbara_gtk_main_application_window_set_test_button_sensitive
func senbara_gtk_main_application_window_set_test_button_sensitive(window unsafe.Pointer, sensitive bool) {
	obj := gobject.ObjectNewFromInternalPtr(uintptr(window))

	goInstance := obj.GetData("go_instance")

	v := (*senbaraGtkMainApplicationWindow)(unsafe.Pointer(goInstance))

	v.setTestButtonSensitive(sensitive)
}

func main() {}
