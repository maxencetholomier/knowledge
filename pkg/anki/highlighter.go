package anki

import (
	"bytes"
	"html"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

func HighlightCodeBlock(code, language string) string {
	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Fallback
	}

	style := styles.Get("github")
	if style == nil {
		style = styles.Fallback
	}

	formatter := chromahtml.New(
		chromahtml.WithClasses(false),
		chromahtml.TabWidth(4),
		chromahtml.Standalone(false),
	)

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return html.EscapeString(code)
	}

	var buf bytes.Buffer
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return html.EscapeString(code)
	}

	return buf.String()
}
