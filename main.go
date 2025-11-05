package main

import (
	app2 "GoTorch/internal/app"
	"context"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

var assetsFS embed.FS

func main() {
	app := app2.New()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "GoTorch - Torchlight Infinite Tracker",
		Width:  1100,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assetsFS,
		},
		OnStartup: func(ctx context.Context) {
			app.Startup(ctx)
		},
		OnShutdown: func(ctx context.Context) {
			app.Shutdown(ctx)
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
