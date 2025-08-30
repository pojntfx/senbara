project('senbara-gtk', 'c',
          version: '0.1.0',
    meson_version: '>= 1.0.0',
  default_options: [ 
    'warning_level=2', 
    'werror=false', 
    'c_std=gnu11',
  ],
)

gnome = import('gnome')
pkg = import('pkgconfig')

config_h = configuration_data()
config_h.set_quoted('PACKAGE_VERSION', meson.project_version())
configure_file(output: 'config.h', configuration: config_h)
add_project_arguments(['-I' + meson.project_build_root()], language: 'c')
add_project_arguments(['-DSENBARA_GTK_COMPILATION'], language: 'c')

api_version = '0.1'
lib_name = 'senbara-gtk'
full_lib_name = lib_name + '-' + api_version

version_split = meson.project_version().split('.')
version_conf = configuration_data()
version_conf.set('VERSION', meson.project_version())
version_conf.set('MAJOR_VERSION', version_split[0])
version_conf.set('MINOR_VERSION', version_split[1])
version_conf.set('MICRO_VERSION', version_split[2])

senbara_gtk_version_h = configure_file(
          input: 'senbara-gtk-version.h.in',
         output: 'senbara-gtk-version.h',
  configuration: version_conf,
        install: true,
    install_dir: get_option('includedir') / lib_name
)

gtk4_dep = dependency('gtk4', version: '>= 4.17.5')
libadwaita_dep = dependency('libadwaita-1', version: '>= 1.7.5')

senbara_gtk_deps = [
  gtk4_dep,
  libadwaita_dep,
]

go = find_program('go', required: false)
glib_compile_resources = find_program('glib-compile-resources', required: true)
blueprint_compiler = find_program('blueprint-compiler', required: false)

if blueprint_compiler.found()
  window_ui = custom_target('window.ui',
    input : 'assets/resources/window.blp',
    output : 'window.ui',
    command : [blueprint_compiler, 'compile', '--output', '@OUTPUT@', '@INPUT@'],
    build_by_default : true,
    install : false
  )
else
  window_ui = files('assets/resources/window.ui')
endif

gresource_xml = files('assets/resources/index.gresource.xml')
gresource = custom_target('index.gresource',
  input : gresource_xml,
  output : 'index.gresource',
  command : [glib_compile_resources,
    '--sourcedir=' + meson.current_source_dir() + '/assets/resources',
    '--sourcedir=' + meson.current_build_dir(),
    '--target=@OUTPUT@',
    '@INPUT@'
  ],
  depends : blueprint_compiler.found() ? [window_ui] : [],
  build_by_default : true,
  install : false
)

senbara_gtk_headers = [
  'senbara-gtk.h',
]

if go.found()
  # Build the Go code as a static archive
  go_built_archive = custom_target('go-build-archive',
    output : 'lib' + full_lib_name + '.a',
    command : [
      find_program('sh'), '-c',
      'cd @0@ && @1@ generate ./... && @1@ build -buildmode=c-archive -o @2@/@OUTPUT@ .'.format(
        meson.current_source_dir(), 
        go.full_path(),
        meson.current_build_dir()
      )
    ],
    depends : [gresource],
    build_by_default : true,
    install : false
  )
  
  # Create a shared_library that includes the Go static archive
  # This satisfies gnome.generate_gir's type requirements
  senbara_gtk_lib = shared_library(full_lib_name,
    sources: [],  
    dependencies: senbara_gtk_deps,
    install: true,
    install_dir: get_option('libdir'),
    link_whole: [go_built_archive],  
  )
else
  warning('Go compiler not found. Build will be skipped.')
  senbara_gtk_lib = []
endif

install_headers(senbara_gtk_headers, subdir: lib_name)

pkg.generate(
  description: 'Example GTK library built with Go',
    libraries: '-L${libdir} -l' + full_lib_name,
         name: lib_name,
     filebase: full_lib_name,
      version: meson.project_version(),
      subdirs: lib_name,
     requires: ['gtk4 >= 4.17.5', 'libadwaita-1 >= 1.7.5'],
  install_dir: get_option('libdir') / 'pkgconfig'
)

if go.found()
  senbara_gtk_sources = [senbara_gtk_headers]
  
  senbara_gtk_gir = gnome.generate_gir(senbara_gtk_lib,
                sources: senbara_gtk_sources,
                 header: 'senbara-gtk.h',
        export_packages: full_lib_name,
              nsversion: api_version,
              namespace: 'SenbaraGtk',
          symbol_prefix: 'senbara_gtk',
      identifier_prefix: 'SenbaraGtk',
               includes: ['Gtk-4.0', 'Adw-1'],
                install: true,
  )
endif

gi_docgen = find_program('gi-docgen', required: false)

if gi_docgen.found() and go.found()
  docs_toml = files('docs.toml')
  
  docs = custom_target('documentation',
    output : ['docs'],
    command : [gi_docgen, 'generate',
      '--config=' + meson.current_source_dir() + '/docs.toml',
      '--output-dir=@OUTPUT@',
      '--no-namespace-dir',
      meson.current_build_dir() + '/SenbaraGtk-' + api_version + '.gir'
    ],
    depends : [senbara_gtk_gir[0]],
    build_by_default : true,
    install : true,
    install_dir : get_option('datadir') / 'doc' / full_lib_name
  )
  
  devhelp_file = custom_target('devhelp',
    input : docs,
    output : full_lib_name + '.devhelp2',
    command : ['cp', '@INPUT@/docs.devhelp2', '@OUTPUT@'],
    build_by_default : true,
    install : true,
    install_dir : get_option('datadir') / 'doc' / full_lib_name
  )
endif
