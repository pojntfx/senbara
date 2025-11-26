project(
    'senbara-gnome-neo',
    'c',
    version: '0.1.0',
    meson_version: '>= 1.0.0',
    default_options: [],
)
add_project_arguments(['-I' + meson.project_build_root()], language: 'c')


i18n = import('i18n')
gnome = import('gnome')

subdir('src')
subdir('po')

gnome.post_install(
     glib_compile_schemas: true,
    gtk_update_icon_cache: true,
  update_desktop_database: true,
)
