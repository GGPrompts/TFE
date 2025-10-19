Launch TFE in a tmux split pane so we can monitor it together.

```bash
cd /home/matt/projects/TFE

# Check if we're in a tmux session
if [ -n "$TMUX" ]; then
    # We're in tmux - create a vertical split
    tmux split-window -v -c "#{pane_current_path}" "./tfe"
    echo "✅ TFE launched in tmux split pane!"
else
    # Not in tmux - start a new session
    tmux new-session -d -s tfe -c "$PWD" "./tfe"
    echo "✅ TFE launched in new tmux session 'tfe'"
    echo "   Attach with: tmux attach -t tfe"
fi
```
