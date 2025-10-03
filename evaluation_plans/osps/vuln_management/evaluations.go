package vuln_management

import (
	"github.com/ossf/gemara/layer4"
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
)

//
// Vulnerability Management Control Family

func OSPS_VM_01() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		ControlID: "OSPS-VM-01",
	}

	evaluation.AddAssessment(
		"OSPS-VM-01.01",
		"While active, the project documentation MUST include a policy for coordinated vulnerability reporting, with a clear timeframe for response.",
		[]string{
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.IsActive,
			reusable_steps.HasSecurityInsightsFile,
			hasVulnerabilityDisclosurePolicy,
		},
	)

	return
}

func OSPS_VM_02() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		ControlID: "OSPS-VM-02",
	}

	evaluation.AddAssessment(
		"OSPS-VM-02.01",
		"While active, the project documentation MUST contain security contacts.",
		[]string{
			"Maturity Level 1",
		},
		[]layer4.AssessmentStep{
			reusable_steps.IsCodeRepo,
			hasSecContact,
		},
	)

	return
}

func OSPS_VM_03() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		ControlID: "OSPS-VM-03",
	}

	evaluation.AddAssessment(
		"OSPS-VM-03.01",
		"While active, the project documentation MUST provide a means for reporting security vulnerabilities privately to the security contacts within the project.",
		[]string{
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.IsActive,
			reusable_steps.HasSecurityInsightsFile,
			hasPrivateVulnerabilityReporting,
		},
	)

	return
}

func OSPS_VM_04() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		ControlID: "OSPS-VM-04",
	}

	evaluation.AddAssessment(
		"OSPS-VM-04.01",
		"While active, the project documentation MUST publicly publish data about discovered vulnerabilities.",
		[]string{
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.IsActive,
			hasPublicVulnerabilityDisclosure,
		},
	)

	evaluation.AddAssessment(
		"OSPS-VM-04.02",
		"While active, any vulnerabilities in the software components not affecting the project MUST be accounted for in a VEX document, augmenting the vulnerability report with non-exploitability details.",
		[]string{
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.NotImplemented,
		},
	)

	return
}

func OSPS_VM_05() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		ControlID: "OSPS-VM-05",
	}

	evaluation.AddAssessment(
		"OSPS-VM-05.01",
		"While active, the project documentation MUST include a policy that defines a threshold for remediation of SCA findings related to vulnerabilities and licenses.",
		[]string{
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.NotImplemented,
		},
	)

	evaluation.AddAssessment(
		"OSPS-VM-05.02",
		"While active, the project documentation MUST include a policy to address SCA violations prior to any release.",
		[]string{
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.NotImplemented,
		},
	)

	evaluation.AddAssessment(
		"OSPS-VM-05.03",
		"While active, all changes to the project's codebase MUST be automatically evaluated against a documented policy for malicious dependencies and known vulnerabilities in dependencies, then blocked in the event of violations, except when declared and suppressed as non-exploitable.",
		[]string{
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.NotImplemented,
		},
	)

	return
}

func OSPS_VM_06() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		ControlID: "OSPS-VM-06",
	}

	evaluation.AddAssessment(
		"OSPS-VM-06.01",
		"While active, the project documentation MUST include a policy that defines a threshold for remediation of SAST findings.",
		[]string{
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.HasDependencyManagementPolicy,
		},
	)

	evaluation.AddAssessment(
		"OSPS-VM-06.02",
		"While active, all changes to the project's codebase MUST be automatically evaluated against a documented policy for security weaknesses and blocked in the event of violations except when declared and suppressed as non-exploitable.",
		[]string{
			"Maturity Level 3",
		},
		[]layer4.AssessmentStep{
			reusable_steps.IsCodeRepo,
			reusable_steps.HasSecurityInsightsFile,
			sastToolDefined,
		},
	)

	return
}
