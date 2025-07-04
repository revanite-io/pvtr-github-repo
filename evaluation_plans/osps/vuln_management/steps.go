package vuln_management

import (
	"slices"

	"github.com/ossf/gemara/layer4"

	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
)

func hasSecContact(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
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

func sastToolDefined(payloadData interface{}, _ map[string]*layer4.Change) (result layer4.Result, message string) {
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
