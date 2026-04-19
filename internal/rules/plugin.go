package rules

import (
	"strings"

	"github.com/RoninForge/hanko/internal/report"
)

// DuplicateHooksDeclaration flags the common v2.1+ footgun where a plugin
// declares `"hooks": "./hooks/hooks.json"` in plugin.json. Claude Code
// auto-loads that path by convention, so the explicit declaration causes
// a double-load error. Rule ID: HANKO001.
func DuplicateHooksDeclaration(manifest any) []report.Finding {
	m, ok := manifest.(map[string]any)
	if !ok {
		return nil
	}
	hooks, present := m["hooks"]
	if !present {
		return nil
	}

	// Walk the common shapes: string, single-element array, or both.
	isDefault := func(p string) bool {
		p = strings.TrimSpace(p)
		return p == "./hooks/hooks.json" || p == "hooks/hooks.json"
	}

	switch v := hooks.(type) {
	case string:
		if isDefault(v) {
			return []report.Finding{defaultHooksFinding(jsonPointer("hooks"))}
		}
	case []any:
		for i, item := range v {
			s, ok := item.(string)
			if !ok {
				continue
			}
			if isDefault(s) {
				return []report.Finding{defaultHooksFinding(jsonPointer("hooks", i))}
			}
		}
	}
	return nil
}

func defaultHooksFinding(path string) report.Finding {
	return report.Finding{
		Severity: report.SeverityError,
		Rule:     "HANKO001",
		Path:     path,
		Message:  "`hooks` explicitly references `./hooks/hooks.json`, which Claude Code v2.1+ auto-loads by convention. The duplicate declaration causes a load error",
		Fix:      "remove the `hooks` field entirely and let Claude Code auto-discover the default path, OR rename your hooks file and keep the explicit reference",
		DocURL:   "https://code.claude.com/docs/en/plugins-reference",
	}
}

// AgentsAsBareDirectory flags `"agents": "./agents"` or similar patterns
// where the value is a string pointing at a directory rather than explicit
// `.md` file paths. The docs example shows a directory string works, but
// the official validator rejects it (issue #44777). Rule ID: HANKO002.
func AgentsAsBareDirectory(manifest any) []report.Finding {
	m, ok := manifest.(map[string]any)
	if !ok {
		return nil
	}
	agents, present := m["agents"]
	if !present {
		return nil
	}
	s, isString := agents.(string)
	if !isString {
		return nil
	}
	// A string that ends in `.md` is a single file, which is fine.
	if strings.HasSuffix(s, ".md") {
		return nil
	}
	return []report.Finding{{
		Severity: report.SeverityError,
		Rule:     "HANKO002",
		Path:     "/agents",
		Message:  "`agents` is set to a bare directory string `" + s + "`. The official validator rejects this despite the docs example",
		Fix:      "use an array of explicit `.md` file paths, e.g. `[\"./agents/reviewer.md\", \"./agents/planner.md\"]`",
		DocURL:   "https://github.com/anthropics/claude-code/issues/44777",
	}}
}

// AuthorMissing is a warning for plugins with no `author` field. Claude
// Desktop's `listAvailablePlugins` validator refuses to load a catalog
// if any plugin is missing its author (issue #33068), which affects every
// user of that marketplace. Rule ID: HANKO003.
func AuthorMissing(manifest any) []report.Finding {
	m, ok := manifest.(map[string]any)
	if !ok {
		return nil
	}
	if _, present := m["author"]; present {
		return nil
	}
	return []report.Finding{{
		Severity: report.SeverityWarning,
		Rule:     "HANKO003",
		Path:     "",
		Message:  "`author` is absent. Claude Desktop's listAvailablePlugins refuses to load marketplaces that contain any plugin without an author",
		Fix:      "add `\"author\": { \"name\": \"Your Name\", \"email\": \"you@example.com\" }`",
		DocURL:   "https://github.com/anthropics/claude-code/issues/33068",
	}}
}

// authorMissingStrict is the strict-Anthropic overlay: same check, error
// severity, used by rules.StrictAnthropic().
func authorMissingStrict(manifest any) []report.Finding {
	out := AuthorMissing(manifest)
	for i := range out {
		if out[i].Rule == "HANKO003" {
			out[i].Severity = report.SeverityError
			out[i].Rule = "HANKO003-strict"
		}
	}
	return out
}

// VersionMissing is a warning for plugins that omit `version`. The field
// is optional per docs but Claude Code's install cache keys on it, so
// users of a version-less plugin silently miss updates. Rule ID: HANKO004.
func VersionMissing(manifest any) []report.Finding {
	m, ok := manifest.(map[string]any)
	if !ok {
		return nil
	}
	if _, present := m["version"]; present {
		return nil
	}
	return []report.Finding{{
		Severity: report.SeverityWarning,
		Rule:     "HANKO004",
		Path:     "",
		Message:  "`version` is absent. Claude Code's install cache keys on this field, so users will not see updates unless you set one",
		Fix:      "add `\"version\": \"0.1.0\"` and bump on every release",
		DocURL:   "https://code.claude.com/docs/en/plugins-reference",
	}}
}

// versionMissingStrict is the cc-marketplace overlay: same check, error
// severity, used by rules.CCMarketplace().
func versionMissingStrict(manifest any) []report.Finding {
	out := VersionMissing(manifest)
	for i := range out {
		if out[i].Rule == "HANKO004" {
			out[i].Severity = report.SeverityError
			out[i].Rule = "HANKO004-strict"
		}
	}
	return out
}

// descriptionMissingStrict is the cc-marketplace overlay: description is
// required, not optional. Rule ID: HANKO005-strict.
func descriptionMissingStrict(manifest any) []report.Finding {
	m, ok := manifest.(map[string]any)
	if !ok {
		return nil
	}
	if _, present := m["description"]; present {
		return nil
	}
	return []report.Finding{{
		Severity: report.SeverityError,
		Rule:     "HANKO005-strict",
		Path:     "",
		Message:  "`description` is required by the cc-marketplace submission rules",
		Fix:      "add a one-sentence `description` field explaining what the plugin does",
		DocURL:   "https://github.com/ananddtyagi/cc-marketplace/blob/main/PLUGIN_SCHEMA.md",
	}}
}
