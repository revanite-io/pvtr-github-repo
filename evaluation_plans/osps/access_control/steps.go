package access_control

import (
	"github.com/gemaraproj/go-gemara"

	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
)

func OrgRequiresMFA(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	payload, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}

	required := payload.RepositoryMetadata.IsMFARequiredForAdministrativeActions()

	if required == nil {
		return gemara.NotRun, "Not evaluated. Two-factor authentication evaluation requires a token with org:admin permissions, or manual review", confidence
	} else if *required {
		return gemara.Passed, "Two-factor authentication is configured as required by the parent organization", confidence
	}
	return gemara.Failed, "Two-factor authentication is NOT configured as required by the parent organization", confidence
}

func BranchProtectionRestrictsPushes(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	payload, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}
	protectionData := payload.Repository.DefaultBranchRef.BranchProtectionRule

	if protectionData.RestrictsPushes {
		result = gemara.Passed
		message = "Branch protection rule restricts pushes"
	} else if protectionData.RequiresApprovingReviews {
		result = gemara.Passed
		message = "Branch protection rule requires approving reviews"
	} else {
		result = gemara.NeedsReview
		message = "Branch protection rule does not restrict pushes or require approving reviews; Rulesets not yet evaluated."
	}
	return
}

func BranchProtectionPreventsDeletion(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	payload, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}

	allowsDeletion := payload.Repository.DefaultBranchRef.RefUpdateRule.AllowsDeletions

	if allowsDeletion {
		result = gemara.Failed
		message = "Branch protection rule allows deletions"
	} else {
		result = gemara.Passed
		message = "Branch protection rule prevents deletions"
	}
	return
}

func WorkflowDefaultReadPermissions(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	payload, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}

	permissions := payload.WorkflowPermissions
	if !payload.WorkflowsEnabled {
		return gemara.NeedsReview, "GitHub Actions is disabled for this repository; manual review required.", confidence
	}

	if permissions.DefaultPermissions == "read" && !permissions.CanApprovePullRequest {
		result = gemara.Passed
		message = "Workflow permissions default to read only."
	} else if permissions.DefaultPermissions == "read" && permissions.CanApprovePullRequest {
		result = gemara.Failed
		message = "Workflow permissions default to read only for contents and packages, but PR approval is permitted."
	} else if permissions.DefaultPermissions == "write" && !permissions.CanApprovePullRequest {
		result = gemara.Failed
		message = "Workflow permissions default to read/write, but PR approval is forbidden."
	} else {
		result = gemara.Failed
		message = "Workflow permissions default to read/write and PR approval is permitted."
	}
	return
}
