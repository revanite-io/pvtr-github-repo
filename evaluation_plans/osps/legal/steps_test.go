package legal

import (
	"testing"

	"github.com/ossf/gemara/layer4"
	"github.com/revanite-io/pvtr-github-repo/data"
	"github.com/stretchr/testify/assert"
)

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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, message := releasesLicensed(test.payloadData, nil)
			assert.Equal(t, test.expectedResult, result)
			assert.Equal(t, test.expectedMessage, message)
		})
	}
}
