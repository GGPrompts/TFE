#!/usr/bin/env bash
# TFE Build & Install Script
# Builds the binary and installs it to ~/.local/bin/tfe

set -e  # Exit on error

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔨 Building TFE...${NC}"

# Build the binary
go build

if [ ! -f "./tfe" ]; then
    echo -e "${RED}❌ Build failed - binary not created${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Build successful${NC}"

# Get binary size for display
SIZE=$(ls -lh ./tfe | awk '{print $5}')
echo -e "${BLUE}📦 Binary size: ${SIZE}${NC}"

# Install to ~/.local/bin/
INSTALL_DIR="$HOME/.local/bin"
INSTALL_PATH="$INSTALL_DIR/tfe"

# Create directory if it doesn't exist
mkdir -p "$INSTALL_DIR"

echo -e "${BLUE}📋 Installing to ${INSTALL_PATH}...${NC}"

# Copy binary
cp ./tfe "$INSTALL_PATH"

# Make it executable (should already be, but just in case)
chmod +x "$INSTALL_PATH"

# Verify installation
if [ -f "$INSTALL_PATH" ]; then
    INSTALLED_SIZE=$(ls -lh "$INSTALL_PATH" | awk '{print $5}')
    echo -e "${GREEN}✓ Installed successfully${NC}"
    echo -e "${BLUE}📍 Location: ${INSTALL_PATH}${NC}"
    echo -e "${BLUE}📦 Installed size: ${INSTALLED_SIZE}${NC}"

    # Verify checksums match
    LOCAL_MD5=$(md5sum ./tfe | awk '{print $1}')
    INSTALLED_MD5=$(md5sum "$INSTALL_PATH" | awk '{print $1}')

    if [ "$LOCAL_MD5" = "$INSTALLED_MD5" ]; then
        echo -e "${GREEN}✓ Checksums match - installation verified${NC}"
    else
        echo -e "${RED}⚠️  Warning: Checksums don't match!${NC}"
        exit 1
    fi
else
    echo -e "${RED}❌ Installation failed${NC}"
    exit 1
fi

echo -e "${GREEN}🎉 Done! You can now run 'tfe' from anywhere.${NC}"
