package legal

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/revanite-io/pvtr-github-repo/data"
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
	"github.com/revanite-io/sci/pkg/layer4"
)

type LicenseList struct {
	Licenses []License `json:"licenses"`
}

type License struct {
	LicenseID             string `json:"licenseId"`
	IsDeprecatedLicenseId bool   `json:"isDeprecatedLicenseId"`
	IsOsiApproved         bool   `json:"isOsiApproved"`
	IsFsfLibre            bool   `json:"isFsfLibre"`
}

const spdxURL = "https://raw.githubusercontent.com/spdx/license-list-data/main/json/licenses.json"

func getLicenseList(data data.Payload) (LicenseList, string) {
	goodLicenseList := LicenseList{}
	response, err := data.MakeApiCall(spdxURL, false)
	if err != nil {
		return goodLicenseList, fmt.Sprintf("Failed to fetch good license data: %s", err.Error())
	}
	err = json.Unmarshal(response, &goodLicenseList)
	if err != nil {
		return goodLicenseList, fmt.Sprintf("Failed to unmarshal good license data: %s", err.Error())
	}
	if goodLicenseList.Licenses == nil || len(goodLicenseList.Licenses) == 0 {
		return goodLicenseList, "Good license data was unexpectedly empty"
	}
	return goodLicenseList, ""
}

func splitSpdxExpression(expression string) (spdx_ids []string) {
	a := strings.Split(expression, " AND ")
	for _, aa := range a {
		b := strings.Split(aa, " OR ")
		for _, bb := range b {
			spdx_ids = append(spdx_ids, bb)
		}
	}
	return
}

func foundLicense(payloadData interface{}, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}
	if data.Repository.LicenseInfo.Url == "" {
		return layer4.Failed, "License was not found in a well known location via the GitHub API"
	}
	return layer4.Passed, "License was found in a well known location via the GitHub API"
}

func goodLicense(payloadData interface{}, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	licenses, errString := getLicenseList(data)
	if errString != "" {
		return layer4.Unknown, errString
	}

	apiInfo := data.Repository.LicenseInfo.SpdxId
	siInfo := data.Insights.Repository.License.Expression
	if apiInfo == "" && siInfo == "" {
		return layer4.Failed, "License SPDX identifier was not found in Security Insights data or via GitHub API"
	}

	spdx_ids_a := splitSpdxExpression(apiInfo)
	spdx_ids_b := splitSpdxExpression(siInfo)
	spdx_ids := append(spdx_ids_a, spdx_ids_b...)
	badLicenses := []string{}
	for _, spdx_id := range spdx_ids {
		var validId bool
		for _, license := range licenses.Licenses {
			if license.LicenseID == spdx_id {
				validId = true
				if (!license.IsOsiApproved && !license.IsFsfLibre) || license.IsDeprecatedLicenseId {
					badLicenses = append(badLicenses, spdx_id)
				}
			}
		}
		if !validId {
			badLicenses = append(badLicenses, spdx_id)
		}
	}
	approvedLicenses := strings.Join(spdx_ids, ", ")
	data.Config.Logger.Trace(fmt.Sprintf("Approved licenses: %s", approvedLicenses))
	data.Config.Logger.Trace(fmt.Sprintf("Non-approved licenses: %s", badLicenses))

	if len(badLicenses) > 0 {
		return layer4.Failed, fmt.Sprintf("These licenses are not OSI or FSF approved: %s", strings.Join(badLicenses, ", "))
	}
	return layer4.NeedsReview, "All license found are OSI or FSF approved"
}
