package anki

import (
	"html"
	"regexp"
	"strings"
)

var lineBreakTag = regexp.MustCompile(`(?i)<br\s*/?>`)
var emptyDiv = regexp.MustCompile(`(?i)<div[^>]*>\s*<br\s*/?>\s*</div>`)
var closingDiv = regexp.MustCompile(`(?i)</div>`)
var openingDiv = regexp.MustCompile(`(?i)<div[^>]*>`)

func EncodeMarkdownField(markdown string) string {
	return strings.ReplaceAll(html.EscapeString(strings.TrimSpace(markdown)), "\n", "<br>")
}

func DecodeMarkdownField(field string) string {
	text := emptyDiv.ReplaceAllString(field, "\n")
	text = closingDiv.ReplaceAllString(text, "\n")
	text = openingDiv.ReplaceAllString(text, "")
	text = lineBreakTag.ReplaceAllString(text, "\n")
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, " ", " ")
	text = html.UnescapeString(text)
	return strings.TrimSpace(strings.ReplaceAll(text, "\r\n", "\n"))
}
