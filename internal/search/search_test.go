package search

import (
	"testing"

	"github.com/maxencetholomier/knowledge/internal/files"
)

func testFiles() []files.FileInfo {
	return []files.FileInfo{
		{Name: "a.md", Content: "# Go tooling\n\nnotes about gofmt"},
		{Name: "b.md", Content: "# Cooking\n\nbake with Go-flour"},
		{Name: "c.md", Content: "# Empty\n\n"},
	}
}

func names(results []SearchResult) []string {
	var out []string
	for _, r := range results {
		out = append(out, r.File.Name)
	}
	return out
}

func TestSearchStrategies(t *testing.T) {
	cases := []struct {
		name     string
		searcher *Searcher
		patterns []string
		want     []string
	}{
		{"or matches any", &Searcher{Files: testFiles()}, []string{"gofmt", "bake"}, []string{"a.md", "b.md"}},
		{"and needs all", &Searcher{Files: testFiles(), MatchingStrategy: StrategyAND},
			[]string{"Go", "bake"}, []string{"b.md"}},
		{"header only", &Searcher{Files: testFiles(), Strategy: StrategyHeader}, []string{"Go"}, []string{"a.md"}},
		{"body only", &Searcher{Files: testFiles(), Strategy: StrategyBody}, []string{"Go"}, []string{"b.md"}},
		{"case insensitive", &Searcher{Files: testFiles(), CaseInsensitive: true}, []string{"cooking"}, []string{"b.md"}},
		{"case sensitive", &Searcher{Files: testFiles()}, []string{"cooking"}, nil},
	}

	for _, c := range cases {
		results, err := c.searcher.Search(c.patterns)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		got := names(results)
		if len(got) != len(c.want) {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("%s: got %v, want %v", c.name, got, c.want)
			}
		}
	}
}

func TestSearchInvalidPattern(t *testing.T) {
	if _, err := (&Searcher{Files: testFiles()}).Search([]string{"("}); err == nil {
		t.Error("invalid regexp must return an error")
	}
}
