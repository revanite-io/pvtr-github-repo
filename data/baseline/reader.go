package baseline

import (
	"embed"

	"github.com/ossf/gemara/layer2"
	"github.com/privateerproj/privateer-sdk/pluginkit"
)

// We have tight control over the catalog right now while it is local, but
// if/when we have dynamic retrieval, we should retain the local files for testing
const dataDir string = "catalog"

//go:embed catalog
var files embed.FS

func GetBaselineCatalog() (layer2.Catalog, error) {
	return pluginkit.GetPluginCatalog(dataDir, files)
}
