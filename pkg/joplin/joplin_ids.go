package joplin

import "strings"

func filterIdsByExtension(ids []string, extension string) []string {
	results := []string{}

	for _, id := range ids {
		if strings.HasSuffix(id, extension) {
			results = append(results, id)
		}
	}

	return results
}
