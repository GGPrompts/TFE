#!/usr/bin/env bash
# TFE Wrapper Script
# This wrapper enables the "Quick CD" feature to change directories after exiting TFE
#
# Usage: Add this to your ~/.bashrc or ~/.zshrc:
#   source /path/to/tfe-wrapper.sh
#
# Then use 'tfe' instead of './tfe' to launch the file explorer

tfe() {
    # Auto-detect TFE binary location (works across all devices)
    # The wrapper is in the same directory as the tfe binary
    local WRAPPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    local TFE_BIN="$WRAPPER_DIR/tfe"
    local CD_TARGET="$HOME/.tfe_cd_target"

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
