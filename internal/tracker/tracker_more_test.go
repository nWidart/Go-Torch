package tracker

import (
	"testing"
	"time"

	"GoTorch/internal/types"
)

func TestLastEventsBufferTrim(t *testing.T) {
	trk := New()
	start := time.Now().Add(-200 * time.Second)
	for i := 0; i < 120; i++ {
		// Use BagInit with distinct timestamps
		trk.OnEvent(&types.Event{Kind: types.EventBagInit, Time: start.Add(time.Duration(i) * time.Second), Bag: &types.BagEvent{PageID: 1, SlotID: i, ConfigBaseID: 1, Num: i}})
	}
	st := trk.GetState()
	if len(st.LastEvents) != 100 {
		t.Fatalf("expected last events length 100, got %d", len(st.LastEvents))
	}
	// Oldest should correspond to i=20 (0..119 -> kept 20..119)
	oldest := st.LastEvents[0]
	wantTime := start.Add(20 * time.Second)
	if !oldest.Time.Equal(wantTime) {
		t.Fatalf("oldest event time = %v; want %v", oldest.Time, wantTime)
	}
}

func TestGetStateDeepCopyIsolation(t *testing.T) {
	trk := New()
	// Seed inventory and tally with a simple in-map increment
	cfg := 777
	now := time.Now()
	trk.OnEvent(&types.Event{Kind: types.EventBagInit, Time: now, Bag: &types.BagEvent{PageID: 1, SlotID: 1, ConfigBaseID: cfg, Num: 1}})
	trk.OnEvent(&types.Event{Kind: types.EventMapStart, Time: now.Add(time.Second)})
	trk.OnEvent(&types.Event{Kind: types.EventBagMod, Time: now.Add(2 * time.Second), Bag: &types.BagEvent{PageID: 1, SlotID: 1, ConfigBaseID: cfg, Num: 3}})

	st := trk.GetState()
	// Mutate returned copy
	st.Inventory[slotKey{PageID: 1, SlotID: 1, ConfigBaseID: cfg}] = 999
	st.Current.Tally[cfg] = 999
	if st.LastEvents != nil && len(st.LastEvents) > 0 {
		st.LastEvents[0].Kind = types.EventUnknown
	}

	// Fetch again; internal state should be unchanged
	st2 := trk.GetState()
	if st2.Inventory[slotKey{PageID: 1, SlotID: 1, ConfigBaseID: cfg}] != 3 { // last Num was 3
		t.Fatalf("internal inventory mutated via shallow copy; got %v", st2.Inventory)
	}
	if st2.Current.Tally[cfg] != 2 {
		t.Fatalf("internal tally mutated; got %d", st2.Current.Tally[cfg])
	}
	if len(st2.LastEvents) == 0 || st2.LastEvents[0].Kind == types.EventUnknown {
		t.Fatalf("internal last events mutated via copy")
	}
}

func TestSessionRestartCreatesNewSession(t *testing.T) {
	trk := New()
	start1 := time.Now()
	start2 := start1.Add(10 * time.Second)
	trk.OnEvent(&types.Event{Kind: types.EventMapStart, Time: start1})
	trk.OnEvent(&types.Event{Kind: types.EventMapStart, Time: start2})
	st := trk.GetState()
	if !st.InMap || !st.Current.Active {
		t.Fatalf("expected active session after restart")
	}
	if !st.Current.StartedAt.Equal(start2) {
		t.Fatalf("expected session start at %v, got %v", start2, st.Current.StartedAt)
	}
}
