# Contributing to hanko

Thanks for considering a contribution. hanko has a sharp scope: **validate Claude Code plugin manifests (`plugin.json`, `marketplace.json`) before submission**. Contributions that stay inside that scope are welcome. Contributions that expand the scope (general CLAUDE.md linting, hook runtime testing, skill quality scoring) should be discussed in an issue first. See the sibling project [shodo](https://github.com/RoninForge/shodo) for AGENTS.md/CLAUDE.md scoring.

## Ground rules

- **Honest positioning.** No marketing-speak in code, comments, or docs. No emoji in copy or UI.
- **Stay boring.** Prefer stdlib over dependencies. Prefer plain data structures over abstractions.
- **No network calls.** hanko must remain a fully-offline validator. The only bytes that leave the machine are the ones you ship in a git commit.

## Development setup

Requires Go 1.25 or later.

```sh
git clone https://github.com/RoninForge/hanko.git
cd hanko
make build     # compile ./bin/hanko
make test      # go test ./... with race detector
make lint      # golangci-lint run
make fmt       # gofmt
```

## Running tests

```sh
go test ./...                       # everything
go test ./internal/validator        # single package
go test -race -cover ./...          # what CI runs
```

Every new rule must ship with:

1. A valid fixture under `testdata/valid/` that exercises it correctly.
2. An invalid fixture under `testdata/invalid/` that triggers the rule.
3. A table entry in the relevant `_test.go` asserting the expected error code.

## Commit style

Conventional Commits, low ceremony:

```
feat(rules): reject duplicate hooks declarations
fix(schema): allow object-form repository per docs
docs: link v0.1 marketplace research notes
```

## Pull requests

1. Open an issue first for anything larger than a typo.
2. Write tests.
3. Run `make check` locally and make sure CI is green.
4. Describe the behavior change in the PR body.

## Code layout

```
cmd/hanko/         thin main() entrypoint
internal/cli/      cobra command tree
internal/version/  build-time version metadata
internal/schema/   embedded JSON schemas (//go:embed)
internal/rules/    Go-coded rules beyond JSON Schema (reserved names, duplicate hooks, etc.)
internal/validator/ schema + rules orchestrator
internal/report/   error formatting (pretty / json)
action.yml         composite GitHub Action wrapper (at repo root for Marketplace)
scripts/           install script and helpers
testdata/          valid and invalid fixtures driving tests
```

## Reporting security issues

See [SECURITY.md](SECURITY.md). Do not file public issues for security bugs.
