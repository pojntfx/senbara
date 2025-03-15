package resources

import (
	_ "embed"
	"path"
)

const (
	AppID = "com.pojtinger.felicitas.Senbara"

	appPath = "/com/pojtinger/felicitas/Senbara/"
)

//go:generate blueprint-compiler compile --output window.ui window.blp
var ResourceWindowPath = path.Join(appPath, "window.ui")

//go:generate glib-compile-resources com.pojtinger.felicitas.Senbara.gresource.xml
//go:embed com.pojtinger.felicitas.Senbara.gresource
var ResourceContents []byte
