package governance

import (
	"regexp"
	"strings"

	"github.com/gemaraproj/go-gemara"
	"github.com/ossf/pvtr-github-repo-scanner/data"
)

// ContributionGuideFiles are the conventional CONTRIBUTING filenames that GitHub
// and the OSPS Baseline recognize as a documented contribution process. Matched
// case-insensitively against repository contents.
var ContributionGuideFiles = []string{
	"CONTRIBUTING.md",
	"CONTRIBUTING",
	"CONTRIBUTING.rst",
	"CONTRIBUTING.txt",
}

// governanceDocDirs are the locations GitHub and common convention recognize for
// governance and ownership documentation: repository root, .github, and docs.
var governanceDocDirs = []string{"", ".github", "docs"}

// coreTeamFiles name a project's maintainers or owners; any one of them
// constitutes a listing of the core team.
var coreTeamFiles = []string{"MAINTAINERS.md", "MAINTAINERS", "CODEOWNERS", "GOVERNANCE.md", "GOVERNANCE"}

// rolesAndResponsibilitiesFiles document how a project is governed and who is
// responsible for what.
var rolesAndResponsibilitiesFiles = []string{"GOVERNANCE.md", "GOVERNANCE", "MAINTAINERS.md", "MAINTAINERS"}

func CoreTeamIsListed(payload data.Payload) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	if len(payload.Insights.Repository.CoreTeam) > 0 {
		return gemara.Passed, "Core team was specified in Security Insights data", gemara.High
	}

	// Fallback: a maintainers/owners file is itself a listing of the core team.
	if payload.RestData != nil {
		if path := payload.FindFileInDirs(governanceDocDirs, coreTeamFiles); path != "" {
			return gemara.Passed, "Core team listing found via GitHub (" + path + ")", gemara.Medium
		}
	}

	return gemara.Failed, "Core team was NOT specified in Security Insights data or via a maintainers/owners file on GitHub", gemara.Medium
}

func ProjectAdminsListed(payload data.Payload) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	if len(payload.Insights.Project.Administrators) > 0 {
		return gemara.Passed, "Project admins were specified in Security Insights data", gemara.High
	}

	// Project administrators hold administrative (destructive) access — a distinct,
	// more privileged role than the maintainers/owners a MAINTAINERS or CODEOWNERS
	// file lists, so such a file is not evidence of who the admins are. Admin
	// membership is not publicly observable, so without a Security Insights
	// declaration it cannot be confirmed.
	return gemara.NeedsReview, "Project administrators are not declared in Security Insights data; admin membership is not determinable from public repository files, so manual review is required", gemara.Low
}

func HasRolesAndResponsibilities(payload data.Payload) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	if payload.Insights.Repository.Documentation.Governance != nil {
		return gemara.Passed, "Roles and responsibilities were specified in Security Insights data", gemara.High
	}

	// Fallback: governance or maintainers documentation defines roles and
	// responsibilities even when it is not declared in Security Insights.
	if payload.RestData != nil {
		if path := payload.FindFileInDirs(governanceDocDirs, rolesAndResponsibilitiesFiles); path != "" {
			return gemara.Passed, "Governance/maintainers documentation found via GitHub (" + path + ")", gemara.Medium
		}
	}

	return gemara.Failed, "Roles and responsibilities were NOT specified in Security Insights data or via governance/maintainers documentation on GitHub", gemara.Medium
}

func HasContributionGuide(payload data.Payload) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	hasCoCLocation := payload.Insights.Project.Documentation.CodeOfConduct != nil

	if hasCoCLocation && payload.Insights.Repository.Documentation.ContributingGuide != nil {
		return gemara.Passed, "Contributing guide specified in Security Insights data (Bonus: code of conduct location also specified)", gemara.High
	}

	// Fallback: an observed contribution guide satisfies the control's requirement
	// for a documented contribution process, so it Passes. The code of conduct
	// location stays a recommendation and never demotes the result.
	if evidence := contributionGuideEvidence(payload); evidence != "" {
		if hasCoCLocation {
			return gemara.Passed, "Contributing guide found via " + evidence + " (Bonus: code of conduct location specified in Security Insights data)", gemara.Medium
		}
		return gemara.Passed, "Contributing guide found via " + evidence + " (Recommendation: add code of conduct location to Security Insights data)", gemara.Medium
	}

	return gemara.Failed, "Contribution guide not found in Security Insights data or via GitHub API", gemara.Medium
}

// contributionGuideEvidence reports where a contribution guide was observed, or
// "" if none was found. It prefers the GitHub contributing-guidelines API, then
// falls back to a deterministic search of the repository root tree and contents
// (root and .github) so the many repositories that document contribution without
// declaring it in Security Insights are still credited.
func contributionGuideEvidence(payload data.Payload) string {
	if payload.GraphqlRepoData != nil {
		if payload.Repository.ContributingGuidelines.Body != "" {
			return "GitHub contributing-guidelines API"
		}
		for _, entry := range payload.Repository.Object.Tree.Entries {
			if entry.Type == "blob" && isContributionGuideName(entry.Name) {
				return "GitHub API (repository file " + entry.Path + ")"
			}
		}
	}
	if payload.RestData != nil {
		if path := payload.FindFile(ContributionGuideFiles...); path != "" {
			return "GitHub API (repository file " + path + ")"
		}
	}
	return ""
}

func isContributionGuideName(name string) bool {
	for _, candidate := range ContributionGuideFiles {
		if strings.EqualFold(name, candidate) {
			return true
		}
	}
	return false
}

func HasContributionReviewPolicy(payload data.Payload) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	if !payload.IsCodeRepo {
		return gemara.NotApplicable, "Repository contains no code - skipping code contribution policy check", confidence
	}
	if payload.Insights.Repository.Documentation.ReviewPolicy != nil {
		return gemara.Passed, "Code review guide was specified in Security Insights data", confidence
	}

	// GV-03.02 wants a contributor guide that documents the requirements for
	// acceptable contributions, which the OSPS recommendation places in
	// CONTRIBUTING.md. When Security Insights does not declare it but a
	// contribution guide is observed, that guide may well cover those
	// requirements — we just cannot confirm the content automatically, so flag it
	// for review rather than failing a repo that most likely documents this.
	if evidence := contributionGuideEvidence(payload); evidence != "" {
		return gemara.NeedsReview, "Contribution guide found via " + evidence + ", but Security Insights does not declare its requirements for acceptable contributions; manual review required to confirm the guide covers them", gemara.Low
	}

	return gemara.Failed, "No contributor guide documenting requirements for acceptable contributions found in Security Insights data or repository files", gemara.Medium
}

// Vocabulary for OSPS-GV-04.01. All patterns are word-boundary anchored so
// unrelated words cannot match as substrings.
var (
	// escalationVocabPattern recognizes text about collaborator roles and the
	// permissions those roles carry.
	escalationVocabPattern = regexp.MustCompile(`(?i)\b(?:maintainer|committer|collaborator|commit access|write access|push access|merge rights|admin(?:istrator)?\s+(?:role|rights|access)|elevated\s+(?:permissions?|privileges?|access)|escalated\s+permissions?|owner\s+role)\b`)

	// reviewVocabPattern recognizes text about a person being reviewed,
	// approved, nominated, or vetted.
	reviewVocabPattern = regexp.MustCompile(`(?i)\b(?:review(?:ed)?|approv(?:e|ed|al)|nominat(?:e|ed|ion)|vote(?:d)?|consensus|sponsor(?:ed)?|vett(?:ed|ing))\b`)

	// escalationObligationPattern recognizes language that makes the review a
	// requirement rather than a description.
	escalationObligationPattern = regexp.MustCompile(`(?i)\b(?:must|shall|required?|requires|only\s+(?:after|by|through|via)|needs?\s+to\s+be\s+approved)\b`)

	// escalationNegationPattern recognizes language that disclaims or weakens
	// the requirement within the same section, so contradictory prose is not
	// credited as a definitive policy (the lesson from the VM-05 prose checks).
	escalationNegationPattern = regexp.MustCompile(`(?i)\b(?:not\s+required|no\s+(?:formal\s+)?review|optional|informal(?:ly)?|do(?:es)?\s+not\s+require|without\s+(?:a\s+)?review)\b`)

	// fencedCodeBlockPattern strips fenced code blocks before prose analysis so
	// example snippets cannot satisfy or negate the policy vocabulary.
	fencedCodeBlockPattern = regexp.MustCompile("(?s)```.*?```")
)

// escalationPolicySections splits a documentation file into markdown-heading
// sections with fenced code blocks removed, so vocabulary co-occurrence is
// judged within one topical section rather than across an entire file.
func escalationPolicySections(content string) []string {
	cleaned := fencedCodeBlockPattern.ReplaceAllString(content, "")
	var sections []string
	var current []string
	flush := func() {
		if len(current) > 0 {
			sections = append(sections, strings.Join(current, "\n"))
			current = nil
		}
	}
	for _, line := range strings.Split(cleaned, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			flush()
		}
		current = append(current, line)
	}
	flush()
	return sections
}

// classifyEscalationSection reports whether a section mentions collaborator
// escalation together with review (mention), and whether it additionally states
// the review as an uncontradicted requirement (policy).
func classifyEscalationSection(section string) (mention bool, policy bool) {
	if !escalationVocabPattern.MatchString(section) || !reviewVocabPattern.MatchString(section) {
		return false, false
	}
	if !escalationObligationPattern.MatchString(section) {
		return true, false
	}
	if escalationNegationPattern.MatchString(section) {
		return true, false
	}
	return true, true
}

// HasEscalatedPermissionsReviewPolicy implements OSPS-GV-04.01: while active,
// the project documentation MUST have a policy that code collaborators are
// reviewed prior to granting escalated permissions to sensitive resources.
//
// This is distinct from GV-03.02 (requirements for acceptable contributions):
// the subject here is the person receiving elevated access, not the change
// being merged. Security Insights has no escalation-policy field, so evidence
// is layered: repository documentation prose is scanned for a section that
// pairs escalation vocabulary with review vocabulary; an uncontradicted
// requirement Passes at Medium (heuristic prose match), a weaker pairing or a
// declared Security Insights governance URL or a governance file needs human
// review, and an observed absence of every signal Fails, mirroring GV-03.02's
// terminal. Unreadable documentation is never reported as a violation.
func HasEscalatedPermissionsReviewPolicy(payload data.Payload) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	if !payload.IsCodeRepo {
		return gemara.NotApplicable, "Repository contains no code - skipping escalated-permissions policy check", gemara.High
	}

	files, docsErr := payload.GetDocumentationFiles()
	policyFile := ""
	mentionFile := ""
	for _, file := range files {
		for _, section := range escalationPolicySections(file.Content) {
			mention, policy := classifyEscalationSection(section)
			if policy && policyFile == "" {
				policyFile = file.Path
			}
			if mention && mentionFile == "" {
				mentionFile = file.Path
			}
		}
	}

	if policyFile != "" {
		return gemara.Passed, "Documentation requires code collaborators to be reviewed before receiving escalated permissions (" + policyFile + ")", gemara.Medium
	}
	if mentionFile != "" {
		return gemara.NeedsReview, "Documentation discusses collaborator roles and review (" + mentionFile + ") but does not state an uncontradicted requirement that collaborators are reviewed before escalated permissions are granted; manual review required", gemara.Medium
	}
	if payload.RestData != nil && payload.Insights.Repository != nil &&
		payload.Insights.Repository.Documentation != nil &&
		payload.Insights.Repository.Documentation.Governance != nil {
		return gemara.NeedsReview, "Governance documentation is declared in Security Insights; confirm it requires review of code collaborators before escalated permissions are granted", gemara.Medium
	}
	if file := governanceFilePresent(files); file != "" {
		return gemara.NeedsReview, "A governance document (" + file + ") is present but no escalation-review policy language was recognized; manual review required", gemara.Low
	}
	if docsErr != nil {
		return gemara.NeedsReview, "Repository documentation could not be fully read; manually review whether a collaborator escalation-review policy is documented", gemara.Low
	}

	return gemara.Failed, "No policy requiring review of code collaborators before granting escalated permissions was found in repository documentation or Security Insights data", gemara.Medium
}

// governanceFilePresent returns the path of a conventional governance or
// maintainers document among the fetched documentation files, or "".
func governanceFilePresent(files []data.DocumentationFile) string {
	for _, file := range files {
		base := strings.ToLower(file.Path)
		if slash := strings.LastIndex(base, "/"); slash >= 0 {
			base = base[slash+1:]
		}
		switch base {
		case "governance.md", "governance", "maintainers.md", "maintainers":
			return file.Path
		}
	}
	return ""
}
