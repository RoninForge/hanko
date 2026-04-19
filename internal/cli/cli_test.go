package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/RoninForge/hanko/internal/report"
)

func runCLI(args ...string) (stdout, stderr string, code int) {
	var out, errb bytes.Buffer
	code = Execute(&out, &errb, args)
	return out.String(), errb.String(), code
}

func TestVersionCommand(t *testing.T) {
	stdout, _, code := runCLI("version")
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if stdout == "" {
		t.Error("version command should write something to stdout")
	}
}

func TestHelpCommand(t *testing.T) {
	stdout, _, code := runCLI("--help")
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	for _, want := range []string{"hanko", "check", "submit-check", "schema"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("help output missing %q", want)
		}
	}
}

func TestSchemaPrintPlugin(t *testing.T) {
	stdout, _, code := runCLI("schema", "print", "--file", "plugin")
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout, "\"Claude Code Plugin Manifest\"") {
		t.Error("schema print --file plugin should emit the plugin schema")
	}
}

func TestSchemaPrintMarketplace(t *testing.T) {
	stdout, _, code := runCLI("schema", "print", "--file", "marketplace")
	if code != 0 {
		t.Errorf("schema print --file marketplace should exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "\"Claude Code Marketplace Catalog\"") {
		t.Error("schema print --file marketplace should emit the marketplace schema")
	}
}

func TestSubmitCheckMissingMarketplaceFlag(t *testing.T) {
	_, _, code := runCLI("submit-check", ".")
	if code == 0 {
		t.Error("submit-check without --marketplace should exit non-zero")
	}
}

// testdataFile resolves a relative path inside the repo-level testdata/
// directory by walking up from the current working dir.
func testdataFile(t *testing.T, rel string) string {
	t.Helper()
	dir, _ := os.Getwd()
	for i := 0; i < 6; i++ {
		candidate := filepath.Join(dir, "testdata", rel)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatalf("could not find testdata/%s", rel)
	return ""
}

func TestCheckOnValidFixture(t *testing.T) {
	file := testdataFile(t, "valid/grafana.json")
	stdout, _, code := runCLI("check", "--kind=plugin", "--color=false", file)
	if code != 0 {
		t.Errorf("valid fixture should exit 0, got %d\nstdout: %s", code, stdout)
	}
	if !strings.Contains(stdout, "clean") {
		t.Errorf("valid fixture should report clean, got: %s", stdout)
	}
}

func TestCheckOnInvalidFixture(t *testing.T) {
	file := testdataFile(t, "invalid/duplicate-hooks-declaration.json")
	stdout, _, code := runCLI("check", "--kind=plugin", "--color=false", file)
	if code == 0 {
		t.Errorf("invalid fixture should exit non-zero, stdout: %s", stdout)
	}
	if !strings.Contains(stdout, "HANKO001") {
		t.Errorf("expected HANKO001 in output, got: %s", stdout)
	}
}

func TestCheckJSONOutput(t *testing.T) {
	file := testdataFile(t, "invalid/reserved-marketplace-name.json")
	stdout, _, code := runCLI("check", "--kind=marketplace", "--json", file)
	if code == 0 {
		t.Error("reserved-marketplace-name fixture should exit non-zero")
	}
	// Single-file JSON output must still be a top-level array so
	// consumers can decode unconditionally into []Report.
	stdout = strings.TrimSpace(stdout)
	if !strings.HasPrefix(stdout, "[") {
		t.Errorf("JSON output should start with [, got: %s", stdout)
	}
	if !strings.Contains(stdout, "HANKO101") {
		t.Errorf("expected HANKO101 in JSON output, got: %s", stdout)
	}
	// Also confirm the array round-trips through json.Unmarshal into []Report.
	var got []report.Report
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("JSON output should round-trip into []Report, got: %v\nstdout: %s", err, stdout)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 report in array, got %d", len(got))
	}
}

func TestSubmitCheckAnthropic(t *testing.T) {
	file := testdataFile(t, "invalid/missing-author.json")
	stdout, _, code := runCLI("submit-check", "--marketplace=anthropic", "--kind=plugin", "--color=false", file)
	if code == 0 {
		t.Errorf("submit-check anthropic on missing-author should exit non-zero, stdout: %s", stdout)
	}
	if !strings.Contains(stdout, "HANKO003-strict") {
		t.Errorf("expected HANKO003-strict under --marketplace=anthropic, got: %s", stdout)
	}
}

func TestResolveKindErrors(t *testing.T) {
	_, _, code := runCLI("check", "--kind=bogus", "testdata/nope.json")
	if code == 0 {
		t.Error("unknown --kind should fail")
	}
}

// TestResolveKindByFilename exercises the filename-detection branch of
// resolveKind which the earlier CLI tests skipped by always passing
// --kind explicitly. Three cases: plugin.json auto-detected, a file
// with an unknown name rejected clearly.
func TestResolveKindByFilename(t *testing.T) {
	dir := t.TempDir()
	// plugin.json auto-detects as KindPlugin.
	pluginPath := filepath.Join(dir, "plugin.json")
	if err := os.WriteFile(pluginPath,
		[]byte(`{"name":"ok","version":"1.0.0","description":"y","author":{"name":"z"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout, _, code := runCLI("check", "--color=false", pluginPath)
	if code != 0 {
		t.Errorf("plugin.json should auto-detect and pass, got %d\n%s", code, stdout)
	}
	if !strings.Contains(stdout, "clean") {
		t.Errorf("auto-detected plugin.json should report clean, got: %s", stdout)
	}

	// marketplace.json auto-detects as KindMarketplace.
	marketplacePath := filepath.Join(dir, "marketplace.json")
	if err := os.WriteFile(marketplacePath,
		[]byte(`{"name":"ok","owner":{"name":"me"},"plugins":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout, _, code = runCLI("check", "--color=false", marketplacePath)
	if code != 0 {
		t.Errorf("marketplace.json should auto-detect and pass, got %d\n%s", code, stdout)
	}

	// Any other filename must fail with a clear error message instead of
	// guessing a kind.
	weird := filepath.Join(dir, "totally-unknown-name.json")
	if err := os.WriteFile(weird, []byte(`{"name":"x"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, stderr, code := runCLI("check", weird)
	if code == 0 {
		t.Error("unknown filename without --kind should fail")
	}
	if !strings.Contains(stderr, "cannot infer manifest kind") && !strings.Contains(stderr, "--kind") {
		t.Errorf("error message should guide the user toward --kind, got: %s", stderr)
	}
}

// TestCheckDirectory exercises the auto-discovery path where runCheck
// looks for .claude-plugin/plugin.json and .claude-plugin/marketplace.json
// under a directory. This also exercises the multi-report emit path.
func TestCheckDirectory(t *testing.T) {
	dir := t.TempDir()
	manifestDir := filepath.Join(dir, ".claude-plugin")
	if err := os.MkdirAll(manifestDir, 0o755); err != nil {
		t.Fatal(err)
	}
	plugin := `{"name":"dir-test","version":"1.0.0","description":"x","author":{"name":"x"}}`
	marketplace := `{"name":"dir-test-mkt","owner":{"name":"x"},"plugins":[{"name":"p","source":"./p"}]}`
	if err := os.WriteFile(filepath.Join(manifestDir, "plugin.json"), []byte(plugin), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(manifestDir, "marketplace.json"), []byte(marketplace), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout, _, code := runCLI("check", "--color=false", dir)
	if code != 0 {
		t.Errorf("directory check should succeed, got code=%d, stdout=%s", code, stdout)
	}
	// The output should contain two "clean" lines, one per manifest.
	if strings.Count(stdout, "clean") < 2 {
		t.Errorf("expected two clean reports, got: %s", stdout)
	}
}

func TestCheckDirectoryNoManifests(t *testing.T) {
	dir := t.TempDir()
	_, _, code := runCLI("check", dir)
	if code == 0 {
		t.Error("check on a dir without manifests should exit non-zero")
	}
}

func TestCheckJSONMulti(t *testing.T) {
	dir := t.TempDir()
	manifestDir := filepath.Join(dir, ".claude-plugin")
	_ = os.MkdirAll(manifestDir, 0o755)
	_ = os.WriteFile(filepath.Join(manifestDir, "plugin.json"),
		[]byte(`{"name":"x","version":"1.0.0","description":"y","author":{"name":"z"}}`), 0o644)
	_ = os.WriteFile(filepath.Join(manifestDir, "marketplace.json"),
		[]byte(`{"name":"m","owner":{"name":"o"},"plugins":[{"name":"p","source":"./p"}]}`), 0o644)
	stdout, _, code := runCLI("check", "--json", dir)
	if code != 0 {
		t.Errorf("multi-report JSON should still succeed if clean, got %d\n%s", code, stdout)
	}
	if !strings.HasPrefix(strings.TrimSpace(stdout), "[") {
		t.Errorf("multi-report JSON should be a top-level array, got: %s", stdout)
	}
	// Critical: clean reports must emit `findings: []`, never `findings: null`.
	// A nil-valued field would crash the GitHub Action's Python summary,
	// which was a real bug surfaced in Round 3 review.
	var got []report.Report
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("JSON should be parseable: %v\nstdout: %s", err, stdout)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 reports, got %d", len(got))
	}
	for i, r := range got {
		if r.Findings == nil {
			t.Errorf("report[%d].Findings is nil, must be an empty slice so JSON is []", i)
		}
	}
	// Also verify the raw output string (in case Unmarshal silently tolerates null).
	if strings.Contains(stdout, `"findings": null`) {
		t.Errorf("JSON contained literal \"findings\": null; should be []. output:\n%s", stdout)
	}
}

// TestCheckJSONCleanSingle is the tight regression guard for the Round 3
// blocker: a clean single-file --json run must emit `findings: []`, never
// `findings: null`, or the bundled GitHub Action's Python summary crashes
// on the happy path.
func TestCheckJSONCleanSingle(t *testing.T) {
	file := testdataFile(t, "valid/grafana.json")
	stdout, _, code := runCLI("check", "--kind=plugin", "--json", file)
	if code != 0 {
		t.Errorf("clean fixture should exit 0, got %d\nstdout: %s", code, stdout)
	}
	if strings.Contains(stdout, `"findings": null`) {
		t.Errorf("clean report must emit `\"findings\": []`, got `null`. output:\n%s", stdout)
	}
	var got []report.Report
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("clean JSON should parse: %v\n%s", err, stdout)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 report, got %d", len(got))
	}
	if got[0].Findings == nil {
		t.Errorf("clean report findings must be an empty slice, not nil")
	}
}

func TestCheckNonexistentPath(t *testing.T) {
	_, _, code := runCLI("check", "/does/not/exist/plugin.json")
	if code == 0 {
		t.Error("nonexistent path should fail")
	}
}
