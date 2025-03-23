package data

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/privateerproj/privateer-sdk/config"
	"github.com/shurcooL/githubv4"
)

// GraphqlRepoData is used in a query to get general repository information
type GraphqlRepoData struct {
	Repository struct {
		Name                    string
		IsPrivate               bool
		HasDiscussionsEnabled   bool
		HasIssuesEnabled        bool
		IsSecurityPolicyEnabled bool

		DefaultBranchRef struct {
			Name          string
			RefUpdateRule struct {
				AllowsDeletions              bool
				AllowsForcePushes            bool
				RequiredApprovingReviewCount int
			}
			BranchProtectionRule struct {
				RestrictsPushes             bool // This didn't give an accurate result
				RequiresApprovingReviews    bool // This gave an accurate result
				RequiresCommitSignatures    bool
				RequiresStatusChecks        bool
				RequiredStatusCheckContexts []string
			}
			Target struct {
				OID    string `graphql:"oid"` // Latest commit SHA
				Commit struct {
					Status struct {
						State    string // Overall commit status
						Contexts []struct {
							Context     string
							Description string
							State       string
							TargetURL   string `graphql:"targetUrl"`
						}
					} `graphql:"status"` // Classic status API

					AssociatedPullRequests struct {
						Nodes []struct {
							StatusCheckRollup struct {
								Commit struct {
									CheckSuites struct {
										Nodes []struct {
											CheckRuns struct {
												Nodes []struct {
													Name string `graphql:"name"`
												}
											} `graphql:"checkRuns(first: 25)"`
										}
									} `graphql:"checkSuites(first: 25)"`
								}
							}
						}
					} `graphql:"associatedPullRequests(last: 1)"`
				} `graphql:"... on Commit"`
			} `graphql:"target"`
		}
		LicenseInfo struct {
			Name   string
			SpdxId string
			Url    string
		}
		LatestRelease struct {
			Description string
		}
		ContributingGuidelines struct {
			Body         string
			ResourcePath githubv4.URI
		}
	} `graphql:"repository(owner: $owner, name: $name)"`
}

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

func checkTreeForBinaries(tree *GraphqlRepoTree, binariesFound []string) []string {
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
	binaryExtensions := map[string]bool{
		"": true, ".exe": true, ".dll": true, ".so": true, ".png": true, ".jpg": true,
		".pdf": true, ".zip": true, ".tar": true, ".mp4": true, ".mp3": true,
	}
	knownFilenames := map[string]bool{
		// Extend this with more known filenames as needed
		"README": true, "LICENSE": true, "CHANGELOG": true, "CONTRIBUTING": true,
		"CODE_OF_CONDUCT": true, "TODO": true, "SECURITY": true, "NOTICE": true, "CODEOWNERS": true,
		".gitignore": true, ".gitattributes": true, "Makefile": true, "Dockerfile": true,
		"Vagrantfile": true, "Gemfile": true, "Procfile": true, "Brewfile": true, "MANIFEST": true,
	}
	if knownFilenames[filename] {
		return false
	}
	ext := filepath.Ext(filename)
	return binaryExtensions[ext]
}

func fetchGraphqlRepoTree(config *config.Config, client *githubv4.Client, branch string) (tree *GraphqlRepoTree, err error) {
	path := "" // TODO: I suspected we should be able to target subdirectories this way, but it hasn't succeeded

	fullPath := fmt.Sprintf("%s:%s", branch, path) // Ensure correct format

	variables := map[string]interface{}{
		"owner":  githubv4.String(config.GetString("owner")),
		"name":   githubv4.String(config.GetString("repo")),
		"branch": githubv4.String(fullPath),
	}

	err = client.Query(context.Background(), &tree, variables)

	return tree, err
}

func getSuspectedBinaries(client *githubv4.Client, config *config.Config, branchName string) (suspectedBinaries []string, err error) {
	tree, err := fetchGraphqlRepoTree(config, client, branchName)
	if err != nil {
		return nil, err
	}
	binaryFileNames := checkTreeForBinaries(tree, []string{})
	return binaryFileNames, nil
}
