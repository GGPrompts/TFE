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

    # Check common locations and find the most recent version
    local TFE_LOCATIONS=(
        "$HOME/go/bin/tfe"
        "$HOME/bin/tfe"
        "$HOME/.local/bin/tfe"
        "$HOME/.config/tfe/tfe"
        "/usr/local/bin/tfe"
    )

    local NEWEST_TFE=""
    local NEWEST_TIME=0
    local VERSION_MISMATCH=false
    local FOUND_COUNT=0
    local FIRST_MD5=""

    # Find all TFE installations and check for version mismatches
    for location in "${TFE_LOCATIONS[@]}"; do
        if [ -f "$location" ] && [ -x "$location" ]; then
            FOUND_COUNT=$((FOUND_COUNT + 1))

            # Get modification time
            if [[ "$OSTYPE" == "darwin"* ]]; then
                # macOS
                MODTIME=$(stat -f %m "$location" 2>/dev/null || echo 0)
            else
                # Linux
                MODTIME=$(stat -c %Y "$location" 2>/dev/null || echo 0)
            fi

            # Track newest binary
            if [ "$MODTIME" -gt "$NEWEST_TIME" ]; then
                NEWEST_TIME=$MODTIME
                NEWEST_TFE=$location
            fi

            # Check for version mismatches using md5sum
            if command -v md5sum &> /dev/null; then
                CURRENT_MD5=$(md5sum "$location" 2>/dev/null | awk '{print $1}')
                if [ -z "$FIRST_MD5" ]; then
                    FIRST_MD5=$CURRENT_MD5
                elif [ "$CURRENT_MD5" != "$FIRST_MD5" ]; then
                    VERSION_MISMATCH=true
                fi
            fi
        fi
    done

    # Fallback: search PATH if no specific location found
    if [ -z "$NEWEST_TFE" ]; then
        if command -v tfe &> /dev/null; then
            TFE_BIN="$(command -v tfe)"
        else
            echo "Error: TFE binary not found"
            echo "Please ensure TFE is installed and in your PATH"
            return 1
        fi
    else
        TFE_BIN=$NEWEST_TFE
    fi

    # Warn about version mismatches
    if [ "$VERSION_MISMATCH" = true ] && [ "$FOUND_COUNT" -gt 1 ]; then
        echo "⚠️  Warning: Multiple different versions of TFE found!"
        echo "   Run './build.sh' in your TFE directory to sync all installations"
        echo "   Using: $TFE_BIN"
        echo ""
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
