package resources

import (
	_ "embed"
	"path"
)

const (
	AppID = "com.pojtinger.felicitas.Senbara"

	appPath = "/com/pojtinger/felicitas/Senbara/"
)

//go:generate sh -c "blueprint-compiler batch-compile . . *.blp && scss index.scss index.css && glib-compile-resources index.gresource.xml"
//go:embed index.gresource
var ResourceContents []byte

var (
	ResourceWindowUIPath = path.Join(appPath, "window.ui")
	ResourceIndexCSSPath = path.Join(appPath, "index.css")
)
