package pricing

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchRemotePrices_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]PriceUpdate{
			"1001": {Price: 12.5, LastUpdate: 1700000000},
		})
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	m, err := FetchRemotePrices(ctx, ts.URL+"/get")
	if err != nil {
		t.Fatalf("FetchRemotePrices error: %v", err)
	}
	if len(m) != 1 || m["1001"].Price != 12.5 || m["1001"].LastUpdate != 1700000000 {
		t.Fatalf("unexpected map: %#v", m)
	}
}

func TestFetchRemotePrices_Non200(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("bad"))
	}))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := FetchRemotePrices(ctx, ts.URL)
	if err == nil {
		t.Fatal("expected error on non-200 status")
	}
}

func TestFetchRemotePrices_InvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{"))
	}))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := FetchRemotePrices(ctx, ts.URL)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestWithTimeout_DefaultAndCustom(t *testing.T) {
	start := time.Now()
	ctx, cancel := WithTimeout(context.Background(), 0)
	defer cancel()
	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected deadline to be set")
	}
	d := time.Until(deadline)
	if d < 4*time.Second || d > 6*time.Second {
		t.Fatalf("expected ~5s timeout, got %v (deadline %v, start %v)", d, deadline, start)
	}

	ctx2, cancel2 := WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel2()
	deadline2, ok2 := ctx2.Deadline()
	if !ok2 {
		t.Fatal("expected deadline to be set for custom duration")
	}
	d2 := time.Until(deadline2)
	if d2 < 1*time.Second || d2 > 2*time.Second {
		t.Fatalf("expected ~1.5s timeout, got %v", d2)
	}
}
