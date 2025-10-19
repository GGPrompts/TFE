# TFE Development Plan

**Project:** TFE - Terminal File Explorer
**Status:** v0.4.0 - True file manager with F7/F8 operations
**Updated:** 2025-10-16

---

## Current Status

### ✅ Completed (See CHANGELOG.md for details)
- **Core File Browser** - Navigation, sorting, metadata display
- **View Modes** - List, Grid, Detail, Tree views
- **Dual-Pane System** - Split-screen with live preview
- **File Preview** - Text, markdown (Glamour), binary detection
- **External Editor** - Micro/nano/vim integration
- **F-Key Hotkeys** - Midnight Commander style (F1-F10)
- **Context Menu** - Right-click + F2 keyboard access
- **Favorites System** - Bookmarks with F6 filter
- **Command Prompt** - MC-style always-active shell
- **Clipboard Integration** - Multi-platform path copying
- **Mouse Support** - Click, double-click, scroll
- **TUI Tool Launcher** - lazygit, lazydocker, lnav, htop
- **F7: Create Directory** - ✨ NEW! Dialog system with validation
- **F8: Delete Files** - ✨ NEW! Safe deletion with confirmations
- **Dialog System** - Input, confirmation, and status messages

### 🚧 Known Limitations
- **No directory search** - Can't filter files by name pattern (have fuzzy search via Ctrl+P)
- **No multi-select** - Operations limited to single files
- **No copy/move** - Can't move files between directories yet

---

## Roadmap

### ✅ Phase 1: Complete File Operations - **COMPLETED v0.4.0**

**Status:** ✅ Fully implemented (2025-10-16)

**What was delivered:**
- F7: Create Directory with validation
- F8: Delete File/Folder with confirmations
- Dialog system (input, confirm, status messages)
- Context menu integration
- Auto-dismissing status messages (3s)

See **CHANGELOG.md** for full details.

---

### Phase 2: Code Quality & Stability 🔧

**Goal:** Improve maintainability and fix bugs

#### 2.1 Fix Silent Errors
**Issue:** 4 locations where errors are ignored without user feedback

Fix locations:
- `update.go:491` - Editor not available
- `context_menu.go:136` - Quick CD write failure
- `context_menu.go:167` - Clipboard copy failure
- `update.go:91` - Clipboard copy failure in preview

**Solution:** Use new status message system from Phase 1.4

#### 2.2 Add Search Feature
**Goal:** Filter files in current directory by name

**Keybinding:** `/` to enter search mode

**Implementation:**
```go
// In model:
searchMode   bool
searchQuery  string
filteredFiles []fileItem  // Subset of m.files matching search
```

**Features:**
- Type `/` to enter search mode
- Incremental filtering as you type
- ESC to clear search
- Display: "Search: [query] (5 matches)"
- Fuzzy matching optional

**Files to Update:**
- `types.go` - Add search fields to model
- `update.go` - Add search mode key handling
- `render_file_list.go` - Render filtered files
- `file_operations.go` - Add search filtering function

**Estimated Time:** 2-3 hours

#### 2.3 Fix Grid View Bugs
**Issue:** Potential off-by-one errors in click detection (update.go:762-768)

**Problems:**
- Uses `clickedRow` but should check `actualRow` in validation
- No check for `clickedCol >= m.gridColumns`

**Solution:** Add proper bounds checking

**Estimated Time:** 30 minutes

#### 2.4 Refactor update.go - ✅ **COMPLETED**
**Status:** ✅ Completed during Phase 1

**Result:** Successfully split into 3 focused files:
- `update.go` (138 lines) - Main dispatcher
- `update_keyboard.go` (821 lines) - Keyboard event handling
- `update_mouse.go` (663 lines) - Mouse event handling

**Benefits Achieved:**
- Much easier to maintain and navigate
- Clear separation of concerns
- Each file has single responsibility

---

### Phase 3: Context Visualizer ⭐ **UNIQUE FEATURE**

**Goal:** Show what Claude Code sees from current directory

**Why:** This is TFE's killer feature - no other tool does this

**Status:** Not started

#### Features:
1. **Context Analyzer** (Press `Ctrl+K` or new F-key)
   - Show all files with token estimates (~4 chars = 1 token)
   - Visual indicators: ✅ Included, ❌ Excluded, ⚠️ Too large
   - Display total tokens: "45K / 200K (22%)"
   - Parse .gitignore and .claudeignore patterns

2. **Config Hierarchy Viewer** (Press `Ctrl+Shift+K`)
   - Walk up directory tree finding CLAUDE.md files
   - Show settings precedence (enterprise → local → shared → global)
   - Display as tree with token counts
   - Show which settings files are active

3. **Token Optimizer**
   - Suggest files/folders to exclude
   - Calculate token savings
   - Generate .claudeignore entries
   - "Add to .claudeignore" action

**Implementation:**
- Create `context_analyzer.go` module
- Add context view mode to `viewMode` enum
- Token estimation functions
- .gitignore parser (use filepath.Match)
- .claudeignore parser
- Hierarchy walker (walk up to root)

**See:** Original PLAN.md lines 339-531 for detailed design

**Estimated Time:** 2-3 weeks

---

### Phase 4: Essential File Operations 🔥 **REQUIRED FOR v1.0**

**Goal:** Complete essential file management before public launch

**Status:** Not started - **CRITICAL for v1.0 release**

#### Required for Launch (4-6 hours):
1. **Copy Files** (2-3 hours) - ⚠️ BLOCKING v1.0
   - Context menu: "📋 Copy to..."
   - Input dialog for destination
   - Progress indicator for large files
   - Error handling (permissions, disk space)

2. **Rename Files** (1-2 hours) - ⚠️ BLOCKING v1.0
   - Context menu: "✏️ Rename..."
   - Input dialog pre-filled with current name
   - Validation and error handling

3. **New File Creation** (1 hour) - ⚠️ BLOCKING v1.0
   - Context menu: "📄 New File..."
   - Auto-open in editor after creation
   - Error handling

**Implementation Plan:**
- Create `file_copy.go` module for copy operations
- Extend `context_menu.go` with new actions
- Extend `dialog.go` for filename/path inputs
- Add error feedback for all operations

#### Nice to Have (v1.1+):
- Move files between panes in dual-pane mode
- Batch operations (multi-select with Space)
- Progress bars for long operations
- Undo/trash support

**See:** `docs/LAUNCH_CHECKLIST.md` for complete v1.0 requirements

**Estimated Time:** 4-6 hours (critical features only)

---

### Phase 5: Windows-Friendly Features

**Goal:** Bridge Windows and Linux concepts

#### Features:
- Dual terminology (e.g., "Shortcut (symlink)")
- Visual permissions editor (checkbox UI)
- Plain English command helper
- Symlink indicators

**Note:** Nice-to-have, not critical

**Estimated Time:** 1-2 weeks

---

## Prioritized Next Steps

### 🔥 **CRITICAL - Required for v1.0 Launch (4-6 hours)**
1. **✅ Copy Files** - Context menu + dialog (2-3 hours) - **DO FIRST**
2. **✅ Rename Files** - Context menu + dialog (1-2 hours) - **DO SECOND**
3. **✅ New File Creation** - Context menu + dialog (1 hour) - **DO THIRD**

**After these 3 features:** Ready for v1.0 launch! 🚀

### 📸 **Launch Preparation (8-12 hours total)**
4. **Screenshots/GIFs** - Show off the UI (2 hours)
5. **Documentation Polish** - Installation, features, comparison (1.5 hours)
6. **GitHub Release** - Binaries for Linux/macOS (2-3 hours)
7. **Testing** - Platform testing, edge cases (2 hours)
8. **Marketing Posts** - Reddit, HN, lobste.rs (1 hour)

**See:** `docs/LAUNCH_CHECKLIST.md` for complete checklist

### ⭐ **Post-Launch (v1.1+)**
9. **Context Visualizer** - Unique differentiator (Phase 3)
   - Show Claude Code context and token counts
   - .claudeignore optimization suggestions
   - This makes TFE special!

10. **Multi-Select Operations** - Batch copy/delete/move
11. **Archive Operations** - Extract/create .zip, .tar.gz
12. **Permissions Editor** - GUI for chmod

### ✅ **Already Complete**
- ~~F7/F8 operations~~ ✅
- ~~Silent errors fixed~~ ✅
- ~~Directory search (`/`)~~ ✅
- ~~Refactor update.go~~ ✅
- ~~Grid view bugs~~ ✅
- ~~Prompts library~~ ✅
- ~~Fillable fields for prompts~~ ✅ (Just finished!)

---

## Design Principles

### Keep It Simple
- Don't over-engineer
- Ship working features quickly
- Get feedback early

### Maintainability First
- Modular architecture (13 files, each with clear purpose)
- See CLAUDE.md for architecture guidelines
- Add new features to appropriate modules

### User Experience
- Always provide feedback (success/error messages)
- Confirm destructive operations
- Clear, helpful error messages
- Keyboard shortcuts for everything

### Performance
- Responsive UI (never block on I/O)
- Cache where appropriate
- Handle large directories gracefully

---

## Success Metrics

### Phase 1 Complete When:
- ✅ F7 creates directories
- ✅ F8 deletes files/folders safely
- ✅ All errors show user feedback
- ✅ Search works smoothly

### Phase 3 Complete When:
- ✅ Context view shows all files with token counts
- ✅ Hierarchy view shows CLAUDE.md chain
- ✅ Can identify context optimization opportunities
- ✅ Faster than manually running `/context`

### Daily Driver When:
- ✅ Use it instead of `ls` + `cd`
- ✅ Use it instead of `micro` (for opening files)
- ✅ Use it instead of `mkdir` / `rm`
- ✅ Use it for Claude Code context debugging

---

## Technical Debt

### Current Issues:
1. **Model too large** - 96 fields, mixed responsibilities
   - Consider: PreviewModel, ContextMenuModel, CommandPromptModel sub-structs

2. **Magic numbers** - Scattered throughout code
   - Consider: config.go with all constants

3. **Render side effects** - renderTreeView modifies model
   - Move tree building to Update()

4. **No tests** - Zero test coverage
   - Add tests when codebase stabilizes

### Don't Fix Unless:
- Code becomes hard to work with
- Adding new features is slow
- Bugs are frequent

Right now: **Ship features first, refactor later**

---

## Resources

### Documentation
- [Bubbletea Tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
- [Lipgloss Examples](https://github.com/charmbracelet/lipgloss/tree/master/examples)
- [Claude Code Docs](https://docs.claude.com/en/docs/claude-code)

### Similar Projects (For Inspiration)
- [Midnight Commander](https://github.com/MidnightCommander/mc) - Classic dual-pane
- [Yazi](https://github.com/sxyazi/yazi) - Modern Rust TUI
- [Ranger](https://github.com/ranger/ranger) - Python file manager

### Project Files
- **README.md** - User documentation
- **CLAUDE.md** - Architecture guide (READ THIS when adding features)
- **HOTKEYS.md** - Keyboard reference (shown with F1)
- **CHANGELOG.md** - Version history
- **NEXT_SESSION.md** - Short-term session notes

---

## Notes

- **Current Focus:** Phase 4 (Copy/Rename/New File) - **CRITICAL FOR v1.0**
- **Primary Differentiator:** Prompts library with fillable fields ✨ (unique feature!)
- **Secondary Differentiator:** Context Visualizer (future - Phase 3)
- **Philosophy:** Hybrid approach (native preview + external editor)
- **Target:** AI-assisted developers, Claude Code users, terminal power users
- **Keep:** Fast, simple, modular
- **Launch Goal:** v1.0 ready in 1 week (4-6 hours coding + 8-12 hours polish)

---

**Last Updated:** 2025-10-18
**Next Session:** Implement Copy/Rename/New File (Phase 4 - v1.0 blockers)
**Launch Checklist:** See `docs/LAUNCH_CHECKLIST.md` for complete v1.0 requirements
