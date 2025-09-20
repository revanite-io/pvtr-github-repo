package legal

import (
	"fmt"
	"testing"

	"github.com/ossf/gemara/layer4"
	"github.com/privateerproj/privateer-sdk/config"
	"github.com/revanite-io/pvtr-github-repo/data"
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

func TestReleasesLicensed(t *testing.T) {
	tests := []struct {
		name            string
		payloadData     any
		expectedResult  layer4.Result
		expectedMessage string
	}{
		{
			name: "No releases found",
			payloadData: data.Payload{
				RestData: &data.RestData{
					Releases: []data.ReleaseData{},
				},
			},
			expectedResult:  layer4.NotApplicable,
			expectedMessage: "No releases found",
		},
		{
			name: "No licenses found",
			payloadData: data.Payload{
				RestData: &data.RestData{
					Releases: []data.ReleaseData{
						{
							Name: "v1.0.0",
						},
					},
				},
				GraphqlRepoData: &data.GraphqlRepoData{},
			},
			expectedResult:  layer4.Failed,
			expectedMessage: "License was not found in a well known location via the GitHub API",
		},
		{
			name: "Has releases and license",
			payloadData: data.Payload{
				RestData: &data.RestData{
					Releases: []data.ReleaseData{
						{
							Name: "v1.0.0",
						},
					},
				},
				GraphqlRepoData: stubGraphqlRepo("https://api.github.com/licenses/mit"),
			},
			expectedResult:  layer4.Passed,
			expectedMessage: "GitHub releases include the license(s) in the released source code.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, message := releasesLicensed(test.payloadData, nil)
			assert.Equal(t, test.expectedResult, result)
			assert.Equal(t, test.expectedMessage, message)
		})
	}
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
		payloadData     any
		mockLicenses    LicenseList
		mockError       string
		expectedResult  layer4.Result
		expectedMessage string
	}{
		{
			name:            "Invalid payload",
			payloadData:     "invalid",
			expectedResult:  layer4.Unknown,
			expectedMessage: "Malformed assessment: expected payload type data.Payload, got string (invalid)",
		},
		{
			name: "No license identifiers found",
			payloadData: data.Payload{
				RestData:        &data.RestData{},
				GraphqlRepoData: &data.GraphqlRepoData{},
				Config:          &config.Config{},
			},
			mockLicenses: LicenseList{
				Licenses: []License{
					{LicenseID: "MIT", IsOsiApproved: true, IsFsfLibre: false},
				},
			},
			expectedResult:  layer4.Failed,
			expectedMessage: "License SPDX identifier was not found in Security Insights data or via GitHub API",
		},
		{
			name: "OSI approved license (MIT)",
			payloadData: data.Payload{
				RestData: &data.RestData{},
				GraphqlRepoData: func() *data.GraphqlRepoData {
					repo := stubGraphqlRepo("")
					repo.Repository.LicenseInfo.SpdxId = "MIT"
					return repo
				}(),
				Config: &config.Config{},
			},
			mockLicenses: LicenseList{
				Licenses: []License{
					{LicenseID: "MIT", IsOsiApproved: true, IsFsfLibre: false},
				},
			},
			expectedResult:  layer4.NeedsReview,
			expectedMessage: "All license found are OSI or FSF approved",
		},
		{
			name: "Non-approved license",
			payloadData: data.Payload{
				RestData: &data.RestData{},
				GraphqlRepoData: func() *data.GraphqlRepoData {
					repo := stubGraphqlRepo("")
					repo.Repository.LicenseInfo.SpdxId = "BadLicense"
					return repo
				}(),
				Config: &config.Config{},
			},
			mockLicenses: LicenseList{
				Licenses: []License{
					{LicenseID: "BadLicense", IsOsiApproved: false, IsFsfLibre: false},
				},
			},
			expectedResult:  layer4.Failed,
			expectedMessage: "These licenses are not OSI or FSF approved: BadLicense",
		},
		{
			name: "Multiple licenses with mixed approval",
			payloadData: data.Payload{
				RestData: &data.RestData{},
				GraphqlRepoData: func() *data.GraphqlRepoData {
					repo := stubGraphqlRepo("")
					repo.Repository.LicenseInfo.SpdxId = "MIT AND BadLicense"
					return repo
				}(),
				Config: &config.Config{},
			},
			mockLicenses: LicenseList{
				Licenses: []License{
					{LicenseID: "MIT", IsOsiApproved: true, IsFsfLibre: false},
					{LicenseID: "BadLicense", IsOsiApproved: false, IsFsfLibre: false},
				},
			},
			expectedResult:  layer4.Failed,
			expectedMessage: "These licenses are not OSI or FSF approved: BadLicense",
		},
		{
			name: "Unknown license ID",
			payloadData: data.Payload{
				RestData: &data.RestData{},
				GraphqlRepoData: func() *data.GraphqlRepoData {
					repo := stubGraphqlRepo("")
					repo.Repository.LicenseInfo.SpdxId = "UnknownLicense"
					return repo
				}(),
				Config: &config.Config{},
			},
			mockLicenses: LicenseList{
				Licenses: []License{
					{LicenseID: "MIT", IsOsiApproved: true, IsFsfLibre: false},
				},
			},
			expectedResult:  layer4.Failed,
			expectedMessage: "These licenses are not OSI or FSF approved: UnknownLicense",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Mock getLicenseList function
			mockGetLicenseList := func(data data.Payload, makeApiCall func(string, bool) ([]byte, error)) (LicenseList, string) {
				if test.mockError != "" {
					return LicenseList{}, test.mockError
				}
				return test.mockLicenses, ""
			}

			result, message := goodLicense(test.payloadData, nil, mockGetLicenseList)
			assert.Equal(t, test.expectedResult, result)
			assert.Equal(t, test.expectedMessage, message)
		})
	}
}
