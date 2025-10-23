---
description: Monitor TFE while manually testing a specific feature
---

# Test TFE Feature

You are helping test a specific feature in TFE by monitoring the application during manual testing.

## Your Task

Use bash and tmux to safely monitor TFE while the user manually tests a feature.

1. **Prepare for testing**
   - Rebuild TFE if needed (ask me first)
   - Launch TFE in a detached tmux session (not via start_process to avoid terminal corruption)
   - Use tmux capture-pane to monitor output
   - Provide the attach command so user can interact with TFE in a new terminal

2. **Launch TFE safely**
   ```bash
   # Kill any existing tfe-test session
   tmux kill-session -t tfe-test 2>/dev/null || true

   # Launch TFE in detached tmux session
   cd /home/matt/projects/TFE
   tmux new-session -d -s tfe-test -c "$PWD" "./tfe"

   # Verify it started
   tmux list-sessions | grep tfe-test
   ```

   Then output:
   ```
   ✅ TFE launched in tmux session 'tfe-test'

   📋 To interact with TFE, open a new terminal and run:
      tmux attach -t tfe-test

   (Press Ctrl+B then D to detach without closing TFE)
   ```

3. **Ask me what feature to test**
   - Example: "Testing tree view expansion"
   - Example: "Testing preview mode with large files"
   - Example: "Testing new keyboard shortcuts"

4. **Monitor during testing**
   Use `tmux capture-pane -t tfe-test -p -S -50` to capture output
   - Watch for panics, errors, crashes
   - Check for performance issues (slow responses)
   - Track any warnings or unusual output
   - Parse TFE's TUI to see what file/directory is selected

5. **Real-time reporting**
   - Alert me immediately if errors occur
   - Show me the exact error and location
   - Suggest what might have caused it
   - Recommend fixes
   - Periodically capture pane to see current state

6. **Post-test analysis**
   - Summarize what happened during the test
   - List any issues found
   - Suggest improvements
   - Offer to create TODO items for bugs found

7. **Interactive debugging**
   - If a crash occurs, read the relevant source files
   - Analyze the problematic code path
   - Suggest fixes
   - Ask if I want to apply them and re-test
   - Can kill and restart TFE session as needed

## Communication

- Keep me updated every 10-15 seconds if I'm actively testing
- Use clear status indicators: ✅ ⚠️ ❌ 🔍 🐛
- Show relevant code snippets when discussing errors
- Be proactive about suggesting fixes

Start by asking me what feature I want to test, then prepare the monitoring environment.
