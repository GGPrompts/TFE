# TFE Backlog - Ideas & Future Features

This document contains ideas, feature requests, and brainstorming notes that are **not yet ready** for PLAN.md. Think of this as the "parking lot" for good ideas that need more thought or aren't prioritized yet.

**Status:** When an idea is refined and ready for implementation, move it to PLAN.md.

---

## UI/UX Improvements

### Menu Bar with Dropdowns
**Status:** Needs careful planning
**Complexity:** Medium-High (affects all mouse positioning)

Add a menu bar (File | View | Tools | Help) for better discoverability, especially for:
- Windows users new to Linux
- Mobile/Termux users (touch-friendly)
- Users who don't want to memorize hotkeys

**Blockers:**
- Mouse click positioning throughout codebase assumes 4-line header
- Need to update ~8+ locations that calculate `headerOffset`
- Risk of breaking existing mouse/touch functionality

**Requirements before implementation:**
1. Create `docs/MENU_BAR_SESSION.md` with detailed plan
2. Audit all mouse position calculations
3. Create centralized `getHeaderHeight()` function
4. Comprehensive testing checklist for mouse/touch

---

### Splash Screen on Launch
**Status:** Ready to implement (safe)
**Complexity:** Low (no mouse positioning changes)

Show ASCII art + version on first launch, then hide to save space.

**Benefits:**
- Professional branding
- Save 1-2 lines of space after splash
- No breaking changes to mouse positioning

**Could implement:** Anytime, low risk

---

## File Operations

### Copy/Move Files
**Status:** Brainstorming
**Complexity:** Medium

Add F6 (copy) and allow moving files between panes in dual-pane mode.

**Considerations:**
- Progress bars for large files?
- Error handling for permissions
- Overwrite confirmations

---

## Search & Navigation

### Quick Search / Jump to File
**Status:** Idea stage
**Complexity:** Medium

Type-to-filter files in current directory (like VS Code's Ctrl+P).

**UI Options:**
1. Inline filter (show as you type in command bar)
2. Overlay search box
3. Highlight matches as you type

---

### Recursive Search
**Status:** Low priority
**Complexity:** High

Search file contents across directory tree.

**Considerations:**
- Performance on large directories
- UI for search results
- Integration with ripgrep/fzf?

---

## Archive this file when it reaches ~300 lines

Move implemented items to PLAN.md or delete. Archive old ideas to `docs/archive/BACKLOG_2025.md`.

---

**Last Updated:** 2025-10-16
