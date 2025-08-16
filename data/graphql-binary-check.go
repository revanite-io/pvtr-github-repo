package data

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/privateerproj/privateer-sdk/config"
	"github.com/shurcooL/githubv4"
)

var (
	binaryExtensions = map[string]bool{
		"":       true,
		"tar":    true,
		"gz":     true,
		"tgz":    true,
		"zip":    true,
		"rar":    true,
		"7z":     true,
		"bz2":    true,
		"xz":     true,
		"lzma":   true,
		"lz4":    true,
		"zst":    true,
		"apk":    true,
		"crx":    true,
		"deb":    true,
		"dex":    true,
		"dey":    true,
		"elf":    true,
		"o":      true,
		"a":      true,
		"so":     true,
		"macho":  true,
		"iso":    true,
		"class":  true,
		"jar":    true,
		"bundle": true,
		"dylib":  true,
		"lib":    true,
		"msi":    true,
		"dll":    true,
		"drv":    true,
		"efi":    true,
		"exe":    true,
		"ocx":    true,
		"pyc":    true,
		"pyo":    true,
		"par":    true,
		"rpm":    true,
		"wasm":   true,
		"whl":    true,
	}

	// Extend this with more known filenames as needed
	knownFilenames = map[string]bool{
		"README":          true,
		"LICENSE":         true,
		"CHANGELOG":       true,
		"CONTRIBUTING":    true,
		"CODE_OF_CONDUCT": true,
		"TODO":            true,
		"SECURITY":        true,
		"NOTICE":          true,
		"CODEOWNERS":      true,
		".gitignore":      true,
		".gitattributes":  true,
		"Makefile":        true,
		"Dockerfile":      true,
		"Vagrantfile":     true,
		"Gemfile":         true,
		"Procfile":        true,
		"Brewfile":        true,
		"MANIFEST":        true,
		"DCO":             true,
		"MAINTAINERS":     true,
	}
)

// GraphqlRepoTree is used in a query to get top 3 levels of the repository contents
type GraphqlRepoTree struct {
	Repository struct {
		Object struct {
			Tree struct {
				Entries []struct {
					Name   string
					Type   string // "blob" for files, "tree" for directories
					Path   string
					Object *struct {
						Tree struct {
							Entries []struct {
								Name   string
								Type   string
								Path   string
								Object *struct {
									Tree struct {
										Entries []struct {
											Name string
											Type string
											Path string
										}
									} `graphql:"... on Tree"`
								} `graphql:"object"`
							}
						} `graphql:"... on Tree"`
					} `graphql:"object"`
				}
			} `graphql:"... on Tree"`
		} `graphql:"object(expression: $branch)"`
	} `graphql:"repository(owner: $owner, name: $name)"`
}

func checkTreeForBinaries(tree *GraphqlRepoTree) (binariesFound []string) {
	for _, entry := range tree.Repository.Object.Tree.Entries {
		binariesFound = identifyBinaries(binariesFound, entry.Type, entry.Name)
		if entry.Type == "tree" {
			for _, subEntry := range entry.Object.Tree.Entries {
				binariesFound = identifyBinaries(binariesFound, subEntry.Type, subEntry.Name)
				if subEntry.Type == "tree" {
					for _, subSubEntry := range subEntry.Object.Tree.Entries {
						binariesFound = identifyBinaries(binariesFound, subSubEntry.Type, subSubEntry.Name)
						// if subSubEntry.Type == "tree" {
						// TODO: The current GraphQL call stops after 3 levels of depth.
						// Additional API calls will be required for recursion if another tree is found.
						// }
					}
				}
			}
		}
	}
	return binariesFound
}

func identifyBinaries(binariesFound []string, filetype string, filename string) []string {
	if filetype == "blob" {
		if isBinaryFile(filename) {
			binariesFound = append(binariesFound, filename)
		}
	}
	return binariesFound
}

// TODO: this is a lightweight check, looking at filenames only.
// GitHub's GraphQL API has an 'isBinary' field that could be used for a more accurate check,
// but I didn't manage to get that query working as expected.
func isBinaryFile(filename string) bool {
	if knownFilenames[filename] {
		return false
	}
	ext := filepath.Ext(filename)
	return binaryExtensions[ext]
}

func fetchGraphqlRepoTree(config *config.Config, client *githubv4.Client, branch string) (tree *GraphqlRepoTree, err error) {
	path := "" // TODO: I suspected we should be able to target subdirectories this way, but it hasn't succeeded

	fullPath := fmt.Sprintf("%s:%s", branch, path) // Ensure correct format

	variables := map[string]any{
		"owner":  githubv4.String(config.GetString("owner")),
		"name":   githubv4.String(config.GetString("repo")),
		"branch": githubv4.String(fullPath),
	}

	err = client.Query(context.Background(), &tree, variables)

	return tree, err
}
