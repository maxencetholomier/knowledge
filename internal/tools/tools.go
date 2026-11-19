package tools

import (
	"fmt"
	"os/exec"
)

func Require(name string) error {
	if _, err := exec.LookPath(name); err != nil {
		return fmt.Errorf("%s is required for this feature but is not installed", name)
	}
	return nil
}
