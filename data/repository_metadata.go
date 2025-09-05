package data

import (
	"context"

	"github.com/google/go-github/v74/github"
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

func loadRepositoryMetadata(ghClient *github.Client, owner, repo string) (ghRepo *github.Repository, data RepositoryMetadata, err error) {
	repository, _, err := ghClient.Repositories.Get(context.Background(), owner, repo)
	if err != nil {
		return repository, &GitHubRepositoryMetadata{}, err
	}
	organization, _, err := ghClient.Organizations.Get(context.Background(), owner)
	if err != nil {
		return repository, &GitHubRepositoryMetadata{
			ghRepo: repository,
		}, nil
	}
	return repository, &GitHubRepositoryMetadata{
		ghRepo: repository,
		ghOrg:  organization,
	}, nil
}
