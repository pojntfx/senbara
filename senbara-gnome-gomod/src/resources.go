package main

import (
	_ "embed"
	"path"
)

const (
	AppID      = "com.pojtinger.felicitas.SenbaraGnomeGomod"
	AppVersion = "0.1.0"
)

//go:generate sh -c "blueprint-compiler batch-compile . . *.blp && glib-compile-resources *.gresource.xml"
//go:embed senbara-gnome-gomod.gresource
var ResourceContents []byte

var (
	AppPath = path.Join("/com", "pojtinger", "felicitas", "SenbaraGnomeGomod")

	ResourceWindowUIPath = path.Join(AppPath, "window.ui")
)
