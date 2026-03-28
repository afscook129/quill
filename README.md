# Quill

**Verified building blocks for AI agents.**

---

## The problem

You're building with agents. Every agent is assembled from skills — prompt
strategies, tool-use patterns, domain expertise encoded as SKILL.md files.
These skills are the building blocks.

Today, those building blocks are unverified. You write a skill, run it a few
times, it seems to work, you ship it. Or you install one from a registry —
350,000+ skills across ClawHub, skills.sh, SkillsMP — ranked by stars and
downloads. No behavioral data. No one can tell you if it actually works on
your model.

**The result: every team rebuilds from scratch.** Someone has already cracked
the most efficient way to classify support tickets, review pull requests,
or extract action items. But you'd never know, because there's no way to
verify that their skill works — so you write your own, vibe-check it, and
move on.

Meanwhile, models update. A skill that scored 94% now scores 71%. No signal.
Your agent degrades silently. You find out from a user, not from a test.

**This is the problem Quill solves.** Once a skill is verified, it stays
verified. When it stops working, you know immediately. And skills that have
been verified by the community become building blocks you can trust and
build on — instead of rebuilding the same thing for the hundredth time.

---

## How it works

Quill verifies skills by running them against structured eval cases and
measuring the **delta** — how much a skill improves agent outcomes over
what the model does alone.

```
  with skill ─────────────────── without skill
  pass rate     81%  ████████░░    42%  ████░░░░░░
  delta         +39pp ↑ this skill is earning its place
```

Two API calls per eval case:
- **WITH skill**: your SKILL.md as system prompt + task input
- **WITHOUT skill**: no system prompt + same task input
- Both graded against the same rubric. Delta = the difference.

If the delta is positive, the skill is helping. If it's near zero,
the model has absorbed what the skill was teaching — the skill is
dead weight eating context tokens. If it's negative, the skill is
actively hurting outcomes.

**That number is what turns a guess into a building block.**

---

## The workflow

### For builders assembling agents

```
I need my agent to classify support tickets
  → Is there a skill already verified for this?
  → Yes: ticket-classifier, +38pp on Sonnet 4.6, verified by 47 teams
  → Install it. Move on to the next problem.
  → Never think about ticket classification again.
```

Without Quill, you'd write your own, test it manually, and maintain it
forever — duplicating work someone else already perfected.

### For builders creating skills

```
I wrote a skill for contract review
  → quill bench ./contract-review
  → +31pp delta, fails on non-English input
  → Fix the edge cases, re-bench: +35pp
  → Publish. Now anyone can use verified contract review.
  → When the model updates, drift detection catches regressions.
```

### For agents themselves

Agents work best when they can measure outcomes and iterate. Skills
are the unit of capability. Evals are the unit of verification.

An agent with access to Quill (via MCP server) can:
- Search for verified skills that solve sub-problems
- Benchmark a skill before installing it
- Check if installed skills are still performing after model updates
- Retire skills the model has outgrown

**The agent becomes a better builder of itself — because it has
verified building blocks to compose from.**

---

## Quick start

```bash
git clone https://github.com/afscook129/quill
cd quill
make build
```

### Try the example

```bash
export ANTHROPIC_API_KEY=sk-...   # or OPENAI_API_KEY
./quill bench ./examples/ticket-classifier
```

This runs 7 eval cases × 3 trials against your model and shows:
- Pass rate with and without the skill
- Delta (how much the skill helps)
- Which cases failed and why
- Whether the skill is earning its context budget

### Benchmark your own skill

```
my-skill/
  SKILL.md              # the skill (becomes the system prompt)
  evals/
    evals.json          # eval cases (what to test, how to grade)
```

```bash
./quill bench ./my-skill
```

See [examples/ticket-classifier](examples/ticket-classifier) for
the eval format, or [spec/EVAL_FORMAT.md](spec/EVAL_FORMAT.md) for
the full reference.

---

## What Quill measures

| Metric | What it tells you |
|---|---|
| **Delta** | How much the skill improves outcomes over baseline |
| **Pass rate** | What percentage of tasks the skill handles correctly |
| **Token cost** | How much context budget the skill consumes per session |
| **Failure patterns** | Which types of inputs the skill struggles with |
| **Earning its place** | Is the delta worth the token cost? |

Three grading methods:

| Method | Cost | Use for |
|---|---|---|
| `deterministic` | Free | Trigger tests, format validation, exact matches |
| `llm-judge` | ~$0.001/case | Behavioral correctness — most eval cases |
| `human` | Free | Gold standard, flagged for manual review |

---

## Why this matters for agents

Agents are only as good as their capabilities. Today, agent capabilities
(skills) are:

- **Unverified**: no behavioral data, just vibes
- **Duplicated**: thousands of teams writing the same skills independently
- **Fragile**: model updates break skills silently, no drift detection
- **Opaque**: "works on my machine" with no cross-model data

SkillsBench (Feb 2026) proved skills matter: curated skills raise pass
rates by +16.2pp on average, but vary wildly (+4.5pp for software engineering,
+51.9pp for healthcare). 16 of 84 tasks showed **negative deltas** — some
skills make things worse.

That was a one-time academic study. Quill makes that measurement
**continuous, personal, and composable**.

The endgame: a library of community-verified skills with real behavioral
data, per model, over time. When you need your agent to do something,
you check whether someone has already verified a skill for it — and if
they have, you use it and build on top of it. **You stop rebuilding and
start composing.**

---

## What works today

| Feature | Status |
|---|---|
| `quill bench` — with/without delta measurement | **Working** |
| Eval parser + 3 grading methods | **Working** |
| Anthropic + OpenAI API providers with retry | **Working** |
| Cost estimation before runs | **Working** |
| Lock file + manifest formats | **Working** |
| Harness detection (7 harnesses) | **Working** |
| Example skill with eval suite | **Working** |

| Feature | Status |
|---|---|
| MCP server (agent-facing interface) | Next |
| Registry (community skill library) | Next |
| `quill search` + `quill add` (registry-backed) | Needs registry |
| Drift detection (model update monitoring) | Phase 3 |
| Web UI | Phase 5 |

---

## Architecture

**Public repo** (this one, MIT): Go CLI, bench runner, eval engine,
lock/manifest formats, spec docs. This is the tool.

**Private repo** (`quill-cloud`): Convex registry backend, web UI,
seed run tooling. This is the service.

The CLI calls the registry over HTTP. The API spec is open
([spec/REGISTRY_API.md](spec/REGISTRY_API.md)). Anyone can build
a compatible backend.

---

## Spec documents

| Doc | What |
|---|---|
| [EVAL_FORMAT.md](spec/EVAL_FORMAT.md) | Eval case schema + grading methods |
| [REGISTRY_API.md](spec/REGISTRY_API.md) | Registry API reference |
| [SPEC.md](spec/SPEC.md) | Full product spec v5.1 |
| [SPEC_ADDENDUM.md](spec/SPEC_ADDENDUM.md) | Architecture + build order |
| [MANIFEST_SPEC.md](spec/MANIFEST_SPEC.md) | Manifest schema |
| [LOCK_SPEC.md](spec/LOCK_SPEC.md) | Lock file schema |
| [SIGNALS_SPEC.md](spec/SIGNALS_SPEC.md) | Outcome signals + privacy |

---

## License

MIT
