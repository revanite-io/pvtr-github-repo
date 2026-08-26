package legal

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gemaraproj/go-gemara"
	"github.com/ossf/pvtr-github-repo-scanner/data"
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

func getLicenseList(payload data.Payload, makeApiCall func(string, bool) ([]byte, error)) (LicenseList, string) {
	GoodLicenseList := LicenseList{}
	if makeApiCall == nil {
		makeApiCall = payload.MakeApiCall
	}
	response, err := makeApiCall(spdxURL, false)
	if err != nil {
		return GoodLicenseList, fmt.Sprintf("Failed to fetch good license data: %s", err.Error())
	}
	err = json.Unmarshal(response, &GoodLicenseList)
	if err != nil {
		return GoodLicenseList, fmt.Sprintf("Failed to unmarshal good license data: %s", err.Error())
	}
	if len(GoodLicenseList.Licenses) == 0 {
		return GoodLicenseList, "Good license data was unexpectedly empty"
	}
	return GoodLicenseList, ""
}

func splitSpdxExpression(expression string) (spdx_ids []string) {
	// Remove grouping parentheses; they affect evaluation order, not approval.
	expression = strings.ReplaceAll(expression, "(", " ")
	expression = strings.ReplaceAll(expression, ")", " ")
	a := strings.Split(expression, " AND ")
	for _, aa := range a {
		b := strings.Split(aa, " OR ")
		for _, token := range b {
			token = strings.TrimSpace(token)
			// Strip " WITH <exception>" — the exception clause is irrelevant to
			// OSI/FSF approval; only the base license ID matters.
			if idx := strings.Index(token, " WITH "); idx != -1 {
				token = token[:idx]
			}
			// Strip the "+" (or-later) operator so any "id+" form (e.g. "Apache-2.0+")
			// resolves to its base license ID.
			token = strings.TrimSuffix(token, "+")
			spdx_ids = append(spdx_ids, token)
		}
	}
	return
}

// noAssertion is what GitHub's license detector reports when a license file is
// present but cannot be classified. It means "unidentified", not "disapproved".
const noAssertion = "NOASSERTION"

// licenseFilePrefix matches per-license files such as LICENSE-MIT / LICENSE-APACHE.
const licenseFilePrefix = "license-"

// rootLicenseFiles are conventional license filenames GitHub's detector
// sometimes fails to classify (nonstandard text, multi-license repos, or an
// unexpected location). Compared case-insensitively.
var rootLicenseFiles = []string{
	"LICENSE",
	"LICENCE",
	"COPYING",
	"COPYING.LESSER",
	"LICENSE.md",
	"LICENSE.txt",
}

// isLicenseFileName reports whether a filename is a conventional standalone
// license file: an exact well-known name (case-insensitive) or a per-license
// LICENSE-* file.
func isLicenseFileName(name string) bool {
	if strings.HasPrefix(strings.ToLower(name), licenseFilePrefix) {
		return true
	}
	for _, known := range rootLicenseFiles {
		if strings.EqualFold(name, known) {
			return true
		}
	}
	return false
}

// findRootLicenseFile returns the name of a license file present in the
// repository root tree, or "" if none is found. The repository root is itself a
// well-known location, so a conventionally-named file there is independent
// evidence a license exists even when GitHub's detector returns nothing.
func findRootLicenseFile(payload data.Payload) string {
	if payload.GraphqlRepoData == nil {
		return ""
	}
	for _, entry := range payload.Repository.Object.Tree.Entries {
		if entry.Type != "blob" {
			continue
		}
		if isLicenseFileName(entry.Name) {
			return entry.Name
		}
	}
	return ""
}

func FoundLicense(payload data.Payload) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	if payload.Repository.LicenseInfo.Url != "" {
		return gemara.Passed, "License was found in a well known location via the GitHub API", gemara.High
	}
	if file := findRootLicenseFile(payload); file != "" {
		return gemara.Passed, fmt.Sprintf("License file %q found in the repository root, a well known location; GitHub could not identify the license type", file), gemara.Medium
	}
	return gemara.Failed, "License was not found in a well known location via the GitHub API", gemara.Medium
}

// latestPublishedRelease returns the most recent non-draft release. GitHub
// lists releases newest-first, so the first non-draft entry is the latest
// published one. Draft releases are not published software and must not affect
// the assessment, even when an authenticated caller can observe them.
func latestPublishedRelease(releases []data.ReleaseData) (data.ReleaseData, bool) {
	for _, release := range releases {
		if !release.Draft {
			return release, true
		}
	}
	return data.ReleaseData{}, false
}

// licenseAssets returns the names of release assets that look like standalone
// license files attached at the top level of the release.
func licenseAssets(release data.ReleaseData) []string {
	var names []string
	for _, asset := range release.Assets {
		if isLicenseFileName(asset.Name) {
			names = append(names, asset.Name)
		}
	}
	return names
}

// releaseLabel names a release for messages, preferring the tag over the
// free-form release name.
func releaseLabel(release data.ReleaseData) string {
	if release.TagName != "" {
		return release.TagName
	}
	return release.Name
}

// ReleasesLicensed assesses whether released software assets include their
// license (OSPS-LE-03.02, and the license-presence half of OSPS-LE-02.02).
//
// The assessment is scoped to the latest published release, which reflects the
// project's current licensing posture. Two release-time hazards drive the
// design (see ossf/pvtr-github-repo-scanner#70):
//
//  1. A license file attached directly to the release can supersede whatever
//     ships inside the source archives. Its content is not observable here, so
//     its presence downgrades the result to NeedsReview rather than passing on
//     source evidence alone.
//  2. The license can be modified or removed at the release tag while the
//     default branch still shows an approved one. GitHub's auto-generated
//     source archives contain the tree at the tag, so the authoritative
//     evidence is the license GitHub detects at that tag, not at HEAD.
//
// Default-branch evidence is used only as a lower-confidence fallback when the
// tag-level lookup is unavailable.
func ReleasesLicensed(payload data.Payload) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	if payload.RestData == nil {
		return gemara.NeedsReview, "Release data is unavailable; review the released assets for license coverage", gemara.Low
	}
	if payload.ReleasesError != nil {
		return gemara.NeedsReview, fmt.Sprintf("Release data could not be retrieved: %v. Review the released assets for license coverage", payload.ReleasesError), gemara.Low
	}

	latest, ok := latestPublishedRelease(payload.Releases)
	if !ok {
		return gemara.NotApplicable, "No releases found", gemara.High
	}

	if assets := licenseAssets(latest); len(assets) > 0 {
		return gemara.NeedsReview, fmt.Sprintf("Release %q attaches standalone license file(s) as release assets (%s), which may supersede the license in the released source code; manual review is required to confirm they match an approved license", releaseLabel(latest), strings.Join(assets, ", ")), gemara.High
	}

	if latest.TagName != "" {
		license, found, err := payload.GetLicenseAtRef(latest.TagName)
		switch {
		case err == nil && found && license.SpdxId != "" && license.SpdxId != noAssertion:
			return gemara.Passed, fmt.Sprintf("GitHub identifies license %s at %q in the released source code at tag %q; the auto-generated release archives include it", license.SpdxId, license.Path, latest.TagName), gemara.High
		case err == nil && found:
			return gemara.Passed, fmt.Sprintf("License file %q is present in the released source code at tag %q, but GitHub could not identify the license type", license.Path, latest.TagName), gemara.Medium
		case err == nil:
			if payload.GraphqlRepoData != nil && payload.Repository.LicenseInfo.Url != "" {
				// The tamper/removal case #70 exists to catch: HEAD is
				// licensed, but the tree actually shipped is not.
				return gemara.Failed, fmt.Sprintf("A license exists on the default branch, but none was found in the released source code at tag %q", latest.TagName), gemara.High
			}
			return gemara.Failed, fmt.Sprintf("No license was found in the released source code at tag %q", latest.TagName), gemara.High
		default:
			if payload.Config != nil && payload.Config.Logger != nil {
				payload.Config.Logger.Warn(fmt.Sprintf("could not fetch license at release tag %q, falling back to default-branch evidence: %s", latest.TagName, err.Error()))
			}
		}
	}

	// Fallback: the released tree could not be checked directly (no tag name,
	// or the tag-level lookup failed), so the default branch stands in as a
	// weaker proxy for what the release archives contain.
	if payload.GraphqlRepoData != nil && payload.Repository.LicenseInfo.Url != "" {
		return gemara.Passed, "A license was found on the default branch; the released source code could not be checked directly, so this is default-branch evidence only", gemara.Medium
	}
	if file := findRootLicenseFile(payload); file != "" {
		return gemara.Passed, fmt.Sprintf("License file %q found in the repository root; the released source code could not be checked directly, so this is default-branch evidence only", file), gemara.Low
	}
	return gemara.Failed, "License was not found in a well known location via the GitHub API", gemara.Medium
}

func GoodLicense(payload data.Payload) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	licenses, errString := getLicenseList(payload, nil)

	if errString != "" {
		return gemara.Unknown, errString, confidence
	}

	apiInfo := payload.Repository.LicenseInfo.SpdxId
	siInfo := payload.Insights.Repository.License.Expression

	spdx_ids := append(splitSpdxExpression(apiInfo), splitSpdxExpression(siInfo)...)
	badLicenses := []string{}
	deprecatedApproved := []string{}
	usableIDs := []string{}
	for _, spdx_id := range spdx_ids {
		// An empty string comes from splitting an empty expression, and
		// NOASSERTION is GitHub's marker for a license it could not identify.
		// Neither is an SPDX identifier we can evaluate for approval.
		if spdx_id == "" || spdx_id == noAssertion {
			continue
		}
		usableIDs = append(usableIDs, spdx_id)
		var validId bool
		for _, license := range licenses.Licenses {
			if license.LicenseID == spdx_id {
				validId = true
				if !license.IsOsiApproved && !license.IsFsfLibre {
					badLicenses = append(badLicenses, spdx_id)
				} else if license.IsDeprecatedLicenseId {
					deprecatedApproved = append(deprecatedApproved, spdx_id)
				}
			}
		}
		if !validId {
			badLicenses = append(badLicenses, spdx_id)
		}
	}

	if len(usableIDs) == 0 {
		// No SPDX identifier is available from GitHub or Security Insights. A
		// license file in the root means one exists but its identity is unknown,
		// so a human must judge OSI/FSF approval rather than us passing or failing.
		if file := findRootLicenseFile(payload); file != "" {
			return gemara.NeedsReview, fmt.Sprintf("License file %q is present but its SPDX identity could not be determined; manual review is required to confirm OSI or FSF approval", file), gemara.High
		}
		return gemara.Failed, "License SPDX identifier was not found in Security Insights data or via GitHub API", gemara.Medium
	}

	approvedLicenses := strings.Join(usableIDs, ", ")
	if payload.Config.Logger != nil {
		payload.Config.Logger.Trace(fmt.Sprintf("Requested licenses: %s", approvedLicenses))
		payload.Config.Logger.Trace(fmt.Sprintf("Non-approved licenses: %s", badLicenses))
		payload.Config.Logger.Trace(fmt.Sprintf("Deprecated-but-approved licenses: %s", deprecatedApproved))
	}

	if len(badLicenses) > 0 {
		return gemara.Failed, fmt.Sprintf("These licenses are not OSI or FSF approved: %s", strings.Join(badLicenses, ", ")), gemara.High
	}
	if len(deprecatedApproved) > 0 {
		return gemara.Passed, fmt.Sprintf("All licenses found are OSI or FSF approved. Note: the following SPDX IDs are deprecated and should be migrated to their -only/-or-later form: %s", strings.Join(deprecatedApproved, ", ")), gemara.High
	}
	return gemara.Passed, "All license found are OSI or FSF approved", gemara.High
}
