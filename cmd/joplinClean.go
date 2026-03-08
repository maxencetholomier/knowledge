package cmd

import (
	"fmt"
	"kl/pkg/files"
	"kl/pkg/joplin"
	"kl/pkg/prompt"
	"strings"

	"github.com/spf13/cobra"
)

var joplinCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean notes in Joplin that are not present locally",
	Long:  `Remove notes from Joplin that do not have corresponding local files in the knowledge base.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		scanner := files.NewScanner(DirZet).WithExtensions("md")
		localFiles, err := scanner.ListFiles()
		if err != nil {
			return fmt.Errorf("failed to scan local files: %w", err)
		}

		localTimestamps := files.GetTimestamps(localFiles)

		joplinIds, err := joplin.GetIds("notes")
		if err != nil {
			return fmt.Errorf("failed to get Joplin note IDs: %w", err)
		}

		type noteToDelete struct {
			id    string
			title string
		}

		var notesToDelete []noteToDelete
		
		for _, joplinId := range joplinIds {
			title, err := joplin.GetField(joplinId, "title")
			if err != nil {
				title = "Unknown"
			}

			filename := joplin.DecryptFilename(joplinId)
			if filename != "" {
				noteTimestamp := strings.Split(filename, ".")[0]
				if len(noteTimestamp) == 14 {
					found := false
					for _, localTimestamp := range localTimestamps {
						if noteTimestamp == localTimestamp {
							found = true
							break
						}
					}
					if !found {
						notesToDelete = append(notesToDelete, noteToDelete{
							id:    joplinId,
							title: title,
						})
					}
				} else {
					notesToDelete = append(notesToDelete, noteToDelete{
						id:    joplinId,
						title: title,
					})
				}
			} else {
				notesToDelete = append(notesToDelete, noteToDelete{
					id:    joplinId,
					title: title,
				})
			}
		}

		if len(notesToDelete) == 0 {
			fmt.Println("No notes to clean from Joplin.")
			return nil
		}

		fmt.Printf("Found %d notes in Joplin that are not present locally:\n", len(notesToDelete))
		for _, note := range notesToDelete {
			filename := joplin.DecryptFilename(note.id)
			if filename != "" {
				timestamp := strings.Split(filename, ".")[0]
				if note.title != "" {
					fmt.Printf("  • %s - %s\n", timestamp, note.title)
				} else {
					fmt.Printf("  • %s (no title)\n", timestamp)
				}
			} else {
				if note.title != "" {
					fmt.Printf("  • %s (unrecognized ID format)\n", note.title)
				} else {
					fmt.Printf("  • (no title, unrecognized ID format)\n")
				}
			}
		}

		confirmed, err := prompt.Confirm("Do you want to delete these notes from Joplin?")
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Println("Operation cancelled.")
			return nil
		}

		fmt.Printf("\nDeleting %d notes from Joplin...\n", len(notesToDelete))
		
		deletedCount := 0
		for _, note := range notesToDelete {
			displayName := note.title
			if displayName == "" {
				displayName = "untitled note"
			}
			
			err := joplin.DeleteNoteFromJoplin(note.id)
			if err != nil {
				fmt.Printf("✗ Failed to delete \"%s\": %v\n", displayName, err)
			} else {
				fmt.Printf("✓ Deleted \"%s\"\n", displayName)
				deletedCount++
			}
		}

		fmt.Printf("\nCleaning completed. Deleted %d notes from Joplin.\n", deletedCount)
		return nil
	},
}

func init() {
	joplinCmd.AddCommand(joplinCleanCmd)
}