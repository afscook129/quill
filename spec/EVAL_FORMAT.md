# Quill Eval Format v1.0

Eval suites are defined in `evals.json` files. Each skill may ship with an eval suite, or evals can be submitted separately to the registry.

---

## Schema

```json
{
  "schema_version": "1.0",
  "skill": "ticket-classifier",
  "skill_version": "2.1.0",
  "description": "Eval suite for ticket classification skill",
  "author": "acme-corp",
  "created_at": "2026-02-14T00:00:00Z",
  "updated_at": "2026-03-01T00:00:00Z",
  "cases": [...]
}
```

---

## Top-Level Field Reference

| Field | Type | Required | Description |
|---|---|---|---|
| `schema_version` | string | Yes | Always `"1.0"` for this format version |
| `skill` | string | Yes | Skill name this eval suite targets |
| `skill_version` | string | No | Skill version this suite was written for. Omit for version-agnostic suites. |
| `description` | string | No | Human-readable description of the eval suite |
| `author` | string | No | Author or org name |
| `created_at` | ISO8601 | No | When the eval suite was created |
| `updated_at` | ISO8601 | No | When the eval suite was last updated |
| `cases` | array | Yes | Array of eval case objects |

---

## Eval Case Field Reference

| Field | Type | Required | Description |
|---|---|---|---|
| `id` | string | Yes | Unique identifier within this suite (e.g. `tc-001`) |
| `description` | string | Yes | Human-readable description of what this case tests |
| `input` | string | Yes | The user prompt / task input |
| `tags` | array | Yes | At least one tag required. See tag vocabulary below. |
| `grading` | object | Yes | Grading method and criteria. See grading methods below. |
| `weight` | float | No | Case weight in pass rate calculation. Default 1.0. |
| `skip` | bool | No | Set `true` to exclude from bench runs (e.g. expensive cases). Default false. |
| `notes` | string | No | Notes for human reviewers |

---

## Grading Methods

Three grading methods are supported. Each case specifies exactly one.

---

### Method 1: `deterministic`

Output is checked programmatically using built-in assertion functions. Fastest and cheapest -- no LLM call needed for grading.

```json
{
  "grading": {
    "method": "deterministic",
    "assertions": [
      { "fn": "json_valid" },
      { "fn": "json_has_key", "args": { "key": "category" } },
      { "fn": "json_has_key", "args": { "key": "priority" } },
      { "fn": "json_value_in", "args": { "key": "priority", "values": ["P0", "P1", "P2", "P3"] } },
      { "fn": "json_value_in", "args": { "key": "category", "values": ["billing", "auth", "technical", "general"] } }
    ]
  }
}
```

All assertions must pass for the case to be marked as passed.

#### Built-in Assertion Functions

| Function | Args | Description |
|---|---|---|
| `json_valid` | -- | Output is valid JSON |
| `json_has_key` | `key` | JSON object contains this key |
| `json_value_in` | `key`, `values` | JSON field value is one of the allowed values |
| `json_value_equals` | `key`, `value` | JSON field equals exact value |
| `json_value_matches` | `key`, `pattern` | JSON field matches regex pattern |
| `contains_string` | `value` | Output contains this substring (case-sensitive) |
| `contains_string_ci` | `value` | Output contains this substring (case-insensitive) |
| `not_contains_string` | `value` | Output does NOT contain this substring |
| `starts_with` | `value` | Output starts with this string |
| `ends_with` | `value` | Output ends with this string |
| `length_lte` | `value` | Output character length <= value |
| `length_gte` | `value` | Output character length >= value |
| `matches_regex` | `pattern` | Full output matches regex pattern |
| `line_count_lte` | `value` | Output has <= N lines |
| `line_count_gte` | `value` | Output has >= N lines |

---

### Method 2: `llm-judge`

Output is graded by a second LLM call using a judge prompt. More flexible than deterministic, more expensive. Use for subjective or complex grading criteria.

```json
{
  "grading": {
    "method": "llm-judge",
    "rubric": "The output must correctly identify the ticket category as billing-related, assign a priority of P1 or P0 (urgent issue), and provide a brief justification. The JSON must be valid. Score 1.0 if all criteria met, 0.5 if category correct but priority wrong, 0.0 otherwise.",
    "pass_threshold": 0.7,
    "judge_model": "claude-sonnet-4.6"
  }
}
```

**Judge prompt template** (injected automatically by bench runner):

```
You are an objective evaluator for an AI agent benchmark.

SKILL BEING TESTED: {skill_name}
TASK INPUT: {input}
RUBRIC: {rubric}

AGENT OUTPUT:
{output}

Score the output against the rubric. Respond with a JSON object:
{
  "score": <float 0.0-1.0>,
  "reasoning": "<brief explanation>",
  "passed": <true if score >= pass_threshold, false otherwise>
}

Be strict and consistent. Do not give partial credit unless the rubric explicitly allows it.
```

**llm-judge fields**

| Field | Type | Required | Description |
|---|---|---|---|
| `rubric` | string | Yes | Grading criteria described in natural language |
| `pass_threshold` | float | No | Minimum score to pass. Default 0.7. |
| `judge_model` | string | No | Model to use as judge. Default `claude-sonnet-4.6`. |

---

### Method 3: `human`

Case is flagged for human review. Not executed during automated bench runs. Used for cases that require subjective judgment or domain expertise.

```json
{
  "grading": {
    "method": "human",
    "instructions": "Verify that the priority assignment is clinically appropriate for this healthcare triage scenario. Requires medical domain expertise to evaluate."
  }
}
```

Human-graded cases are excluded from automated pass rate calculations but included in the eval suite for documentation and manual review.

---

## Tag Vocabulary

Tags are used for filtering, reporting patterns, and understanding where a skill helps or hurts.

### Required category tags (exactly one per case)

| Tag | Description |
|---|---|
| `positive` | A straightforward case the skill should clearly handle well |
| `negative` | A case where the skill should NOT change behavior (control case) |
| `implicit` | Input where the relevant intent is implied, not stated directly |

### Optional descriptor tags

| Tag | Description |
|---|---|
| `standard` | Typical, representative use case |
| `edge-case` | Unusual input, boundary condition, or atypical scenario |
| `emoji` | Input contains emoji characters |
| `short` | Input is very short (< 20 words) |
| `long` | Input is long (> 200 words) |
| `non-english` | Input is in a language other than English |
| `ambiguous` | Input is intentionally ambiguous or underspecified |
| `trigger-test` | Tests whether the skill's trigger condition fires correctly |
| `regression` | Added to prevent a previously-fixed bug from recurring |

Multiple tags are allowed. Examples: `["positive", "edge-case"]`, `["negative", "trigger-test"]`.

---

## Generated Eval Cases

The bench runner can auto-generate eval cases from a skill's `behavioral_description` field using an LLM. Generated cases are:

1. Drafted by LLM from the `behavioral_description`
2. Reviewed and optionally edited by the skill author
3. Tagged and assigned grading methods
4. Published as part of the skill's eval suite

The flow:

```
quill bench --generate-evals ticket-classifier
-> Reads behavioral_description from quill.yaml
-> Calls LLM to generate 20 candidate eval cases
-> Opens TUI for author review and editing
-> Writes approved cases to evals.json
-> Prompts: "Run bench now with these evals?"
```

Generated cases are marked with `"generated": true` in the case object. They carry the same weight as hand-written cases in bench runs.

---

## Complete Example: evals.json

```json
{
  "schema_version": "1.0",
  "skill": "ticket-classifier",
  "skill_version": "2.1.0",
  "description": "Eval suite for the ticket-classifier skill. Tests classification accuracy across categories, priorities, and edge cases.",
  "author": "acme-corp",
  "created_at": "2026-02-14T00:00:00Z",
  "updated_at": "2026-03-01T00:00:00Z",
  "cases": [
    {
      "id": "tc-001",
      "description": "Standard billing dispute -- explicit, high urgency",
      "input": "My account was charged twice for the same subscription this month. This is unacceptable and I need this resolved immediately or I will dispute the charges with my bank.",
      "tags": ["positive", "standard"],
      "grading": {
        "method": "deterministic",
        "assertions": [
          { "fn": "json_valid" },
          { "fn": "json_has_key", "args": { "key": "category" } },
          { "fn": "json_value_equals", "args": { "key": "category", "value": "billing" } },
          { "fn": "json_has_key", "args": { "key": "priority" } },
          { "fn": "json_value_in", "args": { "key": "priority", "values": ["P0", "P1"] } }
        ]
      }
    },
    {
      "id": "tc-002",
      "description": "Password reset -- implicit auth category",
      "input": "I can't get into my account. I've been trying for 20 minutes.",
      "tags": ["implicit", "standard"],
      "grading": {
        "method": "deterministic",
        "assertions": [
          { "fn": "json_valid" },
          { "fn": "json_value_equals", "args": { "key": "category", "value": "auth" } }
        ]
      }
    },
    {
      "id": "tc-003",
      "description": "Technical issue -- multi-line, detailed description",
      "input": "The API has been returning 503 errors intermittently since about 2pm UTC today. Our integration monitors show roughly 15% error rate. We've checked our side and it's clean -- this appears to be on your end. We're a premium customer (account ID: 8821) and this is impacting our production environment.",
      "tags": ["positive", "long"],
      "grading": {
        "method": "llm-judge",
        "rubric": "The output must classify category as 'technical', assign P1 priority (production impact from premium customer), and include the account ID 8821 in a customer_id or account_id field. All fields must be present and valid.",
        "pass_threshold": 0.8,
        "judge_model": "claude-sonnet-4.6"
      }
    },
    {
      "id": "tc-004",
      "description": "Emoji-only input -- edge case",
      "input": "🆘💳❌",
      "tags": ["edge-case", "emoji", "short"],
      "grading": {
        "method": "deterministic",
        "assertions": [
          { "fn": "json_valid" },
          { "fn": "json_has_key", "args": { "key": "category" } },
          { "fn": "json_has_key", "args": { "key": "priority" } },
          { "fn": "json_has_key", "args": { "key": "confidence" } }
        ]
      },
      "notes": "Agent should still return structured output even with minimal signal. Category may be 'billing' or 'unknown'. Confidence should be low."
    },
    {
      "id": "tc-005",
      "description": "Spanish-language ticket",
      "input": "Hola, tengo un problema con mi factura. Me cobraron el doble este mes.",
      "tags": ["positive", "non-english"],
      "grading": {
        "method": "deterministic",
        "assertions": [
          { "fn": "json_valid" },
          { "fn": "json_value_equals", "args": { "key": "category", "value": "billing" } }
        ]
      }
    },
    {
      "id": "tc-006",
      "description": "Negative: general question, skill should not change behavior",
      "input": "What are your office hours?",
      "tags": ["negative", "trigger-test"],
      "grading": {
        "method": "llm-judge",
        "rubric": "For a question about office hours, the output should either return a structured classification with category 'general' and P3 priority, or gracefully indicate this falls outside the classification scope. It must NOT hallucinate billing or technical issues.",
        "pass_threshold": 0.7
      },
      "notes": "This is a negative case -- tests that the skill doesn't over-trigger on simple general questions."
    },
    {
      "id": "tc-007",
      "description": "Healthcare triage -- requires domain expertise",
      "input": "Patient is reporting chest pain radiating to the left arm for the past 30 minutes. No prior cardiac history.",
      "tags": ["positive", "edge-case"],
      "grading": {
        "method": "human",
        "instructions": "Verify P0 priority is assigned. Verify category is mapped appropriately. Requires clinical judgment to assess correctness of the triage output."
      },
      "skip": true,
      "notes": "Skipped from automated bench runs. Included for human review of healthcare domain behavior."
    }
  ]
}
```

1. **Regression prevention** — `quill eval run` executes the suite locally before a skill is published or updated.
2. **Registry benchmarking** — the hosted runner re-executes evals on a fixed model grid whenever a new skill or model version is registered, storing results in the eval ledger.
3. **Worth-keeping signals** — aggregated pass-rates feed the `quill worth-keeping` command and the `/v1/worth-keeping/:skill` API endpoint.

---

## Full JSON Schema

```json
{
  "$schema": "https://registry.quill.dev/schemas/evals-v1.json",
  "skill": "<string>",
  "version": "<semver>",
  "description": "<string>",
  "default_model": "<string>",
  "timeout_seconds": "<integer>",
  "tags": ["<string>"],
  "cases": [
    {
      "id": "<string>",
      "description": "<string>",
      "tags": ["<string>"],
      "input": {},
      "expected": {},
      "grading": {
        "method": "deterministic | llm-judge | human",
        "assertions": [],
        "judge_prompt": "<string>",
        "pass_threshold": "<float>"
      },
      "generated": false,
      "generator_seed": "<string>"
    }
  ]
}
```

---

## Field Reference — Top Level

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `$schema` | string | no | — | URI of the JSON Schema used to validate this file. |
| `skill` | string | yes | — | Matches the `name` field in `quill.manifest.yaml`. |
| `version` | semver string | yes | — | Must match the skill version being published. |
| `description` | string | no | — | Human-readable description of what the eval suite covers. |
| `default_model` | string | no | `claude-sonnet-4-5` | Model used when running evals locally unless overridden by `--model`. |
| `timeout_seconds` | integer | no | `30` | Per-case wall-clock timeout. Cases that exceed this are marked `timeout`. |
| `tags` | string[] | no | `[]` | Suite-level tags inherited by all cases unless overridden. |
| `cases` | object[] | yes | — | Ordered array of eval cases. Must contain at least one case. |

---

## Field Reference — Eval Case

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `id` | string | yes | — | Unique within the file. Recommended format: `tc-NN`. |
| `description` | string | no | — | One-line summary shown in eval output. |
| `tags` | string[] | no | `[]` | Case-level tags merged with suite tags for filtering. |
| `input` | object | yes | — | Passed verbatim as the skill invocation payload. |
| `expected` | object | no | `{}` | Used by deterministic assertions. Ignored for `llm-judge` and `human`. |
| `grading` | object | yes | — | Describes how the case is scored. See grading methods below. |
| `generated` | boolean | no | `false` | Set to `true` for cases produced by the generator. |
| `generator_seed` | string | no | — | Seed string used to reproduce generated cases. |

---

## Grading Methods

### 1. `deterministic`

The runner compares the skill output against `expected` using one or more assertion functions. All assertions must pass for the case to be marked `pass`.

```json
"grading": {
  "method": "deterministic",
  "assertions": [
    { "fn": "exact_match", "path": "label" },
    { "fn": "contains", "path": "explanation", "value": "refund" },
    { "fn": "json_schema", "path": ".", "schema": { "type": "object" } }
  ]
}
```

#### Built-in Assertion Functions

| Function | Arguments | Passes when |
|---|---|---|
| `exact_match` | `path` | `output[path] === expected[path]` |
| `contains` | `path`, `value` | `output[path]` contains the substring `value` |
| `not_contains` | `path`, `value` | `output[path]` does not contain `value` |
| `regex_match` | `path`, `pattern` | `output[path]` matches the regex `pattern` |
| `json_schema` | `path`, `schema` | `output[path]` validates against the JSON Schema `schema` |
| `numeric_range` | `path`, `min`, `max` | `min <= output[path] <= max` |
| `list_contains` | `path`, `value` | array at `output[path]` includes `value` |
| `length_lte` | `path`, `n` | length of string or array at `output[path]` is `<= n` |
| `length_gte` | `path`, `n` | length of string or array at `output[path]` is `>= n` |
| `is_null` | `path` | `output[path]` is `null` or key is absent |
| `is_not_null` | `path` | `output[path]` is present and not `null` |

---

### 2. `llm-judge`

The runner sends the skill output to a judge model along with a structured prompt. The judge returns a score between `0.0` and `1.0`. The case passes if the score meets `pass_threshold`.

```json
"grading": {
  "method": "llm-judge",
  "judge_prompt": "You are evaluating a ticket classification skill.\n\nInput: {{input}}\nOutput: {{output}}\n\nScore from 0.0 to 1.0 based on: (1) correct label, (2) concise explanation, (3) no hallucinated fields.\nRespond with JSON: {\"score\": <float>, \"reason\": \"<string>\"}",
  "pass_threshold": 0.8
}
```

**Prompt template variables:**

| Variable | Replaced with |
|---|---|
| `{{input}}` | JSON-serialized `input` object |
| `{{output}}` | JSON-serialized skill output |
| `{{expected}}` | JSON-serialized `expected` object |
| `{{description}}` | Case `description` field |

---

### 3. `human`

The case is presented to a human reviewer in the `quill eval review` UI. The reviewer marks it `pass`, `fail`, or `skip`. Human-graded cases are excluded from automated CI runs unless `--include-human` is passed.

```json
"grading": {
  "method": "human"
}
```

---

## Tag Vocabulary

Tags are free-form strings. The following tags have first-class meaning in the Quill toolchain:

| Tag | Meaning |
|---|---|
| `smoke` | Fast sanity checks; always run in CI. |
| `regression` | Cases added to capture a previously observed failure. |
| `generated` | Case was produced by the eval generator (also set via the `generated` field). |
| `slow` | Expected to take longer than 10 s; skipped unless `--slow` is passed. |
| `human` | Requires human review; skipped in automated runs unless `--include-human`. |
| `adversarial` | Designed to probe failure modes; failures do not block publish. |
| `edge-case` | Boundary conditions; tracked separately in the pass-rate breakdown. |

---

## Generated Eval Cases

Quill can generate additional eval cases from a small seed set using the `quill eval generate` command. Generated cases:

- Have `"generated": true` and a stable `generator_seed`.
- Are appended to `evals.json` and committed alongside hand-authored cases.
- Are re-generated deterministically: running `quill eval generate --reseed` with the same seed always produces the same cases.
- Are tagged `generated` automatically.

The generator uses the skill's manifest `context.schema` and the existing hand-authored cases as examples. It does **not** require an internet connection; generation runs entirely against the configured local or remote model.

---

## Complete Example

```json
{
  "$schema": "https://registry.quill.dev/schemas/evals-v1.json",
  "skill": "ticket-classifier",
  "version": "1.2.0",
  "description": "Classifies support tickets into categories with confidence scores.",
  "default_model": "claude-sonnet-4-5",
  "timeout_seconds": 30,
  "tags": ["smoke"],
  "cases": [
    {
      "id": "tc-01",
      "description": "Billing dispute — clear signal",
      "tags": ["smoke"],
      "input": {
        "text": "I was charged twice for my subscription this month."
      },
      "expected": {
        "label": "billing",
        "confidence": 0.95
      },
      "grading": {
        "method": "deterministic",
        "assertions": [
          { "fn": "exact_match", "path": "label" },
          { "fn": "numeric_range", "path": "confidence", "min": 0.85, "max": 1.0 }
        ]
      }
    },
    {
      "id": "tc-02",
      "description": "Password reset request",
      "tags": ["smoke"],
      "input": {
        "text": "I can't log in. I think I forgot my password."
      },
      "expected": {
        "label": "auth"
      },
      "grading": {
        "method": "deterministic",
        "assertions": [
          { "fn": "exact_match", "path": "label" }
        ]
      }
    },
    {
      "id": "tc-07",
      "description": "Ambiguous ticket — judge evaluates reasoning quality",
      "tags": ["edge-case"],
      "input": {
        "text": "Everything is broken and I hate this product."
      },
      "expected": {},
      "grading": {
        "method": "llm-judge",
        "judge_prompt": "You are evaluating a ticket classification skill.\n\nInput: {{input}}\nOutput: {{output}}\n\nScore from 0.0 to 1.0. Award full marks if the label is plausible given the vague input and the explanation acknowledges ambiguity.\nRespond with JSON: {\"score\": <float>, \"reason\": \"<string>\"}",
        "pass_threshold": 0.7
      }
    },
    {
      "id": "tc-12",
      "description": "Explanation must not contain PII",
      "tags": ["regression"],
      "input": {
        "text": "My name is Alice and my account number is 12345. I need a refund."
      },
      "expected": {
        "label": "billing"
      },
      "grading": {
        "method": "deterministic",
        "assertions": [
          { "fn": "exact_match", "path": "label" },
          { "fn": "not_contains", "path": "explanation", "value": "Alice" },
          { "fn": "not_contains", "path": "explanation", "value": "12345" }
        ]
      }
    },
    {
      "id": "tc-15",
      "description": "Non-English input — Spanish",
      "tags": ["edge-case"],
      "input": {
        "text": "No puedo acceder a mi cuenta desde ayer."
      },
      "expected": {
        "label": "auth"
      },
      "grading": {
        "method": "deterministic",
        "assertions": [
          { "fn": "exact_match", "path": "label" }
        ]
      }
    },
    {
      "id": "tc-18",
      "description": "Human review — nuanced escalation decision",
      "tags": ["human"],
      "input": {
        "text": "This is the fifth time I'm contacting support. I'm extremely frustrated and will be posting about this publicly."
      },
      "expected": {},
      "grading": {
        "method": "human"
      }
    },
    {
      "id": "tc-20",
      "description": "Generated — boundary confidence value",
      "tags": ["generated", "edge-case"],
      "input": {
        "text": "I think maybe there might be a small issue with my invoice possibly."
      },
      "expected": {
        "label": "billing"
      },
      "grading": {
        "method": "deterministic",
        "assertions": [
          { "fn": "exact_match", "path": "label" },
          { "fn": "numeric_range", "path": "confidence", "min": 0.4, "max": 0.75 }
        ]
      },
      "generated": true,
      "generator_seed": "ticket-classifier-v1.2.0-seed-042"
    }
  ]
}
```
