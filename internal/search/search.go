package search

import (
	"github.com/maxencetholomier/knowledge/internal/files"
	"os/exec"
	"regexp"
)

type SearchStrategy string

const (
	StrategyAll    SearchStrategy = "all"
	StrategyHeader SearchStrategy = "header"
	StrategyBody   SearchStrategy = "body"
)

type MatchingStrategy string

const (
	StrategyAND MatchingStrategy = "AND"
	StrategyOR  MatchingStrategy = "OR"
)

type SearchResult struct {
	File    files.FileInfo
	Match   string
	Line    int
	Content string
}

type Searcher struct {
	Files            []files.FileInfo
	Strategy         SearchStrategy
	CaseInsensitive  bool
	MatchingStrategy MatchingStrategy
}

func (s *Searcher) Search(patterns []string) ([]SearchResult, error) {
	var results []SearchResult

	for _, file := range s.Files {
		match, err := s.searchInFile(file, patterns)
		if err != nil {
			return nil, err
		}

		if match != "" {
			results = append(results, SearchResult{
				File:    file,
				Match:   match,
				Content: match,
			})
		}
	}

	return results, nil
}

func (s *Searcher) searchInFile(file files.FileInfo, patterns []string) (string, error) {
	var searchContent string
	var err error

	switch s.Strategy {
	case StrategyHeader:
		searchContent, err = file.GetTitle()
	case StrategyBody:
		searchContent, err = file.GetBody()
	default:
		if err := file.LoadContent(); err != nil {
			return "", err
		}
		searchContent = file.Content
	}

	if err != nil {
		return "", err
	}

	if s.MatchingStrategy == StrategyAND {
		for _, pattern := range patterns {
			flags := ""
			if s.CaseInsensitive {
				flags = "(?i)"
			}
			re, err := regexp.Compile(flags + pattern)
			if err != nil {
				return "", err
			}

			if !re.MatchString(searchContent) {
				return "", nil
			}
		}
	} else {
		matched := false
		for _, pattern := range patterns {
			flags := ""
			if s.CaseInsensitive {
				flags = "(?i)"
			}
			re, err := regexp.Compile(flags + pattern)
			if err != nil {
				return "", err
			}

			if re.MatchString(searchContent) {
				matched = true
				break
			}
		}
		if !matched {
			return "", nil
		}
	}

	title, err := file.GetTitle()
	if err != nil {
		return "", err
	}

	return title, nil
}

type RipgrepSearcher struct {
	Dir             string
	FileTypes       []string
	CaseInsensitive bool
	Colors          bool
}

func (r *RipgrepSearcher) Search(pattern string) ([]byte, error) {
	var args []string

	if r.Colors {
		args = []string{
			"--color=always",
			"--line-number",
			"--no-heading",
			"--with-filename",
			"--colors", "path:fg:magenta",
			"--colors", "path:style:bold",
			"--colors", "line:fg:green",
		}

		if r.CaseInsensitive {
			args = append(args, "--ignore-case")
		} else {
			args = append(args, "--smart-case")
		}

		if pattern != "" {
			args = append(args, "--colors", "match:fg:red")
			args = append(args, "--colors", "match:style:bold")
		} else {
			args = append(args, "--colors", "match:none")
		}

		for _, fileType := range r.FileTypes {
			args = append(args, "--type", fileType)
		}

		if pattern != "" {
			args = append(args, pattern)
		} else {
			args = append(args, ".")
		}

		args = append(args, ".")

		cmd := exec.Command("rg", args...)
		cmd.Dir = r.Dir
		return cmd.Output()
	}

	args = []string{
		"--line-number",
		"--smart-case",
	}

	for _, fileType := range r.FileTypes {
		args = append(args, "--type", fileType)
	}

	if pattern != "" {
		args = append(args, pattern)
	} else {
		args = append(args, ".*")
	}
	args = append(args, r.Dir)

	cmd := exec.Command("rg", args...)
	return cmd.Output()
}
