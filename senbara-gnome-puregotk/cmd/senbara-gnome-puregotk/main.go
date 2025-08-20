package main

import (
	"fmt"
	"log"
	"unsafe"

	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gobject"
	"github.com/jwijenbergh/puregotk/v4/gobject/types"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/pojntfx/senbara/senbara-gnome-puregotk/assets/resources"
)

import "C"

var (
	gTypeSenbaraPureGoTKMainWindow gobject.Type
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
	signalID := gobject.SignalLookup("button-test-clicked", gTypeSenbaraPureGoTKMainWindow)
	gobject.SignalEmit(obj, signalID, 0)
}

// C exports for GObject Introspection

//export senbara_pure_go_tk_main_window_get_type
func senbara_pure_go_tk_main_window_get_type() C.ulong {
	if gTypeSenbaraPureGoTKMainWindow == 0 {
		senbara_init_types()
	}
	return C.ulong(gTypeSenbaraPureGoTKMainWindow)
}

//export senbara_pure_go_tk_main_window_new
func senbara_pure_go_tk_main_window_new() unsafe.Pointer {
	fmt.Println("Calling constructor for widget in Go")
	window := newSenbaraPureGoTKMainWindow()

	return unsafe.Pointer(window.ApplicationWindow.Ptr)
}

// //export senbara_pure_go_tk_main_window_emit_button_test_clicked
// func senbara_pure_go_tk_main_window_emit_button_test_clicked(window unsafe.Pointer) {
// 	// if currentSenbaraPureGoTKMainWindow != nil {
// 	// 	currentSenbaraPureGoTKMainWindow.emitButtonTestClicked()
// 	// }
// }

// //export senbara_pure_go_tk_main_window_connect_button_test_clicked
// func senbara_pure_go_tk_main_window_connect_button_test_clicked(window unsafe.Pointer, callback unsafe.Pointer) {
// 	if currentSenbaraPureGoTKMainWindow != nil {
// 		// Note: This is a simplified callback connection
// 		// In a real implementation, you'd need proper callback marshaling
// 		cb := func() {
// 			// Call the C callback function
// 		}
// 		currentSenbaraPureGoTKMainWindow.ConnectButtonTestClicked(cb)
// 	}
// }

//export senbara_init_types
func senbara_init_types() {
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

	}

	var instanceInit gobject.InstanceInitFunc = func(ti *gobject.TypeInstance, tc *gobject.TypeClass) {
		typeClass := (*gtk.WidgetClass)(unsafe.Pointer(tc))
		typeClass.SetTemplateFromResource(resources.ResourceWindowUIPath)

		var callbackSymbol gobject.Callback = func() {
			log.Println("Callback on_button_test_clicked called")

			obj := (*gobject.Object)(unsafe.Pointer(ti))
			signalID := gobject.SignalLookup("button-test-clicked", gTypeSenbaraPureGoTKMainWindow)
			gobject.SignalEmit(obj, signalID, 0)
		}

		typeClass.BindTemplateCallbackFull("on_button_test_clicked", &callbackSymbol)
	}

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
	// Required for c-shared build mode
}
