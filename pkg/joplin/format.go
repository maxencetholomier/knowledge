package joplin

import (
	"fmt"
	"regexp"
	"strings"
)

func replaceTimestampToIds(line string) (string, error) {
	re := regexp.MustCompile(`[0-9]{14}(?:_[0-9]+)?\.(md|png|jpeg|jpg|svg)?`)

	index := 0
	result := re.ReplaceAllStringFunc(line, func(match string) string {
		new_match := ":/" + FilenameToId(match, index)
		index = index + 1
		return new_match
	})

	return result, nil
}

func ReplaceIdsToLink(line string) string {
	reImage := regexp.MustCompile(`!\[(.*?)\]\(:/([a-zA-Z0-9]{1,32})\)`)

	result := reImage.ReplaceAllStringFunc(line, func(match string) string {
		parts := reImage.FindStringSubmatch(match)
		alt_name := parts[1]
		id := parts[2]
		new_match := IdToFilename(id)
		return "![" + alt_name + "](" + new_match + ")"
	})

	reLink := regexp.MustCompile(`\[(.*?)\]\(:/([a-zA-Z0-9]{1,32})\)`)

	result = reLink.ReplaceAllStringFunc(result, func(match string) string {
		parts := reLink.FindStringSubmatch(match)
		alt_name := parts[1]
		id := parts[2]
		new_match := IdToFilename(id)
		return "[" + alt_name + "](" + new_match + ")"
	})

	return result
}

func StripLeadingHeading(body string) string {
	if !strings.HasPrefix(body, "# ") {
		return body
	}
	newline := strings.Index(body, "\n")
	if newline == -1 {
		return ""
	}
	return strings.TrimLeft(body[newline+1:], "\n")
}

func NoteToMarkdown(title, body, timestamp string) (string, error) {
	content := "# " + title + "\n\n" + StripLeadingHeading(body)
	content, err := convertResourceLinks(content, timestamp)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(content, "&nbsp;", ""), nil
}

func convertResourceLinks(input, timestamp string) (string, error) {
	id, err := replaceTimestampToIds(timestamp)
	if err != nil {
		return "", err
	}
	return replacingJoplinLink(input, id)
}

func replacingJoplinLink(input string, timestamp string) (string, error) {
	pattern := `\[.*?\]\(:/[a-zA-Z0-9]{1,32}\)`
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}
	index := 0
	result := re.ReplaceAllStringFunc(input, func(match string) string {
		replacement := "![](" + timestamp + "_" + fmt.Sprintf("%d", index) + ".png)"
		index++
		return replacement
	})

	return result, nil
}
