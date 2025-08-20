#ifndef LIBSENBARA_GTK_H
#define LIBSENBARA_GTK_H

#include <glib-object.h>
#include <gtk/gtk.h>

G_BEGIN_DECLS

#define LIBSENBARA_GTK_MAIN_APPLICATION_WINDOW (libsenbara_gtk_main_application_window_get_type())

typedef struct _LibSenbaraGtkMainApplicationWindow LibSenbaraGtkMainApplicationWindow;
typedef struct _LibSenbaraGtkMainApplicationWindowClass LibSenbaraGtkMainApplicationWindowClass;

struct _LibSenbaraGtkMainApplicationWindow {
    GtkApplicationWindow parent_instance;
};

struct _LibSenbaraGtkMainApplicationWindowClass {
    GtkApplicationWindowClass parent_class;
};

/**
 * libsenbara_gtk_main_application_window_get_type:
 *
 * Gets the GType for LibSenbaraGtkMainApplicationWindow.
 *
 * Returns: the GType for LibSenbaraGtkMainApplicationWindow
 */
GType libsenbara_gtk_main_application_window_get_type(void);

/**
 * libsenbara_gtk_main_application_window_new:
 *
 * Creates a new LibSenbaraGtkMainApplicationWindow.
 *
 * Returns: (transfer full): A new LibSenbaraGtkMainApplicationWindow widget
 */
LibSenbaraGtkMainApplicationWindow* libsenbara_gtk_main_application_window_new(void);

/**
 * libsenbara_gtk_init_types:
 *
 * Initializes the libsenbara-gtk type system.
 * Call this before using any libsenbara-gtk widgets.
 */
void libsenbara_gtk_init_types(void);

G_END_DECLS

#endif /* LIBSENBARA_GTK_H */
