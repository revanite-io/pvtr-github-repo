package data

import (
	"testing"
)

func TestCheckTreeForBinaries(t *testing.T) {
	tests := []struct {
		name     string
		tree     *GraphqlRepoTree
		expected []string
	}{
		{
			name:     "nil tree returns nil",
			tree:     nil,
			expected: nil,
		},
		{
			name:     "empty tree returns no binaries",
			tree:     &GraphqlRepoTree{},
			expected: nil,
		},
		{
			name:     "text files are not flagged as binary",
			tree:     buildTestTree([]testFile{{name: "README.md"}, {name: "LICENSE"}, {name: "OWNERS"}, {name: "Tiltfile"}}),
			expected: nil,
		},
		{
			name:     "binary files are correctly detected",
			tree:     buildTestTree([]testFile{{name: "app.jar", binary: true}, {name: "README.md"}}),
			expected: []string{"app.jar"},
		},
		{
			name:     "multiple binary files detected",
			tree:     buildTestTree([]testFile{{name: "app.exe", binary: true}, {name: "lib.dll", binary: true}, {name: "main.go"}}),
			expected: []string{"app.exe", "lib.dll"},
		},
		{
			name:     "nested binary files at level 2",
			tree:     buildTestTreeNested(nil, []testFile{{name: "wrapper.jar", binary: true}}, nil),
			expected: []string{"wrapper.jar"},
		},
		{
			name:     "nested binary files at level 3",
			tree:     buildTestTreeNested(nil, nil, []testFile{{name: "deep.exe", binary: true}, {name: "data.bin", binary: true}}),
			expected: []string{"deep.exe", "data.bin"},
		},
		{
			name:     "binaries at all levels",
			tree:     buildTestTreeNested([]testFile{{name: "root.jar", binary: true}}, []testFile{{name: "level2.dll", binary: true}}, []testFile{{name: "level3.exe", binary: true}}),
			expected: []string{"root.jar", "level2.dll", "level3.exe"},
		},
		{
			name: "extensionless text files not flagged (the bug we fixed)",
			tree: buildTestTree([]testFile{
				{name: "OWNERS"}, {name: "OWNERS_ALIASES"}, {name: "SECURITY_CONTACTS"},
				{name: "Tiltfile"}, {name: "dockerignore"}, {name: "TECHNICAL_ADVISORY_MEMBERS"},
			}),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkTreeForBinaries(tt.tree)

			if len(result) != len(tt.expected) {
				t.Errorf("got %d binaries, want %d\ngot: %v\nwant: %v",
					len(result), len(tt.expected), result, tt.expected)
				return
			}

			for i, name := range tt.expected {
				if result[i] != name {
					t.Errorf("binary[%d] = %q, want %q", i, result[i], name)
				}
			}
		})
	}
}

type testFile struct {
	name   string
	binary bool
}

// buildTestTree creates a GraphqlRepoTree with files at root level
func buildTestTree(files []testFile) *GraphqlRepoTree {
	tree := &GraphqlRepoTree{}
	for _, f := range files {
		entry := newLevel1Entry(f.name, "blob")
		entry.Object.Blob.IsBinary = f.binary
		tree.Repository.Object.Tree.Entries = append(tree.Repository.Object.Tree.Entries, entry)
	}
	return tree
}

// buildTestTreeNested creates a tree with files at all 3 levels
func buildTestTreeNested(level1, level2, level3 []testFile) *GraphqlRepoTree {
	tree := buildTestTree(level1)

	if len(level2) > 0 || len(level3) > 0 {
		dir := newLevel1Entry("subdir", "tree")

		for _, f := range level2 {
			entry := newLevel2Entry(f.name, "blob")
			entry.Object.Blob.IsBinary = f.binary
			dir.Object.Tree.Entries = append(dir.Object.Tree.Entries, entry)
		}

		if len(level3) > 0 {
			subdir := newLevel2Entry("deep", "tree")
			for _, f := range level3 {
				entry := newLevel3Entry(f.name, "blob")
				entry.Object.Blob.IsBinary = f.binary
				subdir.Object.Tree.Entries = append(subdir.Object.Tree.Entries, entry)
			}
			dir.Object.Tree.Entries = append(dir.Object.Tree.Entries, subdir)
		}

		tree.Repository.Object.Tree.Entries = append(tree.Repository.Object.Tree.Entries, dir)
	}
	return tree
}

// Entry type aliases for readability
type (
	level1Entry = struct {
		Name   string
		Type   string
		Path   string
		Object *struct {
			Blob struct{ IsBinary bool } `graphql:"... on Blob"`
			Tree struct {
				Entries []level2Entry
			} `graphql:"... on Tree"`
		} `graphql:"object"`
	}
	level2Entry = struct {
		Name   string
		Type   string
		Path   string
		Object *struct {
			Blob struct{ IsBinary bool } `graphql:"... on Blob"`
			Tree struct {
				Entries []level3Entry
			} `graphql:"... on Tree"`
		} `graphql:"object"`
	}
	level3Entry = struct {
		Name   string
		Type   string
		Path   string
		Object *struct {
			Blob struct{ IsBinary bool } `graphql:"... on Blob"`
		} `graphql:"object"`
	}
)

func newLevel1Entry(name, typ string) level1Entry {
	e := level1Entry{Name: name, Type: typ, Path: name}
	e.Object = &struct {
		Blob struct{ IsBinary bool } `graphql:"... on Blob"`
		Tree struct {
			Entries []level2Entry
		} `graphql:"... on Tree"`
	}{}
	return e
}

func newLevel2Entry(name, typ string) level2Entry {
	e := level2Entry{Name: name, Type: typ, Path: name}
	e.Object = &struct {
		Blob struct{ IsBinary bool } `graphql:"... on Blob"`
		Tree struct {
			Entries []level3Entry
		} `graphql:"... on Tree"`
	}{}
	return e
}

func newLevel3Entry(name, typ string) level3Entry {
	e := level3Entry{Name: name, Type: typ, Path: name}
	e.Object = &struct {
		Blob struct{ IsBinary bool } `graphql:"... on Blob"`
	}{}
	return e
}
