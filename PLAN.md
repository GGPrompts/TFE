# TFE Development Plan

**Project:** TFE - Terminal File Explorer
**Status:** v0.3.0 - Feature-complete file viewer/browser
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

### 🚧 Known Limitations
- **F7/F8 are placeholders** - No create directory or delete file operations yet
- **Silent error handling** - Some operations fail without user feedback
- **No search** - Can't filter/find files within current directory
- **No multi-select** - Operations limited to single files
- **Large update.go** - 991 lines, needs refactoring

---

## Roadmap

### Phase 1: Complete File Operations 🎯 **HIGH PRIORITY**

**Goal:** Make TFE a true file *manager*, not just a viewer

**Why:** F7/F8 are the most visible incomplete features. Users can press them and nothing happens.

#### 1.1 Dialog System (NEW)
**Create:** `dialog.go`

```go
type dialogType int
const (
    dialogNone dialogType = iota
    dialogInput       // For F7 (directory name)
    dialogConfirm     // For F8 (yes/no delete)
    dialogError       // For error messages
    dialogSuccess     // For success messages
)

type dialogModel struct {
    dialogType dialogType
    title      string
    message    string
    input      string      // For text input
    callback   func()      // Action on confirm
}
```

**Features:**
- Text input dialog for directory names (F7)
- Yes/No confirmation dialog for destructive operations (F8)
- Error/Success toast notifications (auto-dismiss after 3s)
- Overlay rendering (similar to context menu)
- ESC to cancel, Enter to confirm
- Input validation

#### 1.2 F7: Create Directory
**Location:** `update.go` + `file_operations.go`

**Implementation:**
1. Detect F7 keypress → Show input dialog
2. Get directory name from user
3. Validate name (no /, \\, special chars)
4. Call `os.Mkdir()` with 0755 permissions
5. Handle errors gracefully (show error dialog)
6. Refresh file list
7. Move cursor to newly created directory

**Context Menu Integration:**
- Add "New Folder..." option to context menu when in directory

#### 1.3 F8: Delete File/Folder
**Location:** `update.go` + `file_operations.go`

**Implementation:**
1. Detect F8 keypress → Show confirmation dialog
2. Display: "Delete [filename]? This cannot be undone."
3. If confirmed:
   - For files: `os.Remove()`
   - For empty directories: `os.Remove()`
   - For non-empty directories: Show warning, require second confirmation, then `os.RemoveAll()`
4. Handle errors (permissions, file in use, etc.)
5. Show success message
6. Refresh file list
7. Move cursor to previous item (or next if was last)

**Context Menu Integration:**
- Add "Delete" option to context menu
- Same behavior as F8

**Safety Features:**
- Always confirm before delete
- Warn on non-empty directories
- Clear error messages for permission issues
- Don't delete if file doesn't exist anymore

#### 1.4 Error Feedback System
**Location:** `types.go`, `view.go`

Add status message system:
```go
type statusMessage struct {
    text      string
    isError   bool
    timestamp time.Time
}

// In model:
statusMsg *statusMessage  // Auto-dismiss after 3 seconds
```

Show in status bar:
- Success: Green background, "✓ Directory created: project/"
- Error: Red background, "✗ Permission denied: /root/folder"
- Auto-dismiss after 3 seconds or ESC

**Files to Update:**
- `types.go` - Add dialog state, status message to model
- `dialog.go` - NEW file for dialog system
- `update.go` - Add F7/F8 handlers, dialog event handling
- `view.go` - Render dialog overlay, status messages
- `context_menu.go` - Add "New Folder" and "Delete" menu items
- `file_operations.go` - Add mkdir and delete helper functions

**Testing Checklist:**
- [ ] F7 creates directory successfully
- [ ] F7 handles invalid names (/, *, etc.)
- [ ] F7 handles existing directory name
- [ ] F7 ESC cancels without creating
- [ ] F8 shows confirmation dialog
- [ ] F8 deletes file after confirmation
- [ ] F8 deletes empty directory
- [ ] F8 warns on non-empty directory
- [ ] F8 ESC cancels without deleting
- [ ] Context menu "Delete" works
- [ ] Context menu "New Folder" works
- [ ] Error messages display correctly
- [ ] Success messages display correctly

**Estimated Time:** 4-6 hours

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

#### 2.4 Refactor update.go (OPTIONAL)
**Issue:** 991 lines with 55+ case statements

**Goal:** Split into focused handler files

**New Structure:**
```
keyboard_handler.go  - All keyboard event cases
mouse_handler.go     - All mouse event cases
window_handler.go    - Window resize events
spinner_handler.go   - Spinner/background task events
```

**Benefits:**
- Easier to maintain
- Better testability
- Clear separation of concerns

**Note:** This is optional - only do if codebase feels unmanageable

**Estimated Time:** 3-4 hours

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

### Phase 4: File Operations (Extended)

**Goal:** Complete file management capabilities

#### Features:
- Copy file/directory
- Move/rename file/directory
- Batch operations (multi-select with Space)
- Progress indicators for long operations
- Undo/trash support (optional)

**Note:** Lower priority than Phase 1-3

**Estimated Time:** 2-3 weeks

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

### 🔥 Immediate (Do First)
1. **Implement F7/F8** - Most visible gap
   - Create dialog system
   - F7 create directory
   - F8 delete file/folder
   - Add error feedback

2. **Fix Silent Errors** - Critical UX issue
   - Add status message system
   - Show all error messages

3. **Add Search** - Highly valuable, quick to implement
   - `/` to search
   - Filter current directory

### ⭐ High Value (Do Soon)
4. **Context Visualizer** - Unique differentiator
   - This makes TFE special
   - No other tool has this

5. **Fix Grid View Bugs** - Prevent potential crashes

### 🔧 Nice to Have (Do Later)
6. **Refactor update.go** - Only if feeling unwieldy
7. **Extended File Ops** - Copy, move, batch operations
8. **Windows Features** - Terminology, permissions editor

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

- **Focus:** Ship F7/F8 next - most visible incomplete feature
- **Differentiator:** Context Visualizer is what makes TFE unique
- **Philosophy:** Hybrid approach (native preview + external editor)
- **Target:** Windows→WSL developers, Claude Code users
- **Keep:** Fast, simple, modular

---

**Last Updated:** 2025-10-16
**Next Session:** Implement F7/F8 + Dialog System (Phase 1)
