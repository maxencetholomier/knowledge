package anki

import (
	"sort"
	"strings"
)

func CleanDeckLines(data string, localNotes map[string]string) (kept, removed []string, changed bool) {
	seen := make(map[string]bool)
	for line := range strings.SplitSeq(data, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		entry := trimmed
		if idx := strings.Index(entry, " #"); idx != -1 {
			entry = strings.TrimSpace(entry[:idx])
		}
		key := trimmed
		if entry != "" && !strings.HasPrefix(entry, "#") {
			key = entry
			if _, exists := localNotes[strings.TrimSuffix(entry, ".md")]; !exists {
				removed = append(removed, entry)
				continue
			}
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		kept = append(kept, trimmed)
	}

	sort.Sort(sort.Reverse(sort.StringSlice(kept)))
	return kept, removed, DeckContent(kept) != data
}

func DeckContent(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}
