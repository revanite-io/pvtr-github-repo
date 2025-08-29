package legal

import (
	"testing"

	"github.com/ossf/gemara/layer4"
	"github.com/revanite-io/pvtr-github-repo/data"
	"github.com/stretchr/testify/assert"
)

// FakeGraphqlRepo provides a minimal implementation of GraphqlRepoData
// for testing OSPS-LE-03.02 license verification. It only includes
// the LicenseInfo structure needed for license URL verification.
type FakeGraphqlRepo struct {
	Repository struct {
		LicenseInfo struct {
			Url string
		}
	}
}

// stubGraphqlRepo creates a GraphqlRepoData instance with a specified license URL.
// This helper function is used to test the successful case of license verification
// where both releases and a valid license exist.
func stubGraphqlRepo(licenseUrl string) *data.GraphqlRepoData {
	repo := &data.GraphqlRepoData{}
	repo.Repository.LicenseInfo.Url = licenseUrl // Set the license URL
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
		//Added test case for when the license and releases are found
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
