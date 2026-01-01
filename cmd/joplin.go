package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var joplinCmd = &cobra.Command{
	Use:   "joplin",
	Short: "Joplin integration commands",
	Long:  `Parent command for Joplin integration subcommands including import, export, and merge operations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("subcommands missing")
	},
}

func init() {
	rootCmd.AddCommand(joplinCmd)
}
