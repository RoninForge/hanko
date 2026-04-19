package rules

import (
	"strings"
	"testing"

	"github.com/RoninForge/hanko/internal/report"
)

func TestDuplicateHooksDeclaration(t *testing.T) {
	tests := []struct {
		name     string
		manifest map[string]any
		wantRule bool
	}{
		{
			name:     "no hooks",
			manifest: map[string]any{"name": "x"},
			wantRule: false,
		},
		{
			name:     "default hooks path",
			manifest: map[string]any{"name": "x", "hooks": "./hooks/hooks.json"},
			wantRule: true,
		},
		{
			name:     "default hooks path without dot-slash",
			manifest: map[string]any{"name": "x", "hooks": "hooks/hooks.json"},
			wantRule: true,
		},
		{
			name:     "custom hooks path",
			manifest: map[string]any{"name": "x", "hooks": "./hooks/custom.json"},
			wantRule: false,
		},
		{
			name:     "default in an array",
			manifest: map[string]any{"name": "x", "hooks": []any{"./hooks/hooks.json"}},
			wantRule: true,
		},
		{
			name:     "inline hooks object",
			manifest: map[string]any{"name": "x", "hooks": map[string]any{"PreToolUse": []any{}}},
			wantRule: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findings := DuplicateHooksDeclaration(tt.manifest)
			got := hasRule(findings, "HANKO001")
			if got != tt.wantRule {
				t.Errorf("HANKO001 fired=%v want=%v (findings=%v)", got, tt.wantRule, findings)
			}
		})
	}
}

func TestAgentsAsBareDirectory(t *testing.T) {
	tests := []struct {
		name     string
		manifest map[string]any
		wantRule bool
	}{
		{"absent", map[string]any{"name": "x"}, false},
		{"bare directory", map[string]any{"name": "x", "agents": "./agents/"}, true},
		{"directory without trailing slash", map[string]any{"name": "x", "agents": "./agents"}, true},
		{"single .md file as string", map[string]any{"name": "x", "agents": "./agents/one.md"}, false},
		{"array of files", map[string]any{"name": "x", "agents": []any{"./agents/one.md"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findings := AgentsAsBareDirectory(tt.manifest)
			got := hasRule(findings, "HANKO002")
			if got != tt.wantRule {
				t.Errorf("HANKO002 fired=%v want=%v", got, tt.wantRule)
			}
		})
	}
}

func TestAuthorMissing(t *testing.T) {
	withAuthor := map[string]any{
		"name":   "x",
		"author": map[string]any{"name": "Someone"},
	}
	withoutAuthor := map[string]any{"name": "x"}

	if got := AuthorMissing(withAuthor); len(got) != 0 {
		t.Errorf("author present should produce no findings, got %v", got)
	}
	findings := AuthorMissing(withoutAuthor)
	if !hasRule(findings, "HANKO003") {
		t.Errorf("HANKO003 should fire when author is missing, got %v", findings)
	}
	// All findings must be warnings at this severity.
	for _, f := range findings {
		if f.Severity != report.SeverityWarning {
			t.Errorf("HANKO003 should be warning severity, got %v", f.Severity)
		}
	}
}

func TestVersionMissing(t *testing.T) {
	withVersion := map[string]any{"name": "x", "version": "0.1.0"}
	withoutVersion := map[string]any{"name": "x"}

	if got := VersionMissing(withVersion); len(got) != 0 {
		t.Errorf("version present should produce no findings, got %v", got)
	}
	findings := VersionMissing(withoutVersion)
	if !hasRule(findings, "HANKO004") {
		t.Error("HANKO004 should fire when version is missing")
	}
}

func TestReservedMarketplaceName(t *testing.T) {
	for name := range reservedMarketplaceNames {
		findings := ReservedMarketplaceName(map[string]any{"name": name})
		if !hasRule(findings, "HANKO101") {
			t.Errorf("reserved name %q should fire HANKO101", name)
		}
	}

	for _, safe := range []string{"acme-plugins", "mytools", "cool-claude-addons"} {
		findings := ReservedMarketplaceName(map[string]any{"name": safe})
		if hasRule(findings, "HANKO101") {
			t.Errorf("non-reserved name %q should not fire HANKO101", safe)
		}
	}
}

func TestImpersonationPattern(t *testing.T) {
	// Names that should warn via HANKO102.
	shouldWarn := []string{
		"official-claude-plugins",
		"anthropic-tools-v2",
		"claude-marketplace-official",
	}
	for _, n := range shouldWarn {
		findings := ImpersonationPattern(map[string]any{"name": n})
		if !hasRule(findings, "HANKO102") {
			t.Errorf("impersonation name %q should fire HANKO102", n)
		}
	}

	// Exact-match reserved names are caught by HANKO101 and must NOT
	// double-fire HANKO102.
	for name := range reservedMarketplaceNames {
		findings := ImpersonationPattern(map[string]any{"name": name})
		if hasRule(findings, "HANKO102") {
			t.Errorf("reserved name %q should be caught by HANKO101 only, not HANKO102", name)
		}
	}
}

func TestDuplicatePluginNames(t *testing.T) {
	manifest := map[string]any{
		"plugins": []any{
			map[string]any{"name": "foo", "source": "./p/foo"},
			map[string]any{"name": "bar", "source": "./p/bar"},
			map[string]any{"name": "foo", "source": "./p/foo2"},
		},
	}
	findings := DuplicatePluginNames(manifest)
	if !hasRule(findings, "HANKO103") {
		t.Errorf("duplicate plugin names should fire HANKO103, got %v", findings)
	}
	// Sanity: the path points at the second occurrence.
	for _, f := range findings {
		if f.Rule == "HANKO103" && !strings.Contains(f.Path, "plugins/2") {
			t.Errorf("HANKO103 path should point at the duplicate (plugins/2), got %q", f.Path)
		}
	}
}

func TestForMarketplace(t *testing.T) {
	// All recognised marketplace names must return a non-empty rule set.
	for _, m := range MarketplaceNames() {
		s := ForMarketplace(m)
		if len(s.Plugin) == 0 && len(s.Marketplace) == 0 {
			t.Errorf("ForMarketplace(%q) returned empty rule set", m)
		}
	}
	// Unknown falls back to base.
	base := Base()
	unknown := ForMarketplace("completely-made-up")
	if len(base.Plugin) != len(unknown.Plugin) {
		t.Errorf("unknown marketplace should fall back to Base, got %d plugin rules want %d",
			len(unknown.Plugin), len(base.Plugin))
	}
}

// TestForMarketplaceAppliesStrictOverlays runs the rule sets returned by
// ForMarketplace against a minimal manifest missing recommended fields,
// and asserts the marketplace name actually picked up the strict overlay.
// The lighter TestForMarketplace above would pass even if the switch
// accidentally returned Base() for every name.
func TestForMarketplaceAppliesStrictOverlays(t *testing.T) {
	manifest := map[string]any{"name": "x"} // missing author, version, description

	tests := []struct {
		marketplace      string
		wantErrorRules   []string
		forbidErrorRules []string
		wantWarningRules []string
	}{
		{
			marketplace:      "anthropic",
			wantErrorRules:   []string{"HANKO003-strict"},
			forbidErrorRules: []string{"HANKO003"},
			wantWarningRules: []string{"HANKO004"},
		},
		{
			marketplace:      "cc-marketplace",
			wantErrorRules:   []string{"HANKO004-strict", "HANKO005-strict"},
			forbidErrorRules: []string{"HANKO004"},
			wantWarningRules: []string{"HANKO003"},
		},
		{
			marketplace:      "buildwithclaude",
			wantErrorRules:   nil,
			forbidErrorRules: []string{"HANKO003-strict", "HANKO004-strict"},
			wantWarningRules: []string{"HANKO003", "HANKO004"},
		},
		{
			marketplace:      "claudemarketplaces",
			wantErrorRules:   nil,
			forbidErrorRules: []string{"HANKO003-strict", "HANKO004-strict"},
			wantWarningRules: []string{"HANKO003", "HANKO004"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.marketplace, func(t *testing.T) {
			findings := Apply(ForMarketplace(tt.marketplace).Plugin, manifest)
			var errs, warns []string
			for _, f := range findings {
				switch f.Severity {
				case report.SeverityError:
					errs = append(errs, f.Rule)
				case report.SeverityWarning:
					warns = append(warns, f.Rule)
				}
			}
			for _, rule := range tt.wantErrorRules {
				if !hasRule(findings, rule) {
					t.Errorf("%s: expected error rule %q, got errors=%v warnings=%v",
						tt.marketplace, rule, errs, warns)
				}
			}
			for _, rule := range tt.forbidErrorRules {
				for _, f := range findings {
					if f.Rule == rule && f.Severity == report.SeverityError {
						t.Errorf("%s: rule %q must not fire as error (it should have been replaced by the strict variant)",
							tt.marketplace, rule)
					}
				}
			}
			for _, rule := range tt.wantWarningRules {
				if !hasRule(findings, rule) {
					t.Errorf("%s: expected warning rule %q, got errors=%v warnings=%v",
						tt.marketplace, rule, errs, warns)
				}
			}
		})
	}
}

func TestJSONPointer(t *testing.T) {
	tests := []struct {
		name     string
		segments []any
		want     string
	}{
		{"empty", nil, ""},
		{"single string", []any{"name"}, "/name"},
		{"nested", []any{"plugins", 2, "name"}, "/plugins/2/name"},
		{"escaped slash", []any{"a/b"}, "/a~1b"},
		{"escaped tilde", []any{"a~b"}, "/a~0b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := jsonPointer(tt.segments...)
			if got != tt.want {
				t.Errorf("jsonPointer(%v) = %q, want %q", tt.segments, got, tt.want)
			}
		})
	}
}

func hasRule(findings []report.Finding, rule string) bool {
	for _, f := range findings {
		if f.Rule == rule {
			return true
		}
	}
	return false
}

func TestApply(t *testing.T) {
	manifest := map[string]any{"name": "x"} // no author, no version
	rs := Base().Plugin
	findings := Apply(rs, manifest)
	// Base plugin rules should flag AuthorMissing + VersionMissing as warnings.
	if !hasRule(findings, "HANKO003") {
		t.Error("Apply should surface HANKO003 from AuthorMissing")
	}
	if !hasRule(findings, "HANKO004") {
		t.Error("Apply should surface HANKO004 from VersionMissing")
	}
}

func TestStrictOverlays(t *testing.T) {
	manifest := map[string]any{"name": "x"} // missing author, version, description

	// Anthropic overlay promotes HANKO003 → HANKO003-strict at error severity.
	// Critically: HANKO003 (the base warning) must NOT also fire, or users
	// see a duplicate finding for the same root cause.
	anthropic := StrictAnthropic()
	findings := Apply(anthropic.Plugin, manifest)
	if !hasRule(findings, "HANKO003-strict") {
		t.Errorf("Anthropic overlay should fire HANKO003-strict, got %v", findings)
	}
	if hasRule(findings, "HANKO003") {
		t.Errorf("Anthropic overlay must replace HANKO003 with HANKO003-strict, not fire both. got %v", findings)
	}
	var sawStrictError bool
	for _, f := range findings {
		if f.Rule == "HANKO003-strict" && f.Severity == report.SeverityError {
			sawStrictError = true
		}
	}
	if !sawStrictError {
		t.Error("HANKO003-strict should carry error severity under Anthropic overlay")
	}

	// cc-marketplace overlay promotes HANKO004 → HANKO004-strict and adds
	// HANKO005-strict for the missing description. Same replace-not-append
	// invariant applies: HANKO004 (base warning) must not fire alongside.
	cc := CCMarketplace()
	ccFindings := Apply(cc.Plugin, manifest)
	if !hasRule(ccFindings, "HANKO004-strict") {
		t.Error("cc-marketplace overlay should fire HANKO004-strict")
	}
	if hasRule(ccFindings, "HANKO004") {
		t.Errorf("cc-marketplace overlay must replace HANKO004 with HANKO004-strict, not fire both. got %v", ccFindings)
	}
	if !hasRule(ccFindings, "HANKO005-strict") {
		t.Error("cc-marketplace overlay should fire HANKO005-strict")
	}

	// buildwithclaude currently returns base rules.
	bw := BuildWithClaude()
	if len(bw.Plugin) != len(Base().Plugin) {
		t.Error("BuildWithClaude should currently return the base rule set")
	}

	// claudemarketplaces is auto-discovery and currently a base pass-through,
	// but the case is explicit in ForMarketplace so the CLI can document it.
	cm := ClaudeMarketplaces()
	if len(cm.Plugin) != len(Base().Plugin) {
		t.Error("ClaudeMarketplaces should currently return the base rule set")
	}
}

func TestVersionMissingStrictOnlyFiresWhenMissing(t *testing.T) {
	withVersion := map[string]any{"name": "x", "version": "1.0.0"}
	findings := versionMissingStrict(withVersion)
	if len(findings) != 0 {
		t.Errorf("versionMissingStrict should be silent when version is present, got %v", findings)
	}
}

func TestDescriptionMissingStrictOnlyFiresWhenMissing(t *testing.T) {
	withDesc := map[string]any{"name": "x", "description": "a thing"}
	findings := descriptionMissingStrict(withDesc)
	if len(findings) != 0 {
		t.Errorf("descriptionMissingStrict should be silent when description is present, got %v", findings)
	}
}
