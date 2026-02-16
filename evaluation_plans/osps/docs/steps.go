package docs

import (
	"github.com/gemaraproj/go-gemara"

	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
)

func HasSupportDocs(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}

	if data.HasSupportMarkdown() {
		return gemara.Passed, "A support.md file or support statements in the readme.md was found", confidence

	}

	return gemara.Failed, "A support.md file or support statements in the readme.md was NOT found", confidence
}

func HasUserGuides(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}

	if data.Insights.Project.Documentation.DetailedGuide == nil {
		return gemara.Failed, "User guide was NOT specified in Security Insights data", confidence
	}

	return gemara.Passed, "User guide was specified in Security Insights data", confidence
}

func AcceptsVulnReports(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}

	if data.Insights.Project.VulnerabilityReporting.ReportsAccepted {
		return gemara.Passed, "Repository accepts vulnerability reports", confidence
	}

	return gemara.Failed, "Repository does not accept vulnerability reports", confidence
}

func HasSignatureVerificationGuide(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}

	if data.Insights.Project.Documentation.SignatureVerification == nil {
		return gemara.Failed, "Signature verification guide was NOT specified in Security Insights data", confidence
	}

	return gemara.Passed, "Signature verification guide was specified in Security Insights data", confidence
}

func HasDependencyManagementPolicy(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}

	if data.Insights.Repository.Documentation.DependencyManagementPolicy == nil {
		return gemara.Failed, "Dependency management policy was NOT specified in Security Insights data", confidence
	}

	return gemara.Passed, "Dependency management policy was specified in Security Insights data", confidence
}

func HasIdentityVerificationGuide(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}

	if data.Insights.Project.Documentation.SignatureVerification == nil {
		return gemara.Failed, "Identity verification guide was NOT specified in Security Insights data (checked signature-verification field)", confidence
	}

	return gemara.Passed, "Identity verification guide was specified in Security Insights data (found in signature-verification field)", confidence
}
