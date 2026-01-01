package files

import (
	"fmt"
	"kl/pkg/config"
	"kl/pkg/utils"
	"os"
	"os/exec"
	"strings"
	"time"
)

func Edit(filePath string) {
	editor := config.GetEditor()
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

func GetLastUpdate(fileName string, DirZet string) (time.Time, error) {
	cmd := exec.Command("git", "-C", DirZet, "log", "-1", "--format=%ad", "--date=unix", fileName)
	var last_update time.Time

	zet_last_update, err := cmd.Output()
	if err != nil {
		return last_update, err
	}

	var unixTimestamp int64
	_, err = fmt.Sscanf(string(zet_last_update), "%d", &unixTimestamp)
	if err != nil {
		return last_update, err
	}

	last_update = time.Unix(unixTimestamp, 0)

	return last_update, nil
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

func CreateMediaNote(dirZet, tool string, toolParams []string, terminal string, terminalParams []string, fileExtension, templateType string, useNewTerminal bool) error {
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
