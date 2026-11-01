package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Config struct {
	FZF                bool     `json:"fzf"`
	FZFEnv             string   `json:"fzfEnv"`
	FZFOptions         string   `json:"fzfOption"`
	FZFTmuxOptions     string   `json:"fzfTmuxOption"`
	ScreenshotTool     string   `json:"screenshotTool"`
	ScreenshotParams   []string `json:"screenshotParams"`
	ScreenrecordTool   string   `json:"screenrecordTool"`
	ScreenrecordParams []string `json:"screenrecordParams"`
	SchemaTool         string   `json:"schemaTool"`
	SchemaToolParams   []string `json:"schemaToolParams"`
	Terminal           string   `json:"terminal"`
	TerminalParams     []string `json:"terminalParams"`
	JoplinConfigFile   string   `json:"joplinConfigFile"`
	DirExport          string   `json:"dirExport"`
	JoplinNotebook     string   `json:"joplinNotebook"`
}

var Current Config

var CacheFile = filepath.Join(os.TempDir(), "kl-search-cache")

var joplinTokenPattern = regexp.MustCompile(`\s*"api.token":\s*"(.*)"`)

func Load() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to read home directory: %w", err)
	}

	data, err := os.ReadFile(filepath.Join(home, ".config", "kl", "config.json"))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("failed to read kl config file: %w", err)
	}
	if err == nil {
		if err := json.Unmarshal(data, &Current); err != nil {
			return fmt.Errorf("failed to parse kl config file: %w", err)
		}
	}

	applyDefaults(&Current)
	initCacheFile()
	return nil
}

func initCacheFile() {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Join(cacheDir, "kl"), 0755); err != nil {
		return
	}
	CacheFile = filepath.Join(cacheDir, "kl", "search-cache")
}

func applyDefaults(c *Config) {
	setDefault(&c.FZFEnv, "fzf")
	setDefault(&c.ScreenshotTool, "flameshot")
	setDefault(&c.SchemaTool, "inkscape")
	setDefault(&c.Terminal, "x-terminal-emulator")
	setDefault(&c.DirExport, "/tmp/kl-export/")
	if len(c.ScreenshotParams) == 0 {
		c.ScreenshotParams = []string{"gui", "--path"}
	}
	if len(c.SchemaToolParams) == 0 {
		c.SchemaToolParams = []string{"--export-filename"}
	}
	if len(c.TerminalParams) == 0 {
		c.TerminalParams = []string{"--execute"}
	}
}

func setDefault(field *string, value string) {
	if *field == "" {
		*field = value
	}
}

func Editor() string {
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	return "vi"
}

func JoplinToken() (string, error) {
	file, err := joplinConfigFile()
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("failed to read Joplin config file: %w", err)
	}

	matches := joplinTokenPattern.FindStringSubmatch(string(content))
	if len(matches) < 2 {
		return "", fmt.Errorf("api.token not found in %s", file)
	}
	return matches[1], nil
}

func joplinConfigFile() (string, error) {
	if Current.JoplinConfigFile != "" {
		return Current.JoplinConfigFile, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	candidates := []string{
		filepath.Join(home, ".config/joplin-desktop/settings.json"),
		filepath.Join(home, "snap/joplin-desktop/current/.config/joplin-desktop/settings.json"),
		filepath.Join(home, ".var/app/net.cozic.joplin_desktop/config/joplin-desktop/settings.json"),
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("joplin config file not found (tried: %s)", strings.Join(candidates, ", "))
}
