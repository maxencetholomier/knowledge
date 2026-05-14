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

var joplinImportNotebook string

type localNote struct {
	id          string
	title       string
	body        string
	fileName    string
	fileContent string
}

var joplinImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import notes from Joplin",
	Long:  `Import notes from Joplin application into the kl knowledge base.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		notebookName, notebookId, err := joplin.GetNotebookInfo(joplinImportNotebook)
		if err != nil {
			return err
		}

		query := joplin.NoteQuery{Fields: []string{"title", "body"}, NotebookID: notebookId}
		joplinNotes, err := joplin.GetNotes(query)
		if err != nil {
			return err
		}

		notesToImport := collectNotesToImport(joplinNotes)

		confirmed, err := confirmImport(notesToImport, notebookName)
		if err != nil {
			return err
		}
		if confirmed {
			downloadResources(notesToImport)
			writeNotesToFiles(notesToImport)
		}

		notesToDelete, err := collectLocalNotesInJoplinTrash()
		if err != nil {
			return err
		}
		confirmed, err = confirmDeleteLocal(notesToDelete)
		if err != nil {
			return err
		}
		if confirmed {
			deleteLocalNotes(notesToDelete)
		}

		return nil
	},
}

func downloadResources(notes []localNote) {
	for _, note := range notes {
		timestamp := strings.TrimSuffix(note.fileName, ".md")
		if err := joplin.DownloadLinkedResources(note.body, timestamp, DirZet); err != nil {
			if strings.Contains(err.Error(), "404") {
				fmt.Printf("Warning: Some resources not found for %s\n", timestamp)
			} else {
				fmt.Printf("Error downloading resources for %s: %v\n", timestamp, err)
			}
		}
	}
}

func writeNotesToFiles(notesToImport []localNote) {
	fmt.Printf("\nImporting %d notes...\n", len(notesToImport))

	for _, note := range notesToImport {
		file, err := files.Create(DirZet+"/"+note.fileName, note.fileContent)
		if err != nil {
			fmt.Printf("Error creating %s: %v\n", strings.TrimSuffix(note.fileName, ".md"), err)
			continue
		}
		defer file.Close()
	}

	fmt.Printf("\nSuccessfully imported %d notes from Joplin.\n", len(notesToImport))
}

func confirmImport(notesToImport []localNote, notebookName string) (bool, error) {
	if len(notesToImport) == 0 {
		fmt.Println("No new notes to import from Joplin.")
		return false, nil
	}

	if notebookName != "" {
		fmt.Printf("Will import %d new notes from notebook '%s':\n", len(notesToImport), notebookName)
	} else {
		fmt.Printf("Will import %d new notes from Joplin:\n", len(notesToImport))
	}

	for _, note := range notesToImport {
		timestamp := strings.TrimSuffix(note.fileName, ".md")
		if note.title != "" {
			fmt.Printf("  • %s - %s\n", timestamp, note.title)
		} else {
			fmt.Printf("  • %s (no title)\n", timestamp)
		}
	}

	confirmed, err := prompt.Confirm("Proceed with import?")
	if err != nil {
		return false, err
	}
	if !confirmed {
		fmt.Println("Import cancelled.")
		return false, nil
	}

	return true, nil
}

func collectNotesToImport(notes []joplin.Note) []localNote {
	var notesToImport []localNote

	for _, note := range notes {
		fileName := joplin.NoteIDToFilename(note.ID)
		if fileName == "" {
			fileName = utils.CreateTimestamp() + ".md"
		}
		timestamp := strings.TrimSuffix(fileName, ".md")

		if _, err := os.Stat(DirZet + "/" + fileName); os.IsNotExist(err) {
			fileContent, err := joplin.NoteToMarkdown(note.Title, note.Body, timestamp)
			if err != nil {
				fmt.Printf("Error processing %s: %v\n", timestamp, err)
				continue
			}
			notesToImport = append(notesToImport, localNote{
				id:          note.ID,
				title:       note.Title,
				body:        note.Body,
				fileName:    fileName,
				fileContent: fileContent,
			})
		}
	}

	return notesToImport
}

func collectLocalNotesInJoplinTrash() ([]localNote, error) {
	query := joplin.NoteQuery{Fields: []string{"title"}, OnlyDeleted: true}
	trashed, err := joplin.GetNotes(query)
	if err != nil {
		return nil, err
	}

	var result []localNote
	for _, note := range trashed {
		fileName := joplin.NoteIDToFilename(note.ID)
		if fileName == "" {
			continue
		}
		if _, err := os.Stat(DirZet + "/" + fileName); err == nil {
			result = append(result, localNote{fileName: fileName, title: note.Title})
		}
	}
	return result, nil
}

func confirmDeleteLocal(notes []localNote) (bool, error) {
	if len(notes) == 0 {
		fmt.Println("No local notes to delete.")
		return false, nil
	}

	fmt.Printf("\nFound %d local note(s) deleted in Joplin:\n", len(notes))
	for _, note := range notes {
		timestamp := strings.TrimSuffix(note.fileName, ".md")
		fmt.Printf("  • %s - %s\n", timestamp, note.title)
	}

	confirmed, err := prompt.Confirm("Delete these local notes?")
	if err != nil {
		return false, err
	}
	if !confirmed {
		fmt.Println("Local deletion cancelled.")
	}
	return confirmed, nil
}

func deleteLocalNotes(notes []localNote) {
	fmt.Printf("\nDeleting %d local notes...\n", len(notes))
	deleted := 0
	for _, note := range notes {
		if err := os.Remove(DirZet + "/" + note.fileName); err != nil {
			fmt.Printf("  ✗ Failed to delete %s: %v\n", note.fileName, err)
		} else {
			fmt.Printf("  ✓ Deleted %s - %s\n", strings.TrimSuffix(note.fileName, ".md"), note.title)
			deleted++
		}
	}
	fmt.Printf("\nDeleted %d local notes.\n", deleted)
}

func init() {
	joplinImportCmd.Flags().StringVarP(&joplinImportNotebook, "notebook", "n", "", "specify the notebook to import notes from")
	joplinCmd.AddCommand(joplinImportCmd)
}
