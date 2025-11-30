#pragma once

#include <adwaita.h>

G_BEGIN_DECLS

#define SENBARA_GTK_MESON_INSIDE

/**
 * SenbaraGtkMesonMainApplicationWindow:
 *
 * Example application window with a test button and toast notifications.
 */
#define SENBARA_GTK_MESON_MAIN_APPLICATION_WINDOW                                    \
  (senbara_gtk_meson_main_application_window_get_type())

typedef struct _SenbaraGtkMesonMainApplicationWindow SenbaraGtkMesonMainApplicationWindow;

/**
 * senbara_gtk_meson_main_application_window_get_type:
 *
 * Gets the GType for SenbaraGtkMesonMainApplicationWindow.
 *
 * Returns: the GType for SenbaraGtkMesonMainApplicationWindow
 */
GType senbara_gtk_meson_main_application_window_get_type(void);

/**
 * senbara_gtk_meson_main_application_window_show_toast:
 * @window: a SenbaraGtkMesonMainApplicationWindow
 * @message: the message to display in the toast
 *
 * Shows a toast notification with the given message inside the window.
 */
void senbara_gtk_meson_main_application_window_show_toast(
    SenbaraGtkMesonMainApplicationWindow *window, const char *message);

#undef SENBARA_GTK_MESON_INSIDE

G_END_DECLS
