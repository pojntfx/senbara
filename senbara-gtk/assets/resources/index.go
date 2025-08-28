package resources

import (
	_ "embed"
	"path"
)

//go:generate sh -c "find ../../po -name '*.po' | sed 's|^\\../../po/||; s|\\.po$||' > ../../po/LINGUAS"

const (
	AppID = "com.pojtinger.felicitas.senbaragtk.Example"
)

//go:generate sh -c "blueprint-compiler batch-compile . . *.blp && glib-compile-resources *.gresource.xml"
//go:embed index.gresource
var ResourceContents []byte

var (
	AppPath = path.Join("/com", "pojtinger", "felicitas", "senbaragtk", "Example")

	ResourceWindowUIPath = path.Join(AppPath, "window.ui")
)
