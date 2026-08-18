package main

import (
	"strings"
	"testing"
)

// The publish metadata (Publisher, License, catalogNamespaces) is inert at run
// time, so nothing else in this repo notices when it is wrong or when a newly
// embedded catalog is missing a namespace — the release publish is the first
// thing that would fail. Assert the manifest `pvtr publish` reads instead.
func TestPublishManifest(t *testing.T) {
	orchestrator, err := newOrchestrator()
	if err != nil {
		t.Fatalf("newOrchestrator: %v", err)
	}

	manifest, err := orchestrator.PublishManifest()
	if err != nil {
		t.Fatalf("PublishManifest: %v", err)
	}
	if manifest.Coordinate != "ossf/github-repo" {
		t.Errorf("coordinate = %q, want ossf/github-repo", manifest.Coordinate)
	}
	if manifest.License != "Apache-2.0" {
		t.Errorf("license = %q, want Apache-2.0", manifest.License)
	}
	if len(manifest.Evaluates) != len(catalogNamespaces) {
		t.Fatalf("evaluated %d catalogs, want %d", len(manifest.Evaluates), len(catalogNamespaces))
	}
	for _, e := range manifest.Evaluates {
		if !strings.HasPrefix(e.Catalog, "ossf/") {
			t.Errorf("catalog %q is not namespaced under ossf", e.Catalog)
		}
	}
}
