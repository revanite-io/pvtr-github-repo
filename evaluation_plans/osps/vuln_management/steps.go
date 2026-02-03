package vuln_management

import (
	"slices"

	"github.com/gemaraproj/go-gemara"

	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
)

func HasSecContact(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message
	}

	// TODO: Check for a contact email in SECURITY.md

	if data.Insights.Project.Vulnerability.Contact.Email != "" {
		return gemara.Passed, "Security contacts were specified in Security Insights data"
	}
	for _, champion := range data.Insights.Repository.Security.Champions {
		if champion.Email != "" {
			return gemara.Passed, "Security contacts were specified in Security Insights data"
		}
	}

	return gemara.Failed, "Security contacts were not specified in Security Insights data"
}

func SastToolDefined(payloadData interface{}) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message
	}

	for _, tool := range data.Insights.Repository.Security.Tools {
		if tool.Type == "SAST" {

			enabled := []bool{tool.Integration.Adhoc, tool.Integration.CI, tool.Integration.Release}

			if slices.Contains(enabled, true) {
				return gemara.Passed, "Static Application Security Testing documented in Security Insights"
			}
		}
	}

	return gemara.Failed, "No Static Application Security Testing documented in Security Insights"
}

func HasVulnerabilityDisclosurePolicy(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message
	}

	if data.Insights.Project.Vulnerability.SecurityPolicy == "" {
		return gemara.Failed, "Vulnerability disclosure policy was NOT specified in Security Insights data"
	}

	return gemara.Passed, "Vulnerability disclosure policy was specified in Security Insights data"
}

func HasPrivateVulnerabilityReporting(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message
	}

	if !data.Insights.Project.Vulnerability.ReportsAccepted {
		return gemara.Failed, "Project does not accept vulnerability reports according to Security Insights data"
	}

	if data.Insights.Project.Vulnerability.Contact.Email != "" {
		return gemara.Passed, "Private vulnerability reporting available via dedicated contact email in Security Insights data"
	}

	for _, champion := range data.Insights.Repository.Security.Champions {
		if champion.Email != "" {
			return gemara.Passed, "Private vulnerability reporting available via security champions contact in Security Insights data"
		}
	}

	return gemara.Failed, "No private vulnerability reporting contact method found in Security Insights data"
}
