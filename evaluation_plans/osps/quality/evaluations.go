package quality

import (
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
	"github.com/revanite-io/sci/pkg/layer4"
)

//
// Quality Control Family

func OSPS_QA_01() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		Control_Id:        "OSPS-QA-01",
		Remediation_Guide: "",
	}

	evaluation.AddAssessment(
		"OSPS-QA-01.01",
		"While active, the project's source code repository MUST be publicly readable at a static URL.",
		[]string{
			"Maturity Level 1",
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			repoIsPublic,
		},
	)

	evaluation.AddAssessment(
		"OSPS-QA-01.02",
		"The version control system MUST contain a publicly readable record of all changes made, who made the changes, and when the changes were made.",
		[]string{
			"Maturity Level 1",
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.GithubBuiltIn,
		},
	)

	return
}

func OSPS_QA_02() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		Control_Id:        "OSPS-QA-02",
		Remediation_Guide: "",
	}

	evaluation.AddAssessment(
		"OSPS-QA-02.01",
		"When the package management system supports it, the source code repository MUST contain a dependency list that accounts for the direct language dependencies.",
		[]string{
			"Maturity Level 1",
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.NotImplemented,
		},
	)

	evaluation.AddAssessment(
		"OSPS-QA-02.02",
		"When the project has made a release, all compiled released software assets MUST be delivered with a software bill of materials.",
		[]string{
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.NotImplemented,
		},
	)

	return
}

func OSPS_QA_03() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		Control_Id:        "OSPS-QA-03",
		Remediation_Guide: "",
	}

	evaluation.AddAssessment(
		"OSPS-QA-03.01",
		"When a commit is made to the primary branch, any automated status checks for commits MUST pass or be manually bypassed.",
		[]string{
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			statusChecksAreRequiredByRulesets,
			statusChecksAreRequiredByBranchProtection,
		},
	)

	return
}

func OSPS_QA_04() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		Control_Id:        "OSPS-QA-04",
		Remediation_Guide: "",
	}

	evaluation.AddAssessment(
		"OSPS-QA-04.01",
		"While active, the project documentation MUST contain a list of any codebases that are considered subprojects or additional repositories.",
		[]string{
			"Maturity Level 1",
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			insightsListsRepositories,
		},
	)

	// Just run the previous assessments for now
	// TODO: Implement this assessment
	// evaluation.AddAssessment(
	// 	"OSPS-QA-04.02",
	// 	"When the project has made a release comprising multiple source code repositories, all subprojects MUST enforce security requirements that are as strict or stricter than the primary codebase.",
	// 	[]string{
	// 		"Maturity Level 3",
	// 	},
	// 	[]layer4.AssessmentStep{
	// 		reusable_steps.NotImplemented,
	// 	},
	// )

	return
}

func OSPS_QA_05() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		Control_Id:        "OSPS-QA-05",
		Remediation_Guide: "",
	}

	evaluation.AddAssessment(
		"OSPS-QA-05.01",
		"While active, the version control system MUST NOT contain generated executable artifacts.",
		[]string{
			"Maturity Level 1",
			"Maturity Level 2",
			"Maturity Level 3",
		},
		// TODO: Figure out how to walk the repo tree and check for isBinary
		[]layer4.AssessmentStep{
			noBinariesInRepo,
		},
	)

	return
}

func OSPS_QA_06() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		Control_Id:        "OSPS-QA-06",
		Remediation_Guide: "",
	}

	evaluation.AddAssessment(
		"OSPS-QA-06.01",
		"Prior to a commit being accepted, the project's CI/CD pipelines MUST run at least one automated test suite to ensure the changes meet expectations.",
		[]string{
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			hasOneOrMoreStatusChecks,
		},
	)

	evaluation.AddAssessment(
		"OSPS-QA-06.02",
		"While active, project's documentation MUST clearly document when and how tests are run.",
		[]string{
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.NotImplemented,
		},
	)

	evaluation.AddAssessment(
		"OSPS-QA-06.03",
		"While active, the project's documentation MUST include a policy that all major changes to the software produced by the project should add or update tests of the functionality in an automated test suite.",
		[]string{
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.NotImplemented,
		},
	)

	return
}

func OSPS_QA_07() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		Control_Id:        "OSPS-QA-07",
		Remediation_Guide: "",
	}

	evaluation.AddAssessment(
		"OSPS-QA-07.01",
		"When a commit is made to the primary branch, the project's version control system MUST require at least one non-author approval of the changes before merging.",
		[]string{
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			requiresNonAuthorApproval,
		},
	)

	return
}
