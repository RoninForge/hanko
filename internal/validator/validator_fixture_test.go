package validator_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/RoninForge/hanko/internal/report"
	"github.com/RoninForge/hanko/internal/validator"
)

// fixtureCase describes one expectation against a fixture on disk.
//
// wantErrorRules / wantWarningRules are SUBSET checks: every rule listed
// MUST appear in the report, but extra findings are permitted. This keeps
// the table robust when new rules are added that do not contradict the
// intent of the fixture.
type fixtureCase struct {
	file             string // relative to the package's testdata search path
	kind             validator.Kind
	marketplace      string
	wantClean        bool // no errors at all (warnings allowed)
	wantErrorRules   []string
	wantWarningRules []string
}

// testdataPath resolves to the repo-level testdata/ directory regardless of
// which package the test lives in. Using filepath.Join with an absolute-ish
// `../../testdata/...` keeps the Makefile target simple.
func testdataPath(t *testing.T, parts ...string) string {
	t.Helper()
	// Walk upward from the working directory until we find a dir that has
	// a testdata subdir. Go sets the working dir to the package dir for
	// tests, so we go up two levels from internal/validator/.
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	// Start from wd, walk up until testdata/ is present.
	dir := wd
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "testdata")); err == nil {
			return filepath.Join(append([]string{dir, "testdata"}, parts...)...)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatalf("could not locate testdata/ walking up from %s", wd)
	return ""
}

func loadFixture(t *testing.T, rel ...string) []byte {
	t.Helper()
	p := testdataPath(t, rel...)
	data, err := os.ReadFile(p) //nolint:gosec // test-only path
	if err != nil {
		t.Fatalf("read fixture %s: %v", p, err)
	}
	return data
}

func runCase(t *testing.T, c fixtureCase) {
	t.Helper()
	v, err := validator.New()
	if err != nil {
		t.Fatalf("validator.New: %v", err)
	}
	data := loadFixture(t, strings.Split(c.file, "/")...)
	r, err := v.Validate(data, validator.Options{
		Kind:        c.kind,
		Marketplace: c.marketplace,
		File:        c.file,
	})
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	gotErrs, gotWarns := collectRules(r)

	if c.wantClean && r.HasErrors() {
		t.Errorf("%s expected clean but got errors: %v", c.file, gotErrs)
	}
	for _, rule := range c.wantErrorRules {
		if !contains(gotErrs, rule) {
			t.Errorf("%s expected error rule %q to fire. errors=%v warnings=%v",
				c.file, rule, gotErrs, gotWarns)
		}
	}
	for _, rule := range c.wantWarningRules {
		if !contains(gotWarns, rule) {
			t.Errorf("%s expected warning rule %q to fire. errors=%v warnings=%v",
				c.file, rule, gotErrs, gotWarns)
		}
	}
}

func collectRules(r *report.Report) (errs, warns []string) {
	for _, f := range r.Findings {
		switch f.Severity {
		case report.SeverityError:
			errs = append(errs, f.Rule)
		case report.SeverityWarning:
			warns = append(warns, f.Rule)
		}
	}
	return
}

func contains(set []string, s string) bool {
	for _, x := range set {
		if x == s {
			return true
		}
	}
	return false
}

// expectedWarningsForValidFixture hard-codes the warnings each real-world
// valid fixture is known to emit. Keeping this in code (rather than a
// sibling YAML or JSON) makes the expected shape part of the review
// diff: any regression that changes a warning to an error, or drops an
// expected warning, will fail the test with a clear message.
func expectedWarningsForValidFixture(name string) []string {
	switch name {
	case "chrome-devtools-mcp.json":
		// Missing `author`.
		return []string{"HANKO003"}
	case "code-review.json", "commit-commands.json", "example-plugin.json":
		// Missing `version`.
		return []string{"HANKO004"}
	default:
		// Fully clean plugins. Expect no warnings at all.
		return nil
	}
}

// TestValidFixtures walks testdata/valid/ and asserts every file passes
// without errors AND that its warnings match the expected-warnings map.
// This turns what was previously a pass-on-any-finding test into a
// tight regression guard.
func TestValidFixtures(t *testing.T) {
	dir := testdataPath(t, "valid")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read testdata/valid: %v", err)
	}
	v, err := validator.New()
	if err != nil {
		t.Fatalf("validator.New: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		name := e.Name()
		t.Run(name, func(t *testing.T) {
			data := loadFixture(t, "valid", name)
			r, err := v.Validate(data, validator.Options{
				Kind: validator.KindPlugin,
				File: name,
			})
			if err != nil {
				t.Fatalf("Validate: %v", err)
			}
			if r.HasErrors() {
				errs, _ := collectRules(r)
				t.Fatalf("valid fixture produced error-severity findings: %v", errs)
			}
			_, warns := collectRules(r)
			want := expectedWarningsForValidFixture(name)

			// Subset check both ways: every expected warning must fire,
			// and no unexpected warning may fire. If a rule is changed
			// or a new rule is added, this test forces the author to
			// update the expectations map — which is the whole point.
			for _, exp := range want {
				if !contains(warns, exp) {
					t.Errorf("expected warning %q not found. actual warnings=%v", exp, warns)
				}
			}
			for _, got := range warns {
				if !contains(want, got) {
					t.Errorf("unexpected warning %q. expected=%v actual=%v", got, want, warns)
				}
			}
		})
	}
}

// TestInvalidFixtures asserts each invalid fixture triggers its specific
// rule. The expected rule IDs come from the fixture names themselves to
// keep this table easy to extend.
func TestInvalidFixtures(t *testing.T) {
	cases := []fixtureCase{
		{
			file:             "invalid/agents-as-string-directory.json",
			kind:             validator.KindPlugin,
			wantErrorRules:   []string{"HANKO002"},
			wantWarningRules: nil,
		},
		{
			file:           "invalid/duplicate-hooks-declaration.json",
			kind:           validator.KindPlugin,
			wantErrorRules: []string{"HANKO001"},
		},
		{
			file:           "invalid/name-not-kebab-case.json",
			kind:           validator.KindPlugin,
			wantErrorRules: []string{"HANKO-SCHEMA"},
		},
		{
			file:           "invalid/path-traversal-above-root.json",
			kind:           validator.KindPlugin,
			wantErrorRules: []string{"HANKO-SCHEMA"},
		},
		{
			file:           "invalid/reserved-marketplace-name.json",
			kind:           validator.KindMarketplace,
			wantErrorRules: []string{"HANKO101"},
		},
		{
			file:             "invalid/missing-author.json",
			kind:             validator.KindPlugin,
			wantWarningRules: []string{"HANKO003"},
		},
		{
			// Same fixture under --marketplace=anthropic promotes HANKO003
			// to an error via the strict overlay.
			file:           "invalid/missing-author.json",
			kind:           validator.KindPlugin,
			marketplace:    "anthropic",
			wantErrorRules: []string{"HANKO003-strict"},
		},
		{
			file:           "invalid/duplicate-plugin-names.json",
			kind:           validator.KindMarketplace,
			wantErrorRules: []string{"HANKO103"},
		},
	}
	for _, c := range cases {
		name := c.file
		if c.marketplace != "" {
			name += " [" + c.marketplace + "]"
		}
		t.Run(name, func(t *testing.T) {
			runCase(t, c)
		})
	}
}

// TestMarketplaceOverlay verifies the cc-marketplace overlay promotes
// missing-version and missing-description into errors.
func TestMarketplaceOverlay_CCMarketplace(t *testing.T) {
	v, err := validator.New()
	if err != nil {
		t.Fatalf("validator.New: %v", err)
	}
	raw := []byte(`{"name":"x"}`)
	r, err := v.Validate(raw, validator.Options{
		Kind:        validator.KindPlugin,
		Marketplace: "cc-marketplace",
	})
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	errs, _ := collectRules(r)
	for _, rule := range []string{"HANKO004-strict", "HANKO005-strict"} {
		if !contains(errs, rule) {
			t.Errorf("cc-marketplace overlay: expected %q, got errors=%v", rule, errs)
		}
	}
}

// TestInvalidJSON returns a single HANKO000 finding, not a crash.
func TestInvalidJSON(t *testing.T) {
	v, err := validator.New()
	if err != nil {
		t.Fatalf("validator.New: %v", err)
	}
	r, err := v.Validate([]byte(`{"name": "broken" `), validator.Options{
		Kind: validator.KindPlugin,
	})
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if !r.HasErrors() {
		t.Fatal("expected errors for malformed JSON")
	}
	errs, _ := collectRules(r)
	if !contains(errs, "HANKO000") {
		t.Errorf("expected HANKO000, got %v", errs)
	}
}

// TestEmptyObject triggers the HANKO-SCHEMA "required name" path.
func TestEmptyObject(t *testing.T) {
	v, err := validator.New()
	if err != nil {
		t.Fatalf("validator.New: %v", err)
	}
	r, err := v.Validate([]byte(`{}`), validator.Options{Kind: validator.KindPlugin})
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if !r.HasErrors() {
		t.Fatal("expected schema errors for empty object")
	}
}

// TestInstancePathEscaping exercises the RFC 6901 escape on a synthetic
// manifest whose mcpServers inline object has a key with a reserved
// character. The schema would accept the key (mcpServers.additionalProperties
// is permissive on names), so the schema error must come from a deeper
// failure (missing required `command`). The resulting InstanceLocation
// contains the tricky key and must round-trip as a valid JSON pointer.
func TestInstancePathEscaping(t *testing.T) {
	v, err := validator.New()
	if err != nil {
		t.Fatalf("validator.New: %v", err)
	}
	// mcpServer object without required `command` → schema error whose
	// instance location includes the weird key.
	raw := []byte(`{"name":"x","mcpServers":{"my~server":{}}}`)
	r, err := v.Validate(raw, validator.Options{Kind: validator.KindPlugin})
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if !r.HasErrors() {
		t.Fatal("expected schema error for mcpServer missing command")
	}
	var sawEscaped bool
	for _, f := range r.Findings {
		// Must be properly escaped: ~ → ~0. A naive concatenation would
		// produce "/mcpServers/my~server/command" and break consumers.
		if strings.Contains(f.Path, "my~0server") {
			sawEscaped = true
			break
		}
	}
	if !sawEscaped {
		paths := make([]string, 0, len(r.Findings))
		for _, f := range r.Findings {
			paths = append(paths, f.Path)
		}
		t.Errorf("expected a finding whose path contains \"my~0server\" (RFC 6901 escaped), got paths: %v", paths)
	}
}
