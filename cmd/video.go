package cmd

import (
	"kl/pkg/files"

	"github.com/spf13/cobra"
)

var videoUseNewTerminal bool

var videoCmd = &cobra.Command{
	Use:     "video",
	Aliases: []string{"v"},
	Short:   "Create a new note with video recording using screenrecord",
	Long: `Create a new note with video recording using screenrecord.

This command generates a timestamped filename, records a video using the screenrecord tool,
and creates a corresponding markdown note with a link to the video file. The note will
contain a markdown link in the format [Video](timestamp.mp4).

The screenrecord command will be called with the target path where the video should be saved.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return files.CreateMediaNote(DirZet, ScreenrecordTool, ScreenrecordParams, Terminal, TerminalParams, ".mp4", "video", videoUseNewTerminal)
	},
}

func init() {
	videoCmd.Flags().BoolVarP(&videoUseNewTerminal, "new-terminal", "", false, "open in new terminal")
	rootCmd.AddCommand(videoCmd)
}
