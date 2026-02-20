package main

import (
	"time"
	"unsafe"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
)

func glibDateTimeFromGo(t time.Time) *glib.DateTime {
	return glib.NewDateTimeLocal(int32(t.Year()), int32(t.Month()), int32(t.Day()), int32(t.Hour()), int32(t.Minute()), float64(t.Second()))
}

func connectButtonClicked(button *gtk.Button, fn func()) {
	cb := func(_ gtk.Button) { fn() }
	button.ConnectClicked(&cb)
}

func connectSimpleActionActivate(action *gio.SimpleAction, fn func()) {
	cb := func(_ gio.SimpleAction, _ uintptr) { fn() }
	action.ConnectActivate(&cb)
}

func connectSimpleActionActivateWithParam(action *gio.SimpleAction, fn func(parameter *glib.Variant)) {
	cb := func(_ gio.SimpleAction, paramPtr uintptr) {
		var param *glib.Variant
		if paramPtr != 0 {
			param = (*glib.Variant)(unsafe.Pointer(paramPtr))
		}
		fn(param)
	}
	action.ConnectActivate(&cb)
}

func connectSettingsChanged(settings *gio.Settings, fn func(key string)) {
	cb := func(_ gio.Settings, key string) { fn(key) }
	settings.ConnectChanged(&cb)
}

func connectEntryRowChanged(row *adw.EntryRow, fn func()) {
	cb := func() { fn() }
	row.PreferencesRow.ListBoxRow.Widget.InitiallyUnowned.Object.ConnectSignal("changed", &cb)
}

func connectPasswordEntryRowChanged(row *adw.PasswordEntryRow, fn func()) {
	cb := func() { fn() }
	row.EntryRow.PreferencesRow.ListBoxRow.Widget.InitiallyUnowned.Object.ConnectSignal("changed", &cb)
}

func connectSearchEntryChanged(entry *gtk.SearchEntry, fn func()) {
	cb := func(_ gtk.SearchEntry) { fn() }
	entry.ConnectSearchChanged(&cb)
}

func setListBoxFilterFunc(listBox *gtk.ListBox, fn func(row *gtk.ListBoxRow) bool) {
	filterFunc := gtk.ListBoxFilterFunc(func(rowPtr uintptr, _ uintptr) bool {
		row := gtk.ListBoxRowNewFromInternalPtr(rowPtr)
		return fn(row)
	})
	destroyNotify := glib.DestroyNotify(func(_ uintptr) {})
	listBox.SetFilterFunc(&filterFunc, 0, &destroyNotify)
}

func connectTextBufferChanged(buffer *gtk.TextBuffer, fn func()) {
	cb := func(_ gtk.TextBuffer) { fn() }
	buffer.ConnectChanged(&cb)
}

func connectDialogClosed(dialog *adw.Dialog, fn func()) {
	cb := func(_ adw.Dialog) { fn() }
	dialog.ConnectClosed(&cb)
}

func getTextBufferText(buffer *gtk.TextBuffer) string {
	var startIter, endIter gtk.TextIter
	buffer.GetStartIter(&startIter)
	buffer.GetEndIter(&endIter)
	return buffer.GetText(&startIter, &endIter, true)
}

func connectAlertDialogResponse(dialog *adw.AlertDialog, fn func(response string)) {
	cb := func(_ adw.AlertDialog, response string) { fn(response) }
	dialog.ConnectResponse(&cb)
}

func idleAdd(fn func()) {
	sourceFn := glib.SourceFunc(func(_ uintptr) bool {
		fn()
		return false // Return false to stop the idle source after first run
	})
	glib.IdleAdd(&sourceFn, 0)
}

func connectListBoxRowActivated(listBox *gtk.ListBox, fn func(row *gtk.ListBoxRow)) {
	cb := func(_ gtk.ListBox, rowPtr uintptr) {
		row := gtk.ListBoxRowNewFromInternalPtr(rowPtr)
		fn(row)
	}
	listBox.ConnectRowActivated(&cb)
}

func connectNavigationViewPopped(nv *adw.NavigationView, fn func(page *adw.NavigationPage)) {
	cb := func(_ adw.NavigationView, pagePtr uintptr) {
		page := adw.NavigationPageNewFromInternalPtr(pagePtr)
		fn(page)
	}
	nv.ConnectPopped(&cb)
}

func connectNavigationViewPushed(nv *adw.NavigationView, fn func()) {
	cb := func(_ adw.NavigationView) { fn() }
	nv.ConnectPushed(&cb)
}

func connectNavigationViewReplaced(nv *adw.NavigationView, fn func()) {
	cb := func(_ adw.NavigationView) { fn() }
	nv.ConnectReplaced(&cb)
}

func fileDialogSave(fd *gtk.FileDialog, parent *gtk.Window, fn func(file *gio.FileBase, err error)) {
	cb := gio.AsyncReadyCallback(func(_ uintptr, resultPtr uintptr, _ uintptr) {
		result := gio.AsyncResultBase{Ptr: resultPtr}
		file, err := fd.SaveFinish(&result)
		fn(file, err)
	})
	fd.Save(parent, nil, &cb, 0)
}

func fileDialogOpen(fd *gtk.FileDialog, parent *gtk.Window, fn func(file *gio.FileBase, err error)) {
	cb := gio.AsyncReadyCallback(func(_ uintptr, resultPtr uintptr, _ uintptr) {
		result := gio.AsyncResultBase{Ptr: resultPtr}
		file, err := fd.OpenFinish(&result)
		fn(file, err)
	})
	fd.Open(parent, nil, &cb, 0)
}
