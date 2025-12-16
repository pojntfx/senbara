package main

import (
	"os"
	"path"
	"path/filepath"

	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/pojntfx/senbara/senbara-gnome/assets/resources"
)

func initEmbeddedSettings() (*gio.Settings, error) {
	// Self-extract GSettings schema
	st, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(st)

	if err := os.WriteFile(filepath.Join(st, path.Base(resources.ResourceGSchemasCompiledPath)), resources.Schema, os.ModePerm); err != nil {
		return nil, err
	}

	if err := os.Setenv("GSETTINGS_SCHEMA_DIR", st); err != nil {
		return nil, err
	}

	return gio.NewSettings(resources.AppID), nil
}
