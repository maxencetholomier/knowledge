package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var ankiCmd = &cobra.Command{
	Use:   "anki",
	Short: "Anki integration commands",
	Long:  `Parent command for Anki integration subcommands including export operations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("subcommands missing")
	},
}

func init() {
	rootCmd.AddCommand(ankiCmd)
}
