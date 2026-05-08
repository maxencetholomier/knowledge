package cmd

import (
	"fmt"
	"kl/pkg/joplin"

	"github.com/spf13/cobra"
)

var joplinListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Short:   "List all notes in Joplin with timestamps and titles",
	Long:    `Display all notes in Joplin showing their timestamps and titles.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		notes, err := joplin.GetNotes([]string{"title"})
		if err != nil {
			return err
		}

		for _, note := range joplin.FilterLocalNotes(notes) {
			if note.Title != "" {
				fmt.Printf("%s - %s\n", note.Timestamp, note.Title)
			} else {
				fmt.Println(note.Timestamp)
			}
		}
		return nil
	},
}

func init() {
	joplinCmd.AddCommand(joplinListCmd)
}