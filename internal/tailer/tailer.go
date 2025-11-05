package tailer

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"sync"
	"time"
)

// Options control tailing behavior.
type Options struct {
	Path      string        // Log file path
	FromStart bool          // If true, start reading from start, else from end
	PollEvery time.Duration // How often to poll for new data
	ReadChunk int           // Read buffer size per iteration
}

// Tailer tails a single file with simple polling. Cross-platform (Windows/macOS/Linux).
type Tailer struct {
	opt Options
	mu  sync.Mutex
	f   *os.File
	pos int64
	st  os.FileInfo
	ctx context.Context
	can context.CancelFunc
}

func New(opt Options) *Tailer {
	if opt.PollEvery <= 0 {
		opt.PollEvery = 300 * time.Millisecond
	}
	if opt.ReadChunk <= 0 {
		opt.ReadChunk = 64 * 1024
	}
	return &Tailer{opt: opt}
}

// Start begins tailing. It sends complete lines on out. It returns when ctx is done or on fatal error.
func (t *Tailer) Start(ctx context.Context, out chan<- string) error {
	if t.opt.Path == "" {
		return errors.New("tailer: empty path")
	}
	ctx, cancel := context.WithCancel(ctx)
	t.ctx = ctx
	t.can = cancel
	defer cancel()

	retryDelay := 500 * time.Millisecond
	var pending []byte

	openFile := func() error {
		f, err := os.Open(t.opt.Path)
		if err != nil {
			return err
		}
		st, err := f.Stat()
		if err != nil {
			f.Close()
			return err
		}
		var startPos int64
		if t.opt.FromStart {
			startPos = 0
		} else {
			startPos = st.Size()
		}
		if _, err := f.Seek(startPos, io.SeekStart); err != nil {
			f.Close()
			return err
		}
		t.mu.Lock()
		t.f = f
		t.st = st
		t.pos = startPos
		t.mu.Unlock()
		return nil
	}

	// Wait for file to exist if needed
	for {
		if err := openFile(); err != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryDelay):
				continue
			}
		}
		break
	}

	buf := make([]byte, t.opt.ReadChunk)
	reader := bufio.NewReaderSize(nil, t.opt.ReadChunk)

	flushLines := func(b []byte) {
		// Split on \n; handle Windows \r\n
		start := 0
		for i := 0; i < len(b); i++ {
			if b[i] == '\n' {
				line := b[start:i]
				// trim trailing \r
				if len(line) > 0 && line[len(line)-1] == '\r' {
					line = line[:len(line)-1]
				}
				out <- string(line)
				start = i + 1
			}
		}
		pending = append(pending[:0], b[start:]...)
	}

	// readLoop
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(t.opt.PollEvery):
			// fallthrough
		}

		t.mu.Lock()
		f := t.f
		st := t.st
		pos := t.pos
		t.mu.Unlock()

		// Detect rotation or truncation
		curSt, err := os.Stat(t.opt.Path)
		if err != nil {
			// File temporarily missing; try reopen later
			t.closeFile()
			continue
		}
		rotated := false
		if !sameFile(st, curSt) {
			rotated = true
		} else if curSt.Size() < pos {
			rotated = true // truncated
		}
		if rotated {
			t.closeFile()
			if err := openFile(); err != nil {
				// wait and retry next tick
				continue
			}
			pending = pending[:0]
			reader.Reset(t.f)
			continue
		}

		if f == nil {
			if err := openFile(); err != nil {
				continue
			}
			reader.Reset(t.f)
			continue
		}

		// Read new bytes
		n, err := f.Read(buf)
		if errors.Is(err, io.EOF) {
			// no new data yet
			continue
		}
		if err != nil && !errors.Is(err, io.EOF) {
			// transient read error, try reopening
			t.closeFile()
			continue
		}
		if n > 0 {
			data := append(pending, buf[:n]...)
			flushLines(data)
			t.mu.Lock()
			t.pos += int64(n)
			t.mu.Unlock()
		}
	}
}

func (t *Tailer) closeFile() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.f != nil {
		_ = t.f.Close()
	}
	t.f = nil
	t.st = nil
	t.pos = 0
}

// Stop cancels the tailing context if started.
func (t *Tailer) Stop() {
	if t.can != nil {
		t.can()
	}
}

// sameFile reports whether a and b refer to the same file identity.
func sameFile(a, b os.FileInfo) bool {
	if a == nil || b == nil {
		return false
	}
	return os.SameFile(a, b)
}
