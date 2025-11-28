package main

import (
	_ "embed"
	"path"
)

const (
	AppID      = "com.pojtinger.felicitas.SenbaraGnomeNeo"
	AppVersion = "0.1.0"
)

//go:embed index.gresource
var ResourceContents []byte

var (
	AppPath = path.Join("/com", "pojtinger", "felicitas", "SenbaraGnomeNeo")

	ResourceWindowUIPath = path.Join(AppPath, "main-window.ui")
)
