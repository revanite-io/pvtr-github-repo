package data

import (
	"testing"

	"github.com/google/go-github/v71/github"
	"github.com/stretchr/testify/assert"
)

func TestCheckFile(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name     string
		filename string
		toplevel []*github.RepositoryContent
		forgedir []*github.RepositoryContent
		expected string
	}{
		{
			name:     "finds support.md in root",
			filename: "support.md",
			toplevel: []*github.RepositoryContent{
				{
					Type: github.Ptr("file"),
					Name: github.Ptr("support.md"),
					Path: github.Ptr("support.md"),
				},
				{
					Type: github.Ptr("file"),
					Name: github.Ptr("readme.md"),
					Path: github.Ptr("readme.md"),
				},
				{
					Type: github.Ptr("file"),
					Name: github.Ptr("other.md"),
					Path: github.Ptr("other.md"),
				},
			},
			forgedir: []*github.RepositoryContent{},
			expected: "support.md",
		},
		{
			name:     "finds readme.md in root",
			filename: "readme.md",
			toplevel: []*github.RepositoryContent{
				{
					Type: github.Ptr("file"),
					Name: github.Ptr("support.md"),
					Path: github.Ptr("support.md"),
				},
				{
					Type: github.Ptr("file"),
					Name: github.Ptr("readme.md"),
					Path: github.Ptr("readme.md"),
				},
				{
					Type: github.Ptr("file"),
					Name: github.Ptr("other.md"),
					Path: github.Ptr("other.md"),
				},
			},
			forgedir: []*github.RepositoryContent{},
			expected: "readme.md",
		},
		{
			name:     "case insensitive match",
			filename: "readme.md",
			toplevel: []*github.RepositoryContent{
				{
					Type: github.Ptr("file"),
					Name: github.Ptr("support.md"),
					Path: github.Ptr("support.md"),
				},
				{
					Type: github.Ptr("file"),
					Name: github.Ptr("README.md"),
					Path: github.Ptr("README.md"),
				},
				{
					Type: github.Ptr("file"),
					Name: github.Ptr("other.md"),
					Path: github.Ptr("other.md"),
				},
			},
			forgedir: []*github.RepositoryContent{},
			expected: "README.md",
		},
		{
			name:     "finds support.md in forge dir",
			filename: "support.md",
			toplevel: []*github.RepositoryContent{},
			forgedir: []*github.RepositoryContent{
				{
					Type: github.Ptr("file"),
					Name: github.Ptr("support.md"),
					Path: github.Ptr(".github/support.md"),
				},
				{
					Type: github.Ptr("file"),
					Name: github.Ptr("readme.md"),
					Path: github.Ptr(".github/readme.md"),
				},
			},
			expected: ".github/support.md",
		},
		{
			name:     "file not found",
			filename: "nonexistent.md",
			toplevel: []*github.RepositoryContent{},
			forgedir: []*github.RepositoryContent{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RestData{
				Contents: Contents{
					TopLevel: tt.toplevel,
					ForgeDir: tt.forgedir,
				},
			}
			result := r.checkFile(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}
