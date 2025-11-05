package tailer

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeAppend(t *testing.T, path string, s string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		t.Fatalf("open append: %v", err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	if _, err := w.WriteString(s); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}
}

func TestTailerBasicFollowAndTruncate(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "log.txt")

	// initial content with CRLF and LF
	writeAppend(t, logPath, "foo\r\n")
	writeAppend(t, logPath, "bar\n")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	out := make(chan string, 16)
	tlr := New(Options{Path: logPath, FromStart: true, PollEvery: 50 * time.Millisecond, ReadChunk: 1024})
	// start tailer
	go func() { _ = tlr.Start(ctx, out) }()

	// expect first two lines
	want := []string{"foo", "bar"}
	got := make([]string, 0, 8)
	deadline := time.Now().Add(2 * time.Second)
	for len(got) < len(want) && time.Now().Before(deadline) {
		select {
		case s := <-out:
			got = append(got, s)
		case <-time.After(50 * time.Millisecond):
		}
	}
	if len(got) < len(want) {
		t.Fatalf("timeout waiting lines, got=%v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("line %d = %q; want %q", i, got[i], want[i])
		}
	}

	// append a new line and expect it
	writeAppend(t, logPath, "baz\n")
	select {
	case s := <-out:
		if s != "baz" {
			t.Fatalf("got %q want baz", s)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for baz")
	}

	// simulate truncation and new content
	if err := os.Truncate(logPath, 0); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	writeAppend(t, logPath, "new\n")
	select {
	case s := <-out:
		if s != "new" {
			t.Fatalf("got %q want new", s)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for new after truncate")
	}

	// stop
	cancel()
	tlr.Stop()
}
