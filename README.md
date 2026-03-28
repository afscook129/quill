# Quill

**The credibility layer for AI agent skills.**

---

AI agents have a credibility problem.

They work in demos. They fail in production. A model updates and something breaks silently. You can't explain to a stakeholder why the agent works. You can't prove it got better. The investment stalls. "AI" becomes a word that means "expensive experiment."

That's not a model problem. Models are getting better fast. It's an infrastructure problem. There's no layer that makes agent behavior **measurable**, **improvable**, or **defensible**. No way to say "this agent produces correct output 91% of the time, up from 67% before we tuned the skill stack." No CI job that catches skill regressions before they ship. No shared truth about what's actually working across a team.

Quill is that layer.

---

## How it works

Your agent uses Quill. You don't have to.

Add Quill as an MCP server and your agent in Claude Code, Cursor, VS Code, or any compatible harness can search for skills, benchmark them, install them, and monitor them — all on your behalf. You describe what you want. The agent handles the rest.

```
You: "I need something that classifies support tickets by urgency"

Agent: Found ticket-classifier — adds 38% improvement over baseline
       on your model. Want me to install it?

You: "Yes"

Agent: Installed. 94% pass rate confirmed. Your agent handles
       ticket classification much more reliably now.
```

Zero commands typed. Quill searched across registries, ranked by measured outcomes (not stars), validated with real evals, and installed. Your agent got smarter.

For developers who want direct control:

```bash
quill init                           # configures everything in 30s
quill add "describe what you need"   # semantic search → install best match
quill bench ./my-skill               # stop vibe-checking, start measuring
quill status                         # see what's working and what's drifted
quill fix                            # fix what isn't
```

---

## The loops Quill closes

**Today without Quill:** search → guess → install → works or doesn't → hope → repeat

**Vibe coder**: Describe what you want → agent finds best option → installs it → outcomes improve → stack keeps getting better. You didn't do anything. Your agent got smarter.

**Developer**: Build → bench with real numbers → ship knowing it works → model updates → CI catches regression → fix before users notice → trust accumulates. For the first time, "does this work?" has a real answer.

**Domain expert**: Encode your expertise → bench confirms it works (91%) → publish it → see it working across your team → your knowledge is infrastructure, not a document someone ignores.

**Team**: One person validates a skill → it goes in `quill.lock` → everyone has it automatically → new hire runs one command → CI catches regressions for everyone → shared truth, not shared hope.

---

## Install

### MCP server (primary — agents use this)

Add to your harness's MCP config. `quill init` does this automatically.

```json
{
  "mcpServers": {
    "quill": {
      "command": "quill",
      "args": ["mcp", "--serve"]
    }
  }
}
```

### CLI

```bash
# Homebrew (macOS/Linux)
brew tap quill-dev/tap && brew install quill

# npm / bun / pnpm
npm install -g @quill-ai/cli
bun install -g @quill-ai/cli

# Zero install — run immediately
npx @quill-ai/cli <command>
bunx @quill-ai/cli <command>

# CI/CD — downloads binary directly
curl -fsSL https://quill.dev/install.sh | bash
```

---

## Claude Code integration

Quill is built to work with Claude Code out of the box. Run `quill init` and it detects Claude Code, sets up hooks, and adds the MCP server config automatically.

**What you get:**
- **SessionStart hook** — drift check on every session. If a model update broke a skill, you know before you start working.
- **PreToolUse hook** — permission enforcement. Skills declare what they need; Quill enforces it.
- **PostToolUse hook** — anonymous outcome signals. Pass/fail per tool call, latency — never content. Makes the registry smarter for everyone.
- **MCP server** — your agent can search, install, benchmark, and monitor skills as structured tool calls.
- **quill-assistant SKILL.md** — conversational access to all Quill capabilities.

The hooks config (`quill init` generates this in `.claude/settings.json`):

```json
{
  "hooks": {
    "SessionStart": [{
      "hooks": [{
        "type": "command",
        "command": "quill hook session-start --quiet",
        "timeout": 3
      }]
    }],
    "PreToolUse": [{
      "matcher": ".*",
      "hooks": [{
        "type": "command",
        "command": "quill hook pre-tool --event \"$CLAUDE_HOOK_INPUT\"",
        "timeout": 1
      }]
    }],
    "PostToolUse": [{
      "matcher": ".*",
      "hooks": [{
        "type": "command",
        "command": "quill hook post-tool --event \"$CLAUDE_HOOK_INPUT\" --async",
        "timeout": 0
      }]
    }]
  }
}
```

---

## What Quill measures

Every skill gets a **delta** — how much it improves agent outcomes over what the model does alone. Not stars. Not downloads. Measured behavioral improvement.

```
$ quill bench ./my-skill

◆ benchmarking my-skill · claude-sonnet-4.6 · 3 trials

  with skill ────────────────────── without skill
  pass rate     81%  ████████░░      42%  ████░░░░░░
  delta         +39pp ↑ skill is earning its context budget

  failed cases (4)
  ✗ tc-07   "plz help cant pay"        edge: emoji/short
  ✗ tc-12   "aide moi avec mon compte" edge: non-english

  pattern: non-standard input (3 of 4 failures)
  suggestion: add explicit handling for emoji, short, non-english

  still earning its place?
  without: base model passes 42% of these tasks
  with:    base model passes 81% of these tasks
  → yes, this skill is adding real value (+39pp)
```

Models improve. Skills accumulate. A skill that was essential six months ago may now be dead weight — eating context tokens for nothing. Quill catches that:

```
$ quill status

  ○  code-formatter@1.0.0
     → the base model now handles this well on its own
       this skill may no longer be adding value (~890 tokens/session)
       quill retire code-formatter  to check and remove it
```

---

## Works with everything

| Interface | How |
|---|---|
| **Claude Code** | MCP + hooks + SKILL.md (full integration) |
| **Cursor** | MCP + rules file |
| **VS Code** | MCP via Copilot or Claude Code extension |
| **Windsurf** | MCP + Cascade context |
| **Zed** | MCP via context servers |
| **Codex CLI** | MCP config |
| **Gemini CLI** | MCP config |
| **Lovable / v0** | Context file + registry API |
| **Replit** | Remote MCP endpoint |

Skills from: npx skills · skillpm · bun · npm · ClawHub · skills.sh · SkillsMP · manual install

Quill doesn't replace any of them. It tells you which ones are working and keeps them that way.

---

## Three promises

**No lock-in.** Everything Quill produces is plain YAML. Stop using Quill and everything still works.

**No forced migration.** Works on skills with no manifest at all. Quill infers what it can, flags the rest, and never blocks you.

**No single registry.** Point it at any registry, including your own. `quill registry init` scaffolds a private registry you can self-host.

---

## Open data (ODbL)

Every benchmark run contributes to a public cross-model skill delta index — how much skills actually help, by model, over time. This data doesn't exist anywhere else. We're building it in the open.

Share-alike: if you build on it, keep it open too.

```
https://registry.quill.dev/v1/
```

---

## CLI reference

### Everyday

```
quill init                        Set up a project (30 seconds)
quill add <skill-or-description>  Install by name or natural language
quill status                      Health summary
quill search <query>              Find skills ranked by delta
quill fix                         Resolve issues
quill upgrade [skill]             Show available upgrades
```

### Benchmarking

```
quill bench <skill-or-path>       With-skill vs without-skill comparison
quill bench-history <skill>       Results over time
quill bench-compare <v1> <v2>     A/B between versions
quill retire <skill>              Check if still earning its place
```

### Understanding

```
quill explain <skill>             Deep dive into what a skill does
quill compare <a> <b>             Head-to-head comparison
```

### Team & CI

```
quill validate <skill>            Full eval suite for CI
quill lock                        Generate quill.lock
quill team status                 Branch divergence (git-native)
quill audit [path]                Security scan
quill migrate --from <m> --to <m> Analyze model switch
```

### Publishing

```
quill publish                     Publish to registry (Sigstore signed)
quill registry init               Scaffold private registry
quill sbom                        Security surface inventory (SPDX 2.3)
```

---

## License

MIT
