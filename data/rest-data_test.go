package data

import (
	"testing"

	"github.com/google/go-github/v71/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
)

func TestCheckFile(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		toplevel  []*github.RepositoryContent
		githubDir []*github.RepositoryContent
		expected  string
	}{
		{
			name:     "finds support.md in root",
			filename: "support.md",
			toplevel: []*github.RepositoryContent{
				{Type: github.Ptr("file"), Name: github.Ptr("support.md"), Path: github.Ptr("support.md")},
				{Type: github.Ptr("file"), Name: github.Ptr("readme.md"), Path: github.Ptr("readme.md")},
			},
			githubDir: []*github.RepositoryContent{},
			expected:  "support.md",
		},
		{
			name:     "finds readme.md in root",
			filename: "readme.md",
			toplevel: []*github.RepositoryContent{
				{Type: github.Ptr("file"), Name: github.Ptr("readme.md"), Path: github.Ptr("readme.md")},
			},
			githubDir: []*github.RepositoryContent{},
			expected:  "readme.md",
		},
		{
			name:     "case insensitive match",
			filename: "readme.md",
			toplevel: []*github.RepositoryContent{
				{Type: github.Ptr("file"), Name: github.Ptr("README.md"), Path: github.Ptr("README.md")},
			},
			githubDir: []*github.RepositoryContent{},
			expected:  "README.md",
		},
		{
			name:     "finds support.md in .github",
			filename: "support.md",
			toplevel: []*github.RepositoryContent{},
			githubDir: []*github.RepositoryContent{
				{Type: github.Ptr("file"), Name: github.Ptr("support.md"), Path: github.Ptr(".github/support.md")},
			},
			expected: ".github/support.md",
		},
		{
			name:      "file not found",
			filename:  "nonexistent.md",
			toplevel:  []*github.RepositoryContent{},
			githubDir: []*github.RepositoryContent{},
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rest := &RestData{
				contents: RepoContent{
					Content: tt.toplevel,
					SubContent: map[string]RepoContent{
						".github": {Content: tt.githubDir},
					},
				},
			}
			result := rest.checkFile(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetSubdirContentByPath(t *testing.T) {
	subContent := RepoContent{
		Content: []*github.RepositoryContent{
			{Name: github.Ptr("workflow.yaml"), Type: github.Ptr("file"), Path: github.Ptr(".github/workflows/workflow.yaml")},
		},
	}

	root := RepoContent{
		SubContent: map[string]RepoContent{
			".github": {
				SubContent: map[string]RepoContent{
					"workflows": subContent,
				},
			},
		},
	}

	restData := &RestData{
		owner: "test-owner",
		repo:  "test-repo",
	}

	t.Run("successful path", func(t *testing.T) {
		result, err := root.GetSubdirContentByPath(restData, ".github/workflows")
		assert.NoError(t, err)
		assert.Equal(t, 1, len(result.Content))
		assert.Equal(t, "workflow.yaml", *result.Content[0].Name)
	})

	t.Run("nonexistent path", func(t *testing.T) {
		_, err := root.GetSubdirContentByPath(restData, ".github/nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "directory 'nonexistent' not found")
	})

	t.Run("no subdirectories", func(t *testing.T) {
		emptyRoot := RepoContent{}
		_, err := emptyRoot.GetSubdirContentByPath(restData, ".github")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no subdirectories found")
	})
}
func TestIsCodeRepo(t *testing.T) {
	tests := []struct {
		name           string
		responses      []mock.MockBackendOption
		expectedResult bool
		expectedError  bool
	}{
		{
			name: "repository with code languages",
			responses: []mock.MockBackendOption{
				mock.WithRequestMatch(
					mock.GetReposLanguagesByOwnerByRepo,
					map[string]int{"Go": 1000, "JavaScript": 500},
				),
			},
			expectedResult: true,
			expectedError:  false,
		},
		{
			name: "repository with no languages",
			responses: []mock.MockBackendOption{
				mock.WithRequestMatch(
					mock.GetReposLanguagesByOwnerByRepo,
					map[string]int{},
				),
			},
			expectedResult: false,
			expectedError:  false,
		},
		{
			name:           "api error",
			responses:      []mock.MockBackendOption{},
			expectedResult: true,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := mock.NewMockedHTTPClient(tt.responses...)
			ghClient := github.NewClient(mockClient)
			rest := &RestData{
				ghClient: ghClient,
				owner:    "test-owner",
				repo:     "test-repo",
			}
			result, err := rest.IsCodeRepo()

			assert.Equal(t, tt.expectedResult, result)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
