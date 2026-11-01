package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/maxencetholomier/knowledge/internal/anki"
	"github.com/maxencetholomier/knowledge/internal/files"
	"github.com/maxencetholomier/knowledge/internal/prompt"

	"github.com/spf13/cobra"
)

type deckCleanup struct {
	Deck    deckFile
	Kept    []string
	Removed []string
}

var cleanCmd = &cobra.Command{
	Use:     "clean",
	Aliases: []string{"c"},
	Short:   "Remove empty notes, unlinked images and stale deck entries",
	Long: `Remove empty notes (no content beyond title) and image files that are not referenced by any notes.

Also clean the anki_export_* deck files:
  - remove lines that reference notes not present locally
  - remove blank lines and duplicate entries
  - trim surrounding whitespace
  - sort lines in reverse order`,
	RunE: func(cmd *cobra.Command, args []string) error {
		emptyNotes, err := findEmptyNotes()
		if err != nil {
			return fmt.Errorf("error finding empty notes: %w", err)
		}

		unlinkedImages, err := findUnlinkedImages()
		if err != nil {
			return fmt.Errorf("error finding unlinked images: %w", err)
		}

		deckCleanups, err := findDeckCleanups(emptyNotes)
		if err != nil {
			return fmt.Errorf("error reading deck files: %w", err)
		}

		if len(emptyNotes)+len(unlinkedImages)+len(deckCleanups) == 0 {
			fmt.Println("No empty notes, unlinked images or stale deck entries found.")
			return nil
		}

		displayFilesToClean(emptyNotes, unlinkedImages, deckCleanups)

		confirmed, err := prompt.Confirm("Do you want to apply this cleanup?")
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Println("Cleanup cancelled.")
			return nil
		}

		deleteErr := deleteFiles(emptyNotes, unlinkedImages)
		return errors.Join(deleteErr, applyDeckCleanups(deckCleanups))
	},
}

func findDeckCleanups(pendingDeletions []string) ([]deckCleanup, error) {
	deckFiles, err := getDeckFiles(DirZet)
	if err != nil {
		return nil, nil
	}

	localNotes, err := getLocalList()
	if err != nil {
		return nil, err
	}

	for _, note := range pendingDeletions {
		delete(localNotes, strings.TrimSuffix(note, ".md"))
	}

	var cleanups []deckCleanup
	for _, deck := range deckFiles {
		data, err := os.ReadFile(deck.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to read deck file: %w", err)
		}
		kept, removed, changed := anki.CleanDeckLines(string(data), localNotes)
		if !changed {
			continue
		}
		cleanups = append(cleanups, deckCleanup{Deck: deck, Kept: kept, Removed: removed})
	}

	return cleanups, nil
}

func applyDeckCleanups(cleanups []deckCleanup) error {
	if len(cleanups) == 0 {
		return nil
	}

	totalRemoved := 0
	for _, cleanup := range cleanups {
		if err := os.WriteFile(cleanup.Deck.Path, []byte(anki.DeckContent(cleanup.Kept)), 0644); err != nil {
			return fmt.Errorf("failed to write deck file '%s': %w", cleanup.Deck.Name, err)
		}
		fmt.Printf("✓ Cleaned deck '%s' (%d stale entries removed)\n", cleanup.Deck.Name, len(cleanup.Removed))
		totalRemoved += len(cleanup.Removed)
	}

	fmt.Printf("Updated %d deck file(s), removed %d stale entries.\n", len(cleanups), totalRemoved)
	return nil
}

func displayFilesToClean(emptyNotes, unlinkedImages []string, cleanups []deckCleanup) {
	fmt.Printf("Found %d empty notes and %d unlinked images:\n", len(emptyNotes), len(unlinkedImages))

	for _, note := range emptyNotes {
		fmt.Printf("  Empty note: %s\n", note)
	}

	for _, image := range unlinkedImages {
		fmt.Printf("  Unlinked image: %s\n", image)
	}

	for _, cleanup := range cleanups {
		if len(cleanup.Removed) == 0 {
			fmt.Printf("  Deck to normalize: %s\n", cleanup.Deck.Name)
			continue
		}
		for _, entry := range cleanup.Removed {
			fmt.Printf("  Stale deck entry: %s (deck: %s)\n", entry, cleanup.Deck.Name)
		}
	}
}

func deleteFiles(emptyNotes, unlinkedImages []string) error {
	toDelete := append(emptyNotes, unlinkedImages...)
	failed := 0
	for _, name := range toDelete {
		if err := os.Remove(filepath.Join(DirZet, name)); err != nil {
			fmt.Printf("Error deleting %s: %v\n", name, err)
			failed++
		}
	}
	fmt.Printf("Deleted %d files.\n", len(toDelete)-failed)
	return batchError(failed, len(toDelete), "deletions")
}

func findEmptyNotes() ([]string, error) {
	fileList, err := files.NewScanner(DirZet).WithExtensions("md").ListFiles()
	if err != nil {
		return nil, err
	}

	var emptyNotes []string
	for _, file := range fileList {
		if err := file.LoadContent(); err == nil && files.IsEmptyNote(file.Content) {
			emptyNotes = append(emptyNotes, file.Name)
		}
	}
	return emptyNotes, nil
}

func findUnlinkedImages() ([]string, error) {
	entries, err := os.ReadDir(DirZet)
	if err != nil {
		return nil, err
	}

	fileList, err := files.NewScanner(DirZet).WithExtensions("md").ListFiles()
	if err != nil {
		return nil, err
	}
	var contents []string
	for _, file := range fileList {
		if err := file.LoadContent(); err == nil {
			contents = append(contents, file.Content)
		}
	}
	linked := files.LinkedImages(contents...)

	var unlinked []string
	for _, entry := range entries {
		if !entry.IsDir() && files.IsImage(entry.Name()) && !linked[entry.Name()] {
			unlinked = append(unlinked, entry.Name())
		}
	}
	return unlinked, nil
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
