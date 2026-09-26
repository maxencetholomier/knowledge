package cmd

import (
	"fmt"
	"github.com/maxencetholomier/knowledge/pkg/files"
	"github.com/maxencetholomier/knowledge/pkg/prompt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

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

		deletedCount := deleteFiles(emptyNotes, unlinkedImages)
		fmt.Printf("Successfully deleted %d files.\n", deletedCount)

		return applyDeckCleanups(deckCleanups)
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
		kept, removed, changed, err := cleanDeckLines(deck.Path, localNotes)
		if err != nil {
			return nil, err
		}
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
		if err := os.WriteFile(cleanup.Deck.Path, []byte(deckContent(cleanup.Kept)), 0644); err != nil {
			return fmt.Errorf("failed to write deck file '%s': %w", cleanup.Deck.Name, err)
		}
		fmt.Printf("✓ Cleaned deck '%s' (%d stale entries removed)\n", cleanup.Deck.Name, len(cleanup.Removed))
		totalRemoved += len(cleanup.Removed)
	}

	fmt.Printf("Updated %d deck file(s), removed %d stale entries.\n", len(cleanups), totalRemoved)
	return nil
}

func cleanDeckLines(deckPath string, localNotes map[string]string) (kept, removed []string, changed bool, err error) {
	data, err := os.ReadFile(deckPath)
	if err != nil {
		return nil, nil, false, fmt.Errorf("failed to read deck file: %w", err)
	}

	seen := make(map[string]bool)
	for line := range strings.SplitSeq(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		entry := trimmed
		if idx := strings.Index(entry, " #"); idx != -1 {
			entry = strings.TrimSpace(entry[:idx])
		}
		key := trimmed
		if entry != "" && !strings.HasPrefix(entry, "#") {
			key = entry
			timestamp := strings.TrimSuffix(entry, ".md")
			if _, exists := localNotes[timestamp]; !exists {
				removed = append(removed, entry)
				continue
			}
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		kept = append(kept, trimmed)
	}

	sort.Sort(sort.Reverse(sort.StringSlice(kept)))
	changed = deckContent(kept) != string(data)
	return kept, removed, changed, nil
}

func deckContent(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
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

func deleteFiles(emptyNotes, unlinkedImages []string) int {
	deletedCount := 0

	for _, note := range emptyNotes {
		if err := os.Remove(filepath.Join(DirZet, note)); err != nil {
			fmt.Printf("Error deleting %s: %v\n", note, err)
		} else {
			deletedCount++
		}
	}

	for _, image := range unlinkedImages {
		if err := os.Remove(filepath.Join(DirZet, image)); err != nil {
			fmt.Printf("Error deleting %s: %v\n", image, err)
		} else {
			deletedCount++
		}
	}

	return deletedCount
}

func findEmptyNotes() ([]string, error) {
	scanner := files.NewScanner(DirZet).WithExtensions("md")
	fileList, err := scanner.ListFiles()
	if err != nil {
		return nil, err
	}

	var emptyNotes []string
	for _, file := range fileList {
		if err := file.LoadContent(); err != nil {
			continue
		}

		lines := strings.Split(file.Content, "\n")
		hasContent := false

		for i, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			if i == 0 && strings.HasPrefix(line, "#") {
				continue
			}

			hasContent = true
			break
		}

		if !hasContent {
			emptyNotes = append(emptyNotes, file.Name)
		}
	}

	return emptyNotes, nil
}

func findUnlinkedImages() ([]string, error) {
	imageFiles, err := findImageFiles()
	if err != nil {
		return nil, err
	}

	if len(imageFiles) == 0 {
		return []string{}, nil
	}

	linkedImages, err := findLinkedImages()
	if err != nil {
		return nil, err
	}

	var unlinkedImages []string
	for _, image := range imageFiles {
		isLinked := false
		for _, linked := range linkedImages {
			if image == linked {
				isLinked = true
				break
			}
		}
		if !isLinked {
			unlinkedImages = append(unlinkedImages, image)
		}
	}

	return unlinkedImages, nil
}

func findImageFiles() ([]string, error) {
	entries, err := os.ReadDir(DirZet)
	if err != nil {
		return nil, err
	}

	var imageFiles []string
	imageExtensions := []string{".png", ".jpg", ".jpeg", ".gif", ".bmp", ".svg", ".webp"}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		for _, ext := range imageExtensions {
			if strings.HasSuffix(strings.ToLower(name), ext) {
				imageFiles = append(imageFiles, name)
				break
			}
		}
	}

	return imageFiles, nil
}

func findLinkedImages() ([]string, error) {
	scanner := files.NewScanner(DirZet).WithExtensions("md")
	fileList, err := scanner.ListFiles()
	if err != nil {
		return nil, err
	}

	linkedImagesMap := make(map[string]bool)

	markdownImagePattern := regexp.MustCompile(`!\[.*?\]\(([^)]+)\)`)
	htmlImagePattern := regexp.MustCompile(`<img[^>]+src=["']([^"']+)["']`)

	for _, file := range fileList {
		if err := file.LoadContent(); err != nil {
			continue
		}

		markdownMatches := markdownImagePattern.FindAllStringSubmatch(file.Content, -1)
		for _, match := range markdownMatches {
			if len(match) > 1 {
				imagePath := match[1]
				imageName := filepath.Base(imagePath)
				linkedImagesMap[imageName] = true
			}
		}

		htmlMatches := htmlImagePattern.FindAllStringSubmatch(file.Content, -1)
		for _, match := range htmlMatches {
			if len(match) > 1 {
				imagePath := match[1]
				imageName := filepath.Base(imagePath)
				linkedImagesMap[imageName] = true
			}
		}
	}

	var linkedImages []string
	for imageName := range linkedImagesMap {
		linkedImages = append(linkedImages, imageName)
	}

	return linkedImages, nil
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
