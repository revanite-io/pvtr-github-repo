package reusable_steps

import (
	"fmt"

	"github.com/gemaraproj/go-gemara"

	"github.com/revanite-io/pvtr-github-repo/data"
)

func VerifyPayload(payloadData any) (payload data.Payload, message string) {
	payload, ok := payloadData.(data.Payload)
	if !ok {
		message = fmt.Sprintf("Malformed assessment: expected payload type %T, got %T (%v)", data.Payload{}, payloadData, payloadData)
	}
	return
}

func NotImplemented(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	return gemara.NotRun, "Not implemented"
}

func GithubBuiltIn(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	_, message = VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message
	}

	return gemara.Passed, "This control is enforced by GitHub for all projects"
}

func GithubTermsOfService(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	return gemara.Passed, "This control is satisfied by the GitHub Terms of Service"
}

func HasSecurityInsightsFile(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	payload, message := VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message
	}
	if payload.InsightsError {
		return gemara.NeedsReview, "An error was encountered while parsing Security Insights content"
	}
	if payload.Insights.Header.URL == "" {
		return gemara.NeedsReview, "Security insights required for this assessment, but file not found"
	}

	return gemara.Passed, "Security insights file found"
}

func HasMadeReleases(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	payload, message := VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message
	}

	if len(payload.Releases) == 0 {
		return gemara.NotApplicable, "No releases found"
	}

	return gemara.Passed, fmt.Sprintf("Found %v releases", len(payload.Releases))
}

func IsActive(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	payload, message := VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message
	}

	if payload.Insights.Repository.Status == "active" {
		result = gemara.Passed
	} else {
		result = gemara.NotApplicable
	}

	return result, fmt.Sprintf("Repo Status is %s", payload.Insights.Repository.Status)
}

func HasIssuesOrDiscussionsEnabled(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message
	}

	if data.Repository.HasDiscussionsEnabled && data.Repository.HasIssuesEnabled {
		return gemara.Passed, "Both issues and discussions are enabled for the repository"
	}
	if data.Repository.HasDiscussionsEnabled {
		return gemara.Passed, "Discussions are enabled for the repository"
	}
	if data.Repository.HasIssuesEnabled {
		return gemara.Passed, "Issues are enabled for the repository"
	}
	return gemara.Failed, "Both issues and discussions are disabled for the repository"
}

func HasDependencyManagementPolicy(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	payload, message := VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message
	}

	if len(payload.Insights.Repository.Documentation.DependencyManagement) > 0 {
		return gemara.Passed, "Found dependency management policy in documentation"
	}

	return gemara.Failed, "No dependency management file found"
}

func IsCodeRepo(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	payload, message := VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message
	}

	if !payload.IsCodeRepo {
		return gemara.NotApplicable, "Repository does not contain code"
	}

	return gemara.Passed, "Repository contains code"
}
