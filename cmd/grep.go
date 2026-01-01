package cmd

import (
	"kl/pkg/files"
	"kl/pkg/output"
	"kl/pkg/search"
	"kl/pkg/ui"

	"github.com/spf13/cobra"
)

var (
	grepCaseInsensitive  bool
	grepMatchingStrategy string
)

var grepCmd = &cobra.Command{
	Use:     "grep",
	Aliases: []string{"r"},
	Short:   "Search for text within note content",
	Long: `Search for text patterns within note content using ripgrep.
With fzf enabled, provides an interactive search interface with preview.
Without fzf, searches through note body content and outputs matching results.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if Fzf {
			return runInteractiveGrep(args)
		} else {
			return runTraditionalGrep(args)
		}
	},
}

func presearchGrep(args []string) ([]byte, error) {
	searchTerm := ""
	if len(args) > 0 {
		searchTerm = args[0]
	}

	searcher := search.NewRipgrepSearcher(DirZet).WithFileTypes("md").WithColors()
	if grepCaseInsensitive {
		searcher = searcher.WithCaseInsensitive()
	}

	return searcher.Search(searchTerm)
}

func printResultsGrep(results []search.SearchResult) error {
	var allResults string
	for i, result := range results {
		err := output.PrintToStdout(result.Match, i+1)
		if err != nil {
			return err
		}

		if allResults == "" {
			allResults = result.Match
		} else {
			allResults = allResults + "\n" + result.Match
		}
	}

	return output.PrintToCache(allResults)
}

func runInteractiveGrep(args []string) error {
	rgOutput, err := presearchGrep(args)
	if err != nil {
		return nil
	}

	fzfConfig := ui.NewFzfConfigWithCommand(FzfEnv).WithBaseDir(DirZet)

	for _, optStr := range FzfOpts {
		if optStr != "" {
			opts, err := ui.ParseOptions(optStr)
			if err == nil {
				fzfConfig.WithOptions(opts...)
			}
		}
	}

	selector := ui.NewSearchSelector(fzfConfig)
	selectedFile, err := selector.SelectFromSearchResults(rgOutput)
	if err != nil {
		return err
	}

	if selectedFile != "" {
		fullPath := DirZet + "/" + selectedFile
		files.Edit(fullPath)
	}

	return nil
}

func runTraditionalGrep(args []string) error {
	if len(args) == 0 {
		return nil
	}

	scanner := files.NewScanner(DirZet).WithExtensions("md")
	fileList, err := scanner.ListFiles()
	if err != nil {
		return err
	}

	searcher := search.NewSearcher(fileList).WithStrategy(search.StrategyBody)
	if grepCaseInsensitive {
		searcher = searcher.WithCaseInsensitive()
	}
	if grepMatchingStrategy != "" {
		switch grepMatchingStrategy {
		case "AND":
			searcher = searcher.WithMatchingStrategy(search.StrategyAND)
		case "OR":
			searcher = searcher.WithMatchingStrategy(search.StrategyOR)
		}
	}
	results, err := searcher.Search(args)
	if err != nil {
		return err
	}

	return printResultsGrep(results)
}

func init() {
	grepCmd.Flags().BoolVarP(&grepCaseInsensitive, "case-insensitive", "i", false, "case insensitive search")
	grepCmd.Flags().StringVarP(&grepMatchingStrategy, "matching-strategy", "m", "", "matching strategy (AND/OR)")
	rootCmd.AddCommand(grepCmd)
}
