# hanko

Validate Claude Code plugin manifests before you submit them.

[![CI](https://github.com/RoninForge/hanko/actions/workflows/ci.yml/badge.svg)](https://github.com/RoninForge/hanko/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

hanko is a single-binary Go CLI that reads your `.claude-plugin/plugin.json` and `.claude-plugin/marketplace.json`, validates them against the schema extracted from Anthropic's plugin docs, and catches the common footguns that make marketplace submission fail: reserved marketplace names, duplicate hooks declarations, `agents` set to a bare directory string, path traversal, and more.

The name: **判子 (hanko)** is the personal seal stamped on official Japanese documents, the traditional act of approving a submission. Run `hanko check` before you file yours.

## Install

```sh
curl -fsSL https://roninforge.org/hanko/install.sh | sh
```

Or grab a binary from the [latest release](https://github.com/RoninForge/hanko/releases/latest). Prefer Go install:

```sh
go install github.com/RoninForge/hanko/cmd/hanko@latest
```

## Quickstart

```sh
# From inside a plugin repo:
hanko check

# Before submitting to a specific marketplace:
hanko submit-check --marketplace anthropic

# Print the embedded JSON Schema for editor integration:
hanko schema print --file plugin > .claude-plugin/plugin.schema.json

# Use in CI (JSON output):
hanko check --json .
```

## What it catches

- **Schema errors** the official validator reports opaquely (missing `name`, wrong types, non-kebab-case, invalid hook event names).
- **Reserved marketplace names** (eight exact names plus an impersonation-pattern regex).
- **Duplicate hooks declaration** where a plugin points `hooks` at `./hooks/hooks.json` even though Claude Code v2.1+ auto-loads that path by convention.
- **`agents` as a bare directory string**, which the validator rejects despite the docs example showing it works.
- **Path traversal** (`..` or absolute) in any component path.
- **Missing `author`** for plugins that plan to ship via the Anthropic marketplace (Claude Desktop's `listAvailablePlugins` validator refuses to load the whole catalog when even one plugin lacks one).
- **Missing `version`**, which is technically optional but breaks Claude Code's install cache.

See [docs/rules.md](docs/rules.md) for the full list with fix suggestions.

## Marketplaces

`hanko submit-check` layers rules on top of the base schema per marketplace:

| `--marketplace`        | Source                                                                                         | Extra rules                                              |
| ---------------------- | ---------------------------------------------------------------------------------------------- | -------------------------------------------------------- |
| `anthropic`            | [anthropics/claude-plugins-official](https://github.com/anthropics/claude-plugins-official)    | Strict reserved-name check, `author` strongly required   |
| `buildwithclaude`      | [davepoon/buildwithclaude](https://github.com/davepoon/buildwithclaude)                        | No delta from base rules (conventions pending)           |
| `cc-marketplace`       | [ananddtyagi/cc-marketplace](https://github.com/ananddtyagi/cc-marketplace)                    | `name`, `version`, `description` all required            |
| `claudemarketplaces`   | [mertbuilds/claudemarketplaces.com](https://github.com/mertbuilds/claudemarketplaces.com)      | No delta from base rules (auto-discovery)                |

## GitHub Action

```yaml
- uses: RoninForge/hanko@v0
  with:
    path: .
    marketplace: anthropic
```

Appends a findings table to the workflow run summary (the "Summary" tab of the PR's checks view) and exits non-zero on any error-severity finding. Set `fail-on-warnings: true` to also fail on warnings.

## How it works

1. `//go:embed` includes the JSON Schema for `plugin.json` and `marketplace.json`.
2. [`santhosh-tekuri/jsonschema/v6`](https://github.com/santhosh-tekuri/jsonschema) validates the file against the schema.
3. Go-coded rules apply the checks that do not round-trip cleanly through JSON Schema (reserved names, duplicate hooks, cross-references).
4. Errors are formatted with schema path, offending value, a one-line fix suggestion, and a doc link.

Zero network calls. Binary is roughly 6 MB. MIT licensed.

## Development

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Security

See [SECURITY.md](SECURITY.md).
