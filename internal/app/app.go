package app

import (
	"context"
	"sync"
	"time"

	"GoTorch/internal/parser"
	"GoTorch/internal/tailer"
	"GoTorch/internal/tracker"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App exposes methods to Wails frontend and manages the tracking lifecycle.
type App struct {
	mu  sync.Mutex
	ctx context.Context

	trk      *tracker.Tracker
	p        *parser.Parser
	t        *tailer.Tailer
	lines    chan string
	cancel   context.CancelFunc
	emitStop context.CancelFunc
}

func New() *App {
	return &App{trk: tracker.New(), p: parser.New()}
}

// Startup is called by Wails when the app starts.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// Shutdown is called by Wails when the app terminates.
func (a *App) Shutdown(ctx context.Context) {
	a.Stop()
}

// StartTracking starts tailing the given log path and emitting state updates to the UI.
// By default, it tails from the end (does not read historical lines).
func (a *App) StartTracking(logPath string) error {
	return a.startTrackingInternal(logPath, false)
}

// StartTrackingWithOptions allows the caller to control whether to read from the start.
// Set fromStart=true during development to process historical lines; false in production.
func (a *App) StartTrackingWithOptions(logPath string, fromStart bool) error {
	return a.startTrackingInternal(logPath, fromStart)
}

func (a *App) startTrackingInternal(logPath string, fromStart bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// stop previous session if any
	if a.cancel != nil {
		a.cancel()
		a.cancel = nil
	}
	if a.emitStop != nil {
		a.emitStop()
		a.emitStop = nil
	}

	ctx, cancel := context.WithCancel(a.ctx)
	a.cancel = cancel

	lines := make(chan string, 2048)
	a.lines = lines
	a.t = tailer.New(tailer.Options{Path: logPath, FromStart: fromStart})
	// start tailer
	go func() {
		_ = a.t.Start(ctx, lines)
	}()
	// start reader + parser
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-lines:
				if !ok {
					return
				}
				if ev := a.p.Parse(line); ev != nil {
					a.trk.OnEvent(ev)
				}
			}
		}
	}()

	// Periodic state emitter (1s)
	emitCtx, emitStop := context.WithCancel(a.ctx)
	a.emitStop = emitStop
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-emitCtx.Done():
				return
			case <-ticker.C:
				runtime.EventsEmit(a.ctx, "state", a.UIState())
			}
		}
	}()

	return nil
}

// Stop tracking and background goroutines.
func (a *App) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cancel != nil {
		a.cancel()
		a.cancel = nil
	}
	if a.emitStop != nil {
		a.emitStop()
		a.emitStop = nil
	}
	if a.t != nil {
		a.t.Stop()
	}
}

// Reset clears the current tracker state (does not stop tracking).
func (a *App) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.trk = tracker.New()
}

// SelectLogFile opens a file dialog and returns the selected log file path.
func (a *App) SelectLogFile() (string, error) {
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select UE_game.log",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Log Files (*.log)",
				Pattern:     "*.log",
			},
		},
	})
	return selection, err
}

// GetState returns the latest state snapshot for the UI to pull on demand.
func (a *App) GetState() UIState {
	return a.UIState()
}

// UIState converts internal tracker state to a JSON-friendly struct for the UI.
func (a *App) UIState() UIState {
	st := a.trk.GetState()
	// convert tally keys to strings
	uiTally := make(map[string]int, len(st.Current.Tally))
	for id, n := range st.Current.Tally {
		uiTally[intToStr(id)] = n
	}
	// convert recent events
	recent := make([]UIEvent, 0, len(st.LastEvents))
	for _, ev := range st.LastEvents {
		recent = append(recent, UIEvent{Time: ev.Time.UnixMilli(), Kind: ev.Kind.String()})
	}
	return UIState{
		InMap:        st.InMap && st.Current.Active,
		SessionStart: st.Current.StartedAt.UnixMilli(),
		SessionEnd:   st.Current.EndedAt.UnixMilli(),
		TotalDrops:   st.TotalDrops,
		Tally:        uiTally,
		Recent:       recent,
	}
}

type UIState struct {
	InMap        bool           `json:"inMap"`
	SessionStart int64          `json:"sessionStart"`
	SessionEnd   int64          `json:"sessionEnd"`
	TotalDrops   int            `json:"totalDrops"`
	Tally        map[string]int `json:"tally"`
	Recent       []UIEvent      `json:"recent"`
}

type UIEvent struct {
	Time int64  `json:"time"`
	Kind string `json:"kind"`
}

func intToStr(n int) string {
	// Fast int to string without fmt to avoid allocations; ok to use fmt if preferred
	// but here use std conversion for clarity
	return strconvItoa(n)
}

// minimal itoa to avoid importing fmt
func strconvItoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
