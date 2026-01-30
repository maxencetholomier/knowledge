package cmd

import (
	"bufio"
	"fmt"
	"kl/pkg/anki"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// TODO: For note 20251220142237.md  the carraige retrun have disapear
var ankiExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export selected notes to Anki package (.apkg)",
	Long: `Export notes to Anki package format (.apkg) for direct import into Anki.

Notes are organized into decks based on files in your zettelkasten directory.
Each file named 'anki_export_<deck_name>' (without extension) defines a deck.

For example:
  - anki_export_vocabulary → creates deck "vocabulary"
  - anki_export_grammar → creates deck "grammar"

Each deck is exported into a separate .apkg file at <export_dir>/anki_cards_<deck_name>.apkg.

Each deck file should contain a list of note filenames (one per line):
  20240101120000.md
  20240102130000.md

Lines starting with # are treated as comments and ignored.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		deckFiles, err := discoverDeckFiles(DirZet)
		if err != nil {
			return err
		}

		fmt.Printf("Discovered %d deck(s): %s\n",
			len(deckFiles),
			strings.Join(getSortedDeckNames(deckFiles), ", "))

		if _, err := os.Stat(DirExport); os.IsNotExist(err) {
			if err := os.Mkdir(DirExport, 0755); err != nil {
				return fmt.Errorf("failed to create export directory: %w", err)
			}
		}

		fmt.Println("Building note index...")
		allNoteFiles := gatherAllNoteFiles(deckFiles)
		noteTitleMap := buildNoteTitleMap(allNoteFiles)

		deckResults := make(map[string]DeckExportResult)
		totalExported := 0
		totalSkipped := 0

		for deckName, deckFilePath := range deckFiles {
			stats, outputPath, err := processDeck(deckName, deckFilePath, noteTitleMap)
			if err != nil {
				fmt.Printf("Warning: Failed to process deck '%s': %v, skipping\n", deckName, err)
				continue
			}

			if stats.NotesProcessed == 0 {
				fmt.Printf("Warning: Deck '%s' has no notes, skipping\n", deckName)
				continue
			}

			deckResults[deckName] = DeckExportResult{
				Stats:      stats,
				OutputPath: outputPath,
			}
			totalExported += stats.NotesExported
			totalSkipped += stats.NotesSkipped
		}

		if totalExported == 0 {
			fmt.Println("No notes to export")
			return nil
		}

		printExportStats(deckResults, totalExported, totalSkipped)

		return nil
	},
}

func extractNoteTitleFromFile(notePath string) (string, error) {
	content, err := os.ReadFile(notePath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			title := strings.TrimPrefix(trimmed, "#")
			return strings.TrimSpace(title), nil
		}
	}
	return "", fmt.Errorf("no title found")
}

func readNoteList(listFile string) ([]string, error) {
	file, err := os.Open(listFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var notes []string
	seen := make(map[string]bool)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") && !seen[line] {
			seen[line] = true
			notes = append(notes, line)
		}
	}

	return notes, scanner.Err()
}

func init() {
	ankiCmd.AddCommand(ankiExportCmd)
}

type DeckStats struct {
	NotesProcessed int
	NotesExported  int
	NotesSkipped   int
	ImagesAdded    int
}

func sanitizeDeckName(name string) string {
	return strings.TrimSpace(name)
}

func discoverDeckFiles(dir string) (map[string]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	deckFiles := make(map[string]string)
	prefix := "anki_export_"

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasPrefix(name, prefix) {
			deckName := strings.TrimPrefix(name, prefix)
			if deckName == "" {
				fmt.Printf("Warning: Ignoring invalid deck file '%s' (no deck name)\n", name)
				continue
			}

			sanitizedName := sanitizeDeckName(deckName)
			deckFiles[sanitizedName] = filepath.Join(dir, name)
		}
	}

	if len(deckFiles) == 0 {
		return nil, fmt.Errorf("no anki_export_* files found in %s", dir)
	}

	return deckFiles, nil
}

func getSortedDeckNames(deckFiles map[string]string) []string {
	names := make([]string, 0, len(deckFiles))
	for name := range deckFiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func gatherAllNoteFiles(deckFiles map[string]string) []string {
	noteSet := make(map[string]bool)

	for _, deckFile := range deckFiles {
		notes, err := readNoteList(deckFile)
		if err != nil {
			continue
		}
		for _, note := range notes {
			noteSet[note] = true
		}
	}

	result := make([]string, 0, len(noteSet))
	for note := range noteSet {
		result = append(result, note)
	}
	return result
}

func buildNoteTitleMap(noteFiles []string) map[string]string {
	noteTitleMap := make(map[string]string)

	for _, noteFile := range noteFiles {
		noteID := strings.TrimSuffix(noteFile, ".md")
		notePath := filepath.Join(DirZet, noteFile)

		if _, err := os.Stat(notePath); os.IsNotExist(err) {
			continue
		}

		title, err := extractNoteTitleFromFile(notePath)
		if err == nil {
			noteTitleMap[noteID] = title
		} else {
			noteTitleMap[noteID] = noteID
		}
	}

	return noteTitleMap
}

func processDeck(deckName, deckFilePath string, noteTitleMap map[string]string) (DeckStats, string, error) {
	stats := DeckStats{}

	noteFiles, err := readNoteList(deckFilePath)
	if err != nil {
		return stats, "", fmt.Errorf("failed to read note list: %w", err)
	}

	stats.NotesProcessed = len(noteFiles)

	if len(noteFiles) == 0 {
		return stats, "", nil
	}

	pkg, err := anki.CreatePackage()
	if err != nil {
		return stats, "", fmt.Errorf("failed to create package: %w", err)
	}

	if err := pkg.CreateDeck(deckName); err != nil {
		return stats, "", fmt.Errorf("failed to create deck: %w", err)
	}

	fmt.Printf("Processing deck: %s (%d notes)\n", deckName, len(noteFiles))

	for i, noteFile := range noteFiles {
		notePath := filepath.Join(DirZet, noteFile)

		if _, err := os.Stat(notePath); os.IsNotExist(err) {
			fmt.Printf("  Warning: Note file %s not found, skipping\n", noteFile)
			stats.NotesSkipped++
			continue
		}

		fmt.Printf("  Processing note %d of %d: %s\n", i+1, len(noteFiles), noteFile)

		note, mediaFiles, err := anki.ConvertNote(notePath, noteTitleMap)
		if err != nil {
			fmt.Printf("  Warning: Failed to process %s: %v, skipping\n", noteFile, err)
			stats.NotesSkipped++
			continue
		}

		for _, media := range mediaFiles {
			pkg.AddMedia(media.Filename, media.Data)
			stats.ImagesAdded++
		}

		if err := pkg.AddNote(deckName, note); err != nil {
			fmt.Printf("  Warning: Failed to add note to deck: %v, skipping\n", err)
			stats.NotesSkipped++
			continue
		}

		stats.NotesExported++
	}

	outputPath := filepath.Join(DirExport, fmt.Sprintf("anki_cards_%s.apkg", deckName))
	err = pkg.WriteToFile(outputPath)
	if err != nil {
		return stats, "", fmt.Errorf("failed to write package: %w", err)
	}

	return stats, outputPath, nil
}

type DeckExportResult struct {
	Stats      DeckStats
	OutputPath string
}

func printExportStats(deckResults map[string]DeckExportResult, totalExported, totalSkipped int) {
	fmt.Printf("\nExport complete!\n")
	fmt.Printf("- Decks exported: %d\n", len(deckResults))

	deckNames := make([]string, 0, len(deckResults))
	for name := range deckResults {
		deckNames = append(deckNames, name)
	}
	sort.Strings(deckNames)

	for _, name := range deckNames {
		result := deckResults[name]
		fmt.Printf("  - %s: %d notes → %s\n", name, result.Stats.NotesExported, result.OutputPath)
	}

	fmt.Printf("- Total notes exported: %d\n", totalExported)
	if totalSkipped > 0 {
		fmt.Printf("- Notes skipped: %d\n", totalSkipped)
	}
}
