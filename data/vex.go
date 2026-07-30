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
	// OpenVEX naming, e.g. openvex.json or product.openvex.json.
	if strings.Contains(base, "cdx-vex") || strings.Contains(base, "cyclonedx-vex") || strings.Contains(base, "openvex") {
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
// names of files that look like VEX documents. It reuses the tree gathered for
// binary analysis, so it adds no extra API calls. OSPS-VM-04.02 uses the result
// to report whether the project publishes a VEX feed.
func detectVexDocuments(tree *GraphqlRepoTree) []string {
	if tree == nil {
		return nil
	}
	found, err := walkTree(tree, func(_ *bool, _ bool, p string, _ int) (bool, error) {
		return isVexPath(p), nil
	})
	if err != nil {
		return nil
	}
	return found
}
