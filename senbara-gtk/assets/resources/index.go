package resources

import (
	_ "embed"
	"path"
)

//go:generate sh -c "blueprint-compiler batch-compile . . *.blp && glib-compile-resources *.gresource.xml"
//go:embed index.gresource
var ResourceContents []byte

var (
	AppPath = path.Join("/com", "pojtinger", "felicitas", "SenbaraGtk")

	ResourceWindowUIPath = path.Join(AppPath, "window.ui")
)
