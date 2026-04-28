# AGENTS.md - Hanko

> Guidance for AI coding agents working on this repository. Human contributors should read CONTRIBUTING.md first; this file exists to give agents the same context without spelunking.

## What this is

Hanko (判子) is a **validator for Claude Code plugin manifests**. It reads `.claude-plugin/plugin.json` and `.claude-plugin/marketplace.json`, runs them through an embedded JSON Schema (the upstream `hesreallyhim/claude-code-json-schema` is vendored and credited), and layers Go-coded rules for the cases the schema cannot express.

Ships as a Go CLI and a GitHub Action. MIT licensed. Part of the [RoninForge](https://roninforge.org) toolkit.

## Trust pledge (load-bearing - never break this)

1. **Zero network calls in the binary.** The validator runs fully offline. The install script is the only time bytes cross the network, and that is GitHub Releases. If a code change introduces an HTTP client, package fetcher, or schema-update-on-startup, it violates the pledge.
2. **Read-only.** Hanko never modifies the user's files. It prints a report. If you add a `--fix` mode, gate it behind explicit opt-in and write a test that proves the no-flag path is read-only.

## What Hanko catches

The rules layer is the value. Catalog (keep this in sync with `internal/rules/`):

- All 8 reserved marketplace names plus the impersonation pattern.
- Duplicate `hooks` and `hooks.json` declarations (v2.1+ footgun).
- The `agents` field set to a bare directory string instead of an array of agent objects.
- Path traversal in component paths.
- Per-marketplace stricter rules for `anthropic`, `buildwithclaude`, `cc-marketplace`.

When a marketplace's rules change, **add a test fixture under `testdata/<marketplace>/` first**, then make it pass. Catalog-of-record drift is the failure mode.

## Layout

    cmd/hanko/             Cobra entrypoint
    internal/validator/    Schema + rule pipeline
    internal/rules/        Go-coded rules per marketplace
    internal/schema/       Embedded JSON Schema (hesreallyhim upstream)
    internal/report/       Pretty-printer + JSON output
    action.yml             GitHub Action manifest
    testdata/              Plugin/marketplace fixtures (good and bad)

## Build, test, lint

    make check    # fmt + vet + lint + test (run before any commit)
    make build    # → ./bin/hanko
    make test     # race detector + coverage
    make snapshot # local goreleaser dry-run, no publish

CI runs `make check` on every push/PR. Don't merge red.

## Style

- Go 1.22+. `errors.Is` / `errors.As`, not string matching.
- No emoji in code, comments, or output. Dev-tool audience finds them amateurish.
- No em dashes in user-facing strings.
- Errors include schema path, fix suggestion, and docs link. If you add a rule, add the suggestion + link.

## Sibling tool

[Tsuba](https://github.com/RoninForge/tsuba) scaffolds plugin directories that pass Hanko on the first run. `tsuba validate` shells out to Hanko via `os/exec`, so:

- Don't break the JSON output schema (`hanko validate --json`) without a coordinated change in Tsuba.
- Don't change exit codes silently. Exit 0 = pass, 1 = validation failure, 2 = usage error. Tsuba relies on this.

## Releasing

    git tag v0.X.Y
    git push origin v0.X.Y

goreleaser builds binaries, publishes to GitHub Releases, and bumps the Homebrew tap (`roninforge/homebrew-tap`). The GitHub Action listing on the Actions Marketplace is tied to the same tag, so don't tag anything red.

## What you should NOT do

- Do not introduce a network call in the validator. Schema updates ship by recompiling and releasing, not by fetching at runtime.
- Do not modify the user's files in any code path other than an explicit, opt-in `--fix` flag (which does not exist yet).
- Do not break the JSON output schema. Tsuba parses it.
- Do not ship a release without test fixtures for every new rule.

## More context

- Site: https://roninforge.org/hanko
- Markdown digest for AI fetchers: https://roninforge.org/hanko.md
- Sibling tools (same RoninForge toolkit): [BudgetClaw](https://github.com/RoninForge/budgetclaw), [Tsuba](https://github.com/RoninForge/tsuba)
