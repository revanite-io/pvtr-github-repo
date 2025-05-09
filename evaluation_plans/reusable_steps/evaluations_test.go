package reusable_steps

import (
	"testing"

	"github.com/ossf/si-tooling/v2/si"
	"github.com/revanite-io/pvtr-github-repo/data"
	"github.com/revanite-io/sci/pkg/layer4"
	"github.com/stretchr/testify/assert"
)

type testingData struct {
	expectedResult    layer4.Result
	expectedMessage   string
	repoDocumentation *si.RepositoryDocumentation
	name              string
	assertionMessage  string
}

func TestHasDependencyManagementPolicySomethin(t *testing.T) {
	depManagement := si.NewURL("https://example.com/dependency-management")
	emptyDepManagement := si.NewURL("")
	nilRepoDocumentation := (*si.RepositoryDocumentation)(nil)

	payload := data.Payload{
		RestData: &data.RestData{
			Insights: si.SecurityInsights{
				Repository: &si.Repository{},
			},
		},
	}

	testData := []testingData{
		{
			expectedResult:  layer4.Passed,
			expectedMessage: "Found dependency management policy in documentation",
			repoDocumentation: &si.RepositoryDocumentation{
				DependencyManagementPolicy: &depManagement,
			},
			name: "Dependency management policy found when present",
		},
		{
			expectedResult:  layer4.Failed,
			expectedMessage: "No dependency management file found",
			repoDocumentation: &si.RepositoryDocumentation{
				DependencyManagementPolicy: &emptyDepManagement,
			},
			name:             "fail when policy is empty",
			assertionMessage: "Empty string check failed",
		},
		{
			expectedResult:    layer4.Failed,
			expectedMessage:   "No dependency management file found",
			repoDocumentation: nilRepoDocumentation,
			assertionMessage:  "Null String check failed",
		},
	}
	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			payload.Insights.Repository.Documentation = test.repoDocumentation
			result, message := HasDependencyManagementPolicy(payload, nil)
			assert.Equal(t, test.expectedResult, result, test.assertionMessage)
			assert.Equal(t, test.expectedMessage, message, test.assertionMessage)
		})
	}
}
