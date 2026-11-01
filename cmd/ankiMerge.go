package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/maxencetholomier/knowledge/internal/anki"
	"github.com/maxencetholomier/knowledge/internal/files"
	"github.com/maxencetholomier/knowledge/internal/merge"

	"github.com/spf13/cobra"
)

var ankiMergeShowDiff bool
var ankiMergeForceLocal bool
var ankiMergeForceAnki bool

type ankiMergeAction struct {
	action     merge.Direction
	deck       string
	fileName   string
	title      string
	local      string
	remote     string
	fileUpdate time.Time
	ankiUpdate time.Time
}

var ankiMergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge notes with Anki bidirectionally",
	Long: `Synchronize notes between the knowledge base and Anki based on modification timestamps.

The hidden Markdown field of each Anki note is compared with the local file. When they differ,
the newest version wins: the Anki note modification time against the last git commit of the file.
Pulled notes overwrite the local file; pushed notes re-export their whole deck. There is no
conflict resolution, the older version is lost.

Anki notes whose Markdown field is empty (exported before the field existed) are always pushed.
Anki notes without a local file are left to 'kl anki import'. Anki must be closed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if ankiMergeForceLocal && ankiMergeForceAnki {
			return fmt.Errorf("--force-local and --force-anki are mutually exclusive")
		}

		forceAction := merge.None
		if ankiMergeForceLocal {
			forceAction = merge.Push
		} else if ankiMergeForceAnki {
			forceAction = merge.Pull
		}

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

		actions, pending, err := getAnkiMergeActions(notes, forceAction)
		if err != nil {
			return err
		}

		if pending > 0 {
			fmt.Printf("%d Anki note(s) have no local note, run 'kl anki import' to create them.\n", pending)
		}

		if len(actions) == 0 {
			fmt.Println("No notes need synchronization - all notes are up to date.")
			return nil
		}

		if !confirmAnkiMerge(actions) {
			return nil
		}

		return applyAnkiMergeActions(actions, deckFiles)
	},
}

func getAnkiMergeActions(
	notes []anki.CollectionNote, forceAction merge.Direction,
) ([]ankiMergeAction, int, error) {
	fileTimestamps, err := files.GetLastUpdates(DirZet)
	if err != nil {
		return nil, 0, err
	}

	var actions []ankiMergeAction
	pending := 0
	for _, note := range notes {
		if note.Timestamp == "" {
			pending++
			continue
		}

		fileName := note.Timestamp + ".md"
		content, err := os.ReadFile(filepath.Join(DirZet, fileName))
		if err != nil {
			pending++
			continue
		}

		action := ankiMergeAction{
			deck:       note.Deck,
			fileName:   fileName,
			title:      note.Front,
			local:      strings.TrimSpace(string(content)),
			remote:     anki.DecodeMarkdownField(note.Markdown),
			ankiUpdate: time.Unix(note.Modified, 0),
		}
		if action.local == action.remote {
			continue
		}

		action.fileUpdate = fileTimestamps[fileName]

		if action.remote == "" {
			action.action = merge.Push
		} else {
			action.action = merge.Decide(
				action.local, action.remote, action.fileUpdate, action.ankiUpdate, forceAction,
			)
		}

		actions = append(actions, action)
	}

	return actions, pending, nil
}

func confirmAnkiMerge(actions []ankiMergeAction) bool {
	fmt.Printf("Will synchronize %d note(s):\n", len(actions))

	pullCount, pushCount := 0, 0
	for _, action := range actions {
		if action.action == merge.Pull {
			pullCount++
			fmt.Printf("  ← %s - %s (pull from Anki, updated %s)\n",
				action.fileName[:14], action.title, action.ankiUpdate.Format("2006-01-02 15:04"))
			if ankiMergeShowDiff {
				fmt.Printf("    Changes:\n")
				merge.PrintDiff(os.Stdout, action.local, action.remote, 5)
			}
		} else {
			pushCount++
			fmt.Printf("  → %s - %s (push to Anki, updated %s)\n",
				action.fileName[:14], action.title, action.fileUpdate.Format("2006-01-02 15:04"))
			if ankiMergeShowDiff {
				fmt.Printf("    Changes:\n")
				merge.PrintDiff(os.Stdout, action.remote, action.local, 5)
			}
		}
	}

	fmt.Printf("\nSummary: %d note(s) to pull from Anki, %d note(s) to push to Anki\n", pullCount, pushCount)

	confirmed, err := confirm("Proceed with synchronization?")
	if err != nil || !confirmed {
		fmt.Println("Synchronization cancelled.")
		return false
	}
	return true
}

func applyAnkiMergeActions(actions []ankiMergeAction, deckFiles []deckFile) error {
	decksToPush := make(map[string]bool)
	pulls, failed := 0, 0
	for _, action := range actions {
		if action.action == merge.Push {
			decksToPush[action.deck] = true
			continue
		}

		pulls++
		if err := files.Create(filepath.Join(DirZet, action.fileName), action.remote+"\n"); err != nil {
			fmt.Printf("Error pulling %s: %v\n", action.fileName, err)
			failed++
		}
	}

	if pulls > 0 {
		fmt.Printf("\nPulled %d note(s) from Anki.\n", pulls-failed)
	}

	if len(decksToPush) > 0 {
		var toExport []deckFile
		for _, deck := range deckFiles {
			if decksToPush[deck.Name] {
				toExport = append(toExport, deck)
			}
		}

		fmt.Printf("\nPushing %d deck(s) to Anki: %s\n", len(toExport), strings.Join(deckNames(toExport), ", "))
		if err := exportDecks(toExport, true); err != nil {
			return err
		}
	}

	return batchError(failed, pulls, "pulls")
}

func init() {
	ankiCmd.AddCommand(ankiMergeCmd)
	ankiMergeCmd.Flags().StringSliceVar(&ankiDecks, "deck", nil, "merge only the given deck(s) (repeatable)")
	ankiMergeCmd.Flags().BoolVar(&ankiMergeShowDiff, "diff", false, "show diff of changes for each file")
	ankiMergeCmd.Flags().BoolVar(&ankiMergeForceLocal, "force-local", false,
		"push all notes from local to Anki, ignoring timestamps")
	ankiMergeCmd.Flags().BoolVar(&ankiMergeForceAnki, "force-anki", false,
		"pull all notes from Anki to local, ignoring timestamps")
	ankiMergeCmd.RegisterFlagCompletionFunc("deck", completeDeckNames)
}
