package data

import (
	"github.com/google/go-github/v74/github"
	"github.com/ossf/si-tooling/v2/si"
)

// SecurityPosture defines an interface for accessing security-related metadata about a repository.
type SecurityPosture interface {
	PreventsPushingSecrets() bool
	ScansForSecrets() bool
	DefinesPolicyForHandlingSecrets() bool
}

type RepoSecurityPosture struct {
	restData                        RestData
	preventsSecretPushing           bool
	scansForSecrets                 bool
	definesPolicyForHandlingSecrets bool
}

func buildSecurityPosture(repository *github.Repository, rd RestData) (SecurityPosture, error) {
	securityConfig := repository.GetSecurityAndAnalysis()
	if securityConfig == nil {
		return &RepoSecurityPosture{
			restData: rd,
		}, nil
	}
  secretsScanningStatus := securityConfig.GetSecretScanning().GetStatus()
  insightsClaimsSecretsTooling := insightsClaimsSecretsTooling(rd.Insights)
	return &RepoSecurityPosture{
		restData:              rd,
		preventsSecretPushing: secretsScanningStatus == "enabled" || insightsClaimsSecretsTooling,
		scansForSecrets:       secretsScanningStatus == "enabled" || insightsClaimsSecretsTooling,
		// TODO: consider if SecurityInsights should have a policy doc field in ProjectDocumentation to handle this
		// definesPolicyForHandlingSecrets: rd.SecurityInsights != nil && ....
	}, nil
}

func insightsClaimsSecretsTooling(insights si.SecurityInsights) bool {
	if insights.Repository.Security.Tools == nil {
		return false
	}
	for _, tool := range insights.Repository.Security.Tools {
		if tool.Type == "secret-scanning" {
			return true
		}
	}
	return false
}

func (rsp *RepoSecurityPosture) PreventsPushingSecrets() bool {
	return rsp.preventsSecretPushing
}

func (rsp *RepoSecurityPosture) ScansForSecrets() bool {
	return rsp.scansForSecrets
}

func (rsp *RepoSecurityPosture) DefinesPolicyForHandlingSecrets() bool {
	return rsp.definesPolicyForHandlingSecrets
}
