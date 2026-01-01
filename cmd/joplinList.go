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
		ids, err := joplin.GetIds("notes")
		if err != nil {
			return err
		}

		for _, id := range ids {
			if !strings.HasSuffix(id, "aaa") {
				continue
			}

			filename := joplin.DecryptFilename(id)
			if filename == "" {
				continue
			}

			timestamp := strings.Split(filename, ".")[0]
			if len(timestamp) != 14 {
				continue
			}

			titleLine, err := joplin.GetField(id, "title")
			if err != nil {
				fmt.Println(timestamp)
			} else {
				firstLine := strings.Split(titleLine, "\n")[0]
				
				title := firstLine
				if strings.HasPrefix(title, "#") {
					title = strings.TrimPrefix(title, "#")
					title = strings.TrimSpace(title)
				}
				
				if title != "" {
					fmt.Printf("%s - %s\n", timestamp, title)
				} else {
					fmt.Println(timestamp)
				}
			}
		}
		return nil
	},
}

func init() {
	joplinCmd.AddCommand(joplinListCmd)
}