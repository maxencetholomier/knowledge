package anki

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/maxencetholomier/knowledge/internal/tools"
	"os"
	"os/exec"
	"time"
)

const mermaidRenderTimeout = 30 * time.Second

const noSandboxPuppeteerConfig = `{"args":["--no-sandbox"]}`

func isBrowserLaunchFailure(output []byte) bool {
	return bytes.Contains(output, []byte("Failed to launch the browser process"))
}

func ConvertMermaidToPNG(mermaidCode string) ([]byte, string, error) {
	if err := tools.Require("mmdc"); err != nil {
		return nil, "", err
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

	run := func(extraArgs ...string) ([]byte, error) {
		ctx, cancel := context.WithTimeout(context.Background(), mermaidRenderTimeout)
		defer cancel()

		args := append([]string{"-i", tmpIn.Name(), "-o", outPath, "--backgroundColor", "white"}, extraArgs...)
		return exec.CommandContext(ctx, "mmdc", args...).CombinedOutput()
	}

	output, err := run()
	if err != nil && isBrowserLaunchFailure(output) {
		configPath, configErr := writeTempPuppeteerConfig()
		if configErr != nil {
			return nil, "", configErr
		}
		defer os.Remove(configPath)

		output, err = run("-p", configPath)
	}
	if err != nil {
		return nil, "", fmt.Errorf("mmdc conversion failed: %w\nOutput: %s", err, output)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read generated PNG: %w", err)
	}

	return data, filename, nil
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
