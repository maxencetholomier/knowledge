package prompt

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

func Confirm(question string) (bool, error) {
	fmt.Printf("%s (y/N): ", question)

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return false, fmt.Errorf("failed to set terminal raw mode: %w", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	buf := make([]byte, 1)
	if _, err := os.Stdin.Read(buf); err != nil {
		return false, fmt.Errorf("failed to read input: %w", err)
	}

	ch := buf[0]
	fmt.Printf("%s\r\n", string(ch))

	return ch == 'y' || ch == 'Y', nil
}
