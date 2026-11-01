package files

import (
	"fmt"
	"github.com/maxencetholomier/knowledge/internal/config"
	"github.com/maxencetholomier/knowledge/internal/utils"
	"os"
	"os/exec"
	"strings"
	"time"
)

func Edit(filePath string) {
	editor := config.Editor()
	vim := exec.Command(editor, filePath)

	vim.Stdin = os.Stdin
	vim.Stdout = os.Stdout
	vim.Stderr = os.Stderr

	vim.Run()
}

func Create(filePath, fileContent string) error {
	return os.WriteFile(filePath, []byte(fileContent), 0644)
}

func GetLastUpdates(DirZet string) (map[string]time.Time, error) {
	out, err := exec.Command(
		"git", "-C", DirZet, "log", "--format=COMMIT %ad", "--date=unix", "--name-only", "--relative",
	).Output()
	if err != nil {
		return nil, err
	}

	result := make(map[string]time.Time)
	var current time.Time

	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "COMMIT ") {
			var unix int64
			fmt.Sscanf(line[7:], "%d", &unix)
			current = time.Unix(unix, 0)
		} else if line != "" {
			if _, seen := result[line]; !seen {
				result[line] = current
			}
		}
	}

	return result, nil
}

func GetFileType(filename string) string {
	if filename == "" {
		return filename
	}

	if strings.Contains(filename, ".") {
		return strings.ToLower(filename[strings.LastIndex(filename, ".")+1:])
	}

	return ""

}

func CreateMediaNote(
	dirZet, tool string, toolParams []string, terminal string, terminalParams []string,
	fileExtension, templateType string, useNewTerminal bool,
) error {
	now := time.Now()
	timestamp := now.Format("20060102150405")

	mediaArgs := append(toolParams, dirZet+"/"+timestamp+fileExtension)

	if useNewTerminal {
		terminalArgs := append(terminalParams, tool)
		terminalArgs = append(terminalArgs, mediaArgs...)
		_, err := exec.Command(terminal, terminalArgs...).Output()
		if err != nil {
			return err
		}
	} else {
		_, err := exec.Command(tool, mediaArgs...).Output()
		if err != nil {
			return err
		}
	}

	fileName := timestamp + ".md"
	template := utils.CreateTemplate(timestamp, templateType)

	if err := Create(dirZet+"/"+fileName, template); err != nil {
		return err
	}

	Edit(dirZet + "/" + fileName)

	return nil
}

func EnsureDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}
	return nil
}
