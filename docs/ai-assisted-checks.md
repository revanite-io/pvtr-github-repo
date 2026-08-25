# AI-Assisted Checks

This guide explains how to enable AI-assisted assessments in the Privateer
GitHub repository scanner, how to configure a provider, and how to read the
results.

The behavior described here reflects
[`privateer-sdk`](https://github.com/privateerproj/privateer-sdk) **v1.32.2**,
the version this scanner currently builds against. The SDK owns the AI client,
the verdict schema, and the `ai_*` configuration keys, so upgrading the SDK can
change this surface.

For a short introduction, see the
[AI-Assisted Checks](../README.md#ai-assisted-checks) section of the README.

## What AI Assistance Does

A handful of OSPS Baseline requirements ask whether a project *documents*
something, or whether a configuration is *appropriate for what a job actually
does*. These questions cannot be answered by pattern matching alone. For those
requirements, the scanner can send a bounded slice of repository content to a
large language model and ask for a structured verdict.

AI never replaces a deterministic check. It is only consulted where a
deterministic check does not exist or cannot reach a conclusion, and its answer
is recorded as evidence alongside every other observation in the evaluation log.

## AI Is Opt-In

When none of the `ai_*` keys are set, AI is off. The scanner makes no requests
to any provider, and every AI-capable step keeps the same non-AI verdict it
produced before AI support existed.

Setting *any* `ai_*` key while a required one is missing is treated as a
misconfiguration rather than as "disabled" — including setting only an optional
key such as `ai_base_url`. The affected step reports `Needs Review` and logs a
warning; the scan itself still completes.

## Configuration Keys

<!-- markdownlint-disable MD013 -->

| Key | Environment variable | Required | Default | Purpose |
| --- | --- | :---: | --- | --- |
| `ai_provider` | `PVTR_AI_PROVIDER` | yes | -- | Backend adapter: `openai` or `anthropic`. |
| `ai_model` | `PVTR_AI_MODEL` | yes | -- | Provider-specific model identifier. |
| `ai_api_key` | `PVTR_AI_API_KEY` | yes | -- | Provider credential. |
| `ai_base_url` | `PVTR_AI_BASE_URL` | no | adapter default | Alternate endpoint: a proxy, gateway, or self-hosted deployment. |
| `ai_timeout` | `PVTR_AI_TIMEOUT` | no | `30s` | Per-call timeout, as a Go duration string. |
| `ai_max_tokens` | `PVTR_AI_MAX_TOKENS` | no | `1024` | Cap on the length of the model's response. |

<!-- markdownlint-enable MD013 -->

The default endpoints are `https://api.openai.com/v1` for `openai` and
`https://api.anthropic.com/v1` for `anthropic`.

`ai_max_tokens` must be an integer; the remaining keys must be strings. An
`ai_timeout` value that is not a valid Go duration is a hard error rather than a
silent fallback to the default.

Lowering `ai_max_tokens` much below the default is not recommended. The response
has to carry a short message, a long-form explanation, and any citations, so an
aggressive cap can truncate the answer and cause the step to fall back to manual
review.

### Where To Put The Keys

AI settings may be declared at the top level of the config file, in which case
every service inherits them, or inside a single service's `vars` block. For a
given key, the scanner resolves the first of these that is present:

1. `services.<service-name>.vars.<key>`
2. top-level `vars.<key>`
3. a top-level `<key>` entry
4. the corresponding `PVTR_*` environment variable

A per-service value therefore overrides an inherited one, which is how a single
service can use a different model or endpoint from the rest. See
[Running Only Some Services With AI](#running-only-some-services-with-ai) before
trying to use an override to disable AI for one service.

### Keeping The Credential Out Of The Config File

Prefer supplying the API key through `PVTR_AI_API_KEY`, or from whatever secret
store your CI already uses, rather than writing it into `config.yml`:

```sh
export PVTR_AI_API_KEY='<your-provider-api-key>'
./pvtr run --binaries-path .
```

## Provider Examples

### OpenAI

```yaml
ai_provider: openai
ai_model: gpt-4o-mini
# Supply the credential via PVTR_AI_API_KEY rather than this file.

services:
  my-scan:
    plugin: github-repo
    vars:
      owner: <github org or user name>
      repo: <github repo name>
      token: <classic token with permissions repo + admin:org>
```

### Anthropic

```yaml
ai_provider: anthropic
ai_model: <anthropic model id>   # check the provider's current model list
# Supply the credential via PVTR_AI_API_KEY rather than this file.

services:
  my-scan:
    plugin: github-repo
    vars:
      owner: <github org or user name>
      repo: <github repo name>
      token: <classic token with permissions repo + admin:org>
```

### A Gateway Or Self-Hosted Endpoint

Use `ai_base_url` when calls must travel through a corporate gateway, an
observability proxy, or a self-hosted deployment that speaks the selected
provider's wire protocol. Keep `ai_provider` set to the protocol the endpoint
implements.

```yaml
ai_provider: openai
ai_model: <model id exposed by the endpoint>
ai_base_url: https://ai-gateway.internal.example.com/v1
```

### Per-Service Overrides

A service's own `vars` win over anything inherited, so one service can use a
different model from the rest:

```yaml
ai_provider: openai
ai_model: gpt-4o-mini

services:
  routine-scan:
    plugin: github-repo
    vars:
      owner: <github org or user name>
      repo: <github repo name>
      token: <classic token>

  careful-scan:
    plugin: github-repo
    vars:
      owner: <github org or user name>
      repo: <other repo name>
      token: <classic token>
      ai_model: <a more capable model id>
```

### Running Only Some Services With AI

If only a subset of your services should use AI, configure AI **inside those
services** rather than at the top level. Services that say nothing about AI then
run with it disabled.

Opting a single service out of inherited settings is possible but easy to get
wrong. An explicitly empty value does override an inherited one, but AI counts
as "not configured" only when *every* `ai_*` key resolves to empty. Blanking
`ai_provider` and `ai_model` while an `ai_api_key` is still reachable — from a
top-level entry or from `PVTR_AI_API_KEY` — leaves the service configured but
invalid, and its AI-assisted requirements report `Needs Review`. Scoping AI to
the services that need it avoids the problem entirely.

## Validating Your Configuration Without Provider Spend

The scanner has no dry-run mode. AI dry-run existed briefly in the SDK
([privateer-sdk#227](https://github.com/privateerproj/privateer-sdk/pull/227))
and was removed when the AI packages were restructured in
[privateer-sdk#252](https://github.com/privateerproj/privateer-sdk/pull/252).
This scanner never implemented dry-run itself, so there is nothing to enable on
the current SDK.

You can still check most of a configuration cheaply:

- **Typos cost nothing.** The provider name, model, and credential are
  validated locally before any network call is made. An unsupported
  `ai_provider`, an empty `ai_model`, an `ai_api_key` left unset while other AI
  keys are present, or an unparseable `ai_timeout` all fail without contacting a
  provider.
- **Watch the logs.** Abandoned AI assessments are logged at `warn` level with
  the requirement id and the reason, so keep `loglevel` at `info` or lower (the
  default in `example-config.yml`) — at `error` these warnings are suppressed.
  A run that produces no such warning and still reports manual review for the
  requirements below means AI was never picked up at all, which usually points
  at the keys being in the wrong place.
- **Point at a local endpoint.** Set `ai_base_url` to a local
  OpenAI-compatible mock server to exercise the full AI code path, including
  evidence gathering and response validation, without any provider usage.
- **Scope the run.** Use `--service=<service_name>` to run a single service
  while you are iterating on the configuration.

## Reading `[AI-Assisted]` Results

An assessment answered with AI help carries an `[AI-Assisted]` prefix on its
message:

```text
[AI-Assisted] CONTRIBUTING.md tells contributors to run `go test ./...` before opening a pull request.
```

When the model returns a verdict but no usable message, the prefix is followed
by the verdict itself:

```text
[AI-Assisted] verdict: needs_review (medium confidence)
```

The message is always a single line and is capped at 160 characters, so it reads
like every other assessment message. The model's longer reasoning is not
discarded; it is kept in the evidence record described below.

### Verdicts

The model answers with `pass`, `fail`, or `needs_review`, which map to `Passed`,
`Failed`, and `Needs Review`. Anything else the model might return — an
unexpected value, a missing field, a malformed payload — maps to `Needs Review`.

An AI-assisted check therefore **never silently passes a requirement**. The
worst case is that a human is asked to look at it.

### Confidence

Confidence is a `low`, `medium`, or `high` enum reported by the model, not a
numeric score. Treat it as a triage aid: a `Passed` result at `low` confidence
deserves a spot check before you rely on it.

For `OSPS-AC-04.02` the scanner applies an extra guard of its own. Unless the
model returns a definite verdict at `high` confidence, the result is recorded as
`Needs Review` at `low` confidence, with the model's summary appended to the
deterministic finding.

### When AI Cannot Answer

Every failure path degrades to `Needs Review` at `low` confidence, logs a
warning naming the requirement, records no AI evidence, and lets the scan
continue. This covers:

- AI is configured incorrectly and the client cannot be built.
- The repository content needed for the question could not be retrieved.
- The content exceeds the size limits described below.
- The provider returned an error, timed out, rate-limited the request, or
  rejected the credential.
- The response did not conform to the expected verdict schema.

A failed AI call never turns into a `Failed` requirement.

## Evidence And Auditing

When the model answers, the scanner records a `gemara` evidence entry of type
`ai-assessment` in the normal evaluation log. There is no separate evidence file
or directory to collect.

The entry contains:

- the verdict, confidence, short message, long-form explanation, and any
  citations the model supplied;
- the exact prompt and the exact material the model was shown;
- provenance: the provider, the model actually used, and the provider's request
  identifier;
- a description naming the files the assessment was based on, as a permalink to
  each file at the scanned commit where one can be constructed, and otherwise as
  a repository-absolute path such as `/README.md`.

That is enough for a reviewer to judge, reproduce, or dispute the answer without
access to provider-side logs.

> **Do not point AI-assisted checks at content you would not publish.** The
> prompt and the material are written verbatim into the results file, and no
> redaction is performed on them. Anything that should not appear in a results
> file should not be sent to an AI provider in the first place.

## Which Checks Use AI

<!-- markdownlint-disable MD013 -->

| Requirement | Question asked | When AI is consulted |
| --- | --- | --- |
| `OSPS-QA-06.02` | Does project documentation explain when and how tests are run? | Whenever AI is configured. Without AI the requirement reports `Needs Review`. |
| `OSPS-QA-06.03` | Does project documentation state a policy for maintaining tests? | Whenever AI is configured. Without AI the requirement reports `Needs Review`. |
| `OSPS-AC-04.02` | Are the permissions a CI/CD job grants itself the minimum it needs? | Only when the deterministic check is inconclusive. |

<!-- markdownlint-enable MD013 -->

`OSPS-AC-04.02` is evaluated deterministically first. A `write-all` grant fails
outright, permissions set to `none` or left empty pass, and a workflow with no
explicit `permissions:` block is not applicable — none of those consult a model.
AI is only asked about the remaining case, where a job holds a specific grant
whose necessity depends on what the job actually does.

### Size Limits

The scanner bounds what it will send:

- README and CONTRIBUTING material for the `OSPS-QA-06` requirements is capped
  at 64 KiB combined.
- Workflow material for `OSPS-AC-04.02` is capped at 50 workflow files and
  64 KiB. Both caps apply only to the workflows that actually need semantic
  review, not to every workflow in the repository, so a repository with many
  workflows can still be assessed as long as few of them are ambiguous.

Exceeding a limit defers to manual review instead of truncating the input.
Truncation could drop the very passage the verdict depends on and produce a
confidently wrong answer, so the scanner refuses to guess.

## Cost And Operational Notes

- At most one provider call is made per applicable requirement per scan. With
  the checks listed above, a single scan makes at most three calls, and fewer
  when a deterministic check already answered.
- Request size is bounded by the caps above, and response size by
  `ai_max_tokens`.
- Only the specific files a question needs are sent — documentation and
  workflow definitions — never the whole repository or its source code.
- Usage charges are yours. A small, inexpensive model is generally sufficient
  for these questions.
- If a provider is unreachable or over quota, scans still complete; the affected
  requirements report `Needs Review`.
- The repository content is sent to whichever provider you configure. Confirm
  that is acceptable for the repositories you scan, particularly for private
  ones, before enabling AI.
