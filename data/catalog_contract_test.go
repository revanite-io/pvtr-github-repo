package data

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gemaraproj/go-gemara"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
)

func TestSupportedCatalogIDsExist(t *testing.T) {
	// Keep the declared compatibility contract in sync with bundled catalog data.
	catalogDir := "catalogs"
	entries, err := os.ReadDir(catalogDir)
	if err != nil {
		t.Fatalf("failed to read catalog directory: %v", err)
	}

	foundCatalogIDs := make(map[string]string, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		catalogPath := filepath.Join(catalogDir, entry.Name())
		data, err := os.ReadFile(catalogPath)
		if err != nil {
			t.Fatalf("failed to read catalog %s: %v", entry.Name(), err)
		}

		var catalog gemara.ControlCatalog
		if err := yaml.Unmarshal(data, &catalog); err != nil {
			t.Fatalf("failed to parse catalog %s: %v", entry.Name(), err)
		}

		foundCatalogIDs[catalog.Metadata.Id] = entry.Name()
	}

	for _, catalogID := range SupportedCatalogIDs {
		assert.Contains(t, foundCatalogIDs, catalogID, "supported catalog ID %s is missing from data/catalogs", catalogID)
	}
}
