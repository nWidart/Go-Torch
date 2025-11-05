package app

import (
	"testing"
	"time"

	"GoTorch/internal/tracker"
	"GoTorch/internal/types"
)

func TestUIStateConversion(t *testing.T) {
	a := New()
	// build some state via tracker events
	start := time.Now().Add(-5 * time.Second)
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
	if st.Tally["5210"] != 2 {
		t.Fatalf("expected tally[5210]=2, got %d", st.Tally["5210"])
	}
	if st.SessionStart == 0 {
		t.Fatalf("expected non-zero session start")
	}
	if len(st.Recent) == 0 {
		t.Fatalf("expected some recent events")
	}

	// Reset should clear state
	a.Reset()
	st2 := a.UIState()
	if st2.TotalDrops != 0 || st2.InMap {
		t.Fatalf("expected cleared state after reset: %+v", st2)
	}

	// Use tracker directly and compare GetState wrapper
	tr := tracker.New()
	a.trk = tr
	tr.OnEvent(&types.Event{Kind: types.EventMapStart, Time: time.Now()})
	st3a := a.GetState()
	st3b := a.UIState()
	if st3a.InMap != st3b.InMap {
		t.Fatalf("GetState and UIState should agree")
	}
}
