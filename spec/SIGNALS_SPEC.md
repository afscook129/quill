# Outcome Signals Specification

## Overview

Outcome signals are lightweight, anonymous telemetry records emitted after each skill invocation. They serve three purposes:

1. **Drift detection** — the registry monitors aggregate pass-rate trends over time and alerts skill authors when a skill's effectiveness degrades without a corresponding skill update (indicating a model change has shifted behavior).
2. **Worth-keeping scores** — signals from real-world usage supplement bench results; a skill that performs well on evals but poorly in production is surfaced as a candidate for review.
3. **Corpus growth** — anonymized signal metadata (not content) informs the registry's ranking and recommendation algorithms.

---

## Privacy Contract

Signals are designed to be safe to emit by default. The following table defines exactly what is and is not collected.

### What IS collected

```yaml
skill: ticket-classifier
version: "2.1.0"
model: claude-sonnet-4-5
harness: claude-code
outcome: success              # success | failure | timeout | error
trigger_matched: true         # was the skill's trigger pattern matched?
token_cost: 1240              # total tokens consumed (prompt + completion)
latency_ms: 3420              # wall-clock ms from invocation to response
eval_case_id: null            # set only during quill eval run; null in production
session_hash: "sha256:a1b2…"  # one-way hash of the session ID; not reversible
occurred_at: "2026-03-20T14:55:00Z"
quill_version: "0.4.1"
```

### What is NEVER collected

```yaml
# The following are explicitly excluded and never transmitted:
prompt_text: ~          # the system prompt or any variant thereof
input_content: ~        # the user's message or task input
output_content: ~       # the skill's response text
file_paths: ~           # any filesystem paths from the user's project
environment_vars: ~     # any environment variables
user_identity: ~        # no usernames, emails, or GitHub handles
ip_address: ~           # network identity is not logged server-side
project_name: ~         # name or path of the user's project
```

---

## Anonymization Rules

1. **Session hash**: The session identifier is hashed with SHA-256 before transmission. The hash is salted with a per-installation random salt stored in `~/.quill/signal_salt` (generated once on `quill init`, never transmitted). The salt ensures that session hashes cannot be correlated across installations.

2. **Token cost rounding**: `token_cost` is rounded to the nearest 10 tokens before transmission to prevent fingerprinting of unusual prompt lengths.

3. **Latency bucketing**: `latency_ms` is rounded to the nearest 100 ms.

4. **Outcome vocabulary**: `outcome` is restricted to the four-value enum (`success`, `failure`, `timeout`, `error`). No additional detail about the nature of a failure is transmitted.

5. **No free-text fields**: All signal fields are typed scalars or enums. There are no free-text or arbitrary-object fields that could inadvertently capture content.

---

## Opt-In Flow

### On `quill init`

During project initialization, the CLI prompts:

```
Quill can send anonymous outcome signals to improve skill quality and drift detection.
Signals contain no prompt text, input, output, or personally identifiable information.
Full spec: https://quill.dev/docs/signals

Enable outcome signals? [Y/n]
```

The response is written to `~/.quill/config.yaml` under `signals.enabled`. The default is `true` (opt-in on no response / Enter). Users who choose `n` will never have signals collected; this setting is re-checked on every invocation.

### CI Environments

In CI environments (detected via the `CI=true` environment variable), signals are **disabled by default** unless `QUILL_SIGNALS=1` is explicitly set. This prevents CI runs from inflating production signal counts.

---

## Signal Format

### Single Signal (JSON)

```json
{
  "schema_version": "1.0",
  "skill": "ticket-classifier",
  "version": "2.1.0",
  "model": "claude-sonnet-4-5",
  "harness": "claude-code",
  "outcome": "success",
  "trigger_matched": true,
  "token_cost": 1240,
  "latency_ms": 3400,
  "eval_case_id": null,
  "session_hash": "sha256:a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
  "occurred_at": "2026-03-20T14:55:00Z",
  "quill_version": "0.4.1"
}
```

### Batch Envelope (JSON)

Signals are batched and submitted together to reduce network overhead. The batch is flushed at process exit or when the buffer reaches 50 signals, whichever comes first.

```json
{
  "schema_version": "1.0",
  "batch_id": "b_7f3a9c2e",
  "submitted_at": "2026-03-20T15:00:00Z",
  "signal_count": 3,
  "signals": [
    {
      "skill": "ticket-classifier",
      "version": "2.1.0",
      "model": "claude-sonnet-4-5",
      "harness": "claude-code",
      "outcome": "success",
      "trigger_matched": true,
      "token_cost": 1240,
      "latency_ms": 3400,
      "eval_case_id": null,
      "session_hash": "sha256:a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
      "occurred_at": "2026-03-20T14:55:00Z",
      "quill_version": "0.4.1"
    },
    {
      "skill": "normalize-text",
      "version": "1.0.2",
      "model": "claude-sonnet-4-5",
      "harness": "claude-code",
      "outcome": "success",
      "trigger_matched": true,
      "token_cost": 430,
      "latency_ms": 1100,
      "eval_case_id": null,
      "session_hash": "sha256:a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
      "occurred_at": "2026-03-20T14:56:10Z",
      "quill_version": "0.4.1"
    },
    {
      "skill": "ticket-classifier",
      "version": "2.1.0",
      "model": "claude-sonnet-4-5",
      "harness": "claude-code",
      "outcome": "failure",
      "trigger_matched": true,
      "token_cost": 1250,
      "latency_ms": 3500,
      "eval_case_id": null,
      "session_hash": "sha256:b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3",
      "occurred_at": "2026-03-20T14:58:30Z",
      "quill_version": "0.4.1"
    }
  ]
}
```

---

## Collection Implementation

### Claude Code Hooks

When running inside Claude Code, signals are collected via the `PostToolUse` hook. The hook fires after each skill tool call completes. The hook payload provides `outcome`, `token_cost`, and `latency_ms` directly from the tool execution context.

Hook registration is automatic when a skill is installed with `quill install`. The hook definition is written to `.claude/hooks/quill-signals.json` and is scoped to the project.

### Other Harnesses

Harnesses other than Claude Code (e.g. custom scripts, CI runners) can emit signals by calling `POST /v1/signals` directly. See the Registry API reference for the endpoint specification. The `harness` field should be set to a string identifying the calling system (e.g. `"custom-ci"`, `"langchain"`, `"openai-assistants"`).

---

## Drift Detection from Signals

The registry drift detector runs on a 6-hour cycle and processes aggregated signals using the following steps:

1. **Windowed pass-rate calculation**: For each (skill, version, model) triple, compute the 7-day rolling pass-rate from signals where `outcome = success` divided by all non-`error` signals.

2. **Baseline comparison**: Compare the current rolling pass-rate against the most recent bench pass-rate for the same triple. The bench pass-rate is the ground truth; signal pass-rate provides a real-world supplement.

3. **Threshold evaluation**: If the signal pass-rate drops more than `drift_threshold_pp` percentage points below the bench pass-rate (default: 10 pp), a drift alert is generated.

4. **Alert dispatch**: Alerts are written to the registry's drift ledger and, if `bench.notify_on_regression: true` in the skill's manifest, dispatched to the author via GitHub notification.

### Anomaly Detection

In addition to drift, the following anomaly patterns trigger alerts:

| Pattern | Trigger condition | Severity |
|---|---|---|
| **Sudden drop** | Pass-rate falls > 20 pp within a single 6-hour window | high |
| **Token cost spike** | Average `token_cost` increases > 50% vs. 7-day baseline | medium |
| **Timeout surge** | `timeout` outcomes exceed 5% of signals in a window | medium |
| **Zero signals** | No signals received for a skill with > 100 weekly installs for 48 h | low |

---

## Data Retention

| Data type | Retention period |
|---|---|
| Raw signal records | 90 days |
| Aggregated daily pass-rates | 2 years |
| Drift alert records | 2 years |
| Bench results | Indefinite |

After 90 days, raw signal records are deleted and cannot be recovered. Aggregated statistics derived from them are retained.

---

## Configuration in `~/.quill/config.yaml`

```yaml
signals:
  enabled: true                        # master on/off switch
  endpoint: https://registry.quill.dev/v1/signals   # override for self-hosted registries
  batch_size: 50                       # flush after N signals (default 50)
  flush_on_exit: true                  # always flush buffer at process exit
  debug: false                         # log signal payloads to stderr before sending
```
