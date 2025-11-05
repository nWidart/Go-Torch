package main

import (
	app2 "GoTorch/internal/app"
	"context"
	"testing"
)

func TestBuildRootAppOptions(t *testing.T) {
	app := app2.New()
	opts := buildRootAppOptions(app)
	if opts == nil {
		t.Fatal("nil options")
	}
	if opts.AssetServer == nil {
		t.Fatal("expected asset server configured")
	}
	if got, want := opts.Title, "GoTorch - Torchlight Infinite Tracker"; got != want {
		t.Fatalf("title=%q want %q", got, want)
	}
	if len(opts.Bind) != 1 {
		t.Fatalf("expected one bound object, got %d", len(opts.Bind))
	}
	// Call startup/shutdown callbacks to ensure no panic
	ctx := context.Background()
	if opts.OnStartup == nil || opts.OnShutdown == nil {
		t.Fatalf("expect startup/shutdown callbacks")
	}
	opts.OnStartup(ctx)
	opts.OnShutdown(ctx)
}
