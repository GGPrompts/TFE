#!/usr/bin/env bash
#
# TFE Uninstallation Script
# Removes TFE binary and wrapper configuration
#
# Usage: curl -sSL https://raw.githubusercontent.com/GGPrompts/TFE/main/uninstall.sh | bash
#

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   TFE - Uninstallation Script         ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
echo ""

# Remove TFE binary
TFE_BINARY="$HOME/go/bin/tfe"
if [ -f "$TFE_BINARY" ]; then
    echo -e "${BLUE}Removing TFE binary...${NC}"
    rm -f "$TFE_BINARY"
    echo -e "${GREEN}✓${NC} Removed $TFE_BINARY"
else
    echo -e "${YELLOW}⚠${NC}  TFE binary not found at $TFE_BINARY"
fi

# Remove wrapper script
WRAPPER_PATH="$HOME/.config/tfe/tfe-wrapper.sh"
if [ -f "$WRAPPER_PATH" ]; then
    echo -e "${BLUE}Removing wrapper script...${NC}"
    rm -f "$WRAPPER_PATH"
    echo -e "${GREEN}✓${NC} Removed $WRAPPER_PATH"
else
    echo -e "${YELLOW}⚠${NC}  Wrapper script not found"
fi

# Remove wrapper directory if empty
if [ -d "$HOME/.config/tfe" ]; then
    if [ -z "$(ls -A $HOME/.config/tfe)" ]; then
        rmdir "$HOME/.config/tfe"
        echo -e "${GREEN}✓${NC} Removed ~/.config/tfe directory"
    fi
fi

# Detect shell configs to clean
SHELL_CONFIGS=("$HOME/.bashrc" "$HOME/.zshrc")

echo ""
echo -e "${BLUE}Checking shell configurations...${NC}"

for SHELL_RC in "${SHELL_CONFIGS[@]}"; do
    if [ -f "$SHELL_RC" ]; then
        if grep -q "tfe-wrapper" "$SHELL_RC"; then
            echo -e "${YELLOW}Found TFE wrapper reference in:${NC} $SHELL_RC"
            echo -e "   To complete uninstall, remove these lines:"
            echo -e "   ${YELLOW}# TFE Quick CD wrapper${NC}"
            echo -e "   ${YELLOW}source ~/.config/tfe/tfe-wrapper.sh${NC}"
            echo ""
            echo -e "   Or run: ${BLUE}sed -i '/tfe-wrapper/d' $SHELL_RC${NC}"
            echo ""
        fi
    fi
done

echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║   Uninstallation Complete!             ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
echo ""
echo -e "${YELLOW}Note:${NC} You may need to remove wrapper references from shell config manually"
echo -e "      Then reload your shell: ${BLUE}source ~/.bashrc${NC} or ${BLUE}source ~/.zshrc${NC}"
echo ""
