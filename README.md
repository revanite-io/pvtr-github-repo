# Privateer Plugin for GitHub Repositories

This application performs automated assessments against GitHub repositories using controls defined in the [Open Source Project Security Baseline v2025.02.25](https://baseline.openssf.org). The application consumes the OSPS Baseline controls using [Gemara](https://github.com/gemaraproj/go-gemara) layer 2 and produces results of the automated assessments using layer 4.

Many of the assessments depend upon the presence of a [Security Insights](https://github.com/ossf/security-insights) file at the root of the repository, or `./github/security-insights.yml`.

## Work in Progress

Currently 43 control requirements across OSPS Baselines levels 1-3 are covered, with 9 not yet implemented. [Maturity Level 1](https://baseline.openssf.org/versions/2025-02-25.html#level-1) requirements are the most rigorously tested and are recommended for use. The results of these layer 1 assessments are integrated into [LFX Insights](https://insights.linuxfoundation.org/project/k8s/repository/kubernetes-kubernetes/security), powering the [Security & Best Practices results](https://insights.linuxfoundation.org/docs/metrics/security/).

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

## AI-Assisted Checks

A few OSPS Baseline requirements ask whether a project *documents* something, or
whether a setting is *appropriate for what it is used for*. Those questions
cannot be answered by pattern matching, so the scanner can optionally ask a
large language model and record its answer as evidence.

AI is **opt-in**. With no `ai_*` settings the scanner behaves exactly as it
always has, contacts no provider, and sends nothing anywhere.

### Configuration

<!-- markdownlint-disable MD013 -->

| Key | Environment variable | Required | Default | Purpose |
| --- | --- | :---: | --- | --- |
| `ai_provider` | `PVTR_AI_PROVIDER` | yes | -- | Backend adapter: `openai` or `anthropic`. |
| `ai_model` | `PVTR_AI_MODEL` | yes | -- | Provider-specific model identifier. |
| `ai_api_key` | `PVTR_AI_API_KEY` | yes | -- | Provider credential. |
| `ai_base_url` | `PVTR_AI_BASE_URL` | no | adapter default | Alternate endpoint: proxy, gateway, or self-hosted deployment. |
| `ai_timeout` | `PVTR_AI_TIMEOUT` | no | `30s` | Per-call timeout, as a Go duration string. |
| `ai_max_tokens` | `PVTR_AI_MAX_TOKENS` | no | `1024` | Cap on the model's response length. |

<!-- markdownlint-enable MD013 -->

Declare the keys at the top level to have every service inherit them, or inside
a single service's `vars` block to scope them to that service:

```yaml
ai_provider: openai
ai_model: gpt-4o-mini
# Supply the credential via PVTR_AI_API_KEY rather than writing it here.

services:
  my-scan:
    plugin: github-repo
    vars:
      owner: <github org or user name>
      repo: <github repo name>
      token: <classic token with permissions repo + admin:org>
```

### Reading The Results

A requirement answered with AI help is prefixed with `[AI-Assisted]`:

```text
[AI-Assisted] CONTRIBUTING.md tells contributors to run `go test ./...` before opening a pull request.
```

The model's verdict maps to `Passed`, `Failed`, or `Needs Review`, with a
`low` / `medium` / `high` confidence. Anything unexpected — a malformed
response, a provider error, a timeout, a missing credential — becomes
`Needs Review` at low confidence and the scan continues. **An AI-assisted check
never silently passes a requirement.**

The model's full reasoning, the exact prompt, the material it was shown, and the
model used are all recorded as evidence in the evaluation log, so a reviewer can
audit or dispute the answer.

Three requirements use AI today: `OSPS-QA-06.02`, `OSPS-QA-06.03`, and
`OSPS-AC-04.02`. At most one provider call is made per applicable requirement
per scan, and only the specific documentation or workflow files a question needs
are sent — never the whole repository.

Note that repository content is sent to whichever provider you configure, and
that provider usage is billed to you.

For provider examples, configuration precedence, how to validate a
configuration without provider spend, size limits, and the full evidence
format, see [docs/ai-assisted-checks.md](docs/ai-assisted-checks.md).

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
