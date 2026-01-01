package cmd

import (
	"fmt"
	"kl/pkg/files"
	"kl/pkg/joplin"
	"kl/pkg/utils"
	"strings"

	"github.com/spf13/cobra"
)

var debugDiff bool

var joplinDiffCmd = &cobra.Command{
	Use:     "diff",
	Aliases: []string{"d"},
	Short:   "Compare local notes with Joplin notes",
	Long:    `Show differences between local knowledge base and Joplin notes, including missing notes and content differences.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		localNotes, err := getLocalList()
		if err != nil {
			return err
		}

		joplinNotes, err := getJoplinList()
		if err != nil {
			return err
		}

		localTimestamps := make([]string, 0, len(localNotes))
		joplinTimestamps := make([]string, 0, len(joplinNotes))

		for timestamp := range localNotes {
			localTimestamps = append(localTimestamps, timestamp)
		}
		for timestamp := range joplinNotes {
			joplinTimestamps = append(joplinTimestamps, timestamp)
		}

		onlyLocal, err := utils.ANotInB(localTimestamps, joplinTimestamps)
		if err != nil {
			return err
		}

		onlyJoplin, err := utils.ANotInB(joplinTimestamps, localTimestamps)
		if err != nil {
			return err
		}

		var hasAnyDifferences bool

		if len(onlyLocal) > 0 {
			hasAnyDifferences = true
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
			hasAnyDifferences = true
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

		common := getCommonTimestamps(localTimestamps, joplinTimestamps)

		if !hasAnyDifferences {
			fmt.Println("No differences found - local and Joplin notes are in sync!")
		} else {
			fmt.Printf("Summary:\n")
			fmt.Printf("  Local only: %d notes\n", len(onlyLocal))
			fmt.Printf("  Joplin only: %d notes\n", len(onlyJoplin))
			fmt.Printf("  Common notes: %d notes\n", len(common))
		}

		return nil
	},
}

func getCommonTimestamps(local, joplin []string) []string {
	var common []string
	for _, localTs := range local {
		if utils.ItemInSlice(joplin, localTs) {
			common = append(common, localTs)
		}
	}
	return common
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
	ids, err := joplin.GetIds("notes")
	if err != nil {
		return nil, err
	}

	notes := make(map[string]string)
	for _, id := range ids {
		if !strings.HasSuffix(id, "aaa") {
			continue
		}

		filename := joplin.DecryptFilename(id)
		if filename == "" {
			continue
		}

		timestamp := strings.Split(filename, ".")[0]
		if len(timestamp) != 14 {
			continue
		}

		titleLine, err := joplin.GetField(id, "title")
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

func init() {
	joplinDiffCmd.Flags().BoolVar(&debugDiff, "debug", false, "show detailed content comparison for debugging")
	joplinCmd.AddCommand(joplinDiffCmd)
}