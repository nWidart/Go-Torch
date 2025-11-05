package pricing

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// DefaultEndpoint is the default pricing endpoint used to refresh prices.
const DefaultEndpoint = "http://serverp.furtorch.heili.tech/get"

// PriceUpdate holds only the fields we care about updating.
type PriceUpdate struct {
	Price      float64 `json:"price"`
	LastUpdate float64 `json:"last_update"`
}

// FetchRemotePrices fetches the remote pricing map keyed by item id.
// The context should contain a timeout/deadline.
func FetchRemotePrices(ctx context.Context, endpoint string) (map[string]PriceUpdate, error) {
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	// Set a short UA, in case server requires it later
	req.Header.Set("User-Agent", "GoTorch/price-updater")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("pricing endpoint non-200: " + resp.Status)
	}
	var out map[string]PriceUpdate
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	if err := dec.Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

// WithTimeout creates a child context with the specified timeout, defaulting to 5s if d==0.
func WithTimeout(parent context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if d <= 0 {
		d = 5 * time.Second
	}
	return context.WithTimeout(parent, d)
}
