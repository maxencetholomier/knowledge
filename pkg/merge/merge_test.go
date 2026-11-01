package merge

import (
	"bytes"
	"slices"
	"testing"
	"time"
)

func TestDecide(t *testing.T) {
	older := time.Unix(100, 0)
	newer := time.Unix(200, 0)
	var zero time.Time

	cases := []struct {
		name                  string
		local, remote         string
		localTime, remoteTime time.Time
		force                 Direction
		want                  Direction
	}{
		{"identical content", "a", "a", older, newer, None, None},
		{"force wins over identical content", "a", "a", older, newer, Push, Push},
		{"remote newer", "a", "b", older, newer, None, Pull},
		{"local newer", "a", "b", newer, older, None, Push},
		{"same time pulls", "a", "b", older, older, None, Pull},
		{"no local history pulls", "a", "b", zero, older, None, Pull},
		{"no remote time pushes", "a", "b", older, zero, None, Push},
		{"no time at all does nothing", "a", "b", zero, zero, None, None},
		{"force pull ignores times", "a", "b", newer, older, Pull, Pull},
	}

	for _, c := range cases {
		if got := Decide(c.local, c.remote, c.localTime, c.remoteTime, c.force); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

func TestMissingAndCommonKeys(t *testing.T) {
	local := map[string]string{"c": "", "a": "", "b": ""}
	remote := map[string]bool{"b": true, "d": true}

	if got := MissingKeys(local, remote); !slices.Equal(got, []string{"a", "c"}) {
		t.Errorf("MissingKeys got %v", got)
	}
	if got := MissingKeys(remote, local); !slices.Equal(got, []string{"d"}) {
		t.Errorf("MissingKeys reversed got %v", got)
	}
	if got := CommonKeys(local, remote); !slices.Equal(got, []string{"b"}) {
		t.Errorf("CommonKeys got %v", got)
	}
}

func TestPrintDiff(t *testing.T) {
	var out bytes.Buffer
	PrintDiff(&out, "same\nold\n\nx", "same\nnew\n\ny", 3)

	want := "    - old\n    + new\n    ... (showing first 3 lines)\n"
	if out.String() != want {
		t.Errorf("got %q, want %q", out.String(), want)
	}
}
