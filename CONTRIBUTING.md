# Contributing to Quill

Quill is open source under the MIT license. Contributions are welcome.

## What to work on

Check the [issues](https://github.com/afscook129/quill/issues) for
open tasks. If you want to work on something not listed, open an issue
first to discuss.

**Good first contributions:**
- Add a new deterministic grading assertion (see `internal/eval/grader.go`)
- Add eval cases to the example skill (`examples/ticket-classifier/evals/evals.json`)
- Add harness detection for a new editor (`internal/cli/init.go`)
- Improve error messages to be more actionable
- Write tests for untested packages

## Development

```bash
# Build
make build

# Run tests
make test

# Run linter
make lint

# Run with coverage
make coverage
```

## Code style

- Standard Go conventions (`gofmt`, `go vet`)
- Error handling: wrap with context (`fmt.Errorf("doing X: %w", err)`)
- No global state — pass dependencies explicitly
- Tests in `_test.go` files alongside implementation

## Adding a grading assertion

Deterministic grading assertions live in `internal/eval/grader.go`.
To add a new one:

1. Add a case to the `switch` in `GradeDeterministic()`
2. Add tests in `internal/eval/grader_test.go`
3. Document the assertion name in `spec/EVAL_FORMAT.md`

## Adding a provider

Provider API clients live in `internal/provider/provider.go`.
To add support for a new model provider:

1. Implement the `Provider` interface (`Call` method)
2. Add model name detection in `ForModel()`
3. Read the API key from an environment variable
4. Add retry logic using `doWithRetry()`

## Adding harness detection

Harness detection lives in `internal/cli/init.go` in `detectHarnesses()`.
To add a new editor/harness:

1. Add detection logic (check for config directory or files)
2. Add config generation in `writeHarnessConfig()`
3. Test with a project that uses the harness

## Commits

- One logical change per commit
- Descriptive commit messages that explain why, not just what
- All tests must pass before committing

## Architecture

```
cmd/quill/main.go       → thin entry point
internal/cli/            → Cobra command definitions
internal/bench/          → benchmark runner (the core)
internal/eval/           → eval parsing + grading
internal/provider/       → LLM API clients (Anthropic, OpenAI)
internal/mcp/            → MCP server (agent interface)
internal/lock/           → quill.lock read/write
internal/manifest/       → quill.manifest.yaml read/write
internal/config/         → ~/.quill/config.yaml
internal/discovery/      → SKILL.md scanner
internal/tui/            → terminal styling
internal/registry/       → registry HTTP client (mock for now)
```

The bench runner (`internal/bench/runner.go`) is the heart of the project.
The MCP server (`internal/mcp/`) is the primary user-facing interface.
Everything else supports these two.
