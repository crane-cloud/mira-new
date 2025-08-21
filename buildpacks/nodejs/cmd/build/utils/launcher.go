package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var LAUNCH_TOML = []byte(`[types]
launch = true
`)

func WriteFile(path string, content []byte) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create parent dirs: %w", err)
	}

	// Write content to file
	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func CommandToJSONString(cmd string) string {
	parts := strings.Fields(cmd)

	quoted := make([]string, len(parts))
	for i, p := range parts {
		quoted[i] = `"` + p + `"`
	}

	return "[" + strings.Join(quoted, ", ") + "]"
}

func GenerateLauncherFileContents(command string) []byte {
	commandString := CommandToJSONString(command)

	var contents = `
[[processes]]
type = "web"
command = ` + commandString + `
default = true
	`
	return []byte(contents)
}
