package docs

import (
	"testing"

	"github.com/ossf/si-tooling/v2/si"
	"github.com/revanite-io/pvtr-github-repo/data"
	"github.com/revanite-io/sci/pkg/layer4"
	"github.com/stretchr/testify/assert"
)

func TestHasUserGuides(t *testing.T) {
	emptyURL := si.NewURL("")
	arbitraryURL := si.NewURL("https://example.com/user-guide")

	tests := []struct {
		name           string
		payload        interface{}
		expectedResult layer4.Result
		expectedMsg    string
	}{
		{
			name: "No documentation provided",
			payload: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: &si.Project{},
					},
				},
			},
			expectedResult: layer4.Failed,
			expectedMsg:    "User guide was NOT specified in Security Insights data",
		},
		{
			name: "Detailed guide is empty",
			payload: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: &si.Project{
							Documentation: &si.ProjectDocumentation{
								DetailedGuide: &emptyURL,
							},
						},
					},
				},
			},
			expectedResult: layer4.Failed,
			expectedMsg:    "User guide was NOT specified in Security Insights data",
		},
		{
			name: "Detailed guide is provided",
			payload: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: &si.Project{
							Documentation: &si.ProjectDocumentation{
								DetailedGuide: &arbitraryURL,
							},
						},
					},
				},
			},
			expectedResult: layer4.Passed,
			expectedMsg:    "User guide was specified in Security Insights data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, message := hasUserGuides(tt.payload, nil)
			assert.Equal(t, tt.expectedResult, result, "Unexpected result")
			assert.Equal(t, tt.expectedMsg, message, "Unexpected message")
		})
	}
}
func TestAcceptsVulnReports(t *testing.T) {
	tests := []struct {
		name           string
		payload        interface{}
		expectedResult layer4.Result
		expectedMsg    string
	}{
		{
			name: "Vulnerability reports not accepted",
			payload: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: &si.Project{
							VulnerabilityReporting: si.VulnerabilityReporting{
								ReportsAccepted: false,
							},
						},
					},
				},
			},
			expectedResult: layer4.Failed,
			expectedMsg:    "Repository does not accept vulnerability reports",
		},
		{
			name: "Vulnerability reports accepted",
			payload: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: &si.Project{
							VulnerabilityReporting: si.VulnerabilityReporting{
								ReportsAccepted: true,
							},
						},
					},
				},
			},
			expectedResult: layer4.Passed,
			expectedMsg:    "Repository accepts vulnerability reports",
		},
		{
			name: "Vulnerability reporting data missing",
			payload: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: &si.Project{},
					},
				},
			},
			expectedResult: layer4.Failed,
			expectedMsg:    "Repository does not accept vulnerability reports",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, message := acceptsVulnReports(tt.payload, nil)
			assert.Equal(t, tt.expectedResult, result, "Unexpected result")
			assert.Equal(t, tt.expectedMsg, message, "Unexpected message")
		})
	}
}
func TestHasSignatureVerificationGuide(t *testing.T) {
	emptyURL := si.NewURL("")
	arbitraryURL := si.NewURL("https://example.com/signature-verification-guide")

	tests := []struct {
		name           string
		payload        interface{}
		expectedResult layer4.Result
		expectedMsg    string
	}{
		{
			name: "No documentation provided",
			payload: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: &si.Project{},
					},
				},
			},
			expectedResult: layer4.Failed,
			expectedMsg:    "Signature verification guide was NOT specified in Security Insights data",
		},
		{
			name: "Signature verification guide is empty",
			payload: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: &si.Project{
							Documentation: &si.ProjectDocumentation{
								SignatureVerification: &emptyURL,
							},
						},
					},
				},
			},
			expectedResult: layer4.Failed,
			expectedMsg:    "Signature verification guide was NOT specified in Security Insights data",
		},
		{
			name: "Signature verification guide is provided",
			payload: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: &si.Project{
							Documentation: &si.ProjectDocumentation{
								SignatureVerification: &arbitraryURL,
							},
						},
					},
				},
			},
			expectedResult: layer4.Passed,
			expectedMsg:    "Signature verification guide was specified in Security Insights data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, message := hasSignatureVerificationGuide(tt.payload, nil)
			assert.Equal(t, tt.expectedResult, result, "Unexpected result")
			assert.Equal(t, tt.expectedMsg, message, "Unexpected message")
		})
	}
}
