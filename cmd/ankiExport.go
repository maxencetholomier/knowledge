package cmd

import (
	"bufio"
	"fmt"
	"kl/pkg/anki"
	"kl/pkg/files"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

type deckFile struct {
	Name string
	Path string
}

type deckStats struct {
	NotesProcessed int
	NotesExported  int
	NotesSkipped   int
	ImagesAdded    int
}

type deckExportResult struct {
	Stats      deckStats
	OutputPath string
}

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

Lines starting with # are treated as comments and ignored. Inline comments are also supported.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		deckFiles, err := getDeckFiles(DirZet)
		if err != nil {
			return err
		}

		printDeckFilesToExport(deckFiles)

		if err := files.EnsureDir(DirExport); err != nil {
			return err
		}

		deckResults, err := exportDeckFiles(deckFiles)
		if err != nil {
			return err
		}

		printAnkiExportSummary(deckResults)

		return nil
	},
}

func printDeckFilesToExport(deckFiles []deckFile) {
	names := make([]string, len(deckFiles))
	for i, d := range deckFiles {
		names[i] = d.Name
	}
	fmt.Printf("Discovered %d deck(s): %s\n", len(deckFiles), strings.Join(names, ", "))
}

func exportDeckFiles(deckFiles []deckFile) (map[string]deckExportResult, error) {
	noteTitleMap, err := getLocalList()
	if err != nil {
		return nil, err
	}

	results := make(map[string]deckExportResult)
	for _, deck := range deckFiles {
		result, err := exportDeckFile(deck, noteTitleMap)
		if err != nil {
			fmt.Printf("Warning: Failed to export deck '%s': %v, skipping\n", deck.Name, err)
			continue
		}
		if result != nil {
			results[deck.Name] = *result
		}
	}
	return results, nil
}

func exportDeckFile(deck deckFile, noteTitleMap map[string]string) (*deckExportResult, error) {
	stats, outputPath, err := processDeck(deck, noteTitleMap)
	if err != nil {
		return nil, err
	}
	if stats.NotesProcessed == 0 {
		fmt.Printf("Warning: Deck '%s' has no notes, skipping\n", deck.Name)
		return nil, nil
	}
	return &deckExportResult{Stats: stats, OutputPath: outputPath}, nil
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
		if idx := strings.Index(line, " #"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}
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

func getDeckFiles(dir string) ([]deckFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var deckFiles []deckFile
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

			deckFiles = append(deckFiles, deckFile{
				Name: strings.TrimSpace(deckName),
				Path: filepath.Join(dir, name),
			})
		}
	}

	if len(deckFiles) == 0 {
		return nil, fmt.Errorf("no anki_export_* files found in %s", dir)
	}

	sort.Slice(deckFiles, func(i, j int) bool {
		return deckFiles[i].Name < deckFiles[j].Name
	})

	return deckFiles, nil
}

func processDeck(deck deckFile, noteTitleMap map[string]string) (deckStats, string, error) {
	stats := deckStats{}

	noteFiles, err := readNoteList(deck.Path)
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

	if err := pkg.CreateDeck(deck.Name); err != nil {
		return stats, "", fmt.Errorf("failed to create deck: %w", err)
	}

	fmt.Printf("Processing deck: %s (%d notes)\n", deck.Name, len(noteFiles))

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

		if err := pkg.AddNote(deck.Name, note); err != nil {
			fmt.Printf("  Warning: Failed to add note to deck: %v, skipping\n", err)
			stats.NotesSkipped++
			continue
		}

		stats.NotesExported++
	}

	outputPath := filepath.Join(DirExport, fmt.Sprintf("anki_cards_%s.apkg", deck.Name))
	err = pkg.WriteToFile(outputPath)
	if err != nil {
		return stats, "", fmt.Errorf("failed to write package: %w", err)
	}

	return stats, outputPath, nil
}

func printAnkiExportSummary(deckResults map[string]deckExportResult) {
	if len(deckResults) == 0 {
		fmt.Println("No notes to export")
		return
	}

	totalExported := 0
	totalSkipped := 0
	for _, r := range deckResults {
		totalExported += r.Stats.NotesExported
		totalSkipped += r.Stats.NotesSkipped
	}

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
