#ifndef SENBARA_GTK_H
#define SENBARA_GTK_H

#include <glib-object.h>
#include <gtk/gtk.h>

G_BEGIN_DECLS

#define SENBARA_GTK_MAIN_APPLICATION_WINDOW (senbara_gtk_main_application_window_get_type())

typedef struct _SenbaraGtkMainApplicationWindow SenbaraGtkMainApplicationWindow;
typedef struct _SenbaraGtkMainApplicationWindowClass SenbaraGtkMainApplicationWindowClass;

struct _SenbaraGtkMainApplicationWindow {
    GtkApplicationWindow parent_instance;
};

struct _SenbaraGtkMainApplicationWindowClass {
    GtkApplicationWindowClass parent_class;
};

/**
 * senbara_gtk_main_application_window_get_type:
 *
 * Gets the GType for SenbaraGtkMainApplicationWindow.
 *
 * Returns: the GType for SenbaraGtkMainApplicationWindow
 */
GType senbara_gtk_main_application_window_get_type(void);

/**
 * senbara_gtk_main_application_window_new:
 *
 * Creates a new SenbaraGtkMainApplicationWindow.
 *
 * Returns: (transfer full): A new SenbaraGtkMainApplicationWindow widget
 */
SenbaraGtkMainApplicationWindow* senbara_gtk_main_application_window_new(void);

/**
 * senbara_gtk_init_types:
 *
 * Initializes the senbara-gtk type system.
 * Call this before using any senbara-gtk widgets.
 */
void senbara_gtk_init_types(void);

G_END_DECLS

#endif /* SENBARA_GTK_H */
