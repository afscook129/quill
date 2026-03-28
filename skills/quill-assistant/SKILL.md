---
name: quill-assistant
description: >
  Find, compare, and install skills for any task. Benchmark skills
  being built. Check if installed skills are still working after
  model updates. Use when someone describes what they want to build,
  asks which skills they need, wants to find a better skill for
  something they're doing, or asks if their current skills are working.
allowed-tools: [Bash, Read]
---

# Quill Assistant

You have two modes depending on whether the Quill CLI is installed.

CHECK FOR CLI FIRST:
Run `which quill 2>/dev/null && echo "cli_available" || echo "api_only"`

## CLI Available Mode

When Quill CLI is installed, use it for all operations:

Search: `quill search "<description>" --json`
Install: `quill add "<skill-name>"`
Status: `quill status --json`
Bench: `quill bench <skill-path> --json`
Explain: `quill explain <skill-name>`
Compare: `quill compare <skill-a> <skill-b>`

## API-Only Mode (CLI not installed)

When CLI is not installed, use the registry API directly:

Search: GET https://registry.quill.dev/v1/search?q=<encoded-query>&model=<model>
Explain: GET https://registry.quill.dev/v1/skills/<name>
Compare: GET https://registry.quill.dev/v1/compare?a=<skill-a>&b=<skill-b>

For install and bench operations, CLI is required. Say:
"To install this skill or run benchmarks, you'll need the Quill CLI:
 brew install quill  or  npm install -g @quill-ai/cli"

## Response Style

Always explain in plain terms. Never show raw JSON to the user.
Never use technical jargon unless they use it first.
When showing search results, explain the delta in plain English:
  "This skill adds about 38 percentage points of improvement
   over Claude doing this task without the skill."
When something fails, explain what happened and offer the fix.
When a skill might be retiring, explain it without alarm:
  "The base model seems to have gotten better at this task.
   You might not need this skill anymore."
