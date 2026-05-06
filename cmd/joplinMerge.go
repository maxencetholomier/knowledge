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
	fileName             string
	title                string
	action               string // "pull_from_joplin", "push_to_joplin", "no_change"
	localUpdate          time.Time
	joplinUpdate         time.Time
	joplinBody           string
	normalizedJoplinBody string
	localBody            string
}

var joplinMergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge notes with Joplin bidirectionally",
	Long: `Synchronize notes between the knowledge base and Joplin by merging changes bidirectionally based on modification timestamps.

IMPORTANT: The merge process uses modification timestamps to determine which version is newer.
The newest version will be preserved and the older version will be overwritten.
If you have modified the same note both locally and in the cloud, there is no conflict resolution -
data loss may occur as the older version will be replaced by the newer one.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		if joplinMergeForceLocal && joplinMergeForceJoplin {
			return fmt.Errorf("--force-local and --force-joplin are mutually exclusive")
		}

		notebookName, notebookId, err := joplin.GetNotebookInfo(joplinMergeNotebook)
		if err != nil {
			return err
		}

		ids, err := joplin.GetIds("notes")
		if err != nil {
			return err
		}

		forceAction := ""
		if joplinMergeForceLocal {
			forceAction = "push_to_joplin"
		} else if joplinMergeForceJoplin {
			forceAction = "pull_from_joplin"
		}

		mergeActions, err := getMergeActions(ids, forceAction)
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
			var cleanBody string
			strippedBody := joplin.StripLeadingHeading(action.joplinBody)
			if action.title != "" {
				cleanBody = "# " + action.title + "\n\n" + strippedBody
			} else {
				cleanBody = "# \n\n" + strippedBody
			}
			file, err := files.Create(DirZet+"/"+action.fileName, joplin.ReplaceIdsToLink(cleanBody))
			if err != nil {
				fmt.Printf("Error pulling %s: %v\n", action.fileName, err)
				continue
			}
			defer file.Close()

		} else if action.action == "push_to_joplin" {
			zetBody, err := os.ReadFile(DirZet + "/" + action.fileName)
			if err != nil {
				fmt.Printf("Error reading %s: %v\n", action.fileName, err)
				continue
			}

			body := string(zetBody)

			err = joplin.PostResourceFromBody(body, DirZet)
			if err != nil {
				fmt.Printf("Error posting resources for %s: %v\n", action.fileName, err)
			}

			if notebookId != "" {
				err = joplin.PutNoteToJoplinWithNotebook(action.fileName, DirZet, notebookId)
			} else {
				err = joplin.PutNoteToJoplin(action.fileName, DirZet)
			}
			if err != nil {
				fmt.Printf("Error pushing %s: %v\n", action.fileName, err)
				continue
			}
		}
	}
	fmt.Printf("\nSuccessfully synchronized %d notes.\n", len(mergeActions))
}

func getMergeActions(ids []string, forceAction string) ([]mergeAction, error) {
	var mergeActions []mergeAction

	for _, id := range ids {
		body, err := joplin.GetField(id, "body")
		if err != nil {
			continue
		}

		title, err := joplin.GetField(id, "title")
		if err != nil {
			title = ""
		}

		fileName := joplin.DecryptFilename(id)
		if fileName == "" {
			continue
		}

		zet_last_update, localErr := files.GetLastUpdate(fileName, DirZet)
		joplin_last_update, joplinErr := joplin.GetLastUpdate(id)

		if localErr != nil && joplinErr != nil {
			continue
		}

		localContent, readErr := os.ReadFile(DirZet + "/" + fileName)

		cleanedBody := joplin.StripLeadingHeading(joplin.ReplaceIdsToLink(body))
		var reconstructed string
		if title != "" {
			reconstructed = "# " + title + "\n\n" + cleanedBody
		} else {
			reconstructed = "# \n\n" + cleanedBody
		}
		normalizedJoplinContent := strings.TrimSpace(reconstructed)

		action := mergeAction{
			fileName:             fileName,
			title:                title,
			localUpdate:          zet_last_update,
			joplinUpdate:         joplin_last_update,
			joplinBody:           body,
			normalizedJoplinBody: normalizedJoplinContent,
		}

		if readErr == nil {
			action.localBody = strings.TrimSpace(string(localContent))
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
			if action.localBody == normalizedJoplinContent {
				action.action = "no_change"
			} else if zet_last_update.Before(joplin_last_update) {
				action.action = "pull_from_joplin"
			} else if joplin_last_update.Before(zet_last_update) {
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

func confirmMerge(mergeActions []mergeAction, notebookName string) (bool, error) {
	fmt.Printf("Will synchronize %d notes:\n", len(mergeActions))

	pullCount := 0
	pushCount := 0

	for _, action := range mergeActions {
		timestamp := action.fileName[:14]
		if action.action == "pull_from_joplin" {
			fmt.Printf("  ← %s - %s (pull from Joplin, updated %s)\n",
				timestamp, action.title, action.joplinUpdate.Format("2006-01-02 15:04"))
			if joplinMergeShowDiff && action.localBody != "" {
				fmt.Printf("    Changes:\n")
				showDiff(action.localBody, action.normalizedJoplinBody, 5)
			}
			pullCount++
		} else if action.action == "push_to_joplin" {
			fmt.Printf("  → %s - %s (push to Joplin, updated %s)\n",
				timestamp, action.title, action.localUpdate.Format("2006-01-02 15:04"))
			if joplinMergeShowDiff && action.localBody != "" {
				fmt.Printf("    Changes:\n")
				showDiff(action.normalizedJoplinBody, action.localBody, 5)
			}
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

func showDiff(localContent, joplinContent string, maxLines int) {
	localLines := strings.Split(localContent, "\n")
	joplinLines := strings.Split(joplinContent, "\n")

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
	joplinMergeCmd.Flags().BoolVar(&joplinMergeForceLocal, "force-local", false, "push all notes from local to Joplin, ignoring timestamps")
	joplinMergeCmd.Flags().BoolVar(&joplinMergeForceJoplin, "force-joplin", false, "pull all notes from Joplin to local, ignoring timestamps")
	joplinCmd.AddCommand(joplinMergeCmd)
}
