package legal

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ossf/gemara/layer4"
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

			licenses, errString := testGetLicenseListLogic(mockMakeApiCall)

			assert.Equal(t, test.expectedError, errString)
			if test.expectEmpty {
				assert.Empty(t, licenses.Licenses)
			} else {
				assert.NotEmpty(t, licenses.Licenses)
			}
		})
	}
}

func testGetLicenseListLogic(makeApiCall func(string, bool) ([]byte, error)) (LicenseList, string) {
	goodLicenseList := LicenseList{}
	response, err := makeApiCall(spdxURL, false)
	if err != nil {
		return goodLicenseList, fmt.Sprintf("Failed to fetch good license data: %s", err.Error())
	}
	err = json.Unmarshal(response, &goodLicenseList)
	if err != nil {
		return goodLicenseList, fmt.Sprintf("Failed to unmarshal good license data: %s", err.Error())
	}
	if len(goodLicenseList.Licenses) == 0 {
		return goodLicenseList, "Good license data was unexpectedly empty"
	}
	return goodLicenseList, ""
}
