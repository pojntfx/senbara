package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/jwijenbergh/puregotk/pkg/gir/pass"
	"github.com/jwijenbergh/puregotk/pkg/gir/util"
)

//go:generate go run gen.go

func main() {
	dir := "v4"
	os.RemoveAll(dir)
	var girs []string
	var localGirs []string

	// Collect local GIR files
	filepath.Walk("internal/gir/spec", func(path string, f os.FileInfo, err error) error {
		if !strings.HasSuffix(path, ".gir") {
			return nil
		}
		girs = append(girs, path)
		localGirs = append(localGirs, path)
		return nil
	})

	// Find puregotk dependency path and add its GIR files
	puregotk := ""
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/jwijenbergh/puregotk")
	output, err := cmd.Output()
	if err == nil {
		puregotk = strings.TrimSpace(string(output))
		filepath.Walk(filepath.Join(puregotk, "pkg/gir/spec"), func(path string, f os.FileInfo, err error) error {
			if !strings.HasSuffix(path, ".gir") {
				return nil
			}
			girs = append(girs, path)
			return nil
		})
	}

	p, err := pass.New(girs)
	if err != nil {
		panic(err)
	}
	// collect basic type info
	p.First()

	// Create separate pass for generation with only local files if we have dependencies
	if len(girs) > len(localGirs) {
		pLocal, err := pass.New(localGirs)
		if err != nil {
			panic(err)
		}
		pLocal.Types = p.Types
		p = pLocal
	}

	// Create the template
	gotemp, err := template.New("go").Funcs(template.FuncMap{"conv": util.ConvertArgs, "convc": util.ConvertArgsComma}).ParseFiles(filepath.Join(puregotk, "templates/go"))
	if err != nil {
		panic(err)
	}

	// Write go files by making the second pass
	p.Second(dir, gotemp)
}
