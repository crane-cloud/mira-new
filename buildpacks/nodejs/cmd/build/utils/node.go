package utils

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ulikunitz/xz"
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

// DownloadNode downloads the Node.js LTS tarball for linux-x64
func DownloadNode(destinationDir string) (string, error) {
	version, err := GetLatestLTSVersion()
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://nodejs.org/dist/%s/node-%s-linux-x64.tar.xz", version, version)
	filename := filepath.Join(destinationDir, fmt.Sprintf("node-%s-linux-x64.tar.xz", version))

	// Create destination file
	out, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Download file
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download Node.js tarball: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch %s: %s", url, resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filename, nil
}

// ExtractTarXz extracts a .tar.xz file into targetDir and strips the top-level folder
func ExtractTarXz(tarballPath, targetDir string) error {
	// Open the tar.xz file
	f, err := os.Open(tarballPath)
	if err != nil {
		return fmt.Errorf("failed to open tarball: %w", err)
	}
	defer f.Close()

	// Create xz reader
	xzr, err := xz.NewReader(f)
	if err != nil {
		return fmt.Errorf("failed to create xz reader: %w", err)
	}

	// Create tar reader
	tr := tar.NewReader(xzr)

	// Iterate through files in tar
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("error reading tar: %w", err)
		}

		// Strip first path component (simulate --strip-components=1)
		parts := strings.SplitN(hdr.Name, string(os.PathSeparator), 2)
		if len(parts) < 2 {
			// skip top-level dir entry
			continue
		}
		relPath := parts[1]
		destPath := filepath.Join(targetDir, relPath)

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(destPath, os.FileMode(hdr.Mode)); err != nil {
				return fmt.Errorf("failed to create dir: %w", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent dir: %w", err)
			}
			outFile, err := os.Create(destPath)
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to write file: %w", err)
			}
			outFile.Close()
			if err := os.Chmod(destPath, os.FileMode(hdr.Mode)); err != nil {
				return fmt.Errorf("failed to set file mode: %w", err)
			}
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent dir for symlink: %w", err)
			}
			if err := os.Symlink(hdr.Linkname, destPath); err != nil {
				return fmt.Errorf("failed to create symlink: %w", err)
			}
		default:
			// ignore other file types
		}
	}

	return nil
}

func InstallNode() error {
	// Download Node.js
	tarballPath, err := DownloadNode(CNB_LAYERS_DIR)
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
