package cmd

import (
	"fmt"
	"os"

	"github.com/maxencetholomier/knowledge/internal/config"
	"github.com/spf13/cobra"
)

var FzfFlag bool
var NoFzfFlag bool
var DirZetFlag string

var Fzf bool
var FzfEnv string
var DirZet string
var FzfOpts []string

var rootCmd = &cobra.Command{
	Use:   "kl",
	Short: "A note-taking and knowledge management CLI tool",
	Long: `Knowledge is a CLI tool for managing markdown-based notes in a zettelkasten-style system.
It provides commands for creating, searching, editing, and organizing notes with support for
interactive fuzzy finding (fzf) and integration with external tools like Joplin.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Load(); err != nil {
			return fmt.Errorf("configuration error: %w", err)
		}

		Fzf = (config.Current.FZF || FzfFlag) && !NoFzfFlag
		FzfEnv = config.Current.FZFEnv

		if Fzf {
			fzfOpts := envOr(config.Current.FZFOptions, "FZF_DEFAULT_OPTS")
			if FzfEnv == "fzf-tmux" {
				FzfOpts = []string{envOr(config.Current.FZFTmuxOptions, "FZF_TMUX_OPTS"), fzfOpts}
			} else {
				FzfOpts = []string{fzfOpts}
			}
		}

		DirZet = envOr(DirZetFlag, "K_DIR")
		if DirZet == "" {
			return fmt.Errorf("no directory provided. Set K_DIR environment variable or use --dir flag")
		}

		return nil
	},
}

func envOr(value, envName string) string {
	if value != "" {
		return value
	}
	return os.Getenv(envName)
}

func Execute() {
	rootCmd.PersistentFlags().BoolVarP(&FzfFlag, "fzf", "f", false, "activate fzf")
	rootCmd.PersistentFlags().BoolVarP(&NoFzfFlag, "no-fzf", "z", false, "deactivate fzf")
	rootCmd.PersistentFlags().StringVarP(&DirZetFlag, "dir", "d", "", "define zet repo")
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
