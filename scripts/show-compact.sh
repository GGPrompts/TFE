#!/bin/bash
# show-compact.sh
# Display the session summary from docs/NEXT_SESSION.md and copy to clipboard

SUMMARY_FILE="docs/NEXT_SESSION.md"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

if [ ! -f "$SUMMARY_FILE" ]; then
    echo -e "${RED}❌ No session summary found${NC}"
    echo "Expected: $SUMMARY_FILE"
    echo ""
    echo "Run /save-session in Claude Code to create it."
    exit 1
fi

# Show file info
FILE_SIZE=$(wc -c < "$SUMMARY_FILE")
LINE_COUNT=$(wc -l < "$SUMMARY_FILE")
echo -e "${GREEN}📋 Session Summary${NC} ${BLUE}(${LINE_COUNT} lines, ${FILE_SIZE} bytes)${NC}"
echo "═══════════════════════════════════════════════════════════"
cat "$SUMMARY_FILE"
echo "═══════════════════════════════════════════════════════════"
echo ""

# Try to copy to clipboard
if command -v xclip &> /dev/null; then
    cat "$SUMMARY_FILE" | xclip -selection clipboard
    echo -e "${GREEN}✅ Copied to clipboard (xclip)${NC}"
elif command -v clip.exe &> /dev/null; then
    cat "$SUMMARY_FILE" | clip.exe
    echo -e "${GREEN}✅ Copied to clipboard (WSL clip.exe)${NC}"
elif command -v pbcopy &> /dev/null; then
    cat "$SUMMARY_FILE" | pbcopy
    echo -e "${GREEN}✅ Copied to clipboard (pbcopy)${NC}"
else
    echo -e "${YELLOW}⚠️  Clipboard copy not available${NC}"
    echo "Install xclip: sudo apt install xclip"
fi

echo ""
echo "Next steps:"
echo "  Option 1: Use built-in /compact in Claude Code (keeps same session)"
echo "  Option 2: Run 'session-reload' for completely fresh session"
echo "  Option 3: Manual - /clear then paste"
echo ""
echo "Automated reload:"
echo "  session-reload"
