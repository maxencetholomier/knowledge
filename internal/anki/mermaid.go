package anki

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
)

func ConvertMermaidToPNG(mermaidCode string) ([]byte, string, error) {
	if _, err := exec.LookPath("mmdc"); err != nil {
		return nil, "", fmt.Errorf("mmdc not found in PATH (install with: npm install -g @mermaid-js/mermaid-cli): %w", err)
	}

	hash := sha256.Sum256([]byte(mermaidCode))
	filename := fmt.Sprintf("mermaid_%x.png", hash[:8])

	tmpIn, err := os.CreateTemp("", "mermaid_*.mmd")
	if err != nil {
		return nil, "", fmt.Errorf("failed to create temp input file: %w", err)
	}
	defer os.Remove(tmpIn.Name())

	if _, err := tmpIn.WriteString(mermaidCode); err != nil {
		tmpIn.Close()
		return nil, "", fmt.Errorf("failed to write mermaid code: %w", err)
	}
	tmpIn.Close()

	tmpOut, err := os.CreateTemp("", "mermaid_*.png")
	if err != nil {
		return nil, "", fmt.Errorf("failed to create temp output file: %w", err)
	}
	outPath := tmpOut.Name()
	tmpOut.Close()
	defer os.Remove(outPath)

	cmd := exec.Command("mmdc", "-i", tmpIn.Name(), "-o", outPath, "--backgroundColor", "white")
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, "", fmt.Errorf("mmdc conversion failed: %w\nOutput: %s", err, output)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read generated PNG: %w", err)
	}

	return data, filename, nil
}
