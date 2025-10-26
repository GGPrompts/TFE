#!/usr/bin/env bash
# TFE Wrapper Script
# This wrapper enables the "Quick CD" feature to change directories after exiting TFE
#
# Usage: Add this to your ~/.bashrc or ~/.zshrc:
#   source /path/to/tfe-wrapper.sh
#
# Then use 'tfe' instead of './tfe' to launch the file explorer

tfe() {
    # Auto-detect TFE binary location
    local TFE_BIN=""
    local CD_TARGET="$HOME/.tfe_cd_target"

    # Check common locations in order of preference
    # NOTE: Check specific paths first to avoid finding this wrapper function
    if [ -f "$HOME/go/bin/tfe" ]; then
        # Go install default location (preferred)
        TFE_BIN="$HOME/go/bin/tfe"
    elif [ -f "$HOME/.local/bin/tfe" ] && [ -x "$HOME/.local/bin/tfe" ]; then
        # Local installation
        TFE_BIN="$HOME/.local/bin/tfe"
    elif [ -f "$HOME/.config/tfe/tfe" ]; then
        # Alternative local installation
        TFE_BIN="$HOME/.config/tfe/tfe"
    elif [ -f "/usr/local/bin/tfe" ]; then
        # System-wide installation
        TFE_BIN="/usr/local/bin/tfe"
    elif command -v tfe &> /dev/null; then
        # Fallback: search PATH (but may find wrapper itself)
        TFE_BIN="$(command -v tfe)"
    else
        echo "Error: TFE binary not found"
        echo "Please ensure TFE is installed and in your PATH"
        return 1
    fi

    # Clear any previous cd target
    rm -f "$CD_TARGET"

    # Run TFE
    "$TFE_BIN" "$@"

    # Check if TFE wrote a cd target
    if [ -f "$CD_TARGET" ]; then
        local TARGET_DIR="$(cat "$CD_TARGET")"
        if [ -d "$TARGET_DIR" ]; then
            cd "$TARGET_DIR"
            echo "Changed directory to: $TARGET_DIR"
        fi
        rm -f "$CD_TARGET"
    fi
}
