package tailer

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTailerSplitAcrossReads(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "log.txt")
	if err := os.WriteFile(p, []byte{}, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	out := make(chan string, 8)
	tlr := New(Options{Path: p, FromStart: true, PollEvery: 20 * time.Millisecond, ReadChunk: 4})
	go func() { _ = tlr.Start(ctx, out) }()

	// Write a partial line without newline; should not emit yet
	writeAppend(t, p, "hello")
	select {
	case s := <-out:
		t.Fatalf("unexpected line before newline: %q", s)
	case <-time.After(120 * time.Millisecond):
	}
	// Complete the line
	writeAppend(t, p, " world\n")
	select {
	case s := <-out:
		if s != "hello world" {
			t.Fatalf("got %q want 'hello world'", s)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for completed line")
	}

	cancel()
	tlr.Stop()
}

func TestTailerRotationRecover(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "log.txt")
	// initial
	if err := os.WriteFile(p, []byte("a1\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	out := make(chan string, 8)
	tlr := New(Options{Path: p, FromStart: true, PollEvery: 20 * time.Millisecond, ReadChunk: 64})
	go func() { _ = tlr.Start(ctx, out) }()

	// Expect a1
	select {
	case s := <-out:
		if s != "a1" {
			t.Fatalf("got %q want a1", s)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for a1")
	}
	// Rotate: rename old and create new file with same name
	old := p + ".old"
	if err := os.Rename(p, old); err != nil {
		t.Fatalf("rename: %v", err)
	}
	if err := os.WriteFile(p, []byte("b1\n"), 0o644); err != nil {
		t.Fatalf("write new: %v", err)
	}
	// Expect b1 after rotation detection
	select {
	case s := <-out:
		if s != "b1" {
			t.Fatalf("got %q want b1 after rotation", s)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for b1 after rotation")
	}

	cancel()
	tlr.Stop()
}
