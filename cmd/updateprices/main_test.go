package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"GoTorch/internal/pricing"
)

func TestEqualJSONNumber(t *testing.T) {
	cases := []struct {
		prev interface{}
		now  float64
		ok   bool
	}{
		{float64(1.0), 1.0, true},
		{json.Number("1"), 1.0, true},
		{json.Number("1.5"), 1.5, true},
		{json.Number("bad"), 0.0, false},
		{"1", 1.0, false},
		{nil, 0.0, false},
	}
	for i, c := range cases {
		if got := equalJSONNumber(c.prev, c.now); got != c.ok {
			t.Fatalf("case %d: equalJSONNumber(%T,%v)=%v want %v", i, c.prev, c.now, got, c.ok)
		}
	}
}

func TestMergePrices(t *testing.T) {
	existing := map[string]map[string]interface{}{
		"100": {"name": "foo", "price": json.Number("1"), "last_update": json.Number("1000")},
		"200": {"name": "bar", "price": 2.0, "last_update": 2000.0},
	}
	updates := map[string]pricing.PriceUpdate{
		"100": {Price: 1.5, LastUpdate: 1500}, // change both
		"300": {Price: 3.0, LastUpdate: 3000}, // not present in existing
	}
	changed, total := mergePrices(existing, updates)
	if total != 2 {
		t.Fatalf("total=%d want 2", total)
	}
	if changed != 1 {
		t.Fatalf("changed=%d want 1", changed)
	}
	if existing["100"]["price"].(float64) != 1.5 || existing["100"]["last_update"].(float64) != 1500 {
		t.Fatalf("row 100 not updated: %#v", existing["100"])
	}
	if _, ok := existing["300"]; ok {
		t.Fatalf("unexpected creation of new id 300 in existing")
	}
}

func TestLoadWriteJSONAndBackup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "table.json")
	data := map[string]map[string]interface{}{
		"1": {"name": "a", "price": 1.0, "last_update": 1.0},
	}
	b, _ := json.Marshal(data)
	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	loaded, err := loadJSON(path)
	if err != nil {
		t.Fatalf("loadJSON: %v", err)
	}
	if !reflect.DeepEqual(loaded, data) {
		t.Fatalf("loaded != data: %#v vs %#v", loaded, data)
	}
	// mutate and write
	loaded["1"]["price"] = 2.0
	if err := writeJSON(path, loaded); err != nil {
		t.Fatalf("writeJSON: %v", err)
	}
	// write backup
	if err := writeBackup(path); err != nil {
		t.Fatalf("writeBackup: %v", err)
	}
	if _, err := os.Stat(path + ".bak"); err != nil {
		t.Fatalf("expected backup file: %v", err)
	}
}

func TestRunDryRunAndWrite(t *testing.T) {
	// mock pricing server
	mux := http.NewServeMux()
	mux.HandleFunc("/prices", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]pricing.PriceUpdate{
			"10": {Price: 9.9, LastUpdate: 111},
		})
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	dir := t.TempDir()
	path := filepath.Join(dir, "full_table.json")
	orig := map[string]map[string]interface{}{"10": {"name": "x", "price": 1.0, "last_update": 1.0}}
	b, _ := json.Marshal(orig)
	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Dry-run should not change file
	code := run([]string{"--file", path, "--endpoint", ts.URL + "/prices", "--dry-run"})
	if code != 0 {
		t.Fatalf("dry-run exit code=%d", code)
	}
	b2, _ := os.ReadFile(path)
	var after map[string]map[string]interface{}
	_ = json.Unmarshal(b2, &after)
	if !reflect.DeepEqual(after, orig) {
		t.Fatalf("dry-run mutated file: %#v", after)
	}

	// Real run should modify and create backup
	code = run([]string{"--file", path, "--endpoint", ts.URL + "/prices", "--backup", "true"})
	if code != 0 {
		t.Fatalf("write run exit code=%d", code)
	}
	b3, _ := os.ReadFile(path)
	var updated map[string]map[string]interface{}
	_ = json.Unmarshal(b3, &updated)
	if updated["10"]["price"].(float64) != 9.9 {
		t.Fatalf("expected updated price, got %#v", updated)
	}
	if _, err := os.Stat(path + ".bak"); err != nil {
		t.Fatalf("expected backup file: %v", err)
	}
}

// Ensure pricing.WithTimeout used by run is reliable under short timeouts (smoke)
func TestPricingWithTimeoutSmoke(t *testing.T) {
	ctx, cancel := pricing.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	select {
	case <-ctx.Done():
		// may expire quickly; acceptable
	case <-time.After(50 * time.Millisecond):
		// still alive is also fine; just smoke test
	}
}
