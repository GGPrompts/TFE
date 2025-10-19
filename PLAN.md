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
- **Command Prompt** - Vim-style focus with `:` key
- **Clipboard Integration** - Multi-platform path copying
- **Mouse Support** - Click, double-click, scroll
- **TUI Tool Launcher** - lazygit, lazydocker, lnav, htop
- **F7: Create Directory** - Dialog system with validation
- **F8: Delete Files** - Safe deletion with confirmations (moves to trash)
- **Dialog System** - Input, confirmation, and status messages
- **Trash/Recycle Bin** - ✨ NEW! F12 to view trash, restore/permanently delete (F8 moves to trash)
- **Prompts Library** - ✨ NEW! F11 to filter prompts, fillable `{{VARIABLES}}`, auto-shows ~/.prompts and ~/.claude
- **New File Creation** - ✨ NEW! Context menu creates file and opens in editor
- **Suspend/Resume** - ✨ NEW! Ctrl+Z to drop to shell, `fg` to resume
- **Error Feedback** - All operations show success/error status messages
- **Copy Files** - ✨ NEW! Context menu → "📋 Copy to..." with recursive directory support
- **Rename Files** - ✨ NEW! Context menu → "✏️ Rename..." with validation
- **Image Support** - ✨ NEW! View (viu/timg/chafa) and Edit (textual-paint) images in terminal
- **Preview Search** - ✨ NEW! Ctrl-F to search within file previews
- **Mouse Toggle** - ✨ NEW! Press 'm' in preview to toggle border/mouse for clean text selection

### 🚧 Known Limitations
- **No multi-select** - Operations limited to single files (planned for v1.1+)

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

### Phase 2: Code Quality & Stability ✅ **COMPLETED**

**Goal:** Improve maintainability and fix bugs

#### 2.1 Fix Silent Errors - ✅ **COMPLETED**
**Status:** ✅ All 4 locations fixed with status messages

Fixed locations:
- ✅ Editor not available - Shows "No editor available" message
- ✅ Quick CD write failure - Shows error with setStatusMessage
- ✅ Clipboard copy failure (context menu) - Shows error with setStatusMessage
- ✅ Clipboard copy failure (preview) - Shows error with setStatusMessage

**Solution:** Uses status message system throughout

#### 2.2 Add Search Feature - ✅ **COMPLETED**
**Status:** ✅ Fully implemented with `/` key

**Features Delivered:**
- Type `/` to enter search mode
- Incremental filtering as you type
- ESC to clear search
- Search state tracked in model

#### 2.3 Fix Grid View Bugs - ✅ **OBSOLETE**
**Status:** ✅ Grid view removed entirely (3 display modes now: List, Detail, Tree)

**Result:** Simplified to 3 focused view modes, removed ~250 lines of code

#### 2.4 Refactor update.go - ✅ **COMPLETED**
**Status:** ✅ Completed during Phase 1

**Result:** Successfully split into 3 focused files:
- `update.go` (111 lines) - Main dispatcher
- `update_keyboard.go` (714 lines) - Keyboard event handling
- `update_mouse.go` (383 lines) - Mouse event handling

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

### Phase 4: Essential File Operations ✅ **COMPLETED 2025-10-19**

**Goal:** Complete essential file management before public launch

**Status:** ✅ **ALL 3 FEATURES COMPLETE - v1.0 UNBLOCKED!** 🎉

#### ✅ Completed Features:
1. **Copy Files** ✅ **COMPLETED**
   - ✅ Context menu: "📋 Copy to..."
   - ✅ Input dialog for destination (absolute or relative paths)
   - ✅ Recursive directory copying with permission preservation
   - ✅ Error handling (permissions, disk space)
   - ✅ Status messages for feedback
   - Files: `context_menu.go`, `update_keyboard.go`, `file_operations.go`

2. **Rename Files** ✅ **COMPLETED**
   - ✅ Context menu: "✏️ Rename..."
   - ✅ Input dialog pre-filled with current name
   - ✅ Validation (no empty names, no "/" characters)
   - ✅ Cursor moves to renamed file
   - ✅ Error handling with status messages
   - Files: `context_menu.go`, `update_keyboard.go`

3. **New File Creation** ✅ **COMPLETED**
   - ✅ Context menu: "📄 New File..."
   - ✅ Auto-opens in editor after creation
   - ✅ Full error handling with status messages

**BONUS Features Delivered:**
- **Image Support:** View (viu/timg/chafa) and edit (textual-paint) images
- **Preview Search:** Ctrl-F to search within files
- **Mouse Toggle:** Press 'm' in preview for clean text selection
- **Browser Fix:** Fixed WSL cmd.exe bug for opening images/HTML

#### Nice to Have (v1.1+):
- Move files between panes in dual-pane mode
- Batch operations (multi-select with Space)
- Progress bars for long operations

**See:** `docs/LAUNCH_CHECKLIST.md` for complete v1.0 requirements

**Time Spent:** ~4 hours (all critical features + bonuses!)

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

### ✅ **CRITICAL FEATURES - ALL COMPLETE!** 🎉
1. ~~**Copy Files**~~ - ✅ **COMPLETED 2025-10-19**
2. ~~**Rename Files**~~ - ✅ **COMPLETED 2025-10-19**
3. ~~**New File Creation**~~ - ✅ **COMPLETED**

**Status:** Ready for v1.0 launch! 🚀 All coding is DONE!

### 📸 **Launch Preparation (8-12 hours total)**
4. **Screenshots/GIFs** - Show off the UI (2 hours)
5. **Documentation Polish** - Installation, features, comparison (1.5 hours)
6. **GitHub Release** - Binaries for Linux/macOS (2-3 hours)
7. **Testing** - Platform testing, edge cases (2 hours)
8. **Marketing Posts** - Reddit, HN, lobste.rs (1 hour)

**See:** `docs/LAUNCH_CHECKLIST.md` for complete checklist

### ⭐ **Post-Launch (v1.1+)**
9. **Command Pre-filling** - 🔥 **REVOLUTIONARY FEATURE**
   - Pre-fill command line instead of executing operations directly
   - Educational: Users learn Linux commands as they work
   - Safe: Review before execution (Esc to cancel)
   - Powerful: Modify commands, add flags, chain operations
   - Examples:
     - Rename → Pre-fills `mv 'old.txt' '█'`
     - Copy → Pre-fills `cp 'file.txt' '█'`
     - Bulk operations → Pre-fills shell loops/patterns
   - **Marketing angle:** "The File Manager That Teaches You Linux"
   - Platform-aware templates (GNU vs BSD, WSL, Termux)
   - See session notes for full design

10. **Context Visualizer** - Unique differentiator (Phase 3)
    - Show Claude Code context and token counts
    - .claudeignore optimization suggestions
    - This makes TFE special!

11. **Multi-Select Operations** - Batch copy/delete/move
12. **Archive Operations** - Extract/create .zip, .tar.gz
13. **Permissions Editor** - GUI for chmod

### ✅ **Already Complete**
- ~~F7/F8 operations~~ ✅
- ~~Silent errors fixed~~ ✅
- ~~Directory search (`/`)~~ ✅
- ~~Refactor update.go~~ ✅
- ~~Grid view removed~~ ✅ (simplified to 3 view modes)
- ~~Prompts library (F11)~~ ✅
- ~~Fillable fields for prompts~~ ✅
- ~~New file creation~~ ✅
- ~~Trash/Recycle bin (F12)~~ ✅
- ~~Suspend/Resume (Ctrl+Z)~~ ✅
- ~~Command prompt helper text~~ ✅
- ~~Copy files~~ ✅ **COMPLETED 2025-10-19**
- ~~Rename files~~ ✅ **COMPLETED 2025-10-19**
- ~~Image support (view/edit)~~ ✅ **COMPLETED 2025-10-19**
- ~~Preview search (Ctrl-F)~~ ✅ **COMPLETED 2025-10-19**
- ~~Mouse toggle~~ ✅ **COMPLETED 2025-10-19**

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

- **Current Focus:** ✅ Phase 4 COMPLETE - **v1.0 READY FOR LAUNCH!** 🚀
- **Primary Differentiator:** Prompts library with fillable fields ✨ (unique feature!)
- **Secondary Differentiator:** Image support in terminal (view/edit)
- **Future Differentiator:** Command pre-filling (v1.1) - Educational file manager
- **Tertiary Differentiator:** Context Visualizer (v1.2+) - Claude Code integration
- **Philosophy:** Hybrid approach (native preview + external editor)
- **Target:** AI-assisted developers, Claude Code users, Windows→Linux learners
- **Keep:** Fast, simple, modular, educational
- **Launch Status:** All critical coding DONE! Ready for documentation & release (8-12 hours)

---

**Last Updated:** 2025-10-19
**Status:** ✅ ALL v1.0 FEATURES COMPLETE
**Next Session:** Launch preparation (screenshots, docs, binaries, marketing)
**Launch Checklist:** See `docs/LAUNCH_CHECKLIST.md` for complete v1.0 requirements
