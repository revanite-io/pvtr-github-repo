# OpenTofu configuration for managing the revanite-io/example-osps-baseline-level-1 repository

resource "github_repository" "example_osps_baseline_level_1" {
  name        = "example-osps-baseline-level-1"
  description = "Example repository for integration testing of pvtr-github-repo"
  visibility  = "public"
  has_issues  = true
  has_wiki    = true
  has_projects = true
  has_downloads = true
  vulnerability_alerts = true
}

resource "github_repository_ruleset" "default_branch_protection" {
  name        = "default"
  repository  = github_repository.example_osps_baseline_level_1.name
  target      = "branch"
  enforcement = "active"

  conditions {
    ref_name {
      include = ["~DEFAULT_BRANCH"]
      exclude = []
    }
  }

  rules {
    creation                = false
    update                  = true
    deletion                = true
    non_fast_forward        = true
    pull_request {
        required_approving_review_count = 1
    }
  }
}
