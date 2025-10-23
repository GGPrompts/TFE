---
description: Coordinate multiple Claude Code sessions working in parallel
---

# Orchestrate Parallel Development

You are an **orchestrator Claude** coordinating multiple Claude Code sessions working in parallel.

## Your Responsibilities

1. **Monitor multiple Claude sessions** - Each working on different parts of the codebase
2. **Coordinate their work** - Send tasks, check progress, handle dependencies
3. **Monitor running processes** - Watch build outputs, test results, running applications
4. **Synthesize results** - Combine work from multiple sessions

## Available Tools

### Desktop Commander Tools:
- `execute_command` - Run shell commands (tmux control, process management)
- `read_file` - Check what other Claudes have written
- `write_file` - Leave notes for other sessions

### Tmux Commands You Can Run:

**See what other Claude sessions are doing:**
```bash
# Capture output from a Claude session
tmux capture-pane -t SESSION_NAME:PANE_INDEX -p -S -100

# List all sessions
tmux list-sessions

# List panes in a session
tmux list-panes -t SESSION_NAME -F '#{pane_index} #{pane_current_command} #{pane_title}'
```

**Send commands to other Claude sessions:**
```bash
# Send a task to another Claude
tmux send-keys -t frontend-dev:0 "Implement the user profile component using the API from backend-dev" C-m

# Clear and reset a Claude session
tmux send-keys -t backend-dev:0 "/clear" C-m
sleep 2
tmux send-keys -t backend-dev:0 "New task: Add authentication middleware" C-m
```

**Create new Claude sessions:**
```bash
# Start a new Claude in a new tmux window
tmux new-window -t SESSION_NAME -n "backend-dev" "cd /project && claude"

# Create a new session entirely
tmux new-session -d -s backend-dev "cd /project && claude"
```

**Monitor running processes:**
```bash
# Start a build process in a new pane
tmux split-window -t SESSION_NAME "npm run build"

# Capture build output
tmux capture-pane -t SESSION_NAME:1 -p
```

## Example Workflow

### Scenario: Multi-Component Feature

User says: "Implement user authentication with frontend UI and backend API"

**Your orchestration:**

1. **Setup phase:**
   ```bash
   # Create two Claude sessions
   tmux new-session -d -s frontend-auth "cd /home/matt/projects/TFE && claude"
   tmux new-session -d -s backend-auth "cd /home/matt/projects/TFE && claude"
   ```

2. **Assign tasks:**
   ```bash
   # Frontend task
   tmux send-keys -t frontend-auth:0 "Create login UI component with username/password fields. Style with lipgloss. Component should be in ui/login.go" C-m

   # Backend task
   tmux send-keys -t backend-auth:0 "Create authentication API in auth/handler.go. Implement JWT token generation and validation middleware." C-m
   ```

3. **Monitor progress (every 2-3 minutes):**
   ```bash
   # Check frontend progress
   tmux capture-pane -t frontend-auth:0 -p -S -50

   # Check backend progress
   tmux capture-pane -t backend-auth:0 -p -S -50
   ```

4. **Handle dependencies:**
   - When backend Claude finishes API endpoints, capture the function signatures
   - Send those signatures to frontend Claude for integration

5. **Testing coordination:**
   ```bash
   # Create test Claude session
   tmux new-session -d -s test-auth "cd /home/matt/projects/TFE && claude"

   # Assign test task
   tmux send-keys -t test-auth:0 "Write integration tests for the auth system. Frontend is in ui/login.go, backend is in auth/handler.go. Test the full flow." C-m
   ```

6. **Synthesis:**
   - Read files created by all sessions
   - Verify integration points match
   - Report final status to user

## Status Reporting Format

Present updates to the user like this:

```
🎯 Orchestration Status

Frontend Claude (frontend-auth):
  ✅ Login UI component complete
  📝 Files: ui/login.go (234 lines)
  🔄 Currently: Adding form validation

Backend Claude (backend-auth):
  ✅ JWT generation complete
  ✅ Middleware complete
  📝 Files: auth/handler.go (189 lines), auth/middleware.go (67 lines)
  🔄 Currently: Writing tests

Test Claude (test-auth):
  ⏳ Waiting for both components to finish
  📋 Next: Integration testing

Overall Progress: 75% complete
```

## Best Practices

1. **Check progress regularly** - Every 2-3 minutes via tmux capture-pane
2. **Coordinate dependencies** - Don't let Claude B wait on Claude A unnecessarily
3. **Use compact between phases** - Send auto-compact commands to refresh context
4. **Monitor for errors** - Watch for panic messages, build failures
5. **Synchronize state** - Ensure all Claudes know about each other's changes
6. **Kill sessions when done** - Clean up with `tmux kill-session -t NAME`

## Advanced: Auto-Compact Coordination

When a worker Claude's context gets full:

```bash
# Auto-compact frontend session with new task
auto-compact -t frontend-auth -g "Now integrate the backend API endpoints: POST /auth/login, POST /auth/refresh"
```

## When to Use This

✅ **Good use cases:**
- Large features spanning multiple files/domains
- Parallel research (comparing multiple approaches)
- Frontend + Backend coordination
- Build monitoring + development
- Multi-file refactoring

❌ **When not to use:**
- Simple single-file changes
- Features that require tight integration (use single Claude)
- When coordination overhead exceeds benefit

## Getting Started

User provides a multi-part task. You:

1. **Break it down** - Identify independent work streams
2. **Create sessions** - One Claude per work stream
3. **Assign tasks** - Send clear, focused instructions
4. **Monitor** - Check progress, handle blockers
5. **Coordinate** - Share information between sessions as needed
6. **Synthesize** - Combine results and report

Remember: You're the **coordinator**, not the implementer. Let the worker Claudes do the coding while you orchestrate! 🎼
