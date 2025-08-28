package main

import (
	"unsafe"

	"github.com/jwijenbergh/purego"
	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gobject"
	"github.com/jwijenbergh/puregotk/v4/gobject/types"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/pojntfx/senbara/libsenbara-gtk/assets/resources"
)

import "C"

var (
	gTypeSenbaraGtkMainApplicationWindow gobject.Type
)

const (
	dataKeyGoInstance = "go_instance"

	propertyIdTestButtonSensitive = 1
)

//export senbara_gtk_main_application_window_get_type
func senbara_gtk_main_application_window_get_type() C.ulong {
	if gTypeSenbaraGtkMainApplicationWindow == 0 {
		senbara_gtk_init_types()
	}

	return C.ulong(gTypeSenbaraGtkMainApplicationWindow)
}

type senbaraGtkMainApplicationWindow struct {
	*adw.ApplicationWindow

	buttonTest   *gtk.Button
	toastOverlay *adw.ToastOverlay
}

//export senbara_gtk_init_types
func senbara_gtk_init_types() {
	resource, err := gio.NewResourceFromData(glib.NewBytes(resources.ResourceContents, uint(len(resources.ResourceContents))))
	if err != nil {
		panic(err)
	}
	gio.ResourcesRegister(resource)

	var classInit gobject.ClassInitFunc = func(tc *gobject.TypeClass, u uintptr) {
		widgetClass := (*gtk.WidgetClass)(unsafe.Pointer(tc))
		widgetClass.SetTemplateFromResource(resources.ResourceWindowUIPath)
		widgetClass.BindTemplateChildFull("button_test", false, 0)
		widgetClass.BindTemplateChildFull("toast_overlay", false, 0)

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

		objClass := (*gobject.ObjectClass)(unsafe.Pointer(tc))

		objClass.SetProperty = purego.NewCallback(func(obj uintptr, propId uint, value uintptr, pspec uintptr) {
			switch propId {
			case propertyIdTestButtonSensitive:
				var (
					v = (*gobject.Value)(unsafe.Pointer(value))
					w = (*senbaraGtkMainApplicationWindow)(unsafe.Pointer(gobject.ObjectNewFromInternalPtr(obj).GetData(dataKeyGoInstance)))
				)

				w.buttonTest.SetSensitive(v.GetBoolean())
			}
		})
		objClass.GetProperty = purego.NewCallback(func(obj uintptr, propId uint, value uintptr, pspec uintptr) {
			switch propId {
			case propertyIdTestButtonSensitive:
				var (
					v = (*gobject.Value)(unsafe.Pointer(value))
					w = (*senbaraGtkMainApplicationWindow)(unsafe.Pointer(gobject.ObjectNewFromInternalPtr(obj).GetData(dataKeyGoInstance)))
				)

				v.SetBoolean(w.buttonTest.IsSensitive())
			}
		})

		pspec := gobject.NewParamSpecBoolean(
			"test-button-sensitive",
			"Test Button Sensitive",
			"Whether the test button is sensitive",
			true,
			gobject.GParamReadwriteValue,
		)

		objClass.InstallProperty(propertyIdTestButtonSensitive, pspec)
	}

	var instanceInit gobject.InstanceInitFunc = func(ti *gobject.TypeInstance, tc *gobject.TypeClass) {}

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

//export senbara_gtk_main_application_window_new
func senbara_gtk_main_application_window_new() unsafe.Pointer {
	obj := gobject.NewObject(gTypeSenbaraGtkMainApplicationWindow, "application")

	ensureInstanceData(obj)

	obj.Ref()

	return unsafe.Pointer(obj.Ptr)
}

func ensureInstanceData(gobj *gobject.Object) *senbaraGtkMainApplicationWindow {
	if data := gobj.GetData(dataKeyGoInstance); data != 0 {
		return (*senbaraGtkMainApplicationWindow)(unsafe.Pointer(data))
	}

	parent := (*adw.ApplicationWindow)(unsafe.Pointer(gobj))
	parent.InitTemplate()

	rawButtonTest := parent.Widget.GetTemplateChild(
		gTypeSenbaraGtkMainApplicationWindow,
		"button_test",
	)
	buttonTest := (*gtk.Button)(unsafe.Pointer(rawButtonTest))

	rawToastOverlay := parent.Widget.GetTemplateChild(
		gTypeSenbaraGtkMainApplicationWindow,
		"toast_overlay",
	)
	toastOverlay := (*adw.ToastOverlay)(unsafe.Pointer(rawToastOverlay))

	w := &senbaraGtkMainApplicationWindow{
		ApplicationWindow: parent,
		buttonTest:        buttonTest,
		toastOverlay:      toastOverlay,
	}

	var cleanupCallback glib.DestroyNotify = func(data uintptr) {
		gobj.Unref()
	}
	gobj.SetDataFull(dataKeyGoInstance, uintptr(unsafe.Pointer(w)), &cleanupCallback)

	// TODO: Fix this; while it does read the property default value correctly, we get `g_object_ref_sink: assertion 'G_IS_OBJECT (object)' failed`
	// typeClass := gobject.TypeClassRef(gTypeSenbaraGtkMainApplicationWindow)
	// objClass := (*gobject.ObjectClass)(unsafe.Pointer(typeClass))
	// pspec := objClass.FindProperty("test-button-sensitive").Ref()
	// buttonTest.SetSensitive(pspec.GetDefaultValue().GetBoolean())

	cb := func(gtk.Button) {
		gobject.SignalEmit(
			gobj,
			gobject.SignalLookup("button-test-clicked", gTypeSenbaraGtkMainApplicationWindow),
			0,
		)
	}

	buttonTest.ConnectClicked(&cb)

	return w
}

func (w *senbaraGtkMainApplicationWindow) showToast(message string) {
	w.toastOverlay.AddToast(adw.NewToast(message))
}

//export senbara_gtk_main_application_window_show_toast
func senbara_gtk_main_application_window_show_toast(window unsafe.Pointer, message *C.char) {
	w := (*senbaraGtkMainApplicationWindow)(unsafe.Pointer(gobject.ObjectNewFromInternalPtr(uintptr(window)).GetData(dataKeyGoInstance)))

	w.showToast(C.GoString(message))
}

func main() {}
