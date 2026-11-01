package anki

import (
	"strings"
	"testing"
)

func TestGenerateDeterministicID(t *testing.T) {
	a, b := GenerateDeterministicID("20240101120000"), GenerateDeterministicID("20240101120000")
	if a != b || a <= 0 {
		t.Errorf("id must be stable and positive, got %d and %d", a, b)
	}
	if GenerateDeterministicID("20240101120001") == a {
		t.Error("different names must not collide")
	}
}

func TestExtractTitleAndBody(t *testing.T) {
	note := "\n#  Title  \n\nline 1\nline 2\n"
	if got := extractTitle(note); got != "Title" {
		t.Errorf("title: got %q", got)
	}
	if got := extractBody(note); got != "line 1\nline 2" {
		t.Errorf("body: got %q", got)
	}
	if got := extractBody("no title"); got != "no title" {
		t.Errorf("body without title: got %q", got)
	}
}

func TestDetectAndConvertTables(t *testing.T) {
	cases := map[string]string{
		"| a | b |\n| 1 | 2 |":            "| a | b |\n| --- | --- |\n| 1 | 2 |",
		"| a | b |\n| 1 | 2 |\n|---|---|": "| a | b |\n|---|---|\n| 1 | 2 |",
		"| a | b |\n|---|---|\n| 1 | 2 |": "| a | b |\n|---|---|\n| 1 | 2 |",
		"text with | pipe\n\nparagraph":   "text with | pipe\n\nparagraph",
	}
	for in, want := range cases {
		if got := detectAndConvertTables(in); got != want {
			t.Errorf("%q:\n got  %q\n want %q", in, got, want)
		}
	}
}

func TestProcessNoteLinks(t *testing.T) {
	got := ProcessNoteLinks("see [other note](20240101120000.md) and [ext](https://x.y)", nil)
	if !strings.Contains(got, ">other note</span>") || strings.Contains(got, "20240101120000.md") {
		t.Errorf("note link not converted: %q", got)
	}
	if !strings.Contains(got, "[ext](https://x.y)") {
		t.Errorf("external link must be kept: %q", got)
	}
}
