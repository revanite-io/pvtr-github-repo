package quality

import (
	"fmt"
	"strings"

	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
	"github.com/revanite-io/sci/pkg/layer4"
)

func repoIsPublic(payloadData interface{}, _ map[string]*layer4.Change) (result layer4.Result, message string) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return layer4.Unknown, message
	}

	if !data.Repository.IsPrivate {
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
