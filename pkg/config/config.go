package config

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
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

const (
	DirCache string = "/tmp/kl-search-cache"
)

var joplinToken string
var homeDir string
var FileJoplinConfig string
var editor string
var fzf bool
var fzfOptions string
var fzfTmuxOptions string
var fzfEnv string
var screenshotTool string
var screenshotParams []string
var screenrecordTool string
var screenrecordParams []string
var schemaTool string
var schemaToolParams []string
var terminal string
var terminalParams []string
var dirExport string
var joplinNotebook string
var initError error

func GetJoplinToken() (string, error) {
	if initError != nil {
		return "", initError
	}
	return joplinToken, nil
}

func GetEditor() string {
	return editor
}

func GetFzf() bool {
	return fzf
}

func GetFzfEnv() string {
	return fzfEnv
}

func GetFzfOptions() string {
	return fzfOptions
}

func GetFzfTmuxOptions() string {
	return fzfTmuxOptions
}

func GetScreenshotTool() string {
	if screenshotTool == "" {
		return "flameshot"
	}
	return screenshotTool
}

func GetScreenshotParams() []string {
	if len(screenshotParams) == 0 {
		return []string{"gui", "--path"}
	}
	return screenshotParams
}

func GetScreenrecordTool() string {
	return screenrecordTool
}

func GetScreenrecordParams() []string {
	return screenrecordParams
}

func GetSchemaTool() string {
	if schemaTool == "" {
		return "inkscape"
	}
	return schemaTool
}

func GetSchemaToolParams() []string {
	if len(schemaToolParams) == 0 {
		return []string{"--export-filename"}
	}
	return schemaToolParams
}

func GetTerminal() string {
	if terminal == "" {
		return "x-terminal-emulator"
	}
	return terminal
}

func GetTerminalForVideo() string {
	if terminal == "" {
		return "kitty"
	}
	return terminal
}

func GetTerminalParams() []string {
	if len(terminalParams) == 0 {
		return []string{"--execute"}
	}
	return terminalParams
}

func GetDirExport() string {
	if dirExport == "" {
		return "/tmp/kl-export/"
	}
	return dirExport
}

func GetJoplinNotebook() string {
	return joplinNotebook
}

func CheckInitError() error {
	return initError
}

func initHomeDir() error {
	dir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to read home directory: %w", err)
	}
	homeDir = dir
	return nil
}

func initEditor() {
	editor = os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
}

func loadConfigFile() (*Config, error) {
	filePath := homeDir + "/.config/kl/config.json"

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read k config file: %w", err)
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse k config file: %w", err)
	}

	return &config, nil
}

// TODO: ADD  joplinNotebook in environment configuration
func setConfigVars(config *Config) {
	fzf = config.FZF
	fzfEnv = config.FZFEnv
	fzfOptions = config.FZFOptions
	fzfTmuxOptions = config.FZFTmuxOptions
	screenshotTool = config.ScreenshotTool
	screenshotParams = config.ScreenshotParams
	screenrecordTool = config.ScreenrecordTool
	screenrecordParams = config.ScreenrecordParams
	schemaTool = config.SchemaTool
	schemaToolParams = config.SchemaToolParams
	terminal = config.Terminal
	terminalParams = config.TerminalParams
	dirExport = config.DirExport
	joplinNotebook = config.JoplinNotebook
}

func initJoplin(config *Config) error {
	if config.JoplinConfigFile != "" {
		FileJoplinConfig = config.JoplinConfigFile
	} else {
		FileJoplinConfig = homeDir + "/.config/joplin-desktop/settings.json"
	}

	configFile, err := os.ReadFile(FileJoplinConfig)
	if err != nil {
		return fmt.Errorf("failed to read Joplin config file: %w", err)
	}

	re := regexp.MustCompile(`\s*"api.token":\s*"(.*)"`)
	matches := re.FindStringSubmatch(string(configFile))

	if len(matches) > 0 {
		joplinToken = matches[1]
	} else {
		return fmt.Errorf("api-token not found in joplin config file")
	}
	return nil
}

func init() {
	if err := initHomeDir(); err != nil {
		initError = err
		return
	}

	config, err := loadConfigFile()
	if err != nil {
		initError = err
		return
	}

	setConfigVars(config)
	initEditor()

	if err := initJoplin(config); err != nil {
		initError = err
		return
	}
}
