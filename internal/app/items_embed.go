package app

import "embed"

// embeddedItemsFS contains the fallback item table bundled into the binary.
//
//go:embed data/full_table.json
var embeddedItemsFS embed.FS

// readEmbeddedItemTable returns the embedded full_table.json bytes.
func readEmbeddedItemTable() ([]byte, error) {
	return embeddedItemsFS.ReadFile("data/full_table.json")
}
