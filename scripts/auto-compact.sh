#!/bin/bash
# auto-compact.sh
# Automatically compact Claude session in-place using tmux send-keys
#
# This script:
# 1. Sends /save-session to your Claude tmux pane
# 2. Waits for summary to be created
# 3. Sends /clear
# 4. Pastes the summary back
# All in the same session!

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# SUMMARY_FILE will be set later based on tmux pane's working directory
SUMMARY_FILE=""

# Function to show usage
show_usage() {
    echo "Usage: auto-compact [OPTIONS]"
    echo ""
    echo "Automatically compact Claude Code session in-place."
    echo ""
    echo "Options:"
    echo "  -t, --tmux SESSION    Tmux session name (default: auto-detect)"
    echo "  -p, --pane PANE       Tmux pane (default: active pane in session)"
    echo "  -g, --goal \"GOAL\"     Next session goal (optional)"
    echo "  -c, --commit          Commit changes before compacting"
    echo "  -m, --message \"MSG\"   Custom git commit message"
    echo "  -h, --help            Show this help"
    echo ""
    echo "Examples:"
    echo "  auto-compact                          # Auto-detect tmux session"
    echo "  auto-compact -t claude-session        # Specify session"
    echo "  auto-compact -g \"Fix tree view bug\"  # With goal"
    echo "  auto-compact --commit -g \"Phase 2\"   # Commit + compact + goal"
    echo "  auto-compact -c -m \"Phase 1 done\"    # Custom commit message"
}

# Parse arguments
TMUX_SESSION=""
TMUX_PANE=""
GOAL=""
DO_COMMIT=false
COMMIT_MESSAGE=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--tmux)
            TMUX_SESSION="$2"
            shift 2
            ;;
        -p|--pane)
            TMUX_PANE="$2"
            shift 2
            ;;
        -g|--goal)
            GOAL="$2"
            shift 2
            ;;
        -c|--commit)
            DO_COMMIT=true
            shift
            ;;
        -m|--message)
            COMMIT_MESSAGE="$2"
            DO_COMMIT=true
            shift 2
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            show_usage
            exit 1
            ;;
    esac
done

echo -e "${BLUE}🔄 Auto-Compact Claude Session${NC}\n"

# Check if tmux is running
if ! command -v tmux &> /dev/null; then
    echo -e "${RED}❌ Error: tmux not found${NC}"
    echo "This script requires tmux. Install with: sudo apt install tmux"
    exit 1
fi

# Auto-detect tmux session if not specified
if [ -z "$TMUX_SESSION" ]; then
    echo -e "${YELLOW}🔍 Auto-detecting tmux session with Claude...${NC}"

    # Look for sessions with 'claude' in a pane
    DETECTED_SESSION=$(tmux list-sessions -F '#{session_name}' 2>/dev/null | while read session; do
        if tmux list-panes -t "$session" -F '#{pane_current_command}' 2>/dev/null | grep -q 'claude'; then
            echo "$session"
            break
        fi
    done)

    if [ -z "$DETECTED_SESSION" ]; then
        echo -e "${RED}❌ Error: Could not auto-detect Claude session${NC}"
        echo "Please specify session with: auto-compact -t SESSION_NAME"
        echo ""
        echo "Available tmux sessions:"
        tmux list-sessions 2>/dev/null || echo "  (none)"
        exit 1
    fi

    TMUX_SESSION="$DETECTED_SESSION"
    echo -e "${GREEN}✅ Found Claude in session: $TMUX_SESSION${NC}"
fi

# Verify session exists
if ! tmux has-session -t "$TMUX_SESSION" 2>/dev/null; then
    echo -e "${RED}❌ Error: Tmux session '$TMUX_SESSION' not found${NC}"
    echo ""
    echo "Available sessions:"
    tmux list-sessions
    exit 1
fi

# Auto-detect pane with Claude if not specified
if [ -z "$TMUX_PANE" ]; then
    TMUX_PANE=$(tmux list-panes -t "$TMUX_SESSION" -F '#{pane_index} #{pane_current_command}' | grep 'claude' | head -1 | cut -d' ' -f1)

    if [ -z "$TMUX_PANE" ]; then
        # Default to pane 0
        TMUX_PANE="0"
    fi
fi

TARGET="${TMUX_SESSION}:${TMUX_PANE}"

# Get the working directory from the tmux pane
PANE_DIR=$(tmux display-message -t "$TARGET" -p "#{pane_current_path}")
SUMMARY_FILE="$PANE_DIR/docs/NEXT_SESSION.md"

# Ensure docs directory exists
mkdir -p "$PANE_DIR/docs" 2>/dev/null || true

echo -e "${GREEN}Target:${NC} $TARGET"
echo -e "${GREEN}Project:${NC} $PANE_DIR"
echo -e "${GREEN}Summary:${NC} docs/NEXT_SESSION.md"
if [ -n "$GOAL" ]; then
    echo -e "${GREEN}Goal:${NC} $GOAL"
fi
if [ "$DO_COMMIT" = true ]; then
    echo -e "${GREEN}Git Commit:${NC} Enabled"
    if [ -n "$COMMIT_MESSAGE" ]; then
        echo -e "${GREEN}Message:${NC} $COMMIT_MESSAGE"
    fi
fi
echo ""

# Step 0: Git commit (if requested)
if [ "$DO_COMMIT" = true ]; then
    echo -e "${BLUE}📝 Step 0: Committing changes...${NC}"

    # Change to the pane's directory for git operations
    cd "$PANE_DIR"

    # Check if we're in a git repo
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  Not in a git repository - skipping commit${NC}"
    else
        # Check if there are changes to commit
        if git diff --quiet && git diff --cached --quiet; then
            echo -e "${YELLOW}⚠️  No changes to commit${NC}"
        else
            # Auto-generate commit message if not provided
            if [ -z "$COMMIT_MESSAGE" ]; then
                if [ -n "$GOAL" ]; then
                    COMMIT_MESSAGE="$GOAL"
                else
                    COMMIT_MESSAGE="Auto-compact checkpoint

🤖 Generated with Claude Code auto-compact"
                fi
            fi

            # Stage all changes
            git add -A

            # Commit
            git commit -m "$COMMIT_MESSAGE" 2>&1

            if [ $? -eq 0 ]; then
                echo -e "${GREEN}✅ Changes committed${NC}"
                git log --oneline -1
            else
                echo -e "${RED}❌ Commit failed${NC}"
                echo "Continuing with compact anyway..."
            fi
        fi
    fi
    echo ""
fi

# Remove old summary file
rm -f "$SUMMARY_FILE"

# Step 1: Send /save-session command
STEP_NUM=$([[ "$DO_COMMIT" = true ]] && echo "1/5" || echo "1/4")
echo -e "${BLUE}📝 Step $STEP_NUM: Sending /save-session command...${NC}"
if [ -n "$GOAL" ]; then
    tmux send-keys -t "$TARGET" "/save-session $GOAL" C-m
else
    tmux send-keys -t "$TARGET" "/save-session" C-m
fi

# Step 2: Wait for summary file to be created
STEP_NUM=$([[ "$DO_COMMIT" = true ]] && echo "2/5" || echo "2/4")
echo -e "${BLUE}⏳ Step $STEP_NUM: Waiting for summary generation...${NC}"
MAX_WAIT=30  # Maximum 30 seconds
WAITED=0

while [ ! -f "$SUMMARY_FILE" ] && [ $WAITED -lt $MAX_WAIT ]; do
    sleep 1
    WAITED=$((WAITED + 1))
    echo -ne "\r   Waiting... ${WAITED}s"
done
echo ""

if [ ! -f "$SUMMARY_FILE" ]; then
    echo -e "${RED}❌ Error: Summary file not created after ${MAX_WAIT}s${NC}"
    echo "Check if Claude responded to /save-session command"
    exit 1
fi

echo -e "${GREEN}✅ Summary created (${WAITED}s)${NC}"

# Give Claude a moment to finish printing
sleep 2

# Step 3: Send /clear command
STEP_NUM=$([[ "$DO_COMMIT" = true ]] && echo "3/5" || echo "3/4")
echo -e "${BLUE}🧹 Step $STEP_NUM: Clearing conversation...${NC}"
tmux send-keys -t "$TARGET" "/clear" C-m

# Wait for clear to complete
sleep 1

# Step 4: Paste summary back
STEP_NUM=$([[ "$DO_COMMIT" = true ]] && echo "4/5" || echo "4/4")
echo -e "${BLUE}📋 Step $STEP_NUM: Pasting summary...${NC}"

# Read summary and send it line by line (more reliable than one big paste)
# But first, let's try sending it as a single block with proper escaping
SUMMARY_CONTENT=$(cat "$SUMMARY_FILE")

# Create a temporary file with the summary
TEMP_PASTE=$(mktemp)
cat "$SUMMARY_FILE" > "$TEMP_PASTE"

# Use tmux load-buffer and paste-buffer for reliable pasting
tmux load-buffer "$TEMP_PASTE"
tmux paste-buffer -t "$TARGET"

# Clean up
rm -f "$TEMP_PASTE"

# Give it a moment to paste
sleep 1

# Send Enter to submit
tmux send-keys -t "$TARGET" C-m

echo ""
echo -e "${GREEN}✅ Auto-compact complete!${NC}"
echo ""
echo "Your Claude session has been compacted with the summary."
echo "Summary saved to: docs/NEXT_SESSION.md"
if [ -n "$GOAL" ]; then
    echo "Next goal: $GOAL"
fi
if [ "$DO_COMMIT" = true ]; then
    echo ""
    echo "Git commit created - summary is preserved in history!"
    echo "View it later with: git show HEAD:docs/NEXT_SESSION.md"
fi
echo ""
echo "Switch to your Claude tmux session to see the result:"
echo "  tmux attach -t $TMUX_SESSION"
echo ""
echo "Review summary anytime with:"
echo "  cat docs/NEXT_SESSION.md"
