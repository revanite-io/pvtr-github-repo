package governance

import (
	"github.com/ossf/gemara/layer4"
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
)

func coreTeamIsListed(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	if len(data.Insights.Repository.CoreTeam) == 0 {
		return layer4.Failed, "Core team was NOT specified in Security Insights data"
	}

	return layer4.Passed, "Core team was specified in Security Insights data"
}

func projectAdminsListed(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	if len(data.Insights.Project.Administrators) == 0 {
		return layer4.Failed, "Project admins were NOT specified in Security Insights data"
	}

	return layer4.Passed, "Project admins were specified in Security Insights data"
}

func hasRolesAndResponsibilities(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	if data.Insights.Repository.Documentation.Governance == "" {
		return layer4.Failed, "Roles and responsibilities were NOT specified in Security Insights data"
	}

	return layer4.Passed, "Roles and responsibilities were specified in Security Insights data"
}

func hasContributionGuide(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	if data.Insights.Project.Documentation.CodeOfConduct != "" && data.Insights.Repository.Documentation.Contributing != "" {
		return layer4.Passed, "Contributing guide specified in Security Insights data (Bonus: code of conduct location also specified)"
	}

	if data.Repository.ContributingGuidelines.Body != "" && data.Insights.Project.Documentation.CodeOfConduct != "" {
		return layer4.Passed, "Contributing guide was found via GitHub API (Bonus: code of conduct was specified in Security Insights data)"
	}

	if data.Repository.ContributingGuidelines.Body != "" {
		return layer4.NeedsReview, "Contributing guide was found via GitHub API (Recommendation: Add code of conduct location to Security Insights data)"
	}

	return layer4.Failed, "Contribution guide not found in Security Insights data or via GitHub API"
}

func hasContributionReviewPolicy(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}
	if !data.IsCodeRepo {
		return layer4.NotApplicable, "Repository contains no code - skipping code contribution policy check"
	}
	if data.Insights.Repository.Documentation.ReviewPolicy != "" {
		return layer4.Passed, "Code review guide was specified in Security Insights data"
	}

	return layer4.Failed, "Code review guide was NOT specified in Security Insights data"
}
