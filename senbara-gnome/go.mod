module github.com/pojntfx/senbara/senbara-gnome

go 1.24.0

replace github.com/pojntfx/gotk4-secret/pkg => /home/pojntfx/Projets/gotk4-adwaita/pkg

require (
	github.com/diamondburned/gotk4-adwaita/pkg v0.0.0-20250223021911-503726bcfce6
	github.com/diamondburned/gotk4/pkg v0.3.1
	github.com/zalando/go-keyring v0.2.6
)

require (
	al.essio.dev/pkg/shellescape v1.5.1 // indirect
	github.com/KarpelesLab/weak v0.1.1 // indirect
	github.com/danieljoos/wincred v1.2.2 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	go4.org/unsafe/assume-no-moving-gc v0.0.0-20231121144256-b99613f794b6 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
)
