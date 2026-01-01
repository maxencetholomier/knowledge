package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var quietFlag bool

var manCmd = &cobra.Command{
	Use:   "man [OUTPUT_DIR]",
	Short: "Generate man pages for kl CLI",
	Long: `Generate man pages for the kl CLI tool and all its subcommands.

The man pages will be generated in the specified output directory.
If no output directory is provided, man pages will be generated in /tmp/kl-man/`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		outputDir := "/tmp/kl-man"
		if len(args) > 0 {
			outputDir = args[0]
		}

		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		header := &doc.GenManHeader{
			Title:   "KL",
			Section: "1",
			Manual:  "KL Manual",
			Source:  "kl",
		}

		if err := doc.GenManTree(rootCmd, header, outputDir); err != nil {
			return fmt.Errorf("failed to generate man pages: %w", err)
		}

		if !quietFlag {
			absPath, _ := filepath.Abs(outputDir)
			fmt.Printf("Man pages generated successfully in: %s\n", absPath)
			fmt.Printf("To install system-wide, run: sudo cp %s/*.1 /usr/local/man/man1/ && sudo mandb\n", absPath)
		}

		return nil
	},
}

func init() {
	manCmd.Flags().BoolVarP(&quietFlag, "quiet", "q", false, "suppress output messages")
	rootCmd.AddCommand(manCmd)
}