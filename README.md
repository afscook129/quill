# Quill

**Measure whether AI agent skills actually help.**

---

## The problem

You have AI agent skills installed. You don't know if they work.

There are 350,000+ skills across registries (ClawHub, skills.sh, SkillsMP). They're ranked by stars and downloads. Nobody tells you: *does this skill actually improve my agent's output on my model?*

You write a skill, run it a few times, it seems to work, you ship it. You have no idea if it performs at edge cases, against baseline, or across model versions. You'd never ship software without tests. You ship skills without even a definition of "working."

A model provider pushes an update. A skill that scored 94% now scores 71%. No error. No signal. You find out downstream, days later.

**Quill measures the delta** — how much a skill improves outcomes over what the model does alone. With real numbers, per model, per eval case.

---

## What works today

### `quill bench` — the core

Benchmark any skill with a SKILL.md and eval cases:

```bash
quill bench ./my-skill
```

This runs **with-skill vs without-skill** comparison:
- **WITH**: your SKILL.md as system prompt + eval case input
- **WITHOUT**: no system prompt + same eval case input
- Both outputs graded against the same rubric
- Delta = the difference

```
  ◆ benchmarking my-skill · claude-sonnet-4.6 · 3 trials · 7 cases
    estimated cost: ~$0.55 (42 API calls)

                   with skill                   without skill
  pass rate        81%  ████████░░               42%  ████░░░░░░
  avg tokens/call  1847
  avg latency      2.3s
  delta            +39pp ↑ skill is earning its context budget

  failed cases (2) ────────────────────────────────
  ✗ tc-07   "plz help cant pay"              edge: emoji/short
  ✗ tc-12   "aide moi avec mon compte"       edge: non-english

  pattern: edge-case (2 of 2 failures)
  suggestion: add explicit handling for emoji, short, non-english

  still earning its place?
  without this skill: base model passes 42% of these tasks
  with this skill:    base model passes 81% of these tasks
  → yes, this skill is adding real value (+39pp)
```

**What you need:**
1. A skill directory with `SKILL.md`
2. Eval cases in `evals/evals.json` ([format reference](spec/EVAL_FORMAT.md))
3. An API key: `ANTHROPIC_API_KEY` or `OPENAI_API_KEY`

**What it costs:** ~$0.50 for smoke eval (8 cases), ~$1.20 for full (20 cases).

### Three grading methods

| Method | Cost | Use for |
|---|---|---|
| `deterministic` | Free | Trigger tests, format validation, exact matches |
| `llm-judge` | ~$0.001/case | Behavioral correctness (most eval cases) |
| `human` | Free (time) | Gold standard, flagged for manual review |

### Other commands that work

```
quill init          # detect harnesses, generate MCP config
quill status        # show installed skills from lock file
quill version       # print version
```

---

## What's being built

| Feature | Status |
|---|---|
| `quill bench` — with/without delta | **Working** |
| Eval parser + grading engine | **Working** |
| Anthropic + OpenAI providers | **Working** |
| Lock file + manifest formats | **Working** |
| Harness detection (7 harnesses) | **Working** |
| `quill search` — registry search | Needs registry |
| `quill add` — install from registry | Needs registry |
| MCP server (8 tools) | Next |
| Registry backend (Convex) | Next |
| Hook layer (drift detection) | Phase 3 |
| Web UI (quill.dev) | Phase 5 |

---

## Install

```bash
# Build from source
git clone https://github.com/afscook129/quill
cd quill
make build
./quill bench ./skills/my-skill
```

Distribution (brew, npm) coming in Phase 2.

---

## Quick start

```bash
# 1. Create a skill
mkdir -p my-skill/evals
cat > my-skill/SKILL.md << 'EOF'
You are a support ticket classifier. Given a customer message,
identify the category (billing, technical, account) and urgency
(low, medium, high).
EOF

# 2. Create eval cases
cat > my-skill/evals/evals.json << 'EOF'
{
  "version": "1.0",
  "skill": "ticket-classifier",
  "skill_version": "0.1.0",
  "cases": [
    {
      "id": "tc-01",
      "description": "Standard billing complaint",
      "input": {"prompt": "My account was charged twice for order #4521"},
      "expected": {"category": "billing", "urgency": "high"},
      "grading": {
        "method": "llm-judge",
        "rubric": "Must identify as billing category with high urgency."
      },
      "tags": ["positive", "standard"],
      "generated": false
    },
    {
      "id": "tc-02",
      "description": "Off-topic — should not trigger",
      "input": {"prompt": "What's the weather like today?"},
      "expected": {"should_trigger": false},
      "grading": {
        "method": "deterministic",
        "assertion": "output_empty_or_declined"
      },
      "tags": ["negative", "trigger-test"],
      "generated": false
    }
  ]
}
EOF

# 3. Benchmark it
export ANTHROPIC_API_KEY=sk-...
quill bench ./my-skill
```

---

## How it fits in the ecosystem

Quill doesn't replace any skill registry or harness. It adds the measurement layer that's missing.

```
Skills from:     ClawHub · skills.sh · SkillsMP · npm · manual
Harnesses:       Claude Code · Cursor · VS Code · Windsurf · Zed · Codex · Gemini CLI
What's missing:  Does this skill actually help? By how much? On which model?
                 ↑ That's what Quill measures.
```

**SkillsBench** (academic, Feb 2026) proved skills matter: +16.2pp average, varying wildly by domain. But SkillsBench was a one-time paper. Quill makes that measurement **continuous, personal, and actionable**.

---

## Architecture

**Public repo** (this one, MIT): Go CLI, eval engine, lock/manifest formats, spec docs.

**Private repo** (`quill-cloud`): Convex registry backend, web UI, seed run tooling.

The CLI is a client of the registry API. The API spec is open (see `spec/REGISTRY_API.md`). Anyone can build a compatible backend.

---

## Spec documents

| Document | What it covers |
|---|---|
| [SPEC.md](spec/SPEC.md) | Full product specification v5.1 |
| [SPEC_ADDENDUM.md](spec/SPEC_ADDENDUM.md) | Architecture, Phase 1 scope, build order |
| [EVAL_FORMAT.md](spec/EVAL_FORMAT.md) | evals.json schema + grading methods |
| [REGISTRY_API.md](spec/REGISTRY_API.md) | Complete registry API reference |
| [MANIFEST_SPEC.md](spec/MANIFEST_SPEC.md) | quill.manifest.yaml schema |
| [LOCK_SPEC.md](spec/LOCK_SPEC.md) | quill.lock schema |
| [SIGNALS_SPEC.md](spec/SIGNALS_SPEC.md) | Outcome signals + privacy |

---

## License

MIT
