package vuln_management

import (
	"testing"

	"github.com/gemaraproj/go-gemara"
	"github.com/ossf/pvtr-github-repo-scanner/data"
	"github.com/ossf/si-tooling/v2/si"
	"github.com/stretchr/testify/assert"
)

func ptrTo[T any](v T) *T { return &v }

func TestDetectSastToolInInsights(t *testing.T) {
	tests := []struct {
		name  string
		tools []si.SecurityTool
		want  []string
	}{
		{
			name: "SAST tool integrated in CI is detected by name",
			tools: []si.SecurityTool{{
				Name:        "CodeQL",
				Type:        "SAST",
				Integration: si.SecurityToolIntegration{Ci: true},
			}},
			want: []string{"CodeQL"},
		},
		{
			name: "SAST tool without a name falls back to a generic label",
			tools: []si.SecurityTool{{
				Type:        "SAST",
				Integration: si.SecurityToolIntegration{Ci: true},
			}},
			want: []string{"SAST"},
		},
		{
			name: "SAST tool only run adhoc or at release is not CI evidence",
			tools: []si.SecurityTool{{
				Name:        "CodeQL",
				Type:        "SAST",
				Integration: si.SecurityToolIntegration{Adhoc: true, Release: true},
			}},
			want: nil,
		},
		{
			name: "Non-SAST tool is ignored",
			tools: []si.SecurityTool{{
				Name:        "Dependabot",
				Type:        "SCA",
				Integration: si.SecurityToolIntegration{Ci: true},
			}},
			want: nil,
		},
	}

	for _, test := range tests {
		got := detectSastToolInInsights(test.tools)
		assert.Equal(t, test.want, got, test.name)
	}
}

const codeqlWorkflow = `name: CodeQL
on:
  pull_request:
    branches: [main]
jobs:
  analyze:
    name: Analyze
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: github/codeql-action/analyze@v3
`

const semgrepRunWorkflow = `name: Security
on: [push]
jobs:
  sast:
    runs-on: ubuntu-latest
    steps:
      - run: semgrep ci
`

const scheduledCodeqlWorkflow = `name: Scheduled CodeQL
on:
  schedule:
    - cron: '0 0 * * 0'
jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - uses: github/codeql-action/analyze@v3
`

const nonSastWorkflow = `name: Build
on: [pull_request]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: make build
`

func TestDetectSastInWorkflows(t *testing.T) {
	tests := []struct {
		name         string
		files        []data.WorkflowFile
		wantContains []string
		wantEmpty    bool
	}{
		{
			name:         "CodeQL action on pull_request is detected with job correlation names",
			files:        []data.WorkflowFile{{Name: "codeql.yml", Path: ".github/workflows/codeql.yml", Content: codeqlWorkflow}},
			wantContains: []string{"codeql", "CodeQL", "Analyze"},
		},
		{
			name:         "semgrep run command on push is detected",
			files:        []data.WorkflowFile{{Name: "sec.yml", Path: ".github/workflows/sec.yml", Content: semgrepRunWorkflow}},
			wantContains: []string{"semgrep"},
		},
		{
			name:      "SAST that only runs on a schedule is not evidence of a gate on changes",
			files:     []data.WorkflowFile{{Name: "codeql.yml", Path: ".github/workflows/codeql.yml", Content: scheduledCodeqlWorkflow}},
			wantEmpty: true,
		},
		{
			name:      "workflow without a SAST tool is not detected",
			files:     []data.WorkflowFile{{Name: "build.yml", Path: ".github/workflows/build.yml", Content: nonSastWorkflow}},
			wantEmpty: true,
		},
		{
			name:      "truncated and non-workflow files are skipped",
			files:     []data.WorkflowFile{{Name: "codeql.yml", Path: ".github/workflows/codeql.yml", Truncated: true}, {Name: "README.md", Content: codeqlWorkflow}},
			wantEmpty: true,
		},
	}

	for _, test := range tests {
		got := detectSastInWorkflows(test.files)
		if test.wantEmpty {
			assert.Empty(t, got, test.name)
			continue
		}
		for _, want := range test.wantContains {
			assert.Contains(t, got, want, test.name)
		}
	}
}

func TestRequiredCheckMatchesSast(t *testing.T) {
	tests := []struct {
		name             string
		requiredContexts []string
		sastSources      []string
		want             bool
	}{
		{
			name:             "check context carrying a known SAST identifier matches",
			requiredContexts: []string{"CodeQL"},
			sastSources:      []string{"some-other-source"},
			want:             true,
		},
		{
			name:             "check context matching a SAST job name matches",
			requiredContexts: []string{"Analyze"},
			sastSources:      []string{"codeql", "Analyze"},
			want:             true,
		},
		{
			name:             "unrelated required check does not match",
			requiredContexts: []string{"build", "unit-tests"},
			sastSources:      []string{"codeql", "Analyze"},
			want:             false,
		},
		{
			name:             "short generic source token does not produce a spurious match",
			requiredContexts: []string{"lint"},
			sastSources:      []string{"ci"},
			want:             false,
		},
		{
			name:             "no required contexts cannot match",
			requiredContexts: nil,
			sastSources:      []string{"codeql"},
			want:             false,
		},
	}

	for _, test := range tests {
		got := requiredCheckMatchesSast(test.requiredContexts, test.sastSources)
		assert.Equal(t, test.want, got, test.name)
	}
}

func TestEvaluateSastEnforcement(t *testing.T) {
	tests := []struct {
		name             string
		sastSources      []string
		requiredContexts []string
		adminObservable  bool
		wantResult       gemara.Result
	}{
		{
			name:             "SAST in CI enforced by a matching required check passes",
			sastSources:      []string{"codeql", "Analyze"},
			requiredContexts: []string{"Analyze"},
			wantResult:       gemara.Passed,
		},
		{
			name:             "SAST in CI with required checks but no match needs review",
			sastSources:      []string{"codeql", "Analyze"},
			requiredContexts: []string{"build"},
			wantResult:       gemara.NeedsReview,
		},
		{
			name:            "SAST in CI with admin-observable absence of required checks fails",
			sastSources:     []string{"codeql"},
			adminObservable: true,
			wantResult:      gemara.Failed,
		},
		{
			name:            "SAST in CI with unobservable branch protection needs review",
			sastSources:     []string{"codeql"},
			adminObservable: false,
			wantResult:      gemara.NeedsReview,
		},
		{
			name:        "no SAST in CI fails",
			sastSources: nil,
			wantResult:  gemara.Failed,
		},
	}

	for _, test := range tests {
		result, message, _ := evaluateSastEnforcement(test.sastSources, test.requiredContexts, test.adminObservable)
		assert.Equal(t, test.wantResult, result, test.name)
		assert.NotEmpty(t, message, test.name)
	}
}

func TestHasVulnerabilityDisclosurePolicy(t *testing.T) {
	tests := []struct {
		name                  string
		policy                *si.URL
		securityPolicyEnabled bool
		securityMdPresent     bool
		expectedResult        gemara.Result
		expectedMessage       string
	}{
		{
			name:            "Security Insights policy present",
			policy:          ptrTo(si.URL("https://example.com/SECURITY.md")),
			expectedResult:  gemara.Passed,
			expectedMessage: "Vulnerability disclosure policy was specified in Security Insights data",
		},
		{
			name:                  "No SI policy but GitHub security policy enabled warrants review",
			securityPolicyEnabled: true,
			expectedResult:        gemara.NeedsReview,
			expectedMessage:       "GitHub reports a security policy is enabled for the repository; its CVD policy content and response timeframe need human confirmation",
		},
		{
			name:              "No SI policy but SECURITY.md present warrants review",
			securityMdPresent: true,
			expectedResult:    gemara.NeedsReview,
			expectedMessage:   "A SECURITY.md file was found in the repository via GitHub; its CVD policy content and response timeframe need human confirmation",
		},
		{
			name:            "No policy from any source",
			expectedResult:  gemara.Failed,
			expectedMessage: "No vulnerability disclosure policy found in Security Insights data, a GitHub security policy, or a SECURITY.md file",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload := data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: &si.Project{
							VulnerabilityReporting: si.VulnerabilityReporting{
								Policy: test.policy,
							},
						},
					},
					SecurityPolicy: data.SecurityPolicy{Present: test.securityMdPresent},
				},
				GraphqlRepoData: &data.GraphqlRepoData{},
			}
			payload.Repository.IsSecurityPolicyEnabled = test.securityPolicyEnabled

			result, message, _ := HasVulnerabilityDisclosurePolicy(payload)
			assert.Equal(t, test.expectedResult, result)
			assert.Equal(t, test.expectedMessage, message)
		})
	}
}

func TestHasPrivateVulnerabilityReporting(t *testing.T) {
	tests := []struct {
		name            string
		payload         data.Payload
		expectedResult  gemara.Result
		expectedMessage string
	}{
		{
			name:            "Private reporting via vulnerability contact email",
			expectedResult:  gemara.Passed,
			expectedMessage: "Private vulnerability reporting available via dedicated contact email in Security Insights data",
			payload: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: &si.Project{
							VulnerabilityReporting: si.VulnerabilityReporting{
								ReportsAccepted: true,
								Contact: &si.Contact{
									Email: ptrTo(si.Email("security@example.com")),
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
			expectedResult:  gemara.Passed,
			expectedMessage: "Private vulnerability reporting available via security champions contact in Security Insights data",
			payload: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: &si.Project{
							VulnerabilityReporting: si.VulnerabilityReporting{
								ReportsAccepted: true,
								Contact:         &si.Contact{},
							},
						},
						Repository: &si.Repository{
							SecurityPosture: si.SecurityPosture{
								Champions: []si.Contact{
									{
										Name:  "Security Champion",
										Email: ptrTo(si.Email("champion@example.com")),
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
			name:            "No SI contact but GitHub private reporting enabled",
			expectedResult:  gemara.Passed,
			expectedMessage: "No Security Insights contact, but GitHub private vulnerability reporting is enabled for the repository",
			payload: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: &si.Project{VulnerabilityReporting: si.VulnerabilityReporting{Contact: &si.Contact{}}},
					},
					PrivateVulnReporting: data.PrivateVulnReporting{Enabled: true, Known: true},
				},
				GraphqlRepoData: &data.GraphqlRepoData{},
			},
		},
		{
			name:            "SI reports accepted but no contact and private reporting status unknown",
			expectedResult:  gemara.NeedsReview,
			expectedMessage: "No private vulnerability reporting contact in Security Insights data and GitHub private vulnerability reporting status could not be determined",
			payload: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: &si.Project{
							VulnerabilityReporting: si.VulnerabilityReporting{
								ReportsAccepted: true,
								Contact:         &si.Contact{},
							},
						},
						Repository: &si.Repository{
							SecurityPosture: si.SecurityPosture{
								Champions: []si.Contact{{Name: "Champion Without Email"}},
							},
						},
					},
				},
				GraphqlRepoData: &data.GraphqlRepoData{},
			},
		},
		{
			name:            "No SI contact and private reporting observed disabled",
			expectedResult:  gemara.Failed,
			expectedMessage: "No private vulnerability reporting contact in Security Insights data and GitHub private vulnerability reporting is disabled",
			payload: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: &si.Project{VulnerabilityReporting: si.VulnerabilityReporting{Contact: &si.Contact{}}},
					},
					PrivateVulnReporting: data.PrivateVulnReporting{Enabled: false, Known: true},
				},
				GraphqlRepoData: &data.GraphqlRepoData{},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, message, _ := HasPrivateVulnerabilityReporting(test.payload)
			assert.Equal(t, test.expectedResult, result)
			assert.Equal(t, test.expectedMessage, message)
		})
	}
}

func TestHasSecContact(t *testing.T) {
	tests := []struct {
		name            string
		payload         data.Payload
		expectedResult  gemara.Result
		expectedMessage string
	}{
		{
			name:            "Security Insights contact email",
			expectedResult:  gemara.Passed,
			expectedMessage: "Security contacts were specified in Security Insights data",
			payload: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: &si.Project{
							VulnerabilityReporting: si.VulnerabilityReporting{
								Contact: &si.Contact{Email: ptrTo(si.Email("security@example.com"))},
							},
						},
					},
				},
				GraphqlRepoData: &data.GraphqlRepoData{},
			},
		},
		{
			name:            "Security Insights champion email",
			expectedResult:  gemara.Passed,
			expectedMessage: "Security contacts were specified in Security Insights data",
			payload: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: &si.Project{VulnerabilityReporting: si.VulnerabilityReporting{Contact: &si.Contact{}}},
						Repository: &si.Repository{
							SecurityPosture: si.SecurityPosture{
								Champions: []si.Contact{{Email: ptrTo(si.Email("champion@example.com"))}},
							},
						},
					},
				},
				GraphqlRepoData: &data.GraphqlRepoData{},
			},
		},
		{
			name:            "SECURITY.md with contact email",
			expectedResult:  gemara.Passed,
			expectedMessage: "An email address was found in SECURITY.md (contact email)",
			payload: secContactPayload(data.SecurityPolicy{
				Present: true,
				Content: "Please email security@example.com to report issues.",
			}, data.PrivateVulnReporting{}),
		},
		{
			name:            "SECURITY.md with reporting instructions",
			expectedResult:  gemara.Passed,
			expectedMessage: "An email address was found in SECURITY.md (private-reporting instructions)",
			payload: secContactPayload(data.SecurityPolicy{
				Present: true,
				Content: "Use GitHub private vulnerability reporting to disclose issues.",
			}, data.PrivateVulnReporting{}),
		},
		{
			name:            "SECURITY.md with reporting URL",
			expectedResult:  gemara.Passed,
			expectedMessage: "An email address was found in SECURITY.md (reporting URL)",
			payload: secContactPayload(data.SecurityPolicy{
				Present: true,
				Content: "See our disclosure form at https://example.com/report for details.",
			}, data.PrivateVulnReporting{}),
		},
		{
			name:            "No SI or SECURITY.md but private reporting enabled",
			expectedResult:  gemara.Passed,
			expectedMessage: "GitHub private vulnerability reporting is enabled as a documented reporting channel",
			payload:         secContactPayload(data.SecurityPolicy{}, data.PrivateVulnReporting{Enabled: true, Known: true}),
		},
		{
			name:            "SECURITY.md present but no recognizable contact",
			expectedResult:  gemara.NeedsReview,
			expectedMessage: "A SECURITY.md file was found via GitHub but no recognizable security contact could be identified in it",
			payload: secContactPayload(data.SecurityPolicy{
				Present: true,
				Content: "We take security seriously and patch issues promptly.",
			}, data.PrivateVulnReporting{}),
		},
		{
			name:            "No contact evidence from any source",
			expectedResult:  gemara.Failed,
			expectedMessage: "No security contact found in Security Insights data, a SECURITY.md file, or GitHub private vulnerability reporting",
			payload:         secContactPayload(data.SecurityPolicy{}, data.PrivateVulnReporting{}),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, message, _ := HasSecContact(test.payload)
			assert.Equal(t, test.expectedResult, result)
			assert.Equal(t, test.expectedMessage, message)
		})
	}
}

// secContactPayload builds a payload with empty Security Insights contact data so
// HasSecContact falls through to the GitHub-derived signals under test.
func secContactPayload(policy data.SecurityPolicy, pvr data.PrivateVulnReporting) data.Payload {
	return data.Payload{
		RestData: &data.RestData{
			Insights: si.SecurityInsights{
				Project:    &si.Project{VulnerabilityReporting: si.VulnerabilityReporting{Contact: &si.Contact{}}},
				Repository: &si.Repository{},
			},
			SecurityPolicy:       policy,
			PrivateVulnReporting: pvr,
		},
		GraphqlRepoData: &data.GraphqlRepoData{},
	}
}
