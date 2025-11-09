#!/bin/bash

# updatectl Linux Installer

set -e

echo "Installing updatectl..."

if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go first."
    exit 1
fi

echo "Building updatectl..."
go build -o updatectl main.go

echo "Installing to /usr/local/bin..."
sudo cp updatectl /usr/local/bin/
sudo chmod +x /usr/local/bin/updatectl

echo "updatectl installed successfully!"
echo "Run 'updatectl init' to set up the daemon."