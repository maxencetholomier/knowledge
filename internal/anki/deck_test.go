package anki

import (
	"slices"
	"testing"
)

func TestCleanDeckLines(t *testing.T) {
	local := map[string]string{"20240101120000": "a", "20240101120001": "b"}
	data := "\n20240101120000.md # a\n  20240101120001.md # b  \n20240101120000.md # a\n" +
		"20230101120000.md # gone\n# comment\n# comment\n"

	kept, removed, changed := CleanDeckLines(data, local)

	wantKept := []string{"20240101120001.md # b", "20240101120000.md # a", "# comment"}
	if !slices.Equal(kept, wantKept) {
		t.Errorf("kept = %q, want %q", kept, wantKept)
	}
	if !slices.Equal(removed, []string{"20230101120000.md"}) {
		t.Errorf("removed = %q", removed)
	}
	if !changed {
		t.Error("expected changed")
	}

	if _, _, changed := CleanDeckLines(DeckContent(kept), local); changed {
		t.Error("a clean deck must not be reported as changed")
	}
}
