package quality

import (
	"fmt"
	"strings"

	"github.com/ossf/gemara/layer4"
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
)

func repoIsPublic(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}
	if data.RepositoryMetadata.IsPublic() {
		return layer4.Passed, "Repository is public"
	}
	return layer4.Failed, "Repository is private"
}

func insightsListsRepositories(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}
	
	if len(data.Insights.Project.Repositories) > 0 {
		return layer4.Passed, "Insights contains a list of repositories"
	}

	return layer4.Failed, "Insights does NOT contains a list of repositories"
}

func statusChecksAreRequiredByRulesets(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
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

func statusChecksAreRequiredByBranchProtection(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
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

func noBinariesInRepo(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	// TODO: This only checks the top 3 levels of the repository tree
	// for common binary file extensions and it fails on very large repositories.
	suspectedBinaries, err := data.GetSuspectedBinaries()
	if err != nil {
		data.Config.Logger.Trace(fmt.Sprintf("unexpected response while checking for binaries: %s", err.Error()))
		return layer4.Unknown, "Error while scanning repository for binaries, potentially due to repo size. See logs for details."
	}

	if len(suspectedBinaries) == 0 {
		return layer4.Passed, "No common binary file extensions were found in the repository"
	}
	return layer4.Failed, fmt.Sprintf("Suspected binaries found in the repository: %s", strings.Join(suspectedBinaries, ", "))
}

func requiresNonAuthorApproval(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}
	protection := data.Repository.DefaultBranchRef.BranchProtectionRule

	if !protection.RequiresApprovingReviews {
		return layer4.Failed, "Branch protection rule does not require reviews"
	}

	reviewCount := data.Repository.DefaultBranchRef.RefUpdateRule.RequiredApprovingReviewCount
	if reviewCount < 1 {
		return layer4.Failed, "Branch protection rule requires 0 approving reviews"
	}

	if !protection.RequireLastPushApproval {
		return layer4.Failed, "Branch protection does not require re-approval after new commits"
	}

	return layer4.Passed, fmt.Sprintf("Branch protection requires %d approving reviews and re-approval after new commits", reviewCount)
}

func hasOneOrMoreStatusChecks(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
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

func verifyDependencyManagement(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
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
	// TODO: Do a quality check on the dependency manifests
	return countDependencyManifests(data)
}

func countDependencyManifests(payloadData any) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	manifestsCount := data.DependencyManifestsCount
	if manifestsCount > 0 {
		return layer4.Passed, fmt.Sprintf("Found %d dependency manifests from GitHub API", manifestsCount)
	}
	return layer4.Failed, "No dependency manifests found in the repository by the GitHub API"
}

func documentsTestExecution(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	_, message = reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	return layer4.NeedsReview, "Review project documentation to ensure it explains when and how tests are run"
}

func documentsTestMaintenancePolicy(payloadData any, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	_, message = reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}
	return layer4.NeedsReview, "Review project documentation to ensure it contains a clear policy for maintaining tests"
}
