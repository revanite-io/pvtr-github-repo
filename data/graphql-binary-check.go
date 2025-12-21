package data

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/privateerproj/privateer-sdk/config"
	"github.com/shurcooL/githubv4"
)

type GraphqlRepoTree struct {
	Repository struct {
		Object struct {
			Tree struct {
				Entries []struct {
					Name   string
					Type   string
					Path   string
					Object *struct {
						Blob struct {
							IsBinary    *bool
							IsTruncated bool
						} `graphql:"... on Blob"`
						Tree struct {
							Entries []struct {
								Name   string
								Type   string
								Path   string
								Object *struct {
									Blob struct {
										IsBinary    *bool
										IsTruncated bool
									} `graphql:"... on Blob"`
									Tree struct {
										Entries []struct {
											Name   string
											Type   string
											Path   string
											Object *struct {
												Blob struct {
													IsBinary    *bool
													IsTruncated bool
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

type binaryChecker struct {
	httpClient *http.Client
	logger     hclog.Logger
	owner      string
	repo       string
	branch     string
}

func (bc *binaryChecker) isBinary(isBinaryPtr *bool, isTruncated bool, path string) bool {
	if isBinaryPtr != nil {
		return *isBinaryPtr
	}
	if isTruncated {
		binary, err := bc.checkViaPartialFetch(path)
		if err != nil {
			bc.logger.Trace(fmt.Sprintf("failed to check binary status via partial fetch for %s: %s", path, err.Error()))
			return false
		}
		return binary
	}
	return false
}

func (bc *binaryChecker) checkViaPartialFetch(path string) (bool, error) {
	segments := strings.Split(path, "/")
	for i, seg := range segments {
		segments[i] = url.PathEscape(seg)
	}
	escapedPath := strings.Join(segments, "/")
	rawURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", bc.owner, bc.repo, bc.branch, escapedPath)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("Range", "bytes=0-511")

	resp, err := bc.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	return hasNullBytes(content), nil
}

func hasNullBytes(content []byte) bool {
	return bytes.IndexByte(content, 0) != -1
}

func checkTreeForBinaries(tree *GraphqlRepoTree, bc *binaryChecker) (binariesFound []string) {
	if tree == nil {
		return nil
	}
	for _, entry := range tree.Repository.Object.Tree.Entries {
		if entry.Type == "blob" && entry.Object != nil {
			if bc.isBinary(entry.Object.Blob.IsBinary, entry.Object.Blob.IsTruncated, entry.Path) {
				binariesFound = append(binariesFound, entry.Name)
			}
		}
		if entry.Type == "tree" && entry.Object != nil {
			for _, subEntry := range entry.Object.Tree.Entries {
				if subEntry.Type == "blob" && subEntry.Object != nil {
					if bc.isBinary(subEntry.Object.Blob.IsBinary, subEntry.Object.Blob.IsTruncated, subEntry.Path) {
						binariesFound = append(binariesFound, subEntry.Name)
					}
				}
				if subEntry.Type == "tree" && subEntry.Object != nil {
					for _, subSubEntry := range subEntry.Object.Tree.Entries {
						if subSubEntry.Type == "blob" && subSubEntry.Object != nil {
							if bc.isBinary(subSubEntry.Object.Blob.IsBinary, subSubEntry.Object.Blob.IsTruncated, subSubEntry.Path) {
								binariesFound = append(binariesFound, subSubEntry.Name)
							}
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
	path := ""

	fullPath := fmt.Sprintf("%s:%s", branch, path)

	variables := map[string]any{
		"owner":  githubv4.String(config.GetString("owner")),
		"name":   githubv4.String(config.GetString("repo")),
		"branch": githubv4.String(fullPath),
	}

	err = client.Query(context.Background(), &tree, variables)

	return tree, err
}
