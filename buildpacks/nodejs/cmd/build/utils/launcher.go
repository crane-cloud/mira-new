package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var LAUNCH_TOML = []byte(`[types]
launch = true
`)

var BUILD_TOML = []byte(`[types]
build = true

[env]
PATH = "bin"
`)

func CommandToJSONString(cmd string) string {
	parts := strings.Fields(cmd)

	quoted := make([]string, len(parts))
	for i, p := range parts {
		quoted[i] = `"` + p + `"`
	}

	return "[" + strings.Join(quoted, ", ") + "]"
}

func GenerateLauncherFileContents() []byte {
	command := GetStartCommand(APP_TYPE)
	commandString := CommandToJSONString(command)

	var contents = `
[[processes]]
type = "web"
command = ` + commandString + `
default = true
	`
	return []byte(contents)
}

// RunCommand runs a shell command in the current directory.
// Example: "npm run build"
func RunCommand(command string) error {
	// Split the command into args
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = "." // current dir

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command: %w", err)
	}

	return nil
}

func GetStartCommand(app_type string) string {
	if APP_TYPE == "ssr" {
		return SSR_START_COMMAND
	} else {
		return CADDY_START_COMMAND
	}
}
