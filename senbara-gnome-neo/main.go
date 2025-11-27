package main

//go:generate sh -c "if [ -z \"$FLATPAK_ID\" ]; then GOWORK=off go tool github.com/dennwc/flatpak-go-mod --json . && GOWORK=off go tool github.com/dennwc/flatpak-go-mod --json --module-name senbaragtk ../senbara-gtk; fi"
