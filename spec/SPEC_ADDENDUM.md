# Quill Spec Addendum v1.0
## Everything the builder needs that SPEC.md v5.1 doesn't cover
## Working name: "Quill" — will be renamed before public launch

---

## A. Competitive Landscape — Why This Exists Now

### A.1 The skill registry explosion

As of March 2026, the AI agent skills ecosystem crossed 350,000+ packages in roughly two months — a milestone npm took a decade to reach.

- **SkillsMP**: 200K+ skills. Largest catalog. No behavioral signal beyond metadata.
- **Skills.sh** (Vercel): 83K skills, 8M+ installs, tracks install counts across 18 agents.
- **ClawHub** (OpenClaw): 13K+ skills, vector search, semver, CLI, lock files.

None of these registries carry behavioral delta data. They rank by stars, downloads, or keyword match. Nobody can tell you: "this skill adds +38pp on Sonnet 4.6, drifted -3pp after the March update, and costs 1,240 tokens per session."

### A.2 SkillsBench validates the thesis

SkillsBench (benchflow-ai, February 2026) is an academic benchmark: 84 tasks, 7 models, 3 harnesses, 7,308 trajectories. Key findings:

- Curated skills raise average pass rate by **+16.2 percentage points**
- Effects vary wildly by domain (+4.5pp for software engineering, +51.9pp for healthcare)
- 16 of 84 tasks showed **negative** deltas — some skills make things worse
- Self-generated skills provide **no benefit on average**

SkillsBench is static — run once, paper published. Quill is the **continuous production infrastructure** that makes this measurement ongoing, personal, and actionable.

**Positioning**: "SkillsBench proved skills matter. Quill makes that measurement continuous."

### A.3 What Quill does that nobody else does

| Capability | SkillsMP | Skills.sh | ClawHub | SkillsBench | **Quill** |
|---|---|---|---|---|---|
| Skill catalog | 200K+ | 83K | 13K | 84 tasks | Indexes all |
| Install counts | — | Yes | Yes | — | Yes |
| Behavioral delta per model | — | — | — | Static dataset | **Continuous** |
| Drift detection | — | — | — | — | **Yes** |
| "Still earning its place?" | — | — | — | — | **Yes** |
| Cross-registry search | — | — | — | — | **Yes** |
| Lock file with outcomes | — | — | `.clawhub/lock.json` | — | **Yes** |
| CI gate on regression | — | — | — | — | **Yes** |
| Open delta dataset (ODbL) | — | — | — | Open (paper) | **Yes** |

---

## B. Architecture — Resolved

### B.1 System overview

```
                    USER INTERFACES
  MCP Server  |  CLI (Go)  |  quill-assistant  |  Web UI (Ph5)
  (local or   |            |  (SKILL.md)       |  (Vercel)
   Fly.io)    |            |                   |

       PUBLIC REPO (MIT)
   github.com/afscook129/quill
                  | HTTP/JSON API
   ===============|========================================
                  | PRIVATE REPO (proprietary)
                  | github.com/afscook129/quill-cloud
         Registry API          Web UI
         (Convex cloud)        (Vercel, Phase 5)
         - Skills DB
         - Eval results
         - Bench history
         - Signals
         - Vector index
         - File storage
         - Cron jobs
                  |
         Embedding API
         (OpenAI)
         text-embed-3-small
```

### B.2 Stack decisions — all resolved

| Component | Technology | Rationale |
|---|---|---|
| **CLI + MCP server** | Go + Cobra + Bubbletea/Lipgloss | Single binary, no runtime deps. As spec'd. |
| **Registry backend** | Convex (TypeScript) | DB + vector search + file storage + cron + real-time. Generous free tier. Self-hostable. |
| **Vector search** | Convex built-in vector indexes | No separate vector DB. Embeddings stored alongside relational data. |
| **Embedding model** | OpenAI `text-embedding-3-small` | $0.02/million tokens. 500 skills = $0.002. At 50K skills = $0.20. Negligible cost. |
| **Embedding generation** | At publish time in a Convex action | Registry generates embedding from `behavioral_description` field, stores vector in same DB. |
| **Object storage** | Convex file storage | Manifests, eval suites, bench artifacts. |
| **Hosted MCP server** | Go binary on **Fly.io** | Container deployment, scales to zero. SSE transport. `mcp.quill.dev`. |
| **Web UI** (Phase 5) | Next.js on **Vercel** | `quill.dev`. Free tier. |
| **Auth** (see section C) | GitHub OAuth via Convex auth | Phase 1-4. WorkOS added Phase 6 for enterprise SSO. |
| **Go binary distribution** | GitHub Releases + GoReleaser | Phase 2. Brew tap at `afscook129/homebrew-quill`. |

### B.3 Why Convex, not Postgres + pgvector

- **Single service**: DB, vector search, file storage, cron, real-time subscriptions — all in one. No Postgres cluster + S3 + separate cron service to operate.
- **Free tier**: Generous for early stage. No hosting cost until real scale.
- **Self-hostable**: Open-source backend for orgs that want private registries.
- **TypeScript functions**: Registry logic lives in Convex functions. The Go CLI is a client that calls the Convex API.
- **If Convex doesn't scale**: The API contract is HTTP JSON. The Go CLI doesn't know or care what's behind the API. Registry backend can be swapped later without touching the CLI.

### B.4 Two-repo structure

The public and private code live in separate repositories. The public CLI is a client that calls the registry API over HTTP. It imports nothing from the private repo. The API spec (documented openly) is the boundary.

#### Public repo — `github.com/afscook129/quill` (MIT)

This is what gets stars, adoption, and community contributions.

#### Private repo — `github.com/afscook129/quill-cloud` (proprietary)

This is the business. The registry, the web UI, the infrastructure, the seed run tooling.

### B.5 Hosting map

| Service | Host | When needed | Cost at launch |
|---|---|---|---|
| Registry backend (DB, vector, crons) | **Convex cloud** | Phase 1 | Free tier |
| Hosted MCP server (`mcp.quill.dev`) | **Fly.io** | Phase 1 | ~$0-5/mo (scales to zero) |
| Web UI (`quill.dev`) | **Vercel** | Phase 5 | Free tier |
| Docs site | **Vercel** | Phase 2 | Free tier |
| Go binary distribution | **GitHub Releases** + GoReleaser | Phase 2 | Free |
| npm wrapper (`@quill-ai/cli`) | **npm registry** | Phase 2 | Free |
| Brew tap | **`afscook129/homebrew-quill`** on GitHub | Phase 2 | Free |

### B.6 Open source vs. commercial line

| Component | Repo | License | Rationale |
|---|---|---|---|
| Go CLI + MCP server | public | MIT | Maximum adoption. |
| Lock file / manifest / eval formats | public | MIT | Standards should be open. |
| quill-assistant SKILL.md | public | MIT | Distribution = adoption. |
| Registry API spec | public | MIT | Anyone can build a compatible registry. |
| Convex registry backend | private | Proprietary | The implementation is the business. |
| Web UI | private | Proprietary | Part of the product. |
| Seed run tooling | private | Proprietary | Operational tooling. |
| Delta dataset | public | ODbL | Open data, share-alike. The moat is the data. |

---

## C. Auth Model — Resolved

### C.1 Phase 1-4: GitHub OAuth only

- **Reads**: Fully public. No auth required for search, explain, compare, or reading eval data.
- **Writes**: GitHub OAuth required for publishing skills, submitting eval results, submitting signals.
- **API keys**: Generated after GitHub OAuth login. Used for CI/CD (`QUILL_PUBLISH_TOKEN`).
- **Rate limiting**: 100 req/min unauthenticated, 1000 req/min authenticated.
- **Implementation**: Convex has built-in GitHub OAuth support. One config.

### C.2 Phase 5: Add email/password

For non-technical personas accessing quill.dev who may not have GitHub accounts.

### C.3 Phase 6: Add WorkOS

Enterprise SSO (SAML/OIDC) via WorkOS for paid org tier.

---

## D. Eval Format

See `spec/EVAL_FORMAT.md` for the full evals.json schema, grading methods, and examples.

---

## E. Bench Runner — Execution Model

### E.1 What "with-skill vs without-skill" means mechanically

For each eval case, the bench runner makes **two API calls**:

**Call 1 — WITH skill**: System prompt includes SKILL.md content, user prompt is the eval case input.
**Call 2 — WITHOUT skill**: System prompt is empty/minimal, user prompt is the same eval case input.

Both outputs are graded against the same rubric. The delta = pass rate difference.

### E.2 Token math

Per eval case (3 trials default):
- 2 conditions x 3 trials = 6 API calls
- Per smoke eval (8 cases): ~57,600 tokens
- **Total per smoke eval: ~$0.48**
- **Total per full bench (20 cases): ~$1.20**

### E.3 Provider API calls

The bench runner calls provider APIs directly via Go HTTP clients:
- Anthropic: `POST https://api.anthropic.com/v1/messages`
- OpenAI: `POST https://api.openai.com/v1/chat/completions`
- Google: `POST https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent`

---

## F. Seed Run — Mechanics

500 skills curated from ClawHub, Skills.sh, SkillsMP, and Microsoft repos. Five models. ~75,000 API calls. Estimated cost: ~$150-300. Published as open data (ODbL).

---

## G. `quill add` — Full Install Flow

1. **Resolve**: Query registry for skill metadata
2. **Download**: Fetch SKILL.md + supporting files from source
3. **Place**: Copy into harness skills directory
4. **Validate**: Run smoke evals (8 cases)
5. **Lock**: Write entry to quill.lock
6. **Watch**: Schedule 7-day follow-up bench check

---

## H. Phase 1 Scope Lockdown

### H.1 Phase 1 deliverables

| # | Deliverable | Done when |
|---|---|---|
| 1 | Convex registry schema + seed data loader | Schema deployed, 500 skills loaded |
| 2 | Registry API: `GET /v1/search` | Semantic search returns ranked results |
| 3 | Registry API: `GET /v1/skills/:name` | Returns skill metadata + bench history |
| 4 | Registry API: `POST /v1/evals` | Accepts eval submissions (authenticated) |
| 5 | Registry API: `GET /v1/evals/:skill/:model` | Returns cached eval results |
| 6 | MCP server: all 8 tools | search, add, status, bench, fix, explain, compare, retire |
| 7 | `quill init` command | Auto-detects harnesses, adds MCP config |
| 8 | `quill mcp --serve` | Starts local MCP server (stdio transport) |
| 9 | Hosted MCP at `mcp.quill.dev/sse` | Remote MCP for cloud-native platforms |
| 10 | GitHub OAuth via Convex | Login for publishing + API key generation |
| 11 | Seed run data loaded into registry | 500 skills x 5 models |

### H.2 Deferred to later phases

Phase 2: Full CLI commands, TUI, quill-assistant, brew tap, npm distribution.
Phase 3: Hook layer, signals, drift detection, GitHub Actions.
Phase 4: Trigger accuracy, bench-history, bench-compare.
Phase 5: Web UI, team features.
Phase 6: Org scale, SSO, policy engine.

### H.3 Phase 1 acceptance tests

```
TEST: "Developer adds MCP server and searches for a skill"
  GIVEN: Quill MCP server is configured in Claude Code
  WHEN: User says "I need something that classifies support tickets"
  THEN: Agent calls quill_search, returns ranked results with delta scores

TEST: "Developer installs a skill via MCP"
  GIVEN: quill_search returned ticket-classifier@2.1.0
  WHEN: Agent calls quill_add
  THEN: SKILL.md placed in correct directory, smoke evals pass, quill.lock updated

TEST: "quill init configures everything"
  GIVEN: Claude Code is installed
  WHEN: User runs quill init
  THEN: MCP config created, manifest created, signal prompt shown, status runs

TEST: "Registry search returns ranked results"
  GIVEN: Registry seeded with 500 skills
  WHEN: GET /v1/search?q=classify+support+tickets&model=claude-sonnet-4.6
  THEN: Results ranked by scoring formula, each includes delta/pass_rate/token_cost

TEST: "Bench runner produces correct delta"
  GIVEN: A skill with known behavior and eval cases
  WHEN: quill_bench is called with 3 trials
  THEN: With/without pass rates computed, delta calculated, patterns identified
```

---

## I. Testing Strategy — TDD

Test pyramid: 50-100 unit tests, 15-25 integration tests, 3-5 E2E tests.

---

## K. Resolved Decisions Log

| Decision | Resolution | Date |
|---|---|---|
| Stack | Go CLI + Convex registry + OpenAI embeddings | Mar 24, 2026 |
| Repo structure | Two repos: public MIT + private proprietary | Mar 24, 2026 |
| Hosted MCP | Fly.io (Go binary, SSE transport) | Mar 24, 2026 |
| Web UI hosting | Vercel (Phase 5) | Mar 24, 2026 |
| Auth Phase 1-4 | GitHub OAuth via Convex | Mar 24, 2026 |
| Auth Phase 6 | WorkOS for enterprise SSO | Mar 24, 2026 |
| Seed run cost | ~$150-300, approved | Mar 24, 2026 |

---

## M. Claude Code Build Order

**Build sequence for Phase 1**:

```
Step 1: Public repo scaffolding
  → Go module, Cobra CLI, Makefile, CI workflow

Step 2: Lock file + manifest packages
  → Parse, write, validate, diff

Step 3: Harness detection + quill init
  → Detect harnesses, generate MCP config

Step 4: Private repo scaffolding (quill-cloud)
  → Convex project, schema, seed data loader

Step 5: Registry API — search + read
  → GET /v1/search, GET /v1/skills/:name, GET /v1/evals/:skill/:model

Step 6: Registry API — write
  → POST /v1/evals, POST /v1/skills, GitHub OAuth

Step 7: MCP server implementation
  → All 8 tools, stdio transport

Step 8: Bench runner
  → With/without execution, LLM-judge grading, delta computation

Step 9: quill add flow
  → Resolve → download → place → validate → lock

Step 10: Hosted MCP on Fly.io
  → Dockerfile, fly.toml, SSE transport

Step 11: Seed run
  → Scrape 500 skills, batch bench, load into registry
```
