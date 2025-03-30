package legal

import (
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
	"github.com/revanite-io/sci/pkg/layer4"
)

//
// Legal Control Family

func OSPS_LE_01() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		Control_Id:        "OSPS-LE-01",
		Remediation_Guide: "",
	}

	evaluation.AddAssessment(
		"OSPS-LE-01.01",
		"While active, the version control system MUST require all code contributors to assert that they are legally authorized to make the associated contributions on every commit.",
		[]string{
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.GithubBuiltIn,
		},
	)

	return
}

func OSPS_LE_02() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		Control_Id:        "OSPS-LE-02",
		Remediation_Guide: "",
	}

	evaluation.AddAssessment(
		"OSPS-LE-02.01",
		"While active, the license for the source code MUST meet the OSI Open Source Definition or the FSF Free Software Definition.",
		[]string{
			"Maturity Level 1",
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			foundLicense,
			goodLicense,
		},
	)

	// Just run the previous assessments for now
	// TODO: Implement this assessment
	// evaluation.AddAssessment(
	// 	"OSPS-LE-02.02",
	// 	"While active, the license for the released software assets MUST meet the OSI Open Source Definition or the FSF Free Software Definition.",
	// 	[]string{
	// 		"Maturity Level 1",
	// 		"Maturity Level 2",
	// 		"Maturity Level 3",
	// 	},
	// 	[]layer4.AssessmentStep{
	// 		reusable_steps.HasSecurityInsightsFile,
	// 		reusable_steps.IsActive,
	// 		reusable_steps.NotImplemented,
	// 	},
	// )

	return
}

func OSPS_LE_03() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		Control_Id:        "OSPS-LE-03",
		Remediation_Guide: "",
	}

	evaluation.AddAssessment(
		"OSPS-LE-03.01",
		"While active, the license for the source code MUST be maintained in the corresponding repository's LICENSE file, COPYING file, or LICENSE/ directory.",
		[]string{
			"Maturity Level 1",
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			foundLicense,
		},
	)

	// Just run the previous assessments for now
	// TODO: Implement this assessment
	// evaluation.AddAssessment(
	// 	"OSPS-LE-03.02",
	// 	"While active, the license for the released software assets MUST be included in the released source code, or in a LICENSE file, COPYING file, or LICENSE/ directory alongside the corresponding release assets.",
	// 	[]string{
	// 		"Maturity Level 1",
	// 		"Maturity Level 2",
	// 		"Maturity Level 3",
	// 	},
	// 	[]layer4.AssessmentStep{
	// 		reusable_steps.HasSecurityInsightsFile,
	// 		reusable_steps.IsActive,
	// 		reusable_steps.NotImplemented,
	// 	},
	// )

	return
}
