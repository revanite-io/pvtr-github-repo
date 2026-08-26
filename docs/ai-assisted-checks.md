# AI-Assisted Checks

Some requirements are subjective and cannot be answered by implementing checks for a file, setting or an API call in the code. For
example, `OSPS-QA-06.02` asks whether a project explains when and how to run its
tests. Finding the word "test" is not enough; the documentation must provide
useful instructions.

AI can assess requirements like this when a deterministic check cannot. AI is
optional and never replaces the deterministic checks.

## Enabling AI

<!-- markdownlint-disable MD013 -->

Configure AI with these settings in `config.yml`:

| Setting | Required | Default | Description |
| --- | :---: | --- | --- |
| `ai_provider` | yes | -- | AI provider: `openai` or `anthropic`. |
| `ai_model` | yes | -- | Model name from the provider. |
| `ai_api_key` | yes | -- | API key for the provider. Do not store it in `config.yml`. |
| `ai_base_url` | no | provider default | URL for a compatible gateway, proxy, or self-hosted endpoint. |
| `ai_timeout` | no | `30s` | Maximum time to wait for a response. |
| `ai_max_tokens` | no | `1024` | Maximum response size. |

<!-- markdownlint-enable MD013 -->

Add the settings to the service's `vars` in `config.yml`. This example shows
every setting and enables OpenAI for `my-scan`:

```yaml
services:
  my-scan:
    plugin: github-repo
    vars:
      owner: <github org or user name>
      repo: <github repo name>
      token: <classic token with permissions repo + admin:org>
      ai_provider: openai
      ai_model: gpt-4o-mini
      ai_base_url: https://api.openai.com/v1
      ai_timeout: 30s
      ai_max_tokens: 1024
```

Store the API key in an environment variable or CI secret rather than in
`config.yml`. The scanner reads it from `PVTR_AI_API_KEY`:

```sh
export PVTR_AI_API_KEY='<your-api-key>'
./pvtr run --binaries-path .
```

A CI secret may use any name as long as its value is exposed to the scanner as
`PVTR_AI_API_KEY`.

To use Anthropic, set `ai_provider` to `anthropic`, use an Anthropic model name,
and omit `ai_base_url` to use the provider default. Set `ai_base_url` only for a
compatible gateway, proxy, or self-hosted endpoint.

All three required settings must be present. If one is missing, the affected
checks report `Needs Review`, log a warning, and allow the scan to continue.
`ai_timeout` must use a duration such as `30s` or `2m`; an invalid duration stops
the run. `ai_max_tokens` must be a whole number.

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
