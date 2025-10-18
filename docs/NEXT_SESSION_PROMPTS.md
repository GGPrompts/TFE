# Prompt for Next Session: Implement Prompts Feature Phase 1

## Context

Today we:
1. Fixed terminal cleanup issue (Ctrl+C now properly clears screen)
2. Fixed mouse click accuracy in tree view (4-line offset bug)
3. Added missing fuzzy search button to dual-pane toolbar
4. Restricted dual-pane mode to List/Tree views (Grid/Detail need full width)
5. **Created comprehensive plan for Prompts Library feature**

All changes committed and pushed to `main` branch.

## Current Branch Status

- **Main branch:** Clean, all fixes committed (commit: `0a2b6fb`)
- **Prompts branch:** Created with detailed implementation plan (commits: `2d4e63e`, `09fc716`, `89768e5`)

## The Prompts Feature Plan

**Location:** `docs/PROMPTS_FEATURE.md` on `prompts` branch

**Vision:** Transform TFE into a command center by adding a prompt library system:
- Browse organized prompts as files in `~/.prompts/`
- Template variables auto-fill ({{FILE}}, {{PROJECT}}, etc.)
- Copy rendered prompts to clipboard
- Paste into any AI CLI (Claude Code, aider, cursor, etc.)
- Store CLI command references alongside prompts

**Design Decision:** Simple copy/paste workflow only (no tmux integration, no auto-launch). User already has quick cd (Ctrl+Enter) working great for launching AI CLIs manually.

**Scope:** 4 phases, ~6-8 hours total

## What to Implement Next

### Phase 1: Prompts Filter & UI (~1-2 hours)

**Goal:** Add ability to toggle "Prompt Mode" which filters to show only `.yaml`, `.md`, `.txt` files.

**Tasks (from PROMPTS_FEATURE.md):**

1. Add `showPromptsOnly bool` field to `types.go` model struct
2. Add `isPromptFile()` helper function to `helpers.go`
   - Check for `.yaml`, `.md`, `.txt` extensions
3. Add `getFilteredPromptsFiles()` method (similar to `getFilteredFiles()`)
4. Add toolbar button `[📝]` in `view.go` and `render_preview.go`
   - Position: After `[🔍]` fuzzy search button
   - Toggle `showPromptsOnly` on click
   - Highlight when active: `✨📝`
5. Add keyboard shortcut `F11` to toggle prompt mode in `update_keyboard.go`
6. Add mouse click handler for toolbar button in `update_mouse.go`
   - Click region: X=25-29 (after search button at X=20-24)
7. Update status bar to show "• prompts only" indicator when active
8. Test: Toggle prompt mode, verify only `.yaml/.md/.txt` files shown

**Files to modify:**
- `types.go` - Add field
- `helpers.go` - Add helper function
- `view.go` - Add toolbar button (single-pane)
- `render_preview.go` - Add toolbar button (dual-pane)
- `update_keyboard.go` - Add F11 handler
- `update_mouse.go` - Add click handler for toolbar button

**Pattern to follow:**
Look at how `showFavoritesOnly` works (F6 toggle, star button, filter):
- `types.go:143` - `showFavoritesOnly bool`
- `favorites.go` - Has `isFavorite()` helper
- `file_operations.go:529` - Has `getFilteredFiles()` that respects favorites
- Toolbar buttons in `view.go:70-75` and `render_preview.go:383-391`

**Key decisions:**
- Use F11 for toggle (not currently assigned)
- Button emoji: `📝` (or `💬` or `📋`)
- Active state: `✨📝` (sparkles + emoji)
- Status indicator: "• prompts only" (like "• showing hidden" and "• ⭐ favorites only")

## Complete Prompt for Claude Code

---

I'm working on TFE (Terminal File Explorer) and we're implementing a new Prompts Library feature. We're on the `prompts` branch which contains the complete implementation plan.

**Current state:**
- Branch `prompts` has detailed plan in `docs/PROMPTS_FEATURE.md`
- Ready to start Phase 1: Prompts Filter & UI
- All previous work committed to `main` branch

**What to do:**

1. **Read the plan:**
   ```
   Read docs/PROMPTS_FEATURE.md on the prompts branch to understand the full vision
   ```

2. **Implement Phase 1 tasks (see checklist at line 84-102 in the plan):**
   - Add `showPromptsOnly` field to types.go
   - Add `isPromptFile()` helper to helpers.go (check for .yaml/.md/.txt)
   - Add toolbar button [📝] after fuzzy search button [🔍]
   - Add F11 keyboard shortcut to toggle
   - Add mouse click handler at X=25-29
   - Show "• prompts only" in status bar when active
   - Filter file list to only show prompt files when mode active

3. **Follow TFE's modular architecture:**
   - Check CLAUDE.md for architecture guidelines
   - Keep files focused (single responsibility)
   - Reuse existing patterns (look at how favorites filter works)
   - Test thoroughly before committing

4. **Pattern reference:**
   Study how `showFavoritesOnly` works:
   - Toggle: F6 key
   - Button: [⭐] or [✨]
   - Filter: `getFilteredFiles()` respects the flag
   - Status: Shows "• ⭐ favorites only"

   Do the same for prompts mode.

5. **Mouse click coordinates:**
   Current toolbar buttons:
   - [🏠] Home: X=0-4
   - [⭐] Favorites: X=5-9
   - [>_] Command: X=10-14
   - [📦] CellBlocks: X=15-19
   - [🔍] Search: X=20-24
   - [📝] Prompts (NEW): X=25-29

6. **Testing checklist:**
   - Click toolbar button → activates prompt mode
   - Press F11 → toggles prompt mode
   - Status bar shows "• prompts only"
   - Only .yaml, .md, .txt files visible
   - Can still use tree view, fuzzy search, favorites
   - Press F11 again → exits prompt mode, shows all files
   - Works in both single-pane and dual-pane modes

**Important notes:**
- This is Phase 1 of 4 - just the filter/UI
- No template parsing yet (that's Phase 2)
- No clipboard copying yet (that's Phase 3)
- Just get the basic toggle + filter working

**Expected outcome:**
After Phase 1, users should be able to:
1. Press F11 or click [📝] to enter "Prompt Mode"
2. See only prompt files (.yaml, .md, .txt) in the list
3. Status bar indicates "prompts only" mode is active
4. Press F11 or click button again to exit back to normal file browsing

When Phase 1 is complete, we'll move to Phase 2 (template variable parsing and rendering).

---

## Reference Files

Key files to understand:
- `docs/PROMPTS_FEATURE.md` - Complete implementation plan
- `CLAUDE.md` - TFE architecture guide
- `types.go` - All type definitions
- `helpers.go` - Helper functions
- `favorites.go` - Similar feature to copy pattern from

## Example Prompt Files

When implemented, users will create files like:

```yaml
# ~/.prompts/code-review/general.yaml
name: Code Review Request
description: Request code review with focus areas
template: |
  Please review {{FILE}} for:
  - Code quality
  - Performance
  - Security
```

```markdown
# ~/.prompts/_cli-commands/claude-flags.md

## Claude Code with context
```bash
claude --model sonnet-4 --context ./docs --context ./src
```

## Quick debugging session
```bash
claude --model sonnet-3.5 --no-context
```
```

But for Phase 1, we're just implementing the filter mechanism.

---

**Last session ended:** 2025-10-17
**Next step:** Implement Phase 1 on `prompts` branch
**Estimated time:** 1-2 hours
