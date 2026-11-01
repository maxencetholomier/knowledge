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
	findCaseInsensitive  bool
	findMatchingStrategy string
)

var findCmd = &cobra.Command{
	Use:     "find",
	Aliases: []string{"f"},
	Short:   "Find notes by searching in titles",
	Long: `Find notes by searching in their titles (headers).
With fzf enabled, provides an interactive file selector.
Without fzf, searches through note titles and outputs matching results.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		scanner := files.NewScanner(DirZet).WithExtensions("md")
		fileList, err := scanner.ListFiles()
		if err != nil {
			return err
		}

		if Fzf {
			return runInteractiveFind(args, fileList)
		} else {
			return runTraditionalFind(args, fileList)
		}
	},
}

func presearchFind(args []string, fileList []files.FileInfo) ([]files.FileInfo, error) {
	if len(args) == 0 {
		return fileList, nil
	}

	searcher := &search.Searcher{
		Files: fileList, Strategy: search.StrategyHeader, CaseInsensitive: findCaseInsensitive,
	}
	results, err := searcher.Search(args)
	if err != nil {
		return nil, err
	}

	filteredFiles := make([]files.FileInfo, len(results))
	for i, result := range results {
		filteredFiles[i] = result.File
	}

	return filteredFiles, nil
}

func fileListToResults(fileList []files.FileInfo) []search.SearchResult {
	results := make([]search.SearchResult, 0, len(fileList))
	for _, file := range fileList {
		title, err := file.GetTitle()
		if err != nil {
			title = file.Name
		}
		results = append(results, search.SearchResult{
			File:    file,
			Match:   title,
			Content: title,
		})
	}
	return results
}

func printResultsFind(results []search.SearchResult) error {
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

func runInteractiveFind(args []string, fileList []files.FileInfo) error {
	filteredFiles, err := presearchFind(args, fileList)
	if err != nil {
		return err
	}

	selector := &ui.FileSelector{Config: fzfConfig()}
	selectedFile, err := selector.SelectFile(filteredFiles, DirZet)
	if err != nil {
		return err
	}

	if selectedFile != "" {
		files.Edit(selectedFile)
	}

	return nil
}

func runTraditionalFind(args []string, fileList []files.FileInfo) error {
	var results []search.SearchResult
	var err error

	if len(args) == 0 {
		results = fileListToResults(fileList)
	} else {
		searcher := &search.Searcher{
			Files:            fileList,
			Strategy:         search.StrategyHeader,
			CaseInsensitive:  findCaseInsensitive,
			MatchingStrategy: search.MatchingStrategy(findMatchingStrategy),
		}
		results, err = searcher.Search(args)
		if err != nil {
			return err
		}
	}

	return printResultsFind(results)
}

func init() {
	findCmd.Flags().BoolVarP(&findCaseInsensitive, "case-insensitive", "i", false, "case insensitive search")
	findCmd.Flags().StringVarP(&findMatchingStrategy, "matching-strategy", "m", "", "matching strategy (AND/OR)")
	rootCmd.AddCommand(findCmd)
}
