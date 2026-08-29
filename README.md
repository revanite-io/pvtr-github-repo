# Privateer Plugin for GitHub Repositories

This application performs automated assessments against GitHub repositories using controls defined in the [Open Source Project Security Baseline](https://baseline.openssf.org). The application consumes the OSPS Baseline controls using [Gemara](https://github.com/gemaraproj/go-gemara) layer 2 and produces results of the automated assessments using layer 4.

Many of the assessments depend upon the presence of a [Security Insights](https://github.com/ossf/security-insights) file at the root of the repository, or `./github/security-insights.yml`.

## Catalog Versions

The scanner bundles multiple versions of the OSPS Baseline catalog. Select one in your [Privateer config](https://privateerproj.com/getting-started/quickstart/) under `policy.catalogs`:

- `osps-baseline` — always the latest bundled catalog (currently 2026-08). Use this to pick up new Baseline versions automatically.
- See the [catalog contract](./evaluation_plans/catalog_contract.go) to review the pinnable catalog versions.

> [!NOTE]
> Pinning catalog versions does not carry the same risks as software version pinning. Every release of the baseline is manually copied into this plugin. The content is used to tag steps to assessment requirements, then elevate their prose into the evaluation log.

## Work in Progress

Every assessment requirement in the bundled catalogs has a step implementation, though some return a needs-review result pending manual verification. [maturity-1](https://baseline.openssf.org) requirements are the most rigorously tested and are recommended for use. The results of these assessments are integrated into [LFX Insights](https://insights.linuxfoundation.org/project/k8s/repository/kubernetes-kubernetes/security), powering the [Security & Best Practices results](https://insights.linuxfoundation.org/docs/metrics/security/).

![alt text](kubernetes_insights_baseline.png)

Level 2 and Level 3 requirements are undergoing current development and may be less rigorously tested.

## Local Usage

To run the GitHub scanner locally, you will need the Privateer (`pvtr`) framework and the GitHub repository scanner (`pvtr-github-repo-scanner`) plugin.

1. Install pvtr using one of the methods described [here](https://github.com/privateerproj/privateer/blob/main/README.md#step-2-choose-your-installation-method).
2. Next, download the `pvtr-github-repo-scanner` plugin from the [releases](https://github.com/ossf/pvtr-github-repo-scanner/releases).

The following command is an example where the `pvtr`, the `pvtr-github-repo-scanner`, and the `config.yaml` are in the same directory.
```sh
./pvtr run --binaries-path .
```
If the binaries and the config files are in different directories specify the complete path using `--binaries-path` and `--config` flags.

You may have to adjust the plugin name in the config.yaml file to match them.

## Docker Usage

```sh
# build the image
docker build . -t local
docker run \
  -v ./config.yml:/.privateer/config.yml \
  -v ./evaluation_results:/.privateer/bin/evaluation_results \
  local
```

## GitHub Actions Usage

See the [OSPS Security Baseline Scanner](https://github.com/marketplace/actions/open-source-project-security-baseline-scanner)

## Best Practices Badge Integration

To use scan results with the OpenSSF Best Practices Badge, see the user guide in
[docs/best-practices-badge.md](docs/best-practices-badge.md).

## Contributing

Contributions are welcome! Please see our [Contributing Guidelines](.github/CONTRIBUTING.md) for more information.

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.
