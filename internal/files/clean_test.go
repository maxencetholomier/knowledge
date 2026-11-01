package files

import "testing"

func TestIsEmptyNote(t *testing.T) {
	empty := []string{"", "# Title", "# Title\n\n  \n", "\n\n"}
	for _, content := range empty {
		if !IsEmptyNote(content) {
			t.Errorf("%q should be empty", content)
		}
	}
	filled := []string{"# Title\n\ntext", "text", "\n# Title"}
	for _, content := range filled {
		if IsEmptyNote(content) {
			t.Errorf("%q should not be empty", content)
		}
	}
}

func TestIsImage(t *testing.T) {
	if !IsImage("a.PNG") || !IsImage("b.svg") || IsImage("c.md") || IsImage("d.mp4") {
		t.Error("image detection by extension is wrong")
	}
}

func TestLinkedImages(t *testing.T) {
	linked := LinkedImages("![](20240101120000.png) ![alt](sub/dir/x.jpg)", `<img src="y.svg"> [link](z.md)`)
	for _, want := range []string{"20240101120000.png", "x.jpg", "y.svg"} {
		if !linked[want] {
			t.Errorf("%s should be linked, got %v", want, linked)
		}
	}
	if linked["z.md"] || len(linked) != 3 {
		t.Errorf("unexpected entries: %v", linked)
	}
}
