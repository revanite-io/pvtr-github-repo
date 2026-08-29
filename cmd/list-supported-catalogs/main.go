package main

import (
	"fmt"

	"github.com/ossf/pvtr-github-repo-scanner/data"
)

func main() {
	// Print one catalog ID per line so shell scripts can iterate over the contract.
	for _, catalogID := range data.SupportedCatalogIDs {
		fmt.Println(catalogID)
	}
}
