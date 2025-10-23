---
description: Summarize and export current work session to a file
---

# Save Session: Summarize and Export Current Work Session

You are creating a concise summary of the current work session and saving it to a file, enabling the user to continue in a completely fresh conversation or resume later with minimal context loss.

**Note:** This is different from the built-in `/compact` command which compresses the conversation in-place. This command EXPORTS the summary to a file for external use with reload scripts.

## Command Arguments

The user may provide an optional argument after `/save-session`:
- If provided: Use it as the **opening prompt for the next session**
- If empty: Just create summary without a specific next task

**Examples:**
- `/save-session` - Just create summary
- `/save-session Let's implement syntax highlighting in the preview` - Summary + specific next goal
- `/save-session Debug the tree view expansion issue` - Summary + next debugging task

## Your Task

Create a comprehensive but concise summary of our conversation that captures:

### 1. Session Overview
- What we accomplished
- What features were implemented  
- What problems were solved
- Current state of the project

### 2. Technical Details
**Files Modified:**
- List all files created or modified
- Include key changes made to each file
- Note any new functions, types, or significant refactors

**Architecture Decisions:**
- Any important architectural choices
- New patterns introduced
- Refactoring decisions

**Current State:**
- Where we left off
- What's working
- What's in progress
- Any known issues or TODOs

### 3. Context for Continuation
**Next Steps:**
- What should be worked on next
- Any blockers or dependencies
- Suggested priorities

**Important Context:**
- Key insights or discoveries
- Things to remember
- Gotchas or caveats

## Summary Format

Use this structure for the summary:

```markdown
# Session Summary - [Date/Time]

## What We Accomplished
- [Bullet point list of main achievements]

## Files Modified
### Created:
- `path/to/file.ext` - Description of what it does

### Modified:
- `path/to/file.ext` - What changed and why

## Key Technical Details
- [Important implementation details]
- [Architecture decisions]
- [New patterns or approaches]

## Current State
- ✅ Working: [What's functional]
- 🔄 In Progress: [What's partially done]
- ⚠️ Issues: [Known problems]

## Next Steps
1. [Priority 1]
2. [Priority 2]
3. [Priority 3]

## Important Context
- [Key insights to remember]
- [Gotchas or warnings]
- [Dependencies or requirements]

---

## NEXT SESSION GOAL

[If user provided argument to /save-session, include it here with clear formatting]

**User wants to work on:**
[Their specified goal/task]

[If no argument provided, use "No specific goal set - ready to continue general development"]

```

**IMPORTANT:** If the user provided a specific goal/task as an argument to `/save-session`, make sure to include it prominently at the end of the summary so the next session starts with clear direction.

## After Creating Summary

1. **Save the summary** using Desktop Commander's `write_file` tool:
   ```
   write_file({
     path: "docs/NEXT_SESSION.md",
     content: [the summary you created]
   })
   ```

   **Why this location?**
   - Consistent project-specific location
   - Easy to find and review (`cat docs/NEXT_SESSION.md`)
   - Can be git-tracked for team collaboration (optional)
   - Scripts always know where to look

2. **Confirm to user:**
   ```
   ✅ Summary saved to docs/NEXT_SESSION.md
   [If user provided goal: "Next session goal: [their goal]"]

   To continue with the work:

   Option 1 (Built-in /compact - Fastest):
     Use: /compact
     → Compresses current conversation in-place (1 second)
     → Stay in same session
     → Claude's native feature

   Option 2 (Auto-compact - Automated):
     In tmux? Run from another terminal: auto-compact
     → Automatic in-session compact (10-15 seconds)
     → Uses this summary automatically
     → Stays in same session

   Option 3 (Completely Fresh Session):
     Exit Claude and run: session-reload
     → Starts brand new Claude session
     → Pre-loaded with docs/NEXT_SESSION.md
     → Maximum context reset

   Option 4 (Manual Review):
     Run: show-session
     → View summary and copy to clipboard
     → Start new session manually
     → Paste summary

   TIP: /compact for quick resets (built-in)
        auto-compact for automated tmux workflows
        session-reload for complete fresh start
   ```

3. **Display the summary** to the user so they can review it

## Style Guidelines

- **Be concise but complete** - Include everything needed, nothing extra
- **Use clear structure** - Make it easy to scan and understand
- **Prioritize actionable information** - Focus on what helps continue work
- **Preserve technical details** - Don't lose important implementation specifics
- **Maintain context** - Include enough background for the next session

Now create the summary for our current session and save it to `/tmp/claude-session-summary.md`.
