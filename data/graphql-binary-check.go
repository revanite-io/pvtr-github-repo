package data

import (
	"context"
	"fmt"

	"github.com/privateerproj/privateer-sdk/config"
	"github.com/shurcooL/githubv4"
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
						Blob struct {
							IsBinary bool
						} `graphql:"... on Blob"`
						Tree struct {
							Entries []struct {
								Name   string
								Type   string
								Path   string
								Object *struct {
									Blob struct {
										IsBinary bool
									} `graphql:"... on Blob"`
									Tree struct {
										Entries []struct {
											Name   string
											Type   string
											Path   string
											Object *struct {
												Blob struct {
													IsBinary bool
												} `graphql:"... on Blob"`
											} `graphql:"object"`
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
	if tree == nil {
		return nil
	}
	for _, entry := range tree.Repository.Object.Tree.Entries {
		if entry.Type == "blob" && entry.Object != nil && entry.Object.Blob.IsBinary {
			binariesFound = append(binariesFound, entry.Name)
		}
		if entry.Type == "tree" && entry.Object != nil {
			for _, subEntry := range entry.Object.Tree.Entries {
				if subEntry.Type == "blob" && subEntry.Object != nil && subEntry.Object.Blob.IsBinary {
					binariesFound = append(binariesFound, subEntry.Name)
				}
				if subEntry.Type == "tree" && subEntry.Object != nil {
					for _, subSubEntry := range subEntry.Object.Tree.Entries {
						if subSubEntry.Type == "blob" && subSubEntry.Object != nil && subSubEntry.Object.Blob.IsBinary {
							binariesFound = append(binariesFound, subSubEntry.Name)
						}
						// TODO: The current GraphQL call stops after 3 levels of depth.
						// Additional API calls will be required for recursion if another tree is found.
					}
				}
			}
		}
	}
	return binariesFound
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
