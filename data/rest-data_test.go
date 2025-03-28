package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckFile(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name     string
		filename string
		toplevel []DirContents
		forgedir []DirContents
		expected string
	}{
		{
			name:     "finds support.md in root",
			filename: "support.md",
			toplevel: []DirContents{
				{
					Name: "support.md",
					Path: "support.md",
				},
				{
					Name: "readme.md",
					Path: "readme.md",
				},
				{
					Name: "other.md",
					Path: "other.md",
				},
			},
			forgedir: []DirContents{},
			expected: "support.md",
		},
		{
			name:     "finds readme.md in root",
			filename: "readme.md",
			toplevel: []DirContents{
				{
					Name: "support.md",
					Path: "support.md",
				},
				{
					Name: "readme.md",
					Path: "readme.md",
				},
				{
					Name: "other.md",
					Path: "other.md",
				},
			},
			forgedir: []DirContents{},
			expected: "readme.md",
		},
		{
			name:     "case insensitive match",
			filename: "readme.md",
			toplevel: []DirContents{
				{
					Name: "support.md",
					Path: "support.md",
				},
				{
					Name: "README.md",
					Path: "README.md",
				},
				{
					Name: "other.md",
					Path: "other.md",
				},
			},
			forgedir: []DirContents{},
			expected: "README.md",
		},
		{
			name:     "finds support.md in forge dir",
			filename: "support.md",
			toplevel: []DirContents{},
			forgedir: []DirContents{
				{
					Name: "support.md",
					Path: ".github/support.md",
				},
				{
					Name: "readme.md",
					Path: ".github/readme.md",
				},
			},
			expected: ".github/support.md",
		},
		{
			name:     "file not found",
			filename: "nonexistent.md",
			toplevel: []DirContents{},
			forgedir: []DirContents{},
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
