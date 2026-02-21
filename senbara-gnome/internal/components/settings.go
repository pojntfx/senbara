package components

import (
	"errors"
	"os"
	"path"
	"path/filepath"

	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/pojntfx/senbara/senbara-gnome/assets/resources"
)

func InitEmbeddedSettings() (*gio.Settings, error) {
	// Self-extract GSettings schema
	st, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(st)

	if err := os.WriteFile(filepath.Join(st, path.Base(resources.ResourceGSchemasCompiledPath)), resources.Schema, os.ModePerm); err != nil {
		return nil, err
	}

	source, err := gio.NewSettingsSchemaSourceFromDirectory(st, gio.SettingsSchemaSourceGetDefault(), true)
	if err != nil {
		return nil, err
	}

	schema := source.Lookup(resources.AppID, false)
	if schema == nil {
		return nil, errors.New("could not find schema")
	}

	return gio.NewSettingsFull(schema, nil, schema.GetPath()), nil
}
