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

func canonicalSastToolID(id string) string {
	if strings.HasPrefix(id, "sonar") {
		return "sonar"
	}
	return id
}

// matchesGitHubGlob matches the branch-filter glob syntax needed here: * and **.
// GitHub's ref-filter grammar treats ?, +, and character classes as quantifiers
// or extended patterns whose semantics differ from ordinary globbing, so those
// are conservatively treated as no match rather than risking a false coverage
// claim for a branch GitHub would not actually match.
func matchesGitHubGlob(value, pattern string) bool {
	var expression strings.Builder
	expression.WriteByte('^')
	for i := 0; i < len(pattern); i++ {
		switch pattern[i] {
		case '*':
			if i+1 < len(pattern) && pattern[i+1] == '*' {
				expression.WriteString(".*")
				i++
			} else {
				expression.WriteString("[^/]*")
			}
		case '?', '+', '[', ']':
			return false
		default:
			expression.WriteString(regexp.QuoteMeta(pattern[i : i+1]))
		}
	}
	expression.WriteByte('$')
	return regexp.MustCompile(expression.String()).MatchString(value)
}

func branchFilterIncludes(filter *actionlint.WebhookEventFilter, branch string) bool {
	if filter.IsEmpty() {
		return true
	}
	included := false
	for _, pattern := range filter.Values {
		if pattern == nil {
			continue
		}
		value := pattern.Value
		negated := strings.HasPrefix(value, "!")
		value = strings.TrimPrefix(value, "!")
		if matchesGitHubGlob(branch, value) {
			included = !negated
		}
	}
	return included
}

func branchFilterExcludes(filter *actionlint.WebhookEventFilter, branch string) bool {
	if filter.IsEmpty() {
		return false
	}
	for _, pattern := range filter.Values {
		if pattern != nil && matchesGitHubGlob(branch, pattern.Value) {
			return true
		}
	}
	return false
}

// workflowCoversChanges reports whether a workflow runs for every pull request
// targeting the default branch or every pushed commit. The second return value
// marks relevant triggers whose filters prevent proving full coverage.
func workflowCoversChanges(workflow *actionlint.Workflow, defaultBranch string) (bool, bool) {
	uncertain := false
	for _, event := range workflow.On {
		switch event.EventName() {
		case "pull_request_target":
			// This event runs in the base repository context and does not prove
			// that the pull request's code is what the SAST tool analyzes.
			uncertain = true
			continue
		case "pull_request", "push":
		default:
			continue
		}

		webhook, ok := event.(*actionlint.WebhookEvent)
		if !ok {
			uncertain = true
			continue
		}
		types := make(map[string]bool)
		for _, eventType := range webhook.Types {
			if eventType != nil {
				types[eventType.Value] = true
			}
		}
		pathsFiltered := !webhook.Paths.IsEmpty() || !webhook.PathsIgnore.IsEmpty()

		if event.EventName() == "push" {
			branchesUnfiltered := webhook.Branches.IsEmpty() && webhook.BranchesIgnore.IsEmpty()
			tagsUnfiltered := webhook.Tags.IsEmpty() && webhook.TagsIgnore.IsEmpty()
			if pathsFiltered || !branchesUnfiltered || !tagsUnfiltered {
				uncertain = true
				continue
			}
			return true, false
		}

		typesCoverChanges := len(types) == 0 || (types["opened"] && types["synchronize"])
		branchCovered := defaultBranch != "" &&
			branchFilterIncludes(webhook.Branches, defaultBranch) &&
			!branchFilterExcludes(webhook.BranchesIgnore, defaultBranch)
		if typesCoverChanges && !pathsFiltered && branchCovered {
			return true, false
		}
		uncertain = true
	}
	return false, uncertain
}

// sastSource is a detected signal that SAST runs on changes: either a tool
// identifier or a workflow/job status-check context. toolID links workflow
// aliases to policy evidence for the same tool in Security Insights.
// workflowAlias prevents separate Insights declarations in the same tool family
// from lending policy evidence to each other.
type sastSource struct {
	name             string
	toolID           string
	workflowAlias    bool
	policyDocumented bool
}

type sastDetection struct {
	sources           []sastSource
	inspectionBlocked bool
	coverageProven    bool
}

// detectSastToolInInsights identifies SAST tools integrated into continuous
// integration. A non-empty rulesets declaration documents the policy applied by
// the tool; adhoc or release-only integrations do not evaluate incoming changes.
func detectSastToolInInsights(tools []si.SecurityTool) sastDetection {
	var detection sastDetection
	for _, tool := range tools {
		if strings.EqualFold(tool.Type, "SAST") && tool.Integration.Ci {
			name := strings.TrimSpace(tool.Name)
			if name == "" {
				name = "SAST"
			}
			toolID := canonicalSastToolID(matchSast(name))
			if toolID == "" {
				toolID = strings.ToLower(name)
			}
			detection.sources = append(detection.sources, sastSource{
				name:             name,
				toolID:           toolID,
				policyDocumented: len(tool.Rulesets) > 0,
			})
		}
	}
	return detection
}

// detectSastInWorkflows inspects workflow files for a known SAST tool invoked by
// a workflow that runs on changes. Local reusable workflows are resolved from
// the fetched files, while recognizable remote workflow names are detected
// directly. Any file or remote call that cannot be inspected is recorded so an
// absence of SAST evidence cannot be reported as a definitive failure.
func detectSastInWorkflows(files []data.WorkflowFile, defaultBranch string) sastDetection {
	var detection sastDetection
	parsed := make(map[string]*actionlint.Workflow)
	for _, file := range files {
		isWorkflowYAML := strings.HasSuffix(file.Name, ".yml") || strings.HasSuffix(file.Name, ".yaml")
		if !isWorkflowYAML {
			continue
		}
		if file.Truncated {
			detection.inspectionBlocked = true
			continue
		}
		workflow, err := actionlint.Parse([]byte(file.Content))
		if err != nil || workflow == nil {
			detection.inspectionBlocked = true
			continue
		}
		parsed[strings.TrimPrefix(file.Path, "./")] = workflow
	}

	var inspectJob func(*actionlint.Job, map[string]bool) ([]string, bool)
	inspectJob = func(job *actionlint.Job, visiting map[string]bool) ([]string, bool) {
		if id := jobUsesSast(job); id != "" {
			return []string{id}, false
		}
		if job.WorkflowCall == nil || job.WorkflowCall.Uses == nil {
			return nil, false
		}

		uses := strings.TrimSpace(job.WorkflowCall.Uses.Value)
		if id := matchSast(uses); id != "" {
			return []string{id}, false
		}
		if !strings.HasPrefix(uses, "./") {
			return nil, true
		}

		path := strings.TrimPrefix(uses, "./")
		if visiting[path] {
			return nil, true
		}
		called, ok := parsed[path]
		if !ok {
			return nil, true
		}
		visiting[path] = true
		defer delete(visiting, path)

		blocked := false
		for _, calledJob := range called.Jobs {
			if calledJob == nil {
				continue
			}
			sources, callBlocked := inspectJob(calledJob, visiting)
			blocked = blocked || callBlocked
			if len(sources) > 0 {
				if calledJob.Name != nil && calledJob.Name.Value != "" {
					sources = append(sources, calledJob.Name.Value)
				}
				if calledJob.ID != nil && calledJob.ID.Value != "" {
					sources = append(sources, calledJob.ID.Value)
				}
				return sources, blocked
			}
		}
		return nil, blocked
	}

	for path, workflow := range parsed {
		coversChanges, coverageUncertain := workflowCoversChanges(workflow, defaultBranch)
		detection.inspectionBlocked = detection.inspectionBlocked || coverageUncertain
		if !coversChanges {
			continue
		}

		for _, job := range workflow.Jobs {
			if job == nil {
				continue
			}
			sources, blocked := inspectJob(job, map[string]bool{path: true})
			detection.inspectionBlocked = detection.inspectionBlocked || blocked
			if len(sources) == 0 {
				continue
			}
			detection.coverageProven = true

			jobName := ""
			if job.Name != nil {
				jobName = strings.TrimSpace(job.Name.Value)
			}
			jobID := ""
			if job.ID != nil {
				jobID = strings.TrimSpace(job.ID.Value)
			}
			checkName := jobName
			if checkName == "" {
				checkName = jobID
			}
			toolID := canonicalSastToolID(matchSast(sources[0]))

			for _, source := range sources {
				// Tool identifiers are reliable exact contexts in their own
				// right. Workflow-detected aliases receive policy evidence later
				// only when Security Insights documents this same tool.
				if matchSast(source) != "" {
					detection.sources = append(detection.sources, sastSource{name: source, toolID: toolID, workflowAlias: true})
				}
			}
			if job.WorkflowCall == nil {
				if checkName != "" {
					detection.sources = append(detection.sources, sastSource{name: checkName, toolID: toolID, workflowAlias: true})
				}
			} else if checkName != "" {
				for _, calledJob := range sources[1:] {
					calledJob = strings.TrimSpace(calledJob)
					if calledJob != "" {
						detection.sources = append(detection.sources, sastSource{
							name:          checkName + " / " + calledJob,
							toolID:        toolID,
							workflowAlias: true,
						})
					}
				}
			}
		}
	}
	return detection
}

// associateSastPolicies propagates documented policy evidence from Security
// Insights to workflow check aliases for the same canonical SAST tool.
func associateSastPolicies(sources []sastSource) {
	documentedTools := make(map[string]bool)
	for _, source := range sources {
		if source.toolID != "" && source.policyDocumented {
			documentedTools[source.toolID] = true
		}
	}
	for i := range sources {
		if sources[i].workflowAlias && documentedTools[sources[i].toolID] {
			sources[i].policyDocumented = true
		}
	}
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

// requiredCheckMatchesSast reports whether a required status-check context
// exactly matches a detected tool or complete workflow/job context. withPolicy
// is true only when at least one matched source carries a documented SAST
// ruleset or policy, so enforcement of a policy-less tool is not mistaken for a
// documented gate even when an unrelated tool does document one.
func requiredCheckMatchesSast(requiredContexts []string, sastSources []sastSource) (matched bool, withPolicy bool) {
	for _, context := range requiredContexts {
		normalizedContext := strings.ToLower(strings.TrimSpace(context))
		if normalizedContext == "" {
			continue
		}
		for _, source := range sastSources {
			if normalizedContext == strings.ToLower(strings.TrimSpace(source.name)) {
				matched = true
				if source.policyDocumented {
					withPolicy = true
				}
			}
		}
	}
	return matched, withPolicy
}

// SastEnforcedOnChanges implements OSPS-VM-06.02: all changes to the codebase
// must be automatically evaluated for security weaknesses and blocked on
// violations. It confirms both that a SAST tool runs on changes (in CI, per
// Security Insights or a workflow triggered by pull_request/push) and that it is
// enforced as a required status check that blocks merges to the default branch.
func SastEnforcedOnChanges(payload data.Payload) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	var detection sastDetection
	if payload.RestData != nil && payload.Insights.Repository != nil {
		insightsDetection := detectSastToolInInsights(payload.Insights.Repository.SecurityPosture.Tools)
		detection.sources = append(detection.sources, insightsDetection.sources...)
	}
	files, err := payload.GetWorkflowFiles()
	if err != nil {
		detection.inspectionBlocked = true
	} else {
		defaultBranch := ""
		if payload.GraphqlRepoData != nil {
			defaultBranch = payload.Repository.DefaultBranchRef.Name
		}
		workflowDetection := detectSastInWorkflows(files, defaultBranch)
		detection.sources = append(detection.sources, workflowDetection.sources...)
		detection.inspectionBlocked = workflowDetection.inspectionBlocked
		detection.coverageProven = workflowDetection.coverageProven
	}
	associateSastPolicies(detection.sources)

	// Union the status-check contexts required on the default branch from both
	// sources the scanner can observe: classic branch protection (admin-only)
	// and repository rulesets (publicly readable).
	var requiredContexts []string
	if payload.GraphqlRepoData != nil {
		requiredContexts = append(requiredContexts, payload.Repository.DefaultBranchRef.BranchProtectionRule.RequiredStatusCheckContexts...)
	}
	if payload.RepositoryMetadata != nil {
		requiredContexts = append(requiredContexts, payload.RepositoryMetadata.RequiredStatusCheckContexts()...)
	}

	adminObservable := payload.RepositoryMetadata != nil && payload.RepositoryMetadata.ViewerCanAdminister()

	return evaluateSastEnforcement(detection, requiredContexts, adminObservable)
}

// evaluateSastEnforcement applies the OSPS-VM-06.02 decision matrix to the
// gathered signals, kept separate from data access so it can be unit tested with
// plain inputs.
func evaluateSastEnforcement(detection sastDetection, requiredContexts []string, adminObservable bool) (gemara.Result, string, gemara.ConfidenceLevel) {
	if detection.inspectionBlocked && !detection.coverageProven {
		return gemara.NeedsReview, "Workflow inspection was incomplete, so the scanner could not determine whether SAST runs on all changes", gemara.Low
	}
	if len(detection.sources) == 0 {
		return gemara.Failed, "No Static Application Security Testing runs on changes, in Security Insights or a CI workflow triggered by pull requests or pushes", gemara.Medium
	}

	if matched, matchedWithPolicy := requiredCheckMatchesSast(requiredContexts, detection.sources); matched {
		if !matchedWithPolicy {
			return gemara.NeedsReview, "A SAST tool runs in CI and is enforced as a required status check, but no documented SAST ruleset or policy was found in Security Insights", gemara.Medium
		}
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
