# Privateer Plugin for GitHub Repositories

This application performs automated assessments against GitHub repositories using controls defined in the [Open Source Project Security Baseline v2025.02.25](https://baseline.openssf.org). The application consumes the OSPS Baseline controls using [Gemara](https://github.com/ossf/gemara) layer 2 and produces results of the automated assessments using layer 4.

Many of the assessments depend upon the presence of a [Security Insights](https://github.com/ossf/security-insights) file at the root of the repository, or `./github/security-insights.yml`.

## Work in Progress

Currently 39 control requirements across OSPS Baselines levels 1-3 are covered, with 13 not yet implemented. [Maturity Level 1](https://baseline.openssf.org/versions/2025-02-25.html#level-1) requirements are the most rigorously tested and are recommended for use. The results of these layer 1 assessments are integrated into [LFX Insights](https://insights.linuxfoundation.org/project/k8s/repository/kubernetes-kubernetes/security), powering the [Security & Best Practices results](https://insights.linuxfoundation.org/docs/metrics/security/).

![alt text](kubernetes_insights_baseline.png)

Level 2 and Level 3 requirements are undergoing current development and may be less rigorously tested.

## Docker Usage

```sh
# build the image
docker build . -t local
docker run \
  --mount type=bind,source=./config.yml,destination=/.privateer/config.yml \
  --mount type=bind,source=./evaluation_results,destination=/.privateer/bin/evaluation_results \
  local
```

## GitHub Actions Usage

We've pushed images to GitHub Container Registry and Docker Hub for use in GitHub Actions.

You will need a GitHub personal access token with repository read permissions. This token should be added to your config file or as a secret in your repository.

### Example GHA Setup

- [Config](https://github.com/privateerproj/.github/blob/main/.github/pvtr-config.yml)
- [Workflow Definition](https://github.com/privateerproj/.github/blob/main/.github/workflows/osps-baseline.yml)
- [Action Results](https://github.com/privateerproj/.github/actions/runs/13691384519/job/38285134201)

### CI/CD Pipeline

This repository uses GitHub Actions for continuous integration and deployment. The workflow automatically builds, tests, performs security scanning, and publishes multi-platform binaries and Docker images.

For Docker Hub publishing, add `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN` secrets to your repository settings.

## Contributing

Contributions are welcome! Please see our [Contributing Guidelines](.github/CONTRIBUTING.md) for more information.

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.
