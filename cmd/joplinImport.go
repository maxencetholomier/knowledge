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
	id       string
	title    string
	body     string
	fileName string
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

		joplinNotes, err := joplin.GetNotes([]string{"title", "body", "parent_id"})
		if err != nil {
			return err
		}

		notesToImport, err := collectNotesToImport(joplinNotes, notebookId)
		if err != nil {
			return err
		}

		confirmed, err := confirmImport(notesToImport, notebookName)
		if err != nil {
			return err
		}
		if confirmed {
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

func writeNotesToFiles(notesToImport []localNote) {
	fmt.Printf("\nImporting %d notes...\n", len(notesToImport))

	for _, note := range notesToImport {
		timestamp := strings.TrimSuffix(note.fileName, ".md")

		new_id, err := joplin.ReplaceTimestampToIds(timestamp)
		if err != nil {
			fmt.Printf("Error processing %s: %v\n", timestamp, err)
			continue
		}

		cleanBody := joplin.ReconstructBody(note.title, note.body)

		err = joplin.GetResourcesFromBody(cleanBody, timestamp, DirZet)
		if err != nil {
			if strings.Contains(err.Error(), "404") {
				fmt.Printf("Warning: Some resources not found for %s\n", timestamp)
			} else {
				fmt.Printf("Error getting resources for %s: %v\n", timestamp, err)
			}
		}

		new_body, err := joplin.ReplacingJoplinLink(cleanBody, new_id)
		if err != nil {
			fmt.Printf("Error replacing links for %s: %v\n", timestamp, err)
			continue
		}

		new_body = strings.ReplaceAll(new_body, "&nbsp;", "")

		file, err := files.Create(DirZet+"/"+note.fileName, new_body)
		if err != nil {
			fmt.Printf("Error creating %s: %v\n", timestamp, err)
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

func collectNotesToImport(notes []joplin.Note, notebookId string) ([]localNote, error) {
	var notesToImport []localNote

	for _, note := range notes {
		if notebookId != "" && note.ParentID != notebookId {
			continue
		}

		fileName := joplin.DecryptFilename(note.ID)
		if fileName == "" {
			timestamp := utils.CreateTimestamp()
			fileName = timestamp + ".md"
		}

		if _, err := os.Stat(DirZet + "/" + fileName); os.IsNotExist(err) {
			notesToImport = append(notesToImport, localNote{
				id:       note.ID,
				title:    note.Title,
				body:     note.Body,
				fileName: fileName,
			})
		}
	}

	return notesToImport, nil
}

type localNoteToDelete struct {
	fileName string
	title    string
}

func collectLocalNotesInJoplinTrash() ([]localNoteToDelete, error) {
	trashed, err := joplin.GetTrashedNotes()
	if err != nil {
		return nil, err
	}

	var result []localNoteToDelete
	for _, note := range trashed {
		fileName := joplin.DecryptFilename(note.ID)
		if fileName == "" {
			continue
		}
		if _, err := os.Stat(DirZet + "/" + fileName); err == nil {
			result = append(result, localNoteToDelete{fileName: fileName, title: note.Title})
		}
	}
	return result, nil
}

func confirmDeleteLocal(notes []localNoteToDelete) (bool, error) {
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

func deleteLocalNotes(notes []localNoteToDelete) {
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
