---
description: Monitor TFE application for errors, panics, and issues in real-time
---

# Monitor TFE Application Logs

You are monitoring the TFE (Terminal File Explorer) application for errors, panics, and issues.

## Your Task

Use Desktop Commander MCP tools to:

1. **Check monitoring method**
   - First, ask if the user is running TFE in a tmux session
   - If yes, ask for the tmux session name (e.g., "tfe-dev")
   - If no tmux, check for background processes or offer to start one

2. **TMUX Monitoring (Preferred Method)**
   - If monitoring a tmux session:
     - Use `execute_command` to run: `tmux list-sessions` to verify it exists
     - Use `tmux capture-pane -t SESSION_NAME -p -S -100` to capture last 100 lines
     - Parse the output for errors, panics, TUI rendering issues
     - Can see the actual Bubbletea visual output!
     - Periodically capture new output (every 10-15 seconds when actively monitoring)
     - Use `tmux capture-pane -t SESSION_NAME:WINDOW.PANE -p -S -50` for specific panes

   **Tmux Commands Reference:**
   ```bash
   # List sessions
   tmux list-sessions

   # List windows in a session
   tmux list-windows -t SESSION_NAME

   # List panes in a window
   tmux list-panes -t SESSION_NAME:WINDOW

   # Capture pane output (last N lines)
   tmux capture-pane -t SESSION_NAME -p -S -N

   # Capture entire scrollback
   tmux capture-pane -t SESSION_NAME -p -S -
   ```

3. **Background Process Monitoring (Alternative)**
   - Use `list_processes` to find TFE process
   - If not running, offer to start it with `start_process("./tfe")`
   - If running as a background process, use `read_process_output` to check output
   - Look for any log files in `/tmp/tfe-*.log` or similar

4. **Monitor for errors, panics, warnings**

3. **Watch for these issues:**
   - Go panic messages
   - Runtime errors
   - Stack traces
   - Bubbletea rendering issues
   - File operation errors
   - Segmentation faults
   - Build failures

4. **Continuous monitoring:**
   - Check output periodically (you can read latest lines with negative offset)
   - Alert me immediately if you detect any issues
   - Provide the full error message and stack trace
   - Identify which source file and line caused the issue
   - Suggest potential fixes based on the error

5. **Proactive analysis:**
   - If you see an error in a specific file (e.g., `file_operations.go:234`):
     - Read that file
     - Analyze the problematic code
     - Suggest a fix
     - Ask if I want you to apply it

6. **Stay vigilant:**
   - Keep monitoring throughout our conversation
   - Report any new errors as they occur
   - Track if errors repeat or change

## Expected Workflow

### Tmux Monitoring:
```
1. Ask: "Are you running TFE in a tmux session?"
2. If yes: "What's the session name?"
3. Verify session exists: tmux list-sessions
4. Capture initial output: tmux capture-pane -t SESSION -p -S -100
5. Parse for errors/panics
6. Report: "✅ Monitoring tmux session 'tfe-dev'. TFE running cleanly" or "⚠️ Issues detected"
7. Continue monitoring periodically
8. Alert to any new errors immediately
```

### Background Process Monitoring:
```
1. Check if TFE is running
2. Start monitoring output/logs
3. Report current status: "✅ TFE running cleanly" or "⚠️ Issues detected"
4. Continue monitoring and alert me to any new issues
5. Provide file:line references for all errors
6. Suggest fixes proactively
```

## Communication Style

- Use emojis for quick status: ✅ 🔄 ⚠️ ❌ 🔍
- Be concise but thorough
- Show error context (a few lines around the issue)
- Always provide actionable information

Start monitoring now and report the current status.
