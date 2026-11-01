package merge

import (
	"fmt"
	"io"
	"slices"
	"strings"
	"time"
)

type Direction string

const (
	None Direction = ""
	Pull Direction = "pull"
	Push Direction = "push"
)

func Decide(local, remote string, localTime, remoteTime time.Time, force Direction) Direction {
	if force != None {
		return force
	}
	if local == remote {
		return None
	}
	switch {
	case localTime.IsZero() && remoteTime.IsZero():
		return None
	case localTime.IsZero():
		return Pull
	case remoteTime.IsZero():
		return Push
	case remoteTime.Before(localTime):
		return Push
	default:
		return Pull
	}
}

func MissingKeys[A, B any](from map[string]A, in map[string]B) []string {
	var missing []string
	for key := range from {
		if _, exists := in[key]; !exists {
			missing = append(missing, key)
		}
	}
	slices.Sort(missing)
	return missing
}

func CommonKeys[A, B any](a map[string]A, b map[string]B) []string {
	var common []string
	for key := range a {
		if _, exists := b[key]; exists {
			common = append(common, key)
		}
	}
	slices.Sort(common)
	return common
}

func PrintDiff(w io.Writer, from, to string, maxLines int) {
	fromLines := strings.Split(from, "\n")
	toLines := strings.Split(to, "\n")
	shown := min(max(len(fromLines), len(toLines)), maxLines)

	for i := range shown {
		var fromLine, toLine string
		if i < len(fromLines) {
			fromLine = fromLines[i]
		}
		if i < len(toLines) {
			toLine = toLines[i]
		}
		if fromLine == toLine {
			continue
		}
		if fromLine != "" {
			fmt.Fprintf(w, "    - %s\n", fromLine)
		}
		if toLine != "" {
			fmt.Fprintf(w, "    + %s\n", toLine)
		}
	}

	if len(fromLines) > maxLines || len(toLines) > maxLines {
		fmt.Fprintf(w, "    ... (showing first %d lines)\n", maxLines)
	}
}
