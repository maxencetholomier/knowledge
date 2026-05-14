package cmd

import (
	"fmt"
	"kl/pkg/files"
	"kl/pkg/joplin"
	"kl/pkg/prompt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var joplinMergeNotebook string
var joplinMergeShowDiff bool
var joplinMergeForceLocal bool
var joplinMergeForceJoplin bool

type mergeAction struct {
	action       string // "pull_from_joplin", "push_to_joplin", "no_change"
	fileContent  string
	fileName     string
	fileUpdate   time.Time
	joplinBody   string
	joplinTitle  string
	joplinUpdate time.Time
}

var joplinMergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge notes with Joplin bidirectionally",
	Long: `Synchronize notes between the knowledge base and Joplin by merging changes bidirectionally based on modification fileTimestamps.

IMPORTANT: The merge process uses modification fileTimestamps to determine which version is newer.
The newest version will be preserved and the older version will be overwritten.
If you have modified the same note both locally and in the cloud, there is no conflict resolution -
data loss may occur as the older version will be replaced by the newer one.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		if joplinMergeForceLocal && joplinMergeForceJoplin {
			return fmt.Errorf("--force-local and --force-joplin are mutually exclusive")
		}

		forceAction := ""
		if joplinMergeForceLocal {
			forceAction = "push_to_joplin"
		} else if joplinMergeForceJoplin {
			forceAction = "pull_from_joplin"
		}

		notebookName, notebookId, err := joplin.GetNotebookInfo(joplinMergeNotebook)
		if err != nil {
			return err
		}

		query := joplin.GetQuery{Fields: []string{"title", "body", "updated_time"}, NotebookID: notebookId}
		joplinNotes, err := joplin.GetNotes(query)
		if err != nil {
			return err
		}

		mergeActions, err := getMergeActions(joplinNotes, forceAction)
		if err != nil {
			return err
		}

		if len(mergeActions) == 0 {
			fmt.Println("No notes need synchronization - all notes are up to date.")
			return nil
		}

		confirmed, err := confirmMerge(mergeActions, notebookName)
		if err != nil {
			return err
		}
		if !confirmed {
			return nil
		}

		applyMergeActions(mergeActions, notebookId)
		return nil
	},
}

func applyMergeActions(mergeActions []mergeAction, notebookId string) {
	fmt.Printf("\nSynchronizing %d notes...\n", len(mergeActions))

	for _, action := range mergeActions {
		if action.action == "pull_from_joplin" {
			fileContent := "# " + action.joplinTitle + "\n\n" + joplin.StripLeadingHeading(action.joplinBody)
			file, err := files.Create(DirZet+"/"+action.fileName, joplin.ReplaceIdsToLink(fileContent))
			if err != nil {
				fmt.Printf("Error pulling %s: %v\n", action.fileName, err)
				continue
			}
			defer file.Close()

		} else if action.action == "push_to_joplin" {
			body, err := os.ReadFile(DirZet + "/" + action.fileName)
			if err != nil {
				fmt.Printf("Error reading %s: %v\n", action.fileName, err)
				continue
			}

			err = joplin.SendResourceFromBody(string(body), DirZet)
			if err != nil {
				fmt.Printf("Error posting resources for %s: %v\n", action.fileName, err)
			}

			query := joplin.WriteQuery{Method: joplin.PUT, FileName: action.fileName, DirZet: DirZet, NotebookId: notebookId}
			if err = joplin.Send(query); err != nil {
				fmt.Printf("Error pushing %s: %v\n", action.fileName, err)
				continue
			}
		}
	}
	fmt.Printf("\nSuccessfully synchronized %d notes.\n", len(mergeActions))
}

func getMergeActions(joplinNotes []joplin.Note, forceAction string) ([]mergeAction, error) {
	fileTimestamps, err := files.GetLastUpdates(DirZet)
	if err != nil {
		return nil, err
	}

	var mergeActions []mergeAction
	for _, note := range joplinNotes {
		fileName := joplin.IdToFilename(note.ID)
		if fileName == "" {
			continue
		}

		fileLastUpdate, inGit := fileTimestamps[fileName]
		var localErr error
		if !inGit {
			localErr = fmt.Errorf("no git history for %s", fileName)
		}

		joplinLastUpdate := note.UpdatedTime
		var joplinErr error
		if joplinLastUpdate.IsZero() {
			joplinErr = fmt.Errorf("no updated_time for note %s", note.ID)
		}

		if localErr != nil && joplinErr != nil {
			continue
		}

		fileContent, readErr := os.ReadFile(DirZet + "/" + fileName)
		joplinAsLocal := strings.TrimSpace("# " + note.Title + "\n\n" + joplin.StripLeadingHeading(joplin.ReplaceIdsToLink(note.Body)))

		action := mergeAction{
			fileName:     fileName,
			joplinTitle:  note.Title,
			fileUpdate:   fileLastUpdate,
			joplinUpdate: joplinLastUpdate,
			joplinBody:   note.Body,
		}

		if readErr == nil {
			action.fileContent = strings.TrimSpace(string(fileContent))
		}

		if forceAction != "" {
			if forceAction == "push_to_joplin" && (localErr != nil || readErr != nil) {
				continue
			}
			action.action = forceAction
		} else if localErr != nil {
			action.action = "pull_from_joplin"
		} else if joplinErr != nil {
			action.action = "push_to_joplin"
		} else if readErr != nil {
			action.action = "pull_from_joplin"
		} else {
			if action.fileContent == joplinAsLocal {
				action.action = "no_change"
			} else if fileLastUpdate.Before(joplinLastUpdate) {
				action.action = "pull_from_joplin"
			} else if joplinLastUpdate.Before(fileLastUpdate) {
				action.action = "push_to_joplin"
			} else {
				action.action = "pull_from_joplin"
			}
		}

		if action.action != "no_change" {
			mergeActions = append(mergeActions, action)
		}
	}

	return mergeActions, nil
}

func printMergeAction(action mergeAction) {
	joplinAsLocal := strings.TrimSpace("# " + action.joplinTitle + "\n\n" + joplin.StripLeadingHeading(joplin.ReplaceIdsToLink(action.joplinBody)))

	var arrow, direction string
	var updateTime time.Time
	var diffFrom, diffTo string

	if action.action == "pull_from_joplin" {
		arrow, direction = "←", "pull from Joplin"
		updateTime = action.joplinUpdate
		diffFrom, diffTo = action.fileContent, joplinAsLocal
	} else if action.action == "push_to_joplin" {
		arrow, direction = "→", "push to Joplin"
		updateTime = action.fileUpdate
		diffFrom, diffTo = joplinAsLocal, action.fileContent
	}

	fmt.Printf("  %s %s - %s (%s, updated %s)\n",
		arrow, action.fileName[:14], action.joplinTitle, direction, updateTime.Format("2006-01-02 15:04"))
	if joplinMergeShowDiff && action.fileContent != "" {
		fmt.Printf("    Changes:\n")
		showDiff(diffFrom, diffTo, 5)
	}
}

func confirmMerge(mergeActions []mergeAction, notebookName string) (bool, error) {
	fmt.Printf("Will synchronize %d notes:\n", len(mergeActions))

	pullCount := 0
	pushCount := 0

	for _, action := range mergeActions {
		printMergeAction(action)
		if action.action == "pull_from_joplin" {
			pullCount++
		} else if action.action == "push_to_joplin" {
			pushCount++
		}
	}

	fmt.Printf("\nSummary: %d notes to pull from Joplin, %d notes to push to Joplin\n", pullCount, pushCount)
	if pushCount > 0 && notebookName != "" {
		fmt.Printf("Notes pushed to Joplin will be moved to notebook '%s'\n", notebookName)
	}

	confirmed, err := prompt.Confirm("Proceed with synchronization?")
	if err != nil {
		return false, err
	}
	if !confirmed {
		fmt.Println("Synchronization cancelled.")
		return false, nil
	}

	return true, nil
}

func showDiff(localContent, joplinAsLocal string, maxLines int) {
	localLines := strings.Split(localContent, "\n")
	joplinLines := strings.Split(joplinAsLocal, "\n")

	maxLen := len(localLines)
	if len(joplinLines) > maxLen {
		maxLen = len(joplinLines)
	}

	if maxLen > maxLines {
		maxLen = maxLines
	}

	for i := 0; i < maxLen; i++ {
		var localLine, joplinLine string
		if i < len(localLines) {
			localLine = localLines[i]
		}
		if i < len(joplinLines) {
			joplinLine = joplinLines[i]
		}

		if localLine != joplinLine {
			if localLine != "" {
				fmt.Printf("    - %s\n", localLine)
			}
			if joplinLine != "" {
				fmt.Printf("    + %s\n", joplinLine)
			}
		}
	}

	if len(localLines) > maxLines || len(joplinLines) > maxLines {
		fmt.Printf("    ... (showing first %d lines)\n", maxLines)
	}
}

func init() {
	joplinMergeCmd.Flags().StringVarP(&joplinMergeNotebook, "notebook", "n", "", "specify the notebook to move local notes to when pushing to Joplin")
	joplinMergeCmd.Flags().BoolVar(&joplinMergeShowDiff, "diff", false, "show diff of changes for each file")
	joplinMergeCmd.Flags().BoolVar(&joplinMergeForceLocal, "force-local", false, "push all notes from local to Joplin, ignoring fileTimestamps")
	joplinMergeCmd.Flags().BoolVar(&joplinMergeForceJoplin, "force-joplin", false, "pull all notes from Joplin to local, ignoring fileTimestamps")
	joplinCmd.AddCommand(joplinMergeCmd)
}
