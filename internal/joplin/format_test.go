package joplin

import "testing"

func TestIdRoundTrip(t *testing.T) {
	for _, name := range []string{"20240101120000.md", "20240101120000_3.png", "20240101120000_3.svg"} {
		id := FilenameToId(name, 3)
		if len(id) != 32 {
			t.Errorf("%s: id %q is not 32 characters", name, id)
		}
		if got := IdToFilename(id); got != name {
			t.Errorf("%s: round trip gave %q (id %s)", name, got, id)
		}
	}
	if got := IdToFilename("short"); got != "" {
		t.Errorf("foreign id should give empty name, got %q", got)
	}
}

func TestReplaceIdsToLink(t *testing.T) {
	body := "see [note](:/" + FilenameToId("20240101120000.md", 0) +
		") and ![img](:/" + FilenameToId("20240101120000_1.png", 1) + ")"
	want := "see [note](20240101120000.md) and ![img](20240101120000_1.png)"
	if got := ReplaceIdsToLink(body); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStripLeadingHeading(t *testing.T) {
	cases := map[string]string{
		"# Title\n\nbody":  "body",
		"# Title":          "",
		"no heading\nbody": "no heading\nbody",
	}
	for in, want := range cases {
		if got := StripLeadingHeading(in); got != want {
			t.Errorf("%q: got %q, want %q", in, got, want)
		}
	}
}

func TestNoteToMarkdown(t *testing.T) {
	got, err := NoteToMarkdown("Title", "# Title\n\ntext&nbsp;![a](:/abc123)\n[b](:/def456)", "20240101120000")
	if err != nil {
		t.Fatal(err)
	}
	want := "# Title\n\ntext![](20240101120000_0.png)\n![](20240101120000_1.png)"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
