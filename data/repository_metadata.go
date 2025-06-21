package data

import (
	"context"

	"github.com/google/go-github/v71/github"
)

type RepositoryMetadata interface {
	IsActive() bool
	IsPublic() bool
	OrganizationBlogURL() *string
	IsMFARequiredForAdministrativeActions() *bool
}

type GitHubRepositoryMetadata struct {
	Releases []ReleaseData
	Rulesets []Ruleset
	ghRepo   *github.Repository
	ghOrg    *github.Organization
}

func (r *GitHubRepositoryMetadata) IsActive() bool {
	return !r.ghRepo.GetArchived() && !r.ghRepo.GetDisabled()
}

func (r *GitHubRepositoryMetadata) IsPublic() bool {
	return !r.ghRepo.GetPrivate()
}

func (r *GitHubRepositoryMetadata) OrganizationBlogURL() *string {
	if r.ghOrg != nil {
		return r.ghOrg.Blog
	}
	return nil
}

func (r *GitHubRepositoryMetadata) IsMFARequiredForAdministrativeActions() *bool {
	if r.ghOrg == nil {
		return nil
	}
	return r.ghOrg.TwoFactorRequirementEnabled
}

func loadRepositoryMetadata(ghClient *github.Client, owner, repo string) (data RepositoryMetadata, err error) {
	repository, _, err := ghClient.Repositories.Get(context.Background(), owner, repo)
	if err != nil {
		return &GitHubRepositoryMetadata{}, err
	}
	organization, _, err := ghClient.Organizations.Get(context.Background(), owner)
	if err != nil {
		return &GitHubRepositoryMetadata{
			ghRepo: repository,
		}, nil
	}
	return &GitHubRepositoryMetadata{
		ghRepo: repository,
		ghOrg:  organization,
	}, nil
}
