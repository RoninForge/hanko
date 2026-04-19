# Hanko rule catalog

Every finding hanko emits carries a stable rule identifier so CI scripts can pin on specific rules, ignore others, or report on them over time. This document is the authoritative list.

Rule identifiers are **stable across releases**. If a rule is removed or renamed, the old ID stays reserved so a pinned CI script never silently matches a different check.

Severity classes:

- **error** - blocks submission; CLI exits non-zero.
- **warning** - advisory; CLI exits zero unless `--fail-on-warnings` is passed to the GitHub Action wrapper.

Marketplace overlays can promote a base rule to a strict variant (`-strict` suffix) that runs at error severity. When that happens the base rule does NOT also fire, so a single root cause never produces duplicate findings.

---

## Parser

### HANKO000 - invalid JSON

**Severity:** error
**Condition:** The manifest file is not parseable as JSON.
**Fix:** Run the file through `jq .` or your editor's JSON linter to find the syntax error.

### HANKO-SCHEMA - JSON Schema validation failure

**Severity:** error
**Condition:** The decoded JSON violates the embedded hanko schema for the requested kind (plugin or marketplace). One finding per leaf violation.
**Source of truth:** `internal/schema/plugin.schema.json` and `internal/schema/marketplace.schema.json`, derived from [Anthropic's plugin docs](https://code.claude.com/docs/en/plugins-reference) and [plugin marketplace docs](https://code.claude.com/docs/en/plugin-marketplaces). See `docs/research/phase-1-schema.md` for the provenance of each field.

Common instances:

- missing required `name` field
- `name` not kebab-case (`^[a-z0-9]+(-[a-z0-9]+)*$`)
- `version` not semver (`MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]`)
- path field that does not start with `./`
- path field containing `..`
- `source` object with wrong type constant
- unknown top-level field (schema is `additionalProperties: false`)

---

## Plugin-manifest rules

### HANKO001 - duplicate hooks declaration

**Severity:** error
**Condition:** `hooks` is set to `./hooks/hooks.json` (or `hooks/hooks.json`), either as a string or as a single-element array. Claude Code v2.1+ auto-loads that path by convention, so the explicit declaration causes a double-load error at install time.
**Fix:** Remove the `hooks` field entirely and let Claude Code auto-discover the default path. If you have additional hooks files, point `hooks` at those paths only.
**Evidence:** `affaan-m/everything-claude-code/.claude-plugin/PLUGIN_SCHEMA_NOTES.md`, confirmed in that repo's commit history showing repeated fix/revert cycles.

### HANKO002 - agents as a bare directory string

**Severity:** error
**Condition:** `agents` is a string that does not end in `.md` (e.g. `"./agents"` or `"./agents/"`). The docs example shows a directory string as legal, but the official validator rejects it in practice.
**Fix:** Use an array of explicit file paths: `"agents": ["./agents/reviewer.md", "./agents/planner.md"]`.
**Evidence:** [anthropics/claude-code#44777](https://github.com/anthropics/claude-code/issues/44777).

### HANKO003 - author missing

**Severity:** warning (base) / error (`--marketplace=anthropic`)
**Condition:** `author` field absent. Claude Desktop's `listAvailablePlugins` validator refuses to load any marketplace that contains a plugin without an author, so one missing `author` breaks the entire catalog for every user.
**Fix:** `"author": { "name": "Your Name", "email": "you@example.com" }`.
**Evidence:** [anthropics/claude-code#33068](https://github.com/anthropics/claude-code/issues/33068).

**Strict variant:** `HANKO003-strict` fires at error severity under `--marketplace=anthropic`. When the strict variant fires, the base `HANKO003` is suppressed to avoid double-reporting.

### HANKO004 - version missing

**Severity:** warning (base) / error (`--marketplace=cc-marketplace`)
**Condition:** `version` field absent. The docs list it as optional, but Claude Code's install cache keys on this field; omitting it means users will not see updates you ship.
**Fix:** `"version": "0.1.0"` and bump on every release.

**Strict variant:** `HANKO004-strict` fires at error severity under `--marketplace=cc-marketplace`. The base `HANKO004` is suppressed when the strict variant fires.

### HANKO005-strict - description missing (cc-marketplace only)

**Severity:** error
**Condition:** `description` field absent. The cc-marketplace submission rules require it.
**Fix:** Add a one-sentence `description` field explaining what the plugin does.
**Evidence:** [ananddtyagi/cc-marketplace/PLUGIN_SCHEMA.md](https://github.com/ananddtyagi/cc-marketplace/blob/main/PLUGIN_SCHEMA.md).

---

## Marketplace-manifest rules

### HANKO101 - reserved marketplace name

**Severity:** error
**Condition:** Top-level `name` matches one of the eight marketplace names Anthropic reserves for their own first-party listings:

- `claude-code-marketplace`
- `claude-code-plugins`
- `claude-plugins-official`
- `anthropic-marketplace`
- `anthropic-plugins`
- `agent-skills`
- `knowledge-work-plugins`
- `life-sciences`

**Fix:** Rename your marketplace. Common patterns are `<your-github-username>-plugins` or `<topic>-claude-plugins`.
**Evidence:** [plugin marketplace docs](https://code.claude.com/docs/en/plugin-marketplaces), cross-referenced with [anthropics/claude-code#14145](https://github.com/anthropics/claude-code/issues/14145) and [#18232](https://github.com/anthropics/claude-code/issues/18232).

### HANKO102 - impersonation pattern

**Severity:** warning (all marketplaces)
**Condition:** Top-level `name` matches the impersonation regex (captures combinations like `official-claude-plugins`, `anthropic-tools-v2`, `claude-marketplace-official`).
**Fix:** Consider renaming to avoid the ambiguity. Anthropic's own validator is stricter than ours and may reject the submission even if hanko only warns.
**Note:** Hanko surfaces this as a warning in every context. The upstream validator has been known to over-match ([#18232](https://github.com/anthropics/claude-code/issues/18232)), so we intentionally stop short of promoting it to an error even under `--marketplace=anthropic`.

### HANKO103 - duplicate plugin names in marketplace

**Severity:** error
**Condition:** Two or more entries in the `plugins` array share the same `name`.
**Fix:** Plugin names must be unique within a marketplace. Rename one of the entries.

---

## CI and pinning

Each rule ID is a string. The JSON output (`hanko check --json` or `hanko submit-check --json`) emits findings with their rule IDs, making it easy to:

- **Count errors by rule** - `jq '[.[] | .findings[] | select(.rule=="HANKO001")] | length'`.
- **Ignore a specific warning** - post-process the report and filter out that rule ID before failing.
- **Lock on a known-good set** - snapshot the report, diff on every run.

If a rule needs to be renamed or removed in a future release, the old ID will stay reserved in this document, so any CI that pins on it either keeps working or fails loudly.

---

## Severity promotion rule

A rule has at most one strict variant, named `<RULE>-strict`. When a marketplace overlay promotes the base rule to error severity:

- The base rule is **replaced** in the rule set, not layered alongside.
- Only the `-strict` variant appears in the output.
- Severity on the `-strict` variant is always error.

This keeps a single root cause from producing two findings (one warning, one error) in the same report.
