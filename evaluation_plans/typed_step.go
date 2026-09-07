package evaluation_plans

import (
	"github.com/gemaraproj/go-gemara"
	"github.com/ossf/pvtr-github-repo-scanner/data"
)

// TypedStep is the signature every step in this plugin uses: it receives a
// fully-typed data.Payload instead of an untyped any. The SDK's typed
// registration helpers (pluginkit.AddEvaluationSuiteTyped and
// ...TypedForAllCatalogs) adapt it to gemara.AssessmentStep and perform the
// payload type assertion once, so steps need no payload guard of their own.
//
// Always register through those helpers. Wrapping a TypedStep in a local
// closure erases its function name from the benchmark report and evaluation
// log; see pluginkit.FuncName for why.
type TypedStep func(data.Payload) (gemara.Result, string, gemara.ConfidenceLevel)

// AllSteps merges every step map in this package into one map, keyed by
// assessment ID, for registration against all loaded catalogs.
//
// One merged map is safe to register against every catalog: the SDK looks up
// steps by the requirement IDs a catalog declares, so IDs it does not declare
// are never run. It is also safe across OSPS Baseline versions, because the
// maintenance policy
// (https://github.com/ossf/security-baseline/blob/main/docs/maintenance.md#identifiers)
// gives a control a new identifier whenever its meaning changes substantially,
// so one implementation per ID holds for every version.
//
// Every map merged here must take data.Payload; steps that consume a different
// payload type need their own registration call.
func AllSteps() map[string][]TypedStep {
	merged := make(map[string][]TypedStep, len(OSPS))
	for id, steps := range OSPS {
		merged[id] = append(merged[id], steps...)
	}
	// Merge additional step maps here.
	return merged
}
