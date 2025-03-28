package docs

import (
	"github.com/revanite-io/sci/pkg/layer4"

	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
)

func hasSupportDocs(payloadData interface{}, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	if data.HasSupportMarkdown() {
		return layer4.Passed, "A support.md file or support statements in the readme.md was found"

	}

	return layer4.Failed, "A support.md file or support statements in the readme.md was NOT found"
}

func hasUserGuides(payloadData interface{}, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	if data.Insights.Project.Documentation.DetailedGuide == "" {
		return layer4.Failed, "User guide was NOT specified in Security Insights data"
	}

	return layer4.Passed, "User guide was specified in Security Insights data"
}

func acceptsVulnReports(payloadData interface{}, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	if data.Insights.Project.Vulnerability.ReportsAccepted {
		return layer4.Passed, "Repository accepts vulnerability reports"
	}

	return layer4.Failed, "Repository does not accept vulnerability reports"
}

func hasSignatureVerificationGuide(payloadData interface{}, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	if data.Insights.Project.Documentation.SignatureVerification == "" {
		return layer4.Failed, "Signature verification guide was NOT specified in Security Insights data"
	}

	return layer4.Passed, "Signature verification guide was specified in Security Insights data"
}

func hasDependencyManagementPolicy(payloadData interface{}, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	if data.Insights.Repository.Documentation.DependencyManagement == "" {
		return layer4.Failed, "Dependency management policy was NOT specified in Security Insights data"
	}

	return layer4.Passed, "Dependency management policy was specified in Security Insights data"
}
