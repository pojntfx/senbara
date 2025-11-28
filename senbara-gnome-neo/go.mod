module github.com/pojntfx/senbara/senbara-gnome-neo

go 1.25.0

tool github.com/dennwc/flatpak-go-mod

require (
	github.com/jwijenbergh/puregotk v0.0.0-20251022075221-eae1610c7d83
	github.com/pojntfx/go-gettext v0.2.0
	github.com/pojntfx/senbara/senbara-gtk-go v0.0.0-20251125083721-e474d86bebcc
)

require (
	github.com/dennwc/flatpak-go-mod v0.1.1-0.20251127123506-956509dd96ba // indirect
	github.com/goccy/go-yaml v1.18.0 // indirect
	github.com/jwijenbergh/purego v0.0.0-20251017112123-b71757b9ba42 // indirect
	golang.org/x/mod v0.30.0 // indirect
)

replace (
	github.com/jwijenbergh/puregotk => github.com/pojntfx/puregotk v0.0.0-20251127054829-d0e087e37740
	github.com/pojntfx/senbara/senbara-gtk-go v0.0.0-20250826075235-cbb2c7573805 => ../senbara-gtk-go
)
