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
	appPath = path.Join("/com", "pojtinger", "felicitas", "SenbaraGtk")

	resourceWindowUIPath = path.Join(appPath, "window.ui")
)

//go:embed senbara-gtk.gresource
var ResourceContents []byte
