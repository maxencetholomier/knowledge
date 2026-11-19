package anki

import (
	"testing"

	"github.com/maxencetholomier/knowledge/internal/tools"
)

func TestConvertMermaidToPNG(t *testing.T) {
	if err := tools.Require("mmdc"); err != nil {
		t.Skip("mmdc is not installed")
	}

	data, filename, err := ConvertMermaidToPNG(sandboxProbeDiagram)
	if err != nil {
		t.Fatalf("ConvertMermaidToPNG() = %v, want nil", err)
	}

	if len(data) < 4 || string(data[1:4]) != "PNG" {
		t.Errorf("got %d bytes that are not a PNG", len(data))
	}

	if filename != "mermaid_917423685aa27a50.png" {
		t.Errorf("filename = %q, want a deterministic hash of the diagram", filename)
	}
}
