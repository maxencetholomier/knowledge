package anki

import "testing"

func TestMarkdownFieldRoundTrip(t *testing.T) {
	note := "# Title <a> & b\n\n```go\nif a < b {\n    return &c\n}\n```\n\n| x | y |\n|---|---|\n| 1 | 2 |"
	if got := DecodeMarkdownField(EncodeMarkdownField(note)); got != note {
		t.Fatalf("round trip mismatch:\n%q\n%q", got, note)
	}
}

func TestDecodeMarkdownFieldEditorMarkup(t *testing.T) {
	field := "<div># Title</div><div><br></div><div>a&nbsp;&nbsp;b<br/>c &lt; d</div>"
	want := "# Title\n\na  b\nc < d"
	if got := DecodeMarkdownField(field); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
