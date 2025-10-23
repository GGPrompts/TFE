---
description: Monitor TFE in tmux session with full TUI visibility
---

# Watch TFE in Tmux Session

You are monitoring TFE running in a tmux session, giving you visibility into the exact visual output the user sees.

## Your Task

Use Desktop Commander's `execute_command` tool to monitor a tmux session running TFE.

### Step 1: Identify the Session

Ask the user:
- "What's your tmux session name?" (e.g., "tfe-dev", "0", "main")
- Or detect automatically with: `tmux list-sessions`

### Step 2: Verify and Identify the Pane

```bash
# List all sessions
tmux list-sessions

# List windows in the session
tmux list-windows -t SESSION_NAME

# List panes in a window
tmux list-panes -t SESSION_NAME:0

# Get pane details
tmux display-message -t SESSION_NAME -p '#{session_name}:#{window_index}.#{pane_index}'
```

### Step 3: Capture Initial State

```bash
# Capture last 100 lines from the pane
tmux capture-pane -t SESSION_NAME -p -S -100

# For specific window/pane:
tmux capture-pane -t SESSION_NAME:WINDOW.PANE -p -S -100
```

**Parse the output for:**
- Current directory TFE is browsing
- Files visible in the list
- Preview content (if in preview mode)
- Any error messages or panics
- Bubbletea rendering artifacts
- Current display mode (list/grid/detail/tree)

### Step 4: Continuous Monitoring

Every 15-20 seconds (or when user asks), capture fresh output:

```bash
# Get latest 50 lines
tmux capture-pane -t SESSION_NAME -p -S -50

# Compare with previous capture
# Alert on changes:
#   - New errors appeared
#   - TFE crashed (pane shows shell prompt)
#   - Directory changed
#   - Different file selected
```

### Step 5: Parse TFE's TUI Output

**Look for:**
- **Title bar**: Shows current path
- **File list**: What files are visible
- **Status bar**: Current mode, file count
- **Preview pane**: What file content is shown
- **Error indicators**:
  - Go panic stack traces
  - "panic: runtime error"
  - "fatal error:"
  - Bubbletea crash messages

**Example parsing:**
```
Top line: /home/matt/projects/TFE
File list: Contains "main.go", "types.go", etc.
Preview: Showing README.md content
Status: Detail mode, 15 files
```

### Step 6: Proactive Assistance

**If you detect:**

**The user navigated to a file with issues:**
```
Tmux shows: preview of broken.go
You: "I see you're viewing broken.go. I notice there's a
     nil pointer dereference on line 45. Want me to fix it?"
```

**The user is browsing a directory:**
```
Tmux shows: /home/matt/projects/TFE in tree view
You: "I see you're in tree view. Just FYI, expanding
     node_modules might be slow - we could add lazy loading."
```

**An error occurred:**
```
Tmux shows: panic: index out of range
You: "⚠️ TFE crashed! Panic in getCurrentFile() when cursor
     exceeded file count. Here's the fix..."
```

## Advanced Tmux Features

### Monitor Multiple Panes

If user has split panes (TFE in one, logs in another):
```bash
# Capture all panes
tmux list-panes -t SESSION_NAME -F '#{pane_index}'
tmux capture-pane -t SESSION_NAME:0.0 -p  # TFE pane
tmux capture-pane -t SESSION_NAME:0.1 -p  # Log pane
```

### Send Commands to TFE (Advanced)

If user wants automation:
```bash
# Send keystrokes to the tmux pane
tmux send-keys -t SESSION_NAME "j" Enter  # Down arrow
tmux send-keys -t SESSION_NAME "F5"       # Switch mode
```

**Use cautiously!** Only with explicit user permission.

### Monitor tmux Status

```bash
# Check if pane is still alive
tmux list-panes -t SESSION_NAME

# If TFE crashed, pane will show shell prompt instead
```

## Communication Style

**Initial Report:**
```
🔍 Monitoring tmux session 'tfe-dev'
📍 Current: /home/matt/projects/TFE
📁 Viewing: 15 files in detail mode
👁️  Preview: README.md (42 lines)
✅ TFE running cleanly
```

**During Monitoring:**
```
🔄 Directory changed → /home/matt/projects
📄 Previewing main.go
```

**On Error:**
```
❌ PANIC DETECTED!
📍 Location: file_operations.go:234
🐛 Error: index out of range [5] with length 5
📋 Stack trace: [show first 10 lines]

Reading file_operations.go...
Found the issue: cursor not bounds-checked
Here's the fix: [show code]

Apply this fix?
```

## Real-World Example

```
User runs in terminal 1:
  tmux new -s tfe-dev
  cd ~/projects/TFE
  ./tfe

User in Claude Code (terminal 2):
  /watch-tmux
  > Session name: tfe-dev

Claude:
  🔍 Connected to tmux session 'tfe-dev'
  📍 TFE is browsing: /home/matt/projects/TFE
  📁 15 files visible, cursor on "main.go"
  👁️  Preview pane showing main.go (21 lines)
  ✅ Monitoring... I'll alert you to any issues

[User navigates to types.go in TFE]

Claude:
  📄 I see you're viewing types.go. The model struct
      is getting large (173 lines). Want me to suggest
      how we could split it into smaller pieces?
```

## Benefits of Tmux Monitoring

✅ **You control TFE** - Full interactive access
✅ **I see what you see** - Including TUI layout
✅ **Real-time awareness** - I know what you're looking at
✅ **Proactive help** - I can suggest things based on your actions
✅ **Non-intrusive** - I'm watching, not controlling
✅ **Works with manual testing** - Perfect for exploratory work

## Start Monitoring

Ask for the tmux session name and begin monitoring now.
