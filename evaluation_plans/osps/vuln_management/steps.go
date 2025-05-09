package vuln_management

import (
	"slices"

	"github.com/revanite-io/sci/pkg/layer4"

	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
)

func hasSecContact(payloadData interface{}, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	// TODO: Check for a contact email in SECURITY.md
	proj := data.Insights.Project
	if proj != nil && proj.VulnerabilityReporting.Contact != nil && data.Insights.Project.VulnerabilityReporting.Contact.Email.String() != "" {
		return layer4.Passed, "Security contacts were specified in Security Insights data"
	}
	for _, champion := range data.Insights.Repository.SecurityPosture.Champions {
		if champion.Email != nil && len(champion.Email.String()) > 0 {
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

	for _, tool := range data.Insights.Repository.SecurityPosture.Tools {
		if tool.Type == "SAST" {

			enabled := []bool{tool.Integration.Adhoc, tool.Integration.Ci, tool.Integration.Release}

			if slices.Contains(enabled, true) {
				return layer4.Passed, "Static Application Security Testing documented in Security Insights"
			}
		}
	}

	return layer4.Failed, "No Static Application Security Testing documented in Security Insights"
}
