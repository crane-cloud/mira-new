package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

// NodeVersion represents a Node.js release entry in the index.json
type NodeVersion struct {
	Version string `json:"version"`
	LTS     any    `json:"lts"` // can be false or string
}

// GetLatestLTSVersion fetches the latest LTS version from Node.js index.json
func GetLatestLTSVersion() (string, error) {
	resp, err := http.Get("https://nodejs.org/dist/index.json")
	if err != nil {
		return "", fmt.Errorf("failed to fetch index.json: %w", err)
	}
	defer resp.Body.Close()

	var versions []NodeVersion
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return "", fmt.Errorf("failed to decode index.json: %w", err)
	}

	// Find the first LTS (the list is ordered latest first)
	for _, v := range versions {
		if v.LTS != false {
			return v.Version, nil
		}
	}

	return "", fmt.Errorf("no LTS version found")
}

func InstallNode() error {
	// Get the latest Node.js LTS version
	version, err := GetLatestLTSVersion()
	if err != nil {
		return fmt.Errorf("failed to get latest LTS version: %w", err)
	}

	dest := filepath.Join(CNB_LAYERS_DIR, fmt.Sprintf("node-%s-linux-x64.tar.xz", version))
	nodeURL := fmt.Sprintf("https://nodejs.org/dist/%s/node-%s-linux-x64.tar.xz", version, version)

	// Download Node.js
	tarballPath, err := DownloadFile(nodeURL, dest)
	if err != nil {
		return fmt.Errorf("failed to download Node.js: %w", err)
	}

	// Extract the tarballs
	if err := ExtractTarXz(tarballPath, NODE_JS_LAYERS_DIR); err != nil {
		return fmt.Errorf("failed to extract Node.js tarball: %w", err)
	}

	// Clean up the downloaded tarball
	if err := os.Remove(tarballPath); err != nil {
		return fmt.Errorf("failed to remove tarball: %w", err)
	}

	return nil
}
