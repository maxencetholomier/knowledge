package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/maxencetholomier/knowledge/pkg/anki"

	"github.com/spf13/cobra"
)

var ankiCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean notes in Anki that are not present locally",
	Long: `Remove notes from Anki that no longer have a local note listed in the anki_export_* deck files.

Anki must be closed: the collection file is opened directly and stays locked while Anki runs.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		deckFiles, err := getDeckFiles(DirZet)
		if err != nil {
			return err
		}

		deckFiles, err = filterDeckFiles(deckFiles, ankiDecks)
		if err != nil {
			return err
		}

		specs, err := deckSpecsFromFiles(deckFiles)
		if err != nil {
			return err
		}

		collectionPath, err := anki.FindCollection()
		if err != nil {
			return err
		}

		return pruneAnkiDecks(collectionPath, specs)
	},
}

func deckSpecsFromFiles(deckFiles []deckFile) ([]anki.DeckSpec, error) {
	specs := make([]anki.DeckSpec, 0, len(deckFiles))

	for _, deck := range deckFiles {
		noteFiles, err := readNoteList(deck.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to read deck file '%s': %w", deck.Name, err)
		}

		timestamps := make([]string, 0, len(noteFiles))
		for _, noteFile := range noteFiles {
			if _, err := os.Stat(filepath.Join(DirZet, noteFile)); err != nil {
				continue
			}
			timestamps = append(timestamps, strings.TrimSuffix(noteFile, ".md"))
		}

		specs = append(specs, anki.DeckSpec{Name: deck.Name, Timestamps: timestamps})
	}

	return specs, nil
}

func init() {
	ankiCmd.AddCommand(ankiCleanCmd)
	ankiCleanCmd.Flags().StringSliceVar(&ankiDecks, "deck", nil, "clean only the given deck(s) (repeatable)")
	ankiCleanCmd.Flags().BoolVarP(&ankiAssumeYes, "yes", "y", false, "remove orphan notes without confirmation")
	ankiCleanCmd.RegisterFlagCompletionFunc("deck", completeDeckNames)
}
