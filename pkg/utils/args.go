package utils

import (
	"fmt"
	"kl/pkg/output"
	"os"
	"strconv"
	"strings"
)

func ResolveFileName(args []string, dirCache string) (string, error) {
	if len(args) > 1 {
		return "", fmt.Errorf("too many arguments. Maximum is one.")
	}

	readFromCache := false
	var lineNum int = 1
	var fileName string

	if len(args) == 1 {
		trimmed := strings.TrimSuffix(args[0], ".md")
		if len(trimmed) == 14 {
			fileName = trimmed
		} else {
			lineNumber, err := strconv.Atoi(args[0])
			if err != nil {
				return "", fmt.Errorf("invalid line number: %s", args[0])
			}
			readFromCache = true
			lineNum = lineNumber
		}
	} else {
		readFromCache = true
	}

	if readFromCache {
		_, err := os.Stat(dirCache)
		if err != nil {
			return "", fmt.Errorf("cache file does not exist. Run 'kl list' first or provide a 14-character timestamp")
		}
		line, err := output.GetLineFromCache(dirCache, lineNum)
		if err != nil {
			return "", err
		}

		fmt.Println("line:" + line)
		fileName = strings.Split(line, " ")[1]
	}

	return fileName, nil
}