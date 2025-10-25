#!/usr/bin/env bash
#
# TFE Installation Script
# Installs TFE with Quick CD feature (like Midnight Commander)
#
# Usage: curl -sSL https://raw.githubusercontent.com/GGPrompts/TFE/main/install.sh | bash
#

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   TFE - Terminal File Explorer        ║${NC}"
echo -e "${BLUE}║   Installation Script                  ║${NC}"
echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}✗ Error: Go is not installed${NC}"
    echo "Please install Go 1.24+ from https://go.dev/dl/"
    exit 1
fi

echo -e "${GREEN}✓${NC} Go found: $(go version)"

# Install TFE binary
echo ""
echo -e "${BLUE}Installing TFE binary...${NC}"
if go install github.com/GGPrompts/tfe@latest; then
    echo -e "${GREEN}✓${NC} TFE binary installed to ~/go/bin/tfe"
else
    echo -e "${RED}✗ Failed to install TFE${NC}"
    exit 1
fi

# Verify binary is in PATH
if ! command -v tfe &> /dev/null; then
    echo -e "${YELLOW}⚠${NC}  TFE binary not found in PATH"
    echo "   Add ~/go/bin to your PATH:"
    echo "   export PATH=\$PATH:~/go/bin"
    echo ""
fi

# Detect shell
SHELL_NAME=$(basename "$SHELL")
SHELL_RC=""

case "$SHELL_NAME" in
    bash)
        SHELL_RC="$HOME/.bashrc"
        ;;
    zsh)
        SHELL_RC="$HOME/.zshrc"
        ;;
    fish)
        echo -e "${YELLOW}⚠${NC}  Fish shell detected - manual setup required"
        echo "   See: https://github.com/GGPrompts/TFE#fish-shell-setup"
        SHELL_RC=""
        ;;
    *)
        echo -e "${YELLOW}⚠${NC}  Unknown shell: $SHELL_NAME"
        echo "   Manual wrapper setup required"
        SHELL_RC=""
        ;;
esac

# Download and setup wrapper for bash/zsh
if [ -n "$SHELL_RC" ]; then
    echo ""
    echo -e "${BLUE}Setting up Quick CD feature...${NC}"

    # Create ~/.config/tfe directory for wrapper
    mkdir -p "$HOME/.config/tfe"

    # Download wrapper script
    WRAPPER_PATH="$HOME/.config/tfe/tfe-wrapper.sh"
    echo "Downloading wrapper script..."

    if curl -sSL https://raw.githubusercontent.com/GGPrompts/TFE/main/tfe-wrapper.sh -o "$WRAPPER_PATH"; then
        chmod +x "$WRAPPER_PATH"
        echo -e "${GREEN}✓${NC} Wrapper downloaded to $WRAPPER_PATH"
    else
        echo -e "${RED}✗ Failed to download wrapper${NC}"
        echo "You can manually download it from:"
        echo "https://github.com/GGPrompts/TFE/blob/main/tfe-wrapper.sh"
        exit 1
    fi

    # Check if wrapper is already sourced
    if grep -q "tfe-wrapper.sh" "$SHELL_RC" 2>/dev/null; then
        echo -e "${YELLOW}⚠${NC}  Wrapper already configured in $SHELL_RC"
    else
        # Add wrapper to shell config
        echo "" >> "$SHELL_RC"
        echo "# TFE Quick CD wrapper" >> "$SHELL_RC"
        echo "source $WRAPPER_PATH" >> "$SHELL_RC"
        echo -e "${GREEN}✓${NC} Wrapper added to $SHELL_RC"
    fi

    echo ""
    echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║   Installation Complete! 🎉            ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "To start using TFE with Quick CD:"
    echo -e "${BLUE}1.${NC} Reload your shell:"
    echo -e "   ${YELLOW}source $SHELL_RC${NC}"
    echo -e "${BLUE}2.${NC} Launch TFE:"
    echo -e "   ${YELLOW}tfe${NC}"
    echo ""
    echo -e "Features enabled:"
    echo -e "  ${GREEN}✓${NC} Quick CD - Right-click folder → 'Quick CD' exits and changes directory"
    echo -e "  ${GREEN}✓${NC} All TFE features available"
    echo ""
else
    echo ""
    echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║   Binary Installed! ⚠️                 ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${YELLOW}Note:${NC} Quick CD requires manual wrapper setup for your shell"
    echo "See: https://github.com/GGPrompts/TFE#installation"
    echo ""
fi

echo -e "Documentation: ${BLUE}https://github.com/GGPrompts/TFE${NC}"
echo -e "Hotkeys: Press ${YELLOW}F1${NC} in TFE for help"
echo ""
