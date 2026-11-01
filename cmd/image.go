package cmd

import (
	"github.com/maxencetholomier/knowledge/pkg/config"
	"github.com/maxencetholomier/knowledge/pkg/files"

	"github.com/spf13/cobra"
)

var imageUseNewTerminal bool

var imageCmd = &cobra.Command{
	Use:     "image",
	Aliases: []string{"i"},
	Short:   "Create a new note with screenshot using Flameshot",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Current
		return files.CreateMediaNote(
			DirZet, cfg.ScreenshotTool, cfg.ScreenshotParams, cfg.Terminal, cfg.TerminalParams,
			".png", "image", imageUseNewTerminal,
		)
	},
}

func init() {
	imageCmd.Flags().BoolVarP(&imageUseNewTerminal, "new-terminal", "", false, "open in new terminal")
	rootCmd.AddCommand(imageCmd)
}
