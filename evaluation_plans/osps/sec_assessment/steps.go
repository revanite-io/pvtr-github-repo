package sec_assessment

import (
	"strings"

	"github.com/gemaraproj/go-gemara"
	"github.com/ossf/pvtr-github-repo-scanner/data"
	"github.com/ossf/pvtr-github-repo-scanner/evaluation_plans/reusable_steps"
)

// DesignDocFiles are common file names for design/architecture documentation
var DesignDocFiles = []string{
	"architecture.md",
	"design.md",
	"architecture.rst",
	"design.rst",
	"architecture.txt",
	"design.txt",
}

// DesignDocDirectories are common directory names that typically contain design documentation
var DesignDocDirectories = []string{
	"adr",
	"adrs",
	"architecture",
	"design",
	"docs",
	"doc",
}

func HasDesignDocumentation(payload data.Payload) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	var foundDirectories []string

	// Check for design documentation files and directories in repository root
	if payload.GraphqlRepoData != nil {
		for _, entry := range payload.Repository.Object.Tree.Entries {
			// Check for design doc files (blobs only)
			if entry.Type == "blob" {
				for _, designFile := range DesignDocFiles {
					if strings.EqualFold(entry.Name, designFile) {
						return gemara.Passed, "Design documentation found: " + entry.Name, confidence
					}
				}
			}

			// Check for directories that typically contain design documentation
			if entry.Type == "tree" {
				for _, designDir := range DesignDocDirectories {
					if strings.EqualFold(entry.Name, designDir) {
						foundDirectories = append(foundDirectories, entry.Name)
					}
				}
			}
		}
	}

	// If we found directories that typically contain design docs, flag for manual review
	if len(foundDirectories) > 0 {
		return gemara.NeedsReview, "No design documentation file found in root, but found directories that may contain design documentation: " + strings.Join(foundDirectories, ", ") + " - manual review needed", confidence
	}

	// Fallback: check if DetailedGuide is specified in Security Insights
	if payload.RestData != nil && payload.Insights.Project.Documentation.DetailedGuide != nil {
		return gemara.NeedsReview, "No design documentation file found, but detailed guide specified in Security Insights - manual review needed to confirm design documentation with actions and actors", confidence
	}

	return gemara.Failed, "Design documentation demonstrating all actions and actors was NOT found", confidence
}

// InterfaceDocFiles are common file names for external interface / API documentation.
var InterfaceDocFiles = []string{
	"api.md",
	"api.rst",
	"api.txt",
	"api.yaml",
	"api.yml",
	"api.json",
	"apidocs.md",
	"api-reference.md",
	"api-reference.rst",
	"openapi.yaml",
	"openapi.yml",
	"openapi.json",
	"swagger.yaml",
	"swagger.yml",
	"swagger.json",
}

// InterfaceDocDirectories are common directory names that typically contain
// external interface / API documentation.
var InterfaceDocDirectories = []string{
	"api",
	"apis",
	"apidocs",
	"api-docs",
	"documentation",
	"reference",
	"references",
	"docs",
	"doc",
	"spec",
	"specs",
	"openapi",
	"swagger",
	"schema",
	"proto",
}

// HasExternalInterfaceDocumentation assesses OSPS-SA-02.01: when the project has
// made a release, its documentation must describe all external software
// interfaces (APIs) of the released assets.
func HasExternalInterfaceDocumentation(payload data.Payload) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	// The requirement only applies once a release exists.
	released, observable := reusable_steps.HasPublishedRelease(payload)
	if !observable {
		return gemara.NeedsReview, "Release data is unavailable; manually review whether the documentation describes all external software interfaces", gemara.Low
	}
	if !released {
		return gemara.NotApplicable, "No published releases found; the external interface documentation requirement does not apply", gemara.High
	}

	var foundDirectories []string

	if payload.GraphqlRepoData != nil {
		for _, entry := range payload.Repository.Object.Tree.Entries {
			if entry.Type == "blob" {
				for _, docFile := range InterfaceDocFiles {
					if strings.EqualFold(entry.Name, docFile) {
						// A matching root filename indicates interface docs likely
						// exist, but does not prove they cover every external
						// interface of the released assets.
						return gemara.NeedsReview, "External interface documentation found (" + entry.Name + "), but coverage of all external interfaces requires manual review", gemara.Low
					}
				}
			}

			if entry.Type == "tree" {
				for _, docDir := range InterfaceDocDirectories {
					if strings.EqualFold(entry.Name, docDir) {
						foundDirectories = append(foundDirectories, entry.Name)
					}
				}
			}
		}
	}

	// A directory that typically holds API docs is a weaker signal that still
	// cannot be confirmed to document every interface.
	if len(foundDirectories) > 0 {
		return gemara.NeedsReview, "No external interface documentation file found in root, but found directories that may contain API documentation: " + strings.Join(foundDirectories, ", ") + " - manual review needed to confirm all external interfaces are documented", gemara.Low
	}

	// Fallback: a detailed or quickstart guide in Security Insights may describe
	// the interfaces, but this cannot be verified automatically.
	if payload.RestData != nil && payload.Insights.Project != nil && payload.Insights.Project.Documentation != nil {
		if payload.Insights.Project.Documentation.DetailedGuide != nil {
			return gemara.NeedsReview, "No external interface documentation file or directory found, but detailed guide specified in Security Insights - manual review needed to confirm all external interfaces are documented", gemara.Low
		}
		if payload.Insights.Project.Documentation.QuickstartGuide != nil {
			return gemara.NeedsReview, "No external interface documentation file or directory found, but quickstart guide specified in Security Insights - manual review needed to confirm all external interfaces are documented", gemara.Low
		}
	}

	// No interface-doc file, API-doc directory, or Security Insights guide was
	// found for a released project, so the MUST requirement is unmet.
	return gemara.Failed, "No documentation file, API-documentation directory, or Security Insights guide describing the external software interfaces of released assets was found", gemara.Medium
}
