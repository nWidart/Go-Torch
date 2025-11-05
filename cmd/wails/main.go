package main

import (
	app2 "GoTorch/internal/app"
	"context"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
)

func main() {
	app := app2.New()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "GoTorch - Torchlight Infinite Tracker",
		Width:  1100,
		Height: 800,
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
