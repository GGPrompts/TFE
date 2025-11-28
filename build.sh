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
# CGO_ENABLED=0 is needed on Termux (Android) where the CGO toolchain is often broken
# It also produces a fully static binary which is more portable
CGO_ENABLED=0 go build

if [ ! -f "./tfe" ]; then
    echo -e "${RED}❌ Build failed - binary not created${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Build successful${NC}"

# Get binary size for display
SIZE=$(ls -lh ./tfe | awk '{print $5}')
echo -e "${BLUE}📦 Binary size: ${SIZE}${NC}"

# Check if TFE is currently running
if pgrep -x tfe > /dev/null; then
    echo -e "${YELLOW}⚠️  TFE is currently running${NC}"
    echo -e "${YELLOW}   The binary will be updated, but running instances will use the old version${NC}"
    echo -e "${YELLOW}   Restart TFE to use the new version${NC}"
    echo ""
fi

# Auto-discover all TFE installations
echo -e "${BLUE}🔍 Discovering existing TFE installations...${NC}"
TFE_LOCATIONS=(
    "$HOME/.local/bin/tfe"
    "$HOME/go/bin/tfe"
    "$HOME/bin/tfe"
    "/usr/local/bin/tfe"
)

FOUND_LOCATIONS=()
for location in "${TFE_LOCATIONS[@]}"; do
    if [ -f "$location" ]; then
        FOUND_LOCATIONS+=("$location")
        echo -e "${BLUE}  Found: ${location}${NC}"
    fi
done

# If no installations found, use default location
if [ ${#FOUND_LOCATIONS[@]} -eq 0 ]; then
    FOUND_LOCATIONS=("$HOME/.local/bin/tfe")
    echo -e "${BLUE}  No existing installations found${NC}"
    echo -e "${BLUE}  Will install to: ~/.local/bin/tfe${NC}"
fi

echo ""

# Update all found locations
for location in "${FOUND_LOCATIONS[@]}"; do
    INSTALL_DIR=$(dirname "$location")

    # Create directory if it doesn't exist
    mkdir -p "$INSTALL_DIR"

    echo -e "${BLUE}📋 Installing to ${location}...${NC}"

    if cp ./tfe "$location" 2>/dev/null; then
        chmod +x "$location"
        echo -e "${GREEN}✓ Updated ${location}${NC}"
    else
        echo -e "${RED}⚠️  Failed to update ${location}${NC}"
        echo -e "${RED}   (may be in use or permission denied)${NC}"
    fi
done

# Primary installation path for verification
INSTALL_PATH="${FOUND_LOCATIONS[0]}"

# Copy HOTKEYS.md to all installation directories for F1 help
if [ -f "./HOTKEYS.md" ]; then
    echo ""
    echo -e "${BLUE}📖 Installing HOTKEYS.md for F1 help...${NC}"
    for location in "${FOUND_LOCATIONS[@]}"; do
        HOTKEYS_DIR=$(dirname "$location")
        if cp ./HOTKEYS.md "$HOTKEYS_DIR/HOTKEYS.md" 2>/dev/null; then
            echo -e "${GREEN}✓ Copied to ${HOTKEYS_DIR}/${NC}"
        fi
    done
fi

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
