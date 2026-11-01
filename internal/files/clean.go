package files

import (
	"path/filepath"
	"regexp"
	"strings"
)

var imageExtensions = []string{".png", ".jpg", ".jpeg", ".gif", ".bmp", ".svg", ".webp"}

var markdownImagePattern = regexp.MustCompile(`!\[.*?\]\(([^)]+)\)`)
var htmlImagePattern = regexp.MustCompile(`<img[^>]+src=["']([^"']+)["']`)

func IsEmptyNote(content string) bool {
	for i, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || (i == 0 && strings.HasPrefix(line, "#")) {
			continue
		}
		return false
	}
	return true
}

func IsImage(name string) bool {
	lower := strings.ToLower(name)
	for _, ext := range imageExtensions {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

func LinkedImages(contents ...string) map[string]bool {
	linked := make(map[string]bool)
	for _, content := range contents {
		for _, pattern := range []*regexp.Regexp{markdownImagePattern, htmlImagePattern} {
			for _, match := range pattern.FindAllStringSubmatch(content, -1) {
				linked[filepath.Base(match[1])] = true
			}
		}
	}
	return linked
}
