package quality

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
	"github.com/revanite-io/sci/pkg/layer4"
)

func repoIsPublic(payloadData interface{}, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}
	if data.RepositoryMetadata.IsPublic() {
		return layer4.Passed, "Repository is public"
	}
	return layer4.Failed, "Repository is private"
}

func insightsListsRepositories(payloadData interface{}, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	if len(data.Insights.Project.Repositories) > 0 {
		return layer4.Passed, "Insights contains a list of repositories"
	}

	return layer4.Failed, "Insights does NOT contains a list of repositories"
}

func statusChecksAreRequiredByRulesets(payloadData interface{}, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	// get the name of all status checks that were run
	var statusChecks []string
	for _, check := range data.Repository.DefaultBranchRef.Target.Commit.AssociatedPullRequests.Nodes {
		for _, run := range check.StatusCheckRollup.Commit.CheckSuites.Nodes {
			for _, checkRun := range run.CheckRuns.Nodes {
				statusChecks = append(statusChecks, checkRun.Name)
			}
		}
	}

	// get the rules that apply to the default branch
	rules := data.GetRulesets(data.Repository.DefaultBranchRef.Name)
	if len(rules) == 0 {
		return layer4.Passed, "No rulesets found for default branch, continuing to evaluate branch protection"
	}

	// get the name of all required status checks
	var requiredChecks []string
	for _, rule := range data.Rulesets {
		for _, requiredCheck := range rule.Parameters.RequiredChecks {
			requiredChecks = append(requiredChecks, requiredCheck.Context)
		}
	}

	// check whether all executed checks are required
	missingChecks := []string{}
	for _, check := range statusChecks {
		found := false
		for _, requiredCheck := range requiredChecks {
			if check == requiredCheck {
				found = true
				break
			}
		}
		if !found {
			missingChecks = append(missingChecks, check)
		}
	}

	if len(missingChecks) > 0 {
		return layer4.Failed, fmt.Sprintf("Some executed status checks are not mandatory but all should be: %s (NOTE: Not continuing to evaluate branch protection: combining requirements in rulesets and branch protection is not recommended)", strings.Join(missingChecks, ", "))
	}

	return layer4.Passed, "No status checks were run that are not required by the rules"
}

func statusChecksAreRequiredByBranchProtection(payloadData interface{}, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	// get the name of all status checks that were run
	var statusChecks []string
	for _, check := range data.Repository.DefaultBranchRef.Target.Commit.AssociatedPullRequests.Nodes {
		for _, run := range check.StatusCheckRollup.Commit.CheckSuites.Nodes {
			for _, checkRun := range run.CheckRuns.Nodes {
				statusChecks = append(statusChecks, checkRun.Name)
			}
		}
	}

	requiredChecks := data.Repository.DefaultBranchRef.BranchProtectionRule.RequiredStatusCheckContexts

	// check whether all executed checks are required
	missingChecks := []string{}
	for _, check := range statusChecks {
		found := false
		for _, requiredCheck := range requiredChecks {
			if check == requiredCheck {
				found = true
				break
			}
		}
		if !found {
			missingChecks = append(missingChecks, check)
		}
	}

	if len(missingChecks) > 0 {
		return layer4.Failed, fmt.Sprintf("Some executed status checks are not mandatory but all should be: %s", strings.Join(missingChecks, ", "))
	}

	return layer4.Passed, "No status checks were run that are not required by branch protection"
}

// TODO: after 3 layers of depth, make additional API calls if a tree is found.
// TODO: Examine more than just the file name.
func noBinariesInRepo(payloadData interface{}, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	if len(data.SuspectedBinaries) == 0 {
		return layer4.Passed, "No binaries were found in the repository (Note: this check only examines file names at this time)"
	}
	return layer4.Failed, fmt.Sprintf("Suspected binaries found in the repository: %s", strings.Join(data.SuspectedBinaries, ", "))
}

func requiresNonAuthorApproval(payloadData interface{}, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}
	protection := data.Repository.DefaultBranchRef.BranchProtectionRule

	// Check if reviews are required
	if !protection.RequiresApprovingReviews {
		return layer4.Failed, "Branch protection rule does not require reviews"
	}

	// Check if at least one review is required
	reviewCount := data.Repository.DefaultBranchRef.RefUpdateRule.RequiredApprovingReviewCount
	if reviewCount < 1 {
		return layer4.Failed, "Branch protection rule requires 0 approving reviews"
	}

	// Check if new commits dismiss previous approvals
	if !protection.RequireLastPushApproval {
		return layer4.Failed, "Branch protection does not require re-approval after new commits"
	}

	return layer4.Passed, fmt.Sprintf("Branch protection requires %d approving reviews and re-approval after new commits", reviewCount)
}

func hasOneOrMoreStatusChecks(payloadData interface{}, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	// get the name of all status checks that were run
	var statusChecks []string
	for _, check := range data.Repository.DefaultBranchRef.Target.Commit.AssociatedPullRequests.Nodes {
		for _, run := range check.StatusCheckRollup.Commit.CheckSuites.Nodes {
			for _, checkRun := range run.CheckRuns.Nodes {
				statusChecks = append(statusChecks, checkRun.Name)
			}
		}
	}

	if len(statusChecks) > 0 {
		return layer4.Passed, fmt.Sprintf("%d status checks were run", len(statusChecks))
	}

	return layer4.Failed, "No status checks were run"
}

func verifyDependencyManagement(payloadData interface{}, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	// Validate required fields
	if data.Repository.Name == "" || data.Repository.DefaultBranchRef.Name == "" ||
		data.Repository.DefaultBranchRef.Target.OID == "" {
		return layer4.Unknown, "Missing required repository data"
	}

	// Check dependency manifests
	result, message = verifyDependencyManifests(data)
	if result != layer4.Passed {
		return result, message
	}

	return layer4.Passed, "Dependency management files and SBOM requirements met"
}

type DependencyManifest struct {
	Filename string
	Language string
}

// Known dependency management files for various languages
var knownManifests = []DependencyManifest{
	{"go.mod", "Go"},
	{"package.json", "JavaScript/Node.js"},
	{"pom.xml", "Java (Maven)"},
	{"build.gradle", "Java (Gradle)"},
	{"requirements.txt", "Python"},
	{"Pipfile", "Python (Pipenv)"},
	{"pyproject.toml", "Python (Poetry)"},
	{"Gemfile", "Ruby"},
	{"composer.json", "PHP"},
	{"Cargo.toml", "Rust"},
	{"*.csproj", ".NET"},
	{"mix.exs", "Elixir"},
}

// ManifestResult stores the validation result for a manifest
type ManifestResult struct {
	Found    bool
	Language string
	HasDeps  bool
	Manifest string
}

func isManifestFile(filename string, pattern string) bool {
	if strings.Contains(pattern, "*") {
		matched, _ := filepath.Match(pattern, filename)
		return matched
	}
	return filename == pattern
}

func verifyDependencyManifests(payloadData interface{}) (layer4.Result, string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	if data.Repository.Object.Tree.Entries == nil {
		return layer4.Unknown, "Repository tree entries not available"
	}

	entries := data.Repository.Object.Tree.Entries
	manifests := data.Repository.DependencyGraphManifests.Nodes

	foundManifests := make(map[string]ManifestResult)
	manifestErrors := []string{}

	// Check repository contents for dependency files
	for _, entry := range entries {
		for _, manifest := range knownManifests {
			if isManifestFile(entry.Name, manifest.Filename) {
				result := ManifestResult{
					Found:    true,
					Language: manifest.Language,
					Manifest: entry.Name,
				}
				foundManifests[manifest.Language] = result
			}
		}
	}

	if len(manifestErrors) > 0 {
		return layer4.Failed, fmt.Sprintf("Dependency management issues found:\n%s",
			strings.Join(manifestErrors, "\n"))
	}

	if len(foundManifests) == 0 {
		return layer4.Failed, "No dependency management files found"
	}

	for _, manifest := range manifests {
		if manifest.Dependencies.TotalCount > 0 {
			for lang, result := range foundManifests {
				if manifest.Filename == result.Manifest {
					result.HasDeps = true
					foundManifests[lang] = result
					break
				}
			}
		}
	}

	missingDeps := []string{}
	for lang, result := range foundManifests {
		if !result.HasDeps {
			missingDeps = append(missingDeps,
				fmt.Sprintf("%s (%s)", result.Manifest, lang))
		}
	}

	if len(missingDeps) > 0 {
		return layer4.Failed, fmt.Sprintf("No dependencies declared in: %s",
			strings.Join(missingDeps, ", "))
	}

	return layer4.Passed, "Dependency management files present and properly configured"
}

func documentsTestExecution(payloadData interface{}, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	_, message = reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}
	return layer4.NeedsReview, "Review project documentation to ensure it explains when and how tests are run"
}

func documentsTestMaintenancePolicy(payloadData interface{}, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	_, message = reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}
	return layer4.NeedsReview, "Review project documentation to ensure it contains a clear policy for maintaining tests"
}
