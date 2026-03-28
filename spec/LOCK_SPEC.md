# Lock File Reference — `quill.lock`

## Overview

`quill.lock` is a machine-generated YAML file that records the exact resolved state of every skill installed in a project. It is written by `quill install` and read by `quill run`, `quill eval`, and the runtime loader. The lock file ensures reproducible installs: two developers on the same project always run identical skill versions and configurations.

The lock file must be committed to version control. It should never be edited by hand.

---

## Full YAML Schema

```yaml
# ── Meta ─────────────────────────────────────────────────────────────────────
lock_version: <integer>               # schema version; currently 1
generated_at: <iso8601-datetime>
generated_by: <string>               # e.g. "quill/0.4.1"
quill_min_version: <semver>          # minimum quill CLI version required to read this lock

# ── Resolved Skills ───────────────────────────────────────────────────────────
resolved:
  - name: <string>
    version: <semver>
    source: <registry-url>
    integrity: <hash-algorithm>:<hex-digest>
    manifest_hash: <hash-algorithm>:<hex-digest>
    system_prompt_hash: <hash-algorithm>:<hex-digest>
    resolved_variant: <string | null>  # matched model_pattern, or null if no variant matched
    dependencies:
      - name: <string>
        version: <semver>
    install_path: <relative-path>
    installed_at: <iso8601-datetime>

# ── System ────────────────────────────────────────────────────────────────────
system:
  platform: <string>                  # e.g. "darwin/arm64"
  node_version: <string | null>
  quill_version: <string>
```

### Example with three skills

```yaml
lock_version: 1
generated_at: "2026-03-20T14:30:00Z"
generated_by: "quill/0.4.1"
quill_min_version: "0.4.0"

resolved:
  - name: ticket-classifier
    version: "2.1.0"
    source: "https://registry.quill.dev/v1/skills/ticket-classifier/2.1.0"
    integrity: "sha256:a3f9c2d1e4b7089f6a1c3e5d7b9f0e2a4c6d8f0a2b4c6d8e0f2a4b6c8d0e2f4"
    manifest_hash: "sha256:1b3d5f7a9c0e2b4d6f8a0c2e4f6a8b0d2f4a6c8e0b2d4f6a8c0e2b4d6f8a0c2"
    system_prompt_hash: "sha256:9e7c5a3f1d8b6e4c2a0f8d6b4a2c0e8f6d4b2a0c8e6d4b2a0f8c6e4d2b0a8f6"
    resolved_variant: "claude-3-5-*"
    dependencies:
      - name: normalize-text
        version: "1.0.2"
    install_path: ".quill/skills/ticket-classifier@2.1.0"
    installed_at: "2026-03-20T14:30:00Z"

  - name: normalize-text
    version: "1.0.2"
    source: "https://registry.quill.dev/v1/skills/normalize-text/1.0.2"
    integrity: "sha256:c1e3f5a7b9d0e2c4f6a8b0d2e4f6c8a0b2d4f6e8a0c2e4f6b8d0a2c4e6f8b0d2"
    manifest_hash: "sha256:2c4e6a8b0d2f4c6e8a0b2d4e6f8a2c4e6b8d0f2a4c6e8b0d2f4a6c8e0b2d4f6"
    system_prompt_hash: "sha256:8f6d4b2a0c8e6f4d2b0a8f6e4d2c0b8a6f4e2d0c8b6a4f2e0d8c6b4a2f0e8d6"
    resolved_variant: null
    dependencies: []
    install_path: ".quill/skills/normalize-text@1.0.2"
    installed_at: "2026-03-20T14:30:00Z"

  - name: my-custom-skill
    version: "0.1.0"
    source: "local"
    integrity: "sha256:f0e2d4c6b8a0f2e4d6c8b0a2f4e6d8c0b2a4f6e8d0c2b4a6f8e0d2c4b6a8f0e2"
    manifest_hash: "sha256:3d5f7a9b1c3e5f7a9b1c3e5f7a9b1c3e5f7a9b1c3e5f7a9b1c3e5f7a9b1c3e5"
    system_prompt_hash: "sha256:7a9b1c3e5d7f9a1b3c5e7d9f1a3b5c7e9d1f3a5b7c9e1d3f5a7b9c1e3d5f7a9"
    resolved_variant: "*"
    dependencies: []
    install_path: ".quill/skills/my-custom-skill@0.1.0"
    installed_at: "2026-03-20T14:30:00Z"

system:
  platform: "darwin/arm64"
  node_version: null
  quill_version: "0.4.1"
```

---

## Field Reference — `meta`

| Field | Type | Description |
|---|---|---|
| `lock_version` | integer | Schema version of this lock file. Currently `1`. Quill will refuse to read a lock file with an unsupported version. |
| `generated_at` | ISO 8601 datetime | Timestamp of when the lock file was last written. |
| `generated_by` | string | Version of the quill CLI that wrote the file, in the form `quill/<semver>`. |
| `quill_min_version` | semver | Minimum quill CLI version that can correctly parse this lock file. |

---

## Field Reference — `resolved[]`

| Field | Type | Description |
|---|---|---|
| `name` | string | Skill name, matching the `name` field in the skill's manifest. |
| `version` | semver | Exact pinned version installed. |
| `source` | URL or `"local"` | Full URL of the registry endpoint the skill was fetched from, or `"local"` for skills installed from a local path. |
| `integrity` | hash string | Content hash of the downloaded skill archive (format: `<algorithm>:<hex>`). Verified on every `quill install` and at runtime load. |
| `manifest_hash` | hash string | SHA-256 of the `quill.manifest.yaml` file at install time. Divergence triggers a warning. |
| `system_prompt_hash` | hash string | SHA-256 of the resolved system prompt file at install time. Divergence triggers a warning. |
| `resolved_variant` | string or null | The `model_pattern` of the variant that was selected at install time, or `null` if no variant matched and the catch-all was absent. |
| `dependencies` | array | Flattened list of direct skill dependencies resolved for this skill. Each entry has `name` and `version`. |
| `install_path` | relative path | Path within the project where the skill files were written, relative to the project root. |
| `installed_at` | ISO 8601 datetime | Timestamp of when this specific skill entry was installed or last updated. |

---

## Field Reference — `system`

| Field | Type | Description |
|---|---|---|
| `platform` | string | OS and architecture string at install time (e.g. `linux/amd64`, `darwin/arm64`). |
| `node_version` | string or null | Node.js version if relevant to the skill runtime; `null` otherwise. |
| `quill_version` | string | Exact version of the quill CLI that performed the install. |

---

## Behaviors

### Creation

`quill.lock` is created or fully regenerated whenever `quill install` is run in a project that either has no lock file or has a `quill.manifest.yaml` that has changed since the last install. The resolver:

1. Reads all `dependencies.skills` entries from the project manifest.
2. Resolves transitive dependencies using the registry's dependency graph API.
3. Selects the highest version of each skill that satisfies all declared version ranges.
4. Writes the fully resolved set to `quill.lock`, replacing any previous content.

### Updates

Running `quill upgrade [skill]` re-resolves the specified skill (or all skills if no name is given) against the latest compatible versions and rewrites the affected entries in the lock file. The `generated_at` timestamp is updated for the entire file on any change.

### Divergence Detection

At runtime, `quill run` compares the on-disk state of each installed skill against the hashes recorded in the lock file. The following table describes the behavior on divergence:

| Divergence type | Detected by | Default behavior |
|---|---|---|
| `integrity` mismatch | Hash of skill archive vs. `integrity` field | Hard error; skill will not load. |
| `manifest_hash` mismatch | Hash of `quill.manifest.yaml` vs. `manifest_hash` | Warning printed to stderr; skill loads. |
| `system_prompt_hash` mismatch | Hash of system prompt file vs. `system_prompt_hash` | Warning printed to stderr; skill loads. |
| Missing skill directory | `install_path` does not exist | Hard error; user must re-run `quill install`. |
| `lock_version` unsupported | `lock_version` > CLI's max supported version | Hard error; user must upgrade quill CLI. |

Divergence checking can be disabled with `--skip-integrity` (not recommended outside of development).

---

## Hash Computation Algorithm

All hashes in the lock file use SHA-256 unless a future lock version specifies otherwise. The hex digest is lowercase.

**Skill archive integrity hash**: SHA-256 of the raw bytes of the `.tar.gz` archive downloaded from the registry, before extraction.

**Manifest hash**: SHA-256 of the UTF-8 encoded bytes of `quill.manifest.yaml` as it exists on disk after install, with no normalization applied.

**System prompt hash**: SHA-256 of the UTF-8 encoded bytes of the resolved system prompt file (the file pointed to by the matched variant's `system_prompt_file`, or the default `system_prompt.md` if no variant matched). If no system prompt file exists, the hash is the SHA-256 of an empty byte string: `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`.

---

## Git Integration Notes

- `quill.lock` **must** be committed to version control alongside `quill.manifest.yaml`.
- The `.quill/skills/` directory (the install path prefix) **must** be listed in `.gitignore`; only the lock file, not the installed artifacts, is tracked.
- On CI, run `quill install --frozen` to install from the lock file without re-resolving. This command fails if the lock file is absent or if any resolved version is no longer available in the registry.
- `quill install --frozen` will not write any changes to `quill.lock`; it is safe to run in read-only CI environments.

---

## Relationship to Other Files

| File | Relationship |
|---|---|
| `quill.manifest.yaml` | Declares version ranges; `quill.lock` pins those ranges to exact versions. |
| `evals.json` | Not referenced by the lock file directly; evals are versioned together with the skill and covered by the `integrity` hash. |
| `~/.quill/config.yaml` | User-level configuration (default model, signal opt-in, etc.). Not project-scoped; not referenced in the lock file. |
| `.quill/skills/` | The on-disk install location. Contents are verified against hashes in the lock file at runtime. |
