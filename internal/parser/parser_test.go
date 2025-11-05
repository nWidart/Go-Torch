package parser

import (
	"testing"
	"time"

	"GoTorch/internal/types"
)

func TestParseBagInitAndBagMod(t *testing.T) {
	p := New()
	// Example lines inspired by UE log
	lineInit := "[2025.11.04-19.20.45:474][302]GameLog: Display: [Game] BagMgr@:InitBagData PageId = 1 SlotId = 2 ConfigBaseId = 5210 Num = 5"
	ev := p.Parse(lineInit)
	if ev == nil || ev.Kind != types.EventBagInit {
		t.Fatalf("expected BagInit event, got %#v", ev)
	}
	if ev.Bag == nil || ev.Bag.PageID != 1 || ev.Bag.SlotID != 2 || ev.Bag.ConfigBaseID != 5210 || ev.Bag.Num != 5 {
		t.Fatalf("unexpected bag payload: %#v", ev.Bag)
	}

	lineMod := "[2025.11.04-19.21.45:100][111]GameLog: Display: [Game] BagMgr@:Modfy BagItem PageId = 1 SlotId = 2 ConfigBaseId = 5210 Num = 7"
	ev2 := p.Parse(lineMod)
	if ev2 == nil || ev2.Kind != types.EventBagMod {
		t.Fatalf("expected BagMod event, got %#v", ev2)
	}
}

func TestParseMapStartAndEnd(t *testing.T) {
	p := New()
	start := "[2025.11.04-19.20.45:474][302]GameLog: Display: [Game] PageApplyBase@ _UpdateGameEnd: LastSceneName = /Game/Art/Maps/UI/LoginScene/LoginScene NextSceneName = World'/Game/Art/Maps/07YJ/YJ_YongZhouHuiLang200/YJ_YongZhouHuiLang200.YJ_YongZhouHuiLang200'"
	if ev := p.Parse(start); ev == nil || ev.Kind != types.EventMapStart {
		t.Fatalf("expected MapStart, got %#v", ev)
	}
	end := "[2025.11.04-19.26.24:480][420]GameLog: Display: [Game] PageApplyBase@ _UpdateGameEnd: LastSceneName = World'/Game/Art/Maps/07YJ/YJ_YongZhouHuiLang200/YJ_YongZhouHuiLang200.YJ_YongZhouHuiLang200' NextSceneName = World'" + refugePath + "'"
	if ev := p.Parse(end); ev == nil || ev.Kind != types.EventMapEnd {
		t.Fatalf("expected MapEnd, got %#v", ev)
	}
}

func TestTimestampParsing(t *testing.T) {
	p := New()
	line := "[2025.11.04-19.20.45:463][302]GameLog: Display: [Game] Something else"
	ev := p.Parse(line)
	if ev != nil {
		// Unrecognized line should return nil; but we can still test parseTimestamp directly via a wrapper
	}
	// Use unexported parser method through known line patterns that will return an event and carry the parsed time
	lineMod := "[2025.11.04-19.21.45:100][111]GameLog: Display: [Game] BagMgr@:Modfy BagItem PageId = 1 SlotId = 2 ConfigBaseId = 5210 Num = 7"
	ev2 := p.Parse(lineMod)
	if ev2 == nil {
		t.Fatalf("expected event, got nil")
	}
	// Expect year=2025, month=11, day=04, hour=19, minute=21
	got := ev2.Time
	if got.Year() != 2025 || got.Month() != time.November || got.Day() != 4 || got.Hour() != 19 || got.Minute() != 21 {
		t.Fatalf("unexpected parsed timestamp: %v", got)
	}
}
