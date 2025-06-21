package data

import (
	"testing"

	"github.com/google/go-github/v71/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
)

func TestLoadRepositoryMetadata(t *testing.T) {
	testCases := []struct {
		name              string
		owner             string
		repo              string
		responses         []mock.MockBackendOption
		expectedRepoError bool
		expectedOrgError  bool
	}{
		{
			name:  "valid repository",
			owner: "test-owner",
			repo:  "test-repo",
			responses: []mock.MockBackendOption{
				mock.WithRequestMatch(
					mock.GetReposByOwnerByRepo,
					github.Repository{
						Owner: &github.User{
							Login: github.Ptr("test-owner"),
						},
						Name:     github.Ptr("test-repo"),
						Private:  github.Ptr(false),
						Archived: github.Ptr(false),
						Disabled: github.Ptr(false),
					},
				),
				mock.WithRequestMatch(
					mock.GetOrgsByOrg,
					github.Organization{
						Login:                       github.Ptr("test-owner"),
						TwoFactorRequirementEnabled: github.Ptr(true),
					},
				),
			},
			expectedRepoError: false,
			expectedOrgError:  false,
		},
		{
			name:              "invalid repository",
			owner:             "test-owner",
			repo:              "test-repo",
			expectedRepoError: true,
			expectedOrgError:  false,
		},
		{
			name:  "invalid organization",
			owner: "test-owner",
			repo:  "test-repo",
			responses: []mock.MockBackendOption{
				mock.WithRequestMatch(
					mock.GetReposByOwnerByRepo,
					github.Repository{
						Owner: &github.User{
							Login: github.Ptr("test-owner"),
						},
						Name:     github.Ptr("test-repo"),
						Private:  github.Ptr(false),
						Archived: github.Ptr(false),
						Disabled: github.Ptr(false),
					},
				),
			},
			expectedRepoError: false,
			expectedOrgError:  true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockClient := mock.NewMockedHTTPClient(
				testCase.responses...,
			)
			ghClient := github.NewClient(mockClient)
			repoMetadata, err := loadRepositoryMetadata(ghClient, testCase.owner, testCase.repo)
			if testCase.expectedRepoError {
				assert.Error(t, err)
			} else {
				if testCase.expectedOrgError {
					// When org data can't be retrieved, IsMFARequiredForAdministrativeActions returns nil
					assert.Nil(t, repoMetadata.IsMFARequiredForAdministrativeActions())
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, repoMetadata)
					assert.True(t, repoMetadata.IsActive())
					assert.True(t, repoMetadata.IsPublic())
					assert.NotNil(t, repoMetadata.IsMFARequiredForAdministrativeActions())
					assert.True(t, *repoMetadata.IsMFARequiredForAdministrativeActions())
					assert.Nil(t, repoMetadata.OrganizationBlogURL())
				}
			}
		})
	}
}
