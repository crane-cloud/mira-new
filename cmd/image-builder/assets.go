package imagebuilder

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-resty/resty/v2"
)

// CloneGitRepo clones a git repository
func CloneGitRepo(app *ImageBuild) error {
	fmt.Println("Cloning git repository")
	_, err := git.PlainClone("/usr/local/crane/git/"+app.Name, false, &git.CloneOptions{
		URL: app.Spec.Source.GitRepo.URL,
		Auth: &http.BasicAuth{
			Username: app.Spec.Source.GitRepo.Username,
			Password: app.Spec.Source.GitRepo.Password,
		},
	})
	if err != nil {
		return fmt.Errorf("error cloning git repository: %v", err)
	}
	fmt.Println("Cloned")

	return nil
}

// DownloadFile downloads a zip file from a URL
func DownloadFile(app *ImageBuild) error {
	client := resty.New()
	resp, err := client.R().
		SetOutput("/usr/local/crane/blobs/" + app.Name + ".zip").
		Get(app.Spec.Source.BlobFile.Source)
	if err != nil {
		return fmt.Errorf("error downloading file: %v", err)
	}
	if resp.IsError() {
		return fmt.Errorf("error downloading file: %v", resp.Status())
	}
	fmt.Println("Downloaded")
	return nil

}

// UnzipFile unzips a zip file
func UnzipFile(app *ImageBuild) error {
	saveto := "/usr/local/crane/zip/" + app.Name
	zipFile := "/usr/local/crane/blobs/" + app.Name + ".zip"
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return fmt.Errorf("error opening zip file: %v", err)
	}

	defer zipReader.Close()

	for _, file := range zipReader.File {
		// Create the file in the destination directory
		destPath := filepath.Join(saveto, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(destPath, os.ModePerm)
			continue
		}

		destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return fmt.Errorf("error creating file: %v", err)
		}
		defer destFile.Close()

		srcFile, err := file.Open()
		if err != nil {
			return fmt.Errorf("error opening file: %v", err)
		}
		defer srcFile.Close()

		if _, err := io.Copy(destFile, srcFile); err != nil {
			return fmt.Errorf("error copying file: %v", err)
		}
	}
	fmt.Println("Unzipped")
	return nil
}

// HandleFileSource handles the file source type
func HandleFileSource(app *ImageBuild) error {

	// Download the file
	fmt.Println("Downloading file")
	err := DownloadFile(app)
	if err != nil {
		return err
	}
	fmt.Println("Downloaded file")

	fmt.Println("Unzipping file")
	// Unzip the file
	err = UnzipFile(app)
	if err != nil {
		return err
	}
	fmt.Println("Unzipped file")

	return nil
}
