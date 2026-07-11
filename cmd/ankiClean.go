package cmd

import (
	"fmt"
	"github.com/maxencetholomier/knowledge/pkg/prompt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var ankiCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean deck files: drop stale entries, blank lines and duplicates, sort entries",
	Long: `Clean anki_export_* deck files:
  - remove lines that reference notes not present locally
  - remove blank lines and duplicate entries
  - trim surrounding whitespace
  - sort lines in reverse order`,
	RunE: func(cmd *cobra.Command, args []string) error {
		deckFiles, err := getDeckFiles(DirZet)
		if err != nil {
			return err
		}

		localNotes, err := getLocalList()
		if err != nil {
			return err
		}

		var staleEntries []string
		for _, deck := range deckFiles {
			_, removed, _, err := cleanDeckLines(deck.Path, localNotes)
			if err != nil {
				return err
			}
			for _, entry := range removed {
				staleEntries = append(staleEntries, fmt.Sprintf("%s (deck: %s)", entry, deck.Name))
			}
		}

		if len(staleEntries) > 0 {
			fmt.Printf("Found %d deck entries without local note:\n", len(staleEntries))
			for _, entry := range staleEntries {
				fmt.Printf("  • %s\n", entry)
			}
			confirmed, err := prompt.Confirm("Do you want to remove these entries from the deck files?")
			if err != nil {
				return err
			}
			if !confirmed {
				fmt.Println("Operation cancelled.")
				return nil
			}
		}

		totalRemoved := 0
		updatedDecks := 0
		for _, deck := range deckFiles {
			kept, removed, changed, err := cleanDeckLines(deck.Path, localNotes)
			if err != nil {
				return err
			}
			if !changed {
				continue
			}
			if err := os.WriteFile(deck.Path, []byte(deckContent(kept)), 0644); err != nil {
				return fmt.Errorf("failed to write deck file '%s': %w", deck.Name, err)
			}
			fmt.Printf("✓ Cleaned deck '%s' (%d stale entries removed)\n", deck.Name, len(removed))
			totalRemoved += len(removed)
			updatedDecks++
		}

		if updatedDecks == 0 {
			fmt.Println("All Anki deck files are already clean.")
			return nil
		}

		fmt.Printf("\nCleaning completed. Updated %d deck file(s), removed %d stale entries.\n", updatedDecks, totalRemoved)
		return nil
	},
}

func cleanDeckLines(deckPath string, localNotes map[string]string) (kept []string, removed []string, changed bool, err error) {
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

func init() {
	ankiCmd.AddCommand(ankiCleanCmd)
}
