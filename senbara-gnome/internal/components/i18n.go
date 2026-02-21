package components

import (
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/pojntfx/go-gettext/pkg/i18n"
	"github.com/pojntfx/senbara/senbara-gnome/po"
)

func InitEmbeddedI18n() error {
	// Self-extract locale files for i18n
	i18t, err := os.MkdirTemp("", "")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(i18t)

	if err := fs.WalkDir(po.FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if err := os.MkdirAll(filepath.Join(i18t, path), os.ModePerm); err != nil {
				return err
			}

			return nil
		}

		src, err := po.FS.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		dst, err := os.Create(filepath.Join(i18t, path))
		if err != nil {
			return err
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return i18n.InitI18n(gettextPackage, i18t, slog.Default())
}
