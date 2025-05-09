package vuln_management

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

func TestSastToolDefined(t *testing.T) {

	testData := []testingData{
		{
			expectedResult:   layer4.Passed,
			expectedMessage:  "Static Application Security Testing documented in Security Insights",
			assertionMessage: "Test for SAST integration enabled",
			payloadData: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Repository: &si.Repository{
							SecurityPosture: si.SecurityPosture{
								Tools: []si.SecurityTool{
									{
										Type: "SAST",
										Integration: si.SecurityToolIntegration{
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
						Repository: &si.Repository{
							SecurityPosture: si.SecurityPosture{
								Tools: []si.SecurityTool{
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
						Repository: &si.Repository{
							SecurityPosture: si.SecurityPosture{
								Tools: []si.SecurityTool{
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
						Repository: &si.Repository{
							SecurityPosture: si.SecurityPosture{},
						},
					},
				},
			},
		},
	}

	for _, test := range testData {
		result, message := sastToolDefined(test.payloadData, nil)

		assert.Equal(t, test.expectedResult, result, test.assertionMessage)
		assert.Equal(t, test.expectedMessage, message, test.assertionMessage)
	}

}

func TestHasSecContact(t *testing.T) {
	arbitraryEmail := si.NewEmail("champion@example.com")

	tests := []struct {
		name            string
		payloadData     interface{}
		expectedResult  layer4.Result
		expectedMessage string
	}{
		{
			name: "Valid contact in VulnerabilityReporting",
			payloadData: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: &si.Project{
							VulnerabilityReporting: si.VulnerabilityReporting{
								Contact: &si.Contact{
									Name:  "Security Team",
									Email: &arbitraryEmail,
								},
							},
						},
					},
				},
			},
			expectedResult:  layer4.Passed,
			expectedMessage: "Security contacts were specified in Security Insights data",
		},
		{
			name: "Valid contact in Champions",
			payloadData: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Repository: &si.Repository{
							SecurityPosture: si.SecurityPosture{
								Champions: []si.Contact{
									{
										Name:  "Security Champion",
										Email: &arbitraryEmail,
									},
								},
							},
						},
					},
				},
			},
			expectedResult:  layer4.Passed,
			expectedMessage: "Security contacts were specified in Security Insights data",
		},
		{
			name: "No security contacts specified",
			payloadData: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: &si.Project{
							VulnerabilityReporting: si.VulnerabilityReporting{},
						},
						Repository: &si.Repository{
							SecurityPosture: si.SecurityPosture{
								Champions: []si.Contact{},
							},
						},
					},
				},
			},
			expectedResult:  layer4.Failed,
			expectedMessage: "Security contacts were not specified in Security Insights data",
		},
		{
			name:            "Invalid payload data",
			payloadData:     nil,
			expectedResult:  layer4.Unknown,
			expectedMessage: "Malformed assessment: expected payload type data.Payload, got <nil> (<nil>)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, message := hasSecContact(test.payloadData, nil)
			assert.Equal(t, test.expectedResult, result)
			assert.Equal(t, test.expectedMessage, message)
		})
	}
}
