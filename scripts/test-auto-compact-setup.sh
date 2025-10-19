#!/bin/bash
# Quick tmux + Claude setup for testing auto-compact

echo "Setting up tmux session for Claude Code..."
echo ""
echo "This will:"
echo "  1. Create a tmux session called 'claude-test'"
echo "  2. Start Claude Code in that session"
echo "  3. Show you how to use auto-compact from another terminal"
echo ""
read -p "Press Enter to continue..."

# Create tmux session and start Claude
tmux new-session -d -s claude-test
tmux send-keys -t claude-test "cd /home/matt/projects/TFE" C-m
tmux send-keys -t claude-test "claude" C-m

echo ""
echo "✅ Tmux session 'claude-test' created!"
echo ""
echo "Now do this:"
echo ""
echo "1. Attach to the session:"
echo "   tmux attach -t claude-test"
echo ""
echo "2. Talk to Claude normally in that session"
echo ""
echo "3. Open a SECOND terminal and run:"
echo "   auto-compact -t claude-test -g \"Testing auto-compact\""
echo ""
echo "4. Watch as Claude automatically:"
echo "   - Runs /save-session"
echo "   - Runs /clear"
echo "   - Pastes the summary back"
echo ""
echo "Detach from tmux with: Ctrl+B then D"
echo "Kill the session with: tmux kill-session -t claude-test"
