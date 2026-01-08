package governance

import (
	"github.com/gemaraproj/go-gemara"
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
)

func CoreTeamIsListed(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, 0 // TODO
	}

	if len(data.Insights.Repository.CoreTeam) == 0 {
		return gemara.Failed, "Core team was NOT specified in Security Insights data", 0 // TODO
	}

	return gemara.Passed, "Core team was specified in Security Insights data", 0 // TODO
}

func ProjectAdminsListed(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, 0 // TODO
	}

	if len(data.Insights.Project.Administrators) == 0 {
		return gemara.Failed, "Project admins were NOT specified in Security Insights data", 0 // TODO
	}

	return gemara.Passed, "Project admins were specified in Security Insights data", 0 // TODO
}

func HasRolesAndResponsibilities(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, 0 // TODO
	}

	if data.Insights.Repository.Documentation.Governance == "" {
		return gemara.Failed, "Roles and responsibilities were NOT specified in Security Insights data", 0 // TODO
	}

	return gemara.Passed, "Roles and responsibilities were specified in Security Insights data", 0 // TODO
}

func HasContributionGuide(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, 0 // TODO
	}

	if data.Insights.Project.Documentation.CodeOfConduct != "" && data.Insights.Repository.Documentation.Contributing != "" {
		return gemara.Passed,
			"Contributing guide specified in Security Insights data (Bonus: code of conduct location also specified)",
			0 // TODO
	}

	if data.Repository.ContributingGuidelines.Body != "" && data.Insights.Project.Documentation.CodeOfConduct != "" {
		return gemara.Passed,
			"Contributing guide was found via GitHub API (Bonus: code of conduct was specified in Security Insights data)",
			0 // TODO
	}

	if data.Repository.ContributingGuidelines.Body != "" {
		return gemara.NeedsReview,
			"Contributing guide was found via GitHub API (Recommendation: Add code of conduct location to Security Insights data)",
			0 // TODO
	}

	return gemara.Failed,
		"Contribution guide not found in Security Insights data or via GitHub API",
		0 // TODO
}

func HasContributionReviewPolicy(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, 0 // TODO
	}
	if !data.IsCodeRepo {
		return gemara.NotApplicable,
			"Repository contains no code - skipping code contribution policy check",
			0 // TODO
	}
	if data.Insights.Repository.Documentation.ReviewPolicy != "" {
		return gemara.Passed,
			"Code review guide was specified in Security Insights data",
			0 // TODO
	}

	return gemara.Failed,
		"Code review guide was NOT specified in Security Insights data",
		0 // TODO
}
