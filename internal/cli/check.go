package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/RoninForge/hanko/internal/report"
	"github.com/RoninForge/hanko/internal/validator"
	"github.com/spf13/cobra"
)

// commonFlags is what check and submit-check both expose.
type commonFlags struct {
	json  bool
	color bool
}

func (f *commonFlags) register(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&f.json, "json", false, "emit the report as JSON on stdout instead of a pretty summary")
	cmd.Flags().BoolVar(&f.color, "color", ttyColor(), "colorize the pretty output (auto-detected from TTY)")
}

func newCheckCmd(stdout, _ io.Writer) *cobra.Command {
	var flags commonFlags
	var kindFlag string
	cmd := &cobra.Command{
		Use:   "check [path]",
		Short: "Validate a plugin manifest, a marketplace manifest, or a directory containing either",
		Long: `hanko check reads .claude-plugin/plugin.json and/or
.claude-plugin/marketplace.json and reports any schema violations or
rule findings.

Pass a directory (defaults to the current directory) and hanko will
auto-discover both manifest files. Pass a specific file path and hanko
will validate that file alone; pass --kind to override detection when
the filename is non-standard.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := "."
			if len(args) == 1 {
				target = args[0]
			}
			reports, err := runCheck(target, "", kindFlag)
			if err != nil {
				return err
			}
			return emit(stdout, reports, flags)
		},
	}
	flags.register(cmd)
	cmd.Flags().StringVar(&kindFlag, "kind", "", "override manifest kind detection (plugin|marketplace)")
	return cmd
}

func newSubmitCheckCmd(stdout, _ io.Writer) *cobra.Command {
	var flags commonFlags
	var marketplace string
	var kindFlag string
	cmd := &cobra.Command{
		Use:   "submit-check [path]",
		Short: "Validate with the additional rules of a specific marketplace",
		Long: `hanko submit-check layers marketplace-specific rules on top of the
base schema. Use it right before you submit to that marketplace.

Recognised marketplace names: anthropic, buildwithclaude, cc-marketplace,
claudemarketplaces. Unknown names fall back to the base rules.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if marketplace == "" {
				return errors.New("--marketplace is required; use `hanko check` for base rules only")
			}
			target := "."
			if len(args) == 1 {
				target = args[0]
			}
			reports, err := runCheck(target, marketplace, kindFlag)
			if err != nil {
				return err
			}
			return emit(stdout, reports, flags)
		},
	}
	flags.register(cmd)
	cmd.Flags().StringVar(&marketplace, "marketplace", "", "marketplace name to layer on top of the base rules (anthropic|buildwithclaude|cc-marketplace|claudemarketplaces)")
	cmd.Flags().StringVar(&kindFlag, "kind", "", "override manifest kind detection (plugin|marketplace)")
	_ = cmd.MarkFlagRequired("marketplace")
	return cmd
}

// runCheck resolves target (file or dir) into one or two validation runs
// and returns the reports. An explicit kindOverride ("plugin"/"marketplace")
// skips filename-based detection for single-file targets.
func runCheck(target, marketplace, kindOverride string) ([]*report.Report, error) {
	v, err := validator.New()
	if err != nil {
		return nil, fmt.Errorf("initialize validator: %w", err)
	}

	info, err := os.Stat(target)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", target, err)
	}

	// Single-file case.
	if !info.IsDir() {
		kind, err := resolveKind(target, kindOverride)
		if err != nil {
			return nil, err
		}
		r, err := validateFile(v, target, kind, marketplace)
		if err != nil {
			return nil, err
		}
		return []*report.Report{r}, nil
	}

	// Directory case: look for both manifests inside .claude-plugin/.
	var reports []*report.Report
	pluginPath := filepath.Join(target, ".claude-plugin", "plugin.json")
	marketplacePath := filepath.Join(target, ".claude-plugin", "marketplace.json")
	if _, err := os.Stat(pluginPath); err == nil {
		r, err := validateFile(v, pluginPath, validator.KindPlugin, marketplace)
		if err != nil {
			return nil, err
		}
		reports = append(reports, r)
	}
	if _, err := os.Stat(marketplacePath); err == nil {
		r, err := validateFile(v, marketplacePath, validator.KindMarketplace, marketplace)
		if err != nil {
			return nil, err
		}
		reports = append(reports, r)
	}
	if len(reports) == 0 {
		return nil, fmt.Errorf("no .claude-plugin/plugin.json or .claude-plugin/marketplace.json found under %s", target)
	}
	return reports, nil
}

func validateFile(v *validator.Validator, path string, kind validator.Kind, marketplace string) (*report.Report, error) {
	data, err := os.ReadFile(path) //nolint:gosec // user-supplied path, by design
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return v.Validate(data, validator.Options{
		Kind:        kind,
		Marketplace: marketplace,
		File:        path,
	})
}

func resolveKind(path, override string) (validator.Kind, error) {
	switch override {
	case "plugin":
		return validator.KindPlugin, nil
	case "marketplace":
		return validator.KindMarketplace, nil
	case "":
		// fall through to filename detection
	default:
		return "", fmt.Errorf("unknown --kind %q (expected plugin or marketplace)", override)
	}
	base := filepath.Base(path)
	switch base {
	case "plugin.json":
		return validator.KindPlugin, nil
	case "marketplace.json":
		return validator.KindMarketplace, nil
	default:
		return "", fmt.Errorf("cannot infer manifest kind from filename %q; pass --kind=plugin or --kind=marketplace", base)
	}
}

// emit writes the reports to stdout in the requested format and returns
// ErrValidationFailed if any report has errors, so Execute() can map it
// to exit code 1. (stderr writes happen in Execute itself, not here, so
// pretty output and the error line are never interleaved.)
func emit(stdout io.Writer, reports []*report.Report, flags commonFlags) error {
	anyErr := false
	if flags.json {
		// Always emit a top-level array, even for single-file runs, so
		// consumers can decode into `[]Report` unconditionally without
		// branching on the number of files found.
		_, _ = fmt.Fprintln(stdout, "[")
		for i, r := range reports {
			if err := r.WriteJSON(stdout); err != nil {
				return err
			}
			if i < len(reports)-1 {
				_, _ = fmt.Fprintln(stdout, ",")
			}
		}
		_, _ = fmt.Fprintln(stdout, "]")
	} else {
		for _, r := range reports {
			r.WritePretty(stdout, flags.color)
		}
	}

	for _, r := range reports {
		if r.HasErrors() {
			anyErr = true
		}
	}
	if anyErr {
		// Signal failure without printing a second error line on top of
		// the pretty report. Execute() maps this sentinel to exit code 1.
		return ErrValidationFailed
	}
	return nil
}

// ttyColor returns true when stdout is a terminal. We keep the check
// minimal so hanko does not grow a dependency for it.
func ttyColor() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
