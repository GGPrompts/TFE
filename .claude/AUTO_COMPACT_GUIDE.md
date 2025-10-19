# Auto-Compact: Automated In-Session Compacting

## The Brilliant Idea

Instead of exiting Claude and reloading, we can **automatically compact the session in-place** using tmux automation!

## How It Works

```
┌─────────────────────────────────────────────────────────┐
│ Claude Running in Tmux                                  │
│ ┌─────────────────────────────────────────────────────┐ │
│ │ > /save-session Fix the tree view bug              │ │ ← Script sends this
│ │                                                     │ │
│ │ [Claude creates summary...]                        │ │
│ │ ✅ Summary saved to /tmp/claude-session-summary.md │ │
│ │                                                     │ │
│ │ > /clear                                            │ │ ← Script sends this
│ │                                                     │ │
│ │ Conversation cleared.                               │ │
│ │                                                     │ │
│ │ > [Summary pasted here...]                          │ │ ← Script pastes
│ │                                                     │ │
│ │ Claude: Ready to work on: Fix the tree view bug!   │ │
│ └─────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
                    ↑
        Script controls everything automatically!
```

## Three Compact Methods Compared

### **Method 1: Built-in `/compact`** ⚡ Fastest
```bash
# In Claude:
/compact
```

**How it works:**
- Claude's native feature
- Compresses conversation in-place
- No external tools needed

**Pros:**
- ✅ Instant (1 second)
- ✅ No scripts needed
- ✅ Stays in session
- ✅ Official feature

**Cons:**
- ❌ Fixed format
- ❌ No goal support
- ❌ Can't customize

**Use when:** Quick reset, routine work

---

### **Method 2: Auto-Compact** 🤖 Automated (NEW!)
```bash
# In a separate terminal:
auto-compact -t claude-session -g "Fix tree view bug"
```

**How it works:**
1. Script sends `/save-session` to your Claude tmux pane
2. Waits for summary to be created (~5-10 seconds)
3. Sends `/clear`
4. Pastes summary back
5. All automatic!

**Pros:**
- ✅ Fully automated
- ✅ Stays in same session
- ✅ Custom summary format
- ✅ Goal support
- ✅ Can trigger from anywhere
- ✅ No need to switch to Claude window

**Cons:**
- ⚠️ Requires tmux
- ⚠️ ~10-15 second process
- ⚠️ Needs separate terminal

**Use when:**
- Claude is in tmux
- Want automation
- Want custom format + goals
- Compacting from a different window

---

### **Method 3: Session Reload** 🔄 Complete Fresh Start
```bash
# In Claude:
/save-session Fix tree view bug

# Exit (Ctrl+D)

# In shell:
session-reload
```

**How it works:**
1. Creates summary in Claude
2. Exit completely
3. Start brand new Claude process
4. Load with summary

**Pros:**
- ✅ Completely fresh Claude instance
- ✅ Maximum context reset
- ✅ New process = max performance
- ✅ Can review summary first

**Cons:**
- ⚠️ Must exit Claude
- ⚠️ New process startup (~2-3 seconds)
- ⚠️ More manual steps

**Use when:**
- Want completely fresh start
- Session is truly huge
- Performance is degrading
- End of day/next day resumption

---

## Auto-Compact Usage

### **Basic (Auto-detect)**
```bash
# Script finds Claude automatically
auto-compact
```

### **With Specific Session**
```bash
auto-compact -t my-claude-session
```

### **With Goal**
```bash
auto-compact -g "Implement syntax highlighting"
```

### **Full Options**
```bash
auto-compact -t claude-dev -p 0 -g "Debug tree view cursor positioning"
```

## Real-World Workflows

### **Workflow 1: Two-Terminal Setup**

```
Terminal 1 (tmux):              Terminal 2:
┌──────────────────┐            ┌──────────────────┐
│ tmux session     │            │ Development      │
│                  │            │                  │
│ [Claude Code]    │            │ $ vim main.go    │
│ > Working...     │            │ $ git commit     │
│                  │            │                  │
│                  │            │ [Context full]   │
│                  │            │                  │
│                  │            │ $ auto-compact   │
│                  │            │   -t claude-dev  │
│ [Auto /clear]    │◄───────────┤                  │
│ [Auto paste]     │            │ $ # Done!        │
│                  │            │                  │
│ > Fresh context  │            │ $ continue work  │
└──────────────────┘            └──────────────────┘
```

**Benefit:** Never leave your work terminal! 🎯

---

### **Workflow 2: Tmux Split Panes**

```
┌─────────────────────────────────────────────────┐
│ Tmux Session                                    │
│ ┌──────────────┬────────────────────────────┐  │
│ │ Pane 0       │ Pane 1                     │  │
│ │              │                            │  │
│ │ [Claude Code]│ $ auto-compact -p 0        │  │
│ │              │   ✅ Compacted!            │  │
│ │              │                            │  │
│ │              │ $ # Continue development   │  │
│ └──────────────┴────────────────────────────┘  │
└─────────────────────────────────────────────────┘
```

**Setup:**
```bash
tmux new -s dev
tmux split-window -h
# Left pane: claude
# Right pane: auto-compact -p 0
```

---

### **Workflow 3: Scripted Development Flow**

Create `~/bin/dev-compact`:
```bash
#!/bin/bash
# Compact Claude, then run tests

auto-compact -t dev-session -g "Continue from where we left off"
sleep 5
tmux send-keys -t dev-session:1 "go test ./..." C-m
```

**Usage:**
```bash
dev-compact
# → Claude compacted
# → Tests running
# → Back to coding!
```

---

## Advanced Examples

### **Example 1: Context Switch with Goal**

```bash
# You're working on feature A in Claude
# Boss says: "Fix production bug NOW!"

# Compact with context for later:
auto-compact -g "Resume: Implementing preview variable highlighting, was about to add the input fields UI"

# Now use Claude for the bug
# Later, compact again:
auto-compact -g "Back to feature A: Add input fields below preview"

# Claude picks right back up!
```

---

### **Example 2: End of Day Automation**

Add to `~/.bashrc`:
```bash
alias eod='auto-compact -g "Tomorrow: Continue from where we left off" && tmux detach'
```

**Usage:**
```bash
# At end of day:
eod

# Tomorrow:
tmux attach
# Claude is ready with yesterday's context!
```

---

### **Example 3: Timed Auto-Compact**

```bash
#!/bin/bash
# Auto-compact Claude every 2 hours

while true; do
    sleep 7200  # 2 hours
    auto-compact -t claude-main -g "Continue from where we left off"
    notify-send "Claude Compacted" "Context refreshed!"
done
```

**Benefit:** Never hit context limits! 🚀

---

## How Auto-Compact Actually Works

### **Behind the Scenes:**

```bash
# 1. Send command to tmux pane
tmux send-keys -t claude-session:0 "/save-session Fix bug" C-m

# 2. Wait for Desktop Commander to create summary
while [ ! -f /tmp/claude-session-summary.md ]; do
    sleep 1
done

# 3. Clear conversation
tmux send-keys -t claude-session:0 "/clear" C-m
sleep 1

# 4. Load summary into tmux buffer
tmux load-buffer /tmp/claude-session-summary.md

# 5. Paste into Claude
tmux paste-buffer -t claude-session:0

# 6. Submit
tmux send-keys -t claude-session:0 C-m
```

**Key Technical Details:**
- Uses `tmux send-keys` for typing simulation
- Uses `tmux load-buffer` + `paste-buffer` for reliable large text pasting
- Polls summary file to know when generation is complete
- Adds delays for Claude to process each step

---

## Comparison Matrix

| Feature | /compact | auto-compact | session-reload |
|---------|----------|--------------|----------------|
| **Speed** | ⚡ 1s | 🔄 10-15s | 🐢 5-20s |
| **Automation** | Manual | Automatic | Semi-auto |
| **Stays in Session** | ✅ | ✅ | ❌ |
| **Goal Support** | ❌ | ✅ | ✅ |
| **Custom Format** | ❌ | ✅ | ✅ |
| **Requires Tmux** | ❌ | ✅ | ❌ |
| **Fresh Instance** | ❌ | ❌ | ✅ |
| **Trigger from Anywhere** | ❌ | ✅ | ❌ |

---

## Troubleshooting

### **"Could not auto-detect Claude session"**
```bash
# List sessions:
tmux list-sessions

# Specify manually:
auto-compact -t YOUR_SESSION_NAME
```

### **"Summary file not created"**
- Check Claude responded to `/save-session`
- Look in tmux pane to see any errors
- Desktop Commander might not be available

### **Paste didn't work**
- Large summaries might fail
- Try manually: `cat /tmp/claude-session-summary.md`
- Check tmux buffer size: `tmux show-buffer`

### **Wrong pane**
```bash
# List panes:
tmux list-panes -t SESSION_NAME

# Specify pane:
auto-compact -t SESSION_NAME -p PANE_INDEX
```

---

## Tips & Tricks

### **Alias for Quick Compact**
```bash
alias ac='auto-compact'
alias acg='auto-compact -g'

# Usage:
ac                              # Quick compact
acg "Fix tree view bug"        # With goal
```

### **Integration with Git Workflow**
```bash
# Before committing:
auto-compact -g "Committed feature X, now working on feature Y"
git add .
git commit -m "Implement feature X"
```

### **Tmux Key Binding**
Add to `~/.tmux.conf`:
```
bind-key C run-shell "auto-compact -t #S -p #{pane_index}"
```

**Usage:** Press `Prefix + Shift+C` to auto-compact current pane!

---

## Summary

**Auto-compact is the best of both worlds:**

✅ **Fast** - Automated in ~10 seconds
✅ **Convenient** - No need to exit Claude
✅ **Powerful** - Custom format + goals
✅ **Smart** - Works from any terminal
✅ **Flexible** - Auto-detect or specify session

**Perfect for:**
- Active development sessions
- Tmux-based workflows
- Frequent compacting needs
- Multi-terminal setups
- Automated workflows

**Use it when:** You want the benefits of `session-reload` without leaving your Claude session! 🎯
