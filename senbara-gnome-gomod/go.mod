module github.com/pojntfx/senbara/senbara-gnome-gomod

go 1.25.0

tool github.com/dennwc/flatpak-go-mod

require (
	github.com/jwijenbergh/puregotk v0.0.0-20251022075221-eae1610c7d83
	github.com/pojntfx/go-gettext v0.2.0
	github.com/pojntfx/senbara/senbara-gtk-go-meson v0.0.0-00010101000000-000000000000
)

require (
	github.com/dennwc/flatpak-go-mod v0.1.1-0.20251127123506-956509dd96ba // indirect
	github.com/goccy/go-yaml v1.18.0 // indirect
	github.com/jwijenbergh/purego v0.0.0-20251017112123-b71757b9ba42 // indirect
	golang.org/x/mod v0.30.0 // indirect
)

replace (
	github.com/jwijenbergh/puregotk => github.com/pojntfx/puregotk v0.0.0-20251127054829-d0e087e37740
	github.com/pojntfx/senbara/senbara-gtk-go-meson v0.0.0-00010101000000-000000000000 => ../senbara-gtk-go-meson
)
