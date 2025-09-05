package main

import (
	"fmt"
	"os"
)

func main() {
	if DetectNodeJSApp() {
		fmt.Println("Detected Node.js application.")
		os.Exit(0) // Exit with success code
	} else {
		fmt.Println("No Node.js application detected.")
		os.Exit(1) // Exit with failure code
	}
}

// Checks if the codebase is a Node.js application by looking for a package.json file.
func DetectNodeJSApp() bool {
	// Check if the package.json file exists in the current directory
	return fileExists("package.json")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true // file exists
	}
	if os.IsNotExist(err) {
		return false // file does not exist
	}
	return false // some other error (e.g., permission denied)
}
