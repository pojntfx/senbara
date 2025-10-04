package main

import (
	"path"
	"runtime"
	"unsafe"

	_ "embed"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gobject"
	"github.com/jwijenbergh/puregotk/v4/gobject/types"
	"github.com/jwijenbergh/puregotk/v4/gtk"
)

/*
#cgo pkg-config: glib-2.0
#include <locale.h>
#include <glib/gi18n.h>
*/
import "C"

//go:embed senbara-gtk.gresource
var ResourceContents []byte

var (
	gTypeSenbaraGtkMainApplicationWindow gobject.Type

	appPath = path.Join("/com", "pojtinger", "felicitas", "SenbaraGtk")

	resourceWindowUIPath = path.Join(appPath, "senbara-gtk-window.ui")
)

const (
	dataKeyGoInstance = "go_instance"

	propertyIdTestButtonSensitive = 1
)

//export senbara_gtk_main_application_window_get_type
func senbara_gtk_main_application_window_get_type() C.ulong {
	return C.ulong(gTypeSenbaraGtkMainApplicationWindow)
}

type senbaraGtkMainApplicationWindow struct {
	*adw.ApplicationWindow

	buttonTest   *gtk.Button
	toastOverlay *adw.ToastOverlay
}

func init() {
	if C.bindtextdomain(C.CString(GettextPackage), C.CString(LocaleDir)) == nil {
		panic("failed to bind text domain")
	}

	if C.bind_textdomain_codeset(C.CString(GettextPackage), C.CString("UTF-8")) == nil {
		panic("failed to set text domain codeset")
	}

	if C.textdomain(C.CString(GettextPackage)) == nil {
		panic("failed to set text domain")
	}

	resource, err := gio.NewResourceFromData(glib.NewBytes(ResourceContents, uint(len(ResourceContents))))
	if err != nil {
		panic(err)
	}
	gio.ResourcesRegister(resource)

	var classInit gobject.ClassInitFunc = func(tc *gobject.TypeClass, u uintptr) {
		typeClass := (*gtk.WidgetClass)(unsafe.Pointer(tc))
		typeClass.SetTemplateFromResource(resourceWindowUIPath)

		typeClass.BindTemplateChildFull("button_test", false, 0)
		typeClass.BindTemplateChildFull("toast_overlay", false, 0)

		gobject.SignalNewv(
			"button-test-clicked",
			gTypeSenbaraGtkMainApplicationWindow,
			gobject.GSignalRunFirstValue,
			nil,
			nil,
			0,
			nil,
			types.GType(gobject.TypeNoneVal),
			0,
			nil,
		)

		objClass := (*gobject.ObjectClass)(unsafe.Pointer(tc))

		objClass.SetCallbackConstructed(func(o *gobject.Object) {
			parentObjClass := (*gobject.ObjectClass)(unsafe.Pointer(tc.PeekParent()))

			parentObjClass.GetCallbackConstructed()(o)

			var parent adw.ApplicationWindow
			o.Cast(&parent)

			parent.InitTemplate()

			var buttonTest gtk.Button
			parent.Widget.GetTemplateChild(
				gTypeSenbaraGtkMainApplicationWindow,
				"button_test",
			).Cast(&buttonTest)

			var toastOverlay adw.ToastOverlay
			parent.Widget.GetTemplateChild(
				gTypeSenbaraGtkMainApplicationWindow,
				"toast_overlay",
			).Cast(&toastOverlay)

			w := &senbaraGtkMainApplicationWindow{
				ApplicationWindow: &parent,
				buttonTest:        &buttonTest,
				toastOverlay:      &toastOverlay,
			}

			var pinner runtime.Pinner
			pinner.Pin(w)

			var cleanupCallback glib.DestroyNotify = func(data uintptr) {
				pinner.Unpin()
			}
			o.SetDataFull(dataKeyGoInstance, uintptr(unsafe.Pointer(w)), &cleanupCallback)

			cb := func(gtk.Button) {
				gobject.SignalEmit(
					o,
					gobject.SignalLookup("button-test-clicked", gTypeSenbaraGtkMainApplicationWindow),
					0,
				)
			}

			buttonTest.ConnectClicked(&cb)
		})

		objClass.SetCallbackSetProperty(func(o *gobject.Object, u uint, v *gobject.Value, ps *gobject.ParamSpec) {
			switch u {
			case propertyIdTestButtonSensitive:
				w := (*senbaraGtkMainApplicationWindow)(unsafe.Pointer(o.GetData(dataKeyGoInstance)))

				w.buttonTest.SetSensitive(v.GetBoolean())
			}
		})

		objClass.SetCallbackGetProperty(func(o *gobject.Object, u uint, v *gobject.Value, ps *gobject.ParamSpec) {
			switch u {
			case propertyIdTestButtonSensitive:
				w := (*senbaraGtkMainApplicationWindow)(unsafe.Pointer(o.GetData(dataKeyGoInstance)))

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

	var parentQuery gobject.TypeQuery
	gobject.NewTypeQuery(adw.ApplicationWindowGLibType(), &parentQuery)

	var instanceInit gobject.InstanceInitFunc = func(ti *gobject.TypeInstance, tc *gobject.TypeClass) {}

	gTypeSenbaraGtkMainApplicationWindow = gobject.TypeRegisterStaticSimple(
		adw.ApplicationWindowGLibType(),
		"SenbaraGtkMainApplicationWindow",
		parentQuery.ClassSize,
		&classInit,
		parentQuery.InstanceSize+uint(unsafe.Sizeof(senbaraGtkMainApplicationWindow{}))+uint(unsafe.Sizeof(&senbaraGtkMainApplicationWindow{})),
		&instanceInit,
		0,
	)
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
