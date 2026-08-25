# AI-Assisted Checks

AI-assisted checks help evaluate requirements that need judgment rather than a
simple file or setting check.

For example, `OSPS-QA-06.02` asks whether a project explains when and how to run
its tests. Finding the word "test" is not enough: the documentation must provide
useful instructions. An AI model can review the relevant documentation and
assess whether it meets the requirement.

AI is disabled by default.

## Enabling AI

Set an AI provider, model, and API key. The scanner supports `openai` and
`anthropic`.

The recommended setup uses environment variables so the API key is not stored
in the configuration file:

```sh
export PVTR_AI_PROVIDER='openai'
export PVTR_AI_MODEL='gpt-4o-mini'
export PVTR_AI_API_KEY='<your-api-key>'

./pvtr run --binaries-path .
```

For Anthropic, change the provider and model:

```sh
export PVTR_AI_PROVIDER='anthropic'
export PVTR_AI_MODEL='<anthropic-model-id>'
```

Environment variables apply to every service in the run. Use
`--service=<service-name>` to run only one configured service.

You can also add non-secret settings to the configuration file:

```yaml
ai_provider: openai
ai_model: gpt-4o-mini

services:
  my-scan:
    plugin: github-repo
    vars:
      owner: <github org or user name>
      repo: <github repo name>
      token: <classic token with permissions repo + admin:org>
```

Top-level settings apply to every service. To override a setting for one
service, add it to that service's `vars` block:

```yaml
services:
  my-scan:
    plugin: github-repo
    vars:
      ai_model: <different-model-id>
```

## Configuration

<!-- markdownlint-disable MD013 -->

| Setting | Environment variable | Required | Default | Description |
| --- | --- | :---: | --- | --- |
| `ai_provider` | `PVTR_AI_PROVIDER` | yes | -- | AI provider: `openai` or `anthropic`. |
| `ai_model` | `PVTR_AI_MODEL` | yes | -- | Model name from the provider. |
| `ai_api_key` | `PVTR_AI_API_KEY` | yes | -- | API key for the provider. |
| `ai_base_url` | `PVTR_AI_BASE_URL` | no | provider default | URL for a compatible gateway, proxy, or self-hosted endpoint. |
| `ai_timeout` | `PVTR_AI_TIMEOUT` | no | `30s` | Maximum time to wait for a response. |
| `ai_max_tokens` | `PVTR_AI_MAX_TOKENS` | no | `1024` | Maximum response size. |

<!-- markdownlint-enable MD013 -->

All three required settings must be present. If one is missing, the affected
checks report `Needs Review`, log a warning, and allow the scan to continue.
`ai_timeout` must use a duration such as `30s` or `2m`; an invalid duration stops
the run. `ai_max_tokens` must be a whole number.

### Custom Endpoints

Set `ai_base_url` to use a compatible gateway, proxy, or self-hosted endpoint.
The endpoint must accept the request format for the selected provider.

```yaml
ai_provider: openai
ai_model: <model-id>
ai_base_url: https://ai-gateway.example.com/v1
```

## Security And Privacy

The scanner sends repository content to the configured AI provider. It sends
only the files needed by the check: README, CONTRIBUTING, or relevant GitHub
Actions workflow files.

> **The scanner does not detect or redact secrets in repository files.** Do not
> enable AI for repositories whose relevant files contain secrets or other data
> that must not be sent to the provider.

The same content is stored without redaction in the scan evidence, together with
the question sent to the model and its response. Protect the results file
accordingly.

Keep the provider API key in `PVTR_AI_API_KEY` or your CI secret store. Do not
commit it to the configuration file.

Confirm that your organization permits the selected provider to process the
repository content.

## Reading Results

Results produced with AI include the `[AI-Assisted]` prefix:

```text
[AI-Assisted] CONTRIBUTING.md explains how and when contributors should run the tests.
```

The result is `Passed`, `Failed`, or `Needs Review`, with `low`, `medium`, or
`high` confidence. Review low-confidence results before relying on them.

An invalid response, provider error, timeout, missing setting, or oversized
input produces `Needs Review` at low confidence. The scan continues and logs a
warning. An AI failure never becomes a `Failed` requirement.

The scanner stores successful AI assessments as `ai-assessment` evidence in:

```text
<write-directory>/<service>/<service>.yaml
```

The evidence includes the result, confidence, explanation, question, content,
provider, model, request ID, and source file references.

## Checks That Use AI

AI is used only where a regular check does not exist or cannot answer the
requirement.

<!-- markdownlint-disable MD013 -->

| Requirement | AI assessment |
| --- | --- |
| `OSPS-QA-06.02` | Whether project documentation explains when and how tests are run. |
| `OSPS-QA-06.03` | Whether project documentation defines a policy for maintaining tests. |
| `OSPS-AC-04.02` | Whether specific GitHub Actions permissions are required by the job. |

<!-- markdownlint-enable MD013 -->

`OSPS-QA-06.02` and `OSPS-QA-06.03` report `Needs Review` when AI is disabled.
For `OSPS-AC-04.02`, regular checks handle clear cases first; AI reviews only
permissions that require context.

Each requirement makes at most one provider request per scan. Documentation sent
for `OSPS-QA-06` is limited to 64 KiB. `OSPS-AC-04.02` is limited to 50 workflow
files and 64 KiB. Inputs over these limits report `Needs Review`.
