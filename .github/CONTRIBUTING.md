# Contributing to Privateer

We'd love to accept your patches and contributions to this project. There are just a few small guidelines you need to follow.

## Code of Conduct

Participation in this project comes under the [Contributor Covenant Code of Conduct](./CODE_OF_CONDUCT.md)

## Code Submission

Thank you for considering submitting code to Privateer!

- We follow the [GitHub Pull Request Model](https://help.github.com/articles/about-pull-requests/) for all contributions.
- For large bodies of work, we recommend creating an issue using the "Feature Request" template to outline the feature that you wish to build, and describe how it will be implemented. This gives a chance for review to happen early, and ensures no wasted effort occurs.
- For new features, documentation must be included. Currently we do not have a formalized documentation process, so please use your best judgment until a process is in place.
- All submissions, including submissions by project members, will require review before being merged.
- Once review has occurred, please rebase your PR down to a single commit. This will ensure a nice clean Git history.
- Please write a [good Git Commit message](https://chris.beams.io/posts/git-commit/)
- Please follow the code formatting instructions below

## Forking

If you come from another language, such as Python, imports behave a bit differently in Go projects than you may be familiar with.

Please review [this guide](https://blog.sgmansfield.com/2016/06/working-with-forks-in-go/) for suggestions on how to successfully develop on a forked branch.

One key to remember: only use `go get` once! The rest of the time you should use two remotes: one to pull code from the primary repo, and another to push code to your fork.

## Formatting

When submitting pull requests, make sure to do the following:

- Format all Go code with `gofmt`. Many people use `goimports` which fixes import statements and formats code in the same style of `gofmt`.
- Remove trailing whitespace. Many editors will do this automatically.
- Ensure any new files have a trailing newline

## Continuous Integration

Privateer uses Github Actions for all CI tasks. You may review the existing workflows in `.github/workflows`. Results of checks will automatically be pushed to PRs and may block merging if checks fail.
