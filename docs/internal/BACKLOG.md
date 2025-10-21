# TFE Backlog - Ideas & Future Features

This document contains ideas, feature requests, and brainstorming notes that are **not yet ready** for PLAN.md. Think of this as the "parking lot" for good ideas that need more thought or aren't prioritized yet.

**Status:** When an idea is refined and ready for implementation, move it to PLAN.md.

---

## UI/UX Improvements

### Icon Fallback System (3-Tier Support)
**Status:** Post-launch v1.1+ feature
**Complexity:** Medium
**Priority:** Medium (for broader terminal compatibility)

Add automatic emoji detection and fallback to ASCII icons when emojis don't render properly.

**Three icon modes:**
1. **Emoji Mode** (default, current) - `📁 folder/  🐹 main.go  📄 README.md`
2. **ASCII Mode** (fallback) - `[D] folder/  [G] main.go  [T] README.md`
3. **None Mode** (minimal) - `> folder/  - main.go  - README.md`

**Implementation:**
- Add `--icons=emoji|ascii|none` command-line flag
- Auto-detect emoji support on startup (test render a known emoji)
- Graceful fallback if test shows boxes
- Store icon mode in model
- Create parallel ASCII icon mapping in `file_operations.go`

**Why:**
- TFE works perfectly in Windows Terminal and Termux with emojis
- Some users on older terminals or specific configurations might not see emojis
- Provides universal compatibility without sacrificing beautiful defaults

**Branch:** `feature/icon-fallback` (create when ready)

---

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

**Last Updated:** 2025-10-20
