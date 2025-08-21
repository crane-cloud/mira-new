package main

import (
	"fmt"
	"log"
	"mira_buildpack/nodejs/cmd/build/utils"
)

func main() {
	fmt.Println("Starting Node.js buildpack setup...")

	fmt.Println("Downloading Node.js LTS version...")
	// Install Node.js LTS version
	err := utils.InstallNode()
	if err != nil {
		log.Fatalf("Failed to install Node.js: %v", err)
	}

	// Write the nodejs.toml file
	err = utils.WriteFile(utils.NODE_JS_LAUNCH_FILE, utils.LAUNCH_TOML)
	if err != nil {
		log.Fatalf("Failed to write launch file: %v", err)
	}

	// Write the launcher file
	var launcherContents = utils.GenerateLauncherFileContents(utils.START_COMMAND)
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
