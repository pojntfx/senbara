#ifndef SENBARA_H
#define SENBARA_H

#include <glib-object.h>
#include <gtk/gtk.h>

G_BEGIN_DECLS

#define SENBARA_TYPE_PURE_GO_TK_MAIN_WINDOW (senbara_pure_go_tk_main_window_get_type())

typedef struct _SenbaraPureGoTKMainWindow SenbaraPureGoTKMainWindow;
typedef struct _SenbaraPureGoTKMainWindowClass SenbaraPureGoTKMainWindowClass;

struct _SenbaraPureGoTKMainWindow {
    GtkApplicationWindow parent_instance;
};

struct _SenbaraPureGoTKMainWindowClass {
    GtkApplicationWindowClass parent_class;
};

/**
 * senbara_pure_go_tk_main_window_get_type:
 *
 * Gets the GType for SenbaraPureGoTKMainWindow.
 *
 * Returns: the GType for SenbaraPureGoTKMainWindow
 */
GType senbara_pure_go_tk_main_window_get_type(void);

/**
 * senbara_pure_go_tk_main_window_new:
 *
 * Creates a new SenbaraPureGoTKMainWindow.
 *
 * Returns: (transfer full): A new SenbaraPureGoTKMainWindow widget
 */
SenbaraPureGoTKMainWindow* senbara_pure_go_tk_main_window_new(void);

// /**
//  * senbara_pure_go_tk_main_window_connect_button_test_clicked:
//  * @window: a SenbaraPureGoTKMainWindow
//  * @callback: (scope call): The callback function to connect
//  *
//  * Connects a callback to the "button-test-clicked" signal.
//  */
// void senbara_pure_go_tk_main_window_connect_button_test_clicked(SenbaraPureGoTKMainWindow *window, GCallback callback);

/**
 * senbara_init_types:
 *
 * Initializes the Senbara type system.
 * Call this before using any Senbara widgets.
 */
void senbara_init_types(void);

G_END_DECLS

#endif /* SENBARA_H */
