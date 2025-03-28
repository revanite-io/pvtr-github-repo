package governance

import (
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
	"github.com/revanite-io/sci/pkg/layer4"
)

//
// Governance Control Family

func OSPS_GV_01() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		Control_Id:        "OSPS-GV-01",
		Remediation_Guide: "",
	}

	evaluation.AddAssessment(
		"OSPS-GV-01.01",
		"While active, the project documentation MUST include a list of project members with access to sensitive resources.",
		[]string{
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.HasSecurityInsightsFile,
			reusable_steps.IsActive,
			coreTeamIsListed,
			projectAdminsListed,
		},
	)

	evaluation.AddAssessment(
		"OSPS-GV-01.02",
		"While active, the project documentation MUST include descriptions of the roles and responsibilities for members of the project.",
		[]string{
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			hasRolesAndResponsibilities,
		},
	)

	return
}

func OSPS_GV_02() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		Control_Id:        "OSPS-GV-02",
		Remediation_Guide: "",
	}

	evaluation.AddAssessment(
		"OSPS-GV-02.01",
		"While active, the project MUST have one or more mechanisms for public discussions about proposed changes and usage obstacles.",
		[]string{
			"Maturity Level 1",
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.HasIssuesOrDiscussionsEnabled,
		},
	)

	return
}

func OSPS_GV_03() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		Control_Id:        "OSPS-GV-03",
		Remediation_Guide: "",
	}

	evaluation.AddAssessment(
		"OSPS-GV-03.01",
		"While active, the project documentation MUST include an explanation of the contribution process.",
		[]string{
			"Maturity Level 1",
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			hasContributionGuide,
		},
	)

	evaluation.AddAssessment(
		"OSPS-GV-03.02",
		"While active, the project documentation MUST include a guide for code contributors that includes requirements for acceptable contributions.",
		[]string{
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.HasSecurityInsightsFile,
			reusable_steps.IsActive,
			hasContributionReviewPolicy,
		},
	)

	return
}

func OSPS_GV_04() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		Control_Id:        "OSPS-GV-04",
		Remediation_Guide: "",
	}

	evaluation.AddAssessment(
		"OSPS-GV-04.01",
		"While active, the project documentation MUST have a policy that code contributors are reviewed prior to granting escalated permissions to sensitive resources.",
		[]string{
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.NotImplemented,
		},
	)

	return
}
