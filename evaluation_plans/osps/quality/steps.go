package quality

import (
	"fmt"
	"strings"

	"github.com/gemaraproj/go-gemara"
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
)

func RepoIsPublic(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, 0 // TODO
	}
	if data.RepositoryMetadata.IsPublic() {
		return gemara.Passed, "Repository is public", 0 // TODO
	}
	return gemara.Failed, "Repository is private", 0 // TODO
}

func InsightsListsRepositories(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, 0 // TODO
	}

	if len(data.Insights.Project.Repositories) > 0 {
		return gemara.Passed, "Insights contains a list of repositories", 0 // TODO
	}

	return gemara.Failed, "Insights does not contain a list of repositories", 0 // TODO
}

func StatusChecksAreRequiredByRulesets(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, 0 // TODO
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
		return gemara.Passed, "No rulesets found for default branch, continuing to evaluate branch protection", 0 // TODO
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
		return gemara.Failed,
			fmt.Sprintf(
				"Some executed status checks are not mandatory but all should be: %s (NOTE: Not continuing to evaluate branch protection: combining requirements in rulesets and branch protection is not recommended)",
				strings.Join(missingChecks, ", ")),
			0 // TODO
	}

	return gemara.Passed, "No status checks were run that are not required by the rules", 0 // TODO
}

func StatusChecksAreRequiredByBranchProtection(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, 0 // TODO
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
		return gemara.Failed,
			fmt.Sprintf("Some executed status checks are not mandatory but all should be: %s",
				strings.Join(missingChecks, ", ")),
			0 // TODO
	}

	return gemara.Passed, "No status checks were run that are not required by branch protection", 0 // TODO
}

func NoBinariesInRepo(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, 0 // TODO
	}

	// TODO: This only checks the top 3 levels of the repository tree
	// for common binary file extensions and it fails on very large repositories.
	suspectedBinaries, err := data.GetSuspectedBinaries()
	if err != nil {
		data.Config.Logger.Trace(fmt.Sprintf("unexpected response while checking for binaries: %s", err.Error()))
		return gemara.Unknown,
			"Error while scanning repository for binaries, potentially due to repo size. See logs for details.",
			0 // TODO
	}

	if len(suspectedBinaries) == 0 {
		return gemara.Passed, "No common binary file extensions were found in the repository", 0 // TODO
	}
	return gemara.Failed,
		fmt.Sprintf("Suspected binaries found in the repository: %s", strings.Join(suspectedBinaries, ", ")),
		0 // TODO
}

func RequiresNonAuthorApproval(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, 0 // TODO
	}
	protection := data.Repository.DefaultBranchRef.BranchProtectionRule

	if !protection.RequiresApprovingReviews {
		return gemara.Failed, "Branch protection rule does not require reviews", 0 // TODO
	}

	reviewCount := data.Repository.DefaultBranchRef.RefUpdateRule.RequiredApprovingReviewCount
	if reviewCount < 1 {
		return gemara.Failed, "Branch protection rule requires 0 approving reviews", 0 // TODO
	}

	if !protection.RequireLastPushApproval {
		return gemara.Failed, "Branch protection does not require re-approval after new commits", 0 // TODO
	}

	return gemara.Passed,
		fmt.Sprintf("Branch protection requires %d approving reviews and re-approval after new commits", reviewCount),
		0 // TODO
}

func HasOneOrMoreStatusChecks(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, 0 // TODO
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
		return gemara.Passed, fmt.Sprintf("%d status checks were run", len(statusChecks)), 0 // TODO
	}

	return gemara.Failed, "No status checks were run", 0 // TODO
}

func VerifyDependencyManagement(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, 0 // TODO
	}

	// Validate required fields
	if data.Repository.Name == "" || data.Repository.DefaultBranchRef.Name == "" ||
		data.Repository.DefaultBranchRef.Target.OID == "" {
		return gemara.Unknown, "Missing required repository data", 0 // TODO
	}

	// Check dependency manifests
	// TODO: Do a quality check on the dependency manifests
	return countDependencyManifests(data)
}

func countDependencyManifests(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, 0 // TODO
	}

	manifestsCount := data.DependencyManifestsCount
	if manifestsCount > 0 {
		return gemara.Passed, fmt.Sprintf("Found %d dependency manifests from GitHub API", manifestsCount), 0 // TODO
	}
	return gemara.NeedsReview, "No dependency manifests found in the GitHub dependency graph API. Review project to ensure dependencies are managed.", 0 // TODO
}

func DocumentsTestExecution(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	_, message = reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, 0 // TODO
	}

	return gemara.NeedsReview, "Review project documentation to ensure it explains when and how tests are run", 0 // TODO
}

func DocumentsTestMaintenancePolicy(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	_, message = reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, 0 // TODO
	}
	return gemara.NeedsReview, "Review project documentation to ensure it contains a clear policy for maintaining tests", 0 // TODO
}
