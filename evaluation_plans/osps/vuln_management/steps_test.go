package vuln_management

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
	payloadData      interface{}
	assertionMessage string
}

func TestSastToolDefined(t *testing.T) {

	testData := []testingData{
		{
			expectedResult:   layer4.Passed,
			expectedMessage:  "Static Application Security Testing documented in Security Insights",
			assertionMessage: "Test for SAST integration enabled",
			payloadData: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Repository: si.Repository{
							Security: si.SecurityInfo{
								Tools: []si.Tool{
									{
										Type: "SAST",
										Integration: si.Integration{
											Adhoc: true,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			expectedResult:   layer4.Failed,
			expectedMessage:  "No Static Application Security Testing documented in Security Insights",
			assertionMessage: "Test for SAST integration present but not explicitly enabled",
			payloadData: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Repository: si.Repository{
							Security: si.SecurityInfo{
								Tools: []si.Tool{
									{
										Type: "SAST",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			expectedResult:   layer4.Failed,
			expectedMessage:  "No Static Application Security Testing documented in Security Insights",
			assertionMessage: "Test for Non SAST tool defined",
			payloadData: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Repository: si.Repository{
							Security: si.SecurityInfo{
								Tools: []si.Tool{
									{
										Type: "NotSast",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			expectedResult:   layer4.Failed,
			expectedMessage:  "No Static Application Security Testing documented in Security Insights",
			assertionMessage: "Test for no tools defined",
			payloadData: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Repository: si.Repository{
							Security: si.SecurityInfo{},
						},
					},
				},
			},
		},
	}

	for _, test := range testData {
		result, message := SastToolDefined(test.payloadData)

		assert.Equal(t, test.expectedResult, result, test.assertionMessage)
		assert.Equal(t, test.expectedMessage, message, test.assertionMessage)
	}

}

func TestHasVulnerabilityDisclosurePolicy(t *testing.T) {
	tests := []struct {
		name            string
		payloadData     any
		expectedResult  layer4.Result
		expectedMessage string
	}{
		{
			name:            "Vulnerability disclosure policy present",
			expectedResult:  layer4.Passed,
			expectedMessage: "Vulnerability disclosure policy was specified in Security Insights data",
			payloadData: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: si.Project{
							Vulnerability: si.VulnReport{
								SecurityPolicy: "https://example.com/SECURITY.md",
							},
						},
					},
				},
				GraphqlRepoData: &data.GraphqlRepoData{},
			},
		},
		{
			name:            "Vulnerability disclosure policy missing",
			expectedResult:  layer4.Failed,
			expectedMessage: "Vulnerability disclosure policy was NOT specified in Security Insights data",
			payloadData: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: si.Project{
							Vulnerability: si.VulnReport{
								SecurityPolicy: "",
							},
						},
					},
				},
				GraphqlRepoData: &data.GraphqlRepoData{},
			},
		},
		{
			name:            "Invalid payload",
			expectedResult:  layer4.Unknown,
			expectedMessage: "Malformed assessment: expected payload type data.Payload, got string (invalid_payload)",
			payloadData:     "invalid_payload",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, message := HasVulnerabilityDisclosurePolicy(test.payloadData)
			assert.Equal(t, test.expectedResult, result)
			assert.Equal(t, test.expectedMessage, message)
		})
	}
}

func TestHasPrivateVulnerabilityReporting(t *testing.T) {
	tests := []struct {
		name            string
		payloadData     any
		expectedResult  layer4.Result
		expectedMessage string
	}{
		{
			name:            "Private reporting via vulnerability contact email",
			expectedResult:  layer4.Passed,
			expectedMessage: "Private vulnerability reporting available via dedicated contact email in Security Insights data",
			payloadData: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: si.Project{
							Vulnerability: si.VulnReport{
								ReportsAccepted: true,
								Contact: si.Contact{
									Email: "security@example.com",
								},
							},
						},
					},
				},
				GraphqlRepoData: &data.GraphqlRepoData{},
			},
		},
		{
			name:            "Private reporting via security champions",
			expectedResult:  layer4.Passed,
			expectedMessage: "Private vulnerability reporting available via security champions contact in Security Insights data",
			payloadData: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: si.Project{
							Vulnerability: si.VulnReport{
								ReportsAccepted: true,
								Contact: si.Contact{
									Email: "",
								},
							},
						},
						Repository: si.Repository{
							Security: si.SecurityInfo{
								Champions: []si.Contact{
									{
										Name:  "Security Champion",
										Email: "champion@example.com",
									},
								},
							},
						},
					},
				},
				GraphqlRepoData: &data.GraphqlRepoData{},
			},
		},
		{
			name:            "Reports not accepted",
			expectedResult:  layer4.Failed,
			expectedMessage: "Project does not accept vulnerability reports according to Security Insights data",
			payloadData: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: si.Project{
							Vulnerability: si.VulnReport{
								ReportsAccepted: false,
								Contact: si.Contact{
									Email: "security@example.com",
								},
							},
						},
					},
				},
				GraphqlRepoData: &data.GraphqlRepoData{},
			},
		},
		{
			name:            "No contact methods available",
			expectedResult:  layer4.Failed,
			expectedMessage: "No private vulnerability reporting contact method found in Security Insights data",
			payloadData: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: si.Project{
							Vulnerability: si.VulnReport{
								ReportsAccepted: true,
								Contact: si.Contact{
									Email: "",
								},
							},
						},
						Repository: si.Repository{
							Security: si.SecurityInfo{
								Champions: []si.Contact{
									{
										Name:  "Champion Without Email",
										Email: "",
									},
								},
							},
						},
					},
				},
				GraphqlRepoData: &data.GraphqlRepoData{},
			},
		},
		{
			name:            "Invalid payload",
			expectedResult:  layer4.Unknown,
			expectedMessage: "Malformed assessment: expected payload type data.Payload, got string (invalid_payload)",
			payloadData:     "invalid_payload",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, message := HasPrivateVulnerabilityReporting(test.payloadData)
			assert.Equal(t, test.expectedResult, result)
			assert.Equal(t, test.expectedMessage, message)
		})
	}
}
