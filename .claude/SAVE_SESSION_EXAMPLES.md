# /save-session Command - Usage Examples

## Basic Usage

### **Example 1: No Specific Goal**

```bash
# In Claude Code:
/save-session
```

**What happens:**
1. Claude creates summary of current work
2. Saves to `/tmp/claude-session-summary.md`
3. Summary ends with: "No specific goal set - ready to continue general development"

**Next session starts with:**
> I'm continuing from a previous session. Here's the summary:
>
> [Full summary...]
>
> Ready to continue from where we left off. What would you like to work on?

---

### **Example 2: With Specific Next Goal**

```bash
# In Claude Code:
/save-session Let's implement syntax highlighting in the preview pane
```

**What happens:**
1. Claude creates summary of current work
2. Adds your goal to the "NEXT SESSION GOAL" section
3. Saves to `/tmp/claude-session-summary.md`

**Summary includes:**
```markdown
## NEXT SESSION GOAL

**User wants to work on:**
Let's implement syntax highlighting in the preview pane
```

**Next session starts with:**
> **User wants to work on:**
> Let's implement syntax highlighting in the preview pane
>
> ---
>
> Here's the summary from my previous session:
> [Full summary...]
>
> ---
>
> Let's get started on the goal above!

**Result:** Claude immediately starts working on your specified task! 🎯

---

## Real-World Workflow Examples

### **Scenario 1: End of Day Summary**

```bash
# Long development session, time to wrap up
/save-session Tomorrow: Fix the tree view cursor bug and add keyboard navigation

# Exit Claude
Ctrl+D

# Next morning:
session-reload

# Claude starts with:
# "User wants to work on: Tomorrow: Fix the tree view cursor bug and add keyboard navigation"
# [Shows full context]
# "Let's get started on the goal above!"
```

**Benefit:** You don't have to remember what you were working on! 🌅

---

### **Scenario 2: Context Switch Mid-Session**

```bash
# Working on feature A, but need to quickly fix a bug

# Save current state:
/save-session Continue implementing the preview variable highlighting feature

# Exit and work on the bug in a fresh session
Ctrl+D
claude

# After bug is fixed, reload original work:
session-reload

# Claude knows exactly what to continue: preview variable highlighting
```

**Benefit:** Clean context switching without losing your place! 🔄

---

### **Scenario 3: Specific Implementation Task**

```bash
# Just finished research, ready to implement
/save-session Implement the tmux monitoring feature using Desktop Commander's execute_command tool. Start by creating the /watch-tmux slash command.

# Reload:
session-reload

# Claude immediately:
# 1. Shows the summary
# 2. Starts working on tmux monitoring
# 3. Creates /watch-tmux command
# No need to re-explain what you want!
```

**Benefit:** Precise direction for the next session! 🎯

---

### **Scenario 4: Debugging Session**

```bash
# Found a bug, context is getting cluttered
/save-session Debug: TFE crashes when expanding empty directories in tree view. Check getCurrentFile() and buildTreeItems() functions.

session-reload

# Claude focuses on:
# 1. Reading getCurrentFile()
# 2. Reading buildTreeItems()
# 3. Looking for empty directory edge cases
# 4. Suggesting fixes
```

**Benefit:** Focused debugging without the clutter! 🐛

---

### **Scenario 5: Multi-Part Feature**

```bash
# Completed part 1 of a feature
/save-session Part 2: Now add the UI controls for the prompt variable editor. Create input fields below the preview pane.

session-reload

# Claude knows:
# - What was done (from summary)
# - What's next (your goal)
# - Context needed (from summary)
```

**Benefit:** Smooth multi-session feature development! 🚀

---

## Tips & Best Practices

### **Be Specific in Your Goals**

❌ **Too vague:**
```bash
/save-session Work on the preview feature
```

✅ **Good:**
```bash
/save-session Add syntax highlighting to markdown preview using chroma or similar library
```

✅ **Even better:**
```bash
/save-session Add syntax highlighting: 1) Research Go syntax highlighting libraries, 2) Integrate with preview rendering, 3) Add language detection for code blocks
```

---

### **Reference File/Function Names**

✅ **Helpful:**
```bash
/save-session Fix the panic in file_operations.go line 234 - add bounds checking before accessing m.files[m.cursor]
```

**Why:** Claude can immediately navigate to the right place!

---

### **Set Context for Next Day**

✅ **Great for morning sessions:**
```bash
/save-session Tomorrow: Continue the keyboard shortcut refactor. Need to update update_keyboard.go to use a dispatch table instead of the giant switch statement.
```

**Why:** You'll remember exactly where you left off!

---

### **Break Down Complex Tasks**

✅ **Multi-step goals:**
```bash
/save-session Next: Performance optimization sprint
1. Profile the file loading in large directories
2. Add caching to getFileIcon()
3. Optimize renderDetailView() rendering
Test each change with /rebuild-tfe
```

**Why:** Claude gets a clear roadmap!

---

## Quick Comparison

| Command | Next Session Starts With | Use When |
|---------|-------------------------|----------|
| `/save-session` | "What would you like to work on?" | General continuation |
| `/save-session [goal]` | Immediately works on your goal | You know what's next |
| Built-in `/compact` | Compressed context, same session | Quick reset, stay in session |

---

## Pro Tips

### **Combine with Tmux Monitoring**

```bash
# End of session:
/save-session Start /watch-tmux monitoring and test the new tree view expansion

# Next session:
session-reload

# Claude will:
# 1. Load summary
# 2. Start /watch-tmux
# 3. Begin testing tree view
```

---

### **Use for Different Time Zones**

```bash
# Before bed:
/save-session Morning: Review the changes to render_preview.go and add error handling for binary file previews

# Next morning:
session-reload
# Fresh start with clear direction!
```

---

### **Quick Feature Pivots**

```bash
# Midway through feature:
/save-session Pause: Implement the context menu first (user needs it now). Resume preview highlighting after.

session-reload
# Clear priority shift!
```

---

## Summary

**Without goal:**
- `/save-session` → general summary
- Next session: open-ended

**With goal:**
- `/save-session [your goal]` → directed summary
- Next session: focused on your task

**The magic:** Your future self (or fresh Claude session) knows exactly what to do! ✨
