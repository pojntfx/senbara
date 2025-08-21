#ifndef SENBARA_GTK_H
#define SENBARA_GTK_H

#include <gtk/gtk.h>

G_BEGIN_DECLS

#define SENBARA_GTK_MAIN_APPLICATION_WINDOW                                    \
  (senbara_gtk_main_application_window_get_type())

typedef struct _SenbaraGtkMainApplicationWindow SenbaraGtkMainApplicationWindow;

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
SenbaraGtkMainApplicationWindow *senbara_gtk_main_application_window_new(void);

/**
 * senbara_gtk_main_application_window_set_test_button_sensitive:
 * @window: a SenbaraGtkMainApplicationWindow
 * @sensitive: whether the test button should be sensitive
 *
 * Sets the sensitivity of the test button in the window.
 */
void senbara_gtk_main_application_window_set_test_button_sensitive(
    SenbaraGtkMainApplicationWindow *window, gboolean sensitive);

/**
 * senbara_gtk_init_types:
 *
 * Initializes the senbara-gtk type system.
 * Call this before using any senbara-gtk widgets.
 */
void senbara_gtk_init_types(void);

G_END_DECLS

#endif /* SENBARA_GTK_H */
