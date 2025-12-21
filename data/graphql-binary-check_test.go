package data

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/go-hclog"
)

func boolPtr(b bool) *bool {
	return &b
}

func TestCheckTreeForBinaries(t *testing.T) {
	bc := &binaryChecker{logger: hclog.NewNullLogger()}

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
			name: "text files are not flagged as binary",
			tree: buildTree([]testEntry{
				{name: "README.md", isBinary: boolPtr(false)},
				{name: "LICENSE", isBinary: boolPtr(false)},
				{name: "OWNERS", isBinary: boolPtr(false)},
				{name: "Tiltfile", isBinary: boolPtr(false)},
			}),
			expected: nil,
		},
		{
			name: "binary files are correctly detected",
			tree: buildTree([]testEntry{
				{name: "app.jar", isBinary: boolPtr(true)},
				{name: "README.md", isBinary: boolPtr(false)},
			}),
			expected: []string{"app.jar"},
		},
		{
			name: "multiple binary files detected",
			tree: buildTree([]testEntry{
				{name: "app.exe", isBinary: boolPtr(true)},
				{name: "lib.dll", isBinary: boolPtr(true)},
				{name: "main.go", isBinary: boolPtr(false)},
			}),
			expected: []string{"app.exe", "lib.dll"},
		},
		{
			name: "nested binary files detected",
			tree: buildTreeWithNested(
				[]testEntry{{name: "README.md", isBinary: boolPtr(false)}},
				[]testEntry{{name: "wrapper.jar", isBinary: boolPtr(true)}},
			),
			expected: []string{"wrapper.jar"},
		},
		{
			name: "extensionless text files not flagged",
			tree: buildTree([]testEntry{
				{name: "OWNERS", isBinary: boolPtr(false)},
				{name: "OWNERS_ALIASES", isBinary: boolPtr(false)},
				{name: "SECURITY_CONTACTS", isBinary: boolPtr(false)},
				{name: "Tiltfile", isBinary: boolPtr(false)},
				{name: "dockerignore", isBinary: boolPtr(false)},
				{name: "TECHNICAL_ADVISORY_MEMBERS", isBinary: boolPtr(false)},
			}),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkTreeForBinaries(tt.tree, bc)

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

func TestBinaryCheckerIsBinary(t *testing.T) {
	bc := &binaryChecker{logger: hclog.NewNullLogger()}

	t.Run("isBinary true returns true", func(t *testing.T) {
		result := bc.isBinary(boolPtr(true), false, "any-file")
		if !result {
			t.Error("expected isBinary=true to return true")
		}
	})

	t.Run("isBinary false returns false", func(t *testing.T) {
		result := bc.isBinary(boolPtr(false), false, "any-file")
		if result {
			t.Error("expected isBinary=false to return false")
		}
	})

	t.Run("isBinary false takes precedence over truncated", func(t *testing.T) {
		result := bc.isBinary(boolPtr(false), true, "any-file")
		if result {
			t.Error("expected isBinary=false to return false even when truncated")
		}
	})

	t.Run("nil isBinary and not truncated returns false", func(t *testing.T) {
		result := bc.isBinary(nil, false, "any-file")
		if result {
			t.Error("expected nil isBinary with isTruncated=false to return false")
		}
	})
}

func TestHasNullBytes(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		expected bool
	}{
		{"empty content", []byte{}, false},
		{"text only", []byte("hello world"), false},
		{"null at start", []byte{0x00, 'a', 'b'}, true},
		{"null at end", []byte{'a', 'b', 0x00}, true},
		{"null in middle", []byte{'a', 0x00, 'b'}, true},
		{"binary content", []byte{0xcf, 0xfa, 0xed, 0xfe, 0x00}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasNullBytes(tt.content)
			if result != tt.expected {
				t.Errorf("hasNullBytes(%v) = %v, want %v", tt.content, result, tt.expected)
			}
		})
	}
}

func TestCheckViaPartialFetch(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   []byte
		responseStatus int
		wantBinary     bool
		wantErr        bool
	}{
		{
			name:           "binary content detected",
			responseBody:   []byte{0xcf, 0xfa, 0xed, 0xfe, 0x00, 0x01, 0x02},
			responseStatus: http.StatusPartialContent,
			wantBinary:     true,
			wantErr:        false,
		},
		{
			name:           "text content not detected as binary",
			responseBody:   []byte("hello world"),
			responseStatus: http.StatusPartialContent,
			wantBinary:     false,
			wantErr:        false,
		},
		{
			name:           "OK status also works",
			responseBody:   []byte{0x00},
			responseStatus: http.StatusOK,
			wantBinary:     true,
			wantErr:        false,
		},
		{
			name:           "404 returns error",
			responseBody:   nil,
			responseStatus: http.StatusNotFound,
			wantBinary:     false,
			wantErr:        true,
		},
		{
			name:           "500 returns error",
			responseBody:   nil,
			responseStatus: http.StatusInternalServerError,
			wantBinary:     false,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseStatus)
				if tt.responseBody != nil {
					_, _ = w.Write(tt.responseBody)
				}
			}))
			defer server.Close()

			bc := &binaryChecker{
				httpClient: server.Client(),
				logger:     hclog.NewNullLogger(),
				owner:      "test",
				repo:       "repo",
				branch:     "main",
			}

			bc.httpClient.Transport = &testTransport{
				baseURL:   server.URL,
				transport: http.DefaultTransport,
			}

			got, err := bc.checkViaPartialFetch("test-file")
			if (err != nil) != tt.wantErr {
				t.Errorf("checkViaPartialFetch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantBinary {
				t.Errorf("checkViaPartialFetch() = %v, want %v", got, tt.wantBinary)
			}
		})
	}
}

func TestCheckViaPartialFetchURLEncoding(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		expectedPath string
	}{
		{
			name:         "spaces in filename",
			path:         "file with spaces.txt",
			expectedPath: "/test/repo/main/file%20with%20spaces.txt",
		},
		{
			name:         "multi-segment path preserved",
			path:         "dir/subdir/file.txt",
			expectedPath: "/test/repo/main/dir/subdir/file.txt",
		},
		{
			name:         "multi-segment with spaces",
			path:         "my dir/my file.txt",
			expectedPath: "/test/repo/main/my%20dir/my%20file.txt",
		},
		{
			name:         "special characters",
			path:         "file#1.txt",
			expectedPath: "/test/repo/main/file%231.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requestedPath string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestedPath = r.URL.EscapedPath()
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("text content"))
			}))
			defer server.Close()

			bc := &binaryChecker{
				httpClient: server.Client(),
				logger:     hclog.NewNullLogger(),
				owner:      "test",
				repo:       "repo",
				branch:     "main",
			}

			bc.httpClient.Transport = &testTransport{
				baseURL:   server.URL,
				transport: http.DefaultTransport,
			}

			_, _ = bc.checkViaPartialFetch(tt.path)

			if requestedPath != tt.expectedPath {
				t.Errorf("URL path = %q, want %q", requestedPath, tt.expectedPath)
			}
		})
	}
}

type testTransport struct {
	baseURL   string
	transport http.RoundTripper
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	serverURL, err := url.Parse(t.baseURL)
	if err != nil {
		return nil, err
	}
	req.URL.Scheme = serverURL.Scheme
	req.URL.Host = serverURL.Host
	return t.transport.RoundTrip(req)
}

type testEntry struct {
	name     string
	isBinary *bool
}

func buildTree(entries []testEntry) *GraphqlRepoTree {
	tree := &GraphqlRepoTree{}

	for _, e := range entries {
		entry := struct {
			Name   string
			Type   string
			Path   string
			Object *struct {
				Blob struct {
					IsBinary    *bool
					IsTruncated bool
				} `graphql:"... on Blob"`
				Tree struct {
					Entries []struct {
						Name   string
						Type   string
						Path   string
						Object *struct {
							Blob struct {
								IsBinary    *bool
								IsTruncated bool
							} `graphql:"... on Blob"`
							Tree struct {
								Entries []struct {
									Name   string
									Type   string
									Path   string
									Object *struct {
										Blob struct {
											IsBinary    *bool
											IsTruncated bool
										} `graphql:"... on Blob"`
									} `graphql:"object"`
								}
							} `graphql:"... on Tree"`
						} `graphql:"object"`
					}
				} `graphql:"... on Tree"`
			} `graphql:"object"`
		}{
			Name: e.name,
			Type: "blob",
			Path: e.name,
		}
		entry.Object = &struct {
			Blob struct {
				IsBinary    *bool
				IsTruncated bool
			} `graphql:"... on Blob"`
			Tree struct {
				Entries []struct {
					Name   string
					Type   string
					Path   string
					Object *struct {
						Blob struct {
							IsBinary    *bool
							IsTruncated bool
						} `graphql:"... on Blob"`
						Tree struct {
							Entries []struct {
								Name   string
								Type   string
								Path   string
								Object *struct {
									Blob struct {
										IsBinary    *bool
										IsTruncated bool
									} `graphql:"... on Blob"`
								} `graphql:"object"`
							}
						} `graphql:"... on Tree"`
					} `graphql:"object"`
				}
			} `graphql:"... on Tree"`
		}{}
		entry.Object.Blob.IsBinary = e.isBinary

		tree.Repository.Object.Tree.Entries = append(tree.Repository.Object.Tree.Entries, entry)
	}

	return tree
}

func buildTreeWithNested(rootEntries []testEntry, subEntries []testEntry) *GraphqlRepoTree {
	tree := buildTree(rootEntries)

	dirEntry := struct {
		Name   string
		Type   string
		Path   string
		Object *struct {
			Blob struct {
				IsBinary    *bool
				IsTruncated bool
			} `graphql:"... on Blob"`
			Tree struct {
				Entries []struct {
					Name   string
					Type   string
					Path   string
					Object *struct {
						Blob struct {
							IsBinary    *bool
							IsTruncated bool
						} `graphql:"... on Blob"`
						Tree struct {
							Entries []struct {
								Name   string
								Type   string
								Path   string
								Object *struct {
									Blob struct {
										IsBinary    *bool
										IsTruncated bool
									} `graphql:"... on Blob"`
								} `graphql:"object"`
							}
						} `graphql:"... on Tree"`
					} `graphql:"object"`
				}
			} `graphql:"... on Tree"`
		} `graphql:"object"`
	}{
		Name: "subdir",
		Type: "tree",
		Path: "subdir",
	}

	dirEntry.Object = &struct {
		Blob struct {
			IsBinary    *bool
			IsTruncated bool
		} `graphql:"... on Blob"`
		Tree struct {
			Entries []struct {
				Name   string
				Type   string
				Path   string
				Object *struct {
					Blob struct {
						IsBinary    *bool
						IsTruncated bool
					} `graphql:"... on Blob"`
					Tree struct {
						Entries []struct {
							Name   string
							Type   string
							Path   string
							Object *struct {
								Blob struct {
									IsBinary    *bool
									IsTruncated bool
								} `graphql:"... on Blob"`
							} `graphql:"object"`
						}
					} `graphql:"... on Tree"`
				} `graphql:"object"`
			}
		} `graphql:"... on Tree"`
	}{}

	for _, e := range subEntries {
		subEntry := struct {
			Name   string
			Type   string
			Path   string
			Object *struct {
				Blob struct {
					IsBinary    *bool
					IsTruncated bool
				} `graphql:"... on Blob"`
				Tree struct {
					Entries []struct {
						Name   string
						Type   string
						Path   string
						Object *struct {
							Blob struct {
								IsBinary    *bool
								IsTruncated bool
							} `graphql:"... on Blob"`
						} `graphql:"object"`
					}
				} `graphql:"... on Tree"`
			} `graphql:"object"`
		}{
			Name: e.name,
			Type: "blob",
			Path: "subdir/" + e.name,
		}
		subEntry.Object = &struct {
			Blob struct {
				IsBinary    *bool
				IsTruncated bool
			} `graphql:"... on Blob"`
			Tree struct {
				Entries []struct {
					Name   string
					Type   string
					Path   string
					Object *struct {
						Blob struct {
							IsBinary    *bool
							IsTruncated bool
						} `graphql:"... on Blob"`
					} `graphql:"object"`
				}
			} `graphql:"... on Tree"`
		}{}
		subEntry.Object.Blob.IsBinary = e.isBinary

		dirEntry.Object.Tree.Entries = append(dirEntry.Object.Tree.Entries, subEntry)
	}

	tree.Repository.Object.Tree.Entries = append(tree.Repository.Object.Tree.Entries, dirEntry)

	return tree
}
