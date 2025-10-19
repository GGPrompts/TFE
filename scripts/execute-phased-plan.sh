#!/bin/bash
# execute-phased-plan.sh
# Automatically execute a multi-phase plan with auto-compacting between phases
#
# This script:
# 1. Reads a plan file with phases
# 2. Sends each phase to Claude
# 3. Waits for completion
# 4. Auto-compacts between phases
# 5. Continues to next phase
# All fully automated!

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

show_usage() {
    echo "Usage: execute-phased-plan [OPTIONS] PLAN_FILE"
    echo ""
    echo "Automatically execute a multi-phase plan with auto-compacting."
    echo ""
    echo "Options:"
    echo "  -t, --tmux SESSION    Tmux session name (default: auto-detect)"
    echo "  -p, --pane PANE       Tmux pane (default: auto-detect)"
    echo "  -i, --interactive     Pause after each phase for review"
    echo "  -h, --help            Show this help"
    echo ""
    echo "Plan file format:"
    echo "  Phase 1: Research syntax highlighting libraries"
    echo "  Phase 2: Implement basic highlighting"
    echo "  Phase 3: Add language detection"
    echo ""
    echo "Example:"
    echo "  execute-phased-plan my-plan.txt"
    echo "  execute-phased-plan -t claude-dev -i phases.txt"
}

# Parse arguments
TMUX_SESSION=""
TMUX_PANE=""
INTERACTIVE=false
PLAN_FILE=""

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
        -i|--interactive)
            INTERACTIVE=true
            shift
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            PLAN_FILE="$1"
            shift
            ;;
    esac
done

if [ -z "$PLAN_FILE" ]; then
    echo -e "${RED}❌ Error: Plan file required${NC}"
    show_usage
    exit 1
fi

if [ ! -f "$PLAN_FILE" ]; then
    echo -e "${RED}❌ Error: Plan file not found: $PLAN_FILE${NC}"
    exit 1
fi

echo -e "${CYAN}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║         🚀 Phased Plan Executor with Auto-Compact 🚀         ║${NC}"
echo -e "${CYAN}╚══════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Auto-detect tmux session if needed
if [ -z "$TMUX_SESSION" ]; then
    echo -e "${YELLOW}🔍 Auto-detecting Claude session...${NC}"
    TMUX_SESSION=$(tmux list-sessions -F '#{session_name}' 2>/dev/null | while read session; do
        if tmux list-panes -t "$session" -F '#{pane_current_command}' 2>/dev/null | grep -q 'claude'; then
            echo "$session"
            break
        fi
    done)

    if [ -z "$TMUX_SESSION" ]; then
        echo -e "${RED}❌ Error: No Claude session found${NC}"
        exit 1
    fi
    echo -e "${GREEN}✅ Found: $TMUX_SESSION${NC}"
fi

# Auto-detect pane
if [ -z "$TMUX_PANE" ]; then
    TMUX_PANE=$(tmux list-panes -t "$TMUX_SESSION" -F '#{pane_index} #{pane_current_command}' | grep 'claude' | head -1 | cut -d' ' -f1)
    [ -z "$TMUX_PANE" ] && TMUX_PANE="0"
fi

TARGET="${TMUX_SESSION}:${TMUX_PANE}"

# Read phases from file
mapfile -t PHASES < "$PLAN_FILE"

# Filter out empty lines and comments
FILTERED_PHASES=()
for phase in "${PHASES[@]}"; do
    # Skip empty lines and comments
    [[ -z "$phase" || "$phase" =~ ^[[:space:]]*# ]] && continue
    FILTERED_PHASES+=("$phase")
done

TOTAL_PHASES=${#FILTERED_PHASES[@]}

if [ $TOTAL_PHASES -eq 0 ]; then
    echo -e "${RED}❌ Error: No phases found in plan file${NC}"
    exit 1
fi

echo -e "${GREEN}📋 Plan loaded: $TOTAL_PHASES phases${NC}"
echo ""
echo "Phases:"
for i in "${!FILTERED_PHASES[@]}"; do
    echo -e "  $(($i + 1)). ${FILTERED_PHASES[$i]}"
done
echo ""

if [ "$INTERACTIVE" = true ]; then
    echo -e "${YELLOW}Interactive mode: Will pause after each phase${NC}"
fi

echo ""
read -p "Ready to execute? (y/N) " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled"
    exit 0
fi

echo ""
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}                    Starting Execution                          ${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo ""

# Function to send command and wait for completion
send_phase() {
    local phase_num=$1
    local phase_desc=$2
    local total=$3

    echo -e "${MAGENTA}╔═══════════════════════════════════════════════════════════╗${NC}"
    echo -e "${MAGENTA}║  📋 PHASE $phase_num/$total: $phase_desc${NC}"
    echo -e "${MAGENTA}╚═══════════════════════════════════════════════════════════╝${NC}"

    # Create prompt for Claude
    local prompt="/phased-plan-execute Phase $phase_num/$total: $phase_desc"

    # Send to tmux
    tmux send-keys -t "$TARGET" "$prompt" C-m

    echo -e "${BLUE}→ Command sent to Claude${NC}"
    echo -e "${YELLOW}⏳ Waiting for phase completion...${NC}"
    echo ""
    echo "   (Claude is working on this phase)"
    echo "   (This may take a few minutes)"
    echo ""

    if [ "$INTERACTIVE" = true ]; then
        echo -e "${YELLOW}Press ENTER when phase is complete...${NC}"
        read
    else
        # In non-interactive mode, wait a reasonable time
        echo "   Waiting 60 seconds for phase to complete..."
        sleep 60
    fi
}

# Execute each phase
for i in "${!FILTERED_PHASES[@]}"; do
    phase_num=$(($i + 1))
    phase_desc="${FILTERED_PHASES[$i]}"

    send_phase $phase_num "$phase_desc" $TOTAL_PHASES

    # Auto-compact between phases (except after last phase)
    if [ $phase_num -lt $TOTAL_PHASES ]; then
        echo -e "${CYAN}🔄 Auto-compacting before next phase...${NC}"

        next_phase_num=$(($phase_num + 1))
        next_phase_desc="${FILTERED_PHASES[$next_phase_num - 1]}"

        # Use auto-compact script
        auto-compact -t "$TMUX_SESSION" -p "$TMUX_PANE" -g "Continue Phase $next_phase_num: $next_phase_desc"

        echo -e "${GREEN}✅ Compact complete - Ready for Phase $next_phase_num${NC}"
        echo ""

        if [ "$INTERACTIVE" = true ]; then
            read -p "Press ENTER to start Phase $next_phase_num..."
            echo ""
        else
            echo "   Starting Phase $next_phase_num in 3 seconds..."
            sleep 3
        fi
    fi
done

echo ""
echo -e "${CYAN}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║              ✅ All Phases Complete! ✅                   ║${NC}"
echo -e "${CYAN}╚═══════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${GREEN}🎉 Successfully executed all $TOTAL_PHASES phases!${NC}"
echo ""
echo "Summary:"
for i in "${!FILTERED_PHASES[@]}"; do
    phase_num=$(($i + 1))
    echo -e "  ✅ Phase $phase_num: ${FILTERED_PHASES[$i]}"
done
echo ""
echo "Check your Claude session for full results!"
