package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAppStartTrackingFromStartCountsDeltas(t *testing.T) {
	// Prepare a temp log with a MapStart and a bag delta during the map
	dir := t.TempDir()
	p := filepath.Join(dir, "ue.log")
	lines := "" +
		"[2025.11.04-19.20.45:474][302]GameLog: Display: [Game] PageApplyBase@ _UpdateGameEnd: LastSceneName = /Game/Art/Maps/UI/LoginScene/LoginScene NextSceneName = World'/Game/Art/Maps/07YJ/YJ_YongZhouHuiLang200/YJ_YongZhouHuiLang200.YJ_YongZhouHuiLang200'\n" +
		"[2025.11.04-19.20.46:000][302]GameLog: Display: [Game] BagMgr@:InitBagData PageId = 1 SlotId = 1 ConfigBaseId = 1001 Num = 0\n" +
		"[2025.11.04-19.20.47:000][302]GameLog: Display: [Game] BagMgr@:Modfy BagItem PageId = 1 SlotId = 1 ConfigBaseId = 1001 Num = 4\n"
	if err := os.WriteFile(p, []byte(lines), 0o644); err != nil {
		t.Fatalf("write log: %v", err)
	}

	a := New()
	ctx := context.Background()
	a.Startup(ctx)
	defer a.Stop()

	if err := a.StartTrackingWithOptions(p, true); err != nil {
		t.Fatalf("StartTrackingWithOptions: %v", err)
	}
	// Wait for default poll (300ms) + processing
	time.Sleep(900 * time.Millisecond)
	st := a.UIState()
	if !st.InMap {
		t.Fatalf("expected in map, got %+v", st)
	}
	if st.TotalDrops != 4 {
		t.Fatalf("expected total drops 4, got %d", st.TotalDrops)
	}
	if st.Tally["1001"] != 4 {
		t.Fatalf("expected tally[1001]=4, got %d", st.Tally["1001"])
	}

	// Reset clears state
	a.Reset()
	st2 := a.UIState()
	if st2.TotalDrops != 0 || st2.InMap || len(st2.Tally) != 0 {
		t.Fatalf("expected cleared state after reset: %+v", st2)
	}
}

func TestAppStopResetIdempotent(t *testing.T) {
	a := New()
	a.Startup(context.Background())
	// Multiple calls should not panic
	a.Stop()
	a.Stop()
	a.Reset()
	a.Reset()
}
