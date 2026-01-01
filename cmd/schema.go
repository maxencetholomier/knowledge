package cmd

import (
	"kl/pkg/files"

	"github.com/spf13/cobra"
)

var schemaUseNewTerminal bool

var schemaCmd = &cobra.Command{
	Use:     "schema",
	Aliases: []string{"s"},
	Short:   "Create a new note with diagram using Inkscape",
	RunE: func(cmd *cobra.Command, args []string) error {
		return files.CreateMediaNote(DirZet, SchemaTool, SchemaToolParams, Terminal, TerminalParams, ".svg", "schema", schemaUseNewTerminal)
	},
}

func init() {
	schemaCmd.Flags().BoolVarP(&schemaUseNewTerminal, "new-terminal", "t", false, "open in new terminal")
	rootCmd.AddCommand(schemaCmd)
}
