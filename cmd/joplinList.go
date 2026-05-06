package cmd

import (
	"fmt"
	"kl/pkg/joplin"
	"strings"

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

		for _, note := range notes {
			if !strings.HasSuffix(note.ID, "aaa") {
				continue
			}

			filename := joplin.DecryptFilename(note.ID)
			if filename == "" {
				continue
			}

			timestamp := strings.Split(filename, ".")[0]
			if len(timestamp) != 14 {
				continue
			}

			title := strings.Split(note.Title, "\n")[0]
			if strings.HasPrefix(title, "#") {
				title = strings.TrimSpace(strings.TrimPrefix(title, "#"))
			}

			if title != "" {
				fmt.Printf("%s - %s\n", timestamp, title)
			} else {
				fmt.Println(timestamp)
			}
		}
		return nil
	},
}

func init() {
	joplinCmd.AddCommand(joplinListCmd)
}