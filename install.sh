#!/bin/bash

# updatectrl Linux Installer

set -e

echo "Installing updatectrl..."

if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go first."
    exit 1
fi

echo "Building updatectrl..."
go build -o updatectrl main.go

echo "Installing to /usr/local/bin..."
sudo cp updatectrl /usr/local/bin/
sudo chmod +x /usr/local/bin/updatectrl

echo "updatectrl installed successfully!"
echo "Run 'updatectrl init' to set up the daemon."