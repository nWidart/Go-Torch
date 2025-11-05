package tailer

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Ensure tailer can start before the file exists and begin emitting once it appears.
func TestTailerWaitsForFileThenReads(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "later.log")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	out := make(chan string, 8)
	tlr := New(Options{Path: p, FromStart: true, PollEvery: 20 * time.Millisecond, ReadChunk: 64})
	go func() { _ = tlr.Start(ctx, out) }()

	// After a short delay, create the file and write a line
	time.Sleep(120 * time.Millisecond)
	if err := os.WriteFile(p, []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	select {
	case s := <-out:
		if s != "hello" {
			t.Fatalf("got %q want hello", s)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for hello after file creation")
	}

	cancel()
	tlr.Stop()
}
