#!/bin/bash

# GitHub repo and asset details
REPO="crane-cloud/mira-new"
ASSET_NAME="mira-ubuntu"
INSTALL_PATH="/usr/local/bin/mira"

# Get latest release info
echo "Fetching latest release info..."
API_URL="https://api.github.com/repos/${REPO}/releases/latest"
RELEASE_DATA=$(curl -s "$API_URL")

# Extract download URL
DOWNLOAD_URL=$(echo "$RELEASE_DATA" | grep "browser_download_url" | grep "$ASSET_NAME" | cut -d '"' -f 4)

# Ensure the URL was found
if [ -z "$DOWNLOAD_URL" ]; then
    echo "Error: Asset '$ASSET_NAME' not found in latest release."
    exit 1
fi

# Download to a temporary location
TMP_FILE=$(mktemp)
echo "Downloading $ASSET_NAME to temporary file..."
curl -L -o "$TMP_FILE" "$DOWNLOAD_URL"

# Make it executable
chmod +x "$TMP_FILE"

# Move to /usr/local/bin as 'mira' (requires sudo)
echo "Installing to $INSTALL_PATH..."
sudo mv "$TMP_FILE" "$INSTALL_PATH"

echo "Installed 'mira' to $INSTALL_PATH."
