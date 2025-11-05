package tracker

import (
	"testing"
	"time"

	"GoTorch/internal/types"
)

func TestTrackerSessionAndDeltas(t *testing.T) {
	trk := New()
	cfg := 5210
	// BagInit before map should set inventory but not count
	trk.OnEvent(&types.Event{Kind: types.EventBagInit, Time: time.Now(), Bag: &types.BagEvent{PageID: 1, SlotID: 2, ConfigBaseID: cfg, Num: 5}})
	st := trk.GetState()
	if st.TotalDrops != 0 || st.Current.Tally[cfg] != 0 || st.InMap {
		t.Fatalf("unexpected state before map: %+v", st)
	}
	// Start map
	start := time.Now()
	trk.OnEvent(&types.Event{Kind: types.EventMapStart, Time: start})
	// Increase via BagMod: +3 (5->8)
	trk.OnEvent(&types.Event{Kind: types.EventBagMod, Time: start.Add(time.Second), Bag: &types.BagEvent{PageID: 1, SlotID: 2, ConfigBaseID: cfg, Num: 8}})
	// Decrease (8->6) should not count
	trk.OnEvent(&types.Event{Kind: types.EventBagMod, Time: start.Add(2 * time.Second), Bag: &types.BagEvent{PageID: 1, SlotID: 2, ConfigBaseID: cfg, Num: 6}})
	st = trk.GetState()
	if st.TotalDrops != 3 || st.Current.Tally[cfg] != 3 {
		t.Fatalf("expected +3 counted during map, got total=%d tally=%d", st.TotalDrops, st.Current.Tally[cfg])
	}
	if !st.InMap || !st.Current.Active {
		t.Fatalf("expected in-map active session")
	}
	// End map
	end := start.Add(3 * time.Second)
	trk.OnEvent(&types.Event{Kind: types.EventMapEnd, Time: end})
	st = trk.GetState()
	if st.InMap || st.Current.Active {
		t.Fatalf("expected session ended")
	}
	if st.Current.EndedAt.IsZero() || !st.Current.EndedAt.Equal(end) {
		t.Fatalf("expected ended at to be set; got %v", st.Current.EndedAt)
	}
	// Post-map increment should not count
	trk.OnEvent(&types.Event{Kind: types.EventBagMod, Time: end.Add(time.Second), Bag: &types.BagEvent{PageID: 1, SlotID: 2, ConfigBaseID: cfg, Num: 9}})
	st = trk.GetState()
	if st.TotalDrops != 3 || st.Current.Tally[cfg] != 3 {
		t.Fatalf("post-map increments should not count; got total=%d tally=%d", st.TotalDrops, st.Current.Tally[cfg])
	}
}
