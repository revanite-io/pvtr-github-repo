package docs

import (
	"github.com/gemaraproj/go-gemara"

	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
)

func HasSupportDocs(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message
	}

	if data.HasSupportMarkdown() {
		return gemara.Passed, "A support.md file or support statements in the readme.md was found"

	}

	return gemara.Failed, "A support.md file or support statements in the readme.md was NOT found"
}

func HasUserGuides(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message
	}

	if data.Insights.Project.Documentation.DetailedGuide == "" {
		return gemara.Failed, "User guide was NOT specified in Security Insights data"
	}

	return gemara.Passed, "User guide was specified in Security Insights data"
}

func AcceptsVulnReports(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message
	}

	if data.Insights.Project.Vulnerability.ReportsAccepted {
		return gemara.Passed, "Repository accepts vulnerability reports"
	}

	return gemara.Failed, "Repository does not accept vulnerability reports"
}

func HasSignatureVerificationGuide(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message
	}

	if data.Insights.Project.Documentation.SignatureVerification == "" {
		return gemara.Failed, "Signature verification guide was NOT specified in Security Insights data"
	}

	return gemara.Passed, "Signature verification guide was specified in Security Insights data"
}

func HasDependencyManagementPolicy(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message
	}

	if data.Insights.Repository.Documentation.DependencyManagement == "" {
		return gemara.Failed, "Dependency management policy was NOT specified in Security Insights data"
	}

	return gemara.Passed, "Dependency management policy was specified in Security Insights data"
}

func HasIdentityVerificationGuide(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message
	}

	if data.Insights.Project.Documentation.SignatureVerification == "" {
		return gemara.Failed, "Identity verification guide was NOT specified in Security Insights data (checked signature-verification field)"
	}

	return gemara.Passed, "Identity verification guide was specified in Security Insights data (found in signature-verification field)"
}
