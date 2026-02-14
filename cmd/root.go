package cmd

import (
	"fmt"
	"kl/pkg/config"
	"os"

	"github.com/spf13/cobra"
)

var FzfFlag bool
var NoFzfFlag bool
var DirZetFlag string

var Fzf bool
var FzfEnv string
var DirZet string
var FzfOpts []string
var DirExport string
var DirCache string
var ScreenshotTool string
var ScreenshotParams []string
var ScreenrecordTool string
var ScreenrecordParams []string
var SchemaTool string
var SchemaToolParams []string
var Terminal string
var TerminalParams []string

var rootCmd = &cobra.Command{
	Use:   "kl",
	Short: "A note-taking and knowledge management CLI tool",
	Long: `Knowledge is a CLI tool for managing markdown-based notes in a zettelkasten-style system.
It provides commands for creating, searching, editing, and organizing notes with support for
interactive fuzzy finding (fzf) and integration with external tools like Joplin.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := config.CheckInitError(); err != nil {
			return fmt.Errorf("configuration error: %w", err)
		}

		if FzfFlag {
			Fzf = true
		} else {
			Fzf = config.GetFzf()
		}

		if NoFzfFlag {
			Fzf = false
		}

		FzfEnv = config.GetFzfEnv()

		if FzfEnv == "" {
			FzfEnv = "fzf"
		}

		DirExport = config.GetDirExport()
		DirCache = config.DirCache
		ScreenshotTool = config.GetScreenshotTool()
		ScreenshotParams = config.GetScreenshotParams()
		ScreenrecordTool = config.GetScreenrecordTool()
		ScreenrecordParams = config.GetScreenrecordParams()
		SchemaTool = config.GetSchemaTool()
		SchemaToolParams = config.GetSchemaToolParams()
		Terminal = config.GetTerminal()
		TerminalParams = config.GetTerminalParams()

		if Fzf && FzfEnv == "fzf-tmux" {
			var fzfTmuxOpts, fzfOpts string

			if configTmuxOpts := config.GetFzfTmuxOptions(); configTmuxOpts != "" {
				fzfTmuxOpts = configTmuxOpts
			} else {
				fzfTmuxOpts = os.Getenv("FZF_TMUX_OPTS")
			}

			if configOpts := config.GetFzfOptions(); configOpts != "" {
				fzfOpts = configOpts
			} else {
				fzfOpts = os.Getenv("FZF_DEFAULT_OPTS")
			}

			FzfOpts = []string{fzfTmuxOpts, fzfOpts}
		} else if Fzf {
			var fzfOpts string

			if configOpts := config.GetFzfOptions(); configOpts != "" {
				fzfOpts = configOpts
			} else {
				fzfOpts = os.Getenv("FZF_DEFAULT_OPTS")
			}

			FzfOpts = []string{fzfOpts}
		}

		if len(DirZetFlag) > 0 {
			DirZet = DirZetFlag
		} else {
			DirZet = os.Getenv("K_DIR")
		}

		if DirZet == "" {
			return fmt.Errorf("no directory provided. Set K_DIR environment variable or use --dir flag")
		}

		return nil
	},
}

func Execute() {
	rootCmd.PersistentFlags().BoolVarP(&FzfFlag, "fzf", "f", false, "activate fzf")
	rootCmd.PersistentFlags().BoolVarP(&NoFzfFlag, "no-fzf", "z", false, "deactivate fzf")
	rootCmd.PersistentFlags().StringVarP(&DirZetFlag, "dir", "d", "", "define zet repo")
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
}
