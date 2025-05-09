package build_release

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/revanite-io/sci/pkg/layer4"
	"github.com/rhysd/actionlint"

	"github.com/revanite-io/pvtr-github-repo/data"
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
)

// https://securitylab.github.com/resources/github-actions-untrusted-input/
// List of untrusted inputs
var regex = `.*(github\.event\.issue\.title|` +
	`github\.event\.issue\.body|` +
	`github\.event\.pull_request\.title|` +
	`github\.event\.pull_request\.body|` +
	`github\.event\.comment\.body|` +
	`github\.event\.review\.body|` +
	`github\.event\.pages.*\.page_name|` +
	`github\.event\.commits.*\.message|` +
	`github\.event\.head_commit\.message|` +
	`github\.event\.head_commit\.author\.email|` +
	`github\.event\.head_commit\.author\.name|` +
	`github\.event\.commits.*\.author\.email|` +
	`github\.event\.commits.*\.author\.name|` +
	`github\.event\.pull_request\.head\.ref|` +
	`github\.event\.pull_request\.head\.label|` +
	`github\.event\.pull_request\.head\.repo\.default_branch|` +
	`github\.head_ref).*`

func cicdSanitizedInputParameters(payloadData interface{}, _ map[string]*layer4.Change) (result layer4.Result, message string) {

	// parse the payload and see if we pass our checks
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	// For each file in the payload
	for _, file := range data.Contents.WorkFlows {

		if file.Encoding != "base64" {
			return layer4.Failed, fmt.Sprintf("File %v is not base64 encoded", file.Name)
		}

		decoded, err := base64.StdEncoding.DecodeString(file.Content)
		if err != nil {
			return layer4.Failed, fmt.Sprintf("Error decoding workflow file: %v", err)
		}

		workflow, actionError := actionlint.Parse(decoded)
		if actionError != nil {
			return layer4.Failed, fmt.Sprintf("Error parsing workflow: %v", actionError)
		}

		// Check the workflow for untrusted inputs
		ok, message := checkWorkflowFileForUntrustedInputs(workflow)

		if !ok {
			return layer4.Failed, message
		}

	}

	return layer4.Passed, "CI/CD tools input sanitized"

}

func checkWorkflowFileForUntrustedInputs(workflow *actionlint.Workflow) (bool, string) {

	expression, _ := regexp.Compile(regex)
	var message strings.Builder

	for _, job := range workflow.Jobs {

		if job == nil {
			continue
		}

		//Check the step for untrusted inputs
		for _, step := range job.Steps {

			if step == nil {
				continue
			}

			// if it isn't an exec run get out of dodge
			run, ok := step.Exec.(*actionlint.ExecRun)
			if !ok || run.Run == nil {
				continue
			}

			varList := pullVariablesFromScript(run.Run.Value)

			for _, name := range varList {
				if expression.Match([]byte(name)) {
					message.WriteString(fmt.Sprintf("Untrusted input found: %v\n", name))
				}
			}
		}
	}

	if message.Len() > 0 {
		return false, message.String()
	}
	return true, ""

}

func pullVariablesFromScript(script string) []string {

	varlist := []string{}

	for {

		//strings.Inex returns the first instance of a string
		//if the string is not found it returns -1 indicating the end of the scan
		//if the string is found it returns the index of the first character of the string
		start := strings.Index(script, "${{")
		if start == -1 {
			break
		}

		//Scanning a new slice gives us the length of the varialbe at the index of the closing bracket
		len := strings.Index(script[start:], "}}")
		if len == -1 {
			//script is malformed somehow
			return nil
		}

		//Create a new slice starting at the first character after the opening bracket of len
		varlist = append(varlist, strings.TrimSpace(script[start+3:start+len]))

		script = script[start+len:]

	}

	return varlist

}

func releaseHasUniqueIdentifier(payloadData interface{}, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	var noNameCount int
	var sameNameFound []string
	var releaseNames = make(map[string]int)

	for _, release := range data.Releases {
		if release.Name == "" {
			noNameCount++
		} else if _, ok := releaseNames[release.Name]; ok {
			sameNameFound = append(sameNameFound, release.Name)
		} else {
			releaseNames[release.Name] = release.Id
		}
	}
	if noNameCount > 0 || len(sameNameFound) > 0 {
		sameNames := strings.Join(sameNameFound, ", ")
		message := []string{fmt.Sprintf("Found %v releases with no name", noNameCount)}
		if len(sameNameFound) > 0 {
			message = append(message, fmt.Sprintf("Found %v releases with the same name: %v", len(sameNameFound), sameNames))
		}
		return layer4.Failed, strings.Join(message, ". ")
	}
	return layer4.Passed, "All releases found have a unique name"
}

func getLinksFromProjectDocumentation(data data.Payload) (urls []string) {
	doc := data.Insights.Project.Documentation
	if doc == nil {
		return urls
	}
	if doc.DetailedGuide != nil {
		urls = append(urls, doc.DetailedGuide.String())
	}
	if doc.CodeOfConduct != nil {
		urls = append(urls, doc.CodeOfConduct.String())
	}
	if doc.QuickstartGuide != nil {
		urls = append(urls, doc.QuickstartGuide.String())
	}
	if doc.ReleaseProcess != nil {
		urls = append(urls, doc.ReleaseProcess.String())
	}
	if doc.SignatureVerification != nil {
		urls = append(urls, doc.SignatureVerification.String())
	}
	return urls
}

func getLinks(data data.Payload) (links []string) {
	si := data.Insights

	if len(si.Header.URL.String()) > 0 {
		links = append(links, si.Header.URL.String())
	}

	if si.Header.ProjectSISource != nil && len(si.Header.ProjectSISource.String()) > 0 {
		links = append(links, si.Header.ProjectSISource.String())
	}

	if si.Project != nil {
		for _, repo := range si.Project.Repositories {
			links = append(links, repo.Url.String())
		}
		links = append(links, getLinksFromProjectDocumentation(data)...)
		if si.Project.HomePage != nil {
			links = append(links, si.Project.HomePage.String())
		}
		if si.Project.Roadmap != nil {
			links = append(links, si.Project.Roadmap.String())
		}
		if si.Project.Funding != nil {
			links = append(links, si.Project.Funding.String())
		}

		if si.Project.VulnerabilityReporting.BugBountyProgram != nil {
			links = append(links, si.Project.VulnerabilityReporting.BugBountyProgram.String())
		}
		if si.Project.VulnerabilityReporting.SecurityPolicy != nil {
			links = append(links, si.Project.VulnerabilityReporting.SecurityPolicy.String())
		}
	}
	if si.Repository != nil {
		if len(si.Repository.Url.String()) > 0 {
			links = append(links, si.Repository.Url.String())
		}
		if len(si.Repository.License.Url.String()) > 0 {
			links = append(links, si.Repository.License.Url.String())
		}

		for _, tool := range si.Repository.SecurityPosture.Tools {
			links = append(links, tool.Results.Adhoc.Location.String())
			links = append(links, tool.Results.CI.Location.String())
			links = append(links, tool.Results.Release.Location.String())
		}
		for _, repo := range si.Repository.SecurityPosture.Assessments.ThirdPartyAssessment {
			links = append(links, repo.Evidence.String())
		}
		if si.Repository.SecurityPosture.Assessments.Self.Evidence != nil {
			links = append(links, si.Repository.SecurityPosture.Assessments.Self.Evidence.String())
		}
	}

	if data.RepositoryMetadata != nil && data.RepositoryMetadata.OrganizationBlogURL() != nil {
		links = append(links, *data.RepositoryMetadata.OrganizationBlogURL())
	}

	return links
}

func insecureURI(uri string) bool {
	if strings.HasPrefix(uri, "https://") ||
		strings.HasPrefix(uri, "ssh:") ||
		strings.HasPrefix(uri, "git:") ||
		strings.HasPrefix(uri, "git@") {
		return false
	}
	return true
}

func ensureInsightsLinksUseHTTPS(payloadData interface{}, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	links := getLinks(data)
	var badURIs []string
	for _, link := range links {
		if insecureURI(link) {
			badURIs = append(badURIs, link)
		}
	}
	if len(badURIs) > 0 {
		return layer4.Failed, fmt.Sprintf("The following links do not use HTTPS: %v", strings.Join(badURIs, ", "))
	}
	return layer4.Passed, "All links use HTTPS"
}

func ensureLatestReleaseHasChangelog(payloadData interface{}, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	releaseDescription := data.Repository.LatestRelease.Description
	if strings.Contains(releaseDescription, "Change Log") || strings.Contains(releaseDescription, "Changelog") {
		return layer4.Passed, "Mention of a changelog found in the latest release"
	}
	return layer4.Failed, "The latest release does not have mention of a changelog: \n" + releaseDescription
}

func insightsHasSlsaAttestation(payloadData interface{}, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	attestations := data.Insights.Repository.ReleaseDetails.Attestations

	for _, attestation := range attestations {
		if attestation.PredicateURI == "https://slsa.dev/provenance/v1" {
			return layer4.Passed, "Found SLSA attestation in security insights"
		}
	}
	return layer4.Failed, "No SLSA attestation found in security insights"
}

func distributionPointsUseHTTPS(payloadData interface{}, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}
	if data.Insights.Repository.ReleaseDetails == nil || (data.Insights.Repository.ReleaseDetails != nil && len(data.Insights.Repository.ReleaseDetails.DistributionPoints) == 0) {
		return layer4.NotApplicable, "No official distribution points found in Security Insights data"
	}
	distributionPoints := data.Insights.Repository.ReleaseDetails.DistributionPoints

	var badURIs []string
	for _, point := range distributionPoints {
		if insecureURI(point.Uri) {
			badURIs = append(badURIs, point.Uri)
		}
	}
	if len(badURIs) > 0 {
		return layer4.Failed, fmt.Sprintf("The following distribution points do not use HTTPS: %v", strings.Join(badURIs, ", "))
	}
	return layer4.Passed, "All distribution points use HTTPS"
}
