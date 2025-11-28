api_version = '0.1'

pkg = import('pkgconfig')
fs = import('fs')

config_go = configuration_data()
config_go.set_quoted('PACKAGE_VERSION', meson.project_version())
config_go.set_quoted('GETTEXT_PACKAGE', 'senbara-gnome-neo')
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

senbara_gnome_neo_source_dir = meson.current_source_dir()

senbara_gnome_neo_source_go_files = files('senbara-gnome-neo.go')

senbara_gnome_neo_headers = ['senbara-gnome-neo.h']

version_split = meson.project_version().split('.')
version_conf = configuration_data()
version_conf.set('VERSION', meson.project_version())
version_conf.set('MAJOR_VERSION', version_split[0])
version_conf.set('MINOR_VERSION', version_split[1])
version_conf.set('MICRO_VERSION', version_split[2])

senbara_gnome_neo_deps = [
    dependency('gtk4'),
    dependency('libadwaita-1', version: '>= 1.4'),
    dependency('senbara-gtk-0.1', version: '>= 0.1.0'),
]

senbara_gnome_neo_blueprints = custom_target(
    'senbara-gnome-neo-blueprints',
    input: files('main-window.blp'),
    output: '.',
    command: [
        find_program('blueprint-compiler'),
        'batch-compile',
        '@CURRENT_SOURCE_DIR@',
        '@CURRENT_SOURCE_DIR@',
        '@INPUT@',
    ],
)

senbara_gnome_neo_resources = custom_target(
    'senbara-gnome-neo-resources',
    input: gnome.compile_resources(
        'senbara-gnome-neo',
        'senbara-gnome-neo.gresource.xml',
        gresource_bundle: true,
        dependencies: [senbara_gnome_neo_blueprints],
    ),
    output: 'senbara-gnome-neo-resources.gresource',
    command: [
        'cp',
        '@INPUT@',
        meson.current_source_dir() / 'senbara-gnome-neo.gresourc',
    ],
)

senbara_gnome_neo_bin = custom_target(
    'senbara-gnome-neo-lib-overwrite',
    build_always_stale: true,  # Go caches internally
    input: senbara_gnome_neo_source_go_files,
    output: 'senbara-gnome-neo',
    command: [
        find_program('go').full_path(),
        'build',
        '-x',
        '-o',
        '@OUTPUT@',
        senbara_gnome_neo_source_dir,
    ],
    depends: [senbara_gnome_neo_resources, config_go_file],
    install: true,
    install_dir: get_option('prefix') / get_option('bindir'),
)
