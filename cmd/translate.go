package cmd

import (
	"github.com/spf13/cobra"
)

var translateCmd = &cobra.Command{
	Use:     "translate",
	Aliases: []string{"t"},
	Short:   "Translate notes",
	Long:    `Translate notes from one language to another.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

// TODO: make a cmd to convert a timestamp or a list of timestamp into filename with and without timestampS
func init() {
	rootCmd.AddCommand(translateCmd)
}
