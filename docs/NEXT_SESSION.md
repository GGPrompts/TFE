# Next Session: Fix Markdown Preview Scrolling Performance

## Problem Statement

Markdown files with ASCII art / box formatting (like CLAUDE.md, README.md) have extremely laggy scrolling in full-screen preview mode, while simpler markdown files scroll smoothly. The lag is NOT related to file length - some very long files are fine while others with box art are choppy.

## Root Cause Analysis (Credit: OpenAI Codex)

**The Issue:**
- Markdown previews never hit the cache
- In `file_operations.go:708-717`, the markdown branch returns immediately after loading the file
- This skips the `populatePreviewCache()` call
- Result: `render_preview.go:141-193` rebuilds a Glamour renderer and re-renders the **entire document on every frame**
- Glamour walks every rune even for simple monospace art, making this very expensive
- Even worse: when toggling to full-screen, the width changes but cache is never repopulated
- Width mismatch forces re-running Glamour on every redraw → laggy scrolling

## Code Locations

**file_operations.go:**
- Lines 708-717: Markdown loading branch that returns early, skipping cache population
- Need to ensure `populatePreviewCache()` is called after markdown load

**render_preview.go:**
- Lines 141-193: Glamour renderer recreation on every frame (performance hotspot)
- Should be using cached content instead

**Cache-related functions:**
- `populatePreviewCache()`: Needs to be called after markdown load
- Cache invalidation on view mode switch (width changes)

## Required Fixes

### Fix 1: Always Cache Markdown Previews
After loading markdown content in `file_operations.go`, ensure `populatePreviewCache()` is called so the Glamour-rendered output is cached once (it already stores results in `cachedRenderedContent`).

**Current problematic code (~line 708-717):**
```go
// Markdown branch returns early - skips populatePreviewCache()
if isMarkdown {
    // ... render with Glamour
    return  // ← This is the problem!
}
```

**Solution:** Don't return early, let it fall through to `populatePreviewCache()`

### Fix 2: Repopulate Cache on View Mode Changes
When switching between dual-pane and full-screen preview modes, the width changes. Recompute the cache with the new width.

**Locations to add cache repopulation:**
1. After `m.viewMode = viewFullPreview` (in update_keyboard.go)
2. When returning to split view from full-screen
3. Anywhere the preview width changes

**Pseudocode:**
```go
m.viewMode = viewFullPreview
m.calculateLayout()  // Updates widths
m.populatePreviewCache()  // Refresh cache with new width
```

### Fix 3: (Optional Follow-up) Optimize Glamour Renderer
Create the Glamour renderer once per width instead of inside the render loop. However, fixes 1 and 2 should eliminate the stutter on their own.

## Testing Checklist

After implementing fixes, test with:
- [ ] CLAUDE.md (has box art project structure)
- [ ] README.md (has various formatting)
- [ ] A simple markdown file (should still work smoothly)
- [ ] Toggle between dual-pane and full-screen preview
- [ ] Scroll up and down rapidly with arrow keys and mouse wheel
- [ ] Verify no performance regression in non-markdown files

## Expected Outcome

- Smooth scrolling in markdown previews (including ASCII art heavy files)
- Cache is populated once after loading markdown
- Cache is refreshed when view mode / width changes
- No repeated Glamour re-renders on every frame

## Implementation Approach

1. Read `file_operations.go` around lines 690-720
2. Identify the markdown early return
3. Refactor to ensure `populatePreviewCache()` is called
4. Find all view mode switches in `update_keyboard.go`
5. Add `m.populatePreviewCache()` calls after mode changes
6. Build and test with CLAUDE.md and README.md
7. Verify performance improvement

## Quick Start Prompt

```
Hi! There's a performance issue with markdown preview scrolling. Files with ASCII
art (like CLAUDE.md) are very laggy while scrolling.

The root cause (identified by Codex): markdown previews skip populatePreviewCache()
due to an early return in file_operations.go around line 708-717. This causes
Glamour to re-render the entire document on every frame.

Also, when switching view modes, the cache width changes but isn't repopulated.

Can you:
1. Fix the markdown loading to call populatePreviewCache()
2. Add cache repopulation when view modes change (dual-pane ↔ full-screen)
3. Test with CLAUDE.md to verify smooth scrolling

The relevant files are file_operations.go (lines 690-720) and update_keyboard.go
(view mode switches).
```

---

## Recently Completed (Previous Session)

### Browser Support + Paste Bug Fix + 's' Key Fix

**Major Changes:**
1. **Browser support** for images and HTML files (F3 + context menu)
2. **Paste bug fix** - using `msg.Runes` instead of `msg.String()` (credit: Codex)
3. **'s' key fix** - removed hotkey to allow typing 's' in commands
4. **Refactoring Phase 9** - split update.go into 3 files

**Files Modified:**
- editor.go: Browser detection and launch functions
- context_menu.go: "🌐 Open in Browser" option
- update_keyboard.go: F3 handler, paste fix, removed 's' hotkey
- update.go: Removed unnecessary paste helper functions
- HOTKEYS.md, CHANGELOG.md: Documentation updates

**Build Status:** ✅ All changes committed (commit 791c1bb)

---

## Current File Status

```
main.go: 21 lines ✅
styles.go: 35 lines ✅
helpers.go: 69 lines ✅
model.go: 78 lines ✅
update.go: 104 lines ✅ (cleaned up after paste fix)
command.go: 127 lines ✅
dialog.go: 141 lines ✅
favorites.go: 150 lines ✅
editor.go: 156 lines ✅ (browser support)
types.go: 173 lines ✅
view.go: 198 lines ✅
context_menu.go: 318 lines ✅ (browser context menu)
render_file_list.go: 447 lines ✅
render_preview.go: 498 lines ✅ (← performance issue here)
update_keyboard.go: 730 lines ✅ (paste fix applied)
update_mouse.go: 470 lines ✅
file_operations.go: 846 lines ⚠️ (← cache issue here)
```

**Architecture Status:** ✅ All modules under control, modular architecture maintained

---

## Documentation Status

Last checked: 2025-10-16

```
CLAUDE.md: 450 lines ✅ (under 500 limit)
README.md: 375 lines ✅ (under 400 limit)
PLAN.md: 339 lines ✅ (under 400 limit)
CHANGELOG.md: 275 lines ✅ (under 300 limit)
BACKLOG.md: 97 lines ✅ (under 300 limit)
HOTKEYS.md: 174 lines ✅ (under 200 limit)
docs/NEXT_SESSION.md: This file
```

All documentation within limits! ✅
