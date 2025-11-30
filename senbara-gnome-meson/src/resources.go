package main

import (
	_ "embed"
	"path"
)

const (
	AppID      = "com.pojtinger.felicitas.SenbaraGnomeMeson"
	AppVersion = "0.1.0"
)

//go:embed senbara-gnome-meson.gresource
var ResourceContents []byte

var (
	AppPath = path.Join("/com", "pojtinger", "felicitas", "SenbaraGnomeMeson")

	ResourceWindowUIPath = path.Join(AppPath, "window.ui")
)
