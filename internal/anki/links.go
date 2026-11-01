package anki

import (
	"fmt"
	"regexp"
	"strings"
)

type NoteLink struct {
	OriginalText string
	LinkText     string
	TargetID     string
}

func ExtractNoteLinks(content string) []NoteLink {
	linkRegex := regexp.MustCompile(`\[([^\]]+)\]\(([0-9]{14})\.md\)`)
	matches := linkRegex.FindAllStringSubmatch(content, -1)

	var links []NoteLink
	for _, match := range matches {
		if len(match) >= 3 {
			links = append(links, NoteLink{
				OriginalText: match[0],
				LinkText:     match[1],
				TargetID:     match[2],
			})
		}
	}
	return links
}

func ProcessNoteLinks(content string, noteTitleMap map[string]string) string {
	links := ExtractNoteLinks(content)

	result := content
	for _, link := range links {
		ankiLink := ConvertLinkToAnki(link, noteTitleMap)
		result = strings.ReplaceAll(result, link.OriginalText, ankiLink)
	}

	return result
}

func ConvertLinkToAnki(link NoteLink, noteTitleMap map[string]string) string {
	return fmt.Sprintf(`<span style="color: #0066cc; font-weight: 500;">%s</span>`, link.LinkText)
}
