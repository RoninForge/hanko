// Package validator is the entry point other code should call. It loads
// the embedded JSON schema for the requested kind, runs the data through
// santhosh-tekuri/jsonschema, and then layers any Go-coded rules on top.
// The orchestration lives here so CLI, tests, and third-party importers
// all see the same validation surface.
package validator

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/RoninForge/hanko/internal/jsonpointer"
	"github.com/RoninForge/hanko/internal/report"
	"github.com/RoninForge/hanko/internal/rules"
	"github.com/RoninForge/hanko/internal/schema"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Kind identifies which embedded schema to apply.
type Kind = schema.Kind

// Re-exported so callers only need to import internal/validator.
const (
	KindPlugin      = schema.KindPlugin
	KindMarketplace = schema.KindMarketplace
)

// Options customises a single validation run.
type Options struct {
	// Kind selects plugin or marketplace schema.
	Kind Kind
	// Marketplace, if non-empty, layers that marketplace's rule overlay on
	// top of the base rule set (see rules.ForMarketplace).
	Marketplace string
	// File is a display path used only for error messages.
	File string
}

// Validator wraps a compiled JSON schema and the rule set that go with it.
// Reuse a single Validator across many files to avoid re-compiling.
type Validator struct {
	pluginSchema      *jsonschema.Schema
	marketplaceSchema *jsonschema.Schema
}

// New compiles both embedded schemas and returns a ready Validator.
func New() (*Validator, error) {
	pluginSchema, err := compile("plugin.schema.json", schema.PluginSchema())
	if err != nil {
		return nil, fmt.Errorf("compile plugin schema: %w", err)
	}
	marketplaceSchema, err := compile("marketplace.schema.json", schema.MarketplaceSchema())
	if err != nil {
		return nil, fmt.Errorf("compile marketplace schema: %w", err)
	}
	return &Validator{
		pluginSchema:      pluginSchema,
		marketplaceSchema: marketplaceSchema,
	}, nil
}

func compile(name string, raw []byte) (*jsonschema.Schema, error) {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", name, err)
	}
	c := jsonschema.NewCompiler()
	if err := c.AddResource(name, doc); err != nil {
		return nil, fmt.Errorf("add resource %s: %w", name, err)
	}
	s, err := c.Compile(name)
	if err != nil {
		return nil, fmt.Errorf("compile %s: %w", name, err)
	}
	return s, nil
}

// Validate runs the data through the selected schema, then the base rule
// set, then any marketplace overlay. The returned report is never nil.
func (v *Validator) Validate(data []byte, opts Options) (*report.Report, error) {
	// Initialise Findings to a non-nil empty slice so clean reports
	// marshal as `"findings": []`, not `"findings": null`. The GitHub
	// Action's Python summary does `r.get("findings", [])` which only
	// returns the default on a missing key — `null` still becomes
	// Python `None` and crashes the next iteration. This was a real
	// blocker found in Round 3 of review.
	r := &report.Report{
		File:        opts.File,
		Kind:        string(opts.Kind),
		Marketplace: opts.Marketplace,
		Findings:    []report.Finding{},
	}

	// Decode to a generic value so we can feed jsonschema and rules with
	// the same shape. jsonschema expects json.Number and Go primitives
	// matching the draft 2020-12 reference decoder.
	decoded, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		r.Add(report.Finding{
			Severity: report.SeverityError,
			Rule:     "HANKO000",
			Message:  "failed to parse JSON: " + err.Error(),
			Fix:      "run the file through `jq .` or your editor's JSON linter to find the syntax error",
		})
		return r, nil
	}

	// Schema validation.
	compiled := v.schemaFor(opts.Kind)
	if compiled == nil {
		return r, fmt.Errorf("unknown kind: %q", opts.Kind)
	}
	if err := compiled.Validate(decoded); err != nil {
		var verr *jsonschema.ValidationError
		if errors.As(err, &verr) {
			for _, f := range schemaFindings(verr) {
				r.Add(f)
			}
		} else {
			r.Add(report.Finding{
				Severity: report.SeverityError,
				Rule:     "HANKO000",
				Message:  "schema validation failed: " + err.Error(),
			})
		}
	}

	// Rule overlay. Only runs when the decoded value is an object, which
	// is true for any valid manifest; if the schema already rejected the
	// file as non-object, findings are still attached above.
	rs := rules.ForMarketplace(opts.Marketplace)
	var ruleSet []rules.Rule
	switch opts.Kind {
	case KindPlugin:
		ruleSet = rs.Plugin
	case KindMarketplace:
		ruleSet = rs.Marketplace
	}
	if decodedMap, ok := decoded.(map[string]any); ok {
		for _, f := range rules.Apply(ruleSet, decodedMap) {
			r.Add(f)
		}
	}

	return r, nil
}

func (v *Validator) schemaFor(k Kind) *jsonschema.Schema {
	switch k {
	case KindPlugin:
		return v.pluginSchema
	case KindMarketplace:
		return v.marketplaceSchema
	}
	return nil
}

// schemaFindings flattens a jsonschema.ValidationError tree into one
// Finding per leaf. Leaves (no child Causes) are the actual failures;
// internal nodes describe the keyword stack above them. Flattening keeps
// the CLI output scannable and JSON output easy to consume.
func schemaFindings(err *jsonschema.ValidationError) []report.Finding {
	leaves := collectLeaves(err)
	out := make([]report.Finding, 0, len(leaves))
	for _, l := range leaves {
		out = append(out, report.Finding{
			Severity: report.SeverityError,
			Rule:     "HANKO-SCHEMA",
			Path:     instancePath(l),
			Message:  leafMessage(l),
			DocURL:   "https://code.claude.com/docs/en/plugins-reference",
		})
	}
	return out
}

func collectLeaves(e *jsonschema.ValidationError) []*jsonschema.ValidationError {
	if e == nil {
		return nil
	}
	if len(e.Causes) == 0 {
		return []*jsonschema.ValidationError{e}
	}
	var out []*jsonschema.ValidationError
	for _, c := range e.Causes {
		out = append(out, collectLeaves(c)...)
	}
	return out
}

// instancePath turns a jsonschema InstanceLocation ([]string) into a JSON
// pointer string. The library stores segments without the leading slash
// and without RFC 6901 escaping, so we route through jsonpointer.Escape
// here so programmatic consumers of the --json output can address the
// offending node even when a manifest has keys like `my~server`.
func instancePath(e *jsonschema.ValidationError) string {
	if len(e.InstanceLocation) == 0 {
		return ""
	}
	var b strings.Builder
	for _, seg := range e.InstanceLocation {
		b.WriteByte('/')
		b.WriteString(jsonpointer.Escape(seg))
	}
	return b.String()
}

// leafMessage formats a leaf validation error. The library's default Error()
// string includes the schema path which is already shown elsewhere; we just
// want the kind of failure and the offending value.
func leafMessage(e *jsonschema.ValidationError) string {
	msg := e.Error()
	// Strip "jsonschema validation failed with ..." prefix when present.
	const prefix = "jsonschema validation failed with "
	if strings.HasPrefix(msg, prefix) {
		if idx := strings.IndexByte(msg, '\n'); idx > 0 {
			msg = msg[idx+1:]
		}
	}
	return msg
}
