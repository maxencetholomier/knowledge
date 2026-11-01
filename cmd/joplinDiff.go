package cmd

import (
	"fmt"
	"strings"

	"github.com/maxencetholomier/knowledge/internal/files"
	"github.com/maxencetholomier/knowledge/internal/joplin"
	"github.com/maxencetholomier/knowledge/internal/merge"

	"github.com/spf13/cobra"
)

type diffResult struct {
	onlyLocal  []string
	onlyJoplin []string
	common     []string
}

var joplinDiffCmd = &cobra.Command{
	Use:     "diff",
	Aliases: []string{"d"},
	Short:   "Compare local notes with Joplin notes",
	Long: `Show differences between local knowledge base and Joplin notes,
including missing notes and content differences.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		localNotes, err := getLocalList()
		if err != nil {
			return err
		}

		joplinNotes, err := getJoplinList()
		if err != nil {
			return err
		}

		diff, err := collectDiff(localNotes, joplinNotes)
		if err != nil {
			return err
		}

		printDifference(diff)
		return nil
	},
}

func printDifference(diff diffResult) {
	if len(diff.onlyLocal) == 0 && len(diff.onlyJoplin) == 0 {
		fmt.Println("No differences found - local and Joplin notes are in sync!")
	} else {
		fmt.Printf("Summary:\n")
		fmt.Printf("  Local only: %d notes\n", len(diff.onlyLocal))
		fmt.Printf("  Joplin only: %d notes\n", len(diff.onlyJoplin))
		fmt.Printf("  Common notes: %d notes\n", len(diff.common))
	}
}

func collectDiff(localNotes, joplinNotes map[string]string) (diffResult, error) {
	onlyLocal := merge.MissingKeys(localNotes, joplinNotes)
	onlyJoplin := merge.MissingKeys(joplinNotes, localNotes)
	common := merge.CommonKeys(localNotes, joplinNotes)

	if len(onlyLocal) > 0 {
		fmt.Printf("Only in local (%d notes):\n", len(onlyLocal))
		for _, timestamp := range onlyLocal {
			if title, exists := localNotes[timestamp]; exists && title != "" {
				fmt.Printf("  • %s - %s\n", timestamp, title)
			} else {
				fmt.Printf("  • %s\n", timestamp)
			}
		}
		fmt.Println()
	}

	if len(onlyJoplin) > 0 {
		fmt.Printf("Only in Joplin (%d notes):\n", len(onlyJoplin))
		for _, timestamp := range onlyJoplin {
			if title, exists := joplinNotes[timestamp]; exists && title != "" {
				fmt.Printf("  • %s - %s\n", timestamp, title)
			} else {
				fmt.Printf("  • %s\n", timestamp)
			}
		}
		fmt.Println()
	}

	return diffResult{onlyLocal: onlyLocal, onlyJoplin: onlyJoplin, common: common}, nil
}

func getLocalList() (map[string]string, error) {
	scanner := files.NewScanner(DirZet).WithExtensions("md")
	fileList, err := scanner.ListFiles()
	if err != nil {
		return nil, err
	}

	notes := make(map[string]string)
	for _, file := range fileList {
		timestamp := strings.TrimSuffix(file.Name, ".md")

		titleLine, err := file.GetTitle()
		title := ""
		if err == nil {
			firstLine := strings.Split(titleLine, "\n")[0]
			title = firstLine
			if strings.HasPrefix(title, "#") {
				title = strings.TrimPrefix(title, "#")
				title = strings.TrimSpace(title)
			}
		}
		notes[timestamp] = title
	}
	return notes, nil
}

func getJoplinList() (map[string]string, error) {
	query := joplin.GetQuery{Fields: []string{"title"}}
	joplinNotes, err := joplin.GetNotes(query)
	if err != nil {
		return nil, err
	}

	notes := make(map[string]string)
	for _, note := range joplin.ToLocalNotes(joplinNotes) {
		notes[note.Timestamp] = note.Title
	}
	return notes, nil
}

func init() {
	joplinCmd.AddCommand(joplinDiffCmd)
}
