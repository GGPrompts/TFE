# TFE Development Slash Commands

These slash commands leverage Desktop Commander MCP to provide powerful development workflows for TFE.

## 📋 Available Commands

**Legend:**
- 🎯 Core commands - Use these most often
- ⭐ Advanced features - Powerful when you need them

### `/tail-logs` - Monitor TFE Runtime
**Purpose:** Continuously monitor TFE for errors, panics, and issues during development.

**What it does:**
- Checks if TFE is running (tmux or background process)
- Monitors process output in real-time
- **Supports tmux monitoring** - can watch your interactive TFE session!
- Alerts you to errors, panics, stack traces
- Automatically reads and analyzes problematic code
- Suggests fixes proactively

**When to use:**
- While developing new features
- When debugging issues
- During manual testing sessions
- Anytime you want real-time error detection

**Example (tmux mode):**
```
You: /tail-logs
Claude: Are you running TFE in a tmux session?
You: Yes, session name is "tfe-dev"
Claude: ✅ Monitoring tmux session 'tfe-dev'
        📍 TFE browsing: /home/matt/projects/TFE
        👁️ Previewing: README.md
        No errors detected
[You navigate in TFE...]
Claude: 📄 I see you opened types.go. The model struct
        is getting large - want suggestions for refactoring?
```

**Example (background mode):**
```
You: /tail-logs
Claude: Are you running TFE in a tmux session?
You: No
Claude: ✅ Started TFE in background (PID 12345)
        Monitoring output...
[You work on code...]
Claude: ⚠️ Panic detected in file_operations.go:234!
        Nil pointer dereference in loadFiles()
        Here's the issue and a suggested fix...
```

---

### `/watch-tmux` - Advanced Tmux Monitoring ⭐ NEW
**Purpose:** Dedicated tmux session monitoring with full TUI visibility.

**What it does:**
- Captures and parses tmux pane output
- **Sees the actual Bubbletea TUI** - not just logs!
- Detects what directory you're browsing
- Knows what file you're previewing
- Monitors for crashes, errors, panics
- Can see visual rendering issues
- Provides context-aware suggestions

**When to use:**
- Manual testing with full control
- Want me to see what you're doing in TFE
- Need proactive assistance based on your actions
- Prefer interactive TFE over background process

**Example:**
```
You: Start TFE in tmux:
     tmux new -s tfe-dev
     ./tfe

You: In Claude: /watch-tmux
     Session name: tfe-dev

Claude: 🔍 Monitoring tmux 'tfe-dev'
        📍 Current: /home/matt/projects/TFE
        📁 15 files (detail mode)
        👁️ Preview: README.md
        ✅ Running cleanly

[You navigate to a Go file with an error]

Claude: ⚠️ I see you're viewing file_operations.go
        Line 234 has a potential nil pointer issue
        Want me to fix it?
```

---

### `/rebuild-tfe` - Full Rebuild Cycle
**Purpose:** Kill old process, rebuild, and restart TFE with error checking.

**What it does:**
- Kills existing TFE process
- Cleans build artifacts
- Runs `go build` with error capture
- Reports compilation issues with file:line references
- Starts new process if build succeeds
- Monitors initial startup for crashes

**When to use:**
- After making code changes
- When TFE won't respond
- To ensure you're running latest code
- Before testing new features

**Example:**
```
You: /rebuild-tfe
Claude: 🔄 Rebuilding TFE...
        • Killing old process... ✅
        • Building... ✅ (2.1s, 8.4 MB)
        • Starting... ✅ (PID 12456)
        🎯 TFE running cleanly!
```

---

### `/test-tfe-feature` - Interactive Testing Monitor
**Purpose:** Monitor TFE while you manually test a specific feature.

**What it does:**
- Asks what feature you're testing
- Starts/ensures TFE is running
- Monitors output during your manual testing
- Reports errors immediately
- Provides real-time debugging assistance
- Summarizes test results

**When to use:**
- Testing new features (tree view, preview modes, etc.)
- Verifying bug fixes
- Manual QA sessions
- Exploring edge cases

**Example:**
```
You: /test-tfe-feature
Claude: What feature are you testing?
You: Tree view folder expansion
Claude: ✅ TFE ready, monitoring started. Go ahead and test!
[You test tree view...]
Claude: ⚠️ Detected slow response (500ms) when expanding node_modules
        Might want to add lazy loading. Want me to investigate?
```

---

### `/analyze-tfe-perf` - Performance Analysis
**Purpose:** Analyze TFE codebase for performance bottlenecks and optimization opportunities.

**What it does:**
- Scans code for common performance issues
- Looks for O(n²) operations, redundant I/O, etc.
- Runs benchmarks if available
- Prioritizes issues by impact
- Suggests specific optimizations with code examples

**When to use:**
- Before releasing new features
- When TFE feels sluggish
- Planning optimization work
- Regular performance reviews

**Example:**
```
You: /analyze-tfe-perf
Claude: 🔍 Performance Analysis Complete

        📊 High Priority:
        1. loadFiles() calls os.ReadDir twice - 50% slower
        2. renderPreview() recreates styles every frame

        💡 Quick wins found in 5 locations
        Want me to implement the fixes?
```

---

### `/compact` - Summarize Session ⭐ NEW
**Purpose:** Create a concise summary of the current work session to continue in a fresh conversation.

**What it does:**
- Analyzes the entire conversation
- Summarizes what was accomplished
- Lists all files modified/created
- Captures key technical details
- Notes current state and next steps
- Saves summary to `/tmp/claude-compact-summary.md`
- Provides options for reloading

**When to use:**
- Context window getting full
- Session getting long and slow
- Want to start fresh but keep progress
- Before switching to a different task
- End of day summary

**Workflow:**
```
You: /compact

Claude: [Creates detailed summary]
        ✅ Summary saved to /tmp/claude-compact-summary.md

        To continue in fresh session:

        Option 1 (Automated):
          Exit Claude and run: compact-reload

        Option 2 (Manual):
          1. Run: /clear
          2. Paste summary below

        Option 3 (Review first):
          Run: show-compact

[Choose your preferred method]
```

**Automated reload:**
```bash
# After running /compact in Claude:
compact-reload

# This will:
# 1. Show summary preview
# 2. Ask for confirmation
# 3. Start new Claude session with summary pre-loaded
```

**Manual reload:**
```bash
# Show and copy summary
show-compact

# In Claude Code:
/clear
# Then paste the summary
```

---

### `/setup-tfe-logging` - Add Event Logging
**Purpose:** Automatically add debug event logging to TFE for real-time monitoring.

**What it does:**
- Adds debug log file to model
- Creates logging helper functions
- Adds logging calls throughout codebase:
  - File previews
  - Directory loads
  - Key presses
  - Render events
- Sets up JSON event stream to `/tmp/tfe-events.jsonl`

**When to use:**
- First time setup for advanced monitoring
- When you want me to see exactly what you're doing in TFE
- Enabling intelligent assistance based on your actions

**Example:**
```
You: /setup-tfe-logging
Claude: [Adds logging code to multiple files]
        ✅ Debug logging implemented!

        Test it with: TFE_DEBUG=1 ./tfe
        Monitor with: tail -f /tmp/tfe-events.jsonl

        Now I'll be able to see what files you're viewing
        and help proactively!
```

---

## 🎯 Tmux vs Background Process

### **When to use Tmux Monitoring:**
✅ **Best for:** Manual testing, exploratory work, learning
✅ **Benefit:** You see and control TFE directly
✅ **Bonus:** I can see the TUI output and your actions
✅ **Setup:**
```bash
tmux new -s tfe-dev
./tfe
# Detach: Ctrl+b, d
# Or keep it visible in a split pane
```

### **When to use Background Process:**
✅ **Best for:** Automated workflows, quick rebuilds
✅ **Benefit:** Hands-off monitoring while you code
✅ **Bonus:** I control process lifecycle (start/stop/restart)

## 🚀 Workflow Examples

### **Tmux Workflow (Recommended for Interactive Testing)**

```bash
# Terminal 1: Start tmux session with TFE
tmux new -s tfe-dev
cd ~/projects/TFE
./tfe

# Terminal 2: Claude Code
claude

# In Claude:
/watch-tmux
> Session: tfe-dev

# Now:
# - You interact with TFE in terminal 1
# - I watch and help in terminal 2
# - Best of both worlds! 🎉
```

### **Development Session with Full Monitoring**

```bash
# Terminal 1: Start development session
cd ~/projects/TFE
claude

# In Claude conversation:
/setup-tfe-logging   # First time only
/rebuild-tfe         # Build and start TFE
/tail-logs           # Start monitoring

# Now work on your code...
# Claude will alert you to any issues in real-time!
```

### **Feature Testing Workflow**

```bash
# Make code changes to tree view
vim render_file_list.go

# In Claude:
/rebuild-tfe
/test-tfe-feature
> "Testing tree view folder expansion"

# Manually test in TFE while Claude monitors
# Claude will catch any crashes, slowdowns, errors
```

### **Performance Optimization Session**

```bash
# In Claude:
/analyze-tfe-perf
> Claude finds 3 high-priority issues

"Apply the top 2 fixes"
> Claude applies optimizations

/rebuild-tfe
> Test performance improvement

"Run benchmarks"
> Verify the improvements
```

---

## 🎯 Power User Tips

### **Combine Commands**
```
/rebuild-tfe

[After it completes:]
/tail-logs
```

### **With Event Logging Enabled**
Once you've run `/setup-tfe-logging`, I can:
- See what files you're previewing
- Read them automatically
- Suggest improvements
- Answer questions about code you're looking at
- All without you asking!

### **Background Monitoring**
```
/tail-logs

# Now I'm monitoring in the background
# Continue our conversation normally
# I'll interrupt if I detect issues
```

---

## 📋 Quick Reference Card

```bash
# Session management
/compact              # Summarize work and prepare for fresh session

# Monitoring (pick one based on workflow)
/watch-tmux           # Monitor tmux session (best for interactive testing)
/tail-logs            # Monitor tmux OR background process (flexible)

# Development workflow
/rebuild-tfe          # Build and start TFE
/test-tfe-feature     # Test specific features

# Performance work
/analyze-tfe-perf     # Find bottlenecks

# Advanced setup
/setup-tfe-logging    # Add event logging (one-time)
```

### **Shell Helper Commands**
```bash
compact-reload        # Auto-reload Claude with compact summary
show-compact          # Display summary and copy to clipboard
```

### **Quick Start: Tmux + Claude Monitoring** 🚀
```bash
# Terminal 1: Start TFE in tmux
tmux new -s tfe-dev
./tfe

# Terminal 2: Start Claude Code
claude

# In Claude conversation:
/watch-tmux
> tfe-dev

# Done! I can now see what you're doing in TFE! 🎉
```

---

## 📝 Notes

- All these commands use **Desktop Commander MCP** tools
- Desktop Commander must be configured and available
- Event logging requires TFE code modifications (via `/setup-tfe-logging`)
- Log files are written to `/tmp/` (cleaned on reboot)

---

## 🔧 Customization

Feel free to modify these commands! They're just markdown files in `.claude/commands/`.

To create your own:
1. Create `.claude/commands/my-command.md`
2. Write a detailed prompt describing the task
3. Use `/my-command` to invoke it

The prompt should:
- Be specific about what tools to use
- Define expected output format
- Handle error cases
- Provide examples when helpful

---

## 🆘 Troubleshooting

**Command not found:**
- Check the file exists in `.claude/commands/`
- Restart Claude Code to reload commands

**Desktop Commander not available:**
- Verify: `claude mcp list`
- Should show: `desktop-commander: ✓ Connected`
- If not, run: `claude mcp add-json --scope user desktop-commander '{"command": "npx", "args": ["-y", "@wonderwhy-er/desktop-commander@latest"]}'`

**TFE won't start:**
- Run `/rebuild-tfe` to see build errors
- Check permissions on `./tfe`
- Verify you're in the TFE project directory
