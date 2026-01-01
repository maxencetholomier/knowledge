package output

import (
	"bufio"
	"fmt"
	"kl/pkg/config"
	"os"
	"strconv"
	"strings"
)

func PrintToStdout(line string, lineNumber int) error {
	fmt.Println(strconv.Itoa(lineNumber) + " " + line)
	return nil
}

func PrintCache() error {
	file, err := os.Open(config.DirCache)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
	}

	err = scanner.Err()
	if err != nil {
		return err
	}

	return nil
}

func PrintToCache(text string) error {
	file, err := os.OpenFile(config.DirCache, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)

	if text == "" {
		formattedLine := fmt.Sprintf("")
		_, err = writer.WriteString(formattedLine)
		if err != nil {
			return err
		}
	} else {
		lines := strings.Split(text, "\n")
		for i, line := range lines {
			formattedLine := fmt.Sprintf("%d %s\n", i+1, line)
			_, err := writer.WriteString(formattedLine)
			if err != nil {
				return err
			}
		}
	}
	writer.Flush()
	return nil
}

func GetLineFromCache(filename string, lineNumber int) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentLine := 1
	for scanner.Scan() {
		if currentLine == lineNumber {
			return scanner.Text(), nil
		}
		currentLine++
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("line %d does not exist in file", lineNumber)
}
