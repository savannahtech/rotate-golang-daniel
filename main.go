package main

import (
	"embed"
	"log"

	"github.com/danielboakye/filechangestracker/internal/core"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
)

//go:embed all:frontend
var assets embed.FS

func main() {
	app := core.New()
	err := wails.Run(&options.App{
		Title:             "Filechangestracker",
		Width:             1200,
		Height:            800,
		HideWindowOnClose: true,
		LogLevel:          logger.DEBUG,
		Assets:            assets,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		log.Fatal("failed to start app: %w", err)
	}
}
