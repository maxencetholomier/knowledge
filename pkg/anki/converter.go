package anki

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	genanki "github.com/npcnixel/genanki-go"
	"github.com/russross/blackfriday/v2"
)

func generateDeterministicID(noteFilename string) int64 {
	hash := sha256.Sum256([]byte(noteFilename))
	id := int64(binary.BigEndian.Uint64(hash[:8]))

	if id < 0 {
		id = -id
	}
	return id % math.MaxInt64
}

func ConvertNote(notePath string, linkMap map[string]string) (*genanki.Note, []MediaFile, error) {
	content, err := os.ReadFile(notePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read note: %w", err)
	}

	if len(content) == 0 {
		return nil, nil, fmt.Errorf("note is empty")
	}

	text := string(content)
	front := extractTitle(text)
	back := extractBody(text)

	if front == "" {
		filename := filepath.Base(notePath)
		front = strings.TrimSuffix(filename, ".md")
	}

	back = ProcessNoteLinks(back, linkMap)

	baseDir := filepath.Dir(notePath)
	mediaFiles, err := ExtractImages(back, baseDir)
	if err != nil {
		return nil, nil, fmt.Errorf("image error: %w", err)
	}

	back = processCodeBlocks(back)
	backHTML := markdownToHTML(back)

	noteFilename := strings.TrimSuffix(filepath.Base(notePath), ".md")

	csum := int64(0)
	for _, c := range front {
		csum = (csum + int64(c)) % 0xffff
	}

	note := &genanki.Note{
		ID:        generateDeterministicID(noteFilename),
		ModelID:   BasicModelID,
		Fields:    []string{front, string(backHTML)},
		Tags:      []string{noteFilename},
		Modified:  time.Now(),
		SortField: front,
		CheckSum:  csum,
	}

	return note, mediaFiles, nil
}

func extractTitle(content string) string {
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			title := strings.TrimPrefix(line, "#")
			title = strings.TrimSpace(title)
			return title
		}
	}

	return ""
}

func extractBody(content string) string {
	lines := strings.Split(content, "\n")
	foundTitle := false
	var bodyLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !foundTitle && strings.HasPrefix(trimmed, "#") {
			foundTitle = true
			continue
		}
		if foundTitle {
			bodyLines = append(bodyLines, line)
		}
	}

	if !foundTitle {
		return content
	}

	return strings.TrimSpace(strings.Join(bodyLines, "\n"))
}

func processCodeBlocks(markdown string) string {
	codeBlockRegex := regexp.MustCompile("```([a-zA-Z0-9]*)\n([\\s\\S]*?)```")

	result := codeBlockRegex.ReplaceAllStringFunc(markdown, func(match string) string {
		submatches := codeBlockRegex.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match
		}

		language := strings.TrimSpace(submatches[1])
		code := submatches[2]

		if language == "" {
			language = "text"
		}

		highlightedHTML := HighlightCodeBlock(code, language)
		return highlightedHTML
	})

	return result
}

func markdownToHTML(markdown string) string {
	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: blackfriday.CommonHTMLFlags,
	})

	md := blackfriday.New(blackfriday.WithRenderer(renderer), blackfriday.WithExtensions(blackfriday.CommonExtensions|blackfriday.HardLineBreak))

	var buf bytes.Buffer
	ast := md.Parse([]byte(markdown))
	ast.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		return renderer.RenderNode(&buf, node, entering)
	})

	return buf.String()
}
