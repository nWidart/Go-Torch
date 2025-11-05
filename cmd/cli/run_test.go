package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunMissingLogArg(t *testing.T) {
	code := run([]string{})
	if code != 2 {
		t.Fatalf("expected exit code 2 for missing args, got %d", code)
	}
}

func TestRunOnceSucceeds(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "ue.log")
	data := "" +
		"[2025.11.04-19.20.45:474][302]GameLog: Display: [Game] PageApplyBase@ _UpdateGameEnd: LastSceneName = /Game/Art/Maps/UI/LoginScene/LoginScene NextSceneName = World'/Game/Art/Maps/07YJ/YJ_YongZhouHuiLang200/YJ_YongZhouHuiLang200.YJ_YongZhouHuiLang200'\n" +
		"[2025.11.04-19.20.46:000][302]GameLog: Display: [Game] BagMgr@:InitBagData PageId = 1 SlotId = 1 ConfigBaseId = 1001 Num = 0\n" +
		"[2025.11.04-19.20.47:000][302]GameLog: Display: [Game] BagMgr@:Modfy BagItem PageId = 1 SlotId = 1 ConfigBaseId = 1001 Num = 2\n"
	if err := os.WriteFile(p, []byte(data), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	code := run([]string{"--log", p, "--once", "--debug=false"})
	if code != 0 {
		t.Fatalf("expected 0, got %d", code)
	}
}

func TestRunEnvExpansion(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "ue.log")
	if err := os.WriteFile(p, []byte("\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	_ = os.Setenv("FOOPATH", p)
	code := run([]string{"--log", "$FOOPATH", "--once", "--debug=false"})
	if code != 0 {
		t.Fatalf("expected 0 with env-expanded path, got %d", code)
	}
}
