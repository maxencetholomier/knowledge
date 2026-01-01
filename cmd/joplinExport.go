package cmd

import (
	"fmt"
	"kl/pkg/config"
	"kl/pkg/files"
	"kl/pkg/joplin"
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

		fmt.Print("\nProceed with export? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Export cancelled.")
			return nil
		}

		fmt.Printf("\nExporting %d notes...\n", len(timestamps))

		processed := make(map[string]bool)
		for _, timestamp := range timestamps {
			if processed[timestamp] {
				fmt.Printf("Skipping duplicate timestamp: %s\n", timestamp)
				continue
			}
			processed[timestamp] = true

			// Read the note content first to extract resources before conversion
			zetBody, err := os.ReadFile(DirZet + "/" + timestamp + ".md")
			if err != nil {
				return fmt.Errorf("failed to read note file %s: %w", timestamp, err)
			}

			body := string(zetBody)

			// Export resources first (before the note content is converted)
			err = joplin.PostResourceFromBody(body, DirZet)
			if err != nil {
				fmt.Printf("Warning: failed to export resources for note %s:\n%v\n", timestamp, err)
			}

			// Then export the note (which will convert timestamp links to Joplin IDs)
			if notebookId != "" {
				err = joplin.PostToJoplinWithNotebook(timestamp+".md", DirZet, notebookId)
			} else {
				err = joplin.PostToJoplin(timestamp+".md", DirZet)
			}
			if err != nil {
				fmt.Printf("ERROR: failed to export note %s:\n%v\n", timestamp, err)
				return fmt.Errorf("export failed")
			}

		}

		fmt.Printf("\nSuccessfully exported %d notes to Joplin.\n", len(timestamps))
		return nil
	},
}

func init() {
	joplinExportCmd.Flags().StringVarP(&joplinExportNotebook, "notebook", "n", "", "specify the notebook to export notes to")
	joplinCmd.AddCommand(joplinExportCmd)
}
