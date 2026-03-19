package cmd

import (
	"fmt"
	"kl/pkg/config"
	"kl/pkg/files"
	"kl/pkg/joplin"
	"kl/pkg/prompt"
	"kl/pkg/utils"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var joplinImportNotebook string

type noteToImport struct {
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

		notebookName, notebookId, err := getNotebookInfo()
		if err != nil {
			return err
		}

		notesToImport, err := collectNotesToImport(notebookId)
		if err != nil {
			return err
		}

		confirmed, err := confirmImport(notesToImport, notebookName)
		if err != nil {
			return err
		}
		if !confirmed {
			return nil
		}

		writeNotesToFiles(notesToImport)
		return nil
	},
}

func getNotebookInfo() (string, string, error) {
	notebookName := joplinImportNotebook
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

func writeNotesToFiles(notesToImport []noteToImport) {
	fmt.Printf("\nImporting %d notes...\n", len(notesToImport))

	for _, note := range notesToImport {
		timestamp := strings.TrimSuffix(note.fileName, ".md")

		new_id, err := joplin.ReplaceTimestampToIds(timestamp)
		if err != nil {
			fmt.Printf("Error processing %s: %v\n", timestamp, err)
			continue
		}

		var cleanBody string
		if note.title != "" {
			cleanBody = "# " + note.title + "\n\n" + note.body
		} else {
			cleanBody = "# \n\n" + note.body
		}

		err = joplin.GetResourcesFromBody(cleanBody, timestamp, DirZet)
		if err != nil {
			fmt.Printf("Error getting resources for %s: %v\n", timestamp, err)
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

func confirmImport(notesToImport []noteToImport, notebookName string) (bool, error) {
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

func collectNotesToImport(notebookId string) ([]noteToImport, error) {
	ids, err := joplin.GetIds("notes")
	if err != nil {
		return nil, err
	}

	var notesToImport []noteToImport

	for _, id := range ids {
		if notebookId != "" {
			parentId, err := joplin.GetNoteParentId(id)
			if err != nil {
				continue
			}
			if parentId != notebookId {
				continue
			}
		}

		title, err := joplin.GetField(id, "title")
		if err != nil {
			continue
		}

		body, err := joplin.GetField(id, "body")
		if err != nil {
			continue
		}

		fileName := joplin.DecryptFilename(id)
		if fileName == "" {
			timestamp := utils.CreateTimestamp()
			fileName = timestamp + ".md"
		}

		if _, err := os.Stat(DirZet + "/" + fileName); os.IsNotExist(err) {
			notesToImport = append(notesToImport, noteToImport{
				id:       id,
				title:    title,
				body:     body,
				fileName: fileName,
			})
		}
	}

	return notesToImport, nil
}

func init() {
	joplinImportCmd.Flags().StringVarP(&joplinImportNotebook, "notebook", "n", "", "specify the notebook to import notes from")
	joplinCmd.AddCommand(joplinImportCmd)
}
