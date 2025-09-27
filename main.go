package main

import (
	"fmt"

	"os"

	"github.com/revanite-io/pvtr-github-repo/data"
	"github.com/revanite-io/pvtr-github-repo/data/baseline"
	"github.com/revanite-io/pvtr-github-repo/evaluation_plans"

	"github.com/privateerproj/privateer-sdk/command"
	"github.com/privateerproj/privateer-sdk/pluginkit"
)

var (
	// Version is to be replaced at build time by the associated tag
	Version = "0.0.0"
	// VersionPostfix is a marker for the version such as "dev", "beta", "rc", etc.
	VersionPostfix = "dev"
	// GitCommitHash is the commit at build time
	GitCommitHash = ""
	// BuiltAt is the actual build datetime
	BuiltAt = ""

	PluginName   = "github-repo"
	RequiredVars = []string{
		"owner",
		"repo",
		"token",
	}
)

func main() {
	if VersionPostfix != "" {
		Version = fmt.Sprintf("%s-%s", Version, VersionPostfix)
	}

	pvtrVessel := pluginkit.NewEvaluationOrchestrator(PluginName, nil, RequiredVars)
	pvtrVessel.PluginVersion = Version
	pvtrVessel.PluginUri = "github.com/revanite-io/pvtr-github-repo"

	requirements, err := baseline.GetAssessmentRequirements()
	if err != nil {
		fmt.Printf("Error loading assessment requirements: %v\n", err)
		os.Exit(1)
	}

	pvtrVessel.AddEvaluationSuite("OSPS_B", data.Loader, evaluation_plans.OSPS_B, requirements)

	runCmd := command.NewPluginCommands(
		PluginName,
		Version,
		VersionPostfix,
		GitCommitHash,
		pvtrVessel,
	)

	err = runCmd.Execute()
	if err != nil {
		fmt.Printf("Error during runCmd.Execute(): %v\n", err)
		os.Exit(1)
	}
}
