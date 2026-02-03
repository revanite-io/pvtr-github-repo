package build_release

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/gemaraproj/go-gemara"
	"github.com/rhysd/actionlint"

	"github.com/revanite-io/pvtr-github-repo/data"
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
)

// https://securitylab.github.com/resources/github-actions-untrusted-input/
// List of untrusted inputs; Global for use in tests also
var untrustedVarsRegex = `.*(github\.event\.issue\.title|` +
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

func CicdSanitizedInputParameters(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {

	// parse the payload and see if we pass our checks
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message
	}
	workflows, err := data.GetDirectoryContent(".github/workflows")
	if len(workflows) == 0 {
		if err != nil {
			message = err.Error()
		} else {
			message = "No workflows found in .github/workflows directory"
		}
		return gemara.NotApplicable, message
	}

	for _, file := range workflows {
		if !strings.HasSuffix(*file.Name, ".yml") && !strings.HasSuffix(*file.Name, ".yaml") {
			continue
		}

		if *file.Encoding != "base64" {
			return gemara.Failed, fmt.Sprintf("File %v is not base64 encoded", file.Name)
		}

		decoded, err := base64.StdEncoding.DecodeString(*file.Content)
		if err != nil {
			return gemara.Failed, fmt.Sprintf("Error decoding workflow file: %v", err)
		}

		workflow, actionError := actionlint.Parse(decoded)
		if actionError != nil {
			return gemara.Failed, fmt.Sprintf("Error parsing workflow: %v (%s)", actionError, *file.Path)
		}

		// Check the workflow for untrusted inputs
		ok, message := checkWorkflowFileForUntrustedInputs(workflow)

		if !ok {
			return gemara.Failed, message
		}
	}

	return gemara.Passed, "GitHub Workflows variables do not contain untrusted inputs"

}

func checkWorkflowFileForUntrustedInputs(workflow *actionlint.Workflow) (bool, string) {

	untrustedVars, _ := regexp.Compile(untrustedVarsRegex)
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
				if untrustedVars.Match([]byte(name)) {
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

func ReleaseHasUniqueIdentifier(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message
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
		return gemara.Failed, strings.Join(message, ". ")
	}
	return gemara.Passed, "All releases found have a unique name"
}

func getLinks(data data.Payload) []string {
	si := data.Insights
	links := []string{
		si.Header.URL,
		si.Header.ProjectSISource,
		si.Project.Homepage,
		si.Project.Roadmap,
		si.Project.Funding,
		si.Project.Documentation.DetailedGuide,
		si.Project.Documentation.CodeOfConduct,
		si.Project.Documentation.QuickstartGuide,
		si.Project.Documentation.ReleaseProcess,
		si.Project.Documentation.SignatureVerification,
		si.Project.Vulnerability.BugBountyProgram,
		si.Project.Vulnerability.SecurityPolicy,
		si.Repository.URL,
		si.Repository.License.URL,
		si.Repository.Security.Assessments.Self.Evidence,
	}
	if data.RepositoryMetadata.OrganizationBlogURL() != nil {
		links = append(links, *data.RepositoryMetadata.OrganizationBlogURL())
	}
	for _, repo := range si.Project.Repositories {
		links = append(links, repo.URL)
	}

	for _, repo := range si.Repository.Security.Assessments.ThirdParty {
		links = append(links, repo.Evidence)
	}

	for _, tool := range si.Repository.Security.Tools {
		links = append(links, tool.Results.Adhoc.Location)
		links = append(links, tool.Results.CI.Location)
		links = append(links, tool.Results.Release.Location)
	}
	return links
}

func insecureURI(uri string) bool {
	if !strings.HasPrefix(uri, "https://") ||
		!strings.HasPrefix(uri, "ssh:") ||
		!strings.HasPrefix(uri, "git:") ||
		!strings.HasPrefix(uri, "git@") {
		return false
	}
	return true
}

func EnsureInsightsLinksUseHTTPS(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message
	}

	links := getLinks(data)
	var badURIs []string
	for _, link := range links {
		if insecureURI(link) {
			badURIs = append(badURIs, link)
		}
	}
	if len(badURIs) > 0 {
		return gemara.Failed, fmt.Sprintf("The following links do not use HTTPS: %v", strings.Join(badURIs, ", "))
	}
	return gemara.Passed, "All links use HTTPS"
}

func EnsureLatestReleaseHasChangelog(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message
	}

	releaseDescription := data.Repository.LatestRelease.Description
	if strings.Contains(releaseDescription, "Change Log") || strings.Contains(releaseDescription, "Changelog") {
		return gemara.Passed, "Mention of a changelog found in the latest release"
	}
	return gemara.Failed, "The latest release does not have mention of a changelog: \n" + releaseDescription
}

func InsightsHasSlsaAttestation(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message
	}

	attestations := data.Insights.Repository.Release.Attestations

	for _, attestation := range attestations {
		if attestation.PredicateURI == "https://slsa.dev/provenance/v1" {
			return gemara.Passed, "Found SLSA attestation in security insights"
		}
	}
	return gemara.Failed, "No SLSA attestation found in security insights"
}

func DistributionPointsUseHTTPS(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message
	}

	distributionPoints := data.Insights.Repository.Release.DistributionPoints

	if len(distributionPoints) == 0 {
		return gemara.NotApplicable, "No official distribution points found in Security Insights data"
	}

	var badURIs []string
	for _, point := range distributionPoints {
		if insecureURI(point.URI) {
			badURIs = append(badURIs, point.URI)
		}
	}
	if len(badURIs) > 0 {
		return gemara.Failed, fmt.Sprintf("The following distribution points do not use HTTPS: %v", strings.Join(badURIs, ", "))
	}
	return gemara.Passed, "All distribution points use HTTPS"
}

func SecretScanningInUse(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message
	}

	if data.SecurityPosture.PreventsPushingSecrets() && data.SecurityPosture.ScansForSecrets() {
		return gemara.Passed, "Secret scanning is enabled and prevents pushing secrets"
	} else if data.SecurityPosture.PreventsPushingSecrets() || data.SecurityPosture.ScansForSecrets() {
		return gemara.Failed, "Secret scanning is only partially enabled"
	} else {
		return gemara.Failed, "Secret scanning is not enabled"
	}
}
