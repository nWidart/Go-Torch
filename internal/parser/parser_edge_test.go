package parser

import (
	"testing"
	"time"

	"GoTorch/internal/types"
)

func TestParseCRLFAndWhitespace(t *testing.T) {
	p := New()
	line := "[2025.11.04-19.21.45:100][111]GameLog: Display: [Game] BagMgr@:Modfy BagItem PageId = 9 SlotId = 9 ConfigBaseId = 123 Num = 1\r\n"
	if ev := p.Parse(line); ev == nil || ev.Kind != types.EventBagMod {
		t.Fatalf("expected BagMod event from CRLF line, got %#v", ev)
	}
}

func TestParseNonMapTransitionIgnored(t *testing.T) {
	p := New()
	line := "[2025.11.04-19.20.45:474][302]GameLog: Display: [Game] PageApplyBase@ _UpdateGameEnd: LastSceneName = /Game/Art/Maps/UI/LoginScene/LoginScene NextSceneName = World'/NotMaps/Somewhere'"
	if ev := p.Parse(line); ev != nil {
		t.Fatalf("expected nil event for non-map transition, got %#v", ev)
	}
}

func TestTimestampMissingFallsBack(t *testing.T) {
	p := New()
	// Intentionally use a non-timestamp bracket so ts parser fails, but regex prefix still matches
	line := "[NOTATS]GameLog: Display: [Game] BagMgr@:Modfy BagItem PageId = 1 SlotId = 1 ConfigBaseId = 1 Num = 2"
	start := time.Now().Add(-1 * time.Second)
	ev := p.Parse(line)
	if ev == nil {
		t.Fatal("expected event")
	}
	if ev.Time.Before(start) || ev.Time.After(time.Now().Add(2*time.Second)) {
		t.Fatalf("fallback time out of expected range: %v", ev.Time)
	}
}

func TestPrefixWithMultipleBrackets(t *testing.T) {
	p := New()
	line := "[2025.11.04-19.21.45:100][302][Foo]GameLog: Display: [Game] BagMgr@:InitBagData PageId = 3 SlotId = 4 ConfigBaseId = 100 Num = 5"
	if ev := p.Parse(line); ev == nil || ev.Kind != types.EventBagInit {
		t.Fatalf("expected BagInit with multiple bracket prefix, got %#v", ev)
	}
}
