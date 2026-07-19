package cmd

import (
	"bufio"
	"fmt"
	"github.com/maxencetholomier/knowledge/pkg/anki"
	"github.com/maxencetholomier/knowledge/pkg/files"
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

Lines starting with # are treated as comments and ignored. Inline comments are also supported.

After export, decks are automatically imported into your Anki collection using the
official anki Python library. A backup of the collection is created in Anki's
standard backups folder before each import. Anki must be closed during the import
(the collection is locked while the application runs); if Anki is running, the
command is skipped entirely. Use --no-import to only export the .apkg files.

Use --deck to restrict the export to specific decks (repeatable):
  kl anki export --deck vocabulary --deck grammar`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !ankiNoImport && anki.IsAnkiRunning() {
			fmt.Println("Anki is running, skipping export. Close Anki and re-run, or use --no-import to only export the .apkg files.")
			return nil
		}

		deckFiles, err := getDeckFiles(DirZet)
		if err != nil {
			return err
		}

		deckFiles, err = filterDeckFiles(deckFiles, ankiDecks)
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

		if !ankiNoImport && len(deckResults) > 0 {
			importDecksIntoAnki(deckResults)
		}

		return nil
	},
}

var ankiNoImport bool
var ankiDecks []string

func filterDeckFiles(deckFiles []deckFile, requested []string) ([]deckFile, error) {
	if len(requested) == 0 {
		return deckFiles, nil
	}

	byName := make(map[string]deckFile, len(deckFiles))
	available := make([]string, 0, len(deckFiles))
	for _, deck := range deckFiles {
		byName[deck.Name] = deck
		available = append(available, deck.Name)
	}

	var filtered []deckFile
	for _, name := range requested {
		deck, exists := byName[name]
		if !exists {
			return nil, fmt.Errorf("deck '%s' not found, available decks: %s", name, strings.Join(available, ", "))
		}
		filtered = append(filtered, deck)
	}

	return filtered, nil
}

func completeDeckNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	deckFiles, err := getDeckFiles(DirZet)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	names := make([]string, 0, len(deckFiles))
	for _, deck := range deckFiles {
		names = append(names, deck.Name)
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

func importDecksIntoAnki(deckResults map[string]deckExportResult) {
	collectionPath, err := anki.FindCollection()
	if err != nil {
		fmt.Printf("\nWarning: %v\nImport the .apkg files manually into Anki.\n", err)
		return
	}

	deckNames := make([]string, 0, len(deckResults))
	for name := range deckResults {
		deckNames = append(deckNames, name)
	}
	sort.Strings(deckNames)

	apkgPaths := make([]string, 0, len(deckNames))
	for _, name := range deckNames {
		path, err := filepath.Abs(deckResults[name].OutputPath)
		if err != nil {
			fmt.Printf("Warning: Failed to resolve path for deck '%s': %v\n", name, err)
			continue
		}
		apkgPaths = append(apkgPaths, path)
	}

	profile := filepath.Base(filepath.Dir(collectionPath))
	fmt.Printf("\nImporting into Anki profile '%s'...\n", profile)

	if err := anki.ImportPackages(collectionPath, apkgPaths); err != nil {
		fmt.Printf("Warning: Anki import failed: %v\nImport the .apkg files manually into Anki.\n", err)
		return
	}

	fmt.Printf("- Imported into Anki: %d deck(s)\n", len(apkgPaths))
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
	ankiExportCmd.Flags().BoolVar(&ankiNoImport, "no-import", false, "skip the import into Anki, only export .apkg files")
	ankiExportCmd.Flags().StringSliceVar(&ankiDecks, "deck", nil, "export only the given deck(s), matching anki_export_<name> (repeatable)")
	ankiExportCmd.RegisterFlagCompletionFunc("deck", completeDeckNames)
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
