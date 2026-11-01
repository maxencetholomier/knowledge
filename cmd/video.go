package cmd

import (
	"github.com/maxencetholomier/knowledge/internal/config"
	"github.com/maxencetholomier/knowledge/internal/files"

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
		cfg := config.Current
		return files.CreateMediaNote(
			DirZet, cfg.ScreenrecordTool, cfg.ScreenrecordParams, cfg.Terminal, cfg.TerminalParams,
			".mp4", "video", videoUseNewTerminal,
		)
	},
}

func init() {
	videoCmd.Flags().BoolVarP(&videoUseNewTerminal, "new-terminal", "", false, "open in new terminal")
	rootCmd.AddCommand(videoCmd)
}
