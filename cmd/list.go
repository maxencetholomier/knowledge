package cmd

import (
	"fmt"
	"kl/pkg/files"
	"strings"

	"github.com/spf13/cobra"
)

// TODO: Add all linked not with a specific notes
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Short:   "List all notes with timestamps and titles",
	Long:    `Display all notes in the kl knowledge base showing their timestamps and titles.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		scanner := files.NewScanner(DirZet).WithExtensions("md")
		fileList, err := scanner.ListFiles()
		if err != nil {
			return err
		}

		for _, file := range fileList {
			timestamp := strings.TrimSuffix(file.Name, ".md")

			title, err := file.GetTitle()
			if err != nil {
				fmt.Println(timestamp)
			} else {
				title = strings.TrimSuffix(title, "\n")
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
	rootCmd.AddCommand(listCmd)
}
