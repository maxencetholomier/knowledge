package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var ankiDiffLocalOnly bool
var ankiDiffAnkiOnly bool

var ankiDiffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Compare local notes with Anki decks",
	Long: `Show differences between local notes and Anki decks.
By default shows both local-only and Anki-only notes.
Use --local to show only notes not in Anki, --anki to show only notes not local.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		deckFiles, err := discoverDeckFiles(DirZet)
		if err != nil {
			return err
		}

		ankiNotes := buildAnkiNoteSet(deckFiles)

		localNotes, err := getLocalList()
		if err != nil {
			return err
		}

		showLocal := !ankiDiffAnkiOnly
		showAnki := !ankiDiffLocalOnly

		if showLocal {
			var onlyLocal []string
			for timestamp := range localNotes {
				if !ankiNotes[timestamp] {
					onlyLocal = append(onlyLocal, timestamp)
				}
			}
			if len(onlyLocal) == 0 {
				fmt.Println("All local notes are in Anki.")
			} else {
				fmt.Printf("Only local (%d):\n", len(onlyLocal))
				for _, ts := range onlyLocal {
					if title := localNotes[ts]; title != "" {
						fmt.Printf("  • %s.md # %s\n", ts, title)
					} else {
						fmt.Printf("  • %s.md\n", ts)
					}
				}
			}
		}

		if showAnki {
			var onlyAnki []string
			for ts := range ankiNotes {
				if _, exists := localNotes[ts]; !exists {
					onlyAnki = append(onlyAnki, ts)
				}
			}
			if len(onlyAnki) == 0 {
				fmt.Println("All Anki notes exist locally.")
			} else {
				fmt.Printf("Only in Anki (%d):\n", len(onlyAnki))
				for _, ts := range onlyAnki {
					fmt.Printf("  • %s.md\n", ts)
				}
			}
		}

		return nil
	},
}

func buildAnkiNoteSet(deckFiles map[string]string) map[string]bool {
	set := make(map[string]bool)
	for _, deckFilePath := range deckFiles {
		notes, err := readNoteList(deckFilePath)
		if err != nil {
			continue
		}
		for _, note := range notes {
			ts := strings.TrimSuffix(note, ".md")
			set[ts] = true
		}
	}
	return set
}

func init() {
	ankiDiffCmd.Flags().BoolVarP(&ankiDiffLocalOnly, "local", "l", false, "Show only notes not in Anki")
	ankiDiffCmd.Flags().BoolVarP(&ankiDiffAnkiOnly, "anki", "a", false, "Show only Anki notes not found locally")
	ankiCmd.AddCommand(ankiDiffCmd)
}
