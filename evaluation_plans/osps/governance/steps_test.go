package governance

import (
	"testing"

	"github.com/ossf/si-tooling/v2/si"
	"github.com/revanite-io/pvtr-github-repo/data"
	"github.com/revanite-io/sci/pkg/layer4"
	"github.com/stretchr/testify/assert"
)

func TestHasContributionReviewPolicy(t *testing.T) {
	emptyURL := si.NewURL("")
	arbitraryURL := si.NewURL("https://example.com/review-policy")

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
						Repository: &si.Repository{},
					},
				},
			},
			expectedResult: layer4.Failed,
			expectedMsg:    "Code review guide was NOT specified in Security Insights data",
		},
		{
			name: "Review policy is empty",
			payload: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Repository: &si.Repository{
							Documentation: &si.RepositoryDocumentation{
								ReviewPolicy: &emptyURL,
							},
						},
					},
				},
			},
			expectedResult: layer4.Failed,
			expectedMsg:    "Code review guide was NOT specified in Security Insights data",
		},
		{
			name: "Review policy is provided",
			payload: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Repository: &si.Repository{
							Documentation: &si.RepositoryDocumentation{
								ReviewPolicy: &arbitraryURL,
							},
						},
					},
				},
			},
			expectedResult: layer4.Passed,
			expectedMsg:    "Code review guide was specified in Security Insights data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, message := hasContributionReviewPolicy(tt.payload, nil)
			assert.Equal(t, tt.expectedResult, result, "Unexpected result")
			assert.Equal(t, tt.expectedMsg, message, "Unexpected message")
		})
	}
}

func TestHasRolesAndResponsibilities(t *testing.T) {
	emptyGovernance := si.NewURL("")
	arbitraryGovernance := si.NewURL("https://example.com/governance")

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
						Repository: &si.Repository{},
					},
				},
			},
			expectedResult: layer4.Failed,
			expectedMsg:    "Roles and responsibilities were NOT specified in Security Insights data",
		},
		{
			name: "Governance is empty",
			payload: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Repository: &si.Repository{
							Documentation: &si.RepositoryDocumentation{
								Governance: &emptyGovernance,
							},
						},
					},
				},
			},
			expectedResult: layer4.Failed,
			expectedMsg:    "Roles and responsibilities were NOT specified in Security Insights data",
		},
		{
			name: "Governance is provided",
			payload: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Repository: &si.Repository{
							Documentation: &si.RepositoryDocumentation{
								Governance: &arbitraryGovernance,
							},
						},
					},
				},
			},
			expectedResult: layer4.Passed,
			expectedMsg:    "Roles and responsibilities were specified in Security Insights data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, message := hasRolesAndResponsibilities(tt.payload, nil)
			assert.Equal(t, tt.expectedResult, result, "Unexpected result")
			assert.Equal(t, tt.expectedMsg, message, "Unexpected message")
		})
	}
}

func TestProjectAdminsListed(t *testing.T) {
	tests := []struct {
		name           string
		payload        interface{}
		expectedResult layer4.Result
		expectedMsg    string
	}{
		{
			name: "No administrators provided",
			payload: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: &si.Project{},
					},
				},
			},
			expectedResult: layer4.Failed,
			expectedMsg:    "Project admins were NOT specified in Security Insights data",
		},
		{
			name: "Administrators list is empty",
			payload: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: &si.Project{
							Administrators: []si.Contact{},
						},
					},
				},
			},
			expectedResult: layer4.Failed,
			expectedMsg:    "Project admins were NOT specified in Security Insights data",
		},
		{
			name: "Administrators are provided",
			payload: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: &si.Project{
							Administrators: []si.Contact{
								{
									Name:    "admin1",
									Primary: false,
								},
								{
									Primary: true,
									Name:    "admin2",
								},
							},
						},
					},
				},
			},
			expectedResult: layer4.Passed,
			expectedMsg:    "Project admins were specified in Security Insights data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, message := projectAdminsListed(tt.payload, nil)
			assert.Equal(t, tt.expectedResult, result, "Unexpected result")
			assert.Equal(t, tt.expectedMsg, message, "Unexpected message")
		})
	}
}

func TestCoreTeamIsListed(t *testing.T) {
	tests := []struct {
		name           string
		payload        interface{}
		expectedResult layer4.Result
		expectedMsg    string
	}{
		{
			name: "No core team provided",
			payload: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Repository: &si.Repository{},
					},
				},
			},
			expectedResult: layer4.Failed,
			expectedMsg:    "Core team was NOT specified in Security Insights data",
		},
		{
			name: "Core team list is empty",
			payload: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Repository: &si.Repository{
							CoreTeam: []si.Contact{},
						},
					},
				},
			},
			expectedResult: layer4.Failed,
			expectedMsg:    "Core team was NOT specified in Security Insights data",
		},
		{
			name: "Core team is provided",
			payload: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Repository: &si.Repository{
							CoreTeam: []si.Contact{
								{
									Name:    "core1",
									Primary: false,
								},
								{
									Primary: true,
									Name:    "core2",
								},
							},
						},
					},
				},
			},
			expectedResult: layer4.Passed,
			expectedMsg:    "Core team was specified in Security Insights data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, message := coreTeamIsListed(tt.payload, nil)
			assert.Equal(t, tt.expectedResult, result, "Unexpected result")
			assert.Equal(t, tt.expectedMsg, message, "Unexpected message")
		})
	}
}
