package access_control

import (
	"github.com/ossf/gemara/layer4"

	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
)

func orgRequiresMFA(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	payload, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	required := payload.RepositoryMetadata.IsMFARequiredForAdministrativeActions()

	if required == nil {
		return layer4.NeedsReview, "Not evaluated. Two-factor authentication evaluation requires a token with org:admin permissions, or manual review"
	} else if *required {
		return layer4.Passed, "Two-factor authentication is configured as required by the parent organization"
	}
	return layer4.Failed, "Two-factor authentication is NOT configured as required by the parent organization"
}

func branchProtectionRestrictsPushes(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	payload, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}
	protectionData := payload.Repository.DefaultBranchRef.BranchProtectionRule

	if protectionData.RestrictsPushes {
		result = layer4.Passed
		message = "Branch protection rule restricts pushes"
	} else if protectionData.RequiresApprovingReviews {
		result = layer4.Passed
		message = "Branch protection rule requires approving reviews"
	} else {
		result = layer4.NeedsReview
		message = "Branch protection rule does not restrict pushes or require approving reviews; Rulesets not yet evaluated."
	}
	return
}

func branchProtectionPreventsDeletion(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	payload, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	allowsDeletion := payload.Repository.DefaultBranchRef.RefUpdateRule.AllowsDeletions

	if allowsDeletion {
		result = layer4.Failed
		message = "Branch protection rule allows deletions"
	} else {
		result = layer4.Passed
		message = "Branch protection rule prevents deletions"
	}
	return
}

func workflowDefaultReadPermissions(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	payload, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	permissions := payload.WorkflowPermissions
	if !payload.WorkflowsEnabled {
		return layer4.NeedsReview, "GitHub Actions is disabled for this repository; manual review required."
	}

	if permissions.DefaultPermissions == "read" && !permissions.CanApprovePullRequest {
		result = layer4.Passed
		message = "Workflow permissions default to read only."
	} else if permissions.DefaultPermissions == "read" && permissions.CanApprovePullRequest {
		result = layer4.Failed
		message = "Workflow permissions default to read only for contents and packages, but PR approval is permitted."
	} else if permissions.DefaultPermissions == "write" && !permissions.CanApprovePullRequest {
		result = layer4.Failed
		message = "Workflow permissions default to read/write, but PR approval is forbidden."
	} else {
		result = layer4.Failed
		message = "Workflow permissions default to read/write and PR approval is permitted."
	}
	return
}
