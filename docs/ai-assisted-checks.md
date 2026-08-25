# AI-Assisted Checks

This guide explains how to turn on AI-assisted checks in the Privateer GitHub
repository scanner, how to set up a provider, and how to read the results.

The AI client, the answer format, and the `ai_*` settings all come from
[`privateer-sdk`](https://github.com/privateerproj/privateer-sdk). See `go.mod`
for the version this scanner currently builds against; an SDK upgrade can change
any of this.

For a short introduction, see the
[AI-Assisted Checks](../README.md#ai-assisted-checks) section of the README.

## Why Some Requirements Need AI

Most Baseline requirements can be answered by looking for something specific: a
file exists, a setting has a certain value, a release has an SBOM attached. A
few ask whether something is *good enough* rather than whether it is *present*.

`OSPS-QA-06.02` asks whether a project documents when and how its tests are run.
Searching for the word "test" matches almost every repository and proves
nothing: "run the tests before submitting" leaves a contributor without a
command, and a `make test` snippet never says whether tests are expected on
every change or only at release time. Both mention testing; neither satisfies
the requirement. The difference is in meaning, not in the presence of a word, so
the check has to read the documentation rather than search it.

For requirements like these, the scanner can send the relevant files to an AI
model and ask for an answer in a fixed format.

## Turning AI On

AI is off until you configure it. With no `ai_*` settings the scanner makes no
requests to any provider, and every AI-capable check gives the same answer it
gave before AI support existed.

The smallest working configuration is a provider, a model, and an API key:

```yaml
ai_provider: openai
ai_model: gpt-4o-mini
# Pass the API key via PVTR_AI_API_KEY rather than writing it here.

services:
  my-scan:
    plugin: github-repo
    vars:
      owner: <github org or user name>
      repo: <github repo name>
      token: <classic token with permissions repo + admin:org>
```

AI counts as on as soon as *any* `ai_*` setting is present. If the required ones
are not all there, the affected checks report `Needs Review` and log a warning
rather than quietly carrying on without AI, so a half-finished configuration is
visible instead of silent. The scan itself still finishes.

## Configuration Settings

<!-- markdownlint-disable MD013 -->

| Setting | Environment variable | Required | Default | Purpose |
| --- | --- | :---: | --- | --- |
| `ai_provider` | `PVTR_AI_PROVIDER` | yes | -- | Which AI service to use: `openai` or `anthropic`. |
| `ai_model` | `PVTR_AI_MODEL` | yes | -- | Model name, spelled the way your provider spells it. |
| `ai_api_key` | `PVTR_AI_API_KEY` | yes | -- | Your API key for that provider. |
| `ai_base_url` | `PVTR_AI_BASE_URL` | no | provider default | A different endpoint: a proxy, a gateway, or a deployment you host yourself. |
| `ai_timeout` | `PVTR_AI_TIMEOUT` | no | `30s` | How long to wait for one answer, for example `30s` or `2m`. |
| `ai_max_tokens` | `PVTR_AI_MAX_TOKENS` | no | `1024` | Longest answer to allow back from the model. |

<!-- markdownlint-enable MD013 -->

If you do not set `ai_base_url`, the scanner uses `https://api.openai.com/v1`
for `openai` and `https://api.anthropic.com/v1` for `anthropic`.

`ai_max_tokens` must be a whole number; the rest must be text. An `ai_timeout`
the scanner cannot read stops the run instead of quietly using the default.

Avoid setting `ai_max_tokens` much lower than the default. The answer has to
carry a short message, a longer explanation, and any citations, so a tight limit
can cut the answer off and send the check to manual review.

### Where Settings Go

Settings at the top level apply to every service. Settings inside a service's
`vars` block apply to that service only, and win over the top-level value:

```yaml
ai_provider: openai
ai_model: gpt-4o-mini        # every service uses this model...

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
      ai_model: <a more capable model id>   # ...except this one
```

Any setting can also come from its `PVTR_AI_*` environment variable, which is
the usual way to pass the API key. The environment overrides a top-level setting
like the `ai_provider:` above, but not one inside a `vars:` block.

### Keeping The API Key Out Of The Config File

Prefer passing the API key through `PVTR_AI_API_KEY`, or from whatever secret
store your CI already uses, rather than writing it into `config.yml`:

```sh
export PVTR_AI_API_KEY='<your-provider-api-key>'
./pvtr run --binaries-path .
```

### Running Only Some Services With AI

Put the AI settings inside the services that need them rather than at the top
level. A service that says nothing about AI runs with it off.

Going the other way — inheriting top-level settings and then switching them off
for one service — is easy to get wrong. AI counts as off only when *every*
`ai_*` setting ends up empty, so blanking `ai_provider` and `ai_model` while an
`ai_api_key` is still reachable from the top level or the environment leaves the
service switched on but broken, and its AI-assisted checks report `Needs Review`.

## Provider Examples

The scanner has no code specific to any one provider. It asks the SDK for a
client and the SDK picks the provider from `ai_provider`, so the choices are
whichever ones the pinned SDK supports — currently `openai` and `anthropic`.
Each one has its own test suite in the SDK, so neither is an afterthought.

The example above uses OpenAI. To use Anthropic instead, change the provider and
the model:

```yaml
ai_provider: anthropic
ai_model: <anthropic model id>   # check the provider's current model list
```

### A Gateway Or Self-Hosted Endpoint

Use `ai_base_url` when calls have to go through a company gateway, a monitoring
proxy, or a deployment you host yourself that accepts the same requests as the
provider you picked. Leave `ai_provider` set to whichever provider's request
format the endpoint accepts.

```yaml
ai_provider: openai
ai_model: <model id exposed by the endpoint>
ai_base_url: https://ai-gateway.internal.example.com/v1
```

## Checking Your Configuration Without Paying For Calls

There is no dry-run mode, but you can still check most of a configuration
cheaply:

- **Typos cost nothing.** The provider name, model, and API key are checked on
  your machine before any network call happens. A provider name that is not
  supported, an empty `ai_model`, an `ai_api_key` left unset while other AI
  settings are present, or an `ai_timeout` the scanner cannot read all fail
  without contacting a provider.
- **Watch the logs.** Abandoned AI assessments are logged at `warn` level with
  the requirement id and the reason, so keep `loglevel` at `info` or lower (the
  default in `example-config.yml`) — at `error` these warnings are hidden. A run
  that logs no such warning and still asks for manual review on the requirements
  below means AI was never picked up at all, which usually means the settings
  are in the wrong place.
- **Point at a local endpoint.** Set `ai_base_url` to a local mock server that
  accepts OpenAI-style requests to exercise the whole AI path, including
  gathering content and checking the answer, without using your provider account.
- **Run one service.** Use `--service=<service_name>` to run a single service
  while you work on the configuration.

If you are looking for a dry-run flag you saw mentioned elsewhere: it existed
briefly in the SDK ([privateer-sdk#227](https://github.com/privateerproj/privateer-sdk/pull/227))
and was removed when the AI packages were reorganized in
[privateer-sdk#252](https://github.com/privateerproj/privateer-sdk/pull/252).
This scanner never had one of its own.

- **Typos cost nothing.** The provider name, model, and API key are checked on
  your machine before any network call happens. A provider name that is not
  supported, an empty `ai_model`, an `ai_api_key` left unset while other AI
  settings are present, or an `ai_timeout` the scanner cannot read all fail
  without contacting a provider.
- **Watch the logs.** Abandoned AI assessments are logged at `warn` level with
  the requirement id and the reason, so keep `loglevel` at `info` or lower (the
  default in `example-config.yml`) — at `error` these warnings are hidden. A run
  that logs no such warning and still asks for manual review on the requirements
  below means AI was never picked up at all, which usually means the settings
  are in the wrong place.
- **Point at a local endpoint.** Set `ai_base_url` to a local mock server that
  accepts OpenAI-style requests to exercise the whole AI path, including
  gathering content and checking the answer, without using your provider account.
- **Run one service.** Use `--service=<service_name>` to run a single service
  while you work on the configuration.

## Reading `[AI-Assisted]` Results

An assessment answered with AI help carries an `[AI-Assisted]` prefix on its
message:

```text
[AI-Assisted] CONTRIBUTING.md tells contributors to run `go test ./...` before opening a pull request.
```

When the model gives an answer but no usable message, the prefix is followed by
the answer itself, labelled `verdict`:

```text
[AI-Assisted] verdict: needs_review (medium confidence)
```

The message is always a single line, and anything past 160 characters is cut
with an ellipsis, so it reads like every other result message. The model's
longer reasoning is not thrown away; it is kept in the evidence described below
(up to 1500 characters).

### Answers

The model replies `pass`, `fail`, or `needs_review`, which become `Passed`,
`Failed`, and `Needs Review`. Anything else the model might send back — an
unexpected value, a missing field, a garbled reply — becomes `Needs Review`.

So an AI-assisted check **never silently passes a requirement**. The worst case
is that a person is asked to look at it.

### Confidence

The model reports its confidence as `low`, `medium`, or `high` — not as a
number. Treat it as a hint about how much to trust the answer: a `Passed` result
at `low` confidence is worth a quick look before you rely on it.

For `OSPS-AC-04.02` the scanner adds a check of its own. Unless the model gives
a clear answer at `high` confidence, the result is recorded as `Needs Review` at
`low` confidence, with the model's summary added to what the regular check
found.

### When AI Cannot Answer

Every failure ends the same way: `Needs Review` at `low` confidence, a warning
in the log naming the requirement, no AI evidence saved, and the scan carries
on. This covers:

- AI settings are wrong and the client cannot be built.
- The repository content needed for the question could not be fetched.
- The content is larger than the limits described below.
- The provider returned an error, timed out, rate-limited the request, or
  rejected the API key.
- The answer did not match the expected format.

A failed AI call never turns into a `Failed` requirement.

## Evidence And Auditing

When the model answers, the scanner saves an entry of type `ai-assessment` in
the results file the scan already writes. There is no extra evidence file or
directory to collect.

The entry holds:

- the answer, the confidence, the short message, the longer explanation, and any
  citations the model gave;
- the exact question the model was asked and the exact content it was shown;
- where the answer came from: the provider, the model actually used, and the
  provider's request id;
- a description naming the files the answer was based on — as a permanent link
  to each file at the scanned commit where one can be built, and otherwise as a
  path from the repository root such as `/README.md`.

That is enough for a reviewer to judge, repeat, or dispute the answer without
access to the provider's own logs.

> **Do not point AI-assisted checks at content you would not publish.** The
> question and the content are written word for word into the results file, and
> nothing is removed or masked. Anything that should not appear in a results
> file should not be sent to an AI provider in the first place.

## Which Checks Use AI

AI never replaces a regular check. It is used only where a regular check does
not exist or cannot reach a conclusion, and its answer is saved as evidence
alongside every other observation the scan records.

<!-- markdownlint-disable MD013 -->

| Requirement | Question asked | When AI is consulted |
| --- | --- | --- |
| `OSPS-QA-06.02` | Does project documentation explain when and how tests are run? | Whenever AI is set up. Without AI the requirement reports `Needs Review`. |
| `OSPS-QA-06.03` | Does project documentation state a policy for maintaining tests? | Whenever AI is set up. Without AI the requirement reports `Needs Review`. |
| `OSPS-AC-04.02` | Are the permissions a CI/CD job grants itself the minimum it needs? | Only when the regular check cannot decide. |

<!-- markdownlint-enable MD013 -->

`OSPS-AC-04.02` is checked by the regular rules first. Granting everything fails
outright — whether written as `write-all` or as every scope listed at its
highest level. Permissions set to `none` or left empty pass, and a workflow with
no `permissions:` block at all is not applicable. None of those involve a model.
AI is only asked about what is left: a job holding a specific grant, such as
`contents: write`, where whether it is needed depends on what the job actually
does.

### Size Limits

The scanner limits what it will send:

- README and CONTRIBUTING content for the `OSPS-QA-06` requirements is limited
  to 64 KiB in total.
- Workflow content for `OSPS-AC-04.02` is limited to 50 workflow files and
  64 KiB. Both limits count only the workflows that actually need a judgment
  call, not every workflow in the repository, so a repository with many
  workflows can still be checked as long as few of them are unclear.

Going over a limit sends the requirement to manual review instead of cutting the
content short. Cutting it short could drop the very passage the answer depends
on and produce a confident but wrong answer, so the scanner does not guess.

## Cost And Operational Notes

- Each applicable requirement makes at most one call to the provider per scan.
  With the checks listed above, one scan makes at most three calls, and fewer
  when a regular check already answered.
- Request size is limited by the caps above, and answer size by
  `ai_max_tokens`.
- Only the specific files a question needs are sent — documentation and workflow
  files — never the whole repository or its source code.
- The provider charges you for these calls. A small, inexpensive model is
  usually good enough for these questions.
- If a provider is unreachable or you are over quota, scans still finish; the
  affected requirements report `Needs Review`.
- Repository content is sent to whichever provider you configure. Confirm that
  is acceptable for the repositories you scan, especially private ones, before
  turning AI on.
