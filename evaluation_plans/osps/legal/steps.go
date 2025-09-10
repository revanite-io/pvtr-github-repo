package legal

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ossf/gemara/layer4"
	"github.com/revanite-io/pvtr-github-repo/data"
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
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

func getLicenseList(data data.Payload, makeApiCall func(string, bool) ([]byte, error)) (LicenseList, string) {
	goodLicenseList := LicenseList{}
	if makeApiCall == nil {
		makeApiCall = data.MakeApiCall
	}
	response, err := makeApiCall(spdxURL, false)
	if err != nil {
		return goodLicenseList, fmt.Sprintf("Failed to fetch good license data: %s", err.Error())
	}
	err = json.Unmarshal(response, &goodLicenseList)
	if err != nil {
		return goodLicenseList, fmt.Sprintf("Failed to unmarshal good license data: %s", err.Error())
	}
	if len(goodLicenseList.Licenses) == 0 {
		return goodLicenseList, "Good license data was unexpectedly empty"
	}
	return goodLicenseList, ""
}

func splitSpdxExpression(expression string) (spdx_ids []string) {
	a := strings.Split(expression, " AND ")
	for _, aa := range a {
		b := strings.Split(aa, " OR ")
		spdx_ids = append(spdx_ids, b...)
	}
	return
}

func foundLicense(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}
	if data.Repository.LicenseInfo.Url == "" {
		return layer4.Failed, "License was not found in a well known location via the GitHub API"
	}
	return layer4.Passed, "License was found in a well known location via the GitHub API"
}

func releasesLicensed(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	if len(data.Releases) == 0 {
		return layer4.NotApplicable, "No releases found"
	}
	if data.Repository.LicenseInfo.Url == "" {
		return layer4.Failed, "License was not found in a well known location via the GitHub API"
	}
	return layer4.Passed, "GitHub releases include the license(s) in the released source code."
}

func goodLicense(payloadData any, _ map[string]*layer4.Change, mockGetLicenseList ...func(data.Payload, func(string, bool) ([]byte, error)) (LicenseList, string)) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	var licenses LicenseList
	var errString string
	if len(mockGetLicenseList) > 0 && mockGetLicenseList[0] != nil {
		licenses, errString = mockGetLicenseList[0](data, nil)
	} else {
		licenses, errString = getLicenseList(data, nil)
	}
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
		if spdx_id == "" {
			continue
		}
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
	if data.Config.Logger != nil {
		data.Config.Logger.Trace(fmt.Sprintf("Requested licenses: %s", approvedLicenses))
		data.Config.Logger.Trace(fmt.Sprintf("Non-approved licenses: %s", badLicenses))
	}

	if len(badLicenses) > 0 {
		return layer4.Failed, fmt.Sprintf("These licenses are not OSI or FSF approved: %s", strings.Join(badLicenses, ", "))
	}
	return layer4.NeedsReview, "All license found are OSI or FSF approved"
}
