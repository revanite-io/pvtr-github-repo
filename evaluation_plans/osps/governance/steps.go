package governance

import (
	"github.com/gemaraproj/go-gemara"
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
)

func CoreTeamIsListed(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}

	if len(data.Insights.Repository.CoreTeam) == 0 {
		return gemara.Failed, "Core team was NOT specified in Security Insights data", confidence
	}

	return gemara.Passed, "Core team was specified in Security Insights data", confidence
}

func ProjectAdminsListed(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}

	if len(data.Insights.Project.Administrators) == 0 {
		return gemara.Failed, "Project admins were NOT specified in Security Insights data", confidence
	}

	return gemara.Passed, "Project admins were specified in Security Insights data", confidence
}

func HasRolesAndResponsibilities(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}

	if data.Insights.Repository.Documentation.Governance == nil {
		return gemara.Failed, "Roles and responsibilities were NOT specified in Security Insights data", confidence
	}

	return gemara.Passed, "Roles and responsibilities were specified in Security Insights data", confidence
}

func HasContributionGuide(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}

	if data.Insights.Project.Documentation.CodeOfConduct != nil && data.Insights.Repository.Documentation.ContributingGuide != nil {
		return gemara.Passed, "Contributing guide specified in Security Insights data (Bonus: code of conduct location also specified)", confidence
	}

	if data.Repository.ContributingGuidelines.Body != "" && data.Insights.Project.Documentation.CodeOfConduct != nil {
		return gemara.Passed, "Contributing guide was found via GitHub API (Bonus: code of conduct was specified in Security Insights data)", confidence
	}

	if data.Repository.ContributingGuidelines.Body != "" {
		return gemara.NeedsReview, "Contributing guide was found via GitHub API (Recommendation: Add code of conduct location to Security Insights data)", confidence
	}

	return gemara.Failed, "Contribution guide not found in Security Insights data or via GitHub API", confidence
}

func HasContributionReviewPolicy(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}
	if !data.IsCodeRepo {
		return gemara.NotApplicable, "Repository contains no code - skipping code contribution policy check", confidence
	}
	if data.Insights.Repository.Documentation.ReviewPolicy != nil {
		return gemara.Passed, "Code review guide was specified in Security Insights data", confidence
	}

	return gemara.Failed, "Code review guide was NOT specified in Security Insights data", confidence
}
