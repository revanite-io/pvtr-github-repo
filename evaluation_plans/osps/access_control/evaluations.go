package access_control

import (
	"github.com/ossf/gemara/layer4"
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
)

//
// Access Control Control Family

func OSPS_AC_01() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		ControlID:        "OSPS-AC-01",
		RemediationGuide: "",
	}

	evaluation.AddAssessment(
		"OSPS-AC-01.01",
		"When a user attempts to access a sensitive resource in the project's version control system, the system MUST require the user to complete a multi-factor authentication process.",
		[]string{
			"Maturity Level 1",
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			orgRequiresMFA,
		},
	)

	return
}

func OSPS_AC_02() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		ControlID:        "OSPS-AC-02",
		RemediationGuide: "",
	}

	evaluation.AddAssessment(
		"OSPS-AC-02.01",
		"When a new collaborator is added, the version control system MUST require manual permission assignment, or restrict the collaborator permissions to the lowest available privileges by default.",
		[]string{
			"Maturity Level 1",
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.GithubBuiltIn, // This control is enforced by GitHub for all projects
		},
	)

	return
}

func OSPS_AC_03() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		ControlID:        "OSPS-AC-03",
		RemediationGuide: "",
	}

	evaluation.AddAssessment(
		"OSPS-AC-03.01",
		"When a direct commit is attempted on the project's primary branch, an enforcement mechanism MUST prevent the change from being applied.",
		[]string{
			"Maturity Level 1",
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.IsCodeRepo,
			branchProtectionRestrictsPushes, // This checks branch protection, but not rulesets yet
		},
	)

	evaluation.AddAssessment(
		"OSPS-AC-03.02",
		"When an attempt is made to delete the project's primary branch, the version control system MUST treat this as a sensitive activity and require explicit confirmation of intent.",
		[]string{
			"Maturity Level 1",
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			branchProtectionPreventsDeletion, // This checks branch protection, but not rulesets yet
		},
	)

	return
}

func OSPS_AC_04() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		ControlID:        "OSPS-AC-04",
		RemediationGuide: "",
	}

	evaluation.AddAssessment(
		"OSPS-AC-04.01",
		"When a CI/CD task is executed with no permissions specified, the project's version control system MUST default to the lowest available permissions for all activities in the pipeline.",
		[]string{
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			workflowDefaultReadPermissions,
		},
	)

	// Just run the previous assessments for now
	// TODO: Implement this assessment
	// evaluation.AddAssessment(
	// 	"OSPS-AC-04.02",
	// 	"When a job is assigned permissions in a CI/CD pipeline, the source code or configuration MUST only assign the minimum privileges necessary for the corresponding activity.",
	// 	[]string{
	// 		"Maturity Level 3",
	// 	},
	// 	[]layer4.AssessmentStep{
	// 		reusable_steps.NotImplemented,
	// 	},
	// )

	return
}
