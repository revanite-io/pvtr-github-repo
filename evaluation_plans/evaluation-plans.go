package evaluation_plans

import (
	"github.com/ossf/gemara/layer4"
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/osps/access_control"
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/osps/build_release"
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/osps/docs"
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/osps/governance"
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/osps/legal"
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/osps/quality"
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/osps/vuln_management"

	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
)

var (
	// Open Source Project Security Baseline
	OSPS = map[string][]layer4.AssessmentStep{
		"OSPS-AC-01.01": {
			access_control.OrgRequiresMFA,
		},
		"OSPS-AC-02.01": {
			reusable_steps.GithubBuiltIn,
		},
		"OSPS-AC-03.01": {
			access_control.BranchProtectionRestrictsPushes,
		},
		"OSPS-AC-03.02": {
			access_control.BranchProtectionPreventsDeletion,
		},
		"OSPS-AC-04.01": {
			access_control.WorkflowDefaultReadPermissions,
		},
		"OSPS-AC-04.02": {
			reusable_steps.NotImplemented,
		},
		"OSPS-BR-01.01": {
			build_release.CicdSanitizedInputParameters,
		},
		"OSPS-BR-01.02": {
			reusable_steps.NotImplemented,
		},
		"OSPS-BR-02.01": {
			reusable_steps.HasMadeReleases,
			build_release.ReleaseHasUniqueIdentifier,
		},
		"OSPS-BR-02.02": {
			reusable_steps.NotImplemented,
		},
		"OSPS-BR-03.01": {
			reusable_steps.HasSecurityInsightsFile,
			build_release.EnsureInsightsLinksUseHTTPS,
		},
		"OSPS-BR-03.02": {
			build_release.DistributionPointsUseHTTPS,
		},
		"OSPS-BR-04.01": {
			reusable_steps.HasMadeReleases,
			build_release.EnsureLatestReleaseHasChangelog,
		},
		"OSPS-BR-05.01": {
			reusable_steps.NotImplemented,
		},
		"OSPS-BR-06.01": {
			reusable_steps.HasMadeReleases,
			reusable_steps.HasSecurityInsightsFile,
			build_release.InsightsHasSlsaAttestation,
		},
		"OSPS-BR-07.01": {
			build_release.SecretScanningInUse,
		},
		"OSPS-BR-07.02": {
			reusable_steps.NotImplemented,
		},
		"OSPS-DO-01.01": {
			reusable_steps.HasMadeReleases,
			reusable_steps.HasSecurityInsightsFile,
			docs.HasUserGuides,
		},
		"OSPS-DO-02.01": {
			reusable_steps.HasMadeReleases,
			reusable_steps.HasIssuesOrDiscussionsEnabled,
			docs.AcceptsVulnReports,
		},
		"OSPS-DO-03.01": {
			reusable_steps.HasMadeReleases,
			reusable_steps.HasSecurityInsightsFile,
			docs.HasSignatureVerificationGuide,
		},
		"OSPS-DO-03.02": {
			reusable_steps.HasMadeReleases,
			reusable_steps.HasSecurityInsightsFile,
			docs.HasIdentityVerificationGuide,
		},
		"OSPS-DO-04.01": {
			docs.HasSupportDocs,
		},
		"OSPS-DO-05.01": {
			reusable_steps.NotImplemented,
		},
		"OSPS-DO-06.01": {
			reusable_steps.IsCodeRepo,
			reusable_steps.HasMadeReleases,
			reusable_steps.HasSecurityInsightsFile,
			docs.HasDependencyManagementPolicy,
		},
		"OSPS-GV-01.01": {
			reusable_steps.HasSecurityInsightsFile,
			reusable_steps.IsActive,
			governance.CoreTeamIsListed,
			governance.ProjectAdminsListed,
		},
		"OSPS-GV-01.02": {
			governance.HasRolesAndResponsibilities,
		},
		"OSPS-GV-02.01": {
			reusable_steps.HasIssuesOrDiscussionsEnabled,
		},
		"OSPS-GV-03.01": {
			governance.HasContributionGuide,
		},
		"OSPS-GV-03.02": {
			reusable_steps.IsCodeRepo,
			reusable_steps.HasSecurityInsightsFile,
			reusable_steps.IsActive,
			governance.HasContributionReviewPolicy,
		},
		"OSPS-GV-04.01": {
			reusable_steps.NotImplemented,
		},
		"OSPS-LE-01.01": {
			reusable_steps.GithubTermsOfService,
		},
		"OSPS-LE-02.01": {
			legal.FoundLicense,
			legal.GoodLicense,
		},
		"OSPS-LE-02.02": {
			legal.ReleasesLicensed,
			legal.GoodLicense,
		},
		"OSPS-LE-03.01": {
			legal.FoundLicense,
		},
		"OSPS-LE-03.02": {
			legal.ReleasesLicensed,
		},
		"OSPS-QA-01.01": {
			quality.RepoIsPublic,
		},
		"OSPS-QA-01.02": {
			reusable_steps.GithubBuiltIn,
		},
		"OSPS-QA-02.01": {
			quality.VerifyDependencyManagement,
		},
		"OSPS-QA-02.02": {
			reusable_steps.NotImplemented,
		},
		"OSPS-QA-03.01": {
			quality.StatusChecksAreRequiredByRulesets,
			quality.StatusChecksAreRequiredByBranchProtection,
		},
		"OSPS-QA-04.01": {
			reusable_steps.IsCodeRepo,
			reusable_steps.HasSecurityInsightsFile,
			reusable_steps.IsActive,
			quality.InsightsListsRepositories,
		},
		"OSPS-QA-04.02": {
			reusable_steps.NotImplemented,
		},
		"OSPS-QA-05.01": {
			quality.NoBinariesInRepo,
		},
		"OSPS-QA-05.02": {
			reusable_steps.NotImplemented,
		},
		"OSPS-QA-06.01": {
			reusable_steps.IsCodeRepo,
			quality.HasOneOrMoreStatusChecks,
		},
		"OSPS-QA-06.02": {
			quality.DocumentsTestExecution,
		},
		"OSPS-QA-06.03": {
			reusable_steps.IsCodeRepo,
			quality.DocumentsTestMaintenancePolicy,
		},
		"OSPS-QA-07.01": {
			quality.RequiresNonAuthorApproval,
		},
		"OSPS-SA-01.01": {
			reusable_steps.NotImplemented,
		},
		"OSPS-SA-02.01": {
			reusable_steps.NotImplemented,
		},
		"OSPS-SA-03.01": {
			reusable_steps.NotImplemented,
		},
		"OSPS-SA-03.02": {
			reusable_steps.NotImplemented,
		},
		"OSPS-VM-01.01": {
			reusable_steps.IsActive,
			reusable_steps.HasSecurityInsightsFile,
			vuln_management.HasVulnerabilityDisclosurePolicy,
		},
		"OSPS-VM-02.01": {
			reusable_steps.IsCodeRepo,
			vuln_management.HasSecContact,
		},
		"OSPS-VM-03.01": {
			reusable_steps.IsActive,
			reusable_steps.HasSecurityInsightsFile,
			vuln_management.HasPrivateVulnerabilityReporting,
		},
		"OSPS-VM-04.01": {
			reusable_steps.NotImplemented,
		},
		"OSPS-VM-04.02": {
			reusable_steps.NotImplemented,
		},
		"OSPS-VM-05.01": {
			reusable_steps.NotImplemented,
		},
		"OSPS-VM-05.03": {
			reusable_steps.NotImplemented,
		},
		"OSPS-VM-05.02": {
			reusable_steps.NotImplemented,
		},
		"OSPS-VM-06.01": {
			reusable_steps.HasDependencyManagementPolicy,
		},
		"OSPS-VM-06.02": {
			reusable_steps.IsCodeRepo,
			reusable_steps.HasSecurityInsightsFile,
			vuln_management.SastToolDefined,
		},
	}
)
