# Changelog

All notable changes to hanko are documented here. Format based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), versions follow [SemVer](https://semver.org/).

## [Unreleased]

### Added

- Initial release: validate `.claude-plugin/plugin.json` and `.claude-plugin/marketplace.json` against an embedded JSON Schema derived from Anthropic's published plugin docs.
- `hanko check` validates a single manifest with pretty errors, schema path, and docs links.
- `hanko submit-check --marketplace <name>` layers marketplace-specific rules (anthropic, buildwithclaude, cc-marketplace, claudemarketplaces).
- `hanko schema print` emits the embedded JSON Schema for editor integration.
- Reserved-name detection for the 8 marketplace names blocked by Anthropic plus an impersonation-pattern regex.
- Duplicate-hooks declaration detection (common v2.1+ footgun).
- `agents` as bare directory rejection with a fix suggestion.
- `--json` output for CI use.
- Composite GitHub Action wrapper that posts a PR check-run summary.
