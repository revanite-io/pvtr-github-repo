package sec_assessment

import (
	"github.com/ossf/gemara/layer4"
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
)

//
// Security Assessment Control Family

func OSPS_SA_01() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		ControlID:        "OSPS-SA-01",
		RemediationGuide: "",
	}

	evaluation.AddAssessment(
		"OSPS-SA-01.01",
		"When the project has made a release, the project documentation MUST include design documentation demonstrating all actions and actors within the system.",
		[]string{
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.NotImplemented,
		},
	)

	return
}

func OSPS_SA_02() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		ControlID:        "OSPS-SA-02",
		RemediationGuide: "",
	}

	evaluation.AddAssessment(
		"OSPS-SA-02.01",
		"When the project has made a release, the project documentation MUST include descriptions of all external software interfaces of the released software assets.",
		[]string{
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.NotImplemented,
		},
	)

	return
}

func OSPS_SA_03() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		ControlID:        "OSPS-SA-03",
		RemediationGuide: "",
	}

	evaluation.AddAssessment(
		"OSPS-SA-03.01",
		"When the project has made a release, the project MUST perform a security assessment to understand the most likely and impactful potential security problems that could occur within the software.",
		[]string{
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.NotImplemented,
		},
	)

	evaluation.AddAssessment(
		"OSPS-SA-03.02",
		"When the project has made a release, the project MUST perform a threat modeling and attack surface analysis to understand and protect against attacks on critical code paths, functions, and interactions within the system.",
		[]string{
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.NotImplemented,
		},
	)

	return
}
