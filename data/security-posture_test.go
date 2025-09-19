package data

import (
	"testing"

	"github.com/google/go-github/v74/github"
	"github.com/ossf/si-tooling/v2/si"
	"github.com/stretchr/testify/assert"
)

func TestRepoSecurityPostureMethods(t *testing.T) {
	rsp := &RepoSecurityPosture{
		preventsSecretPushing:           true,
		scansForSecrets:                 true,
		definesPolicyForHandlingSecrets: false,
	}

	assert.True(t, rsp.PreventsPushingSecrets())
	assert.True(t, rsp.ScansForSecrets())
	assert.False(t, rsp.DefinesPolicyForHandlingSecrets())
}

func TestBuildSecurityPosture_NoSecurityConfig(t *testing.T) {
	repo := &github.Repository{}
	rd := RestData{}
	sp, err := buildSecurityPosture(repo, rd)
	assert.NoError(t, err)
	assert.NotNil(t, sp)
	assert.False(t, sp.PreventsPushingSecrets())
	assert.False(t, sp.ScansForSecrets())
	assert.False(t, sp.DefinesPolicyForHandlingSecrets())
}

func TestBuildSecurityPosture_SecretScanningEnabled(t *testing.T) {
	repo := &github.Repository{
		SecurityAndAnalysis: &github.SecurityAndAnalysis{
			SecretScanning: &github.SecretScanning{
				Status: github.Ptr("enabled"),
			},
		},
	}
	rd := RestData{}
	sp, err := buildSecurityPosture(repo, rd)
	assert.NoError(t, err)
	assert.True(t, sp.PreventsPushingSecrets())
	assert.True(t, sp.ScansForSecrets())
}

func TestBuildSecurityPosture_SecretScanningDisabledButInsightsTooling(t *testing.T) {
	repo := &github.Repository{
		SecurityAndAnalysis: &github.SecurityAndAnalysis{
			SecretScanning: &github.SecretScanning{
				Status: github.Ptr("disabled"),
			},
		},
	}
	rd := RestData{
		Insights: si.SecurityInsights{
			Repository: si.Repository{
				Security: si.SecurityInfo{
					Tools: []si.Tool{
						{Type: "secret-scanning"},
					},
				},
			},
		},
	}
	sp, err := buildSecurityPosture(repo, rd)
	assert.NoError(t, err)
	assert.True(t, sp.PreventsPushingSecrets())
	assert.True(t, sp.ScansForSecrets())
}

func TestInsightsClaimsSecretsTooling(t *testing.T) {
	insights := si.SecurityInsights{
		Repository: si.Repository{
			Security: si.SecurityInfo{
				Tools: []si.Tool{
					{Type: "secret-scanning"},
					{Type: "other-tool"},
				},
			},
		},
	}
	assert.True(t, insightsClaimsSecretsTooling(insights))

	insights.Repository.Security.Tools = []si.Tool{
		{Type: "other-tool"},
	}
	assert.False(t, insightsClaimsSecretsTooling(insights))

	insights.Repository.Security.Tools = nil
	assert.False(t, insightsClaimsSecretsTooling(insights))
}
