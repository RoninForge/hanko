# Phase 1 Research: Claude Code Plugin/Marketplace Schema

**Research window:** 2026-04-19
**Target product:** Hanko (Go CLI for validating `.claude-plugin/plugin.json` and `.claude-plugin/marketplace.json` before submission to plugin marketplaces)
**Primary sources:** `code.claude.com/docs/en/plugins-reference`, `code.claude.com/docs/en/plugin-marketplaces`, `github.com/anthropics/claude-code` issue tracker, `github.com/hesreallyhim/claude-code-json-schema`, real plugin repos across 3 marketplaces.

---

## 1. Official `plugin.json` Schema

**Canonical source:** [code.claude.com/docs/en/plugins-reference](https://code.claude.com/docs/en/plugins-reference) (fetched 2026-04-19, section "Plugin manifest schema").

**File location (enforced):** `.claude-plugin/plugin.json` at plugin root. Only `plugin.json` belongs in that directory; `commands/`, `agents/`, `skills/`, `output-styles/`, `monitors/`, and `hooks/` must sit at the plugin root. ([docs source](https://code.claude.com/docs/en/plugins-reference), "Directory structure mistakes" Warning.)

**Manifest is optional.** If omitted, Claude Code auto-discovers components in default locations and derives the plugin name from the directory name. Use a manifest to provide metadata or custom component paths. ([docs source](https://code.claude.com/docs/en/plugins-reference), "Plugin manifest schema".)

### 1.1 Required fields

| Field | Type | Required | Description | Constraints |
|---|---|---|---|---|
| `name` | string | **YES** (when manifest exists) | Unique plugin identifier | Docs say "kebab-case, no spaces". Official pattern per unofficial schema: `^[a-z0-9]+(-[a-z0-9]+)*$`. Used for namespacing (`plugin-dev:agent-creator`). |

### 1.2 Metadata fields (optional per docs)

| Field | Type | Description | Constraints / Verified Behavior |
|---|---|---|---|
| `version` | string | Semantic version | Docs say optional. **But** `affaan-m/everything-claude-code/.claude-plugin/PLUGIN_SCHEMA_NOTES.md` asserts the validator requires it and installation may fail marketplace install without it. Official `plugin validate` CLI is lenient and accepts non-semver values like `1.0` per hesreallyhim FIXTURE-EVIDENCE. Format: `MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]`. Docs warning: "If you change your plugin's code but don't bump the version in plugin.json, your plugin's existing users won't see your changes due to caching." |
| `description` | string | Brief explanation | No stated length limit. |
| `author` | object | Author information | Subfields: `name` (string, required), `email` (string, optional), `url` (string, optional). **Observed:** Claude Desktop's `listAvailablePlugins` validator rejects marketplace entries missing `author` entirely (issue #33068). Object form ONLY; string form is not supported. |
| `homepage` | string | Documentation URL | URI format. |
| `repository` | string OR object | Source code URL | Docs example shows string only. Unofficial schema allows object form `{type, url, directory}`. hesreallyhim FIXTURE-EVIDENCE notes CLI sometimes rejects the object form. |
| `license` | string | SPDX license identifier | e.g. `MIT`, `Apache-2.0`. No formal SPDX validation in CLI. |
| `keywords` | array<string> | Discovery tags | Unique items recommended. |

### 1.3 Component path fields

All paths: must be relative to plugin root, must start with `./`, forward slashes only, no `..` traversal. ([docs source](https://code.claude.com/docs/en/plugins-reference), "Path behavior rules".)

| Field | Type | Semantics | Verified Gotchas |
|---|---|---|---|
| `skills` | string OR array<string> | Replaces default `skills/` scan | Accepts directory path or explicit file paths. |
| `commands` | string OR array<string> | Replaces default `commands/` scan | Accepts directory or file paths. |
| `agents` | string OR array<string> | Replaces default `agents/` scan | **FOOTGUN:** Public docs example shows a string, but the validator rejects bare directory strings. Use array of explicit `.md` file paths. See `affaan-m/everything-claude-code/PLUGIN_SCHEMA_NOTES.md` and issue #44777 (string `./.claude/agents` rejected as "Invalid input"). |
| `hooks` | string OR array<string> OR object | Path(s) to hooks JSON, or inline hooks config | **FOOTGUN:** Do NOT reference the default `./hooks/hooks.json` here — Claude Code v2.1+ auto-loads it by convention and duplicates cause a load error. See Section 4. |
| `mcpServers` | string OR array OR object | Path(s) to `.mcp.json`, or inline MCP server map | Inline object = `{serverName: {command, args?, env?, cwd?}}` per unofficial schema. |
| `outputStyles` | string OR array<string> | Replaces default `output-styles/` scan | |
| `lspServers` | string OR array OR object | Path to `.lsp.json` or inline LSP server map | Required subfields per inline: `command`, `extensionToLanguage`. Optional: `args`, `transport` ("stdio" or "socket"), `env`, `initializationOptions`, `settings`, `workspaceFolder`, `startupTimeout`, `shutdownTimeout`, `restartOnCrash`, `maxRestarts`. |
| `monitors` | string OR array OR object | Path to `monitors/monitors.json` or inline array | Requires Claude Code v2.1.105+. Each entry: `name`, `command`, `description` (required); `when` optional. |
| `userConfig` | object | Values prompted at plugin enable-time | Keys must be valid identifiers (`^[A-Za-z_][A-Za-z0-9_]*$`). Each value: `{description: string, sensitive: boolean}`. Sensitive values go to the OS keychain (~2KB total limit shared with OAuth tokens). |
| `channels` | array<object> | Message injection channels (Telegram, Slack, Discord-style) | Each entry: `server` (required, must match a key in `mcpServers`), optional `userConfig` (same shape as top-level). |
| `dependencies` | array<string OR object> | Other plugins this one requires | Entry: `"name"` string OR `{name, version?}` where `version` is a semver range. |
| `settings` | object | Plugin default settings applied when enabled | Docs: "Only the `agent` and `subagentStatusLine` keys are currently supported." |

### 1.4 Inline hook configuration (when `hooks` is an object)

Per [docs/plugins-reference](https://code.claude.com/docs/en/plugins-reference) "Hooks" section:

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          { "type": "command", "command": "${CLAUDE_PLUGIN_ROOT}/scripts/format-code.sh" }
        ]
      }
    ]
  }
}
```

**Valid event names** (complete list from docs): `SessionStart`, `UserPromptSubmit`, `PreToolUse`, `PermissionRequest`, `PermissionDenied`, `PostToolUse`, `PostToolUseFailure`, `Notification`, `SubagentStart`, `SubagentStop`, `TaskCreated`, `TaskCompleted`, `Stop`, `StopFailure`, `TeammateIdle`, `InstructionsLoaded`, `ConfigChange`, `CwdChanged`, `FileChanged`, `WorktreeCreate`, `WorktreeRemove`, `PreCompact`, `PostCompact`, `Elicitation`, `ElicitationResult`, `SessionEnd`. Event names are **case-sensitive**.

**Valid hook types:** `command`, `http`, `prompt`, `agent`.

For `type: "command"`, `command` field is required. For `type: "prompt"` or `type: "agent"`, `prompt` field is required.

### 1.5 Environment variable substitution

Paths and commands may reference:
- `${CLAUDE_PLUGIN_ROOT}` — absolute install path (changes per version)
- `${CLAUDE_PLUGIN_DATA}` — persistent data dir surviving updates
- `${user_config.KEY}` — userConfig values
- `${ENV_VAR}` — any env var

---

## 2. Official `marketplace.json` Schema

**Canonical source:** [code.claude.com/docs/en/plugin-marketplaces](https://code.claude.com/docs/en/plugin-marketplaces) (fetched 2026-04-19).

**File location (enforced):** `.claude-plugin/marketplace.json` at repository root. Rejected in any other location (issue #46786).

### 2.1 Required top-level fields

| Field | Type | Description | Constraints |
|---|---|---|---|
| `name` | string | Marketplace identifier | Kebab-case, no spaces. **Reserved name list enforced** (see Section 3). |
| `owner` | object | Maintainer info | `name` (string, required); `email` (string, optional). |
| `plugins` | array<object> | Plugin entries | May be empty (produces warning). |

### 2.2 Optional top-level fields

| Field | Type | Description |
|---|---|---|
| `$schema` | string | JSON Schema self-reference URI. **Note:** Some older CLI versions incorrectly ran reserved-name regex against `$schema` value (issue #20423, closed fixed). The Anthropic-published URL `https://anthropic.com/claude-code/marketplace.schema.json` is referenced in `anthropics/claude-plugins-official/.claude-plugin/marketplace.json` but the URL itself returns 404. |
| `description` | string | Brief marketplace description (may live at root OR inside `metadata`) |
| `version` | string | Marketplace version (may live at root OR inside `metadata`) |
| `metadata` | object | Alternate container: `{description?, version?, pluginRoot?}`. `pluginRoot` is a base dir prepended to relative plugin source paths. |

### 2.3 `plugins` array entry

Per [docs plugin-marketplaces](https://code.claude.com/docs/en/plugin-marketplaces), each entry may include **any plugin manifest field** plus four marketplace-specific fields.

**Required per entry:**

| Field | Type | Description |
|---|---|---|
| `name` | string | Plugin identifier, kebab-case. Must be unique within the marketplace — "Duplicate plugin name x found in marketplace" is an explicit error. |
| `source` | string OR object | Where to fetch the plugin (see 2.4). |

**Optional per entry** (inherited from plugin manifest): `description`, `version`, `author`, `homepage`, `repository`, `license`, `keywords`, `category`, `tags`, `skills`, `commands`, `agents`, `hooks`, `mcpServers`, `lspServers`.

**Marketplace-only extras:**
- `category` (string) — for organization
- `tags` (array<string>) — searchability
- `strict` (boolean, default `true`) — if `false`, marketplace entry is the authoritative definition and the plugin's own `plugin.json` should not declare components (conflict = load failure).

### 2.4 Plugin `source` types

Per [docs plugin-marketplaces "Plugin sources"](https://code.claude.com/docs/en/plugin-marketplaces):

| Source | Type | Required Fields | Optional Fields | Notes |
|---|---|---|---|---|
| Relative path | string | — | — | Must start with `./`. No `..`. Resolved vs marketplace root (not `.claude-plugin/`). Only works for git-based marketplaces. |
| `github` | object | `source`="github", `repo` (in `owner/repo` format) | `ref`, `sha` | `sha` must be 40-hex. |
| `url` | object | `source`="url", `url` (git URL) | `ref`, `sha` | `.git` suffix optional. |
| `git-subdir` | object | `source`="git-subdir", `url`, `path` | `ref`, `sha` | Sparse clone for monorepos. Previously broke catalog loading in CLI <2.1.77 (issue #33739). |
| `npm` | object | `source`="npm", `package` | `version`, `registry` | Public or private npm registry. |

---

## 3. Reserved Names Catalog

**Primary source:** [code.claude.com/docs/en/plugin-marketplaces#marketplace-schema](https://code.claude.com/docs/en/plugin-marketplaces) — official reserved names note. Cross-referenced against issues #14145, #18232, #18329, #46786.

### 3.1 Explicitly reserved marketplace names (exact-match list from official docs)

- `claude-code-marketplace`
- `claude-code-plugins`
- `claude-plugins-official`
- `anthropic-marketplace`
- `anthropic-plugins`
- `agent-skills`
- `knowledge-work-plugins`
- `life-sciences`

These can only be used when the marketplace source is a GitHub repository under the `anthropics/` organization (issue #14145 confirms: "can only be used with GitHub sources from the 'anthropics' organization").

### 3.2 Impersonation pattern rejection

Official docs: "Names that impersonate official marketplaces (like `official-claude-plugins` or `anthropic-tools-v2`) are also blocked."

Error message observed (issue #20423, #18232):

```
Invalid schema: name: Marketplace name cannot impersonate official
Anthropic/Claude marketplaces. Names containing "official", "anthropic",
or "claude" in official-sounding combinations are reserved.
```

**CAUTION — overly broad enforcement (issue #18232, 2026, CLOSED):** The current regex rejects legitimate third-party names like `dagster-claude-plugins`, `claude-dagster`, `dagster-claude-code`. Validator is over-aggressive. Hanko should reject these at `submit-check --marketplace anthropic` but can flag as warning only for other marketplaces.

### 3.3 Plugin name reservation

**Not found in any source.** Plugin names (inside the `plugins` array) do NOT appear to have a reserved-prefix catalog. The only constraint is kebab-case uniqueness within the marketplace. `claude-*` and `anthropic-*` plugin names exist in the official marketplace (e.g. `claude-code-setup`, `claude-opus-4-5-migration`) confirming no prefix-level block.

### 3.4 `$schema` field cross-contamination bug

Issue #20423 (CLOSED): the reserved-name regex was previously applied to the `$schema` value, causing valid marketplaces that referenced `https://anthropic.com/claude-code/marketplace.schema.json` to fail validation. Bug confirmed fixed but is worth noting — Hanko should only scan the `name` string, not all string fields.

---

## 4. Known Validation Footguns (verified 2026-04-19)

### 4.1 Opaque submission errors — #46786 (CLOSED as duplicate, root issue still open)

**Status:** Live. The whole raison d'être for Hanko.

**What breaks:** Marketplace validator emits vague errors like `expected object, received string` without schema path or examples. Users report a 7-attempt "fix and retry" loop.

**Hanko catches:** Yes. Hanko's contribution is precisely the local/CI validation that gives a clear error up-front with schema path, expected shape, and docs link.

### 4.2 `listAvailablePlugins` missing-author validation — #33068 (CLOSED as not Claude-Code)

**Status:** Live in Claude Desktop. Many external plugins in the official marketplace currently ship without `author` and Claude Desktop's validator refuses to load the entire marketplace catalog.

**Affected manifests in production:** `atlassian`, `figma`, `Notion`, `sentry`, `slack`, `vercel`, `pinecone`, `huggingface-skills`, `circleback`, `superpowers`, `posthog`, `coderabbit`, `sonatype-guide`, `firecrawl`, `qodo-skills`, `semgrep`, `postman`, `greptile`, `serena`, `playwright`, `github`, `supabase`, `asana`, `linear`, `gitlab`, `laravel-boost`, `firebase`, `context7`, `stripe`.

**Hanko catches:** Yes at manifest level. We can flag "author missing — Claude Desktop will refuse to list this marketplace." Note: this error is only enforced by Desktop, not the CLI, so it's a warning not an error.

### 4.3 `claude-plugins-official` marketplace load failure — #33739 (CLOSED, fixed in v2.1.77+, regressed briefly on Windows)

**Status:** Fixed. But the root cause is documentable: a `git-subdir` source (used by `semgrep` at `plugins[56]`) broke schema validation and the ENTIRE marketplace became unavailable rather than just the one plugin. This is a runtime CLI bug; `git-subdir` is now an officially documented source type.

**Hanko catches:** Partial. Hanko can validate that all declared source types are from the documented set (`relative path` / `github` / `url` / `git-subdir` / `npm`). If a future unknown source type appears, we warn.

### 4.4 Hook `${CLAUDE_PLUGIN_ROOT}` version cache — #18517 (OPEN)

**Status:** Live and affects any plugin that ships hooks. Settings.json stores the absolute versioned path of hook scripts at install time. When the plugin updates, settings.json is NOT rewritten. Old cache dir is cleaned and the hook breaks with `No such file or directory`. Also affects `statusLine` configs.

**Hanko catches:** Runtime-only. Not a manifest-level problem; Hanko v0 cannot catch this. **Flag for roadmap:** potential Hanko v1 feature: "lint my installed plugins for broken hook paths" — a separate CLI subcommand.

### 4.5 Duplicate `hooks` declaration — VERIFIED in v2.1+

**Status:** Live. Confirmed in `affaan-m/everything-claude-code/.claude-plugin/PLUGIN_SCHEMA_NOTES.md`, which documents 4 fix/revert commits (`22ad036`, `a7bc5f2`, `779085e`, `e3a1306`) where contributors flipped between adding and removing `"hooks": "./hooks/hooks.json"`. Exact CLI error: `Duplicate hooks file detected: ./hooks/hooks.json resolves to already-loaded file. The standard hooks/hooks.json is loaded automatically, so manifest.hooks should only reference additional hook files.`

**Hanko catches:** **YES at manifest level.** This is a prime Hanko rule: if `hooks` is a string or array equal to `./hooks/hooks.json`, reject with a specific error explaining the auto-load convention.

### 4.6 `agents` as directory string — VERIFIED (issue #44777, CLOSED)

**Status:** Live as of v2.1.92. Docs example shows `"agents": "./agents"` as valid. The runtime validator rejects directory paths for `agents` with generic `agents: Invalid input`. Works only with explicit array of `.md` file paths.

**Hanko catches:** YES — warn when `agents` is a bare directory path. Error message in issue #44777 also shows `.claude/`-prefixed paths are rejected even when the directory is inside the plugin.

### 4.7 Docs / CLI contradictions (from hesreallyhim FIXTURE-EVIDENCE)

- Docs mandate kebab-case plugin names; CLI accepts camelCase (enforcement level 1).
- Docs mandate semver `version`; CLI accepts `1.0` (enforcement level 1).
- Schema rejects additional top-level marketplace fields; CLI accepts them (enforcement level 1).
- Schema accepts `repository` object form; CLI sometimes rejects (enforcement level 1 via "complex-manifest-forms").

**Hanko design implication:** offer both `--strict` (reject docs violations) and `--lenient` (only reject what CLI actually rejects) modes.

---

## 5. Real-world Plugin Manifest Examples

All 10 saved under `fixtures/valid/` and linked by repo below.

| # | Plugin | Repo | Note |
|---|---|---|---|
| 1 | `commit-commands` | [anthropics/claude-plugins-official](https://github.com/anthropics/claude-plugins-official/blob/main/plugins/commit-commands/.claude-plugin/plugin.json) | Official minimal: name + description + author. No version. |
| 2 | `code-review` | [anthropics/claude-plugins-official](https://github.com/anthropics/claude-plugins-official/blob/main/plugins/code-review/.claude-plugin/plugin.json) | Same minimal pattern. |
| 3 | `example-plugin` | [anthropics/claude-plugins-official](https://github.com/anthropics/claude-plugins-official/blob/main/plugins/example-plugin/.claude-plugin/plugin.json) | Anthropic's own reference implementation. No version. |
| 4 | `explanatory-output-style` | [anthropics/claude-plugins-official](https://github.com/anthropics/claude-plugins-official/blob/main/plugins/explanatory-output-style/.claude-plugin/plugin.json) | Uses `version`. Clean. |
| 5 | `claude-code-setup` | [anthropics/claude-plugins-official](https://github.com/anthropics/claude-plugins-official/blob/main/plugins/claude-code-setup/.claude-plugin/plugin.json) | Uses `version`. Note `claude-` prefix allowed for plugin names. |
| 6 | `chrome-devtools-mcp` | [ChromeDevTools/chrome-devtools-mcp](https://github.com/ChromeDevTools/chrome-devtools-mcp/blob/main/.claude-plugin/plugin.json) | Inline `mcpServers` object. Pitfall: **no author** — would trigger #33068 if included in Anthropic marketplace. |
| 7 | `grafana` | [grafana/mcp-grafana](https://github.com/grafana/mcp-grafana/blob/main/.claude-plugin/plugin.json) | Uses `${CLAUDE_PLUGIN_ROOT}` in args. Clean. |
| 8 | `oh-my-claudecode` | [Yeachan-Heo/oh-my-claudecode](https://github.com/Yeachan-Heo/oh-my-claudecode/blob/main/.claude-plugin/plugin.json) | Uses `./skills/` string path and `./.mcp.json` for mcpServers. Comprehensive metadata. |
| 9 | `mempalace` | [MemPalace/mempalace](https://github.com/MemPalace/mempalace/blob/main/.claude-plugin/plugin.json) | Empty `commands: []` array. Inline mcpServers with python3. |
| 10 | `socraticode` | [giancarloerra/SocratiCode](https://github.com/giancarloerra/SocratiCode/blob/main/.claude-plugin/plugin.json) | Full author object with url. AGPL-3.0-only license. Clean example. |

### 5.1 Broken/invalid fixtures

Saved under `fixtures/invalid/` with `_why_invalid` key:

| # | File | Failure mode | Source |
|---|---|---|---|
| 1 | `agents-as-string-directory.json` | `agents: "./agents/"` — directory strings rejected | PLUGIN_SCHEMA_NOTES + #44777 |
| 2 | `name-not-kebab-case.json` | `name: "MyPlugin"` — PascalCase violates docs spec | docs + hesreallyhim FIXTURE-EVIDENCE |
| 3 | `duplicate-hooks-declaration.json` | `hooks: "./hooks/hooks.json"` — conflicts with v2.1+ auto-load | PLUGIN_SCHEMA_NOTES "DO NOT ADD" |
| 4 | `path-traversal-above-root.json` | `agents: ["../shared/..."]` — outside plugin root | docs "Path traversal limitations" |
| 5 | `reserved-marketplace-name.json` | Marketplace name `claude-code-plugins` is reserved | docs + #14145, #18232 |

**Additional real broken manifests identified but not saved** (no pristine raw JSON captured):
- `jasonkneen/agent-skills` uses a top-level `components` field not recognized by any schema.
- `Saik0s/mcp-browser-use` hooks value is `{"SessionStart": "hooks/SessionStart.py"}` — path string (no `./` prefix) and not wrapped in a matcher-group array.
- `anam-org/metaxy` missing `version` (would fail per PLUGIN_SCHEMA_NOTES assertion).
- `microsoft/azure-skills` uses `skills: "./skills/"` (string, docs say valid; runtime behavior for directory strings on `skills` is less clear than for `agents`).

---

## 6. Marketplace-specific Validation Quirks

| Marketplace | Repo | Additional rules beyond spec | Source |
|---|---|---|---|
| **Anthropic official** | [anthropics/claude-plugins-official](https://github.com/anthropics/claude-plugins-official) | Reserved marketplace names enforced (Section 3). `author` object strongly recommended to avoid Desktop validation failure (#33068). Plugin submissions via [clau.de/plugin-directory-submission](https://clau.de/plugin-directory-submission) form. No in-repo CONTRIBUTING; internal vs external plugin structure documented in README. | README at repo root |
| **buildwithclaude** | [davepoon/buildwithclaude](https://github.com/davepoon/buildwithclaude) | Directory naming: `plugins/<type>-<category>/.claude-plugin/plugin.json` (e.g., `agents-security`, `commands-git`). Forces agent/command/hook/skill into typed category bundles. More opinionated than Anthropic. | [CONTRIBUTING.md](https://github.com/davepoon/buildwithclaude/blob/main/CONTRIBUTING.md) |
| **cc-marketplace** | [ananddtyagi/cc-marketplace](https://github.com/ananddtyagi/cc-marketplace) | Has its own `PLUGIN_SCHEMA.md` that **requires** `name`, `version`, and `description` (3 required fields vs Anthropic's 1). Python validator: `scripts/validate-plugin-schema.py`. Requires README.md updates on plugin change. | [PLUGIN_SCHEMA.md](https://github.com/ananddtyagi/cc-marketplace/blob/main/PLUGIN_SCHEMA.md) |
| **claudemarketplaces.com** | [mertbuilds/claudemarketplaces.com](https://github.com/mertbuilds/claudemarketplaces.com) | **Auto-discovery only.** Site crawls GitHub daily for `.claude-plugin/marketplace.json`. No submission form. No PR workflow. No additional rules beyond the Anthropic schema. | Site `/about` page |
| **aitmpl.com** | (hosted at aitmpl.com, no public marketplace repo found; repo `aitmpl/aitmpl-com` returns 404) | UNVERIFIED. Presented on aitmpl.com as "Claude Code Plugins & Marketplaces directory." Likely auto-discovery; no contributing workflow found. | no direct source |

**Hanko CLI implication:** The `--marketplace <name>` flag should layer rules on top of the base Anthropic schema:
- `anthropic` → strict reserved-name check + author-object strongly required (warning)
- `buildwithclaude` → enforce directory naming convention
- `cc-marketplace` → `version` and `description` required (not optional)
- `claudemarketplaces` / `aitmpl` → no delta from base rules

---

## 7. Existing Validators' Feature Sets

### 7.1 `hesreallyhim/claude-code-json-schema` (4 stars, last push Feb 2026)

**Repo:** [github.com/hesreallyhim/claude-code-json-schema](https://github.com/hesreallyhim/claude-code-json-schema)

**What it is:** Raw JSON Schema 2020-12 files (`schemas/plugin.schema.json` 10,340 bytes, `schemas/marketplace.schema.json` 7,635 bytes) plus a thin Node CLI in `bin/claude-code-schema-lint.mjs` that wraps Ajv.

**What it validates:**
- Full plugin.json schema with `required: ["name"]`, kebab-case name pattern, semver version pattern, `additionalProperties: false`
- Marketplace.json with `required: ["name", "owner", "plugins"]`
- Inline hook matcher groups and handler conditional shapes (if type=command require command, etc.)
- MCP and LSP inline server shapes
- Path starts with `./` and no `..` traversal
- 40-hex SHA for github/url sources

**What it misses (Hanko opportunities):**
1. **No CLI UX.** Ajv errors are raw and opaque: `"should be string"` at a JSON Pointer path. No human-readable "Your plugin is missing an author field — here's a fix" formatting.
2. **No reserved-name check.** Schema does not encode the Anthropic reserved marketplace name list.
3. **No duplicate-hooks-declaration check** (Section 4.5).
4. **No docs/CLI divergence mode.** Cannot distinguish "strict docs" vs "what CLI actually enforces" — self-described in README as a known limitation.
5. **No `--marketplace <name>` override.** Per-marketplace rule layering is absent.
6. **Node/npm dependency.** We're building Go.
7. **Dormant.** Last commit Feb 2026; single contributor.
8. **No `--fix`** auto-remediation.
9. **Error aggregation is crude.** First-fail output; no structured "here are all 12 problems" summary.

**Minimum bar for Hanko to beat it:** pretty errors, reserved-name enforcement, duplicate-hooks check, Go binary with zero deps, `--fix-safe` for common auto-fixes (add missing `version`, convert `agents` directory string to array, strip `./hooks/hooks.json` hooks declaration).

### 7.2 `agent-sh/agnix` (189 stars)

**Repo:** [github.com/agent-sh/agnix](https://github.com/agent-sh/agnix)

**What it validates:** CLAUDE.md, AGENTS.md, SKILL.md, hooks configs, MCP configs, plus Codex CLI plugin manifests. Plugin.json support exists but is SHALLOW: `crates/agnix-core/src/schemas/plugin.rs` defines only 8 fields (name, description, version, author, homepage, repository, license, keywords) — NO component path fields, NO hooks/mcpServers/lspServers/monitors/userConfig/channels/dependencies.

**What it misses (confirms Hanko gap):**
- No component-path validation (`commands`, `agents`, `skills`, `hooks`, `mcpServers`, `lspServers`, `monitors`).
- No marketplace.json schema at all (grep found zero hits for marketplace.json validation).
- No reserved-name enforcement.
- No per-marketplace rule layering.

Agnix has rich knowledge base content for hooks/skills/agents (see `knowledge-base/standards/claude-code-HARD-RULES.md`) but hasn't wired it into plugin.json validation. **Competitive gap confirmed.**

### 7.3 `carlrannaberg/cclint` (16 stars)

**Repo:** [github.com/carlrannaberg/cclint](https://github.com/carlrannaberg/cclint)

**Scope explicitly limited** per README: "Agent/Subagent Linting, Command Linting, Settings Validation (`.claude/settings.json`), Documentation Linting (CLAUDE.md)."

**Does NOT validate plugin.json or marketplace.json.** Grep for `plugin.json` in src yielded no hits. **No competitive overlap.**

### 7.4 Competitive summary

| Tool | plugin.json | marketplace.json | Reserved names | Duplicate-hooks | Go native | CLI polish |
|---|---|---|---|---|---|---|
| hesreallyhim schema | deep | deep | no | no | no (Node) | raw Ajv |
| agnix | shallow (8 fields) | no | no | no | no (Rust/Node) | mature |
| cclint | no | no | no | no | no (Node) | mature |
| **Hanko target** | deep | deep | yes | yes | yes | submission-focused |

---

## 8. Recommended JSON Schema for Hanko to Embed

A JSON Schema 2020-12 starting draft. This is a synthesis of (a) the official docs as extracted in Sections 1-2, (b) the unofficial hesreallyhim schema, and (c) additional Hanko rules (reserved names, duplicate-hooks, author warning). Keep as `internal/schema/plugin.schema.json` and `internal/schema/marketplace.schema.json` in the Go repo, embedded via `//go:embed`.

### 8.1 Plugin schema draft

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://roninforge.org/hanko/plugin.schema.json",
  "title": "Claude Code Plugin Manifest",
  "type": "object",
  "additionalProperties": false,
  "required": ["name"],
  "properties": {
    "$schema": { "type": "string", "format": "uri" },
    "name": {
      "type": "string",
      "description": "Plugin identifier. Kebab-case per docs.",
      "pattern": "^[a-z0-9]+(-[a-z0-9]+)*$"
    },
    "version": {
      "type": "string",
      "description": "Semantic version (MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]).",
      "pattern": "^(0|[1-9]\\d*)\\.(0|[1-9]\\d*)\\.(0|[1-9]\\d*)(?:-[0-9A-Za-z-]+(?:\\.[0-9A-Za-z-]+)*)?(?:\\+[0-9A-Za-z-]+(?:\\.[0-9A-Za-z-]+)*)?$"
    },
    "description": { "type": "string", "minLength": 1 },
    "author": { "$ref": "#/$defs/authorObject" },
    "homepage": { "type": "string", "format": "uri" },
    "repository": {
      "oneOf": [
        { "type": "string", "format": "uri" },
        { "$ref": "#/$defs/repositoryObject" }
      ]
    },
    "license": { "type": "string", "minLength": 1 },
    "keywords": { "type": "array", "items": { "type": "string" }, "uniqueItems": true },
    "category": { "type": "string" },
    "tags": { "type": "array", "items": { "type": "string" } },
    "skills": { "$ref": "#/$defs/pathOrPathArray" },
    "commands": { "$ref": "#/$defs/pathOrPathArray" },
    "agents": { "$ref": "#/$defs/pathOrPathArray" },
    "outputStyles": { "$ref": "#/$defs/pathOrPathArray" },
    "hooks": {
      "oneOf": [
        { "$ref": "#/$defs/pathOrPathArray" },
        { "$ref": "#/$defs/hooksInline" }
      ]
    },
    "mcpServers": {
      "oneOf": [
        { "$ref": "#/$defs/pathOrPathArray" },
        { "$ref": "#/$defs/mcpInline" }
      ]
    },
    "lspServers": {
      "oneOf": [
        { "$ref": "#/$defs/pathOrPathArray" },
        { "$ref": "#/$defs/lspInline" }
      ]
    },
    "monitors": {
      "oneOf": [
        { "$ref": "#/$defs/pathOrPathArray" },
        { "$ref": "#/$defs/monitorsInline" }
      ]
    },
    "userConfig": { "$ref": "#/$defs/userConfigObject" },
    "channels": { "type": "array", "items": { "$ref": "#/$defs/channelEntry" } },
    "dependencies": {
      "type": "array",
      "items": {
        "oneOf": [
          { "type": "string" },
          { "$ref": "#/$defs/dependencyEntry" }
        ]
      }
    },
    "settings": { "type": "object", "additionalProperties": true }
  },
  "$defs": {
    "authorObject": {
      "type": "object",
      "additionalProperties": false,
      "required": ["name"],
      "properties": {
        "name": { "type": "string", "minLength": 1 },
        "email": { "type": "string", "format": "email" },
        "url": { "type": "string", "format": "uri" }
      }
    },
    "repositoryObject": {
      "type": "object",
      "additionalProperties": false,
      "required": ["type", "url"],
      "properties": {
        "type": { "type": "string" },
        "url": { "type": "string", "format": "uri" },
        "directory": { "type": "string" }
      }
    },
    "relativePath": {
      "type": "string",
      "pattern": "^\\./",
      "allOf": [
        { "not": { "pattern": "\\.\\." } },
        { "not": { "pattern": "\\\\" } }
      ]
    },
    "pathOrPathArray": {
      "oneOf": [
        { "$ref": "#/$defs/relativePath" },
        {
          "type": "array",
          "minItems": 1,
          "items": { "$ref": "#/$defs/relativePath" }
        }
      ]
    },
    "hooksInline": {
      "type": "object",
      "patternProperties": {
        "^(SessionStart|UserPromptSubmit|PreToolUse|PermissionRequest|PermissionDenied|PostToolUse|PostToolUseFailure|Notification|SubagentStart|SubagentStop|TaskCreated|TaskCompleted|Stop|StopFailure|TeammateIdle|InstructionsLoaded|ConfigChange|CwdChanged|FileChanged|WorktreeCreate|WorktreeRemove|PreCompact|PostCompact|Elicitation|ElicitationResult|SessionEnd)$": {
          "type": "array",
          "items": { "$ref": "#/$defs/hookMatcherGroup" }
        }
      },
      "additionalProperties": false
    },
    "hookMatcherGroup": {
      "type": "object",
      "additionalProperties": false,
      "required": ["hooks"],
      "properties": {
        "matcher": { "type": "string" },
        "hooks": {
          "type": "array",
          "minItems": 1,
          "items": { "$ref": "#/$defs/hookHandler" }
        }
      }
    },
    "hookHandler": {
      "type": "object",
      "required": ["type"],
      "properties": {
        "type": { "enum": ["command", "http", "prompt", "agent"] },
        "timeout": { "type": "number", "minimum": 0 },
        "statusMessage": { "type": "string" },
        "once": { "type": "boolean" },
        "command": { "type": "string" },
        "url": { "type": "string", "format": "uri" },
        "async": { "type": "boolean" },
        "prompt": { "type": "string" },
        "model": { "type": "string" }
      },
      "allOf": [
        { "if": { "properties": { "type": { "const": "command" } } }, "then": { "required": ["command"] } },
        { "if": { "properties": { "type": { "const": "http" } } }, "then": { "required": ["url"] } },
        { "if": { "properties": { "type": { "const": "prompt" } } }, "then": { "required": ["prompt"] } },
        { "if": { "properties": { "type": { "const": "agent" } } }, "then": { "required": ["prompt"] } }
      ]
    },
    "mcpInline": {
      "type": "object",
      "additionalProperties": { "$ref": "#/$defs/mcpServer" }
    },
    "mcpServer": {
      "type": "object",
      "additionalProperties": false,
      "required": ["command"],
      "properties": {
        "command": { "type": "string" },
        "args": { "type": "array", "items": { "type": "string" } },
        "env": { "type": "object", "additionalProperties": { "type": "string" } },
        "cwd": { "type": "string" }
      }
    },
    "lspInline": {
      "type": "object",
      "additionalProperties": { "$ref": "#/$defs/lspServer" }
    },
    "lspServer": {
      "type": "object",
      "additionalProperties": false,
      "required": ["command", "extensionToLanguage"],
      "properties": {
        "command": { "type": "string" },
        "args": { "type": "array", "items": { "type": "string" } },
        "extensionToLanguage": {
          "type": "object",
          "minProperties": 1,
          "propertyNames": { "pattern": "^\\.[A-Za-z0-9]+$" },
          "additionalProperties": { "type": "string" }
        },
        "transport": { "enum": ["stdio", "socket"] },
        "env": { "type": "object", "additionalProperties": { "type": "string" } },
        "initializationOptions": {},
        "settings": {},
        "workspaceFolder": { "type": "string" },
        "startupTimeout": { "type": "integer", "minimum": 0 },
        "shutdownTimeout": { "type": "integer", "minimum": 0 },
        "restartOnCrash": { "type": "boolean" },
        "maxRestarts": { "type": "integer", "minimum": 0 }
      }
    },
    "monitorsInline": {
      "type": "array",
      "items": {
        "type": "object",
        "additionalProperties": false,
        "required": ["name", "command", "description"],
        "properties": {
          "name": { "type": "string" },
          "command": { "type": "string" },
          "description": { "type": "string" },
          "when": { "type": "string", "pattern": "^(always|on-skill-invoke:[a-z0-9-]+)$" }
        }
      }
    },
    "userConfigObject": {
      "type": "object",
      "patternProperties": {
        "^[A-Za-z_][A-Za-z0-9_]*$": {
          "type": "object",
          "additionalProperties": false,
          "required": ["description"],
          "properties": {
            "description": { "type": "string" },
            "sensitive": { "type": "boolean" }
          }
        }
      },
      "additionalProperties": false
    },
    "channelEntry": {
      "type": "object",
      "additionalProperties": false,
      "required": ["server"],
      "properties": {
        "server": { "type": "string" },
        "userConfig": { "$ref": "#/$defs/userConfigObject" }
      }
    },
    "dependencyEntry": {
      "type": "object",
      "additionalProperties": false,
      "required": ["name"],
      "properties": {
        "name": { "type": "string" },
        "version": { "type": "string" }
      }
    }
  }
}
```

### 8.2 Additional Hanko-only rules (implement as Go code alongside JSON Schema)

These cannot easily be expressed in pure JSON Schema — apply after schema validation:

1. **Reserved marketplace names** (Section 3) — exact match list.
2. **Impersonation pattern** — regex: name matches `^(official-)?(anthropic|claude).*|.*anthropic.*|.*claude.*-(official|anthropic|claude).*`. Flag as warning for third-party marketplaces; error for Anthropic submission.
3. **Duplicate hooks declaration** — if `hooks` is a string or single-element array equal to `./hooks/hooks.json`, error with Section 4.5 explanation.
4. **Agents as bare directory** — if `agents` is a string that ends with `/` (directory) rather than `.md`, warn with Section 4.6 explanation.
5. **Author missing** — if `author` is absent, warn (marketplace listAvailablePlugins may reject).
6. **Version missing** — if `version` is absent, warn (per PLUGIN_SCHEMA_NOTES; not all tooling enforces).
7. **Cross-check `channels[].server` references a key in `mcpServers`** — cannot be expressed in draft 2020-12 without `$data`.
8. **Cross-check paths resolve to actual files/dirs on disk** — only when Hanko is run against a plugin source tree, not a manifest in isolation.

### 8.3 Marketplace schema draft

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://roninforge.org/hanko/marketplace.schema.json",
  "title": "Claude Code Marketplace Catalog",
  "type": "object",
  "additionalProperties": false,
  "required": ["name", "owner", "plugins"],
  "properties": {
    "$schema": { "type": "string", "format": "uri" },
    "name": {
      "type": "string",
      "pattern": "^[a-z0-9]+(-[a-z0-9]+)*$"
    },
    "description": { "type": "string" },
    "version": { "type": "string" },
    "owner": {
      "type": "object",
      "additionalProperties": false,
      "required": ["name"],
      "properties": {
        "name": { "type": "string" },
        "email": { "type": "string", "format": "email" },
        "url": { "type": "string", "format": "uri" }
      }
    },
    "metadata": {
      "type": "object",
      "additionalProperties": false,
      "properties": {
        "description": { "type": "string" },
        "version": { "type": "string" },
        "pluginRoot": {
          "type": "string",
          "allOf": [
            { "not": { "pattern": "^/" } },
            { "not": { "pattern": "\\.\\." } }
          ]
        }
      }
    },
    "plugins": {
      "type": "array",
      "items": { "$ref": "#/$defs/pluginEntry" }
    }
  },
  "$defs": {
    "sha40": { "type": "string", "pattern": "^[a-fA-F0-9]{40}$" },
    "relativeOrRootedPath": {
      "type": "string",
      "allOf": [
        { "not": { "pattern": "^/" } },
        { "not": { "pattern": "\\.\\." } },
        { "not": { "pattern": "\\\\" } }
      ]
    },
    "githubSource": {
      "type": "object",
      "additionalProperties": false,
      "required": ["source", "repo"],
      "properties": {
        "source": { "const": "github" },
        "repo": { "type": "string", "pattern": "^[^/\\s]+/[^/\\s]+$" },
        "ref": { "type": "string" },
        "sha": { "$ref": "#/$defs/sha40" }
      }
    },
    "urlSource": {
      "type": "object",
      "additionalProperties": false,
      "required": ["source", "url"],
      "properties": {
        "source": { "const": "url" },
        "url": { "type": "string" },
        "ref": { "type": "string" },
        "sha": { "$ref": "#/$defs/sha40" }
      }
    },
    "gitSubdirSource": {
      "type": "object",
      "additionalProperties": false,
      "required": ["source", "url", "path"],
      "properties": {
        "source": { "const": "git-subdir" },
        "url": { "type": "string" },
        "path": { "type": "string" },
        "ref": { "type": "string" },
        "sha": { "$ref": "#/$defs/sha40" }
      }
    },
    "npmSource": {
      "type": "object",
      "additionalProperties": false,
      "required": ["source", "package"],
      "properties": {
        "source": { "const": "npm" },
        "package": { "type": "string" },
        "version": { "type": "string" },
        "registry": { "type": "string", "format": "uri" }
      }
    },
    "source": {
      "oneOf": [
        { "$ref": "#/$defs/relativeOrRootedPath" },
        { "$ref": "#/$defs/githubSource" },
        { "$ref": "#/$defs/urlSource" },
        { "$ref": "#/$defs/gitSubdirSource" },
        { "$ref": "#/$defs/npmSource" }
      ]
    },
    "pluginEntry": {
      "type": "object",
      "additionalProperties": false,
      "required": ["name", "source"],
      "properties": {
        "name": { "type": "string", "pattern": "^[a-z0-9]+(-[a-z0-9]+)*$" },
        "source": { "$ref": "#/$defs/source" },
        "description": { "type": "string" },
        "version": { "type": "string" },
        "author": {
          "type": "object",
          "additionalProperties": false,
          "required": ["name"],
          "properties": {
            "name": { "type": "string" },
            "email": { "type": "string", "format": "email" },
            "url": { "type": "string", "format": "uri" }
          }
        },
        "homepage": { "type": "string", "format": "uri" },
        "repository": { "type": "string", "format": "uri" },
        "license": { "type": "string" },
        "keywords": { "type": "array", "items": { "type": "string" } },
        "category": { "type": "string" },
        "tags": { "type": "array", "items": { "type": "string" } },
        "strict": { "type": "boolean", "default": true },
        "commands": { "$ref": "#/$defs/pathOrArray" },
        "agents": { "$ref": "#/$defs/pathOrArray" },
        "skills": { "$ref": "#/$defs/pathOrArray" },
        "hooks": {},
        "mcpServers": {},
        "lspServers": {}
      }
    },
    "pathOrArray": {
      "oneOf": [
        { "type": "string", "pattern": "^\\./" },
        { "type": "array", "items": { "type": "string", "pattern": "^\\./" } }
      ]
    }
  }
}
```

**Apply after-schema check:** `plugins[].name` values are unique within the array.

---

## 9. Confidence Register

| # | Claim | Level | Evidence |
|---|---|---|---|
| 1 | `plugin.json` lives in `.claude-plugin/` and `name` is the sole required field | VERIFIED | [code.claude.com/docs/en/plugins-reference](https://code.claude.com/docs/en/plugins-reference) "Required fields" table |
| 2 | Full plugin manifest schema listed in Section 1 (24 fields) | VERIFIED | Same source, "Plugin manifest schema" and sub-sections |
| 3 | `marketplace.json` requires `name`, `owner`, `plugins` | VERIFIED | [code.claude.com/docs/en/plugin-marketplaces](https://code.claude.com/docs/en/plugin-marketplaces) "Required fields" table |
| 4 | Reserved marketplace names are the 8 names in Section 3.1 | VERIFIED | Same source, Note block under "Required fields" |
| 5 | Impersonation regex rejects names with "official"/"anthropic"/"claude" in official-sounding combinations | VERIFIED | Issue #18232, #20423, official error message quoted verbatim |
| 6 | `agents` as bare directory path is rejected | VERIFIED | Issue #44777 (CLOSED), PLUGIN_SCHEMA_NOTES.md in everything-claude-code |
| 7 | Duplicate hooks declaration footgun (`./hooks/hooks.json`) | VERIFIED | PLUGIN_SCHEMA_NOTES.md documents 4 commits of fix/revert with exact CLI error |
| 8 | `version` is mandatory in practice despite docs saying optional | PARTIALLY VERIFIED | PLUGIN_SCHEMA_NOTES.md asserts so; hesreallyhim FIXTURE-EVIDENCE says CLI accepts versions like `1.0` as pass (lenient). So "required for marketplace install" is plausible but not proven universally. |
| 9 | `listAvailablePlugins` rejects plugins missing `author` in Desktop only | VERIFIED | Issue #33068, CLOSED as "not Claude Code" — Desktop-only enforcement |
| 10 | `git-subdir` was added as source type and previously caused catalog load failure in CLI <2.1.77 | VERIFIED | Issue #33739 and official docs now document it |
| 11 | Hook `${CLAUDE_PLUGIN_ROOT}` path caching is a runtime bug not catchable at manifest level | VERIFIED | Issue #18517 OPEN, design clearly runtime |
| 12 | hesreallyhim schema has known CLI divergence (kebab-case, semver, extra props) | VERIFIED | Repo's own docs/FIXTURE-EVIDENCE.md |
| 13 | agnix plugin.rs validates only 8 fields, no component paths | VERIFIED | Direct source read of `crates/agnix-core/src/schemas/plugin.rs` |
| 14 | cclint does not validate plugin.json or marketplace.json | VERIFIED | Grep on repo source returned zero hits; README confirms scope |
| 15 | aitmpl.com hosts a plugin directory but no public repo found | UNVERIFIED | `aitmpl/aitmpl-com` 404s. Site visible in search results only. Hanko can ship without aitmpl-specific rules; treat as base Anthropic schema. |
| 16 | buildwithclaude requires typed-category directory naming (`agents-*`, `commands-*`, etc.) | VERIFIED | [davepoon/buildwithclaude/CONTRIBUTING.md](https://github.com/davepoon/buildwithclaude/blob/main/CONTRIBUTING.md) Project Structure section |
| 17 | cc-marketplace requires `name` + `version` + `description` as REQUIRED (stricter than Anthropic) | VERIFIED | [ananddtyagi/cc-marketplace/PLUGIN_SCHEMA.md](https://github.com/ananddtyagi/cc-marketplace/blob/main/PLUGIN_SCHEMA.md) |
| 18 | claudemarketplaces.com uses automatic GitHub discovery only, no submission form | PARTIALLY VERIFIED | Site description in search results. Repo `mertbuilds/claudemarketplaces.com` contains the static site; did not read it to confirm zero submission flow. |
| 19 | Reserved-name regex previously cross-contaminated `$schema` validation | VERIFIED | Issue #20423, CLOSED |
| 20 | 24 valid hook event names enumerated in Section 1.4 | VERIFIED | [code.claude.com/docs/en/plugins-reference](https://code.claude.com/docs/en/plugins-reference) Hooks section |
| 21 | 4 hook types: command, http, prompt, agent | VERIFIED | Same source, "Hook types" subsection |
| 22 | Plugin name `claude-*` prefix is NOT reserved (only marketplace names) | VERIFIED | `anthropics/claude-plugins-official/plugins/` contains `claude-code-setup`, `claude-opus-4-5-migration`, `claude-md-management` |
| 23 | Paths must start with `./`, no `..`, no backslashes | VERIFIED | docs/plugins-reference "Path behavior rules" + "Path traversal limitations" |
| 24 | `strict: false` in marketplace entry + components in `plugin.json` = load failure | VERIFIED | docs/plugin-marketplaces "Strict mode" table |
| 25 | `${CLAUDE_PLUGIN_ROOT}` vs `${CLAUDE_PLUGIN_DATA}` distinction exists | VERIFIED | docs/plugins-reference "Environment variables" section |
| 26 | Hanko's CLI competitor gap is genuine (no tool validates plugin.json + marketplace.json + reserved names + duplicate-hooks in Go) | VERIFIED | Sections 7.1, 7.2, 7.3 with source reads |

---

## Summary for Hanko builder

**Actionable ship list for v0:**
1. Embed the two schemas from Section 8 as `//go:embed`.
2. Use a Go JSON Schema lib: `github.com/santhosh-tekuri/jsonschema/v6` (supports 2020-12, pure Go, no CGO).
3. Layer Go validators on top of schema for: reserved names (Section 3), duplicate hooks (4.5), agents-directory (4.6), author-missing warning (4.2), version-missing warning (4.7), and `channels[].server` cross-reference.
4. Implement `hanko check <path>` (validate single manifest) and `hanko submit-check --marketplace <name> <path>` (add marketplace-specific rules).
5. Pretty errors with schema path, offending value, docs URL, and a one-line fix suggestion.
6. `--fix-safe` for reversible auto-fixes (add missing `version: 0.1.0`, strip `hooks: ./hooks/hooks.json`, convert `agents` directory-string to empty-array-needs-files placeholder).
7. `--strict` / `--lenient` flag to toggle docs-violations-are-errors vs CLI-enforced-only.
8. Fixtures already at `fixtures/valid/` (10 manifests) and `fixtures/invalid/` (5 manifests) — wire these as test data.

**Known unknowns worth bounty research later (not blockers):**
- aitmpl.com marketplace delta (if any).
- `outputStyles` runtime schema for entries (docs mention but no detailed shape).
- Whether `monitors` `when` accepts values beyond `always` and `on-skill-invoke:<skill>`.
- Exact conflict rules when `strict: true` AND both `plugin.json` and marketplace entry define components.
