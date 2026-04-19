package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestSeverityString(t *testing.T) {
	if got := SeverityError.String(); got != "error" {
		t.Errorf("SeverityError.String() = %q, want error", got)
	}
	if got := SeverityWarning.String(); got != "warning" {
		t.Errorf("SeverityWarning.String() = %q, want warning", got)
	}
}

func TestReportBookkeeping(t *testing.T) {
	r := &Report{}
	if r.HasErrors() {
		t.Error("empty report should not have errors")
	}
	if r.HasFindings() {
		t.Error("empty report should not have findings")
	}
	r.Add(Finding{Severity: SeverityWarning, Rule: "W1"})
	if r.HasErrors() {
		t.Error("warning-only report should not have errors")
	}
	if !r.HasFindings() {
		t.Error("report with a warning should have findings")
	}
	r.Add(Finding{Severity: SeverityError, Rule: "E1"})
	if !r.HasErrors() {
		t.Error("report with an error should have errors")
	}
	errs, warns := r.Counts()
	if errs != 1 || warns != 1 {
		t.Errorf("Counts() = (%d, %d), want (1, 1)", errs, warns)
	}
}

func TestWritePrettyClean(t *testing.T) {
	r := &Report{File: "plugin.json", Kind: "plugin"}
	var buf bytes.Buffer
	r.WritePretty(&buf, false)
	out := buf.String()
	if !strings.Contains(out, "clean") {
		t.Errorf("clean pretty output should include \"clean\", got %q", out)
	}
}

func TestWritePrettyWithFindings(t *testing.T) {
	r := &Report{File: "plugin.json", Kind: "plugin"}
	r.Add(Finding{
		Severity: SeverityError,
		Rule:     "HANKO001",
		Path:     "/hooks",
		Message:  "duplicate hooks declaration",
		Fix:      "remove the field",
		DocURL:   "https://example.com",
	})
	var buf bytes.Buffer
	r.WritePretty(&buf, false)
	out := buf.String()
	for _, want := range []string{"HANKO001", "/hooks", "duplicate hooks declaration", "remove the field", "https://example.com"} {
		if !strings.Contains(out, want) {
			t.Errorf("pretty output missing %q. full output:\n%s", want, out)
		}
	}
}

func TestWritePrettyColorCleanContainsANSI(t *testing.T) {
	r := &Report{File: "plugin.json", Kind: "plugin"}
	var buf bytes.Buffer
	r.WritePretty(&buf, true) // color=true
	out := buf.String()
	// Clean reports get the green check. Verify at least one ANSI escape
	// is present so a regression that drops color silently is caught.
	if !strings.Contains(out, "\x1b[32m") {
		t.Errorf("color=true clean output should contain green ANSI \\x1b[32m, got: %q", out)
	}
}

func TestWritePrettyColorFindingsContainANSI(t *testing.T) {
	r := &Report{File: "plugin.json", Kind: "plugin"}
	r.Add(Finding{Severity: SeverityError, Rule: "HANKO001", Message: "x"})
	r.Add(Finding{Severity: SeverityWarning, Rule: "HANKO003", Message: "y"})
	var buf bytes.Buffer
	r.WritePretty(&buf, true) // color=true
	out := buf.String()
	if !strings.Contains(out, "\x1b[31m") {
		t.Errorf("color=true error line should contain red ANSI \\x1b[31m, got: %q", out)
	}
	if !strings.Contains(out, "\x1b[33m") {
		t.Errorf("color=true warning line should contain yellow ANSI \\x1b[33m, got: %q", out)
	}
}

func TestWriteJSONRoundTrip(t *testing.T) {
	r := &Report{
		File: "plugin.json",
		Kind: "plugin",
		Findings: []Finding{
			{Severity: SeverityError, Rule: "HANKO001", Path: "/hooks", Message: "x"},
			{Severity: SeverityWarning, Rule: "HANKO003", Message: "y"},
		},
	}
	var buf bytes.Buffer
	if err := r.WriteJSON(&buf); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	var back Report
	if err := json.Unmarshal(buf.Bytes(), &back); err != nil {
		t.Fatalf("Unmarshal round-trip: %v", err)
	}
	if back.File != r.File || back.Kind != r.Kind || len(back.Findings) != 2 {
		t.Errorf("round-trip lost data: got %+v", back)
	}
	if back.Findings[0].Severity != SeverityError || back.Findings[1].Severity != SeverityWarning {
		t.Errorf("severities did not round-trip: got %v", back.Findings)
	}
}

// TestSeverityMarshalsAsString is the guard that keeps the GitHub Action
// contract intact. The action's Python consumer compares
// `f["severity"] == "error"`, so int marshaling (the zero-cost default
// for an int-typed Severity) would break every consumer silently.
func TestSeverityMarshalsAsString(t *testing.T) {
	f := Finding{Severity: SeverityError, Rule: "HANKO001"}
	data, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !strings.Contains(string(data), `"severity":"error"`) {
		t.Errorf("severity should marshal as string, got: %s", data)
	}

	f2 := Finding{Severity: SeverityWarning}
	data2, _ := json.Marshal(f2)
	if !strings.Contains(string(data2), `"severity":"warning"`) {
		t.Errorf("warning severity should marshal as \"warning\", got: %s", data2)
	}
}

func TestSeverityUnmarshalRejectsUnknown(t *testing.T) {
	var s Severity
	if err := json.Unmarshal([]byte(`"nonsense"`), &s); err == nil {
		t.Error("Unmarshal should reject unknown severity strings")
	}
	// Valid values must still parse.
	if err := json.Unmarshal([]byte(`"error"`), &s); err != nil {
		t.Errorf("unmarshal \"error\" should succeed, got: %v", err)
	}
	if s != SeverityError {
		t.Errorf("unmarshal \"error\" → %v, want %v", s, SeverityError)
	}
}
