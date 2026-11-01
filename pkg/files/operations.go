package files

import (
	"fmt"
	"github.com/maxencetholomier/knowledge/pkg/config"
	"github.com/maxencetholomier/knowledge/pkg/utils"
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

func Create(filePath, fileContent string) (*os.File, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}

	_, err = file.WriteString(fileContent)
	if err != nil {
		return nil, err
	}
	return file, nil
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

	file, err := Create(dirZet+"/"+fileName, template)
	if err != nil {
		return err
	}
	defer file.Close()

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
