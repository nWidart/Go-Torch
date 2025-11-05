package main

import (
	app2 "GoTorch/internal/app"
	"context"
	"testing"
)

func TestBuildAppOptions(t *testing.T) {
	app := app2.New()
	opts := buildAppOptions(app)
	if opts == nil {
		t.Fatal("expected options")
	}
	if opts.Title == "" || opts.Width == 0 || opts.Height == 0 {
		t.Fatalf("unexpected options: %+v", opts)
	}
	if len(opts.Bind) == 0 {
		t.Fatalf("expected app to be bound")
	}
	// Invoke callbacks to cover code paths
	if opts.OnStartup == nil || opts.OnShutdown == nil {
		t.Fatalf("missing startup/shutdown callbacks")
	}
	opts.OnStartup(context.Background())
	opts.OnShutdown(context.Background())
}
