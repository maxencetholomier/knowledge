package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/maxencetholomier/knowledge/internal/config"
	"github.com/maxencetholomier/knowledge/internal/prompt"
	"github.com/maxencetholomier/knowledge/internal/ui"
	"github.com/spf13/cobra"
)

var FzfFlag bool
var NoFzfFlag bool
var DirZetFlag string
var AssumeYes bool

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
	SilenceUsage: true,
	Version:      version(),
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

func fzfConfig() *ui.FzfConfig {
	options := []string{"--ansi"}
	for _, optStr := range FzfOpts {
		if opts, err := ui.ParseOptions(optStr); err == nil {
			options = append(options, opts...)
		}
	}
	return &ui.FzfConfig{Command: FzfEnv, Options: options, BaseDir: DirZet}
}

func confirm(question string) (bool, error) {
	if AssumeYes {
		return true, nil
	}
	return prompt.Confirm(question)
}

func version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			return setting.Value[:min(12, len(setting.Value))]
		}
	}
	return info.Main.Version
}

func batchError(failed, total int, what string) error {
	if failed == 0 {
		return nil
	}
	return fmt.Errorf("%d of %d %s failed", failed, total, what)
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
	rootCmd.PersistentFlags().BoolVarP(&AssumeYes, "yes", "y", false, "answer yes to every confirmation")
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
