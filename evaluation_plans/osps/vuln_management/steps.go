package vuln_management

import (
	"slices"

	"github.com/ossf/gemara/layer4"

	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
)

func HasSecContact(payloadData any) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	// TODO: Check for a contact email in SECURITY.md

	if data.Insights.Project.Vulnerability.Contact.Email != "" {
		return layer4.Passed, "Security contacts were specified in Security Insights data"
	}
	for _, champion := range data.Insights.Repository.Security.Champions {
		if champion.Email != "" {
			return layer4.Passed, "Security contacts were specified in Security Insights data"
		}
	}

	return layer4.Failed, "Security contacts were not specified in Security Insights data"
}

func SastToolDefined(payloadData interface{}) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	for _, tool := range data.Insights.Repository.Security.Tools {
		if tool.Type == "SAST" {

			enabled := []bool{tool.Integration.Adhoc, tool.Integration.CI, tool.Integration.Release}

			if slices.Contains(enabled, true) {
				return layer4.Passed, "Static Application Security Testing documented in Security Insights"
			}
		}
	}

	return layer4.Failed, "No Static Application Security Testing documented in Security Insights"
}

func HasVulnerabilityDisclosurePolicy(payloadData any) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	if data.Insights.Project.Vulnerability.SecurityPolicy == "" {
		return layer4.Failed, "Vulnerability disclosure policy was NOT specified in Security Insights data"
	}

	return layer4.Passed, "Vulnerability disclosure policy was specified in Security Insights data"
}

func HasPrivateVulnerabilityReporting(payloadData any) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	if !data.Insights.Project.Vulnerability.ReportsAccepted {
		return layer4.Failed, "Project does not accept vulnerability reports according to Security Insights data"
	}

	if data.Insights.Project.Vulnerability.Contact.Email != "" {
		return layer4.Passed, "Private vulnerability reporting available via dedicated contact email in Security Insights data"
	}

	for _, champion := range data.Insights.Repository.Security.Champions {
		if champion.Email != "" {
			return layer4.Passed, "Private vulnerability reporting available via security champions contact in Security Insights data"
		}
	}

	return layer4.Failed, "No private vulnerability reporting contact method found in Security Insights data"
}
