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

func defaultBranchRestrictsPushes(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
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
		if payload.RepositoryMetadata.IsDefaultBranchProtected() != nil && *payload.RepositoryMetadata.IsDefaultBranchProtected() {
			result = layer4.Passed
			message = "Branch rule restricts pushes"
		} else if payload.RepositoryMetadata.DefaultBranchRequiresPRReviews() != nil && *payload.RepositoryMetadata.DefaultBranchRequiresPRReviews() {
			result = layer4.Passed
			message = "Branch rule requires approving reviews"
		} else {
			result = layer4.Failed
			message = "Default branch is not protected"
		}
	}
	return result, message
}

func defaultBranchPreventsDeletion(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	payload, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	branchProtectionAllowsDeletion := payload.Repository.DefaultBranchRef.RefUpdateRule.AllowsDeletions
	deletionRule := payload.RepositoryMetadata.IsDefaultBranchProtectedFromDeletion()
	branchRulesAllowDeletion := deletionRule == nil || !*deletionRule

	if branchProtectionAllowsDeletion && branchRulesAllowDeletion {
		result = layer4.Failed
		message = "Default branch is not protected from deletions"
	} else {
		result = layer4.Passed
		if *deletionRule {
			message = "Default branch is protected from deletions by rulesets"
		} else {
			message = "Default branch is protected from deletions by branch protection rules"
		}
	}
	return result, message
}

func workflowDefaultReadPermissions(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	payload, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	workflowPermissions := payload.WorkflowPermissions.DefaultPermissions

	message = "Workflow permissions default to " + workflowPermissions

	if workflowPermissions == "read" {
		result = layer4.Passed
	} else {
		result = layer4.Failed
	}
	return
}
