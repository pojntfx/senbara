package main

import (
	"fmt"
	"runtime"
	"time"
	"unsafe"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gobject"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	. "github.com/pojntfx/go-gettext/pkg/i18n"
	"github.com/pojntfx/senbara/senbara-gtk-go-meson/senbaragtkmeson"
)

var (
	gTypeWindow gobject.Type
)

type Window struct {
	adw.ApplicationWindow

	buttonTest *gtk.Button
}

func NewWindow(FirstPropertyNameVar string, varArgs ...interface{}) Window {
	obj := gobject.NewObject(gTypeWindow, FirstPropertyNameVar, varArgs...)

	var v Window
	obj.Cast(&v)

	return v
}

func (w *Window) SetButtonTestSensitive(sensitive bool) {
	w.buttonTest.SetSensitive(sensitive)
}

func init() {
	var windowClassInit gobject.ClassInitFunc = func(tc *gobject.TypeClass, u uintptr) {
		typeClass := (*gtk.WidgetClass)(unsafe.Pointer(tc))
		typeClass.SetTemplateFromResource(ResourceWindowUIPath)

		typeClass.BindTemplateChildFull("button_test", false, 0)

		objClass := (*gobject.ObjectClass)(unsafe.Pointer(tc))

		objClass.OverrideConstructed(func(o *gobject.Object) {
			parentObjClass := (*gobject.ObjectClass)(unsafe.Pointer(tc.PeekParent()))
			parentObjClass.GetConstructed()(o)

			var parent adw.ApplicationWindow
			o.Cast(&parent)

			parent.InitTemplate()

			var buttonTest gtk.Button
			parent.Widget.GetTemplateChild(
				gTypeWindow,
				"button_test",
			).Cast(&buttonTest)

			w := &Window{
				ApplicationWindow: parent,

				buttonTest: &buttonTest,
			}

			var pinner runtime.Pinner
			pinner.Pin(w)

			var cleanupCallback glib.DestroyNotify = func(data uintptr) {
				pinner.Unpin()
			}
			o.SetDataFull(dataKeyGoInstance, uintptr(unsafe.Pointer(w)), &cleanupCallback)

			onButtonTestClicked := func(gtk.Button) {
				fmt.Println("senbara-gnome-meson test button clicked, opening senbara-gtk-meson window")

				obj := gobject.NewObject(senbaragtkmeson.MainApplicationWindowGLibType(),
					"application", parent.GetApplication(),
				)

				var senbaraWindow senbaragtkmeson.MainApplicationWindow
				obj.Cast(&senbaraWindow)

				onSenbaraButtonTestClicked := func(sw senbaragtkmeson.MainApplicationWindow) {
					fmt.Println("senbara-gtk test button clicked")

					senbaraWindow.ShowToast(L("Button was clicked!"))
					senbaraWindow.SetPropertyTestButtonSensitive(false)

					var timer *time.Timer
					onDestroy := func(gtk.Widget) {
						if timer != nil {
							timer.Stop()
						}
					}
					senbaraWindow.ConnectDestroy(&onDestroy)

					timer = time.AfterFunc(time.Second*3, func() {
						senbaraWindow.ShowToast(L("Button re-enabled after 3 seconds"))
						senbaraWindow.SetPropertyTestButtonSensitive(true)
					})
				}
				senbaraWindow.ConnectButtonTestClicked(&onSenbaraButtonTestClicked)

				senbaraWindow.Present()
			}
			buttonTest.ConnectClicked(&onButtonTestClicked)
		})
	}

	var windowInstanceInit gobject.InstanceInitFunc = func(ti *gobject.TypeInstance, tc *gobject.TypeClass) {}

	var windowParentQuery gobject.TypeQuery
	gobject.NewTypeQuery(adw.ApplicationWindowGLibType(), &windowParentQuery)

	gTypeWindow = gobject.TypeRegisterStaticSimple(
		windowParentQuery.Type,
		"SenbaraGnomeMesonWindow",
		windowParentQuery.ClassSize,
		&windowClassInit,
		windowParentQuery.InstanceSize+uint(unsafe.Sizeof(Window{}))+uint(unsafe.Sizeof(&Window{}))+uint(unsafe.Sizeof(&gtk.Button{})),
		&windowInstanceInit,
		0,
	)
}
