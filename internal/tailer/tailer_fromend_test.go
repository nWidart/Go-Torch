package tailer

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Test that FromStart=false starts tailing at EOF and doesn't emit historical lines.
func TestTailerFromEndSkipsHistory(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "log.txt")
	if err := os.WriteFile(p, []byte("old1\nold2\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	out := make(chan string, 8)
	tlr := New(Options{Path: p, FromStart: false, PollEvery: 30 * time.Millisecond, ReadChunk: 32})
	go func() { _ = tlr.Start(ctx, out) }()

	// Give it a couple polls to start and confirm no historical lines arrive
	select {
	case s := <-out:
		t.Fatalf("unexpected historical line: %q", s)
	case <-time.After(120 * time.Millisecond):
	}

	// Append a new line and expect it
	writeAppend(t, p, "new\n")
	select {
	case s := <-out:
		if s != "new" {
			t.Fatalf("got %q want new", s)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for new line from end")
	}

	cancel()
	tlr.Stop()
}
