package anki

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
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

	back = detectAndConvertTables(back)
	backHTML, mermaidMedia := markdownToHTML(back)
	mediaFiles = append(mediaFiles, mermaidMedia...)

	noteFilename := strings.TrimSuffix(filepath.Base(notePath), ".md")

	csum := int64(0)
	for _, c := range front {
		csum = (csum + int64(c)) % 0xffff
	}

	note := &genanki.Note{
		ID:        generateDeterministicID(noteFilename),
		ModelID:   BasicModelID,
		Fields:    []string{front, backHTML},
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

func detectAndConvertTables(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	i := 0

	for i < len(lines) {
		line := lines[i]

		if isTableRow(line) && !isSeparatorRow(line) {
			tableStart := i
			tableEnd := i

			for j := i; j < len(lines); j++ {
				if isTableRow(lines[j]) || isSeparatorRow(lines[j]) {
					tableEnd = j
				} else if strings.TrimSpace(lines[j]) == "" && j+1 < len(lines) && isTableRow(lines[j+1]) {
					tableEnd = j
				} else {
					break
				}
			}

			hasSeparator := false
			for j := tableStart; j <= tableEnd; j++ {
				if isSeparatorRow(lines[j]) {
					hasSeparator = true
					break
				}
			}

			if !hasSeparator && tableEnd > tableStart {
				result = append(result, lines[tableStart])
				result = append(result, generateSeparatorRow(lines[tableStart]))
				for j := tableStart + 1; j <= tableEnd; j++ {
					result = append(result, lines[j])
				}
				i = tableEnd + 1
				continue
			}
		}

		result = append(result, line)
		i++
	}

	return strings.Join(result, "\n")
}

func isTableRow(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}
	if !strings.Contains(trimmed, "|") {
		return false
	}
	if isSeparatorRow(trimmed) {
		return false
	}
	return true
}

func isSeparatorRow(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}
	cleaned := strings.ReplaceAll(trimmed, "|", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, ":", "")
	cleaned = strings.TrimSpace(cleaned)
	return cleaned == "" && strings.Contains(trimmed, "-")
}

func generateSeparatorRow(headerRow string) string {
	parts := strings.Split(headerRow, "|")
	columnCount := 0

	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			columnCount++
		}
	}

	var separators []string
	for i := 0; i < columnCount; i++ {
		separators = append(separators, "---")
	}

	return "| " + strings.Join(separators, " | ") + " |"
}

func markdownToHTML(markdown string) (string, []MediaFile) {
	renderer := &chromaRenderer{
		HTMLRenderer: blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
			Flags: blackfriday.CommonHTMLFlags,
		}),
	}

	md := blackfriday.New(blackfriday.WithRenderer(renderer), blackfriday.WithExtensions(blackfriday.CommonExtensions|blackfriday.HardLineBreak))

	var buf bytes.Buffer
	ast := md.Parse([]byte(markdown))
	ast.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		return renderer.RenderNode(&buf, node, entering)
	})

	return buf.String(), renderer.mediaFiles
}

type chromaRenderer struct {
	*blackfriday.HTMLRenderer
	mediaFiles []MediaFile
}

func (r *chromaRenderer) RenderNode(w io.Writer, node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
	if node.Type == blackfriday.CodeBlock {
		language := strings.TrimSpace(string(node.Info))
		if language == "mermaid" {
			data, filename, err := ConvertMermaidToPNG(string(node.Literal))
			if err != nil {
				fmt.Fprintf(w, `<pre><code>%s</code></pre>`, node.Literal)
			} else {
				r.mediaFiles = append(r.mediaFiles, MediaFile{
					Filename: filename,
					Data:     data,
				})
				fmt.Fprintf(w, `<img src="%s" alt="Mermaid diagram">`, filename)
			}
			return blackfriday.GoToNext
		}
		if language == "" {
			language = "text"
		}
		highlighted := HighlightCodeBlock(string(node.Literal), language)
		w.Write([]byte(highlighted))
		return blackfriday.GoToNext
	}
	return r.HTMLRenderer.RenderNode(w, node, entering)
}
