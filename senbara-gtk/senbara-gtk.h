#pragma once

#include <gtk/gtk.h>

G_BEGIN_DECLS

#define SENBARA_GTK_INSIDE

/**
 * SenbaraGtkMainApplicationWindow:
 *
 * Example application window with a test button and toast notifications.
 */
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
 * senbara_gtk_main_application_window_show_toast:
 * @window: a SenbaraGtkMainApplicationWindow
 * @message: the message to display in the toast
 *
 * Shows a toast notification with the given message inside the window.
 */
void senbara_gtk_main_application_window_show_toast(
    SenbaraGtkMainApplicationWindow *window, const char *message);

#undef SENBARA_GTK_INSIDE

G_END_DECLS