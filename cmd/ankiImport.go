package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/maxencetholomier/knowledge/internal/anki"
	"github.com/maxencetholomier/knowledge/internal/files"
	"github.com/maxencetholomier/knowledge/internal/prompt"
	"github.com/maxencetholomier/knowledge/internal/utils"

	"github.com/spf13/cobra"
)

type ankiImport struct {
	note      anki.CollectionNote
	timestamp string
	content   string
	adopt     bool
	rawHTML   bool
}

var ankiImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import notes from Anki",
	Long: `Create local notes from the Anki notes of your decks that have no local file.

Notes created by kl carry a timestamp tag and are written back to <timestamp>.md. Notes created
directly in Anki (note type "Knowledge Basic", no timestamp tag) get a new timestamp; the tag and
the GUID are set in Anki so the next export updates them instead of duplicating them.

The content comes from the hidden Markdown field. When it is empty, the Front and Back fields are
written as-is (raw HTML) for manual cleanup.

Imported notes are appended to the anki_export_<deck> file of their deck. Anki must be closed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureAnkiClosed(); err != nil {
			return err
		}

		deckFiles, err := getDeckFiles(DirZet)
		if err != nil {
			return err
		}

		deckFiles, err = filterDeckFiles(deckFiles, ankiDecks)
		if err != nil {
			return err
		}

		collectionPath, err := anki.FindCollection()
		if err != nil {
			return err
		}

		notes, err := anki.ReadNotes(collectionPath, deckNames(deckFiles))
		if err != nil {
			return err
		}

		imports := collectAnkiImports(notes)
		if len(imports) == 0 {
			fmt.Println("No new notes to import from Anki.")
			return nil
		}

		if !confirmAnkiImport(imports) {
			return nil
		}

		return applyAnkiImports(collectionPath, imports, deckFiles)
	},
}

func collectAnkiImports(notes []anki.CollectionNote) []ankiImport {
	var imports []ankiImport
	seen := make(map[string]bool)

	for _, note := range notes {
		timestamp := note.Timestamp
		adopt := timestamp == ""
		if adopt {
			timestamp = utils.CreateTimestamp()
		} else if seen[timestamp] {
			continue
		} else if _, err := os.Stat(filepath.Join(DirZet, timestamp+".md")); err == nil {
			continue
		}
		seen[timestamp] = true

		content := anki.DecodeMarkdownField(note.Markdown)
		rawHTML := content == ""
		if rawHTML {
			content = "# " + note.Front + "\n\n" + note.Back
		}

		imports = append(imports, ankiImport{
			note:      note,
			timestamp: timestamp,
			content:   content + "\n",
			adopt:     adopt,
			rawHTML:   rawHTML,
		})
	}

	return imports
}

func confirmAnkiImport(imports []ankiImport) bool {
	fmt.Printf("Will import %d note(s) from Anki:\n", len(imports))
	for _, imp := range imports {
		var notes []string
		if imp.adopt {
			notes = append(notes, "new timestamp")
		}
		if imp.rawHTML {
			notes = append(notes, "raw HTML, Markdown field empty")
		}
		suffix := ""
		if len(notes) > 0 {
			suffix = " [" + strings.Join(notes, ", ") + "]"
		}
		fmt.Printf("  • %s - %s (deck: %s)%s\n", imp.timestamp, imp.note.Front, imp.note.Deck, suffix)
	}

	if ankiAssumeYes {
		return true
	}

	confirmed, err := prompt.Confirm("Proceed with import?")
	if err != nil || !confirmed {
		fmt.Println("Import cancelled.")
		return false
	}
	return true
}

func applyAnkiImports(collectionPath string, imports []ankiImport, deckFiles []deckFile) error {
	deckPaths := make(map[string]string, len(deckFiles))
	for _, deck := range deckFiles {
		deckPaths[deck.Name] = deck.Path
	}

	var adoptions []anki.Adoption
	failed := 0
	for _, imp := range imports {
		fileName := imp.timestamp + ".md"
		if err := files.Create(filepath.Join(DirZet, fileName), imp.content); err != nil {
			fmt.Printf("Error creating %s: %v\n", fileName, err)
			failed++
			continue
		}

		if err := appendToDeckFile(deckPaths[imp.note.Deck], fileName, imp.note.Front); err != nil {
			fmt.Printf("Error updating deck file for %s: %v\n", fileName, err)
			failed++
		}

		if imp.adopt {
			adoptions = append(adoptions, anki.Adoption{ID: imp.note.ID, Timestamp: imp.timestamp})
		}
	}

	fmt.Printf("\nImported %d note(s) from Anki.\n", len(imports)-failed)

	if len(adoptions) > 0 {
		if err := anki.AdoptNotes(collectionPath, adoptions); err != nil {
			return fmt.Errorf("failed to tag imported notes in Anki: %w", err)
		}
		fmt.Printf("- Tagged in Anki: %d note(s)\n", len(adoptions))
	}

	return batchError(failed, len(imports), "note imports")
}

func appendToDeckFile(deckPath, fileName, title string) error {
	existing, err := os.ReadFile(deckPath)
	if err != nil {
		return err
	}

	line := fileName + " # " + title + "\n"
	if len(existing) > 0 && !strings.HasSuffix(string(existing), "\n") {
		line = "\n" + line
	}

	file, err := os.OpenFile(deckPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(line)
	return err
}

func init() {
	ankiCmd.AddCommand(ankiImportCmd)
	ankiImportCmd.Flags().StringSliceVar(&ankiDecks, "deck", nil, "import only from the given deck(s) (repeatable)")
	ankiImportCmd.Flags().BoolVarP(&ankiAssumeYes, "yes", "y", false, "import without asking for confirmation")
	ankiImportCmd.RegisterFlagCompletionFunc("deck", completeDeckNames)
}
