package anki

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/maxencetholomier/knowledge/internal/tools"
)

const mermaidRenderTimeout = 30 * time.Second

const noSandboxPuppeteerConfig = `{"args":["--no-sandbox"]}`

const sandboxProbeDiagram = "graph TD\n  A-->B\n"

var sandboxUnavailable = sync.OnceValue(func() bool {
	if _, err := renderMermaid(sandboxProbeDiagram); err == nil {
		return false
	}

	configPath, err := writeTempPuppeteerConfig()
	if err != nil {
		return false
	}
	defer os.Remove(configPath)

	_, err = renderMermaid(sandboxProbeDiagram, "-p", configPath)
	return err == nil
})

func ConvertMermaidToPNG(mermaidCode string) ([]byte, string, error) {
	if err := tools.Require("mmdc"); err != nil {
		return nil, "", err
	}

	var extraArgs []string
	if sandboxUnavailable() {
		configPath, err := writeTempPuppeteerConfig()
		if err != nil {
			return nil, "", err
		}
		defer os.Remove(configPath)

		extraArgs = []string{"-p", configPath}
	}

	data, err := renderMermaid(mermaidCode, extraArgs...)
	if err != nil {
		return nil, "", err
	}

	hash := sha256.Sum256([]byte(mermaidCode))
	return data, fmt.Sprintf("mermaid_%x.png", hash[:8]), nil
}

func renderMermaid(mermaidCode string, extraArgs ...string) ([]byte, error) {
	tmpIn, err := os.CreateTemp("", "mermaid_*.mmd")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp input file: %w", err)
	}
	defer os.Remove(tmpIn.Name())

	if _, err := tmpIn.WriteString(mermaidCode); err != nil {
		tmpIn.Close()
		return nil, fmt.Errorf("failed to write mermaid code: %w", err)
	}
	tmpIn.Close()

	tmpOut, err := os.CreateTemp("", "mermaid_*.png")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp output file: %w", err)
	}
	outPath := tmpOut.Name()
	tmpOut.Close()
	defer os.Remove(outPath)

	ctx, cancel := context.WithTimeout(context.Background(), mermaidRenderTimeout)
	defer cancel()

	args := append([]string{"-i", tmpIn.Name(), "-o", outPath, "--backgroundColor", "white"}, extraArgs...)
	if output, err := exec.CommandContext(ctx, "mmdc", args...).CombinedOutput(); err != nil {
		return nil, fmt.Errorf("mmdc conversion failed: %w\nOutput: %s", err, output)
	}

	return os.ReadFile(outPath)
}

func writeTempPuppeteerConfig() (string, error) {
	file, err := os.CreateTemp("", "puppeteer_*.json")
	if err != nil {
		return "", fmt.Errorf("failed to create temp puppeteer config: %w", err)
	}

	if _, err := file.WriteString(noSandboxPuppeteerConfig); err != nil {
		file.Close()
		os.Remove(file.Name())
		return "", fmt.Errorf("failed to write puppeteer config: %w", err)
	}

	if err := file.Close(); err != nil {
		os.Remove(file.Name())
		return "", fmt.Errorf("failed to close puppeteer config: %w", err)
	}

	return file.Name(), nil
}
