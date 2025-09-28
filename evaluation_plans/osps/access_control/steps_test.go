package access_control

import (
	"testing"

	"github.com/ossf/gemara/layer4"
	"github.com/revanite-io/pvtr-github-repo/data"
	"github.com/stretchr/testify/assert"
)

type FakeRepositoryMetadata struct {
	data.RepositoryMetadata
	twoFactorEnabled *bool
}

func (f *FakeRepositoryMetadata) IsMFARequiredForAdministrativeActions() *bool {
	return f.twoFactorEnabled
}

func stubRepoMetadata(twoFactorEnabled *bool) *FakeRepositoryMetadata {
	return &FakeRepositoryMetadata{
		twoFactorEnabled: twoFactorEnabled,
	}
}

func Test_orgRequiresMFA(t *testing.T) {
	trueVal := true
	falseVal := false

	tests := []struct {
		name        string
		payload     data.Payload
		wantResult  layer4.Result
		wantMessage string
	}{
		{
			name: "org requires MFA",
			payload: data.Payload{
				RepositoryMetadata: stubRepoMetadata(&trueVal),
			},
			wantResult:  layer4.Passed,
			wantMessage: "Two-factor authentication is configured as required by the parent organization",
		},
		{
			name: "org does not require MFA",
			payload: data.Payload{
				RepositoryMetadata: stubRepoMetadata(&falseVal),
			},
			wantResult:  layer4.Failed,
			wantMessage: "Two-factor authentication is NOT configured as required by the parent organization",
		},
		{
			name: "unable to evaluate MFA requirement",
			payload: data.Payload{
				RepositoryMetadata: stubRepoMetadata(nil),
			},
			wantResult:  layer4.NeedsReview,
			wantMessage: "Not evaluated. Two-factor authentication evaluation requires a token with org:admin permissions, or manual review",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, gotMessage := orgRequiresMFA(tt.payload, map[string]*layer4.Change{})
			assert.Equal(t, tt.wantResult, gotResult)
			assert.Equal(t, tt.wantMessage, gotMessage)
		})
	}
}

// See https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/enabling-features-for-your-repository/managing-github-actions-settings-for-a-repository#setting-the-permissions-of-the-github_token-for-your-repository
func Test_workflowDefaultReadPermissions(t *testing.T) {
	tests := []struct {
		name        string
		payload     data.Payload
		wantResult  layer4.Result
		wantMessage string
	}{
		{
			name: "read permissions and no PR permissions",
			payload: data.Payload{
				RestData: &data.RestData{
					WorkflowPermissions: data.WorkflowPermissions{
						DefaultPermissions:    "read", // read access for the contents and packages permissions
						CanApprovePullRequest: false,  // cannot create or approve PRs
					},
				},
			},
			wantResult:  layer4.Passed,
			wantMessage: "Workflow permissions default to read only.",
		},
		{
			name: "read permissions, but allows PR approvals",
			payload: data.Payload{
				RestData: &data.RestData{
					WorkflowPermissions: data.WorkflowPermissions{
						DefaultPermissions:    "read", // read access for the contents and packages permissions
						CanApprovePullRequest: true,   // can create & approve PRs
					},
				},
			},
			wantResult:  layer4.Failed,
			wantMessage: "Workflow permissions default to read only for contents and packages, but PR approval is permitted.",
		},
		{
			name: "write permissions and no PR permissions",
			payload: data.Payload{
				RestData: &data.RestData{
					WorkflowPermissions: data.WorkflowPermissions{
						DefaultPermissions:    "write", // read & write access for all permission scopes
						CanApprovePullRequest: false,   // cannot create or approve PRs (in theory at least)
					},
				},
			},
			wantResult:  layer4.Failed,
			wantMessage: "Workflow permissions default to read/write, but PR approval is forbidden.",
		},
		{
			name: "write permissions and PR permissions",
			payload: data.Payload{
				RestData: &data.RestData{
					WorkflowPermissions: data.WorkflowPermissions{
						DefaultPermissions:    "write",
						CanApprovePullRequest: true,
					},
				},
			},
			wantResult:  layer4.Failed,
			wantMessage: "Workflow permissions default to read/write and PR approval is permitted.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, gotMessage := workflowDefaultReadPermissions(tt.payload, map[string]*layer4.Change{})
			assert.Equal(t, tt.wantResult, gotResult)
			assert.Equal(t, tt.wantMessage, gotMessage)
		})
	}
}
