package data

import "github.com/privateerproj/privateer-sdk/config"

// SupportedCatalogIDs is the declared compatibility contract for catalog IDs
// that must continue to exist in bundled catalog data.
var SupportedCatalogIDs = []string{
	"osps-baseline",
	"osps-baseline-2025-10",
	"osps-baseline-2026-02",
	"osps-baseline-2026-08",
}

var maturityLevelAliases = map[string]string{
	"Maturity Level 1": "maturity-1",
	"maturity-1":       "maturity-1",
	"Maturity Level 2": "maturity-2",
	"maturity-2":       "maturity-2",
	"Maturity Level 3": "maturity-3",
	"maturity-3":       "maturity-3",
}

// NormalizeApplicability resolves the change in applicability level ids
// across catalogs and catalog versions. Our copies of the data are modified
// from the original to use a single consistent applicability set.
func NormalizeApplicability(c *config.Config) {
	for i, level := range c.Policy.Applicability {
		if alias, ok := maturityLevelAliases[level]; ok {
			c.Policy.Applicability[i] = alias
		}
	}
}
