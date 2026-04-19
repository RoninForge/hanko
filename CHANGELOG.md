# Changelog

All notable changes to hanko are documented here. Format based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), versions follow [SemVer](https://semver.org/).

## [Unreleased]

## [0.1.1] - 2026-04-19

### Fixed

- `scripts/install.sh`: checksum verification failed on every install because `grep -F "  $ARCHIVE"` also matched the archive's `.sbom.json` sibling line (substring match), returning two hashes. Replaced with `awk '$2 == f'` for exact-field comparison plus a guard that errors out loudly if the checksum file ever produces multiple matches.
- Pretty-output clean line now shows `ok` instead of `✓`, matching the project's "no emoji in UI" rule and the non-color output branch.
- `SECURITY.md` no longer references a non-existent `--fix` flag.
- Windows release archives now ship as `.zip` (goreleaser `format_overrides`) so double-click extract works on stock Windows.
- Removed unused `id-token: write` permission from the Release workflow (we do not sign with OIDC).

### Changed

- CI release trigger now only fires on exact semver tags (`v[0-9]+.[0-9]+.[0-9]+*`). Moving floating aliases like `v0` with `git tag -f` no longer re-triggers goreleaser against an already-published release.

## [0.1.0] - 2026-04-19

### Added

- Initial release: validate `.claude-plugin/plugin.json` and `.claude-plugin/marketplace.json` against an embedded JSON Schema derived from Anthropic's published plugin docs.
- `hanko check` validates a single manifest with pretty errors, schema path, and docs links.
- `hanko submit-check --marketplace <name>` layers marketplace-specific rules (anthropic, buildwithclaude, cc-marketplace, claudemarketplaces).
- `hanko schema print` emits the embedded JSON Schema for editor integration.
- Reserved-name detection for the 8 marketplace names blocked by Anthropic plus an impersonation-pattern regex.
- Duplicate-hooks declaration detection (common v2.1+ footgun).
- `agents` as bare directory rejection with a fix suggestion.
- `--json` output for CI use.
- Composite GitHub Action wrapper that appends findings to the workflow run summary.
