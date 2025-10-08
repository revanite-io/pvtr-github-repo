package docs

import (
	"github.com/ossf/gemara/layer4"

	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
)

func HasSupportDocs(payloadData any) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	if data.HasSupportMarkdown() {
		return layer4.Passed, "A support.md file or support statements in the readme.md was found"

	}

	return layer4.Failed, "A support.md file or support statements in the readme.md was NOT found"
}

func HasUserGuides(payloadData any) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	if data.Insights.Project.Documentation.DetailedGuide == "" {
		return layer4.Failed, "User guide was NOT specified in Security Insights data"
	}

	return layer4.Passed, "User guide was specified in Security Insights data"
}

func AcceptsVulnReports(payloadData any) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	if data.Insights.Project.Vulnerability.ReportsAccepted {
		return layer4.Passed, "Repository accepts vulnerability reports"
	}

	return layer4.Failed, "Repository does not accept vulnerability reports"
}

func HasSignatureVerificationGuide(payloadData any) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	if data.Insights.Project.Documentation.SignatureVerification == "" {
		return layer4.Failed, "Signature verification guide was NOT specified in Security Insights data"
	}

	return layer4.Passed, "Signature verification guide was specified in Security Insights data"
}

func HasDependencyManagementPolicy(payloadData any) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	if data.Insights.Repository.Documentation.DependencyManagement == "" {
		return layer4.Failed, "Dependency management policy was NOT specified in Security Insights data"
	}

	return layer4.Passed, "Dependency management policy was specified in Security Insights data"
}

func HasIdentityVerificationGuide(payloadData any) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	if data.Insights.Project.Documentation.SignatureVerification == "" {
		return layer4.Failed, "Identity verification guide was NOT specified in Security Insights data (checked signature-verification field)"
	}

	return layer4.Passed, "Identity verification guide was specified in Security Insights data (found in signature-verification field)"
}
