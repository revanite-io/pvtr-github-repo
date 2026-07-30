package vuln_management

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gemaraproj/go-gemara"
	"github.com/ossf/pvtr-github-repo-scanner/data"
	"github.com/ossf/si-tooling/v2/si"
	"github.com/rhysd/actionlint"
)

var (
	// emailPattern matches a plausible contact email address.
	emailPattern = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	// urlPattern matches an http(s) link, treated as a reporting destination.
	urlPattern = regexp.MustCompile(`https?://\S+`)
	// reportingPhrases are wordings that signal documented private-reporting
	// instructions even when no address or link is spelled out inline.
	reportingPhrases = []string{
		"private vulnerability reporting",
		"report a vulnerability",
		"reporting a vulnerability",
		"security advisor", // covers "security advisory" and "security advisories"
		"report it privately",
		"report privately",
	}
)

// securityContactInPolicy reports whether a SECURITY.md body documents a way to
// reach the maintainers about a vulnerability, and names the evidence found so
// the assessment message can state its source.
func securityContactInPolicy(content string) (found bool, via string) {
	if emailPattern.MatchString(content) {
		return true, "contact email"
	}
	lower := strings.ToLower(content)
	for _, phrase := range reportingPhrases {
		if strings.Contains(lower, phrase) {
			return true, "private-reporting instructions"
		}
	}
	if urlPattern.MatchString(content) {
		return true, "reporting URL"
	}
	return false, ""
}

func HasSecContact(payload data.Payload) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	if payload.Insights.Project.VulnerabilityReporting.Contact.Email != nil {
		return gemara.Passed, "Security contacts were specified in Security Insights data", gemara.High
	}
	for _, champion := range payload.Insights.Repository.SecurityPosture.Champions {
		if champion.Email != nil {
			return gemara.Passed, "Security contacts were specified in Security Insights data", gemara.High
		}
	}

	if payload.SecurityPolicy.Present {
		if found, via := securityContactInPolicy(payload.SecurityPolicy.Content); found {
			return gemara.Passed, fmt.Sprintf("An email address was found in SECURITY.md (%s)", via), gemara.Medium
		}
	}

	if payload.PrivateVulnReporting.Enabled {
		return gemara.Passed, "GitHub private vulnerability reporting is enabled as a documented reporting channel", gemara.High
	}

	if payload.SecurityPolicy.Present {
		return gemara.NeedsReview, "A SECURITY.md file was found via GitHub but no recognizable security contact could be identified in it", gemara.Low
	}

	return gemara.Failed, "No security contact found in Security Insights data, a SECURITY.md file, or GitHub private vulnerability reporting", gemara.Medium
}

// knownSastIdentifiers are lowercase substrings that identify a Static
// Application Security Testing tool by its action reference, invoked command, or
// the status-check name it publishes. Kept deliberately narrow to SAST: secret
// scanners, dependency/SCA tools, and container/IaC misconfiguration scanners
// are intentionally excluded so they cannot be mistaken for a code-analysis gate.
var knownSastIdentifiers = []string{
	"codeql",
	"semgrep",
	"opengrep",
	"sonarcloud",
	"sonarqube",
	"sonar-scanner",
	"sonarsource",
	"checkmarx",
	"snyk code",
	"bandit",
	"gosec",
	"brakeman",
	"bearer",
	"horusec",
	"deepsource",
}

// matchSast returns the first known SAST identifier that appears in text, or the
// empty string when none is present. Matching is case-insensitive.
func matchSast(text string) string {
	lower := strings.ToLower(text)
	for _, id := range knownSastIdentifiers {
		if strings.Contains(lower, id) {
			return id
		}
	}
	return ""
}

// isWorkflowYAML reports whether a file name is a GitHub Actions workflow
// definition by extension.
func isWorkflowYAML(name string) bool {
	return strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml")
}

// workflowRunsOnChanges reports whether a workflow is triggered by an event that
// fires on changes to the codebase, i.e. a pull request or a push. Only these
// can act as a gate on incoming changes; scheduled or manual runs cannot.
func workflowRunsOnChanges(workflow *actionlint.Workflow) bool {
	for _, event := range workflow.On {
		switch event.EventName() {
		case "pull_request", "pull_request_target", "push":
			return true
		}
	}
	return false
}

// detectSastToolInInsights returns correlation names for every Security Insights
// tool declared as a SAST tool integrated into continuous integration. Adhoc or
// release-only integrations are ignored because they do not evaluate incoming
// changes.
func detectSastToolInInsights(tools []si.SecurityTool) []string {
	var names []string
	for _, tool := range tools {
		if strings.EqualFold(tool.Type, "SAST") && tool.Integration.Ci {
			name := strings.TrimSpace(tool.Name)
			if name == "" {
				name = "SAST"
			}
			names = append(names, name)
		}
	}
	return names
}

// detectSastInWorkflows inspects workflow files for a known SAST tool invoked by
// a workflow that runs on changes. It returns correlation names (the identifier,
// the workflow name, and the job name/ID) that a required status-check context
// can be matched against. Files that cannot be read or parsed are skipped rather
// than treated as evidence.
func detectSastInWorkflows(files []data.WorkflowFile) []string {
	var sources []string
	for _, file := range files {
		if !isWorkflowYAML(file.Name) || file.Truncated {
			continue
		}
		workflow, err := actionlint.Parse([]byte(file.Content))
		if err != nil || workflow == nil {
			continue
		}
		if !workflowRunsOnChanges(workflow) {
			continue
		}

		workflowName := ""
		if workflow.Name != nil {
			workflowName = workflow.Name.Value
		}

		for _, job := range workflow.Jobs {
			if job == nil {
				continue
			}
			id := jobUsesSast(job)
			if id == "" {
				continue
			}
			sources = append(sources, id)
			if workflowName != "" {
				sources = append(sources, workflowName)
			}
			if job.Name != nil && job.Name.Value != "" {
				sources = append(sources, job.Name.Value)
			}
			if job.ID != nil && job.ID.Value != "" {
				sources = append(sources, job.ID.Value)
			}
		}
	}
	return sources
}

// jobUsesSast returns the SAST identifier a job invokes via a step's `uses:`
// action, `run:` command, or step name, or the empty string when none is found.
func jobUsesSast(job *actionlint.Job) string {
	for _, step := range job.Steps {
		if step == nil {
			continue
		}
		switch exec := step.Exec.(type) {
		case *actionlint.ExecAction:
			if exec.Uses != nil {
				if id := matchSast(exec.Uses.Value); id != "" {
					return id
				}
			}
		case *actionlint.ExecRun:
			if exec.Run != nil {
				if id := matchSast(exec.Run.Value); id != "" {
					return id
				}
			}
		}
		if step.Name != nil {
			if id := matchSast(step.Name.Value); id != "" {
				return id
			}
		}
	}
	return ""
}

// requiredCheckMatchesSast reports whether any required status-check context
// corresponds to a detected SAST tool: either the context name itself carries a
// known SAST identifier, or it correlates with one of the SAST source names
// (workflow/job names) that produced the check.
func requiredCheckMatchesSast(requiredContexts, sastSources []string) bool {
	for _, context := range requiredContexts {
		trimmed := strings.TrimSpace(context)
		if trimmed == "" {
			continue
		}
		if matchSast(trimmed) != "" {
			return true
		}
		lowerContext := strings.ToLower(trimmed)
		for _, source := range sastSources {
			lowerSource := strings.ToLower(strings.TrimSpace(source))
			// Guard against short, generic tokens producing spurious substring
			// matches (e.g. a job called "ci").
			if len(lowerSource) < 4 {
				continue
			}
			if strings.Contains(lowerContext, lowerSource) || strings.Contains(lowerSource, lowerContext) {
				return true
			}
		}
	}
	return false
}

// gatherRequiredCheckContexts unions the status-check contexts required on the
// default branch from both sources the scanner can observe: classic branch
// protection (admin-only) and repository rulesets (publicly readable).
func gatherRequiredCheckContexts(payload data.Payload) []string {
	var contexts []string
	if payload.GraphqlRepoData != nil {
		contexts = append(contexts, payload.Repository.DefaultBranchRef.BranchProtectionRule.RequiredStatusCheckContexts...)
	}
	if payload.RepositoryMetadata != nil {
		contexts = append(contexts, payload.RepositoryMetadata.RequiredStatusCheckContexts()...)
	}
	return contexts
}

// SastEnforcedOnChanges implements OSPS-VM-06.02: all changes to the codebase
// must be automatically evaluated for security weaknesses and blocked on
// violations. It confirms both that a SAST tool runs on changes (in CI, per
// Security Insights or a workflow triggered by pull_request/push) and that it is
// enforced as a required status check that blocks merges to the default branch.
func SastEnforcedOnChanges(payload data.Payload) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	var sastSources []string
	if payload.RestData != nil && payload.Insights.Repository != nil {
		sastSources = append(sastSources, detectSastToolInInsights(payload.Insights.Repository.SecurityPosture.Tools)...)
	}
	if files, err := payload.GetWorkflowFiles(); err == nil {
		sastSources = append(sastSources, detectSastInWorkflows(files)...)
	}

	requiredContexts := gatherRequiredCheckContexts(payload)
	adminObservable := payload.RepositoryMetadata != nil && payload.RepositoryMetadata.ViewerCanAdminister()

	return evaluateSastEnforcement(sastSources, requiredContexts, adminObservable)
}

// evaluateSastEnforcement applies the OSPS-VM-06.02 decision matrix to the
// gathered signals, kept separate from data access so it can be unit tested with
// plain inputs.
func evaluateSastEnforcement(sastSources, requiredContexts []string, adminObservable bool) (gemara.Result, string, gemara.ConfidenceLevel) {
	if len(sastSources) == 0 {
		return gemara.Failed, "No Static Application Security Testing runs on changes, in Security Insights or a CI workflow triggered by pull requests or pushes", gemara.Medium
	}

	if requiredCheckMatchesSast(requiredContexts, sastSources) {
		return gemara.Passed, "A SAST tool runs in CI and is enforced as a required status check on the default branch, blocking merges on violations", gemara.High
	}

	if len(requiredContexts) > 0 {
		return gemara.NeedsReview, "A SAST tool runs in CI and required status checks are configured on the default branch, but none could be matched to the SAST tool; confirm the SAST check must pass before merging", gemara.Medium
	}

	if adminObservable {
		return gemara.Failed, "A SAST tool runs in CI but no required status check enforces it on the default branch, so violations do not block merges", gemara.High
	}

	return gemara.NeedsReview, "A SAST tool runs in CI but default branch protection is not observable with the current token; confirm a required status check enforces the SAST tool before merging", gemara.Low
}

func HasVulnerabilityDisclosurePolicy(payload data.Payload) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	if payload.Insights.Project.VulnerabilityReporting.Policy != nil {
		return gemara.Passed, "Vulnerability disclosure policy was specified in Security Insights data", gemara.High
	}

	// A GitHub-observed security policy document (a SECURITY.md, which is also what
	// sets IsSecurityPolicyEnabled) proves a policy file exists but not that it is
	// a coordinated vulnerability disclosure policy with a clear response
	// timeframe, which the requirement demands. Its presence cannot confirm those
	// clauses, so defer to a human rather than passing on the file alone.
	if payload.SecurityPolicy.Present {
		return gemara.NeedsReview, "A SECURITY.md file was found in the repository via GitHub; its CVD policy content and response timeframe need human confirmation", gemara.High
	}

	if payload.Repository.IsSecurityPolicyEnabled {
		return gemara.NeedsReview, "GitHub reports a security policy is enabled for the repository; its CVD policy content and response timeframe need human confirmation", gemara.High
	}

	return gemara.Failed, "No vulnerability disclosure policy found in Security Insights data, a GitHub security policy, or a SECURITY.md file", gemara.Medium
}

func HasPrivateVulnerabilityReporting(payload data.Payload) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	if payload.Insights.Project.VulnerabilityReporting.ReportsAccepted {
		if payload.Insights.Project.VulnerabilityReporting.Contact.Email != nil {
			return gemara.Passed, "Private vulnerability reporting available via dedicated contact email in Security Insights data", gemara.High
		}

		for _, champion := range payload.Insights.Repository.SecurityPosture.Champions {
			if champion.Email != nil {
				return gemara.Passed, "Private vulnerability reporting available via security champions contact in Security Insights data", gemara.High
			}
		}
	}

	if payload.PrivateVulnReporting.Enabled {
		return gemara.Passed, "No Security Insights contact, but GitHub private vulnerability reporting is enabled for the repository", gemara.Medium
	}

	if !payload.PrivateVulnReporting.Known {
		return gemara.NeedsReview, "No private vulnerability reporting contact in Security Insights data and GitHub private vulnerability reporting status could not be determined", gemara.Low
	}

	return gemara.Failed, "No private vulnerability reporting contact in Security Insights data and GitHub private vulnerability reporting is disabled", gemara.Medium
}
