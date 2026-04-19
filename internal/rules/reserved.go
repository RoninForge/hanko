package rules

import (
	"regexp"

	"github.com/RoninForge/hanko/internal/report"
)

// reservedMarketplaceNames are the eight exact marketplace names Anthropic
// reserves for their own first-party listings. Using one of these names in
// a third-party marketplace.json produces a load failure.
//
// Source: code.claude.com/docs/en/plugin-marketplaces#marketplace-schema
// cross-referenced against anthropics/claude-code issues #14145, #18232.
var reservedMarketplaceNames = map[string]struct{}{
	"claude-code-marketplace": {},
	"claude-code-plugins":     {},
	"claude-plugins-official": {},
	"anthropic-marketplace":   {},
	"anthropic-plugins":       {},
	"agent-skills":            {},
	"knowledge-work-plugins":  {},
	"life-sciences":           {},
}

// impersonationPattern matches names that resemble official Anthropic or
// Claude marketplaces even when not on the exact-match list. Captures
// combinations like "official-claude-plugins", "anthropic-tools-v2",
// "claude-marketplace-official". Anthropic's own validator is stricter
// and is known to reject some legitimate third-party names (issue
// #18232), but our pattern is narrower by design; Hanko surfaces any
// match as a warning, never an error, regardless of the marketplace
// overlay. Authors submitting to Anthropic are responsible for reading
// HANKO102 and judging whether to rename before submission.
var impersonationPattern = regexp.MustCompile(`(?i)^(official-)?(anthropic|claude)[-_a-z0-9]*|[-_a-z0-9]*(anthropic|claude)[-_a-z0-9]*-(official|anthropic|claude)`)

// ReservedMarketplaceName flags marketplace.json files whose top-level
// name is on the reserved list. Rule ID: HANKO101.
func ReservedMarketplaceName(manifest any) []report.Finding {
	m, ok := manifest.(map[string]any)
	if !ok {
		return nil
	}
	name, ok := m["name"].(string)
	if !ok {
		return nil
	}
	if _, reserved := reservedMarketplaceNames[name]; !reserved {
		return nil
	}
	return []report.Finding{{
		Severity: report.SeverityError,
		Rule:     "HANKO101",
		Path:     "/name",
		Message:  "marketplace name \"" + name + "\" is reserved by Anthropic and can only be used by repos under the anthropics/ GitHub org",
		Fix:      "rename your marketplace. Common pattern: \"<your-github-username>-plugins\" or \"<topic>-claude-plugins\"",
		DocURL:   "https://code.claude.com/docs/en/plugin-marketplaces#marketplace-schema",
	}}
}

// ImpersonationPattern flags marketplace names that look like official
// Anthropic listings. Warning by default, promoted to error by the
// strict-anthropic overlay in rules.ForMarketplace("anthropic"). Rule
// ID: HANKO102.
func ImpersonationPattern(manifest any) []report.Finding {
	m, ok := manifest.(map[string]any)
	if !ok {
		return nil
	}
	name, ok := m["name"].(string)
	if !ok {
		return nil
	}
	// Exact-match reserved names are caught by HANKO101; skip them here
	// to avoid double-reporting the same bad name.
	if _, reserved := reservedMarketplaceNames[name]; reserved {
		return nil
	}
	if !impersonationPattern.MatchString(name) {
		return nil
	}
	return []report.Finding{{
		Severity: report.SeverityWarning,
		Rule:     "HANKO102",
		Path:     "/name",
		Message:  "marketplace name \"" + name + "\" matches the impersonation pattern for official Anthropic listings",
		Fix:      "consider renaming to avoid the \"official\"/\"anthropic\"/\"claude\" naming ambiguity. Anthropic's submission validator will reject this name",
		DocURL:   "https://code.claude.com/docs/en/plugin-marketplaces#marketplace-schema",
	}}
}

// DuplicatePluginNames flags marketplace catalogs where the plugins array
// contains two or more entries with the same name. Rule ID: HANKO103.
func DuplicatePluginNames(manifest any) []report.Finding {
	m, ok := manifest.(map[string]any)
	if !ok {
		return nil
	}
	plugins, ok := m["plugins"].([]any)
	if !ok {
		return nil
	}
	seen := make(map[string]int, len(plugins))
	var findings []report.Finding
	for i, p := range plugins {
		entry, ok := p.(map[string]any)
		if !ok {
			continue
		}
		name, ok := entry["name"].(string)
		if !ok || name == "" {
			continue
		}
		if prev, dup := seen[name]; dup {
			findings = append(findings, report.Finding{
				Severity: report.SeverityError,
				Rule:     "HANKO103",
				Path:     jsonPointer("plugins", i, "name"),
				Message:  "duplicate plugin name \"" + name + "\" (also at /plugins/" + itoa(prev) + "/name)",
				Fix:      "plugin names must be unique within a marketplace. Rename one of the entries",
				DocURL:   "https://code.claude.com/docs/en/plugin-marketplaces",
			})
			continue
		}
		seen[name] = i
	}
	return findings
}
