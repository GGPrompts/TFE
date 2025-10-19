#!/bin/bash
# Launch TFE in a tmux split pane
# Usage: ./launch-tfe-tmux.sh

cd "$(dirname "$0")"

# Build first if needed
if [ ! -f "./tfe" ]; then
    echo "Building TFE..."
    go build
fi

# Check if we're in a tmux session
if [ -n "$TMUX" ]; then
    # We're in tmux - create a vertical split in the current pane
    tmux split-window -v -c "#{pane_current_path}" "./tfe"
    echo "✅ TFE launched in tmux split pane below!"
else
    # Not in tmux - start a new session
    tmux new-session -d -s tfe -c "$PWD" "./tfe"
    echo "✅ TFE launched in new tmux session 'tfe'"
    echo "   Attach with: tmux attach -t tfe"
    echo "   (You're not in tmux, so I created a detached session)"
fi
