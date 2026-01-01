package ui

import (
	"fmt"
	"kl/pkg/files"
	"os/exec"
	"strings"
)

type FzfConfig struct {
	Command    string
	Options    []string
	Preview    string
	Dimensions string
	BaseDir    string
}

func NewFzfConfig() *FzfConfig {
	return &FzfConfig{
		Command:    "fzf",
		Options:    []string{"--ansi"},
		Dimensions: "",
	}
}

func NewFzfConfigWithCommand(command string) *FzfConfig {
	return &FzfConfig{
		Command:    command,
		Options:    []string{"--ansi"},
		Dimensions: "",
	}
}

func (f *FzfConfig) WithCommand(cmd string) *FzfConfig {
	f.Command = cmd
	return f
}

func (f *FzfConfig) WithOptions(opts ...string) *FzfConfig {
	f.Options = append(f.Options, opts...)
	return f
}

func (f *FzfConfig) WithPreview(preview string) *FzfConfig {
	f.Preview = preview
	return f
}

func (f *FzfConfig) WithDimensions(dims string) *FzfConfig {
	f.Dimensions = dims
	return f
}

func (f *FzfConfig) WithBaseDir(baseDir string) *FzfConfig {
	f.BaseDir = baseDir
	return f
}

type FileSelector struct {
	Config *FzfConfig
}

func NewFileSelector(config *FzfConfig) *FileSelector {
	if config == nil {
		config = NewFzfConfig()
	}
	return &FileSelector{Config: config}
}

func (fs *FileSelector) SelectFile(fileList []files.FileInfo, baseDir string) (string, error) {
	if len(fileList) == 0 {
		return "", fmt.Errorf("no files to select from")
	}

	args := []string{}

	if fs.Config.Dimensions != "" && strings.Contains(fs.Config.Command, "tmux") {
		args = append(args, fs.Config.Dimensions)
	}

	args = append(args, fs.Config.Options...)

	if fs.Config.Preview != "" {
		args = append(args, "--preview", fs.Config.Preview)
	} else {
		args = append(args, "--delimiter", "\t")
		args = append(args, "--with-nth", "1")
		args = append(args, "--preview", fmt.Sprintf("bat --plain --theme=${BAT_THEME} --color=always %s/{2}.md 2>/dev/null || bat --plain --theme=${BAT_THEME} --color=always %s/{1}.md", baseDir, baseDir))
	}

	cmd := exec.Command(fs.Config.Command, args...)

	var input strings.Builder
	for _, file := range fileList {
		title, err := file.GetTitle()
		if err != nil {
			title = file.Name
		}

		if strings.HasPrefix(title, "#") {
			title = strings.TrimPrefix(title, "#")
			title = strings.TrimSpace(title)
		}
		
		if strings.Contains(title, "|") {
			parts := strings.Split(title, "|")
			title = strings.TrimSpace(parts[0])
		}

		timestamp := strings.TrimSuffix(file.Name, ".md")

		if title != "" {
			fmt.Fprintf(&input, "%s\t%s\n", title, timestamp)
		} else {
			fmt.Fprintf(&input, "%s\n", timestamp)
		}
	}

	cmd.Stdin = strings.NewReader(input.String())

	output, err := cmd.Output()
	if err != nil {
		return "", nil
	}

	selected := strings.TrimSpace(string(output))
	if selected == "" {
		return "", nil
	}

	var timestamp string
	if strings.Contains(selected, "\t") {
		parts := strings.Split(selected, "\t")
		if len(parts) >= 2 {
			timestamp = strings.TrimSpace(parts[len(parts)-1])
		} else {
			return "", fmt.Errorf("invalid selection format")
		}
	} else {
		timestamp = selected
	}

	if timestamp == "" {
		return "", fmt.Errorf("no timestamp found in selection")
	}

	return fmt.Sprintf("%s/%s.md", baseDir, timestamp), nil
}

type SearchSelector struct {
	Config *FzfConfig
}

func NewSearchSelector(config *FzfConfig) *SearchSelector {
	if config == nil {
		config = NewFzfConfig()
	}
	return &SearchSelector{Config: config}
}

func (ss *SearchSelector) SelectFromSearchResults(rgOutput []byte) (string, error) {
	if len(rgOutput) == 0 {
		fmt.Println("No matches found")
		return "", nil
	}

	args := []string{}

	if ss.Config.Dimensions != "" && strings.Contains(ss.Config.Command, "tmux") {
		args = append(args, ss.Config.Dimensions)
	}

	args = append(args, "--ansi")
	args = append(args, "--delimiter", ":")
	if ss.Config.BaseDir != "" {
		args = append(args, "--preview", fmt.Sprintf("bat --plain --color=always %s/{1} --highlight-line {2}", ss.Config.BaseDir))
	} else {
		args = append(args, "--preview", "bat --plain --color=always {1} --highlight-line {2}")
	}

	args = append(args, ss.Config.Options...)

	cmd := exec.Command(ss.Config.Command, args...)
	cmd.Stdin = strings.NewReader(string(rgOutput))

	output, err := cmd.Output()
	if err != nil {
		return "", nil
	}

	selected := strings.TrimSpace(string(output))
	if selected == "" {
		return "", nil
	}

	parts := strings.SplitN(selected, ":", 3)
	if len(parts) < 1 {
		return "", fmt.Errorf("invalid selection format")
	}

	return parts[0], nil
}

func ParseOptions(optionsStr string) ([]string, error) {
	var args []string
	var current strings.Builder
	var inQuotes bool
	var quoteChar rune

	for _, char := range optionsStr {
		switch {
		case char == ' ' && !inQuotes:
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		case char == '\'' || char == '"':
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
			} else {
				current.WriteRune(char)
			}
		default:
			current.WriteRune(char)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args, nil
}
