package reusable_steps

import (
	"fmt"

	"github.com/ossf/gemara/layer4"

	"github.com/revanite-io/pvtr-github-repo/data"
)

func VerifyPayload(payloadData any) (payload data.Payload, message string) {
	payload, ok := payloadData.(data.Payload)
	if !ok {
		message = fmt.Sprintf("Malformed assessment: expected payload type %T, got %T (%v)", data.Payload{}, payloadData, payloadData)
	}
	return
}

func NotImplemented(payloadData any, changes map[string]*layer4.Change) (result layer4.Result, message string) {
	return layer4.NeedsReview, "Not implemented"
}

func GithubBuiltIn(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	_, message = VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	return layer4.Passed, "This control is enforced by GitHub for all projects"
}

func GithubTermsOfService(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	return layer4.Passed, "This control is satisfied by the GitHub Terms of Service"
}

func HasSecurityInsightsFile(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	payload, message := VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	if payload.Insights.Header.URL == "" {
		return layer4.NeedsReview, "Security insights required for this assessment, but file not found"
	}

	return layer4.Passed, "Security insights file found"
}

func HasMadeReleases(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	payload, message := VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	if len(payload.Releases) == 0 {
		return layer4.NotApplicable, "No releases found"
	}

	return layer4.Passed, fmt.Sprintf("Found %v releases", len(payload.Releases))
}

func IsActive(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	payload, message := VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	if payload.Insights.Repository.Status == "active" {
		result = layer4.Passed
	} else {
		result = layer4.NotApplicable
	}

	return result, fmt.Sprintf("Repo Status is %s", payload.Insights.Repository.Status)
}

func HasIssuesOrDiscussionsEnabled(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	if data.Repository.HasDiscussionsEnabled && data.Repository.HasIssuesEnabled {
		return layer4.Passed, "Both issues and discussions are enabled for the repository"
	}
	if data.Repository.HasDiscussionsEnabled {
		return layer4.Passed, "Discussions are enabled for the repository"
	}
	if data.Repository.HasIssuesEnabled {
		return layer4.Passed, "Issues are enabled for the repository"
	}
	return layer4.Failed, "Both issues and discussions are disabled for the repository"
}

func HasDependencyManagementPolicy(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	payload, message := VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	if len(payload.Insights.Repository.Documentation.DependencyManagement) > 0 {
		return layer4.Passed, "Found dependency management policy in documentation"
	}

	return layer4.Failed, "No dependency management file found"
}

func IsCodeRepo(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	payload, message := VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	if !payload.IsCodeRepo {
		return layer4.NotApplicable, "Repository does not contain code"
	}

	return layer4.Passed, "Repository contains code"
}
