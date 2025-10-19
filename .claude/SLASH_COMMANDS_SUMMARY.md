# TFE Slash Commands - Quick Summary

## ⚠️ Important: Built-in vs Custom Commands

### Built-in Claude Code Commands (Use These First!)

- **`/compact`** - Compresses current conversation in-place (RECOMMENDED)
- **`/clear`** - Resets conversation completely  
- **`/help`** - Lists all available commands
- **`/agents`** - Manage AI subagents
- **`/review`** - Request code review
- **`/exit`** - End session

### Custom TFE Commands (Our Project-Specific Tools)

#### Session Management
- **`/save-session`** - Export summary to file for external reload
  - Use when you want a completely fresh Claude instance
  - Use when you want to save progress for later
  - Creates `/tmp/claude-session-summary.md`
  - Pair with `session-reload` script

#### Monitoring
- **`/watch-tmux`** - Monitor TFE running in tmux session
- **`/tail-logs`** - Monitor TFE (tmux or background)

#### Development
- **`/rebuild-tfe`** - Full rebuild and restart cycle
- **`/test-tfe-feature`** - Interactive feature testing

#### Analysis
- **`/analyze-tfe-perf`** - Performance analysis

#### Setup
- **`/setup-tfe-logging`** - Add event logging to TFE

## When to Use What?

### Quick Reset (Stay in Session)
```bash
# In Claude Code:
/compact
```
✅ Fastest
✅ Stays in same session
✅ Claude's native feature

### Complete Fresh Start (New Session)
```bash
# In Claude Code:
/save-session

# Exit Claude (Ctrl+D)

# In shell:
session-reload
```
✅ Brand new Claude instance
✅ Maximum context reset
✅ Can review summary first

### Manual Control
```bash
# In Claude Code:
/save-session

# Exit Claude

# In shell:
show-session  # View & copy

# Start Claude and paste
claude
```
✅ Full control
✅ Can edit summary
✅ Review before using

## Global Helper Commands

Available from any directory:

```bash
session-reload    # Auto-reload Claude with session summary
show-session      # Display summary and copy to clipboard
```

## Recommendations

1. **For routine resets**: Use built-in `/compact`
2. **For fresh starts**: Use `/save-session` + `session-reload`
3. **For tmux monitoring**: Use `/watch-tmux`
4. **For rebuilds**: Use `/rebuild-tfe`

## File Locations

- **Session Summary**: `/tmp/claude-session-summary.md`
- **Slash Commands**: `.claude/commands/*.md`
- **Helper Scripts**: `scripts/*.sh` → `~/bin/*`
- **Documentation**: `.claude/COMPACT_GUIDE.md` (needs update)

