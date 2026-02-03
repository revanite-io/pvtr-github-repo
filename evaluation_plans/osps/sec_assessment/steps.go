package sec_assessment

import (
	"strings"

	"github.com/gemaraproj/go-gemara"

	"github.com/revanite-io/pvtr-github-repo/evaluation_plans/reusable_steps"
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

func HasDesignDocumentation(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	data, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message
	}

	var foundDirectories []string

	// Check for design documentation files and directories in repository root
	if data.GraphqlRepoData != nil {
		for _, entry := range data.Repository.Object.Tree.Entries {
			// Check for design doc files (blobs only)
			if entry.Type == "blob" {
				for _, designFile := range DesignDocFiles {
					if strings.EqualFold(entry.Name, designFile) {
						return gemara.Passed, "Design documentation found: " + entry.Name
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
		return gemara.NeedsReview, "No design documentation file found in root, but found directories that may contain design documentation: " + strings.Join(foundDirectories, ", ") + " - manual review needed"
	}

	// Fallback: check if DetailedGuide is specified in Security Insights
	if data.RestData != nil && data.Insights.Project.Documentation.DetailedGuide != "" {
		return gemara.NeedsReview, "No design documentation file found, but detailed guide specified in Security Insights - manual review needed to confirm design documentation with actions and actors"
	}

	return gemara.Failed, "Design documentation demonstrating all actions and actors was NOT found"
}
