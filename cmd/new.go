package cmd

import (
	"time"

	"kl/pkg/files"
	"kl/pkg/utils"

	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:     "new",
	Aliases: []string{"n"},
	Short:   "Create a new note",
	Long:    `Create a new note with a timestamp-based filename and clean header.
Opens the newly created note in your default editor for immediate editing.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		now := time.Now()
		timestamp := now.Format("20060102150405")

		fileName := timestamp + ".md"
		template := utils.CreateTemplate(timestamp, "")

		file, err := files.Create(DirZet+"/"+fileName, template)
		if err != nil {
			return err
		}

		defer file.Close()

		files.Edit(DirZet + "/" + fileName)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
