# Quill Manifest Reference — `quill.yaml`

## Overview

`quill.yaml` is the manifest file for a Quill skill. It lives at the root of a skill directory and describes everything the registry, installer, and runtime need to know: identity, dependencies, model variants, required permissions, context window hints, the public interface, eval pointers, bench configuration, conflict declarations, and provenance metadata.

The file is validated against a JSON Schema on `quill publish` and on `quill install`. A local schema copy is cached at `~/.quill/schemas/manifest-v1.json`.

---

## Full YAML Schema

```yaml
# ── Identity ────────────────────────────────────────────────────────────────
name: <string>                        # required — kebab-case, globally unique in registry
version: <semver>                     # required
description: <string>                 # required — one or two sentences shown in search results
authors:
  - name: <string>
    github: <string>                  # optional GitHub handle
license: <spdx-identifier>            # optional, e.g. MIT, Apache-2.0

# ── Dependencies ────────────────────────────────────────────────────────────
dependencies:
  skills:
    <skill-name>: <version-range>     # semver range, e.g. "^1.0.0"
  tools:
    <tool-name>: <version-range>      # CLI tools the skill shells out to

# ── Variants ────────────────────────────────────────────────────────────────
variants:
  - model_pattern: <glob>             # e.g. "claude-3-5-*", "gpt-4o*", "*"
    system_prompt_file: <path>        # relative path within skill dir
    temperature: <float 0-2>
    max_tokens: <integer>
    extra_stop_sequences: [<string>]

# ── Permissions ─────────────────────────────────────────────────────────────
permissions:
  filesystem:
    read: [<glob-or-path>]
    write: [<glob-or-path>]
  network:
    outbound: [<hostname-or-pattern>]
  shell:
    allow: [<command-prefix>]
    deny: [<command-prefix>]
  context_scope: <all | project | skill>

# ── Context ─────────────────────────────────────────────────────────────────
context:
  recommended_window: <integer>       # tokens; informational
  injects:
    - type: <file | tool_output | memory>
      source: <path-or-expression>
      max_tokens: <integer>
      required: <boolean>

# ── Interface ────────────────────────────────────────────────────────────────
interface:
  input:
    type: <string | object>
    schema_file: <path>               # JSON Schema file for object inputs
    description: <string>
  output:
    type: <string | object>
    schema_file: <path>
    description: <string>
  examples:
    - input: <string or inline object>
      output: <string or inline object>

# ── Evals ────────────────────────────────────────────────────────────────────
evals:
  file: <path>                        # default: evals.json
  auto_run: <boolean>                 # run evals on quill install? default false
  required_pass_rate: <float 0-1>     # gate publish if bench rate drops below this

# ── Bench ────────────────────────────────────────────────────────────────────
bench:
  schedule: <cron-expression>         # how often the registry re-runs evals
  notify_on_regression: <boolean>     # ping author on regression; default true
  baseline_model: <model-id>          # model used for the "canonical" score

# ── Conflicts ────────────────────────────────────────────────────────────────
conflicts:
  - skill: <skill-name>
    version_range: <semver-range>
    reason: <string>

# ── Provenance ───────────────────────────────────────────────────────────────
provenance:
  source_repo: <url>
  commit: <git-sha>
  generated_by: <tool-name>           # e.g. "quill scaffold"
  generated_at: <iso8601-datetime>
```

---

## Field Reference — Required Fields

| Field | Type | Description |
|---|---|---|
| `name` | string | Kebab-case skill name. Must be globally unique in the registry. Validated against `^[a-z][a-z0-9-]{1,63}$`. |
| `version` | semver | Semantic version of this skill release. Must be incremented on every `quill publish`. |
| `description` | string | Short human-readable description. Shown in `quill search` results and the registry UI. 10–280 characters. |

---

## Field Reference — Optional Fields

| Field | Type | Default | Description |
|---|---|---|---|
| `authors` | array | `[]` | List of author objects. Each has `name` (required) and `github` (optional). |
| `license` | SPDX string | `"MIT"` | License identifier. Validated against the SPDX license list. |
| `dependencies.skills` | map | `{}` | Map of skill-name → semver range. Resolved and pinned into `quill.lock` on install. |
| `dependencies.tools` | map | `{}` | Map of external CLI tool name → version range. Checked at install time; install fails if tool is absent. |
| `variants` | array | `[]` | Model-specific overrides. Evaluated in order; first match wins. A catch-all `"*"` at the end is recommended. |
| `permissions` | object | see below | Declared capability requirements. Runtime enforces these; requests outside declared permissions are blocked. |
| `context.recommended_window` | integer | `8192` | Hint to the runtime about how many context tokens this skill works best with. |
| `context.injects` | array | `[]` | Context injections prepended to the skill's prompt at runtime. |
| `interface.input.schema_file` | path | — | Path to a JSON Schema file that validates structured inputs. Used by `quill eval run` and the registry type-checker. |
| `interface.output.schema_file` | path | — | Path to a JSON Schema file that validates structured outputs. |
| `interface.examples` | array | `[]` | Illustrative input/output pairs shown in registry UI. |
| `evals.file` | path | `evals.json` | Location of the eval suite relative to the manifest. |
| `evals.auto_run` | boolean | `false` | If `true`, `quill install` runs the eval suite immediately after install and prints pass rates. |
| `evals.required_pass_rate` | float | `0.0` | `quill publish` fails if the current bench pass rate is below this threshold. |
| `bench.schedule` | cron | `"0 2 * * *"` | Cron expression for how often the registry bench re-runs evals. |
| `bench.notify_on_regression` | boolean | `true` | Whether to notify the author via GitHub notification on a bench regression. |
| `bench.baseline_model` | model-id | registry default | The model used when computing the canonical pass rate for the skill. |
| `conflicts` | array | `[]` | Skills that must not be installed alongside this one. |
| `provenance` | object | — | Machine-written metadata about where the skill came from. |

---

## Versioning Rules

1. **Patch bump** (`1.2.3` → `1.2.4`): Bug fixes and prompt tweaks that do not change the interface schema or permission requirements.
2. **Minor bump** (`1.2.3` → `1.3.0`): New optional fields in `interface.output`, new variants, new optional context injections.
3. **Major bump** (`1.2.3` → `2.0.0`): Breaking changes to `interface.input` or `interface.output` schemas, removed permissions, renamed skill.

The registry enforces these rules heuristically on `quill publish` by diffing the new manifest against the published one. A violation is a warning by default; use `--strict-semver` to make it a hard failure.

---

## Inference Rules

When optional sections are omitted, the following defaults are inferred:

| Omitted Section | Inferred Behavior |
|---|---|
| `variants` | A single implicit variant with no overrides is used; system prompt is read from `system_prompt.md` if present. |
| `permissions` | `context_scope: project`, no filesystem/network/shell access. |
| `evals` | Evals file is assumed to be `evals.json` in the skill root; if absent, eval commands are no-ops. |
| `bench.baseline_model` | Registry substitutes the current default model from registry config. |
| `conflicts` | None declared; skill is assumed to coexist with all others. |
| `dependencies.skills` | Empty; no skill dependencies. |

---

## Variant Selection Logic

At runtime, the variants array is evaluated top-to-bottom. The first entry whose `model_pattern` glob matches the active model ID is selected. If no entry matches, behavior depends on whether a `"*"` catch-all exists:

- **Catch-all present**: use it.
- **No catch-all**: use the base manifest values with no variant overrides, and emit a warning to stderr.

Variant fields that are not specified inherit from the manifest's top-level defaults (if any). Fields not present in either location fall back to the runtime's own defaults.

---

## Permission Values

### `filesystem`

| Key | Value type | Description |
|---|---|---|
| `read` | glob array | File paths or glob patterns the skill may read. Relative paths are resolved from the project root. |
| `write` | glob array | File paths or glob patterns the skill may write or create. |

### `network`

| Key | Value type | Description |
|---|---|---|
| `outbound` | hostname array | Hostnames or wildcard patterns (e.g. `*.github.com`) the skill may make outbound HTTP/S requests to. |

### `shell`

| Key | Value type | Description |
|---|---|---|
| `allow` | string array | Command prefixes the skill is permitted to run via shell execution. |
| `deny` | string array | Command prefixes explicitly blocked, taking precedence over `allow`. |

### `context_scope`

| Value | Meaning |
|---|---|
| `all` | Skill receives the full conversation and project context. |
| `project` | Skill receives project-level context (file tree, open files) but not conversation history. |
| `skill` | Skill receives only the input passed directly to it; no ambient context. |
