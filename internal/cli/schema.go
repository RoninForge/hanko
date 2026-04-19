package cli

import (
	"fmt"
	"io"

	"github.com/RoninForge/hanko/internal/schema"
	"github.com/spf13/cobra"
)

func newSchemaCmd(stdout io.Writer) *cobra.Command {
	var which string
	cmd := &cobra.Command{
		Use:   "schema",
		Short: "Print the embedded plugin or marketplace JSON schema",
	}
	printCmd := &cobra.Command{
		Use:   "print",
		Short: "Emit the schema to stdout",
		Long: `Writes the embedded JSON schema (plugin or marketplace) to stdout so
you can save it as a local reference for editor autocomplete:

    hanko schema print --file plugin > .claude-plugin/plugin.schema.json

Point the $schema field of your manifest at the downloaded file and your
editor will start flagging issues in real time.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch which {
			case "plugin":
				_, err := stdout.Write(schema.PluginSchema())
				return err
			case "marketplace":
				_, err := stdout.Write(schema.MarketplaceSchema())
				return err
			default:
				return fmt.Errorf("unknown --file %q (expected plugin or marketplace)", which)
			}
		},
	}
	printCmd.Flags().StringVar(&which, "file", "plugin", "which schema to print (plugin|marketplace)")
	cmd.AddCommand(printCmd)
	return cmd
}

func newVersionCmd(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the hanko version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(stdout, formatVersion())
			return nil
		},
	}
}
