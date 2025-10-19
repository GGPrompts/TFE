#!/bin/bash
# compact-reload.sh
# Automated compact and reload for Claude Code sessions
#
# This script:
# 1. Checks for a compact summary file in docs/NEXT_SESSION.md
# 2. Exits the current Claude session
# 3. Starts a new Claude session with the summary pre-loaded

SUMMARY_FILE="docs/NEXT_SESSION.md"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔄 Claude Code Session Reload${NC}\n"

# Check if summary file exists
if [ ! -f "$SUMMARY_FILE" ]; then
    echo -e "${RED}❌ Error: Summary file not found${NC}"
    echo -e "Expected location: $SUMMARY_FILE"
    echo ""
    echo "Please run /save-session in your Claude Code session first."
    echo "It will create docs/NEXT_SESSION.md with the session summary."
    echo ""
    echo "Tip: You can also check if an old summary exists:"
    echo "  cat docs/NEXT_SESSION.md"
    exit 1
fi

# Show summary preview
echo -e "${GREEN}✅ Found summary file${NC}"
echo -e "${YELLOW}Preview (first 10 lines):${NC}"
echo "─────────────────────────────────────────────"
head -n 10 "$SUMMARY_FILE"
echo "─────────────────────────────────────────────"
echo ""

# Get file size and line count
FILE_SIZE=$(wc -c < "$SUMMARY_FILE")
LINE_COUNT=$(wc -l < "$SUMMARY_FILE")
echo -e "Summary: ${LINE_COUNT} lines, ${FILE_SIZE} bytes"
echo ""

# Ask for confirmation
read -p "Continue with this summary? (y/N) " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}⚠️  Cancelled${NC}"
    echo "You can view the full summary with: cat $SUMMARY_FILE"
    exit 0
fi

# Prepare the summary content
SUMMARY_CONTENT=$(cat "$SUMMARY_FILE")

# Check if there's a specific goal in the summary
if grep -q "## NEXT SESSION GOAL" "$SUMMARY_FILE"; then
    # Extract the goal
    GOAL_SECTION=$(sed -n '/## NEXT SESSION GOAL/,/```/p' "$SUMMARY_FILE" | grep -v "^##" | grep -v "^\`\`\`" | grep -v "^\[If" | grep -v "^---")

    # Create prompt with goal first
    INITIAL_PROMPT="$GOAL_SECTION

---

Here's the summary from my previous session:

$SUMMARY_CONTENT

---

Let's get started on the goal above!"
else
    # No specific goal, standard prompt
    INITIAL_PROMPT="I'm continuing from a previous session. Here's the summary:

$SUMMARY_CONTENT

---

Ready to continue from where we left off. What would you like to work on?"
fi

echo ""
echo -e "${BLUE}🚀 Starting new Claude Code session with summary...${NC}"
echo ""

# Start Claude Code with the summary as initial prompt
# Note: This will start in the current directory
claude "$INITIAL_PROMPT"

# Clean up old summary after successful reload (optional)
# Uncomment the line below if you want to auto-delete the summary
# rm "$SUMMARY_FILE"
