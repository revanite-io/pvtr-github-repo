package data

import (
	"context"

	"github.com/google/go-github/v71/github"
)

type RepositoryMetadata interface {
	IsActive() bool
	IsPublic() bool
	IsMFARequiredForAdministrativeActions() bool
	UnableToEvaluateMFARequirement() bool
	OrganizationBlogURL() *string
}

type GitHubRepositoryMetadata struct {
	Releases                       []ReleaseData
	Rulesets                       []Ruleset
	ghRepo                         *github.Repository
	ghOrg                          *github.Organization
	unableToEvaluateMFARequirement bool
}

func (r *GitHubRepositoryMetadata) IsActive() bool {
	return !r.ghRepo.GetArchived() && !r.ghRepo.GetDisabled()
}

func (r *GitHubRepositoryMetadata) IsPublic() bool {
	return !r.ghRepo.GetPrivate()
}

func (r *GitHubRepositoryMetadata) IsMFARequiredForAdministrativeActions() bool {
	return r.ghOrg.GetTwoFactorRequirementEnabled()
}

func (r *GitHubRepositoryMetadata) UnableToEvaluateMFARequirement() bool {
	return r.unableToEvaluateMFARequirement
}

func (r *GitHubRepositoryMetadata) OrganizationBlogURL() *string {
	if r.ghOrg != nil {
		return r.ghOrg.Blog
	}
	return nil
}

func loadRepositoryMetadata(ghClient *github.Client, owner, repo string) (data RepositoryMetadata, err error) {
	repository, _, err := ghClient.Repositories.Get(context.Background(), owner, repo)
	if err != nil {
		return &GitHubRepositoryMetadata{}, err
	}
	organization, _, err := ghClient.Organizations.Get(context.Background(), owner)
	if err != nil {
		return &GitHubRepositoryMetadata{
			ghRepo:                         repository,
			unableToEvaluateMFARequirement: true,
		}, nil
	}
	return &GitHubRepositoryMetadata{
		ghRepo: repository,
		ghOrg:  organization,
	}, nil
}
