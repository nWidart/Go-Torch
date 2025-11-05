package main

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"testing"

	"GoTorch/internal/parser"
	"GoTorch/internal/tracker"
)

func TestChanReaderBasicAndCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := make(chan string, 1)
	r := readerFromChan(ctx, ch)

	// Send a message and read with larger buffer to trigger newline injection
	ch <- "hello"
	buf := make([]byte, 10)
	n, err := r.Read(buf)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if n != 6 { // "hello" + "\n"
		t.Fatalf("read n=%d want 6", n)
	}
	if string(buf[:n]) != "hello\n" {
		t.Fatalf("got %q", string(buf[:n]))
	}

	// Now cancel context and expect error
	cancel()
	_, err = r.Read(buf)
	if err == nil {
		t.Fatal("expected error on canceled context")
	}
}

func TestChanReaderClosedChannel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := make(chan string)
	r := readerFromChan(ctx, ch)
	close(ch)
	buf := make([]byte, 10)
	_, err := r.Read(buf)
	if err == nil {
		t.Fatal("expected error when channel is closed")
	}
}

func TestProcessOnceOnTempLog(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "log.log")
	f, err := os.Create(p)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	w := bufio.NewWriter(f)
	// Craft a small sequence: MapStart -> BagInit 0 -> BagMod 3 -> MapEnd
	lines := []string{
		"[2025.11.04-19.20.45:474][302]GameLog: Display: [Game] PageApplyBase@ _UpdateGameEnd: LastSceneName = /Game/Art/Maps/UI/LoginScene/LoginScene NextSceneName = World'/Game/Art/Maps/07YJ/YJ_YongZhouHuiLang200/YJ_YongZhouHuiLang200.YJ_YongZhouHuiLang200'\n",
		"[2025.11.04-19.20.46:000][302]GameLog: Display: [Game] BagMgr@:InitBagData PageId = 1 SlotId = 1 ConfigBaseId = 999 Num = 0\n",
		"[2025.11.04-19.20.47:000][302]GameLog: Display: [Game] BagMgr@:Modfy BagItem PageId = 1 SlotId = 1 ConfigBaseId = 999 Num = 3\n",
		"[2025.11.04-19.20.48:000][302]GameLog: Display: [Game] PageApplyBase@ _UpdateGameEnd: LastSceneName = World'/Game/Art/Maps/07YJ/YJ_YongZhouHuiLang200/YJ_YongZhouHuiLang200.YJ_YongZhouHuiLang200' NextSceneName = World'/Game/Art/Maps/01SD/XZ_YuJinZhiXiBiNanSuo200/XZ_YuJinZhiXiBiNanSuo200.XZ_YuJinZhiXiBiNanSuo200'\n",
	}
	for _, s := range lines {
		_, _ = w.WriteString(s)
	}
	_ = w.Flush()
	_ = f.Close()

	pzr := parser.New()
	trk := tracker.New()
	if err := processOnce(p, pzr, trk, false); err != nil {
		t.Fatalf("processOnce: %v", err)
	}
	st := trk.GetState()
	if st.TotalDrops != 3 || st.Current.Tally[999] != 3 {
		t.Fatalf("unexpected tracker state: total=%d tally=%d", st.TotalDrops, st.Current.Tally[999])
	}
	if st.InMap {
		t.Fatalf("expected map ended at the end of log")
	}
}
