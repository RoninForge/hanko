# Changelog

All notable changes to hanko are documented here. Format based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), versions follow [SemVer](https://semver.org/).

## [Unreleased]

## [0.2.1] - 2026-04-27

### Changed

- `action.yml` `name:` field renamed from `hanko` to `Hanko Plugin Validator` so the Action passes GitHub Marketplace's global-uniqueness check (`hanko` collides with existing GitHub user / org names). Repo name, `uses: RoninForge/hanko@v0` consumption path, branding, and binary name are unchanged.

## [0.2.0] - 2026-04-27

### Changed

- **BREAKING:** moved `action.yml` from `action/action.yml` to the repository root. The composite Action now lives at the path GitHub's Actions Marketplace expects, which lets hanko be discovered, ranked, and installed from the Marketplace UI rather than only by direct repo reference. Update consuming workflows from `uses: RoninForge/hanko/action@v0` to `uses: RoninForge/hanko@v0`. The `v0` floating tag has been re-pointed at v0.2.0; pinned references to `v0.1.x/action` continue to resolve from the prior release.

## [0.1.2] - 2026-04-19

### Fixed

- `action/action.yml`: a misconfigured action (wrong `path:` input, directory without a `.claude-plugin/`) used to crash the step with a Python `JSONDecodeError` traceback because the summary script tried to parse an empty JSON file. Added an explicit guard that bails early and propagates hanko's own exit code so users see hanko's friendly error instead of a Python stack trace.
- `internal/validator`: the "schema validator returned an unexpected error" branch now emits a distinct `HANKO-INTERNAL` rule ID instead of overloading `HANKO000` ("invalid JSON"). Keeps CI pins on `HANKO000` from silently swallowing internal library bugs.
- `scripts/install.sh`: replaced a Unicode arrow `→` with `->` in the status line, matching the project's plain-ASCII UI rule applied to pretty output in v0.1.1.

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

[Unreleased]: https://github.com/RoninForge/hanko/compare/v0.1.2...HEAD
[0.1.2]: https://github.com/RoninForge/hanko/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/RoninForge/hanko/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/RoninForge/hanko/releases/tag/v0.1.0
