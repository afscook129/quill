# CLAUDE.md — Project Instructions for Quill

> Working name: "Quill" — will be renamed before public launch.
> This is the PUBLIC repo: `github.com/afscook129/quill`

## What this project is

Quill is the credibility layer for AI agent skills. It measures whether skills
actually help — with real numbers, per model, over time. See `spec/SPEC.md` and
`spec/SPEC_ADDENDUM.md` for the full product specification.

## Repository structure

```
quill/
├── CLAUDE.md                    ← you are here
├── cmd/quill/main.go            # CLI entry point
├── internal/                    # All Go packages
│   ├── tui/                     # Bubbletea TUI components
│   ├── discovery/               # Cross-registry semantic search
│   ├── bench/                   # Benchmark engine
│   ├── eval/                    # Eval runner + grading
│   ├── resolver/                # Dependency resolution
│   ├── hooks/                   # Claude Code hook handlers
│   ├── harness/                 # Harness detection + config gen
│   ├── lock/                    # Lock file read/write/validate
│   ├── manifest/                # Manifest parse/validate/infer
│   ├── mcp/                     # MCP server implementation
│   ├── registry/                # HTTP client for registry API
│   ├── signals/                 # Outcome signal collection
│   ├── security/                # Sigstore, permissions, SBOM
│   └── config/                  # ~/.quill/config.yaml
├── skills/quill-assistant/      # quill-assistant SKILL.md
├── spec/                        # All specification documents
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Tech stack

- **Language**: Go 1.22+
- **CLI framework**: Cobra
- **TUI**: Bubbletea + Lipgloss
- **MCP server**: Go stdlib HTTP + SSE (stdio for local, SSE for hosted)
- **Registry client**: Go stdlib `net/http` — calls Convex HTTP API
- **Testing**: Go standard `testing` package + `testify` for assertions
- **Build**: GoReleaser for cross-platform binaries

## Development approach: TDD

Every feature follows this sequence:
1. Write the acceptance test (see SPEC_ADDENDUM.md Section H.3)
2. Write unit tests for the component
3. Implement until tests pass
4. Integration test to verify wiring
5. Commit with descriptive message

## Coding conventions

### Go style
- Follow standard Go conventions (`gofmt`, `go vet`, `golint`)
- Package names: short, lowercase, single word when possible
- Error handling: always wrap with context (`fmt.Errorf("doing X: %w", err)`)
- No global state — pass dependencies explicitly
- Interfaces defined by the consumer, not the implementer

### File organization
- One primary type per file, named after the type
- Tests in `_test.go` files alongside implementation
- Internal packages under `internal/` — not importable by external code
- `cmd/quill/main.go` is thin — delegates to internal packages immediately

### Naming
- Commands: `quill <verb> [noun]` (e.g., `quill bench`, `quill add`, `quill status`)
- MCP tools: `quill_<verb>` (e.g., `quill_search`, `quill_bench`, `quill_status`)
- Config files: `quill.manifest.yaml`, `quill.lock`
- Internal packages: match the spec section names (discovery, bench, eval, etc.)

### Output
- TTY detected: full TUI rendering with Bubbletea
- Non-TTY (pipes, CI): plain text, no color, structured when `--json` flag used
- Errors: always include the fix, not just the problem
- Progress: communicate findings, not just "loading..."

## Registry API

This CLI is a **client** of the registry API. The registry runs separately
in the private `quill-cloud` repo on Convex. The API contract is documented
in `spec/REGISTRY_API.md`.

Base URL: `https://registry.quill.dev/v1/` (production)
Local dev: `http://localhost:3210/v1/` (Convex dev server)

The CLI should use an environment variable `QUILL_REGISTRY` to override the
base URL, defaulting to production.

## Build sequence (Phase 1)

Follow the build order in SPEC_ADDENDUM.md Section M. Each step has
acceptance criteria. Do not skip ahead.

```
Step 1:  Public repo scaffolding (Go module, Cobra, Makefile, CI)
Step 2:  Lock file + manifest packages
Step 3:  Harness detection + quill init
Step 7:  MCP server implementation (after registry API exists)
Step 8:  Bench runner
Step 9:  quill add flow
```

Steps 4-6 are in the private repo (Convex registry). Steps 7-9 depend on
the registry API being available. During development, use mock HTTP responses
to develop and test the CLI independently.

## Key spec documents

| Document | What it covers |
|---|---|
| `spec/SPEC.md` | Full product specification v5.1 |
| `spec/SPEC_ADDENDUM.md` | Architecture, eval format, Phase 1 scope, testing |
| `spec/REGISTRY_API.md` | Every registry endpoint |
| `spec/EVAL_FORMAT.md` | evals.json schema + grading methods |
| `spec/MANIFEST_SPEC.md` | quill.manifest.yaml full schema |
| `spec/LOCK_SPEC.md` | quill.lock full schema |
| `spec/SIGNALS_SPEC.md` | Outcome signal format + privacy |

## Testing

```bash
# Run all tests
make test

# Run tests for a specific package
go test ./internal/lock/...

# Run tests with coverage
make coverage

# Run linter
make lint
```

## Important constraints

1. **Single binary, no runtime dependencies.** The Go binary must not require
   Python, Node.js, or any other runtime. The `--grader-endpoint` flag is the
   escape hatch for teams that want Python graders.

2. **No lock-in.** Everything Quill produces is plain YAML or JSON. If someone
   stops using Quill, their skills, configs, and lock files still work.

3. **Privacy by default.** Outcome signals are opt-in. No content is ever
   collected. See `spec/SIGNALS_SPEC.md` for what is and isn't collected.

4. **The user problem is the framing.** Never use internal metric language
   in user-facing output. "Still earning its place?" not "retirement threshold."
   "The base model now handles this well" not "delta below 8pp for 2 versions."
