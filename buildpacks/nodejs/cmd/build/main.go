package main

import (
	"fmt"
	"log"
	"mira_buildpack/nodejs/cmd/build/utils"
	"os"
)

func main() {
	fmt.Println("Starting Node.js buildpack setup...")

	fmt.Println("Downloading Node.js LTS version...")
	// Install Node.js LTS version
	err := utils.InstallNode()
	if err != nil {
		log.Fatalf("Failed to install Node.js: %v", err)
	}

	// Install Caddy
	err = utils.InstallCaddy()
	if err != nil {
		log.Fatalf("Failed to install Caddy: %v", err)
	}

	// Write caddyfile
	err = utils.WriteFile(utils.CADDY_FILE_PATH, []byte(utils.CADDY_FILE))
	if err != nil {
		log.Fatalf("Failed to write Caddyfile: %v", err)
	}

	// Set PATH globally so all subsequent commands can find node
	fmt.Println("Setting up Node.js in PATH...")
	currentPath := os.Getenv("PATH")
	nodeBinPath := utils.NODE_JS_LAYERS_DIR + "/bin"
	newPath := nodeBinPath + ":" + currentPath
	err = os.Setenv("PATH", newPath)
	if err != nil {
		log.Fatalf("Failed to set PATH: %v", err)
	}

	err = utils.WriteFile(utils.NODE_JS_BUILD_TOML, utils.BUILD_TOML)
	if err != nil {
		log.Fatalf("Failed to write build file: %v", err)
	}

	// Write the nodejs.toml file
	err = utils.WriteFile(utils.NODE_JS_LAUNCH_FILE, utils.LAUNCH_TOML)
	if err != nil {
		log.Fatalf("Failed to write launch file: %v", err)
	}

	// Write the launcher file
	var launcherContents = utils.GenerateLauncherFileContents()
	err = utils.WriteFile(utils.LAUNCHER_FILE, launcherContents)
	if err != nil {
		log.Fatalf("Failed to write launcher file: %v", err)
	}

	fmt.Println("Installing dependencies with npm install...")
	// Run NPM install
	err = utils.RunCommand(utils.NPM_INSTALL_COMMAND)
	if err != nil {
		log.Fatalf("ERROR: `npm install` failed to run")
	}

	fmt.Println("Running build command...")
	// Run the build command
	err = utils.RunCommand(utils.BUILD_COMMAND)
	if err != nil {
		log.Fatalf("Failed to run build command: %v", err)
	}

	fmt.Println("Node.js buildpack setup completed successfully.")

}
