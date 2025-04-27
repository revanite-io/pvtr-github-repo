package access_control

import (
	"testing"

	"github.com/revanite-io/pvtr-github-repo/data"
	"github.com/revanite-io/sci/pkg/layer4"
	"github.com/stretchr/testify/assert"
)

type FakeRepositoryMetadata struct {
	data.RepositoryMetadata
	twoFactorEnabled               bool
	unableToEvaluateMFARequirement bool
}

func (f *FakeRepositoryMetadata) IsMFARequiredForAdministrativeActions() bool {
	return f.twoFactorEnabled
}

func (f *FakeRepositoryMetadata) UnableToEvaluateMFARequirement() bool {
	return f.unableToEvaluateMFARequirement
}

func stubRepoMetadata(twoFactorEnabled bool, unableToEvaluateMFARequirement bool) *FakeRepositoryMetadata {
	return &FakeRepositoryMetadata{
		twoFactorEnabled:               twoFactorEnabled,
		unableToEvaluateMFARequirement: unableToEvaluateMFARequirement,
	}
}

func Test_orgRequiresMFA(t *testing.T) {
	tests := []struct {
		name        string
		payload     data.Payload
		wantResult  layer4.Result
		wantMessage string
	}{
		{
			name: "org requires MFA",
			payload: data.Payload{
				RepositoryMetadata: stubRepoMetadata(true, false),
			},
			wantResult:  layer4.Passed,
			wantMessage: "Two-factor authentication is configured as required by the parent organization",
		},
		{
			name: "org does not require MFA",
			payload: data.Payload{
				RepositoryMetadata: stubRepoMetadata(false, false),
			},
			wantResult:  layer4.Failed,
			wantMessage: "Two-factor authentication is NOT configured as required by the parent organization",
		},
		{
			name: "unable to evaluate MFA requirement",
			payload: data.Payload{
				RepositoryMetadata: stubRepoMetadata(false, true),
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
