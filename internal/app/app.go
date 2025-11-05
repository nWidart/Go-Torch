package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"GoTorch/internal/parser"
	"GoTorch/internal/pricing"
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

	// tracking lifecycle timestamps & pause state (app-level session control)
	trackStartedAt   time.Time
	trackStoppedAt   time.Time
	trackPaused      bool
	trackPausedAt    time.Time
	trackPausedAccum time.Duration
	lastEventAt      time.Time // last parsed event time, used to clamp durations when parsing old logs

	// item table loaded from full_table.json (or embedded fallback)
	items       map[string]ItemInfo
	itemsSource string // diagnostic: where items were loaded from
}

func New() *App {
	return &App{trk: tracker.New(), p: parser.New()}
}

// Startup is called by Wails when the app starts.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	// Load item metadata table on startup
	a.loadItemTable()
	// Refresh prices from remote endpoint with a short timeout; ignore errors.
	a.refreshPrices()
}

// Shutdown is called by Wails when the app terminates.
func (a *App) Shutdown(ctx context.Context) {
	a.Stop()
}

// loadItemTable attempts to load full_table.json from common locations, with env override and embedded fallback.
func (a *App) loadItemTable() {
	// default to empty map
	a.items = map[string]ItemInfo{}
	a.itemsSource = "none"

	readItemFile := func(path string) (map[string]ItemInfo, bool) {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, false
		}
		var m map[string]ItemInfo
		if err := json.Unmarshal(b, &m); err != nil || len(m) == 0 {
			return nil, false
		}
		return m, true
	}

	// 1) Environment variable override
	if p := os.Getenv("GOTORCH_ITEM_TABLE"); p != "" {
		if m, ok := readItemFile(p); ok {
			a.items = m
			a.itemsSource = "env:" + p
			goto DONE
		}
	}
	// 2) Working directory
	if m, ok := readItemFile("full_table.json"); ok {
		a.items = m
		a.itemsSource = "file:./full_table.json"
		goto DONE
	}
	// 3) Executable directory
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		if m, ok := readItemFile(filepath.Join(dir, "full_table.json")); ok {
			a.items = m
			a.itemsSource = "exe_dir:full_table.json"
			goto DONE
		}
	}
	// 4) Embedded fallback
	if b, err := readEmbeddedItemTable(); err == nil && len(b) > 0 {
		var m map[string]ItemInfo
		if err := json.Unmarshal(b, &m); err == nil && len(m) > 0 {
			a.items = m
			a.itemsSource = "embedded"
			goto DONE
		}
	}

DONE:
	if a.isWailsContext() {
		runtime.LogInfof(a.ctx, "item table loaded (%d items) from %s", len(a.items), a.itemsSource)
	}
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
	// mark session start now; clear pause/stop state
	a.trackStartedAt = time.Now()
	a.trackStoppedAt = time.Time{}
	a.trackPaused = false
	a.trackPausedAt = time.Time{}
	a.trackPausedAccum = 0

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
					a.mu.Lock()
					a.lastEventAt = ev.Time
					a.mu.Unlock()
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
				if a.isWailsContext() {
					runtime.EventsEmit(a.ctx, "state", a.UIState())
				}
			}
		}
	}()

	return nil
}

// Stop tracking and background goroutines. Also marks the session as stopped.
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
	// finalize pause accumulation if paused
	now := time.Now()
	if a.trackPaused && !a.trackPausedAt.IsZero() {
		a.trackPausedAccum += now.Sub(a.trackPausedAt)
		a.trackPaused = false
		a.trackPausedAt = time.Time{}
	}
	// mark stopped
	if !a.trackStartedAt.IsZero() && a.trackStoppedAt.IsZero() {
		a.trackStoppedAt = now
	}
}

// Reset clears the current tracker state (does not stop tracking).
func (a *App) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.trk = tracker.New()
	// also reset session timing state
	a.trackStartedAt = time.Time{}
	a.trackStoppedAt = time.Time{}
	a.trackPaused = false
	a.trackPausedAt = time.Time{}
	a.trackPausedAccum = 0
}

// PauseSession pauses the app-level session timer.
func (a *App) PauseSession() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.trackStartedAt.IsZero() || a.trackPaused || !a.trackStoppedAt.IsZero() {
		return
	}
	a.trackPaused = true
	a.trackPausedAt = time.Now()
}

// ResumeSession resumes the app-level session timer.
func (a *App) ResumeSession() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.trackPaused || a.trackPausedAt.IsZero() {
		return
	}
	now := time.Now()
	a.trackPausedAccum += now.Sub(a.trackPausedAt)
	a.trackPaused = false
	a.trackPausedAt = time.Time{}
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
	// snapshot app-level session fields
	a.mu.Lock()
	startedAt := a.trackStartedAt
	stoppedAt := a.trackStoppedAt
	paused := a.trackPaused
	pausedAt := a.trackPausedAt
	pausedAccum := a.trackPausedAccum
	lastEv := a.lastEventAt
	a.mu.Unlock()

	// Build UI tally by enriching with metadata; include unknown IDs as placeholders
	uiTally := make(map[string]UITallyItem)
	for id, n := range st.Current.Tally {
		key := intToStr(id)
		if a.items != nil {
			if info, ok := a.items[key]; ok {
				uiTally[key] = UITallyItem{
					Name:       info.Name,
					Type:       info.Type,
					Price:      info.Price,
					LastUpdate: info.LastUpdate,
					From:       info.From,
					Count:      n,
				}
				continue
			}
		}
		// Fallback: show unknown IDs so the tally is visible even without item table
		uiTally[key] = UITallyItem{
			Name:  "#" + key,
			Type:  "Unknown",
			Price: 0,
			Count: n,
		}
	}
	// Compute per-map earnings for completed maps + current
	maps := make([]UIMap, 0, len(st.Completed)+1)
	var sessionEarnings float64
	var totalMapDurMs int64
	for _, m := range st.Completed {
		// only include maps that belong to this session (started after we clicked start), unless session not started
		if !startedAt.IsZero() && m.StartedAt.Before(startedAt) {
			continue
		}
		var earn float64
		for id, c := range m.Tally {
			key := intToStr(id)
			if info, ok := a.items[key]; ok {
				earn += float64(c) * info.Price
			}
		}
		durMs := m.EndedAt.Sub(m.StartedAt).Milliseconds()
		maps = append(maps, UIMap{Start: m.StartedAt.UnixMilli(), End: m.EndedAt.UnixMilli(), DurationMs: durMs, Earnings: earn})
		sessionEarnings += earn
		totalMapDurMs += durMs
	}
	// current map earnings
	var currentEarn float64
	for id, c := range st.Current.Tally {
		key := intToStr(id)
		if info, ok := a.items[key]; ok {
			currentEarn += float64(c) * info.Price
		}
	}
	// include current map as last entry only if active and belongs to this session
	includedCurrent := false
	if st.Current.Active && !st.Current.StartedAt.IsZero() && (startedAt.IsZero() || !st.Current.StartedAt.Before(startedAt)) {
		end := time.Now()
		if !lastEv.IsZero() && lastEv.After(st.Current.StartedAt) {
			end = lastEv // clamp to last event time to avoid huge durations when parsing old logs
		}
		durMs := end.Sub(st.Current.StartedAt).Milliseconds()
		maps = append(maps, UIMap{Start: st.Current.StartedAt.UnixMilli(), End: 0, DurationMs: durMs, Earnings: currentEarn})
		includedCurrent = true
	}
	if includedCurrent {
		sessionEarnings += currentEarn
	}
	// compute earnings per hour over session active time (excluding pauses)
	var sessionStartMs int64
	var sessionEndMs int64 // exported end timestamp; 0 while session is active (running or paused)
	var pausedAccumMs int64
	var calcEndMs int64
	if !startedAt.IsZero() {
		sessionStartMs = startedAt.UnixMilli()
		pausedAccumMs = pausedAccum.Milliseconds()

		// Determine end timestamp for calculations
		if !stoppedAt.IsZero() {
			calcEndMs = stoppedAt.UnixMilli()
			sessionEndMs = stoppedAt.UnixMilli()
		} else if paused && !pausedAt.IsZero() {
			calcEndMs = pausedAt.UnixMilli()
			sessionEndMs = 0
		} else {
			calcEndMs = time.Now().UnixMilli()
			sessionEndMs = 0
		}
	}
	var eph float64
	if sessionStartMs > 0 && calcEndMs > sessionStartMs {
		durActiveMs := (calcEndMs - sessionStartMs) - pausedAccumMs
		if durActiveMs > 0 {
			eph = sessionEarnings / (float64(durActiveMs) / 3600000.0)
		}
	}
	// average time per completed map (only those included)
	var avgMapMs int64
	if len(maps) > 0 {
		// compute from maps that have end>0 (completed)
		var cnt int64
		for _, m := range maps {
			if m.End > 0 {
				avgMapMs += m.DurationMs
				cnt++
			}
		}
		if cnt > 0 {
			avgMapMs = avgMapMs / cnt
		}
	}
	// convert recent events
	recent := make([]UIEvent, 0, len(st.LastEvents))
	for _, ev := range st.LastEvents {
		recent = append(recent, UIEvent{Time: ev.Time.UnixMilli(), Kind: ev.Kind.String()})
	}
	return UIState{
		InMap:         st.InMap && st.Current.Active,
		SessionStart:  sessionStartMs,
		SessionEnd:    sessionEndMs,
		SessionPaused: paused,
		PausedAt: func() int64 {
			if paused {
				return pausedAt.UnixMilli()
			}
			return int64(0)
		}(),
		PausedAccumMs:      pausedAccumMs,
		MapStart:           st.Current.StartedAt.UnixMilli(),
		MapEnd:             st.Current.EndedAt.UnixMilli(),
		TotalDrops:         st.TotalDrops,
		Tally:              uiTally,
		Recent:             recent,
		Maps:               maps,
		EarningsPerSession: sessionEarnings,
		EarningsPerHour:    eph,
		AvgMapTimeMs:       avgMapMs,
	}
}

// ItemInfo represents an item entry from full_table.json
type ItemInfo struct {
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	Price      float64 `json:"price"`
	LastUpdate float64 `json:"last_update"`
	From       string  `json:"from"`
}

// UITallyItem is sent to the frontend for each counted item id
type UITallyItem struct {
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	Price      float64 `json:"price"`
	LastUpdate float64 `json:"last_update"`
	From       string  `json:"from"`
	Count      int     `json:"count"`
}

type UIMap struct {
	Start      int64   `json:"start"`
	End        int64   `json:"end"`
	DurationMs int64   `json:"durationMs"`
	Earnings   float64 `json:"earnings"`
}

type UIState struct {
	InMap              bool                   `json:"inMap"`
	SessionStart       int64                  `json:"sessionStart"`
	SessionEnd         int64                  `json:"sessionEnd"`
	SessionPaused      bool                   `json:"sessionPaused"`
	PausedAt           int64                  `json:"pausedAt"`
	PausedAccumMs      int64                  `json:"pausedAccumMs"`
	MapStart           int64                  `json:"mapStart"`
	MapEnd             int64                  `json:"mapEnd"`
	TotalDrops         int                    `json:"totalDrops"`
	Tally              map[string]UITallyItem `json:"tally"`
	Recent             []UIEvent              `json:"recent"`
	Maps               []UIMap                `json:"maps"`
	EarningsPerSession float64                `json:"earningsPerSession"`
	EarningsPerHour    float64                `json:"earningsPerHour"`
	AvgMapTimeMs       int64                  `json:"avgMapTimeMs"`
}

type UIEvent struct {
	Time int64  `json:"time"`
	Kind string `json:"kind"`
}

// ItemTableSource returns the source used to load the item table and item count.
func (a *App) ItemTableSource() (string, int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.itemsSource, len(a.items)
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

// refreshPrices fetches remote pricing and merges price + last_update into the in-memory items.
func (a *App) refreshPrices() {
	if a.ctx == nil || a.items == nil {
		return
	}
	ctx, cancel := context.WithTimeout(a.ctx, 5*time.Second)
	defer cancel()
	updates, err := pricing.FetchRemotePrices(ctx, "")
	if err != nil {
		if a.isWailsContext() {
			runtime.LogWarningf(a.ctx, "price refresh failed: %v", err)
		}
		return
	}
	var changed int
	var total int
	a.mu.Lock()
	for id, u := range updates {
		total++
		if info, ok := a.items[id]; ok {
			if info.Price != u.Price || info.LastUpdate != u.LastUpdate {
				info.Price = u.Price
				info.LastUpdate = u.LastUpdate
				a.items[id] = info
				changed++
			}
		}
	}
	a.mu.Unlock()
	if a.isWailsContext() {
		runtime.LogInfof(a.ctx, "price refresh: %d updated (from %d remote items)", changed, total)
	}
}

// isWailsContext checks if the context is valid for Wails runtime calls.
func (a *App) isWailsContext() bool {
	// In tests, a.ctx is context.Background() which is not a Wails context
	// We can check by attempting to use it, but simpler is to check if it has
	// a specific value set by Wails. For now, we'll assume if ctx is non-nil
	// and has a value, it's likely a Wails context.
	// The safest way is to use a defer/recover around runtime calls.
	return a.ctx != nil && a.ctx != context.Background() && a.ctx != context.TODO()
}
