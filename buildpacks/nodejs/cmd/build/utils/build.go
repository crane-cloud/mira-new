package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RunCommand runs a shell command in the current directory.
// Example: "npm run build"
func RunCommand(command string) error {
	// Split the command into args
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = "." // current dir

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command: %w", err)
	}

	return nil
}
