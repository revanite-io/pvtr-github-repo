package build_release

import (
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
	"github.com/revanite-io/sci/pkg/layer4"
)

//
// Build and Release Control Family

func OSPS_BR_01() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		Control_Id:        "OSPS-BR-01",
		Remediation_Guide: "",
	}

	evaluation.AddAssessment(
		"OSPS-BR-01.01",
		"When a CI/CD pipeline accepts an input parameter, that parameter MUST be sanitized and validated prior to use in the pipeline.",
		[]string{
			"Maturity Level 1",
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			cicdSanitizedInputParameters,
		},
	)

	evaluation.AddAssessment(
		"OSPS-BR-01.02",
		"When a CI/CD pipeline uses a branch name in its functionality, that name value MUST be sanitized and validated prior to use in the pipeline.",
		[]string{
			"Maturity Level 1",
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.NotImplemented,
		},
	)

	return
}

func OSPS_BR_02() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		Control_Id:        "OSPS-BR-02",
		Remediation_Guide: "",
	}

	evaluation.AddAssessment(
		"OSPS-BR-02.01",
		"When an official release is created, that release MUST be assigned a unique version identifier.",
		[]string{
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.HasMadeReleases,
			releaseHasUniqueIdentifier,
		},
	)

	// Just run the previous assessments for now
	// TODO: Implement this assessment
	// evaluation.AddAssessment(
	// 	"OSPS-BR-02.02",
	// 	"When an official release is created, all assets within that release MUST be clearly associated with the release identifier or another unique identifier for the asset.",
	// 	[]string{
	// 		"Maturity Level 3",
	// 	},
	// 	[]layer4.AssessmentStep{
	// 		reusable_steps.NotImplemented,
	// 	},
	// )

	return
}

func OSPS_BR_03() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		Control_Id:        "OSPS-BR-03",
		Remediation_Guide: "",
	}

	evaluation.AddAssessment(
		"OSPS-BR-03.01",
		"When the project lists a URI as an official project channel, that URI MUST be exclusively delivered using encrypted channels.",
		[]string{
			"Maturity Level 1",
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.HasSecurityInsightsFile,
			ensureInsightsLinksUseHTTPS,
		},
	)

	evaluation.AddAssessment(
		"OSPS-BR-03.02",
		"When the project lists a URI as an official distribution channel, that URI MUST be exclusively delivered using encrypted channels.",
		[]string{
			"Maturity Level 1",
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			distributionPointsUseHTTPS,
		},
	)

	return
}

func OSPS_BR_04() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		Control_Id:        "OSPS-BR-04",
		Remediation_Guide: "",
	}

	evaluation.AddAssessment(
		"OSPS-BR-04.01",
		"When an official release is created, that release MUST contain a descriptive log of functional and security modifications.",
		[]string{
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.HasMadeReleases,
			ensureLatestReleaseHasChangelog,
		},
	)

	return
}

func OSPS_BR_05() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		Control_Id:        "OSPS-BR-05",
		Remediation_Guide: "",
	}

	evaluation.AddAssessment(
		"OSPS-BR-05.01",
		"When a build and release pipeline ingests dependencies, it MUST use standardized tooling where available.",
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

func OSPS_BR_06() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		Control_Id:        "OSPS-BR-06",
		Remediation_Guide: "",
	}

	evaluation.AddAssessment(
		"OSPS-BR-06.01",
		"When an official release is created, that release MUST be signed or accounted for in a signed manifest including each asset's cryptographic hashes.",
		[]string{
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.HasMadeReleases,
			reusable_steps.HasSecurityInsightsFile,
			insightsHasSlsaAttestation,
		},
	)

	return
}
