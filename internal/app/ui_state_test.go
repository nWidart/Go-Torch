package app

import (
	"testing"
	"time"

	"GoTorch/internal/tracker"
	"GoTorch/internal/types"
)

// Basic smoke test for UI state conversion; now relies on app-level session timer.
func TestUIStateConversion(t *testing.T) {
	a := New()
	// Provide minimal item metadata for item 5210
	a.items = map[string]ItemInfo{
		"5210": {Name: "Test Item", Type: "test", Price: 1.0},
	}
	// Simulate that the user clicked Start Session 5s ago
	start := time.Now().Add(-5 * time.Second)
	a.trackStartedAt = start

	// build some state via tracker events (within the session)
	a.trk.OnEvent(&types.Event{Kind: types.EventMapStart, Time: start})
	a.trk.OnEvent(&types.Event{Kind: types.EventBagInit, Time: start, Bag: &types.BagEvent{PageID: 1, SlotID: 1, ConfigBaseID: 5210, Num: 1}})
	// +2 during map
	a.trk.OnEvent(&types.Event{Kind: types.EventBagMod, Time: start.Add(time.Second), Bag: &types.BagEvent{PageID: 1, SlotID: 1, ConfigBaseID: 5210, Num: 3}})

	st := a.UIState()
	if !st.InMap {
		t.Fatalf("expected in map")
	}
	if st.TotalDrops != 2 {
		t.Fatalf("expected total drops 2, got %d", st.TotalDrops)
	}
	if st.Tally["5210"].Count != 2 {
		t.Fatalf("expected tally[5210]=2, got %d", st.Tally["5210"].Count)
	}
	if st.SessionStart == 0 {
		t.Fatalf("expected non-zero session start")
	}
	if st.SessionEnd != 0 {
		t.Fatalf("expected SessionEnd=0 while running, got %d", st.SessionEnd)
	}
	if len(st.Recent) == 0 {
		t.Fatalf("expected some recent events")
	}

	// Reset should clear state and session timer
	a.Reset()
	st2 := a.UIState()
	if st2.TotalDrops != 0 || st2.InMap {
		t.Fatalf("expected cleared state after reset: %+v", st2)
	}
	if st2.SessionStart != 0 || st2.SessionEnd != 0 {
		t.Fatalf("expected cleared session timing after reset: %+v", st2)
	}

	// Use tracker directly and compare GetState wrapper
	tr := tracker.New()
	a.trk = tr
	// simulate a new session start as well
	a.trackStartedAt = time.Now()
	tr.OnEvent(&types.Event{Kind: types.EventMapStart, Time: time.Now()})
	st3a := a.GetState()
	st3b := a.UIState()
	if st3a.InMap != st3b.InMap {
		t.Fatalf("GetState and UIState should agree")
	}
}

// Verify pause/resume fields and EPH exclude paused time.
func TestPauseResumeAndEPH(t *testing.T) {
	a := New()
	// one item with $1 price
	a.items = map[string]ItemInfo{
		"1": {Name: "X", Type: "t", Price: 1.0},
	}
	now := time.Now()
	// session started 10s ago
	a.trackStartedAt = now.Add(-10 * time.Second)
	// no stop, not paused currently; but previously paused for 4s
	a.trackPausedAccum = 4 * time.Second

	// Start a map inside the session and add +6 items so earnings=6
	ms := now.Add(-8 * time.Second)
	a.trk.OnEvent(&types.Event{Kind: types.EventMapStart, Time: ms})
	// inventory snapshot 1 item, then set to 7 so delta=6
	a.trk.OnEvent(&types.Event{Kind: types.EventBagInit, Time: ms, Bag: &types.BagEvent{PageID: 1, SlotID: 1, ConfigBaseID: 1, Num: 1}})
	a.trk.OnEvent(&types.Event{Kind: types.EventBagMod, Time: ms.Add(1 * time.Second), Bag: &types.BagEvent{PageID: 1, SlotID: 1, ConfigBaseID: 1, Num: 7}})

	st := a.UIState()
	if st.SessionPaused {
		t.Fatalf("should not be paused")
	}
	if st.PausedAccumMs < 3900 || st.PausedAccumMs > 4100 {
		t.Fatalf("expected ~4000ms paused accum, got %d", st.PausedAccumMs)
	}
	if st.EarningsPerSession < 5.9 || st.EarningsPerSession > 6.1 {
		t.Fatalf("expected earnings per session ~6, got %f", st.EarningsPerSession)
	}
	// Active time ~6s => EPH ~ 3600
	if st.EarningsPerHour < 3000 || st.EarningsPerHour > 4200 {
		t.Fatalf("expected EPH around 3600, got %f", st.EarningsPerHour)
	}

	// Now pause and ensure SessionEnd stays 0 and paused flags set
	a.trackPaused = true
	a.trackPausedAt = now
	st2 := a.UIState()
	if !st2.SessionPaused || st2.PausedAt == 0 {
		t.Fatalf("expected paused state with pausedAt set")
	}
	if st2.SessionEnd != 0 {
		t.Fatalf("expected SessionEnd=0 while paused, got %d", st2.SessionEnd)
	}
}

// Verify current map duration is clamped to lastEventAt when parsing historical logs.
func TestCurrentMapDurationClamped(t *testing.T) {
	a := New()
	now := time.Now()
	// session started a while ago
	a.trackStartedAt = now.Add(-1 * time.Minute)
	// a map started 30s ago
	mapStart := now.Add(-30 * time.Second)
	a.trk.OnEvent(&types.Event{Kind: types.EventMapStart, Time: mapStart})
	// last parsed event was 2s after map start
	a.lastEventAt = mapStart.Add(2 * time.Second)

	st := a.UIState()
	// find the current map (last entry, End==0)
	if len(st.Maps) == 0 {
		t.Fatalf("expected at least one map entry")
	}
	cur := st.Maps[len(st.Maps)-1]
	if cur.End != 0 {
		t.Fatalf("expected current map End=0")
	}
	if cur.DurationMs < 1500 || cur.DurationMs > 2500 {
		t.Fatalf("expected clamped duration around 2s, got %dms", cur.DurationMs)
	}
}

// SessionEnd should be set after Stop, and SessionPaused should be false.
func TestSessionEndAfterStop(t *testing.T) {
	a := New()
	now := time.Now()
	a.trackStartedAt = now.Add(-3 * time.Second)
	// Simulate paused just before stopping; Stop should finalize pause and clear paused flag.
	a.trackPaused = true
	a.trackPausedAt = now.Add(-1 * time.Second)

	a.Stop()
	st := a.UIState()
	if st.SessionEnd == 0 {
		t.Fatalf("expected SessionEnd > 0 after stop")
	}
	if st.SessionStart == 0 {
		t.Fatalf("expected SessionStart > 0 after stop")
	}
	if st.SessionEnd <= st.SessionStart {
		t.Fatalf("expected SessionEnd (%d) > SessionStart (%d)", st.SessionEnd, st.SessionStart)
	}
	if st.SessionPaused {
		t.Fatalf("expected SessionPaused=false after stop")
	}
}
