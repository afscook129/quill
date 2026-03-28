# Quill — Complete Build Specification v5.1

**Working name**: Quill (will be renamed before public launch)
**Status**: Active development — Phase 1
**Last updated**: March 2026

---

## 1. Problem Statement

AI agent skills are proliferating faster than anyone can evaluate them. As of March 2026, the ecosystem has crossed 350,000+ skill packages across registries. Every agent harness supports them. Every developer is installing them.

But nobody knows which ones actually work.

Current registries rank by stars, downloads, and keyword match. They tell you nothing about whether a skill meaningfully improves agent performance on your tasks, with your model. Nobody publishes the delta.

The result: developers are flying blind. Skills that add -3pp (negative effect) stay installed. Skills that add +38pp get overlooked because they don't have many stars. And when a model provider ships an update, nobody knows which skills have drifted.

**The credibility gap**: "I think my skills are helping" vs. "I know my skills are helping, by how much, with which model, and whether that's changed."

Quill closes that gap.

---

## 2. What Quill Is

Quill is a skill intelligence layer for AI agent systems. It sits between you and the skill registries and provides what none of them provide: **behavioral delta data**.

Quill answers three questions:
1. **Which skill should I use?** — ranked by actual performance delta, not stars
2. **Is it still working?** — drift detection via continuous bench runs and production signals
3. **Is it worth keeping?** — "earning its place" analysis against token cost

Quill is:
- A CLI tool (Go binary, single install)
- An MCP server (for Claude Code, Cursor, and any MCP-compatible harness)
- A registry (indexes skills from all major sources, adds behavioral data)
- An open dataset (all bench results published under ODbL)

---

## 3. Target Personas

### 3.1 The Vibe Coder

**Profile**: Non-technical. Uses Lovable, v0, or Replit. Wants to extend their AI assistant without touching code.

**Loop**:
1. Agent suggests "I found a skill that could help with this" (via quill_search)
2. User says "yeah add it"
3. Skill is installed, evals run silently in background
4. Quill surfaces: "ticket-classifier is adding +38pp — it's working"

**Key insight**: Never sees a terminal. Interacts entirely through the agent.

### 3.2 The Developer

**Profile**: Technical. Uses Claude Code, Cursor, or Windsurf. Wants to build reliable agent systems.

**Loop**:
1. `quill search "classify support tickets"`
2. Review delta scores and model performance table
3. `quill add ticket-classifier`
4. See smoke eval results immediately
5. `quill status` shows ongoing performance

**Key insight**: Wants the data, not the hand-holding. Show me the numbers.

### 3.3 The Domain Expert

**Profile**: Healthcare, legal, finance. High stakes. Needs to know skills are accurate in their domain.

**Loop**:
1. Search for domain-specific skills
2. See domain-specific eval results (not just average delta)
3. Review human-graded eval cases
4. Add with `--strict` flag (fails install if pass rate below threshold)

**Key insight**: Domain delta matters more than average delta. +51.9pp in healthcare means something specific.

### 3.4 The Team Lead

**Profile**: Engineering manager. Wants a shared, reproducible skill configuration across the team.

**Loop**:
1. `quill init` sets up project-level config
2. `quill.lock` committed to repo
3. CI runs `quill ci` — fails build if any skill regresses
4. Team gets `quill status` in PR comments

**Key insight**: Skills should be as reproducible and testable as code dependencies.

### 3.5 The Org

**Profile**: Enterprise. Multiple teams, compliance requirements, private skills.

**Loop**:
1. Private registry deployment (Convex self-hosted)
2. Org-level skill policy: approved list, blocked list
3. WorkOS SSO, audit logs
4. Skills scoped to teams

**Key insight**: Phase 6. Not the initial focus.

---

## 4. CLI Experience

### 4.1 Core commands

```
quill init                    # detect harness, configure MCP, create manifest
quill search <query>          # semantic search with delta scores
quill add <skill>             # install skill with smoke evals
quill status                  # show installed skills + performance
quill bench <skill>           # run full bench (with/without delta)
quill compare <a> <b>         # head-to-head comparison
quill explain <skill>         # explain what a skill does and its evidence
quill fix <skill>             # diagnose and fix a regressing skill
quill retire <skill>          # remove skill + update lock file
quill upgrade [skill]         # upgrade to latest version
quill publish                 # publish skill to registry
quill login                   # GitHub OAuth login
quill mcp --serve             # start MCP server (stdio)
quill ci                      # CI mode: check for regressions, fail on threshold
```

### 4.2 Key UX principles

- **Numbers first**: Every output that matters shows a delta. Not "this skill is great" — "+38pp on claude-sonnet-4.6"
- **Fast by default**: `quill add` runs smoke evals (8 cases, ~$0.48). Full bench is opt-in.
- **Honest about negatives**: Negative deltas are shown clearly. Skills that hurt performance are surfaced.
- **Lock file is truth**: `quill.lock` is the authoritative record. Commands operate against it.

### 4.3 `quill search` output example

```
$ quill search "classify support tickets" --model claude-sonnet-4.6

  ticket-classifier@2.1.0       clawhub    +38.2pp   94% pass   1,240 tok   ★ verified
  support-triage@1.4.2          skills.sh  +22.7pp   79% pass     980 tok
  helpdesk-router@3.0.0         skillsmp   +18.1pp   74% pass   1,100 tok
  customer-intent@1.1.0         clawhub    +12.4pp   69% pass     760 tok

  4 results for "classify support tickets" | model: claude-sonnet-4.6
  Run: quill add ticket-classifier
```

### 4.4 `quill status` output example

```
$ quill status

  ticket-classifier@2.1.0    +38.2pp   stable    ✓ keep    benched 2d ago
  sentiment-analyzer@1.2.0   +19.4pp   stable    ✓ keep    benched 5d ago
  email-drafter@0.9.1         +4.1pp   ↓ drift   ⚠ review  benched 1d ago

  3 skills installed | 2 healthy | 1 needs attention
  Run: quill fix email-drafter
```

---

## 5. MCP Server

### 5.1 Overview

The Quill MCP server exposes 8 tools. It runs in two modes:

- **Local** (`quill mcp --serve`): stdio transport, started by the harness
- **Hosted** (`mcp.quill.dev/sse`): SSE transport on Fly.io, for cloud-native platforms

### 5.2 MCP tools

#### `quill_search`

Search the registry for skills matching a natural language description.

**Input**:
```json
{
  "query": "classify support tickets by category and priority",
  "model": "claude-sonnet-4.6",
  "min_delta": 10,
  "limit": 5
}
```

**Output**: Array of skills with delta, pass rate, token cost, and source.

---

#### `quill_add`

Install a skill from the registry.

**Input**:
```json
{
  "skill": "ticket-classifier",
  "version": "2.1.0",
  "run_evals": true
}
```

**Output**: Installation result including file placement path, smoke eval pass rate, and lock file update status.

---

#### `quill_status`

Return current status of all installed skills.

**Input**: `{}` (no parameters)

**Output**: Array of installed skills with bench results, verdict, and drift status.

---

#### `quill_bench`

Run a full bench run for a skill (with-skill vs. without-skill delta computation).

**Input**:
```json
{
  "skill": "ticket-classifier",
  "model": "claude-sonnet-4.6",
  "trials": 3
}
```

**Output**: Delta, pass rates, per-case breakdown, identified patterns.

---

#### `quill_fix`

Diagnose a regressing or underperforming skill and suggest fixes.

**Input**:
```json
{
  "skill": "email-drafter"
}
```

**Output**: Diagnosis (drift, version issue, conflict, etc.) and recommended action.

---

#### `quill_explain`

Return a detailed explanation of what a skill does, its evidence base, and best-fit use cases.

**Input**:
```json
{
  "skill": "ticket-classifier"
}
```

**Output**: Description, behavioral explanation, performance evidence, model-by-model table, known limitations.

---

#### `quill_compare`

Head-to-head comparison of two skills.

**Input**:
```json
{
  "a": "ticket-classifier",
  "b": "support-triage",
  "model": "claude-sonnet-4.6"
}
```

**Output**: Side-by-side metrics, verdict, composability assessment.

---

#### `quill_retire`

Remove a skill and update the lock file.

**Input**:
```json
{
  "skill": "email-drafter",
  "reason": "negative delta after model update"
}
```

**Output**: Confirmation, updated lock file state.

---

## 6. Integration Matrix

| Harness | MCP support | Hook layer | Install command | Notes |
|---|---|---|---|---|
| **Claude Code** | Yes (stdio + SSE) | PostToolUse | `quill init` | Primary target. Full support. |
| **Cursor** | Yes (MCP) | Limited | `quill init --harness cursor` | MCP config injection. |
| **VS Code + Copilot** | Yes (MCP) | No | `quill init --harness vscode` | MCP server only. No hook layer. |
| **Windsurf** | Yes (MCP) | Yes | `quill init --harness windsurf` | Cascade supports hooks. |
| **Zed** | Partial | No | Manual | Zed MCP support in progress. |
| **OpenAI Codex** | No | No | `quill init --harness codex` | CLI install only. |
| **Gemini CLI** | Yes (MCP) | No | `quill init --harness gemini` | MCP supported. |
| **Lovable** | Via hosted MCP | No | Add `mcp.quill.dev` in settings | Cloud-first. No local CLI needed. |
| **v0** | Via hosted MCP | No | Add `mcp.quill.dev` in settings | Same as Lovable. |
| **Replit** | Via hosted MCP | No | Add `mcp.quill.dev` in settings | Same as Lovable. |

---

## 7. Manifest Format

See `spec/MANIFEST_SPEC.md` for the full manifest reference.

Skills are described by `quill.yaml`. Key fields:

```yaml
name: ticket-classifier
version: 2.1.0
description: "Classifies support tickets by category, priority, and sentiment"
behavioral_description: |
  When given a support ticket, the agent should identify the category,
  assign a priority (P0-P3), extract sentiment, and return structured JSON.
authors:
  - name: acme-corp
    github: acme-corp
license: MIT
```

The `behavioral_description` field is used for embedding generation (semantic search) and as the basis for generated eval cases.

---

## 8. Lock File

See `spec/LOCK_SPEC.md` for the full lock file reference.

`quill.lock` records exact installed versions, file hashes, and bench outcomes. It must be committed to version control.

Key principle: the lock file is the authoritative record of what is installed and how it is performing. `quill status` reads from it. CI checks against it.

---

## 9. Eval Format

See `spec/EVAL_FORMAT.md` for the full eval format reference.

Evals are defined in `evals.json`. Three grading methods:
- `deterministic`: programmatic assertions, no LLM call
- `llm-judge`: graded by a second LLM call
- `human`: flagged for manual review

The bench runner makes two API calls per eval case: one with the skill active, one without. The delta is the pass rate difference.

---

## 10. Registry

### 10.1 What the registry stores

- Skill metadata (from quill.yaml)
- Behavioral descriptions + vector embeddings
- Bench results (all historical runs)
- Eval suites
- Signals (aggregated, anonymized)
- Drift alerts

### 10.2 Search ranking

Results are ranked by a scoring formula:

```
score = (0.40 * delta_normalized)
      + (0.30 * semantic_similarity)
      + (0.20 * recency_score)
      + (0.10 * popularity_score)
```

Delta is weighted highest because it's the most actionable signal.

### 10.3 Cross-registry indexing

Quill indexes skills from all major registries:
- **ClawHub**: 13K+ skills. Scraped via ClawHub API.
- **Skills.sh**: 83K+ skills. Scraped via Skills.sh API.
- **SkillsMP**: 200K+ skills. Scraped via SkillsMP search.
- **Microsoft**: Various skill repos.
- **Direct publish**: Skills published directly to Quill registry.

When a skill is indexed from an external source, Quill adds behavioral data (delta, pass rate, token cost) that the source registry does not provide.

---

## 11. Intelligence Layer

### 11.1 Drift detection

Drift is detected through two complementary pipelines:

**Bench-based drift**: The registry re-runs bench suites on a schedule. If a skill's pass rate drops significantly between runs (while the skill content hasn't changed), drift is flagged.

**Signal-based drift**: Production signals from installed skills feed a rolling pass-rate calculation. If production performance diverges from bench baseline, drift is flagged.

### 11.2 "Worth keeping" analysis

For each installed skill, Quill computes a verdict:

- **keep**: Delta is meaningful (+10pp or more), token cost is reasonable, no drift detected
- **review**: Delta has declined, or drift detected, or a better alternative exists
- **retire**: Negative delta, or delta < 5pp, or skill has been superseded
- **upgrade**: A newer version exists with significantly better performance

### 11.3 Composability signals

Over time, Quill learns which skills compose well together (additive effects) and which conflict (interference, redundancy). This is tracked in the registry and surfaced in search results.

---

## 12. Security Model

### 12.1 Skill integrity

Every installed skill is hash-verified against the registry's recorded hash at install time. Tampered files are detected on load.

### 12.2 Permissions

Skills declare their required permissions in `quill.yaml`. The runtime enforces these:

- `filesystem.read/write`: Path glob whitelist
- `network.outbound`: Hostname whitelist
- `shell.allow/deny`: Command prefix lists
- `context_scope`: What ambient context the skill receives

Permission escalation (a skill requesting more than declared) is blocked.

### 12.3 Signing (Phase 3)

Sigstore signing for published skills. Publishers sign with their GitHub identity. Verification on install.

### 12.4 SBOM (Phase 4)

Skills generate an SBOM on publish. The registry stores it and exposes it via API.

---

## 13. Signals

See `spec/SIGNALS_SPEC.md` for the full signals specification.

Signals are opt-in, anonymous behavioral telemetry. They are never associated with user identity or prompt content. They feed drift detection and registry improvement.

Key principle: if a user turns signals off, Quill works exactly the same. Signals are an enhancement, not a requirement.

---

## 14. Implementation Phases

### Phase 1 — Foundation (current)

**Goal**: MCP server works, registry has data, developers can search and install.

- Convex registry backend with seed data (500 skills)
- Registry API: search, read, eval submission
- MCP server: all 8 tools, stdio + hosted SSE
- `quill init` and `quill mcp --serve`
- GitHub OAuth

### Phase 2 — CLI

**Goal**: Full CLI experience without MCP.

- All CLI commands (`search`, `add`, `status`, `bench`, `compare`, `explain`, `fix`, `retire`, `upgrade`, `publish`)
- TUI with Bubbletea/Lipgloss
- Go binary distribution via GoReleaser
- Brew tap: `afscook129/homebrew-quill`
- npm wrapper: `@quill-ai/cli`
- quill-assistant SKILL.md for agent-powered usage

### Phase 3 — Hook Layer

**Goal**: Passive signal collection, drift detection.

- Claude Code PostToolUse hooks
- Signal collection and batch submission
- Drift detection pipeline
- GitHub Actions: `quill ci` step
- `quill status` in PR comments

### Phase 4 — Intelligence

**Goal**: Richer behavioral data, upgrade intelligence.

- Trigger accuracy measurement
- Bench history visualization
- Cross-skill composability analysis
- Upgrade recommendations with delta comparison

### Phase 5 — Web UI

**Goal**: Accessible to non-CLI users.

- `quill.dev` on Vercel (Next.js)
- Skill pages with performance charts
- Comparison UI
- Email/password auth for non-GitHub users
- Team features: shared manifest, org-level status

### Phase 6 — Org Scale

**Goal**: Enterprise adoption.

- Private registry deployment (self-hosted Convex)
- WorkOS SSO (SAML/OIDC)
- Org-level policy engine (approved/blocked skills)
- Audit logs
- Team billing

---

## 15. Open Source Strategy

### 15.1 What is open (MIT)

- Go CLI + MCP server (`github.com/afscook129/quill`)
- Lock file, manifest, eval formats
- quill-assistant SKILL.md
- Registry API specification
- CLI integration harness adapters

### 15.2 What is proprietary

- Convex registry backend (`github.com/afscook129/quill-cloud`)
- Web UI
- Seed run tooling
- Internal ops tooling

### 15.3 What is open data (ODbL)

- The delta dataset: all bench results, all skill performance data
- Published at `data.quill.dev`
- Open Database License — share-alike

The moat is the data. The open source CLI drives adoption. The proprietary registry is the business. The open data is the credibility.

---

## 16. Metrics and Success Criteria

### Phase 1 success

- Registry seeded with 500 skills + bench data
- MCP server live at `mcp.quill.dev`
- 10 developers actively using it in Phase 1 beta

### Phase 2 success

- CLI published on Homebrew and npm
- 100 GitHub stars
- 500 CLI installs

### Phase 3 success

- 1,000 signals/day
- First drift event detected and alerted
- GitHub Actions step used in 10+ repos

### 12-month vision

- 10,000 skills indexed with behavioral data
- 50,000 CLI installs
- Used by at least 3 Fortune 500 companies
- Delta dataset cited in research

---

## 17. Registry API

See `spec/REGISTRY_API.md` for the complete API reference.

Base URL: `https://api.quill.dev`

Key endpoints:
- `GET /v1/search` — semantic search with delta ranking
- `GET /v1/skills/:name` — full skill metadata + cross-model performance
- `GET /v1/evals/:skill/:model` — cached eval results
- `POST /v1/evals` — submit eval results (authenticated)
- `POST /v1/skills` — publish skill (authenticated)
- `POST /v1/signals` — submit anonymous signals
- `GET /v1/bench/:skill` — benchmark history
- `GET /v1/compare` — head-to-head comparison
- `GET /v1/worth-keeping/:skill` — worth-keeping verdict
- `GET /v1/upgrade/:skill` — upgrade recommendation

---

## 18. Competitive Positioning

See `spec/SPEC_ADDENDUM.md` section A for the full competitive analysis.

**One-line positioning**: "SkillsBench proved skills matter. Quill makes that measurement continuous."

**The gap**: Every skill registry tells you what's available. None of them tell you what works. Quill is the first infrastructure that answers "is this skill actually improving my agent's performance, right now, with my model?"

**The defensible moat**: The delta dataset. Once Quill has bench data on 10,000 skills across 5 models, that dataset is years of work to replicate. The open-data license (ODbL) means it's shareable and citable, which drives academic credibility and community contribution — but the infrastructure to keep generating new data is Quill's.
