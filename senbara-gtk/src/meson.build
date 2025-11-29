api_version = '0.1'

pkg = import('pkgconfig')
gnome = import('gnome')
fs = import('fs')

config_go = configuration_data()
config_go.set_quoted('PACKAGE_VERSION', meson.project_version())
config_go.set_quoted('GETTEXT_PACKAGE', 'senbara-gtk')
config_go.set_quoted('LOCALEDIR', get_option('prefix') / get_option('localedir'))

config_go_file = custom_target(
    'config-go',
    input: configure_file(
        input: meson.current_source_dir() / 'config.go.in',
        output: 'config_rendered.go',
        configuration: config_go,
    ),
    output: 'config.go',
    command: ['cp', '@INPUT@', meson.current_source_dir() / 'config.go'],
)

senbara_gtk_source_dir = meson.current_source_dir()

senbara_gtk_source_go_files = files('senbara-gtk.go')

senbara_gtk_headers = ['senbara-gtk.h', 'window.h']

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

senbara_gtk_blueprints = custom_target(
    'senbara-gtk-blueprints',
    input: files('window.blp'),
    output: '.',
    command: [
        find_program('blueprint-compiler'),
        'batch-compile',
        '@CURRENT_SOURCE_DIR@',
        '@CURRENT_SOURCE_DIR@',
        '@INPUT@',
    ],
)

senbara_gtk_resources = custom_target(
    'senbara-gtk-resources',
    input: gnome.compile_resources(
        'senbara-gtk',
        'senbara-gtk.gresource.xml',
        gresource_bundle: true,
        dependencies: [senbara_gtk_blueprints],
    ),
    output: 'senbara-gtk-resources.gresource',
    command: [
        'cp',
        '@INPUT@',
        meson.current_source_dir() / 'senbara-gtk.gresource',
    ],
)

senbara_gtk_lib_overwrite = custom_target(
    'senbara-gtk-lib-overwrite',
    build_always_stale: true,  # Go caches internally
    input: senbara_gtk_source_go_files,
    output: 'libsenbara-gtk-overwrite-' + api_version + '.so',
    command: [
        find_program('go').full_path(),
        'build',
        '-buildmode',
        'c-shared',
        '-x',
        '-o',
        '@OUTPUT@',
        senbara_gtk_source_dir,
    ],
    depends: [senbara_gtk_resources, config_go_file],
)

senbara_gtk_archive = custom_target(
    'senbara-gtk-archive',
    build_always_stale: true,  # Go caches internally
    input: senbara_gtk_source_go_files,
    output: 'libsenbara-gtk-' + api_version + '.a',
    command: [
        find_program('go').full_path(),
        'build',
        '-buildmode',
        'c-archive',
        '-x',
        '-o',
        '@OUTPUT@',
        senbara_gtk_source_dir,
    ],
    depends: [senbara_gtk_resources, config_go_file, senbara_gtk_lib_overwrite],
)

senbara_gtk_lib = shared_library(
    'senbara-gtk-' + api_version,
    [],
    link_whole: [senbara_gtk_archive],
    dependencies: senbara_gtk_deps,
    install: true,
)

# We overwrite the manually linked senbara_gtk_lib with senbara_gtk_lib_overwrite here since the manually linked library built with `c-archive` can not be used by another Go program without a `morestack` crash
meson.add_install_script(
    'install',
    senbara_gtk_lib_overwrite,
    get_option('prefix') / get_option('libdir') / fs.name(
        senbara_gtk_lib.full_path(),
    ),
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

custom_target(
    'senbara-gtk-docs',
    input: [
        configure_file(
            input: meson.current_source_dir() / 'docs.toml.in',
            output: 'docs.toml',
            configuration: config_go,
        ),
        senbara_gtk_gir,
    ],
    output: 'senbara-gtk-' + api_version,
    command: [
        find_program('gi-docgen'),
        'generate',
        '--no-namespace-dir',
        '--config=@INPUT0@',
        '--output-dir=@OUTPUT@',
        '@INPUT1@',
    ],
    install: true,
    install_dir: get_option('datadir') / 'doc',
)
