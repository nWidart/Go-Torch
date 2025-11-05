package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"GoTorch/internal/pricing"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("updateprices", flag.ContinueOnError)
	file := fs.String("file", "full_table.json", "Path to the item table JSON file to update")
	endpoint := fs.String("endpoint", "", "Pricing endpoint URL (defaults to built-in)")
	dryRun := fs.Bool("dry-run", false, "Do not write changes, only report what would change")
	backup := fs.Bool("backup", true, "Create a .bak backup before writing")
	timeout := fs.Duration("timeout", 8*time.Second, "HTTP timeout for the pricing request")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		usage(fs)
		return 2
	}

	path := strings.TrimSpace(*file)
	if path == "" {
		fmt.Fprintln(os.Stderr, "--file is required")
		usage(fs)
		return 2
	}

	// Load existing JSON as a generic map to preserve unknown fields
	existing, err := loadJSON(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to read file:", err)
		return 1
	}

	// Fetch remote prices
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	updates, err := pricing.FetchRemotePrices(ctx, *endpoint)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to fetch remote pricing:", err)
		return 1
	}

	// Merge updates
	changed, total := mergePrices(existing, updates)
	fmt.Printf("Remote items: %d, Updated entries: %d\n", total, changed)

	if *dryRun {
		fmt.Println("Dry-run: no changes written.")
		return 0
	}

	if *backup {
		if err := writeBackup(path); err != nil {
			fmt.Fprintln(os.Stderr, "warning: could not create backup:", err)
		}
	}
	if err := writeJSON(path, existing); err != nil {
		fmt.Fprintln(os.Stderr, "failed to write file:", err)
		return 1
	}
	fmt.Println("Prices updated successfully:", path)
	return 0
}

func usage(fs *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "Usage: %s [--file full_table.json] [--endpoint URL] [--dry-run] [--backup=true] [--timeout 8s]\n", fs.Name())
}

// loadJSON reads the file path into a generic nested map.
func loadJSON(path string) (map[string]map[string]interface{}, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m map[string]map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// writeJSON writes the map back to disk with indentation.
func writeJSON(path string, m map[string]map[string]interface{}) error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}

func writeBackup(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !fi.Mode().IsRegular() {
		return errors.New("not a regular file: " + path)
	}
	bak := path + ".bak"
	// if existing, add timestamp suffix
	if _, err := os.Stat(bak); err == nil {
		ext := filepath.Ext(path)
		base := strings.TrimSuffix(filepath.Base(path), ext)
		dir := filepath.Dir(path)
		bak = filepath.Join(dir, fmt.Sprintf("%s.%d%s.bak", base, time.Now().Unix(), ext))
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return os.WriteFile(bak, data, 0644)
}

// mergePrices applies Price and LastUpdate from updates into existing map, only for existing ids.
func mergePrices(existing map[string]map[string]interface{}, updates map[string]pricing.PriceUpdate) (changed int, total int) {
	for id, u := range updates {
		total++
		if row, ok := existing[id]; ok {
			var prevPrice, prevLU interface{}
			if v, ok := row["price"]; ok {
				prevPrice = v
			}
			if v, ok := row["last_update"]; ok {
				prevLU = v
			}
			// Always set price and last_update
			row["price"] = u.Price
			row["last_update"] = u.LastUpdate
			// detect changes conservatively
			if !equalJSONNumber(prevPrice, u.Price) || !equalJSONNumber(prevLU, u.LastUpdate) {
				changed++
			}
		}
	}
	return changed, total
}

func equalJSONNumber(prev interface{}, now float64) bool {
	switch v := prev.(type) {
	case float64:
		return v == now
	case json.Number:
		f, err := v.Float64()
		if err == nil {
			return f == now
		}
		return false
	default:
		return false
	}
}
