package cmd

import (
	"fmt"
	"kl/pkg/files"
	"kl/pkg/joplin"
	"kl/pkg/prompt"
	"kl/pkg/utils"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var joplinExportNotebook string

type localNoteToExport struct {
	timestamp string
	title     string
}

type exportError struct {
	timestamp string
	err       error
}

var joplinExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export notes to Joplin",
	Long:  `Export notes from the knowledge base to Joplin application.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		notebookName, notebookId, err := joplin.GetNotebookInfo(joplinExportNotebook)
		if err != nil {
			return err
		}

		notes, err := collectNotesToExport(notebookId)
		if err != nil {
			return err
		}

		confirmed, err := confirmExport(notes, notebookName)
		if err != nil {
			return err
		}
		if !confirmed {
			return nil
		}

		return exportNotesToJoplin(notes, notebookId)
	},
}

func exportNotesToJoplin(notes []localNoteToExport, notebookId string) error {
	var noteErrors []exportError
	var resourceErrors []exportError

	fmt.Printf("\nExporting %d notes...\n", len(notes))

	seenTimestamps := make(map[string]bool)
	for _, note := range notes {
		if seenTimestamps[note.timestamp] {
			fmt.Printf("Skipping duplicate timestamp: %s\n", note.timestamp)
			continue
		}
		seenTimestamps[note.timestamp] = true

		noteErr, resourceErr := exportNote(note, notebookId)
		if noteErr != nil {
			noteErrors = append(noteErrors, exportError{note.timestamp, noteErr})
			continue
		}
		if resourceErr != nil {
			resourceErrors = append(resourceErrors, exportError{note.timestamp, resourceErr})
		}
	}

	printExportSummary(notes, noteErrors, resourceErrors)

	return nil
}

func printExportSummary(notes []localNoteToExport, noteErrors, resourceErrors []exportError) {
	fmt.Printf("\nSuccessfully exported %d notes to Joplin.\n", len(notes)-len(noteErrors))
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
}

func confirmExport(notes []localNoteToExport, notebookName string) (bool, error) {
	if len(notes) == 0 {
		fmt.Println("No notes to export - all local notes are already in Joplin.")
		return false, nil
	}

	if notebookName != "" {
		fmt.Printf("Will export %d notes to notebook '%s':\n", len(notes), notebookName)
	} else {
		fmt.Printf("Will export %d notes to Joplin default location:\n", len(notes))
	}

	for _, note := range notes {
		if note.title != "" {
			fmt.Printf("  • %s - %s\n", note.timestamp, note.title)
		} else {
			fmt.Printf("  • %s\n", note.timestamp)
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

func collectNotesToExport(notebookId string) ([]localNoteToExport, error) {
	scanner := files.NewScanner(DirZet).WithExtensions("md")
	fileList, err := scanner.ListFiles()
	if err != nil {
		return nil, err
	}

	fileTimestamps := files.GetTimestamps(fileList)

	joplinTimestamps, err := getAlreadyExportedTimestamps(notebookId)
	if err != nil {
		return nil, err
	}

	toExport, err := utils.ANotInB(fileTimestamps, joplinTimestamps)
	if err != nil {
		return nil, err
	}

	var result []localNoteToExport
	for _, timestamp := range toExport {
		result = append(result, noteToExport(timestamp))
	}
	return result, nil
}

func noteToExport(timestamp string) localNoteToExport {
	note := localNoteToExport{timestamp: timestamp}
	fileInfo := files.FileInfo{Name: timestamp + ".md", Path: DirZet + "/" + timestamp + ".md"}
	if title, err := fileInfo.GetTitle(); err == nil {
		note.title = title
	}
	return note
}

func getAlreadyExportedTimestamps(notebookId string) ([]string, error) {
	ids, err := joplin.GetNoteIDs(joplin.NoteQuery{NotebookID: notebookId})
	if err != nil {
		return nil, err
	}

	var timestamps []string
	for _, id := range filterAlreadyExportedIds(ids) {
		filename := joplin.IdToFilename(id)
		if filename != "" {
			timestamp := strings.Split(filename, ".")[0]
			if len(timestamp) == 14 {
				timestamps = append(timestamps, timestamp)
			}
		}
	}
	return timestamps, nil
}

func exportNote(note localNoteToExport, notebookId string) (noteErr, resourceErr error) {
	content, err := os.ReadFile(DirZet + "/" + note.timestamp + ".md")
	if err != nil {
		return fmt.Errorf("failed to read note file: %w", err), nil
	}

	if err = joplin.PostResourceFromBody(string(content), DirZet); err != nil {
		resourceErr = err
	}

	query := joplin.WriteQuery{Method: joplin.POST, FileName: note.timestamp + ".md", DirZet: DirZet, NotebookId: notebookId}
	if err = joplin.Send(query); err != nil {
		noteErr = err
	}

	return
}

func filterAlreadyExportedIds(ids []string) []string {
	var result []string
	for _, id := range ids {
		if strings.HasSuffix(id, "aaa") {
			result = append(result, id)
		}
	}
	return result
}

func init() {
	joplinExportCmd.Flags().StringVarP(&joplinExportNotebook, "notebook", "n", "", "specify the notebook to export notes to")
	joplinCmd.AddCommand(joplinExportCmd)
}
