#!/bin/bash

# GitHub repo info
REPO="crane-cloud/mira-new"
ASSET_NAME="mira-ubuntu"

# Get latest release data from GitHub API
echo "Fetching latest release info..."
API_URL="https://api.github.com/repos/${REPO}/releases/latest"
RELEASE_DATA=$(curl -s "$API_URL")


DOWNLOAD_URL=$(echo "$RELEASE_DATA" | grep "browser_download_url" | grep "$ASSET_NAME" | cut -d '"' -f 4)

# Check if URL was found
if [ -z "$DOWNLOAD_URL" ]; then
    echo "Error: Asset '$ASSET_NAME' not found in the latest release."
    exit 1
fi

# Download the file
echo "Downloading $ASSET_NAME from $DOWNLOAD_URL..."
curl -L -o "$ASSET_NAME" "$DOWNLOAD_URL"

# Make it executable
chmod +x "$ASSET_NAME"

echo "Download complete. File saved as '$ASSET_NAME'."
