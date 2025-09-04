api_version = '0.1'

pkg = import('pkgconfig')
gnome = import('gnome')

senbara_gtk_source = meson.current_source_dir() / meson.project_name() + '.go'

senbara_gtk_headers = [
  'senbara-gtk.h',
]

version_split = meson.project_version().split('.')
version_conf = configuration_data()
version_conf.set('VERSION', meson.project_version())
version_conf.set('MAJOR_VERSION', version_split[0])
version_conf.set('MINOR_VERSION', version_split[1])
version_conf.set('MICRO_VERSION', version_split[2])

configure_file(
  input: 'senbara-gtk-version.h.in',
  output: 'senbara-gtk-version.h',
  configuration: version_conf,
  install: true,
  install_dir: get_option('includedir') / 'senbara-gtk',
)

senbara_gtk_deps = [
  dependency('gtk4'),
  dependency('libadwaita-1', version: '>= 1.4'),
]

senbara_gtk_resources = custom_target(
  'senbara-gtk-resources',
  input: gnome.compile_resources(
    'senbara-gtk',
    'senbara-gtk.gresource.xml',
    gresource_bundle: true,
  ),
  output: 'senbara-gtk-resources.gresource',
  command: ['cp', '@INPUT@', meson.current_source_dir() / 'senbara-gtk.gresource'],
)

senbara_gtk_archive = custom_target(
  'senbara-gtk-archive',
  input: senbara_gtk_source,
  output: 'libsenbara-gtk-0.1.a',
  command: [
    find_program('go').full_path(),
    'build',
    '-buildmode', 'c-archive',
    '-x',
    '-o', '@OUTPUT@',
    '@INPUT@',
  ],
  depends: [
    senbara_gtk_resources,
  ],
)

senbara_gtk_lib = shared_library(
  'senbara-gtk-' + api_version,
  [],
  link_whole: [senbara_gtk_archive],
  dependencies: senbara_gtk_deps,
  install: true,
)

install_headers(senbara_gtk_headers, subdir: 'senbara-gtk')

pkg.generate(
  description: 'Example GTK library built with Go',
  libraries: senbara_gtk_lib,
  name: 'senbara-gtk',
  filebase: 'senbara-gtk-' + api_version,
  version: meson.project_version(),
  subdirs: 'senbara-gtk',
  requires: ['gtk4 >= 4.17.5', 'libadwaita-1 >= 1.7.5'],
  install_dir: get_option('libdir') / 'pkgconfig',
)

senbara_gtk_gir = gnome.generate_gir(
  senbara_gtk_lib,
  sources: [senbara_gtk_headers],
  header: 'senbara-gtk.h',
  export_packages: 'senbara-gtk-' + api_version,
  nsversion: api_version,
  namespace: 'SenbaraGtk',
  symbol_prefix: 'senbara_gtk',
  identifier_prefix: 'SenbaraGtk',
  includes: ['Gtk-4.0', 'Adw-1'],
  install: true,
)
