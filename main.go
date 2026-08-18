package main

import (
	"embed"
	"fmt"
	"path/filepath"

	"os"

	"github.com/gemaraproj/go-gemara"
	"github.com/ossf/pvtr-github-repo-scanner/data"
	"github.com/ossf/pvtr-github-repo-scanner/evaluation_plans"

	"github.com/privateerproj/privateer-sdk/command"
	"github.com/privateerproj/privateer-sdk/config"
	"github.com/privateerproj/privateer-sdk/pluginkit"
	"github.com/privateerproj/privateer-sdk/shared"
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
	// catalogNamespaces declares the grc.store namespace that owns each embedded
	// catalog. The OSPS Baseline is OSSF-owned, but the vendored YAML carries no
	// metadata.author.id and `pvtr publish` refuses to guess an owner.
	catalogNamespaces = map[string]string{
		"osps-baseline":         "ossf",
		"osps-baseline-2025-10": "ossf",
		"osps-baseline-2026-02": "ossf",
	}
	//go:embed data/catalogs
	files   embed.FS
	dataDir = filepath.Join("data", "catalogs")
)

func main() {
	if VersionPostfix != "" {
		Version = fmt.Sprintf("%s-%s", Version, VersionPostfix)
	}

	orchestrator, err := newOrchestrator(Version)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(shared.InternalError)
	}

	runCmd := command.NewPluginCommands(
		PluginName,
		Version,
		VersionPostfix,
		GitCommitHash,
		orchestrator,
	)

	err = runCmd.Execute()
	if err != nil {
		os.Exit(shared.InternalError)
	}
}

// newOrchestrator builds the fully populated orchestrator for the given plugin
// version. Publisher, License and catalogNamespaces are inert at run time and
// exist so `pvtr publish` can derive the plugin's grc.store coordinate
// (ossf/github-repo) and its catalog linkage from the binary itself.
func newOrchestrator(version string) (*pluginkit.EvaluationOrchestrator, error) {
	orchestrator := &pluginkit.EvaluationOrchestrator{
		PluginName:        PluginName,
		PluginVersion:     version,
		PluginUri:         "https://github.com/ossf/pvtr-github-repo-scanner",
		Publisher:         "ossf",
		License:           "Apache-2.0",
		CatalogNamespaces: catalogNamespaces,
	}
	orchestrator.AddLoader(data.Loader)
	orchestrator.AddTargetBuilder(func(c *config.Config) gemara.Resource {
		slug := fmt.Sprintf("%s/%s", c.GetString("owner"), c.GetString("repo"))
		return gemara.Resource{
			Id:   fmt.Sprintf("github.com/%s", slug),
			Name: slug,
			Type: gemara.Software,
			Uri:  fmt.Sprintf("https://github.com/%s", slug),
		}
	})

	if err := orchestrator.AddReferenceCatalogs(dataDir, files); err != nil {
		return nil, fmt.Errorf("error loading catalog: %w", err)
	}

	orchestrator.AddRequiredVars(RequiredVars)

	if err := pluginkit.AddEvaluationSuiteTypedForAllCatalogs(orchestrator, nil, evaluation_plans.AllSteps()); err != nil {
		return nil, fmt.Errorf("error adding evaluation suites: %w", err)
	}
	return orchestrator, nil
}
