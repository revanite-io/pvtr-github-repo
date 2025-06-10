package reusable_steps

import (
	"testing"

	"github.com/ossf/si-tooling/v2/si"
	"github.com/revanite-io/pvtr-github-repo/data"
	"github.com/revanite-io/sci/pkg/layer4"
	"github.com/stretchr/testify/assert"
)

type testingData struct {
	expectedResult   layer4.Result
	expectedMessage  string
	payloadData      interface{}
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
		result, message := HasDependencyManagementPolicy(test.payloadData, nil)
		assert.Equal(t, test.expectedResult, result, test.assertionMessage)
		assert.Equal(t, test.expectedMessage, message, test.assertionMessage)
	}

}
