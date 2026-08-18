package main

import "testing"

// The publish metadata (Publisher, License, catalogNamespaces) is inert at run
// time, so nothing else in this repo notices when it is wrong or when a newly
// embedded catalog is missing a namespace — the release publish is the first
// thing that would fail. Assert the manifest `pvtr publish` reads instead.
func TestPublishManifest(t *testing.T) {
	orchestrator, err := newOrchestrator("1.2.3")
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
	if len(manifest.Evaluates) == 0 {
		t.Fatal("expected at least one evaluated catalog")
	}
	for _, e := range manifest.Evaluates {
		if len(e.Catalog) < len("ossf/") || e.Catalog[:len("ossf/")] != "ossf/" {
			t.Errorf("catalog %q is not namespaced under ossf", e.Catalog)
		}
	}
}
