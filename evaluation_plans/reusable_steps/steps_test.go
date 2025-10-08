package reusable_steps

import (
	"testing"

	"github.com/ossf/gemara/layer4"
	"github.com/ossf/si-tooling/v2/si"
	"github.com/revanite-io/pvtr-github-repo/data"
	"github.com/stretchr/testify/assert"
)

type testingData struct {
	expectedResult   layer4.Result
	expectedMessage  string
	payloadData      any
	assertionMessage string
}

func TestHasDependencyManagementPolicy(t *testing.T) {

	//Ick, remind me to never use anonymous structs in my code
	testData := []testingData{
		{
			expectedResult:  layer4.Passed,
			expectedMessage: "Found dependency management policy in documentation",
			payloadData: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Repository: si.Repository{
							Documentation: struct {
								Contributing         string `yaml:"contributing-guide"`
								DependencyManagement string `yaml:"dependency-management-policy"`
								Governance           string `yaml:"governance"`
								ReviewPolicy         string `yaml:"review-policy"`
								SecurityPolicy       string `yaml:"security-policy"`
							}{
								DependencyManagement: "https://example.com/dependency-management",
							},
						},
					},
				},
			},
			assertionMessage: "Happy Path failed",
		},
		{
			expectedResult:  layer4.Failed,
			expectedMessage: "No dependency management file found",
			payloadData: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Repository: si.Repository{
							Documentation: struct {
								Contributing         string `yaml:"contributing-guide"`
								DependencyManagement string `yaml:"dependency-management-policy"`
								Governance           string `yaml:"governance"`
								ReviewPolicy         string `yaml:"review-policy"`
								SecurityPolicy       string `yaml:"security-policy"`
							}{
								DependencyManagement: "",
							},
						},
					},
				},
			},
			assertionMessage: "Empty string check failed",
		},
		{
			expectedResult:  layer4.Failed,
			expectedMessage: "No dependency management file found",
			payloadData: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Repository: si.Repository{
							Documentation: struct {
								Contributing         string `yaml:"contributing-guide"`
								DependencyManagement string `yaml:"dependency-management-policy"`
								Governance           string `yaml:"governance"`
								ReviewPolicy         string `yaml:"review-policy"`
								SecurityPolicy       string `yaml:"security-policy"`
							}{
								DependencyManagement: *new(string), // empty string pointer effectively nil value
							},
						},
					},
				},
			},
			assertionMessage: "Null String check failed",
		},
	}

	for _, test := range testData {
		result, message := HasDependencyManagementPolicy(test.payloadData)
		assert.Equal(t, test.expectedResult, result, test.assertionMessage)
		assert.Equal(t, test.expectedMessage, message, test.assertionMessage)
	}

}
func TestIsCodeRepo(t *testing.T) {
	tests := []struct {
		name             string
		payloadData      any
		expectedResult   layer4.Result
		expectedMessage  string
		assertionMessage string
	}{
		{
			name: "Repository contains code",
			payloadData: data.Payload{
				IsCodeRepo: true,
			},
			expectedResult:   layer4.Passed,
			expectedMessage:  "Repository contains code",
			assertionMessage: "Should pass when IsCodeRepo is true",
		},
		{
			name: "Repository does not contain code",
			payloadData: data.Payload{
				IsCodeRepo: false,
			},
			expectedResult:   layer4.NotApplicable,
			expectedMessage:  "Repository does not contain code",
			assertionMessage: "Should be not applicable when IsCodeRepo is false",
		},
		{
			name:             "Malformed payload type",
			payloadData:      "not a payload",
			expectedResult:   layer4.Unknown,
			expectedMessage:  "Malformed assessment: expected payload type data.Payload, got string (not a payload)",
			assertionMessage: "Should return Unknown for wrong payload type",
		},
	}

	for _, tt := range tests {
		result, message := IsCodeRepo(tt.payloadData)
		assert.Equal(t, tt.expectedResult, result, tt.assertionMessage)
		assert.Equal(t, tt.expectedMessage, message, tt.assertionMessage)
	}
}
func TestHasSecurityInsightsFile(t *testing.T) {
	tests := []struct {
		name             string
		payloadData      any
		expectedResult   layer4.Result
		expectedMessage  string
		assertionMessage string
	}{
		{
			name: "Security insights file found",
			payloadData: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Header: si.Header{
							URL: "https://example.com/security-insights",
						},
					},
				},
			},
			expectedResult:   layer4.Passed,
			expectedMessage:  "Security insights file found",
			assertionMessage: "Should pass when security insights file URL is present",
		},
		{
			name: "Security insights file not found",
			payloadData: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Header: si.Header{
							URL: "",
						},
					},
				},
			},
			expectedResult:   layer4.NeedsReview,
			expectedMessage:  "Security insights required for this assessment, but file not found",
			assertionMessage: "Should need review when security insights file URL is empty",
		},
		{
			name:             "Malformed payload type",
			payloadData:      "not a payload",
			expectedResult:   layer4.Unknown,
			expectedMessage:  "Malformed assessment: expected payload type data.Payload, got string (not a payload)",
			assertionMessage: "Should return Unknown for wrong payload type",
		},
	}

	for _, tt := range tests {
		result, message := HasSecurityInsightsFile(tt.payloadData)
		assert.Equal(t, tt.expectedResult, result, tt.assertionMessage)
		assert.Equal(t, tt.expectedMessage, message, tt.assertionMessage)
	}
}
func TestIsActive(t *testing.T) {
	tests := []struct {
		name             string
		payloadData      any
		expectedResult   layer4.Result
		expectedMessage  string
		assertionMessage string
	}{
		{
			name: "Active repository",
			payloadData: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Repository: si.Repository{
							Status: "active",
						},
					},
				},
			},
			expectedResult:   layer4.Passed,
			expectedMessage:  "Repo Status is active",
			assertionMessage: "Should pass when repository status is active",
		},
		{
			name: "Inactive repository",
			payloadData: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Repository: si.Repository{
							Status: "inactive",
						},
					},
				},
			},
			expectedResult:   layer4.NotApplicable,
			expectedMessage:  "Repo Status is inactive",
			assertionMessage: "Should be not applicable when repository status is inactive",
		},
		{
			name: "Empty status",
			payloadData: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Repository: si.Repository{
							Status: "",
						},
					},
				},
			},
			expectedResult:   layer4.NotApplicable,
			expectedMessage:  "Repo Status is ",
			assertionMessage: "Should be not applicable when repository status is empty",
		},
		{
			name:             "Malformed payload type",
			payloadData:      "not a payload",
			expectedResult:   layer4.Unknown,
			expectedMessage:  "Malformed assessment: expected payload type data.Payload, got string (not a payload)",
			assertionMessage: "Should return Unknown for wrong payload type",
		},
	}

	for _, tt := range tests {
		result, message := IsActive(tt.payloadData)
		assert.Equal(t, tt.expectedResult, result, tt.assertionMessage)
		assert.Equal(t, tt.expectedMessage, message, tt.assertionMessage)
	}
}
