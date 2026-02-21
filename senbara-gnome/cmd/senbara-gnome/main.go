package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/pojntfx/senbara/senbara-gnome/assets/resources"
	"github.com/pojntfx/senbara/senbara-gnome/internal/components"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	level := new(slog.LevelVar)
	log := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	}))

	if err := components.InitEmbeddedI18n(); err != nil {
		panic(err)
	}

	settings, err := components.InitEmbeddedSettings()
	if err != nil {
		panic(err)
	}

	app := components.NewApplication(
		ctx,
		cancel,

		log,
		level,

		settings,

		"application_id", resources.AppID,
		"flags", gio.GApplicationHandlesOpenValue,
	)

	os.Exit(int(app.Run(int32(len(os.Args)), os.Args)))
}
