package cmd

import (
	"kl/pkg/files"
	"kl/pkg/output"
	"kl/pkg/search"
	"kl/pkg/ui"

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

	searcher := search.NewSearcher(fileList).WithStrategy(search.StrategyHeader)
	if findCaseInsensitive {
		searcher = searcher.WithCaseInsensitive()
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

func runInteractiveFind(args []string, fileList []files.FileInfo) error {
	filteredFiles, err := presearchFind(args, fileList)
	if err != nil {
		return err
	}

	fzfConfig := ui.NewFzfConfigWithCommand(FzfEnv)

	for _, optStr := range FzfOpts {
		if optStr != "" {
			opts, err := ui.ParseOptions(optStr)
			if err == nil {
				fzfConfig.WithOptions(opts...)
			}
		}
	}

	selector := ui.NewFileSelector(fzfConfig)
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
		searcher := search.NewSearcher(fileList).WithStrategy(search.StrategyHeader)
		if findCaseInsensitive {
			searcher = searcher.WithCaseInsensitive()
		}
		if findMatchingStrategy != "" {
			switch findMatchingStrategy {
			case "AND":
				searcher = searcher.WithMatchingStrategy(search.StrategyAND)
			case "OR":
				searcher = searcher.WithMatchingStrategy(search.StrategyOR)
			}
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
