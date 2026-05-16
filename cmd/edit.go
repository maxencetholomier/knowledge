package cmd

import (
	"fmt"
	"kl/pkg/files"
	"kl/pkg/utils"
	"os"

	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:     "edit",
	Aliases: []string{"e"},
	Short:   "Edit a note by line number from cache or timestamp",
	RunE: func(cmd *cobra.Command, args []string) error {

		fileName, err := utils.ResolveFileName(args, DirCache)
		if err != nil {
			return err
		}
		filePath := DirZet + "/" + fileName + ".md"

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", filePath)
		}

		files.Edit(filePath)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}
