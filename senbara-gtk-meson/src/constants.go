package main

import (
	_ "embed"

	"path"
)

const (
	dataKeyGoInstance = "go_instance"

	propertyIdTestButtonSensitive = 1
)

var (
	appPath = path.Join("/com", "pojtinger", "felicitas", "SenbaraGtkMeson")

	resourceWindowUIPath = path.Join(appPath, "window.ui")
)

//go:embed senbara-gtk-meson.gresource
var ResourceContents []byte
