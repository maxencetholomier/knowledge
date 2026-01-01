package cmd

import (
	"kl/pkg/utils"
	"os"

	"github.com/spf13/cobra"
)

// TODO: should be able use fzf selector
// TODO: for fzf should be able to delete multiple note
// TODO: without fzf it should be able to take a list of timestamp
var deleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"d"},
	Short:   "Delete a note by line number from cache",
	Long:    `Delete a note by specifying its line number from the last search or find result cache.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fileName, err := utils.ResolveFileName(args, DirCache)
		if err != nil {
			return err
		}

		os.Remove(DirZet + "/" + fileName + ".md")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
