package evaluation_plans

import (
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/osps/access_control"
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/osps/build_release"
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/osps/docs"
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/osps/governance"
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/osps/legal"
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/osps/quality"
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/osps/sec_assessment"
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/osps/vuln_management"

	"github.com/ossf/gemara/layer4"
)

var (
	// Open Source Project Security Baseline
	OSPS_B = []*layer4.ControlEvaluation{
		access_control.OSPS_AC_01(),
		access_control.OSPS_AC_02(),
		access_control.OSPS_AC_03(),
		access_control.OSPS_AC_04(),
		build_release.OSPS_BR_01(),
		build_release.OSPS_BR_02(),
		build_release.OSPS_BR_03(),
		build_release.OSPS_BR_04(),
		build_release.OSPS_BR_05(),
		build_release.OSPS_BR_06(),
		docs.OSPS_DO_01(),
		docs.OSPS_DO_02(),
		docs.OSPS_DO_03(),
		docs.OSPS_DO_04(),
		docs.OSPS_DO_05(),
		docs.OSPS_DO_06(),
		governance.OSPS_GV_01(),
		governance.OSPS_GV_02(),
		governance.OSPS_GV_03(),
		governance.OSPS_GV_04(),
		legal.OSPS_LE_01(),
		legal.OSPS_LE_02(),
		legal.OSPS_LE_03(),
		quality.OSPS_QA_01(),
		quality.OSPS_QA_02(),
		quality.OSPS_QA_03(),
		quality.OSPS_QA_04(),
		quality.OSPS_QA_05(),
		quality.OSPS_QA_06(),
		quality.OSPS_QA_07(),
		sec_assessment.OSPS_SA_01(),
		sec_assessment.OSPS_SA_02(),
		sec_assessment.OSPS_SA_03(),
		vuln_management.OSPS_VM_01(),
		vuln_management.OSPS_VM_02(),
		vuln_management.OSPS_VM_03(),
		vuln_management.OSPS_VM_04(),
		vuln_management.OSPS_VM_05(),
		vuln_management.OSPS_VM_06(),
	}
)
