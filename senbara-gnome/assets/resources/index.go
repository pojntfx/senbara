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

//go:generate glib-compile-resources index.gresource.xml
//go:embed index.gresource
var ResourceContents []byte
