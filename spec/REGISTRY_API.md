# Quill Registry API v1

## Base URL

```
https://api.quill.dev
```

All endpoints are prefixed with `/v1/`.

---

## Authentication

**Reads**: No authentication required.

**Writes**: GitHub OAuth Bearer token or API key required.

```
Authorization: Bearer <token>
```

API keys are generated after GitHub OAuth login. Use `QUILL_PUBLISH_TOKEN` for CI/CD environments.

---

## Rate Limits

| Auth state | Limit |
|---|---|
| Unauthenticated | 100 req/min |
| Authenticated | 1,000 req/min |

Rate limit headers returned on every response:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-RateLimit-Reset: 1711234567
```

---

## Pagination

Offset-based. Default limit: 20. Max limit: 100.

```
GET /v1/search?q=...&limit=20&offset=0
```

Response envelope includes:

```json
{
  "results": [...],
  "total": 142,
  "limit": 20,
  "offset": 0
}
```

---

## Error Response Format

All errors follow this structure:

```json
{
  "error": {
    "code": "SKILL_NOT_FOUND",
    "message": "Skill 'ticket-classifier' not found in registry",
    "details": {}
  }
}
```

### Error Codes

| Code | HTTP Status | Description |
|---|---|---|
| `SKILL_NOT_FOUND` | 404 | Skill name does not exist in registry |
| `VERSION_NOT_FOUND` | 404 | Skill exists but requested version does not |
| `INVALID_REQUEST` | 400 | Missing or malformed request parameters |
| `UNAUTHORIZED` | 401 | Write operation attempted without valid auth token |
| `FORBIDDEN` | 403 | Authenticated user lacks permission for this operation |
| `CONFLICT` | 409 | Skill + version already exists (use new version to update) |
| `RATE_LIMITED` | 429 | Request rate exceeded; see `X-RateLimit-Reset` header |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## Endpoints

---

### GET /v1/search

Semantic search over the skill registry. Returns skills ranked by a scoring formula combining behavioral delta, relevance, freshness, and popularity.

**Scoring formula**:

```
score = (0.40 * delta_normalized)
      + (0.30 * semantic_similarity)
      + (0.20 * recency_score)
      + (0.10 * popularity_score)
```

Where:
- `delta_normalized` = delta for the requested model, normalized 0–1 across corpus
- `semantic_similarity` = cosine similarity of query embedding vs. skill `behavioral_description` embedding
- `recency_score` = decay function on `last_benched_at`
- `popularity_score` = log-normalized install count

**Query Parameters**

| Parameter | Type | Required | Description |
|---|---|---|---|
| `q` | string | Yes | Natural language search query |
| `model` | string | No | Model slug to rank by (e.g. `claude-sonnet-4.6`). Defaults to aggregate. |
| `min_delta` | float | No | Minimum delta (pp). Filter out low-delta skills. |
| `min_score` | float | No | Minimum composite score (0.0–1.0). |
| `verified_only` | bool | No | Only return skills with verified bench data. |
| `source` | string | No | Filter by source registry (`clawhub`, `skills.sh`, `skillsmp`). |
| `composable_with` | string | No | Return skills known to compose well with this skill name. |
| `limit` | int | No | Results per page. Default 20, max 100. |
| `offset` | int | No | Pagination offset. Default 0. |

**Example Request**

```
GET /v1/search?q=classify+support+tickets&model=claude-sonnet-4.6&min_delta=5&limit=5
```

**Example Response**

```json
{
  "results": [
    {
      "name": "ticket-classifier",
      "version": "2.1.0",
      "description": "Classifies support tickets by category, priority, and sentiment",
      "source": "clawhub",
      "score": 0.91,
      "delta": 38.2,
      "pass_rate_with": 0.94,
      "pass_rate_without": 0.56,
      "model": "claude-sonnet-4.6",
      "token_cost_avg": 1240,
      "last_benched_at": "2026-03-20T14:22:00Z",
      "install_count": 4821,
      "verified": true,
      "tags": ["classification", "support", "nlp"]
    },
    {
      "name": "support-triage",
      "version": "1.4.2",
      "description": "Triages incoming support requests with priority scoring",
      "source": "skills.sh",
      "score": 0.74,
      "delta": 22.7,
      "pass_rate_with": 0.79,
      "pass_rate_without": 0.56,
      "model": "claude-sonnet-4.6",
      "token_cost_avg": 980,
      "last_benched_at": "2026-03-18T09:10:00Z",
      "install_count": 1203,
      "verified": true,
      "tags": ["triage", "support", "priority"]
    }
  ],
  "total": 14,
  "limit": 5,
  "offset": 0
}
```

---

### GET /v1/skills/:name

Returns full skill metadata including cross-model performance table and bench history summary.

**Path Parameters**

| Parameter | Description |
|---|---|
| `name` | Skill name (e.g. `ticket-classifier`) |

**Example Request**

```
GET /v1/skills/ticket-classifier
```

**Example Response**

```json
{
  "name": "ticket-classifier",
  "latest_version": "2.1.0",
  "description": "Classifies support tickets by category, priority, and sentiment",
  "behavioral_description": "When given a support ticket, the agent should identify the ticket category from a taxonomy, assign a priority level (P0-P3), extract the primary sentiment, and return a structured classification object.",
  "source": "clawhub",
  "source_url": "https://clawhub.dev/skills/ticket-classifier",
  "author": "acme-corp",
  "license": "MIT",
  "tags": ["classification", "support", "nlp"],
  "install_count": 4821,
  "verified": true,
  "published_at": "2025-11-03T08:00:00Z",
  "updated_at": "2026-02-14T16:30:00Z",
  "performance": {
    "claude-sonnet-4.6": {
      "delta": 38.2,
      "pass_rate_with": 0.94,
      "pass_rate_without": 0.56,
      "token_cost_avg": 1240,
      "trials": 3,
      "eval_cases": 20,
      "last_benched_at": "2026-03-20T14:22:00Z",
      "trend": "stable"
    },
    "claude-opus-4": {
      "delta": 29.1,
      "pass_rate_with": 0.91,
      "pass_rate_without": 0.62,
      "token_cost_avg": 2100,
      "trials": 3,
      "eval_cases": 20,
      "last_benched_at": "2026-03-20T14:22:00Z",
      "trend": "stable"
    },
    "gpt-4o": {
      "delta": 31.5,
      "pass_rate_with": 0.88,
      "pass_rate_without": 0.57,
      "token_cost_avg": 1380,
      "trials": 3,
      "eval_cases": 20,
      "last_benched_at": "2026-03-20T14:22:00Z",
      "trend": "improving"
    }
  },
  "versions": ["1.0.0", "1.1.0", "2.0.0", "2.1.0"],
  "composable_with": ["sentiment-analyzer", "ticket-router"],
  "negative_composites": ["ticket-auto-responder"]
}
```

---

### GET /v1/skills/:name/:version

Returns metadata for a specific version of a skill. Same schema as `GET /v1/skills/:name` but pinned to the requested version, with performance data from bench runs against that exact version.

**Example Request**

```
GET /v1/skills/ticket-classifier/2.0.0
```

---

### GET /v1/evals/:skill/:model

Returns cached eval results for a skill+model combination. Includes per-case pass/fail breakdown.

**Path Parameters**

| Parameter | Description |
|---|---|
| `skill` | Skill name, optionally with version (e.g. `ticket-classifier` or `ticket-classifier@2.1.0`) |
| `model` | Model slug (e.g. `claude-sonnet-4.6`) |

**Query Parameters**

| Parameter | Type | Description |
|---|---|---|
| `version` | string | Specific version. Defaults to latest. |

**Example Request**

```
GET /v1/evals/ticket-classifier/claude-sonnet-4.6
```

**Example Response**

```json
{
  "skill": "ticket-classifier",
  "version": "2.1.0",
  "model": "claude-sonnet-4.6",
  "benched_at": "2026-03-20T14:22:00Z",
  "summary": {
    "delta": 38.2,
    "pass_rate_with": 0.94,
    "pass_rate_without": 0.56,
    "total_cases": 20,
    "passed_with": 19,
    "passed_without": 11,
    "token_cost_avg": 1240,
    "trials": 3
  },
  "cases": [
    {
      "id": "tc-001",
      "description": "Billing dispute ticket, high urgency",
      "tags": ["positive", "billing", "urgent"],
      "with_skill": {
        "passed": true,
        "score": 1.0,
        "token_cost": 1180
      },
      "without_skill": {
        "passed": false,
        "score": 0.3,
        "token_cost": 890
      }
    },
    {
      "id": "tc-002",
      "description": "Password reset request, implicit category",
      "tags": ["implicit", "auth", "edge-case"],
      "with_skill": {
        "passed": true,
        "score": 0.9,
        "token_cost": 1210
      },
      "without_skill": {
        "passed": true,
        "score": 0.7,
        "token_cost": 950
      }
    }
  ],
  "patterns": {
    "consistently_better": ["billing", "urgent", "multi-category"],
    "no_improvement": ["simple-password-reset"],
    "negative_cases": []
  }
}
```

---

### POST /v1/evals

Submit eval results for a skill. Authenticated. Results are stored and incorporated into the registry's performance data after validation.

**Authentication**: Required. Bearer token or API key.

**Content-Type**: `application/json`

**Request Body**

```json
{
  "skill": "ticket-classifier",
  "version": "2.1.0",
  "model": "claude-sonnet-4.6",
  "harness": "claude-code",
  "runner_version": "0.4.1",
  "benched_at": "2026-03-20T14:22:00Z",
  "trials": 3,
  "cases": [
    {
      "id": "tc-001",
      "with_skill": {
        "passed": true,
        "score": 1.0,
        "token_cost": 1180,
        "raw_output": "{ \"category\": \"billing\", \"priority\": \"P1\", ... }"
      },
      "without_skill": {
        "passed": false,
        "score": 0.3,
        "token_cost": 890,
        "raw_output": "I can help you with your billing issue..."
      }
    }
  ],
  "environment": {
    "os": "darwin",
    "arch": "arm64",
    "quill_version": "0.4.1"
  }
}
```

**Response**

```json
{
  "accepted": true,
  "eval_id": "ev_01HX7K2M3N4P5Q6R7S8T",
  "skill": "ticket-classifier",
  "version": "2.1.0",
  "model": "claude-sonnet-4.6",
  "delta": 38.2,
  "message": "Eval accepted. Results will be incorporated after validation (typically <5 minutes)."
}
```

---

### POST /v1/skills

Publish a new skill or new version to the registry. Authenticated.

**Authentication**: Required. Bearer token or API key.

**Content-Type**: `multipart/form-data`

**Form Fields**

| Field | Type | Required | Description |
|---|---|---|---|
| `manifest` | file | Yes | `quill.yaml` manifest file |
| `skill_file` | file | Yes | `SKILL.md` content file |
| `evals` | file | No | `evals.json` eval suite |
| `readme` | file | No | `README.md` documentation |

**Example Response**

```json
{
  "published": true,
  "name": "ticket-classifier",
  "version": "2.1.0",
  "url": "https://quill.dev/skills/ticket-classifier",
  "message": "Skill published. Smoke evals will run within 10 minutes."
}
```

---

### POST /v1/signals

Submit anonymous behavioral signals in batch. Signals are used for drift detection and anomaly detection. No authentication required — signals are anonymous by design.

See `spec/SIGNALS_SPEC.md` for the full signal format and privacy contract.

**Content-Type**: `application/json`

**Request Body**

```json
{
  "schema_version": "1.0",
  "batch_id": "b_7f3a9c",
  "submitted_at": "2026-03-20T15:00:00Z",
  "signals": [
    {
      "skill": "ticket-classifier",
      "version": "2.1.0",
      "model": "claude-sonnet-4.6",
      "harness": "claude-code",
      "outcome": "success",
      "trigger_matched": true,
      "token_cost": 1240,
      "session_hash": "sha256:a1b2c3d4...",
      "occurred_at": "2026-03-20T14:55:00Z"
    }
  ]
}
```

**Response**

```json
{
  "accepted": 1,
  "rejected": 0,
  "batch_id": "b_7f3a9c"
}
```

---

### GET /v1/bench/:skill

Returns full benchmark history for a skill across all models and versions. Useful for visualizing drift over time.

**Path Parameters**

| Parameter | Description |
|---|---|
| `skill` | Skill name |

**Query Parameters**

| Parameter | Type | Description |
|---|---|---|
| `model` | string | Filter to specific model |
| `version` | string | Filter to specific version |
| `since` | ISO8601 | Only return bench runs after this date |
| `limit` | int | Max records. Default 50. |

**Example Request**

```
GET /v1/bench/ticket-classifier?model=claude-sonnet-4.6&since=2026-01-01T00:00:00Z
```

**Example Response**

```json
{
  "skill": "ticket-classifier",
  "model": "claude-sonnet-4.6",
  "history": [
    {
      "version": "2.1.0",
      "benched_at": "2026-03-20T14:22:00Z",
      "delta": 38.2,
      "pass_rate_with": 0.94,
      "pass_rate_without": 0.56,
      "token_cost_avg": 1240
    },
    {
      "version": "2.1.0",
      "benched_at": "2026-03-01T09:00:00Z",
      "delta": 41.0,
      "pass_rate_with": 0.97,
      "pass_rate_without": 0.56,
      "token_cost_avg": 1220
    },
    {
      "version": "2.0.0",
      "benched_at": "2026-02-01T11:00:00Z",
      "delta": 35.5,
      "pass_rate_with": 0.91,
      "pass_rate_without": 0.56,
      "token_cost_avg": 1300
    }
  ],
  "drift_alerts": [
    {
      "detected_at": "2026-03-20T14:22:00Z",
      "type": "delta_decline",
      "magnitude": -2.8,
      "description": "Delta declined 2.8pp since last bench. Within normal variance."
    }
  ]
}
```

---

### GET /v1/compare

Head-to-head comparison of two skills for a given model. Returns side-by-side performance metrics.

**Query Parameters**

| Parameter | Type | Required | Description |
|---|---|---|---|
| `a` | string | Yes | First skill name (optionally with @version) |
| `b` | string | Yes | Second skill name (optionally with @version) |
| `model` | string | No | Model slug. Defaults to aggregate. |

**Example Request**

```
GET /v1/compare?a=ticket-classifier&b=support-triage&model=claude-sonnet-4.6
```

**Example Response**

```json
{
  "model": "claude-sonnet-4.6",
  "a": {
    "name": "ticket-classifier",
    "version": "2.1.0",
    "delta": 38.2,
    "pass_rate_with": 0.94,
    "pass_rate_without": 0.56,
    "token_cost_avg": 1240,
    "last_benched_at": "2026-03-20T14:22:00Z"
  },
  "b": {
    "name": "support-triage",
    "version": "1.4.2",
    "delta": 22.7,
    "pass_rate_with": 0.79,
    "pass_rate_without": 0.56,
    "token_cost_avg": 980,
    "last_benched_at": "2026-03-18T09:10:00Z"
  },
  "verdict": {
    "winner": "ticket-classifier",
    "delta_advantage": 15.5,
    "cost_advantage": "support-triage",
    "cost_difference_tokens": 260,
    "recommendation": "ticket-classifier has significantly higher delta (+15.5pp) at a cost of 260 more tokens per session. Prefer ticket-classifier unless token budget is constrained."
  },
  "overlap": {
    "composable": false,
    "reason": "Both skills handle ticket classification. Using both simultaneously is redundant."
  }
}
```

---

### GET /v1/worth-keeping/:skill

Evaluates whether a currently installed skill is still earning its place. Takes into account drift, model updates, and token cost vs. delta tradeoff.

**Path Parameters**

| Parameter | Description |
|---|---|
| `skill` | Skill name |

**Query Parameters**

| Parameter | Type | Description |
|---|---|---|
| `model` | string | Model in use (required for accurate assessment) |
| `version` | string | Installed version. Defaults to latest. |

**Example Request**

```
GET /v1/worth-keeping/ticket-classifier?model=claude-sonnet-4.6&version=2.1.0
```

**Example Response**

```json
{
  "skill": "ticket-classifier",
  "version": "2.1.0",
  "model": "claude-sonnet-4.6",
  "verdict": "keep",
  "confidence": 0.92,
  "reasoning": "Delta is 38.2pp — well above the 10pp threshold. No significant drift detected in past 30 days. Token cost is moderate at 1,240/session. No better alternative found in registry.",
  "alternatives": [],
  "drift_status": "stable",
  "last_assessed_at": "2026-03-20T14:22:00Z",
  "next_check_due": "2026-04-03T14:22:00Z"
}
```

**Verdict values**: `keep`, `review`, `retire`, `upgrade`

---

### GET /v1/upgrade/:skill

Returns upgrade recommendation for a skill, including whether a newer version has meaningfully better performance.

**Path Parameters**

| Parameter | Description |
|---|---|
| `skill` | Skill name |

**Query Parameters**

| Parameter | Type | Description |
|---|---|---|
| `from_version` | string | Currently installed version |
| `model` | string | Model in use |

**Example Request**

```
GET /v1/upgrade/ticket-classifier?from_version=2.0.0&model=claude-sonnet-4.6
```

**Example Response**

```json
{
  "skill": "ticket-classifier",
  "current_version": "2.0.0",
  "latest_version": "2.1.0",
  "model": "claude-sonnet-4.6",
  "upgrade_recommended": true,
  "current_delta": 35.5,
  "latest_delta": 38.2,
  "delta_improvement": 2.7,
  "breaking_changes": false,
  "changelog_summary": "v2.1.0: Added P0 priority tier, improved multi-category handling, reduced token usage by ~60 tokens.",
  "risk": "low"
}
```

---

## API Versioning

The API is versioned via the URL path (`/v1/`). Breaking changes increment the major version. The current version is `v1`.

Backwards-compatible additions (new optional fields, new endpoints) are made without a version bump. The `X-API-Version` response header reflects the exact deployed version.

```
X-API-Version: 1.4.2
```
