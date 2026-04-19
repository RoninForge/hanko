// Package report represents the outcome of validating a plugin or
// marketplace manifest. A Report is a file-level container of Findings;
// each Finding names a single rule violation with enough context that a
// human or CI can act on it without reading the source.
package report

import (
	"encoding/json"
	"fmt"
)

// Severity classifies how seriously a Finding should be taken.
type Severity int

const (
	// SeverityWarning marks advisory issues. Strict mode can promote
	// these to errors; lenient mode always tolerates them.
	SeverityWarning Severity = iota
	// SeverityError marks issues that block submission. Always fails
	// the CLI exit code.
	SeverityError
)

// String returns a lowercase tag used in CLI output and JSON.
func (s Severity) String() string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warning"
	default:
		return "unknown"
	}
}

// MarshalJSON emits the string form. CI consumers (including the bundled
// GitHub Action) compare `severity == "error"`, so the default int
// marshaling of the underlying type would silently break every consumer.
func (s Severity) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON accepts the string form emitted by MarshalJSON so the
// report JSON can round-trip through external tooling.
func (s *Severity) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	switch str {
	case "error":
		*s = SeverityError
	case "warning":
		*s = SeverityWarning
	default:
		return fmt.Errorf("unknown severity %q", str)
	}
	return nil
}

// Finding is a single rule violation.
type Finding struct {
	// Severity is error or warning.
	Severity Severity `json:"severity"`
	// Rule is a short machine-readable identifier, e.g. "HANKO001".
	// Stable across releases so CI scripts can pin on it.
	Rule string `json:"rule"`
	// Path is a JSON Pointer into the manifest ("/name", "/plugins/0/source"),
	// or empty when the rule is file-level.
	Path string `json:"path,omitempty"`
	// Message describes what's wrong in one sentence.
	Message string `json:"message"`
	// Fix is a one-line human-readable suggestion. Optional.
	Fix string `json:"fix,omitempty"`
	// DocURL points at the authoritative docs for the rule. Optional.
	DocURL string `json:"doc_url,omitempty"`
}

// Report is a file-level aggregation of findings.
type Report struct {
	// File is the path that was validated.
	File string `json:"file"`
	// Kind identifies which schema was applied ("plugin" or "marketplace").
	Kind string `json:"kind"`
	// Marketplace is the --marketplace flag value if any marketplace-specific
	// rules were layered on top of the base schema. Empty otherwise.
	Marketplace string `json:"marketplace,omitempty"`
	// Findings is the ordered list of rule violations.
	Findings []Finding `json:"findings"`
}

// HasErrors returns true if any finding is SeverityError. This is what
// drives the CLI exit code.
func (r *Report) HasErrors() bool {
	for _, f := range r.Findings {
		if f.Severity == SeverityError {
			return true
		}
	}
	return false
}

// HasFindings returns true if the report contains at least one finding of
// any severity.
func (r *Report) HasFindings() bool {
	return len(r.Findings) > 0
}

// Add appends a finding.
func (r *Report) Add(f Finding) {
	r.Findings = append(r.Findings, f)
}

// Counts returns the number of errors and warnings.
func (r *Report) Counts() (errors, warnings int) {
	for _, f := range r.Findings {
		switch f.Severity {
		case SeverityError:
			errors++
		case SeverityWarning:
			warnings++
		}
	}
	return
}
