package legal

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gemaraproj/go-gemara"
	"github.com/ossf/pvtr-github-repo-scanner/data"
	"github.com/privateerproj/privateer-sdk/config"
	"github.com/stretchr/testify/assert"
)

type FakeGraphqlRepo struct {
	Repository struct {
		LicenseInfo struct {
			Url string
		}
	}
}

func stubGraphqlRepo(licenseUrl string) *data.GraphqlRepoData {
	repo := &data.GraphqlRepoData{}
	repo.Repository.LicenseInfo.Url = licenseUrl
	return repo
}

type treeEntry struct {
	name string
	typ  string
}

// stubGraphqlRepoWithTree builds repo data with a license URL and the given root
// tree entries, mirroring the anonymous entry struct in the GraphQL query.
func stubGraphqlRepoWithTree(licenseUrl string, entries ...treeEntry) *data.GraphqlRepoData {
	repo := stubGraphqlRepo(licenseUrl)
	for _, e := range entries {
		typ := e.typ
		if typ == "" {
			typ = "blob"
		}
		repo.Repository.Object.Tree.Entries = append(
			repo.Repository.Object.Tree.Entries,
			struct {
				Name string
				Type string
				Path string
			}{Name: e.name, Type: typ},
		)
	}
	return repo
}

func TestFoundLicense(t *testing.T) {
	tests := []struct {
		name            string
		payload         data.Payload
		expectedResult  gemara.Result
		expectedMessage string
	}{
		{
			name:            "GitHub identifies the license",
			payload:         data.Payload{GraphqlRepoData: stubGraphqlRepo("https://api.github.com/licenses/mit")},
			expectedResult:  gemara.Passed,
			expectedMessage: "License was found in a well known location via the GitHub API",
		},
		{
			name:            "unclassified LICENSE file in root",
			payload:         data.Payload{GraphqlRepoData: stubGraphqlRepoWithTree("", treeEntry{name: "LICENSE"})},
			expectedResult:  gemara.Passed,
			expectedMessage: `License file "LICENSE" found in the repository root, a well known location; GitHub could not identify the license type`,
		},
		{
			name:            "per-license LICENSE-MIT file in root",
			payload:         data.Payload{GraphqlRepoData: stubGraphqlRepoWithTree("", treeEntry{name: "LICENSE-MIT"})},
			expectedResult:  gemara.Passed,
			expectedMessage: `License file "LICENSE-MIT" found in the repository root, a well known location; GitHub could not identify the license type`,
		},
		{
			name:            "COPYING file in root, case-insensitive",
			payload:         data.Payload{GraphqlRepoData: stubGraphqlRepoWithTree("", treeEntry{name: "copying"})},
			expectedResult:  gemara.Passed,
			expectedMessage: `License file "copying" found in the repository root, a well known location; GitHub could not identify the license type`,
		},
		{
			name:            "a directory named LICENSE is not a license file",
			payload:         data.Payload{GraphqlRepoData: stubGraphqlRepoWithTree("", treeEntry{name: "LICENSE", typ: "tree"})},
			expectedResult:  gemara.Failed,
			expectedMessage: "License was not found in a well known location via the GitHub API",
		},
		{
			name:            "no license anywhere",
			payload:         data.Payload{GraphqlRepoData: stubGraphqlRepoWithTree("", treeEntry{name: "README.md"})},
			expectedResult:  gemara.Failed,
			expectedMessage: "License was not found in a well known location via the GitHub API",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, message, _ := FoundLicense(test.payload)
			assert.Equal(t, test.expectedResult, result)
			assert.Equal(t, test.expectedMessage, message)
		})
	}
}

// withLicenseEndpoint points data.APIBase at a stub license endpoint for the
// duration of the test. statusCode 404 simulates "no license at the ref";
// 200 responds with the given SPDX id and path; anything else is a hard error.
func withLicenseEndpoint(t *testing.T, statusCode int, spdxId, path string) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if statusCode != http.StatusOK {
			w.WriteHeader(statusCode)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"path": %q, "license": {"spdx_id": %q}}`, path, spdxId)
	}))
	t.Cleanup(server.Close)
	oldAPIBase := data.APIBase
	data.APIBase = server.URL
	t.Cleanup(func() { data.APIBase = oldAPIBase })
}

// restDataWithReleases builds a RestData carrying the given releases and a
// working HTTP client for tag-level license lookups.
func restDataWithReleases(releases ...data.ReleaseData) *data.RestData {
	return &data.RestData{
		Releases:   releases,
		HttpClient: http.DefaultClient,
	}
}

func TestReleasesLicensed(t *testing.T) {
	t.Run("no RestData", func(t *testing.T) {
		result, message, confidence := ReleasesLicensed(data.Payload{})
		assert.Equal(t, gemara.NeedsReview, result)
		assert.Equal(t, "Release data is unavailable; review the released assets for license coverage", message)
		assert.Equal(t, gemara.Low, confidence)
	})

	t.Run("release fetch failed", func(t *testing.T) {
		payload := data.Payload{
			RestData: &data.RestData{ReleasesError: fmt.Errorf("boom")},
		}
		result, message, confidence := ReleasesLicensed(payload)
		assert.Equal(t, gemara.NeedsReview, result)
		assert.Equal(t, "Release data could not be retrieved: boom. Review the released assets for license coverage", message)
		assert.Equal(t, gemara.Low, confidence)
	})

	t.Run("no releases found", func(t *testing.T) {
		payload := data.Payload{RestData: &data.RestData{}}
		result, message, _ := ReleasesLicensed(payload)
		assert.Equal(t, gemara.NotApplicable, result)
		assert.Equal(t, "No releases found", message)
	})

	t.Run("draft releases do not count as published", func(t *testing.T) {
		payload := data.Payload{
			RestData: restDataWithReleases(data.ReleaseData{TagName: "v1.0.0", Draft: true}),
		}
		result, message, _ := ReleasesLicensed(payload)
		assert.Equal(t, gemara.NotApplicable, result)
		assert.Equal(t, "No releases found", message)
	})

	t.Run("license asset attached to the release needs review", func(t *testing.T) {
		payload := data.Payload{
			RestData: restDataWithReleases(data.ReleaseData{
				TagName: "v1.0.0",
				Assets: []data.ReleaseAsset{
					{Name: "scanner-linux-amd64.tar.gz"},
					{Name: "LICENSE.txt"},
					{Name: "LICENSE-MIT"},
				},
			}),
			GraphqlRepoData: stubGraphqlRepo("https://api.github.com/licenses/mit"),
		}
		result, message, confidence := ReleasesLicensed(payload)
		assert.Equal(t, gemara.NeedsReview, result)
		assert.Equal(t, `Release "v1.0.0" attaches standalone license file(s) as release assets (LICENSE.txt, LICENSE-MIT), which may supersede the license in the released source code; manual review is required to confirm they match an approved license`, message)
		assert.Equal(t, gemara.High, confidence)
	})

	t.Run("only the latest published release's assets are considered", func(t *testing.T) {
		withLicenseEndpoint(t, http.StatusOK, "MIT", "LICENSE")
		payload := data.Payload{
			RestData: restDataWithReleases(
				data.ReleaseData{TagName: "v2.0.0-rc1", Draft: true, Assets: []data.ReleaseAsset{{Name: "LICENSE"}}},
				data.ReleaseData{TagName: "v1.1.0"},
				data.ReleaseData{TagName: "v1.0.0", Assets: []data.ReleaseAsset{{Name: "LICENSE"}}},
			),
		}
		result, message, confidence := ReleasesLicensed(payload)
		assert.Equal(t, gemara.Passed, result)
		assert.Equal(t, `GitHub identifies license MIT at "LICENSE" in the released source code at tag "v1.1.0"; the auto-generated release archives include it`, message)
		assert.Equal(t, gemara.High, confidence)
	})

	t.Run("license identified at the release tag", func(t *testing.T) {
		withLicenseEndpoint(t, http.StatusOK, "Apache-2.0", "LICENSE")
		payload := data.Payload{
			RestData: restDataWithReleases(data.ReleaseData{TagName: "v1.0.0"}),
		}
		result, message, confidence := ReleasesLicensed(payload)
		assert.Equal(t, gemara.Passed, result)
		assert.Equal(t, `GitHub identifies license Apache-2.0 at "LICENSE" in the released source code at tag "v1.0.0"; the auto-generated release archives include it`, message)
		assert.Equal(t, gemara.High, confidence)
	})

	t.Run("unidentified license at the release tag", func(t *testing.T) {
		withLicenseEndpoint(t, http.StatusOK, "NOASSERTION", "COPYING")
		payload := data.Payload{
			RestData: restDataWithReleases(data.ReleaseData{TagName: "v1.0.0"}),
		}
		result, message, confidence := ReleasesLicensed(payload)
		assert.Equal(t, gemara.Passed, result)
		assert.Equal(t, `License file "COPYING" is present in the released source code at tag "v1.0.0", but GitHub could not identify the license type`, message)
		assert.Equal(t, gemara.Medium, confidence)
	})

	t.Run("license on default branch but missing at the release tag", func(t *testing.T) {
		withLicenseEndpoint(t, http.StatusNotFound, "", "")
		payload := data.Payload{
			RestData:        restDataWithReleases(data.ReleaseData{TagName: "v1.0.0"}),
			GraphqlRepoData: stubGraphqlRepo("https://api.github.com/licenses/mit"),
		}
		result, message, confidence := ReleasesLicensed(payload)
		assert.Equal(t, gemara.Failed, result)
		assert.Equal(t, `A license exists on the default branch, but none was found in the released source code at tag "v1.0.0"`, message)
		assert.Equal(t, gemara.High, confidence)
	})

	t.Run("no license anywhere at the release tag", func(t *testing.T) {
		withLicenseEndpoint(t, http.StatusNotFound, "", "")
		payload := data.Payload{
			RestData:        restDataWithReleases(data.ReleaseData{TagName: "v1.0.0"}),
			GraphqlRepoData: &data.GraphqlRepoData{},
		}
		result, message, confidence := ReleasesLicensed(payload)
		assert.Equal(t, gemara.Failed, result)
		assert.Equal(t, `No license was found in the released source code at tag "v1.0.0"`, message)
		assert.Equal(t, gemara.High, confidence)
	})

	t.Run("tag lookup failure falls back to default-branch license", func(t *testing.T) {
		withLicenseEndpoint(t, http.StatusInternalServerError, "", "")
		payload := data.Payload{
			RestData:        restDataWithReleases(data.ReleaseData{TagName: "v1.0.0"}),
			GraphqlRepoData: stubGraphqlRepo("https://api.github.com/licenses/mit"),
		}
		result, message, confidence := ReleasesLicensed(payload)
		assert.Equal(t, gemara.Passed, result)
		assert.Equal(t, "A license was found on the default branch; the released source code could not be checked directly, so this is default-branch evidence only", message)
		assert.Equal(t, gemara.Medium, confidence)
	})

	t.Run("tag lookup failure falls back to unclassified root license file", func(t *testing.T) {
		withLicenseEndpoint(t, http.StatusInternalServerError, "", "")
		payload := data.Payload{
			RestData:        restDataWithReleases(data.ReleaseData{TagName: "v1.0.0"}),
			GraphqlRepoData: stubGraphqlRepoWithTree("", treeEntry{name: "LICENSE"}),
		}
		result, message, confidence := ReleasesLicensed(payload)
		assert.Equal(t, gemara.Passed, result)
		assert.Equal(t, `License file "LICENSE" found in the repository root; the released source code could not be checked directly, so this is default-branch evidence only`, message)
		assert.Equal(t, gemara.Low, confidence)
	})

	t.Run("tag lookup failure with no license evidence fails", func(t *testing.T) {
		withLicenseEndpoint(t, http.StatusInternalServerError, "", "")
		payload := data.Payload{
			RestData:        restDataWithReleases(data.ReleaseData{TagName: "v1.0.0"}),
			GraphqlRepoData: &data.GraphqlRepoData{},
		}
		result, message, confidence := ReleasesLicensed(payload)
		assert.Equal(t, gemara.Failed, result)
		assert.Equal(t, "License was not found in a well known location via the GitHub API", message)
		assert.Equal(t, gemara.Medium, confidence)
	})

	t.Run("release without a tag name falls back to default-branch license", func(t *testing.T) {
		payload := data.Payload{
			RestData:        restDataWithReleases(data.ReleaseData{Name: "v1.0.0"}),
			GraphqlRepoData: stubGraphqlRepo("https://api.github.com/licenses/mit"),
		}
		result, message, confidence := ReleasesLicensed(payload)
		assert.Equal(t, gemara.Passed, result)
		assert.Equal(t, "A license was found on the default branch; the released source code could not be checked directly, so this is default-branch evidence only", message)
		assert.Equal(t, gemara.Medium, confidence)
	})
}

func TestGetLicenseList(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockError     error
		expectedError string
		expectEmpty   bool
	}{
		{
			name:          "Successful Fetch and Parse",
			mockResponse:  `{"licenses": [{"licenseId": "MIT", "isOsiApproved": true, "isFsfLibre": true}]}`,
			mockError:     nil,
			expectedError: "",
			expectEmpty:   false,
		},
		{
			name:          "Fetch Error",
			mockResponse:  "",
			mockError:     fmt.Errorf("fetch error"),
			expectedError: "Failed to fetch good license data: fetch error",
			expectEmpty:   true,
		},
		{
			name:          "Parse Error",
			mockResponse:  "invalid json",
			mockError:     nil,
			expectedError: "Failed to unmarshal good license data: invalid character 'i' looking for beginning of value",
			expectEmpty:   true,
		},
		{
			name:          "Empty License List",
			mockResponse:  `{"licenses": []}`,
			mockError:     nil,
			expectedError: "Good license data was unexpectedly empty",
			expectEmpty:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockMakeApiCall := func(endpoint string, isGithub bool) ([]byte, error) {
				if test.mockError != nil {
					return nil, test.mockError
				}
				return []byte(test.mockResponse), nil
			}

			payload := data.Payload{}
			licenses, errString := getLicenseList(payload, mockMakeApiCall)

			assert.Equal(t, test.expectedError, errString)
			if test.expectEmpty {
				assert.Empty(t, licenses.Licenses)
			} else {
				assert.NotEmpty(t, licenses.Licenses)
			}
		})
	}
}

func TestSplitSpdxExpression(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Single license",
			input:    "MIT",
			expected: []string{"MIT"},
		},
		{
			name:     "Simple AND",
			input:    "MIT AND Apache-2.0",
			expected: []string{"MIT", "Apache-2.0"},
		},
		{
			name:     "Simple OR",
			input:    "MIT OR GPL-3.0",
			expected: []string{"MIT", "GPL-3.0"},
		},
		{
			name:     "Multiple AND",
			input:    "MIT AND Apache-2.0 AND BSD-3-Clause",
			expected: []string{"MIT", "Apache-2.0", "BSD-3-Clause"},
		},
		{
			name:     "Mixed AND and OR",
			input:    "MIT AND Apache-2.0 OR GPL-3.0",
			expected: []string{"MIT", "Apache-2.0", "GPL-3.0"},
		},
		{
			name:     "Empty string",
			input:    "",
			expected: []string{""},
		},
		{
			name:     "WITH exception suffix",
			input:    "Apache-2.0 WITH LLVM-exception",
			expected: []string{"Apache-2.0"},
		},
		{
			name:     "WITH exception inside grouped OR",
			input:    "(GPL-2.0 WITH Classpath-exception-2.0) OR MIT",
			expected: []string{"GPL-2.0", "MIT"},
		},
		{
			name:     "Deprecated + suffix",
			input:    "GPL-2.0+",
			expected: []string{"GPL-2.0"},
		},
		{
			name:     "Parentheses grouping",
			input:    "(MIT OR Apache-2.0)",
			expected: []string{"MIT", "Apache-2.0"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := splitSpdxExpression(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestGoodLicense(t *testing.T) {
	tests := []struct {
		name            string
		payload         data.Payload
		apiResponse     []byte
		apiError        error
		expectedResult  gemara.Result
		expectedMessage string
	}{
		{
			name: "No license identifiers found",
			payload: data.Payload{
				GraphqlRepoData: &data.GraphqlRepoData{},
				Config:          &config.Config{},
			},
			apiResponse:     []byte(`{"licenses":[{"licenseId":"MIT","isOsiApproved":true,"isFsfLibre":false}]}`),
			apiError:        nil,
			expectedResult:  gemara.Failed,
			expectedMessage: "License SPDX identifier was not found in Security Insights data or via GitHub API",
		},
		{
			name: "OSI approved license (MIT)",
			payload: data.Payload{
				GraphqlRepoData: func() *data.GraphqlRepoData {
					repo := stubGraphqlRepo("")
					repo.Repository.LicenseInfo.SpdxId = "MIT"
					return repo
				}(),
				Config: &config.Config{},
			},
			apiResponse:     []byte(`{"licenses":[{"licenseId":"MIT","isOsiApproved":true,"isFsfLibre":false}]}`),
			apiError:        nil,
			expectedResult:  gemara.Passed,
			expectedMessage: "All license found are OSI or FSF approved",
		},
		{
			name: "Non-approved license",
			payload: data.Payload{
				GraphqlRepoData: func() *data.GraphqlRepoData {
					repo := stubGraphqlRepo("")
					repo.Repository.LicenseInfo.SpdxId = "BadLicense"
					return repo
				}(),
				Config: &config.Config{},
			},
			apiResponse:     []byte(`{"licenses":[{"licenseId":"BadLicense","isOsiApproved":false,"isFsfLibre":false}]}`),
			apiError:        nil,
			expectedResult:  gemara.Failed,
			expectedMessage: "These licenses are not OSI or FSF approved: BadLicense",
		},
		{
			name: "Multiple licenses with mixed approval",
			payload: data.Payload{
				GraphqlRepoData: func() *data.GraphqlRepoData {
					repo := stubGraphqlRepo("")
					repo.Repository.LicenseInfo.SpdxId = "MIT AND BadLicense"
					return repo
				}(),
				Config: &config.Config{},
			},
			apiResponse:     []byte(`{"licenses":[{"licenseId":"MIT","isOsiApproved":true,"isFsfLibre":false},{"licenseId":"BadLicense","isOsiApproved":false,"isFsfLibre":false}]}`),
			apiError:        nil,
			expectedResult:  gemara.Failed,
			expectedMessage: "These licenses are not OSI or FSF approved: BadLicense",
		},
		{
			name: "Unknown license ID",
			payload: data.Payload{
				GraphqlRepoData: func() *data.GraphqlRepoData {
					repo := stubGraphqlRepo("")
					repo.Repository.LicenseInfo.SpdxId = "UnknownLicense"
					return repo
				}(),
				Config: &config.Config{},
			},
			apiResponse:     []byte(`{"licenses":[{"licenseId":"MIT","isOsiApproved":true,"isFsfLibre":false}]}`),
			apiError:        nil,
			expectedResult:  gemara.Failed,
			expectedMessage: "These licenses are not OSI or FSF approved: UnknownLicense",
		},
		{
			name: "Deprecated but OSI and FSF approved license (AGPL-3.0)",
			payload: data.Payload{
				GraphqlRepoData: func() *data.GraphqlRepoData {
					repo := stubGraphqlRepo("")
					repo.Repository.LicenseInfo.SpdxId = "AGPL-3.0"
					return repo
				}(),
				Config: &config.Config{},
			},
			apiResponse:     []byte(`{"licenses":[{"licenseId":"AGPL-3.0","isOsiApproved":true,"isFsfLibre":true,"isDeprecatedLicenseId":true}]}`),
			apiError:        nil,
			expectedResult:  gemara.Passed,
			expectedMessage: "All licenses found are OSI or FSF approved. Note: the following SPDX IDs are deprecated and should be migrated to their -only/-or-later form: AGPL-3.0",
		},
		{
			name: "Mix of deprecated-approved and non-deprecated approved licenses",
			payload: data.Payload{
				GraphqlRepoData: func() *data.GraphqlRepoData {
					repo := stubGraphqlRepo("")
					repo.Repository.LicenseInfo.SpdxId = "MIT AND AGPL-3.0"
					return repo
				}(),
				Config: &config.Config{},
			},
			apiResponse:     []byte(`{"licenses":[{"licenseId":"MIT","isOsiApproved":true,"isFsfLibre":false},{"licenseId":"AGPL-3.0","isOsiApproved":true,"isFsfLibre":true,"isDeprecatedLicenseId":true}]}`),
			apiError:        nil,
			expectedResult:  gemara.Passed,
			expectedMessage: "All licenses found are OSI or FSF approved. Note: the following SPDX IDs are deprecated and should be migrated to their -only/-or-later form: AGPL-3.0",
		},
		{
			name: "Mix of deprecated-approved and non-approved licenses",
			payload: data.Payload{
				GraphqlRepoData: func() *data.GraphqlRepoData {
					repo := stubGraphqlRepo("")
					repo.Repository.LicenseInfo.SpdxId = "AGPL-3.0 AND BadLicense"
					return repo
				}(),
				Config: &config.Config{},
			},
			apiResponse:     []byte(`{"licenses":[{"licenseId":"AGPL-3.0","isOsiApproved":true,"isFsfLibre":true,"isDeprecatedLicenseId":true},{"licenseId":"BadLicense","isOsiApproved":false,"isFsfLibre":false}]}`),
			apiError:        nil,
			expectedResult:  gemara.Failed,
			expectedMessage: "These licenses are not OSI or FSF approved: BadLicense",
		},
		{
			name: "NOASSERTION with a license file present needs review",
			payload: data.Payload{
				GraphqlRepoData: func() *data.GraphqlRepoData {
					repo := stubGraphqlRepoWithTree("", treeEntry{name: "LICENSE"})
					repo.Repository.LicenseInfo.SpdxId = "NOASSERTION"
					return repo
				}(),
				Config: &config.Config{},
			},
			apiResponse:     []byte(`{"licenses":[{"licenseId":"MIT","isOsiApproved":true,"isFsfLibre":false}]}`),
			apiError:        nil,
			expectedResult:  gemara.NeedsReview,
			expectedMessage: `License file "LICENSE" is present but its SPDX identity could not be determined; manual review is required to confirm OSI or FSF approval`,
		},
		{
			name: "NOASSERTION with no license file fails",
			payload: data.Payload{
				GraphqlRepoData: func() *data.GraphqlRepoData {
					repo := stubGraphqlRepo("")
					repo.Repository.LicenseInfo.SpdxId = "NOASSERTION"
					return repo
				}(),
				Config: &config.Config{},
			},
			apiResponse:     []byte(`{"licenses":[{"licenseId":"MIT","isOsiApproved":true,"isFsfLibre":false}]}`),
			apiError:        nil,
			expectedResult:  gemara.Failed,
			expectedMessage: "License SPDX identifier was not found in Security Insights data or via GitHub API",
		},
		{
			name: "No SPDX id but a license file present needs review",
			payload: data.Payload{
				GraphqlRepoData: stubGraphqlRepoWithTree("", treeEntry{name: "COPYING"}),
				Config:          &config.Config{},
			},
			apiResponse:     []byte(`{"licenses":[{"licenseId":"MIT","isOsiApproved":true,"isFsfLibre":false}]}`),
			apiError:        nil,
			expectedResult:  gemara.NeedsReview,
			expectedMessage: `License file "COPYING" is present but its SPDX identity could not be determined; manual review is required to confirm OSI or FSF approval`,
		},
		{
			name: "Deprecated and non-approved license fails",
			payload: data.Payload{
				GraphqlRepoData: func() *data.GraphqlRepoData {
					repo := stubGraphqlRepo("")
					repo.Repository.LicenseInfo.SpdxId = "DeprecatedBadLicense"
					return repo
				}(),
				Config: &config.Config{},
			},
			apiResponse:     []byte(`{"licenses":[{"licenseId":"DeprecatedBadLicense","isOsiApproved":false,"isFsfLibre":false,"isDeprecatedLicenseId":true}]}`),
			apiError:        nil,
			expectedResult:  gemara.Failed,
			expectedMessage: "These licenses are not OSI or FSF approved: DeprecatedBadLicense",
		},
		{
			name: "OSI approved license with WITH exception clause (Kokkos pattern)",
			payload: data.Payload{
				GraphqlRepoData: func() *data.GraphqlRepoData {
					repo := stubGraphqlRepo("")
					repo.Repository.LicenseInfo.SpdxId = "Apache-2.0 WITH LLVM-exception"
					return repo
				}(),
				Config: &config.Config{},
			},
			apiResponse:     []byte(`{"licenses":[{"licenseId":"Apache-2.0","isOsiApproved":true,"isFsfLibre":false}]}`),
			apiError:        nil,
			expectedResult:  gemara.Passed,
			expectedMessage: "All license found are OSI or FSF approved",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := data.NewPayloadWithHTTPMock(test.payload, test.apiResponse, 200, test.apiError)
			test.payload = data

			result, message, _ := GoodLicense(test.payload)
			assert.Equal(t, test.expectedResult, result)
			assert.Equal(t, test.expectedMessage, message)
		})
	}
}
