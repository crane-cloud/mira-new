package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

var CADDY_FILE string = `
:8080 {
	root * /workspace/` + OUTPUT_DIR + `
	file_server
}
`

var CADDY_VERSION = "2.4.6"

func InstallCaddy() error {
	dest := filepath.Join(CNB_LAYERS_DIR, fmt.Sprintf("caddy_%s_linux_amd64.tar.gz", CADDY_VERSION))
	caddyURL := "https://github.com/caddyserver/caddy/releases/download/v2.10.2/caddy_2.10.2_linux_amd64.tar.gz"

	// Download Caddy
	tarballPath, err := DownloadFile(caddyURL, dest)
	if err != nil {
		return fmt.Errorf("failed to download Caddy: %w", err)
	}

	// Extract the tarballs
	if err := ExtractTarGz(tarballPath, CADDY_LAYERS_DIR); err != nil {
		return fmt.Errorf("failed to extract Caddy tarball: %w", err)
	}

	// Clean up the downloaded tarball
	if err := os.Remove(tarballPath); err != nil {
		return fmt.Errorf("failed to remove tarball: %w", err)
	}

	return nil
}
