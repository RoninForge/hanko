// Package cli builds the cobra command tree. Keeping the tree in a
// dedicated package (rather than inside cmd/hanko) makes it testable
// without spawning a binary.
package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/RoninForge/hanko/internal/version"
	"github.com/spf13/cobra"
)

// ErrValidationFailed is returned by subcommands that successfully ran
// but found rule violations. Callers (Execute, tests) distinguish this
// from cobra usage errors so the process exit code can reflect the
// difference: 1 for validation failures, 2 for bad invocation.
var ErrValidationFailed = errors.New("validation failed")

// Execute is the single entry point that cmd/hanko/main.go calls. It
// returns the process exit code rather than calling os.Exit so the
// main function is still testable.
//
// Exit codes:
//
//	0 — no findings (or only warnings, if the caller tolerates them)
//	1 — validation failed (at least one error-severity finding)
//	2 — invocation error (bad flags, missing files, internal failure)
func Execute(stdout, stderr io.Writer, args []string) int {
	root := newRoot(stdout, stderr)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		if errors.Is(err, ErrValidationFailed) {
			// Validation failures already printed their pretty report.
			// Adding a cobra-style "Error: validation failed" line here
			// would be redundant noise in the CI log.
			return 1
		}
		// Any other error means the invocation itself was wrong: bad
		// flag, missing file, inference failure. Cobra would normally
		// print this, but SilenceErrors=true on the root command
		// suppresses cobra's print — so we print it here ourselves.
		fmt.Fprintln(stderr, "Error:", err)
		return 2
	}
	return 0
}

func newRoot(stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hanko",
		Short: "Validate Claude Code plugin manifests before submission",
		Long: `hanko validates .claude-plugin/plugin.json and .claude-plugin/marketplace.json
against the schema derived from Anthropic's plugin docs, plus a catalog of
additional rules that catch the cases the official validator reports
opaquely.

The name means "personal seal" (判子) in Japanese — the traditional act of
approving an official document before submitting it.`,
		SilenceUsage: true,
		// Silence errors so ErrValidationFailed doesn't get printed
		// twice (once via the pretty report, once by cobra's default
		// error handler). Unknown-flag usage errors still print via
		// cobra's usage-help machinery because SilenceUsage only
		// affects the usage printout, not the error text.
		SilenceErrors: true,
		Version:       formatVersion(),
	}

	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	cmd.AddCommand(newCheckCmd(stdout, stderr))
	cmd.AddCommand(newSubmitCheckCmd(stdout, stderr))
	cmd.AddCommand(newSchemaCmd(stdout))
	cmd.AddCommand(newVersionCmd(stdout))

	return cmd
}

func formatVersion() string {
	v := version.Get()
	return fmt.Sprintf("%s (commit %s, built %s)", v.Version, v.Commit, v.BuildDate)
}
