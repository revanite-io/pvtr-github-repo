package data

import (
	"path"
	"regexp"
	"strings"
)

// vexTokenPattern matches "vex" as a delimited token (its own path/filename
// segment, or bounded by common separators) so that VEX documents are detected
// without matching unrelated words that merely contain the letters (e.g.
// "convex", "vexing"). Case-insensitive.
var vexTokenPattern = regexp.MustCompile(`(?i)(^|[/._\-])vex([/._\-]|$)`)

// structuredDataExtensions are the file extensions a VEX document is normally
// published in. A bare "vex" directory segment is enough on its own, but a file
// only counts when it is one of these machine-readable formats.
var structuredDataExtensions = map[string]bool{
	".json": true,
	".yaml": true,
	".yml":  true,
	".xml":  true,
	".csaf": true,
}

// isVexPath reports whether a repository path looks like a VEX document. It
// recognises the common conventions — an OpenVEX/CSAF/CycloneDX-VEX file
// (e.g. product.vex.json, openvex.json, bom.cdx-vex.json, advisory.csaf.json)
// or any structured-data file living under a vex/ directory — while avoiding
// false positives from words that merely contain "vex".
func isVexPath(p string) bool {
	lower := strings.ToLower(p)
	base := path.Base(lower)
	dir := path.Dir(lower)
	ext := path.Ext(base)

	// A structured-data file inside a vex/ directory segment.
	if structuredDataExtensions[ext] && vexTokenPattern.MatchString("/"+dir+"/") {
		return true
	}

	// CSAF advisories are frequently VEX; treat a .csaf.json/.csaf file as VEX.
	if strings.HasSuffix(base, ".csaf.json") || ext == ".csaf" {
		return true
	}

	// CycloneDX VEX naming, e.g. bom.cdx-vex.json or app.cyclonedx-vex.json, and
	// OpenVEX naming, e.g. openvex.json or product.openvex.json. Gated on a
	// structured-data extension so VEX *tooling* source/docs (openvex.go,
	// cyclonedx-vex.md) are not mistaken for a published VEX document.
	if structuredDataExtensions[ext] &&
		(strings.Contains(base, "cdx-vex") || strings.Contains(base, "cyclonedx-vex") || strings.Contains(base, "openvex")) {
		return true
	}

	// A file whose name carries a delimited "vex" token in a structured format,
	// e.g. product.vex.json, openvex.json, my-vex.yaml.
	if structuredDataExtensions[ext] && vexTokenPattern.MatchString(base) {
		return true
	}

	return false
}

// detectVexDocuments walks the already-fetched repository tree and returns the
// repo-root-relative paths of files that look like VEX documents. It reuses the
// tree gathered for binary analysis, so it adds no extra API calls. Paths (not
// bare names) are collected so the OSPS-VM-04.02 evidence message is unambiguous
// when a VEX file lives in a subdirectory (e.g. security/vex/bom.json).
func detectVexDocuments(tree *GraphqlRepoTree) []string {
	if tree == nil {
		return nil
	}
	var found []string
	_, _ = walkTree(tree, func(_ *bool, _ bool, p string, _ int) (bool, error) {
		if isVexPath(p) {
			found = append(found, p)
		}
		// Return false so walkTree does not also collect the bare name; this
		// closure captures the full path itself.
		return false, nil
	})
	return found
}
