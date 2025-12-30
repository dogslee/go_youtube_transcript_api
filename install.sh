#!/bin/bash
# Installation script for youtube-transcript-api
# This script builds and installs the command-line tool with a proper name

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default installation directory
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Check if running as root for system-wide installation
if [ "$EUID" -eq 0 ]; then
    INSTALL_DIR="/usr/local/bin"
    echo -e "${YELLOW}Installing system-wide to $INSTALL_DIR${NC}"
else
    # Use user's local bin directory
    if [ -n "$GOBIN" ]; then
        INSTALL_DIR="$GOBIN"
    elif [ -n "$GOPATH" ]; then
        INSTALL_DIR="$GOPATH/bin"
    else
        INSTALL_DIR="$HOME/go/bin"
    fi
    echo -e "${YELLOW}Installing to user directory: $INSTALL_DIR${NC}"
fi

# Create install directory if it doesn't exist
mkdir -p "$INSTALL_DIR"

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Build the binary
echo -e "${GREEN}Building youtube-transcript-api...${NC}"
cd "$SCRIPT_DIR"
go build -o youtube-transcript-api ./cmd

if [ ! -f "youtube-transcript-api" ]; then
    echo -e "${RED}Error: Build failed${NC}"
    exit 1
fi

# Install the binary
echo -e "${GREEN}Installing to $INSTALL_DIR...${NC}"
if [ "$EUID" -eq 0 ]; then
    mv youtube-transcript-api "$INSTALL_DIR/"
else
    cp youtube-transcript-api "$INSTALL_DIR/"
    rm youtube-transcript-api
fi

# Make it executable
chmod +x "$INSTALL_DIR/youtube-transcript-api"

# Check if the directory is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo -e "${YELLOW}Warning: $INSTALL_DIR is not in your PATH${NC}"
    echo -e "${YELLOW}Add the following to your ~/.bashrc, ~/.zshrc, or ~/.profile:${NC}"
    echo -e "${GREEN}export PATH=\"\$PATH:$INSTALL_DIR\"${NC}"
else
    echo -e "${GREEN}Installation complete!${NC}"
    echo -e "${GREEN}You can now use 'youtube-transcript-api' command${NC}"
fi

# Verify installation
if command -v youtube-transcript-api &> /dev/null; then
    echo -e "${GREEN}Verification: youtube-transcript-api is available${NC}"
    youtube-transcript-api --version 2>/dev/null || echo -e "${YELLOW}Note: Run 'youtube-transcript-api --version' to verify${NC}"
else
    echo -e "${YELLOW}Note: You may need to restart your terminal or run:${NC}"
    echo -e "${GREEN}export PATH=\"\$PATH:$INSTALL_DIR\"${NC}"
fi


