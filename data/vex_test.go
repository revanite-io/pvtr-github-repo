package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsVexPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"product.vex.json", true},
		{"openvex.json", true},
		{"my-vex.yaml", true},
		{"security/report.vex.yml", true},
		{"vex/bom.json", true},
		{"docs/vex/product.json", true},
		{"bom.cdx-vex.json", true},
		{"app.cyclonedx-vex.json", true},
		{"advisory.csaf.json", true},
		{"data.csaf", true},
		{"README.md", false},
		{"sbom.json", false},
		{"convex.json", false},
		{"vexing.json", false},
		{"src/vex.go", false},
		{"vex/helper.go", false},
		{"config.yaml", false},
		// VEX tooling source/docs must not be mistaken for a VEX document.
		{"openvex.go", false},
		{"openvex.md", false},
		{"docs/openvex.png", false},
		{"pkg/cyclonedx-vex.go", false},
		{"cdx-vex-writer.py", false},
		{"README-openvex.txt", false},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			assert.Equal(t, test.want, isVexPath(test.path))
		})
	}
}

func TestDetectVexDocuments(t *testing.T) {
	t.Run("nil tree returns nil", func(t *testing.T) {
		assert.Nil(t, detectVexDocuments(nil))
	})

	t.Run("finds vex documents and ignores others", func(t *testing.T) {
		tree := buildTree([]testEntry{
			{name: "README.md"},
			{name: "product.vex.json"},
			{name: "vex/bom.json"},
			{name: "src/main.go"},
		})

		found := detectVexDocuments(tree)
		assert.ElementsMatch(t, []string{"product.vex.json", "vex/bom.json"}, found)
	})

	t.Run("no vex documents returns empty", func(t *testing.T) {
		tree := buildTree([]testEntry{
			{name: "README.md"},
			{name: "go.mod"},
		})

		assert.Empty(t, detectVexDocuments(tree))
	})

	t.Run("returns full path for nested vex document", func(t *testing.T) {
		tree := buildTreeWithNested(
			[]testEntry{{name: "README.md"}},
			[]testEntry{{name: "product.vex.json"}, {name: "notes.md"}},
		)

		found := detectVexDocuments(tree)
		assert.Equal(t, []string{"subdir/product.vex.json"}, found)
	})
}
