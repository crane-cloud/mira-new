#!/bin/bash

set -e

# Function to download a specific asset from the latest GitHub release
download_asset() {
  local repo=$1
  local asset_name=$2

  echo "Fetching latest release for $repo..."
  release_url="https://api.github.com/repos/$repo/releases/latest"
  
  # Get download URL for the asset
  download_url=$(curl -s "$release_url" | grep "browser_download_url" | grep "$asset_name" | cut -d '"' -f 4)

  if [[ -z "$download_url" ]]; then
    echo "Error: Could not find asset $asset_name in $repo"
    exit 1
  fi

  echo "Downloading $asset_name from $repo..."
  curl -LO "$download_url"

  if [[ "$asset_name" == "mira-ubuntu-latest" ]]; then
    echo "Making mira-ubuntu executable..."
    chmod +x mira-ubuntu-latest
  fi
}

# Download compose.yml and loki.yml from open-ug/conveyor
download_asset "open-ug/conveyor" "compose.yml"
download_asset "open-ug/conveyor" "loki.yml"

# Download mira-ubuntu from crane-cloud/mira-new
download_asset "crane-cloud/mira-new" "mira-ubuntu-latest"

echo "All assets downloaded successfully."
