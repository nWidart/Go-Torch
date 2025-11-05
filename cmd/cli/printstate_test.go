package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"GoTorch/internal/tracker"
	"GoTorch/internal/types"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String()
}

func TestPrintState_NoSessionYet(t *testing.T) {
	trk := tracker.New()
	out := captureStdout(t, func() { printState(trk) })
	if !strings.Contains(out, "No session yet.") {
		t.Fatalf("expected 'No session yet.' in output\n%s", out)
	}
}

func TestPrintState_ActiveAndEndedSession(t *testing.T) {
	trk := tracker.New()
	start := time.Now().Add(-3 * time.Second)
	trk.OnEvent(&types.Event{Kind: types.EventMapStart, Time: start})
	// +5 items
	trk.OnEvent(&types.Event{Kind: types.EventBagInit, Time: start, Bag: &types.BagEvent{PageID: 1, SlotID: 1, ConfigBaseID: 42, Num: 0}})
	trk.OnEvent(&types.Event{Kind: types.EventBagMod, Time: start.Add(time.Second), Bag: &types.BagEvent{PageID: 1, SlotID: 1, ConfigBaseID: 42, Num: 5}})

	// Active session print
	outActive := captureStdout(t, func() { printState(trk) })
	if !strings.Contains(outActive, "Status: In Map") {
		t.Fatalf("expected 'In Map' in output\n%s", outActive)
	}
	if !strings.Contains(outActive, "Tally:") {
		t.Fatalf("expected Tally in output\n%s", outActive)
	}

	// End session and print again to hit ended branch and items/hour
	end := start.Add(2 * time.Second)
	trk.OnEvent(&types.Event{Kind: types.EventMapEnd, Time: end})
	outEnded := captureStdout(t, func() { printState(trk) })
	if !strings.Contains(outEnded, "Status: Idle") {
		t.Fatalf("expected 'Idle' in output\n%s", outEnded)
	}
	if !strings.Contains(outEnded, "Items/hour:") {
		t.Fatalf("expected Items/hour in output\n%s", outEnded)
	}
}
