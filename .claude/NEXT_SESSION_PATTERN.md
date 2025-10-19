# The NEXT_SESSION.md Pattern

**Simple Insight:** Always save session summaries to the same location: `docs/NEXT_SESSION.md`

## Why This Is Better

### Old Approach (Complex):
```
/save-session → /tmp/claude-session-summary.md
session-reload → looks for /tmp/claude-session-summary.md
auto-compact → uses /tmp/claude-session-summary.md
```

**Problems:**
- ❌ `/tmp` files can be lost
- ❌ Not project-specific
- ❌ Scripts need to pass file path around
- ❌ Can't git track it
- ❌ Hard to review manually

### New Approach (Simple):
```
/save-session → docs/NEXT_SESSION.md
session-reload → always reads docs/NEXT_SESSION.md
auto-compact → always uses docs/NEXT_SESSION.md
```

**Benefits:**
- ✅ **Consistent location** - Always know where to find it
- ✅ **Project-specific** - Each project has its own
- ✅ **Git-trackable** - Can commit for team collaboration
- ✅ **Easy to review** - `cat docs/NEXT_SESSION.md`
- ✅ **Editable** - Manually adjust before starting
- ✅ **Simpler scripts** - No path configuration needed

## The Brilliant Bonus: Git History

When you use `auto-compact --commit`:

```
Commit 1: "Phase 1 complete"
  → docs/NEXT_SESSION.md contains: "Start Phase 2"

Commit 2: "Phase 2 complete"
  → docs/NEXT_SESSION.md contains: "Start Phase 3"

Commit 3: "Phase 3 complete"
  → docs/NEXT_SESSION.md contains: "Start Phase 4"
```

Even though `docs/NEXT_SESSION.md` is **overwritten** each time:
- ✅ Git history preserves every session summary
- ✅ See what the plan was at any commit
- ✅ Track how the plan evolved
- ✅ Compare planned vs actual work

**View old summaries:**
```bash
# What was the plan at commit abc123?
git show abc123:docs/NEXT_SESSION.md

# See all changes to the plan
git log -p docs/NEXT_SESSION.md

# Compare what you planned vs did
git diff HEAD~1 HEAD -- docs/NEXT_SESSION.md
```

## Workflow Examples

### Example 1: Simple Compact and Continue

```bash
# In Claude:
/save-session Implement syntax highlighting

# Exit
Ctrl+D

# In shell:
session-reload
```

**Result:** New Claude session starts with the summary from `docs/NEXT_SESSION.md`

---

### Example 2: Auto-Compact in Tmux

```bash
# Terminal 1: Claude running in tmux
tmux new -s dev
claude

# Terminal 2: Trigger auto-compact
auto-compact -t dev -g "Add error handling"
```

**Result:**
- Claude session automatically compacts
- Summary saved to `docs/NEXT_SESSION.md`
- Session continues with fresh context
- Next goal: "Add error handling"

---

### Example 3: Phased Work with Git Commits

```bash
# Phase 1 work in Claude...

# From another terminal:
auto-compact --commit -g "Phase 2: Backend API"
```

**What happens:**
1. Git commit with Phase 1 changes + current `docs/NEXT_SESSION.md`
2. Claude creates new summary: "Start Phase 2"
3. Saves to `docs/NEXT_SESSION.md` (overwrites)
4. Session compacts
5. Claude starts Phase 2

**Git history shows:**
- Commit has Phase 1 code + summary saying "do Phase 2"
- Summary is preserved in git even though file is overwritten

---

### Example 4: Review Before Starting

```bash
# Check what you were working on
cat docs/NEXT_SESSION.md

# Edit it if needed
vim docs/NEXT_SESSION.md

# Start with edited summary
session-reload
```

---

### Example 5: Team Collaboration

```bash
# Commit your session summary
git add docs/NEXT_SESSION.md
git commit -m "Session summary: Implemented auth API"
git push

# Teammate pulls and sees what you were working on
git pull
cat docs/NEXT_SESSION.md

# They continue from where you left off
session-reload
```

## File Location Options

We chose `docs/NEXT_SESSION.md`, but you could use:

| Location | Pros | Cons |
|----------|------|------|
| `docs/NEXT_SESSION.md` | Organized, standard docs folder | Need to create docs/ |
| `NEXT_SESSION.md` | Root, most visible | Clutters project root |
| `.claude/NEXT_SESSION.md` | Organized with Claude files | Hidden folder, less visible |

**Recommendation:** `docs/NEXT_SESSION.md` - organized and visible

## Git Integration

### Option 1: Track It (Recommended for Solo)
```bash
# Add to git
git add docs/NEXT_SESSION.md
git commit -m "Session summary: Feature complete"

# Benefits:
# - Full history of session summaries
# - Can see evolution of work
# - Resume from any commit
```

### Option 2: Ignore It (For Teams)
```bash
# Add to .gitignore
echo "docs/NEXT_SESSION.md" >> .gitignore

# Benefits:
# - Won't conflict with teammates
# - Keep session notes private
# - Clean git history
```

### Hybrid: Auto-Commit with Git Commits
When using `auto-compact --commit`:
- Code changes are committed
- `docs/NEXT_SESSION.md` is included in commit
- Shows what you planned next
- Full context preserved

## Usage Patterns

### Pattern 1: Active Development
```bash
# Work for 30-60 minutes
# When context gets full:
auto-compact --commit -g "Continue: Add tests"

# Repeat every hour or so
```

### Pattern 2: End of Day
```bash
# Before leaving:
/save-session Tomorrow: Fix the tree view bug

# Next morning:
session-reload
# Claude knows exactly what to work on
```

### Pattern 3: Multi-Phase Features
```bash
# After each phase:
auto-compact --commit -g "Phase N: [description]"

# Or fully automated:
execute-phased-plan plan.txt
# Handles all compacts + commits automatically
```

## Scripts Updated

All scripts now use `docs/NEXT_SESSION.md`:

1. **`/save-session`** slash command
   - Writes to `docs/NEXT_SESSION.md`
   - Creates `docs/` folder if needed

2. **`auto-compact`** script
   - Detects project directory from tmux pane
   - Uses `$PANE_DIR/docs/NEXT_SESSION.md`
   - Creates `docs/` if needed

3. **`session-reload`** script
   - Reads `docs/NEXT_SESSION.md` from current directory
   - Starts new Claude with that summary

4. **`show-session`** script
   - Displays `docs/NEXT_SESSION.md`
   - Copies to clipboard

## Quick Reference

```bash
# Create summary
/save-session [optional goal]

# View summary
cat docs/NEXT_SESSION.md

# Show and copy to clipboard
show-session

# Fresh session with summary
session-reload

# Auto-compact in tmux
auto-compact -g "Next goal"

# Auto-compact with git commit
auto-compact --commit -g "Next goal"

# View old summary from git
git show HEAD~1:docs/NEXT_SESSION.md
git log -p docs/NEXT_SESSION.md
```

## Migration from Old System

If you have summaries in `/tmp/claude-session-summary.md`:

```bash
# Copy to new location
cp /tmp/claude-session-summary.md docs/NEXT_SESSION.md

# Or move
mv /tmp/claude-session-summary.md docs/NEXT_SESSION.md

# Continue with new system
session-reload
```

## Summary

**One file, consistent location, git-tracked history.**

This simple pattern:
- ✅ Simplifies all scripts
- ✅ Makes summaries easy to find
- ✅ Enables git history of plans
- ✅ Supports team collaboration
- ✅ Works with all workflows

**Location:** `docs/NEXT_SESSION.md`
**Commands:** `/save-session`, `session-reload`, `auto-compact`
**Git:** Optionally tracked for full history

Simple, powerful, elegant. 🎯
