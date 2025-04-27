package data

import (
	"context"
	"fmt"

	"github.com/google/go-github/v71/github"
	"github.com/privateerproj/privateer-sdk/config"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type Payload struct {
	*GraphqlRepoData
	*RestData
	Config             *config.Config
	SuspectedBinaries  []string
	RepositoryMetadata RepositoryMetadata
}

func Loader(config *config.Config) (payload interface{}, err error) {
	graphql, client, err := getGraphqlRepoData(config)
	if err != nil {
		return nil, err
	}
	ghClient := github.NewClient(oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.GetString("token")},
	)))
	repositoryMetadata, err := loadRepositoryMetadata(ghClient, config.GetString("owner"), config.GetString("repo"))
	if err != nil {
		return nil, err
	}
	suspectedBinaries, err := getSuspectedBinaries(client, config, graphql.Repository.DefaultBranchRef.Name)
	if err != nil {
		return nil, err
	}

	rest, err := getRestData(ghClient, config)
	if err != nil {
		return nil, err
	}
	return interface{}(Payload{
		GraphqlRepoData:    graphql,
		RestData:           rest,
		Config:             config,
		SuspectedBinaries:  suspectedBinaries,
		RepositoryMetadata: repositoryMetadata,
	}), nil
}

func getGraphqlRepoData(config *config.Config) (data *GraphqlRepoData, client *githubv4.Client, err error) {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.GetString("token")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client = githubv4.NewClient(httpClient)

	variables := map[string]interface{}{
		"owner": githubv4.String(config.GetString("owner")),
		"name":  githubv4.String(config.GetString("repo")),
	}

	err = client.Query(context.Background(), &data, variables)
	if err != nil {
		config.Logger.Error(fmt.Sprintf("Error querying GitHub GraphQL API: %s", err.Error()))
	}
	return data, client, err
}

func getRestData(ghClient *github.Client, config *config.Config) (data *RestData, err error) {
	r := &RestData{
		ghClient: ghClient,
		Config:   config,
	}
	return r, r.Setup()
}
