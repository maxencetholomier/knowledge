package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func Confirm(question string) (bool, error) {
	fmt.Printf("%s (y/N): ", question)

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read input: %w", err)
	}

	line = strings.TrimSpace(line)
	return line == "y" || line == "Y", nil
}
