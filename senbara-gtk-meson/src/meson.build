api_version = '0.1'

pkg = import('pkgconfig')
gnome = import('gnome')
fs = import('fs')

config_go = configuration_data()
config_go.set_quoted('PACKAGE_VERSION', meson.project_version())
config_go.set_quoted('GETTEXT_PACKAGE', 'senbara-gtk-meson')
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

senbara_gtk_meson_source_dir = meson.current_source_dir()

senbara_gtk_meson_source_go_files = files('senbara-gtk-meson.go')

senbara_gtk_meson_headers = ['senbara-gtk-meson.h', 'window.h']

version_split = meson.project_version().split('.')
version_conf = configuration_data()
version_conf.set('VERSION', meson.project_version())
version_conf.set('MAJOR_VERSION', version_split[0])
version_conf.set('MINOR_VERSION', version_split[1])
version_conf.set('MICRO_VERSION', version_split[2])

configure_file(
    input: 'senbara-gtk-meson-version.h.in',
    output: 'senbara-gtk-meson-version.h',
    configuration: version_conf,
    install: true,
    install_dir: get_option('includedir') / 'senbara-gtk-meson',
)

senbara_gtk_meson_deps = [
    dependency('gtk4'),
    dependency('libadwaita-1', version: '>= 1.4'),
]

senbara_gtk_meson_blueprints = custom_target(
    'senbara-gtk-meson-blueprints',
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

senbara_gtk_meson_resources = custom_target(
    'senbara-gtk-meson-resources',
    input: gnome.compile_resources(
        'senbara-gtk-meson',
        'senbara-gtk-meson.gresource.xml',
        gresource_bundle: true,
        dependencies: [senbara_gtk_meson_blueprints],
    ),
    output: 'senbara-gtk-meson-resources.gresource',
    command: [
        'cp',
        '@INPUT@',
        meson.current_source_dir() / 'senbara-gtk-meson.gresource',
    ],
)

senbara_gtk_meson_lib_overwrite = custom_target(
    'senbara-gtk-meson-lib-overwrite',
    build_always_stale: true,  # Go caches internally
    input: senbara_gtk_meson_source_go_files,
    output: 'libsenbara-gtk-meson-overwrite-' + api_version + '.so',
    command: [
        find_program('go').full_path(),
        'build',
        '-buildmode',
        'c-shared',
        '-x',
        '-o',
        '@OUTPUT@',
        senbara_gtk_meson_source_dir,
    ],
    depends: [senbara_gtk_meson_resources, config_go_file],
)

senbara_gtk_meson_archive = custom_target(
    'senbara-gtk-meson-archive',
    build_always_stale: true,  # Go caches internally
    input: senbara_gtk_meson_source_go_files,
    output: 'libsenbara-gtk-meson-' + api_version + '.a',
    command: [
        find_program('go').full_path(),
        'build',
        '-buildmode',
        'c-archive',
        '-x',
        '-o',
        '@OUTPUT@',
        senbara_gtk_meson_source_dir,
    ],
    depends: [senbara_gtk_meson_resources, config_go_file, senbara_gtk_meson_lib_overwrite],
)

senbara_gtk_meson_lib = shared_library(
    'senbara-gtk-meson-' + api_version,
    [],
    link_whole: [senbara_gtk_meson_archive],
    dependencies: senbara_gtk_meson_deps,
    install: true,
)

# We overwrite the manually linked senbara_gtk_meson_lib with senbara_gtk_meson_lib_overwrite here since the manually linked library built with `c-archive` can not be used by another Go program without a `morestack` crash
meson.add_install_script(
    'install',
    senbara_gtk_meson_lib_overwrite,
    get_option('prefix') / get_option('libdir') / fs.name(
        senbara_gtk_meson_lib.full_path(),
    ),
)

install_headers(senbara_gtk_meson_headers, subdir: 'senbara-gtk-meson')

pkg.generate(
    description: 'Example GTK library built with Go',
    libraries: senbara_gtk_meson_lib,
    name: 'senbara-gtk-meson',
    filebase: 'senbara-gtk-meson-' + api_version,
    version: meson.project_version(),
    subdirs: 'senbara-gtk-meson',
    requires: ['gtk4 >= 4.17.5', 'libadwaita-1 >= 1.7.5'],
    install_dir: get_option('libdir') / 'pkgconfig',
)

senbara_gtk_meson_gir = gnome.generate_gir(
    senbara_gtk_meson_lib,
    sources: [senbara_gtk_meson_headers],
    header: 'senbara-gtk-meson.h',
    export_packages: 'senbara-gtk-meson-' + api_version,
    nsversion: api_version,
    namespace: 'SenbaraGtkMeson',
    symbol_prefix: 'senbara_gtk_meson',
    identifier_prefix: 'SenbaraGtkMeson',
    includes: ['Gtk-4.0', 'Adw-1'],
    install: true,
)

custom_target(
    'senbara-gtk-meson-docs',
    input: [
        configure_file(
            input: meson.current_source_dir() / 'docs.toml.in',
            output: 'docs.toml',
            configuration: config_go,
        ),
        senbara_gtk_meson_gir,
    ],
    output: 'senbara-gtk-meson-' + api_version,
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
