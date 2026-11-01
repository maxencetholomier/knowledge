package cmd

import (
	"strings"

	"github.com/maxencetholomier/knowledge/internal/files"
	"github.com/maxencetholomier/knowledge/internal/output"
	"github.com/maxencetholomier/knowledge/internal/search"
	"github.com/maxencetholomier/knowledge/internal/ui"

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

	searcher := &search.RipgrepSearcher{
		Dir: DirZet, FileTypes: []string{"md"}, Colors: true, CaseInsensitive: grepCaseInsensitive,
	}
	return searcher.Search(searchTerm)
}

func printResultsGrep(results []search.SearchResult) error {
	var allResults string
	for i, result := range results {
		err := output.PrintToStdout(result.Match, i+1, len(results))
		if err != nil {
			return err
		}

		timestamp := strings.TrimSuffix(result.File.Name, ".md")
		line := timestamp + " " + result.Match
		if allResults == "" {
			allResults = line
		} else {
			allResults = allResults + "\n" + line
		}
	}

	return output.PrintToCache(allResults)
}

func runInteractiveGrep(args []string) error {
	rgOutput, err := presearchGrep(args)
	if err != nil {
		return nil
	}

	selector := &ui.SearchSelector{Config: fzfConfig()}
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

	searcher := &search.Searcher{
		Files:            fileList,
		Strategy:         search.StrategyBody,
		CaseInsensitive:  grepCaseInsensitive,
		MatchingStrategy: search.MatchingStrategy(grepMatchingStrategy),
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
