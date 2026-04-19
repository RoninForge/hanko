package report

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

// WritePretty emits a human-readable report. It groups findings by file
// and severity, using color escapes only when the caller has asked for
// them (the CLI detects a TTY and decides).
func (r *Report) WritePretty(w io.Writer, color bool) {
	if !r.HasFindings() {
		if color {
			fmt.Fprintf(w, "\033[32mok\033[0m %s: clean\n", displayFile(r))
		} else {
			fmt.Fprintf(w, "ok %s: clean\n", displayFile(r))
		}
		return
	}

	fmt.Fprintf(w, "%s (%s)\n", displayFile(r), r.Kind)
	if r.Marketplace != "" {
		fmt.Fprintf(w, "  marketplace: %s\n", r.Marketplace)
	}
	fmt.Fprintln(w)

	// Sort so errors float above warnings at the top; stable within each
	// severity so original rule order is preserved per severity.
	findings := make([]Finding, len(r.Findings))
	copy(findings, r.Findings)
	sort.SliceStable(findings, func(i, j int) bool {
		return findings[i].Severity > findings[j].Severity
	})

	for _, f := range findings {
		writeFinding(w, f, color)
	}

	errs, warns := r.Counts()
	fmt.Fprintf(w, "\n%d error(s), %d warning(s)\n", errs, warns)
}

func writeFinding(w io.Writer, f Finding, color bool) {
	label := f.Severity.String()
	if color {
		switch f.Severity {
		case SeverityError:
			label = "\033[31merror\033[0m"
		case SeverityWarning:
			label = "\033[33mwarning\033[0m"
		}
	}
	if f.Path != "" {
		fmt.Fprintf(w, "  %s  [%s]  %s\n", label, f.Rule, f.Path)
	} else {
		fmt.Fprintf(w, "  %s  [%s]\n", label, f.Rule)
	}
	fmt.Fprintf(w, "    %s\n", f.Message)
	if f.Fix != "" {
		fmt.Fprintf(w, "    fix:  %s\n", f.Fix)
	}
	if f.DocURL != "" {
		fmt.Fprintf(w, "    docs: %s\n", f.DocURL)
	}
	fmt.Fprintln(w)
}

// WriteJSON emits a stable JSON representation suitable for CI. One
// object per line is NOT used — a single object is easier to parse with
// `jq` and aligns with tools like eslint/--format json.
func (r *Report) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

func displayFile(r *Report) string {
	if r.File != "" {
		return r.File
	}
	return "<stdin>"
}
