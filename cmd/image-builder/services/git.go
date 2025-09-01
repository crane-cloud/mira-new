package services

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	common "mira/cmd/common"
	"mira/cmd/image-builder/models"
	fileUtils "mira/cmd/utils"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-resty/resty/v2"
)

// GitService handles git operations and file downloads
type GitService struct{}

// NewGitService creates a new git service
func NewGitService() *GitService {
	return &GitService{}
}

// CloneRepository clones a git repository using improved logic from utils
func (g *GitService) CloneRepository(buildSpec *models.BuildSpec, logger common.Logger) (string, error) {
	destPath := "/usr/local/crane/git/" + buildSpec.Name

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Remove existing directory if it exists
	if _, err := os.Stat(destPath); err == nil {
		if err := os.RemoveAll(destPath); err != nil {
			return "", fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}

	fmt.Println("Cloning git repository")
	_, err := git.PlainClone(destPath, false, &git.CloneOptions{
		URL: buildSpec.Spec.Source.GitRepo.URL,
		Auth: &http.BasicAuth{
			Username: buildSpec.Spec.Source.GitRepo.Username,
			Password: buildSpec.Spec.Source.GitRepo.Password,
		},
	})
	if err != nil {
		return "", fmt.Errorf("error cloning git repository: %w", err)
	}

	// Detect frameworks using existing utility
	frameworks, err := fileUtils.DetectJavaScriptFrameworksLocal(destPath)
	if err != nil {
		logger.InfoWithStep("clone", "Framework detection failed, using defaults")
	} else {
		logger.InfoWithStep("clone", fmt.Sprintf("Detected frameworks: %v", frameworks))
	}

	return destPath, nil
}

// HandleFileSource handles file download and extraction
func (g *GitService) HandleFileSource(buildSpec *models.BuildSpec, logger common.Logger) (string, error) {
	// Download the file
	fmt.Println("Downloading file")
	err := g.downloadFile(buildSpec)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	fmt.Println("Downloaded file")

	// Unzip the file
	fmt.Println("Unzipping file")
	destPath, err := g.unzipFile(buildSpec)
	if err != nil {
		return "", fmt.Errorf("unzip failed: %w", err)
	}
	fmt.Println("Unzipped file")

	return destPath, nil
}

// downloadFile downloads a zip file from a URL
func (g *GitService) downloadFile(buildSpec *models.BuildSpec) error {
	client := resty.New()

	// Ensure destination directory exists
	destDir := "/usr/local/crane/blobs"
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	filePath := filepath.Join(destDir, buildSpec.Name+".zip")

	resp, err := client.R().
		SetOutput(filePath).
		Get(buildSpec.Spec.Source.BlobFile.Source)
	if err != nil {
		return fmt.Errorf("error downloading file: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("error downloading file: %v", resp.Status())
	}
	return nil
}

// unzipFile unzips a zip file with better error handling
func (g *GitService) unzipFile(buildSpec *models.BuildSpec) (string, error) {
	saveto := "/usr/local/crane/zip/" + buildSpec.Name
	zipFile := "/usr/local/crane/blobs/" + buildSpec.Name + ".zip"

	// Ensure destination directory exists
	if err := os.MkdirAll(saveto, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return "", fmt.Errorf("error opening zip file: %w", err)
	}
	defer zipReader.Close()

	for _, file := range zipReader.File {
		destPath := filepath.Join(saveto, file.Name)

		// Security check: prevent zip slip
		if !isPathSafe(destPath, saveto) {
			continue
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, file.Mode()); err != nil {
				return "", fmt.Errorf("error creating directory: %w", err)
			}
			continue
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
			return "", fmt.Errorf("error creating parent directory: %w", err)
		}

		destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return "", fmt.Errorf("error creating file: %w", err)
		}

		srcFile, err := file.Open()
		if err != nil {
			destFile.Close()
			return "", fmt.Errorf("error opening file: %w", err)
		}

		_, err = io.Copy(destFile, srcFile)
		srcFile.Close()
		destFile.Close()

		if err != nil {
			return "", fmt.Errorf("error copying file: %w", err)
		}
	}

	return saveto, nil
}

// isPathSafe checks if the path is safe (prevents zip slip attacks)
func isPathSafe(destPath, baseDir string) bool {
	cleanBase := filepath.Clean(baseDir)
	cleanDest := filepath.Clean(destPath)
	rel, err := filepath.Rel(cleanBase, cleanDest)
	if err != nil {
		return false
	}
	return !filepath.IsAbs(rel) && !filepath.HasPrefix(rel, "..")
}
