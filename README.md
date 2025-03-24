# Privateer Plugin for GitHub Repositories

This plugin is designed to test a GitHub repository using automated assessments compatible with the [Simplified Compliance Infrastructure](https://github.com/revanite-io/sci) Layer 4 data types.

Many of the assessments require a [Security Insights](https://github.com/ossf/security-insights-spec) file to be present at the root of the repository, or `./github/security-insights.yml`.

## Work in Progress

Assessment development is currently addressing the [Open Source Project Security Baseline v2025.02.25](https://baseline.openssf.org).

As possible, the goal is to work on the OSPS Baseline maturity levels from the lowest to highest.

## GitHub Actions Usage

We've pushed an image to docker hub for use in GitHub Actions. Many tests are currently pending implementation, and only `Maturity Level 1` is currently recommended for use.

You will also need to set up a GitHub personal access token with the repository read permissions. This token should be added to your config file, or — if using the example pipeline below — as a secret in your repository.

### Example GHA Setup

- [Config](https://github.com/privateerproj/.github/blob/main/.github/pvtr-config.yml)
- [Workflow Definition](https://github.com/privateerproj/.github/blob/main/.github/workflows/osps-baseline.yml)
- [Action Results](https://github.com/privateerproj/.github/actions/runs/13691384519/job/38285134201)

## Local Development

While working on tests, the best way to run the plugin is via `go run . debug --service=<your-service>`. Ensure your local `config` file is set up correctly beforehand.

You may also pull the code locally and run the local Dockerfile:

1. Pull the repo
2. Modify `example-config.yml` to use your values, and rename it to `config.yml`
3. Build the Docker Image: `make docker-build`
4. Run the Docker Image: `make docker-run`
5. Review the output in the directory you've specified in your config file

## Required Token Scopes

![Token Scopes](./token-scopes.png)
