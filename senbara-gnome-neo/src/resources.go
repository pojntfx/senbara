package main

import (
	_ "embed"
	"path"
)

const (
	AppID      = "com.pojtinger.felicitas.SenbaraGnomeNeo"
	AppVersion = "0.1.0"
)

//go:embed senbara-gnome-neo.gresource
var ResourceContents []byte

var (
	AppPath = path.Join("/com", "pojtinger", "felicitas", "SenbaraGnomeNeo")

	ResourceWindowUIPath = path.Join(AppPath, "window.ui")
)
