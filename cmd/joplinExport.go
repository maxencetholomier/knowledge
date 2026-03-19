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

		notebookName, notebookId, err := getNotebookInfoForExport()
		if err != nil {
			return err
		}

		timestamps, err := collectNotesToExport()
		if err != nil {
			return err
		}

		confirmed, err := confirmExport(timestamps, notebookName)
		if err != nil {
			return err
		}
		if !confirmed {
			return nil
		}

		return exportNotesToJoplin(timestamps, notebookId)
	},
}

type exportError struct {
	timestamp string
	err       error
}

func exportNotesToJoplin(timestamps []string, notebookId string) error {
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
}

func confirmExport(timestamps []string, notebookName string) (bool, error) {
	if len(timestamps) == 0 {
		fmt.Println("No notes to export - all local notes are already in Joplin.")
		return false, nil
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
		return false, err
	}
	if !confirmed {
		fmt.Println("Export cancelled.")
		return false, nil
	}

	return true, nil
}

func collectNotesToExport() ([]string, error) {
	scanner := files.NewScanner(DirZet).WithExtensions("md")
	fileList, err := scanner.ListFiles()
	if err != nil {
		return nil, err
	}

	fileTimestamps := files.GetTimestamps(fileList)

	joplinTimestamps, err := joplin.GetTimestamps("notes")
	if err != nil {
		return nil, err
	}

	timestamps, err := utils.ANotInB(fileTimestamps, joplinTimestamps)
	if err != nil {
		return nil, err
	}

	return timestamps, nil
}

func getNotebookInfoForExport() (string, string, error) {
	notebookName := joplinExportNotebook
	if notebookName == "" {
		notebookName = config.GetJoplinNotebook()
	}

	var notebookId string
	if notebookName != "" {
		var err error
		notebookId, err = joplin.GetNotebookIdByName(notebookName)
		if err != nil {
			return "", "", err
		}
	}

	return notebookName, notebookId, nil
}

func init() {
	joplinExportCmd.Flags().StringVarP(&joplinExportNotebook, "notebook", "n", "", "specify the notebook to export notes to")
	joplinCmd.AddCommand(joplinExportCmd)
}
