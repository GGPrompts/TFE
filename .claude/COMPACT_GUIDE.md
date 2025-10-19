# Compact Feature - Complete Guide

## What is Compact?

The **compact** feature lets you summarize a long Claude Code session and continue in a fresh conversation without losing context. This is similar to OpenCode's auto-compact feature, but implemented as a manual slash command with helper scripts.

## Why Use Compact?

✅ **Context window getting full** - Free up tokens for new work
✅ **Session getting slow** - Fresh sessions are faster
✅ **Long conversation** - Easier to work with compact summary
✅ **End of day** - Summarize progress before closing
✅ **Task switching** - Clean slate while preserving context

## How It Works

### **The Two-Part System:**

1. **`/compact` slash command** - Runs inside Claude Code, creates summary
2. **Helper scripts** - Run outside Claude Code, handle reload

This separation is necessary because slash commands can't execute `/clear` or restart Claude Code.

## Usage Workflows

### **Option 1: Fully Automated** ⭐ Recommended

```bash
# Inside Claude Code session:
/compact

# After summary is created, exit Claude Code (Ctrl+D or type 'exit')

# In your shell:
compact-reload

# This will:
# 1. Show summary preview (first 10 lines)
# 2. Ask for confirmation
# 3. Start new Claude session with summary loaded
```

**Pros:**
- Fastest method
- No manual copy/paste
- Preview before committing

---

### **Option 2: Manual with Clipboard**

```bash
# Inside Claude Code:
/compact

# Exit Claude Code

# In your shell:
show-compact

# This displays the summary and copies to clipboard

# Start Claude Code again:
claude

# Inside new session:
# Just paste (Ctrl+Shift+V)
```

**Pros:**
- Can review summary before using
- Can edit summary if needed
- More control over the process

---

### **Option 3: Fully Manual**

```bash
# Inside Claude Code:
/compact

# Copy the summary that Claude displays

# Inside same Claude session:
/clear

# Paste the summary
```

**Pros:**
- Everything in one Claude session
- No need to exit
- Immediate continuation

**Cons:**
- Manual copy/paste
- Claude still has some context (not fully fresh)

## The Summary Format

When you run `/compact`, Claude creates a structured summary:

```markdown
# Session Summary - 2025-10-18 20:00

## What We Accomplished
- [Main achievements]
- [Features implemented]
- [Problems solved]

## Files Modified
### Created:
- `path/to/file.ext` - What it does

### Modified:
- `path/to/file.ext` - What changed

## Key Technical Details
- [Important implementation notes]
- [Architecture decisions]
- [New patterns introduced]

## Current State
- ✅ Working: [What's functional]
- 🔄 In Progress: [What's partial]
- ⚠️ Issues: [Known problems]

## Next Steps
1. [Priority 1]
2. [Priority 2]
3. [Priority 3]

## Important Context
- [Key insights]
- [Gotchas]
- [Dependencies]
```

## Technical Details

### **File Locations:**

```bash
# Summary saved here:
/tmp/claude-compact-summary.md

# Slash command:
.claude/commands/compact.md

# Helper scripts:
scripts/compact-reload.sh → ~/bin/compact-reload
scripts/show-compact.sh → ~/bin/show-compact
```

### **How compact-reload Works:**

```bash
# 1. Check if summary exists
[ -f /tmp/claude-compact-summary.md ]

# 2. Show preview
head -n 10 /tmp/claude-compact-summary.md

# 3. Ask for confirmation
read -p "Continue? (y/N)"

# 4. Start Claude with summary as initial prompt
claude "I'm continuing from a previous session. Here's the summary:

$(cat /tmp/claude-compact-summary.md)

Ready to continue from where we left off."
```

### **How show-compact Works:**

```bash
# 1. Display the summary
cat /tmp/claude-compact-summary.md

# 2. Try to copy to clipboard (WSL/Linux/Mac)
cat /tmp/claude-compact-summary.md | clip.exe  # WSL
cat /tmp/claude-compact-summary.md | xclip -selection clipboard  # Linux
cat /tmp/claude-compact-summary.md | pbcopy  # Mac
```

## Customization

### **Modify Summary Format:**

Edit `.claude/commands/compact.md` to change what's included in summaries:

```bash
vim .claude/commands/compact.md
```

You can:
- Add new sections
- Change the structure
- Adjust verbosity
- Add project-specific details

### **Change Summary Location:**

Edit both scripts to use a different path:

```bash
# In compact-reload.sh and show-compact.sh:
SUMMARY_FILE="/your/custom/path/summary.md"
```

### **Add Auto-Cleanup:**

Uncomment this line in `compact-reload.sh`:

```bash
# Clean up old summary after successful reload (optional)
rm "$SUMMARY_FILE"
```

## Comparison with OpenCode

| Feature | OpenCode Auto-Compact | TFE Compact |
|---------|----------------------|-------------|
| **Trigger** | Automatic at 95% context | Manual via /compact |
| **Summary** | System-generated | Claude-generated |
| **Reload** | Automatic | Semi-automatic script |
| **Customization** | Limited | Fully customizable |
| **File saved** | Internal | /tmp (accessible) |
| **Review before use** | No | Yes |

**Advantages of TFE approach:**
- ✅ Full control over when to compact
- ✅ Can review and edit summary
- ✅ Customize summary format
- ✅ Choose reload method
- ✅ Summary file accessible for other uses

**Advantages of OpenCode approach:**
- ✅ Fully automatic
- ✅ Never hit context limits
- ✅ Zero user intervention

## Troubleshooting

### **"Summary file not found"**

```bash
# Check if summary exists:
ls -lh /tmp/claude-compact-summary.md

# If missing, run /compact in Claude Code first
```

### **"compact-reload: command not found"**

```bash
# Check if ~/bin is in PATH:
echo $PATH | grep -o "$HOME/bin"

# If not, add to ~/.bashrc or ~/.zshrc:
export PATH="$HOME/bin:$PATH"

# Reload shell:
source ~/.bashrc
```

### **Script not executable**

```bash
# Make scripts executable:
chmod +x ~/projects/TFE/scripts/*.sh
```

### **Summary too large**

If the summary is massive (>10KB), Claude might have included too much detail. You can:

1. Edit the summary manually: `vim /tmp/claude-compact-summary.md`
2. Ask Claude to create a more concise version
3. Modify the `/compact` command prompt to request brevity

## Best Practices

### **When to Compact:**

✅ **Good times:**
- After completing a major feature
- When conversation >100 messages
- Context window warning appears
- Before switching tasks
- End of work session

❌ **Avoid compacting:**
- Mid-debugging session
- During active problem-solving
- When referencing lots of history
- Very short sessions (<20 messages)

### **Summary Quality:**

**For best summaries:**
- Let Claude include all modified files
- Keep technical details specific
- Preserve command outputs
- Note any workarounds or gotchas
- Include next steps clearly

**Review before using:**
- Check file paths are correct
- Verify important context included
- Make sure next steps are clear
- Add any missing details

## Advanced Usage

### **Compact + Git Commit:**

```bash
# Use summary for git commit message
/compact

# In shell:
git commit -F /tmp/claude-compact-summary.md
```

### **Compact + Daily Log:**

```bash
# Append to daily log
/compact

# In shell:
cat /tmp/claude-compact-summary.md >> ~/logs/$(date +%Y-%m-%d).md
```

### **Compact + Share Context:**

```bash
# Share progress with team
/compact

# In shell:
cat /tmp/claude-compact-summary.md | mail -s "TFE Progress" team@example.com
```

## Example Session

```bash
# 1. Long development session in Claude Code
You: /watch-tmux
You: /rebuild-tfe
[... lots of work ...]
[... conversation getting long ...]

# 2. Create compact summary
You: /compact

Claude: [Creates detailed summary]
        ✅ Summary saved to /tmp/claude-compact-summary.md

        To continue in fresh session:
        Option 1: compact-reload
        Option 2: show-compact + manual paste
        Option 3: /clear + paste

# 3. Exit Claude Code
You: exit

# 4. Review and reload
$ show-compact
# [Shows summary and copies to clipboard]

$ compact-reload
🔄 Claude Code Compact & Reload

✅ Found summary file
Preview (first 10 lines):
─────────────────────────────────────────────
# Session Summary - 2025-10-18 20:54

## What We Accomplished
- Set up Desktop Commander MCP globally
- Created 6 powerful slash commands
- Implemented tmux monitoring for TFE
- Added /compact feature with helper scripts
─────────────────────────────────────────────

Summary: 87 lines, 4521 bytes

Continue with this summary? (y/N) y

🚀 Starting new Claude Code session with summary...

# 5. New Claude session starts with summary pre-loaded
Claude: I'm continuing from a previous session. Here's the summary:

[Full summary displayed]

Ready to continue from where we left off. What would you like to work on?

You: Let's test the /watch-tmux command now!
```

## Summary

The **compact** feature gives you OpenCode-like session summarization with more control and customization. Use `/compact` when your session gets long, then choose your preferred reload method:

- **Fast & automated:** `compact-reload`
- **Review first:** `show-compact` → paste
- **Stay in session:** `/clear` → paste

All methods preserve your work context while giving you a fresh, fast Claude Code session! 🚀
