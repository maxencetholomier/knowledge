package cmd

import (
	"fmt"
	"kl/pkg/config"
	"kl/pkg/files"
	"kl/pkg/joplin"
	"kl/pkg/prompt"
	"kl/pkg/utils"
	"os"

	"github.com/spf13/cobra"
)

var joplinExportNotebook string

var joplinExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export notes to Joplin",
	Long:  `Export notes from the knowledge base to Joplin application.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		var notebookId string
		notebookName := joplinExportNotebook
		if notebookName == "" {
			notebookName = config.GetJoplinNotebook()
		}
		if notebookName != "" {
			var err error
			notebookId, err = joplin.GetNotebookIdByName(notebookName)
			if err != nil {
				return err
			}
		}

		scanner := files.NewScanner(DirZet).WithExtensions("md")
		fileList, err := scanner.ListFiles()
		if err != nil {
			return err
		}

		fileTimestamps := files.GetTimestamps(fileList)

		joplinTimestamps, err := joplin.GetTimestamps("notes")
		if err != nil {
			return err
		}

		timestamps, err := utils.ANotInB(fileTimestamps, joplinTimestamps)
		if err != nil {
			return err
		}

		if len(timestamps) == 0 {
			fmt.Println("No notes to export - all local notes are already in Joplin.")
			return nil
		}

		if notebookName != "" {
			fmt.Printf("Will export %d notes to notebook '%s':\n", len(timestamps), notebookName)
		} else {
			fmt.Printf("Will export %d notes to Joplin default location:\n", len(timestamps))
		}

		for _, timestamp := range timestamps {
			fileName := timestamp + ".md"
			filePath := DirZet + "/" + fileName
			
			file, err := os.Open(filePath)
			if err != nil {
				fmt.Printf("  • %s (unable to read title)\n", timestamp)
				continue
			}
			defer file.Close()

			fileInfo := files.FileInfo{Name: fileName, Path: filePath}
			title, err := fileInfo.GetTitle()
			if err != nil {
				fmt.Printf("  • %s (unable to read title)\n", timestamp)
			} else {
				if title != "" {
					fmt.Printf("  • %s - %s\n", timestamp, title)
				} else {
					fmt.Printf("  • %s\n", timestamp)
				}
			}
		}

		confirmed, err := prompt.Confirm("Proceed with export?")
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Println("Export cancelled.")
			return nil
		}

		type exportError struct {
			timestamp string
			err       error
		}
		var noteErrors []exportError
		var resourceErrors []exportError

		fmt.Printf("\nExporting %d notes...\n", len(timestamps))

		processed := make(map[string]bool)
		for _, timestamp := range timestamps {
			if processed[timestamp] {
				fmt.Printf("Skipping duplicate timestamp: %s\n", timestamp)
				continue
			}
			processed[timestamp] = true

			zetBody, err := os.ReadFile(DirZet + "/" + timestamp + ".md")
			if err != nil {
				noteErrors = append(noteErrors, exportError{timestamp, fmt.Errorf("failed to read note file: %w", err)})
				continue
			}

			body := string(zetBody)

			err = joplin.PostResourceFromBody(body, DirZet)
			if err != nil {
				resourceErrors = append(resourceErrors, exportError{timestamp, err})
			}

			if notebookId != "" {
				err = joplin.PostToJoplinWithNotebook(timestamp+".md", DirZet, notebookId)
			} else {
				err = joplin.PostToJoplin(timestamp+".md", DirZet)
			}
			if err != nil {
				noteErrors = append(noteErrors, exportError{timestamp, err})
				continue
			}

		}

		fmt.Printf("\nSuccessfully exported %d notes to Joplin.\n", len(timestamps)-len(noteErrors))
		if len(noteErrors) > 0 {
			fmt.Printf("Warning: %d note(s) could not be exported:\n", len(noteErrors))
			for _, e := range noteErrors {
				fmt.Printf("  - %s: %v\n", e.timestamp, e.err)
			}
		}
		if len(resourceErrors) > 0 {
			fmt.Printf("Warning: %d note(s) had resource export failures:\n", len(resourceErrors))
			for _, e := range resourceErrors {
				fmt.Printf("  - %s: %v\n", e.timestamp, e.err)
			}
		}
		return nil
	},
}

func init() {
	joplinExportCmd.Flags().StringVarP(&joplinExportNotebook, "notebook", "n", "", "specify the notebook to export notes to")
	joplinCmd.AddCommand(joplinExportCmd)
}
