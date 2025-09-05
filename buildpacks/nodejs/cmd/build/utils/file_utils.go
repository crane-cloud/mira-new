package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ulikunitz/xz"
)

func DownloadFile(
	fileURL string,
	destination string) (string, error) {

	// Create destination file
	out, err := os.Create(destination)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Download file
	resp, err := http.Get(fileURL)
	if err != nil {
		return "", fmt.Errorf("failed to download Node.js tarball: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch %s: %s", fileURL, resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return destination, nil
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

// ExtractTarGz extracts a .tar.gz archive into the targetDir.
func ExtractTarGz(tarballPath, targetDir string) error {
	// Open the tarball file
	f, err := os.Open(tarballPath)
	if err != nil {
		return fmt.Errorf("failed to open tarball: %w", err)
	}
	defer f.Close()

	// Create a gzip reader
	gzr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	// Create a tar reader
	tr := tar.NewReader(gzr)

	// Iterate through the files in the tar archive
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // end of archive
		}
		if err != nil {
			return fmt.Errorf("failed to read tar entry: %w", err)
		}

		// Determine the target path
		targetPath := filepath.Join(targetDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}

		case tar.TypeReg:
			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent dir: %w", err)
			}

			// Create the file
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}

			// Copy file contents
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to write file: %w", err)
			}
			outFile.Close()
		default:
			// Skip other file types
		}
	}

	return nil
}
