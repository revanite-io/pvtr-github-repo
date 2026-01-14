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
	return gemara.NotRun, "Not implemented", 0 // TODO
}

func GithubBuiltIn(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	_, message = VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, 0 // TODO
	}

	return gemara.Passed, "This control is enforced by GitHub for all projects", 0 // TODO
}

func GithubTermsOfService(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	return gemara.Passed, "This control is satisfied by the GitHub Terms of Service", 0 // TODO
}

func HasSecurityInsightsFile(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	payload, message := VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, 0 // TODO
	}
	if payload.InsightsError {
		return gemara.NeedsReview, "An error was encountered while parsing Security Insights content", 0 // TODO
	}
	if payload.Insights.Header.URL == "" {
		return gemara.NeedsReview, "Security insights required for this assessment, but file not found", 0 // TODO
	}

	return gemara.Passed, "Security insights file found", 0 // TODO
}

func HasMadeReleases(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	payload, message := VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, 0 // TODO
	}

	if len(payload.Releases) == 0 {
		return gemara.NotApplicable, "No releases found", 0 // TODO
	}

	return gemara.Passed, fmt.Sprintf("Found %v releases", len(payload.Releases)), 0 // TODO
}

func IsActive(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	payload, message := VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, 0 // TODO
	}

	if payload.Insights.Repository.Status == "active" {
		result = gemara.Passed
	} else {
		result = gemara.NotApplicable
	}

	return result, fmt.Sprintf("Repo Status is %s", payload.Insights.Repository.Status), 0 // TODO
}

func HasIssuesOrDiscussionsEnabled(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, 0 // TODO
	}

	if data.Repository.HasDiscussionsEnabled && data.Repository.HasIssuesEnabled {
		return gemara.Passed, "Both issues and discussions are enabled for the repository", 0 // TODO
	}
	if data.Repository.HasDiscussionsEnabled {
		return gemara.Passed, "Discussions are enabled for the repository", 0 // TODO
	}
	if data.Repository.HasIssuesEnabled {
		return gemara.Passed, "Issues are enabled for the repository", 0 // TODO
	}
	return gemara.Failed, "Both issues and discussions are disabled for the repository", 0 // TODO
}

func HasDependencyManagementPolicy(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	payload, message := VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, 0 // TODO
	}

	if len(payload.Insights.Repository.Documentation.DependencyManagement) > 0 {
		return gemara.Passed, "Found dependency management policy in documentation", 0 // TODO
	}

	return gemara.Failed, "No dependency management file found", 0 // TODO
}

func IsCodeRepo(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	payload, message := VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, 0 // TODO
	}

	if !payload.IsCodeRepo {
		return gemara.NotApplicable, "Repository does not contain code", 0 // TODO
	}

	return gemara.Passed, "Repository contains code", 0 // TODO
}
