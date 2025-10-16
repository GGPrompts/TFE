# TFE Architecture & Development Guide

This document describes the architecture of the TFE (Terminal File Explorer) project and provides guidelines for maintaining and extending the codebase.

## Architecture Overview

TFE follows a **modular architecture** where each file has a single, clear responsibility. This organization was established through a comprehensive refactoring that reduced `main.go` from 1668 lines to just 21 lines, distributing functionality across 11 focused modules.

### Core Principle

**When adding new features, always maintain this modular architecture by creating new files or extending existing modules rather than adding everything to `main.go`.**

## File Structure

```
tfe/
├── main.go (21 lines)           - Entry point ONLY
├── types.go (173 lines)         - Type definitions & enums
├── styles.go (35 lines)         - Lipgloss style definitions
├── model.go (78 lines)          - Model initialization & layout calculations
├── update.go (111 lines)        - Main update dispatcher & initialization
├── update_keyboard.go (714)     - Keyboard event handling
├── update_mouse.go (383)        - Mouse event handling
├── view.go (189 lines)          - View dispatcher & single-pane rendering
├── render_preview.go (468)      - Preview rendering (full & dual-pane)
├── render_file_list.go (447)    - File list views (List/Grid/Detail/Tree)
├── file_operations.go (657)     - File operations & formatting
├── editor.go (90 lines)         - External editor integration
├── command.go (127 lines)       - Command execution system
├── dialog.go (141 lines)        - Dialog system (input/confirm)
├── context_menu.go (313 lines)  - Right-click context menu
├── favorites.go (150 lines)     - Favorites/bookmarks system
└── helpers.go (69 lines)        - Helper functions for model
```

## Module Responsibilities

### 1. `main.go` - Application Entry Point
**Purpose**: ONLY contains the main() function
**Contents**:
- Creates the Bubbletea program
- Configures terminal options (alt screen, mouse support)
- Runs the application loop
- Handles top-level errors

**Rule**: Never add business logic to this file. It should remain minimal.

### 2. `types.go` - Type Definitions
**Purpose**: All type definitions, structs, enums, and constants
**Contents**:
- `model` struct - main application state
- `fileItem` struct - file/directory representation
- `previewModel` struct - preview pane state
- Enums: `displayMode`, `viewMode`, `focusPane`
- Custom message types: `editorFinishedMsg`

**When to extend**: Add new types here when introducing new data structures or state fields.

**Recent additions**: `treeItem` struct for tree view, favorites fields in model.

### 3. `styles.go` - Visual Styling
**Purpose**: All Lipgloss style definitions
**Contents**:
- `titleStyle` - application title
- `pathStyle` - path display
- `statusStyle` - status bar
- `selectedStyle` - selected item
- `folderStyle`, `fileStyle` - item styling

**When to extend**: Add new styles here when introducing new visual components or changing color schemes.

### 4. `model.go` - Model Management
**Purpose**: Model initialization and layout calculations
**Contents**:
- `initialModel()` - creates the initial application state
- `calculateGridLayout()` - computes grid column layout
- `calculateLayout()` - computes dual-pane widths

**When to extend**: Add new initialization logic or layout calculation functions here.

### 5. `update.go` - Main Update Dispatcher
**Purpose**: Message dispatching and non-input event handling
**Contents**:
- `Init()` - Bubbletea initialization
- `Update()` - Main message dispatcher (calls keyboard/mouse handlers)
- Window resize handling
- Editor/command finished message handling
- Spinner tick handling
- Helper functions: `isSpecialKey()`, `cleanBracketedPaste()`

**When to extend**: Add new message types or top-level event handlers here

### 5a. `update_keyboard.go` - Keyboard Event Handling
**Purpose**: All keyboard input processing
**Contents**:
- `handleKeyEvent()` - Main keyboard event handler
- Preview mode keys (F10, F4, F5, arrow keys, pageup/pagedown)
- Dialog input handling (input/confirm dialogs)
- Context menu keyboard navigation
- Command prompt input (enter, backspace, history)
- All file browser keyboard shortcuts (F1-F10, navigation, display modes)

**When to extend**: Add new keyboard shortcuts or key bindings here

### 5b. `update_mouse.go` - Mouse Event Handling
**Purpose**: All mouse input processing
**Contents**:
- `handleMouseEvent()` - Main mouse event handler
- Left/right click handling
- Double-click detection (navigate folders, preview files)
- Context menu mouse interaction
- Mouse wheel scrolling (file list, preview, context menu)
- Dual-pane click focus switching
- Clickable UI elements (home button)

**When to extend**: Add new mouse interactions or clickable elements here

### 6. `view.go` - View Rendering
**Purpose**: Top-level view dispatching and single-pane rendering
**Contents**:
- `View()` - main render dispatcher
- `renderSinglePane()` - single-pane mode rendering

**When to extend**: Add new view modes here or modify single-pane layout.

### 7. `render_preview.go` - Preview Rendering
**Purpose**: File preview rendering for all preview modes
**Contents**:
- `renderPreview()` - preview pane content with line numbers
- `renderFullPreview()` - full-screen preview mode
- `renderDualPane()` - split-pane layout

**When to extend**: Modify preview rendering logic, add preview features (syntax highlighting, etc.).

### 8. `render_file_list.go` - File List Rendering
**Purpose**: File list rendering in all display modes
**Contents**:
- `renderListView()` - simple list view
- `renderGridView()` - grid layout view
- `renderDetailView()` - detailed view with metadata (default)
- `renderTreeView()` - expandable hierarchical tree view
- `buildTreeItems()` - recursively builds tree with expanded folders

**When to extend**: Add new display modes or modify existing view layouts.

### 9. `file_operations.go` - File Operations
**Purpose**: All file system operations and formatting
**Contents**:
- `loadFiles()` - reads directory contents
- `loadSubdirFiles()` - loads subdirectory for tree view expansion
- `loadPreview()` - loads file for preview
- Icon mapping functions (`getFileIcon()`, `getIconForExtension()`)
- Formatting functions (`formatFileSize()`, `formatModTime()`)
- File type detection (`isBinaryFile()`, `isClaudeContextFile()`)

**When to extend**:
- Add new file operations (copy, move, delete) here
- Add new file type detection logic
- Add new formatting utilities

### 10. `editor.go` - External Tool Integration
**Purpose**: External tool launching and integration (editors, browsers, clipboard)
**Contents**:
- **Editor functions:**
  - `getAvailableEditor()` - finds available editors (micro, nano, vim, vi)
  - `editorAvailable()` - checks if specific editor exists
  - `openEditor()` - launches editor
- **Browser functions:**
  - `isImageFile()` - detects image files (.png, .jpg, .gif, .svg, etc.)
  - `isHTMLFile()` - detects HTML files (.html, .htm)
  - `isBrowserFile()` - combined check for browser-openable files
  - `getAvailableBrowser()` - platform detection (wslview, cmd.exe, xdg-open, open)
  - `openInBrowser()` - launches file in default browser
- **Clipboard:**
  - `copyToClipboard()` - clipboard integration (termux-api, xclip, xsel, pbcopy, clip.exe)
- **TUI tools:**
  - `openTUITool()` - launches TUI applications (lazygit, htop, etc.)

**When to extend**: Add new editor/browser support, clipboard features, or TUI tool integrations.

### 11. `favorites.go` - Favorites System
**Purpose**: Bookmarking files and directories
**Contents**:
- `loadFavorites()` / `saveFavorites()` - persistence to ~/.config/tfe/favorites.json
- `toggleFavorite()` - add/remove favorites
- `getFilteredFiles()` - filter by favorites

**When to extend**: Add favorite management features (import/export, categories, etc.).

### 12. `helpers.go` - Helper Functions
**Purpose**: Utility functions for model operations
**Contents**:
- `getCurrentFile()` - gets selected file (handles tree view expansion)
- `getMaxCursor()` - calculates cursor bounds for current display mode

**When to extend**: Add reusable helper functions that don't fit other modules.

## Development Guidelines

### Adding New Features

When adding new features, follow this decision tree:

1. **Is it a new type or data structure?** → Add to `types.go`
2. **Is it a visual style?** → Add to `styles.go`
3. **Is it event handling (keyboard/mouse)?** → Add to `update_keyboard.go` or `update_mouse.go`
4. **Is it a rendering function?** → Add to `view.go`, `render_preview.go`, or `render_file_list.go`
5. **Is it a file operation?** → Add to `file_operations.go`
6. **Is it external tool integration?** → Add to `editor.go` or create new module
7. **Is it complex enough to need its own module?** → Create a new file

### Creating New Modules

If a feature is substantial enough to warrant its own module:

1. **Create a new `.go` file** with a descriptive name (e.g., `search.go`, `bookmarks.go`)
2. **Keep it in `package main`** (all files share the same package)
3. **Document the module's purpose** at the top of the file
4. **Add it to this document** under "Module Responsibilities"

Example structure for a new module:

```go
package main

// Module: search.go
// Purpose: File and content search functionality
// Responsibilities:
// - Search indexing
// - Pattern matching
// - Search result filtering

import (
    // ... imports
)

// ... implementation
```

### Code Organization Principles

1. **Single Responsibility**: Each file should have one clear purpose
2. **Keep main.go minimal**: Only the entry point belongs here
3. **Group related functions**: Keep related functionality together
4. **Separate concerns**: UI rendering separate from business logic
5. **DRY (Don't Repeat Yourself)**: Extract common logic into helper functions
6. **Clear naming**: File names should immediately convey their purpose

### Testing Strategy

When adding tests (future):
- Create corresponding `*_test.go` files alongside each module
- Test files should mirror the structure: `file_operations_test.go`, `render_preview_test.go`, etc.
- Keep test files focused on their corresponding module

## Common Patterns

### Adding a New Keyboard Shortcut

1. Go to `update_keyboard.go`
2. Find the appropriate switch statement (preview mode vs regular mode)
3. Add a new case for your key
4. Implement the logic or call a function from another module

Example:
```go
case "s":
    // Save bookmark
    m.saveBookmark(m.files[m.cursor].path)
```

### Adding a New Display Mode

1. Add the enum to `types.go`:
```go
const (
    modeList displayMode = iota
    modeGrid
    modeDetail
    modeTree
    modeYourNewMode  // Add here
)
```

2. Add rendering function to `render_file_list.go`:
```go
func (m model) renderYourNewMode(maxVisible int) string {
    // Implementation
}
```

3. Update the switch in `render_file_list.go` or `view.go` to call your renderer

4. Add keyboard shortcut in `update_keyboard.go` if needed

### Adding a New File Operation

1. Add the function to `file_operations.go`:
```go
func (m *model) yourNewOperation(path string) error {
    // Implementation
    return nil
}
```

2. Add keyboard shortcut in `update_keyboard.go` to call it:
```go
case "x":
    if err := m.yourNewOperation(m.files[m.cursor].path); err != nil {
        // Handle error
    }
```

## Refactoring History

This modular architecture was achieved through a systematic refactoring process:

- **Original**: Single `main.go` file with 1668 lines
- **Phases 1-4**: Extracted types, styles, file operations, editor integration (Commit: 9befa48)
- **Phase 5**: Extracted file list rendering functions (Commit: 3d992c6)
- **Phase 6**: Extracted preview rendering and view functions (Commit: 49d6ece)
- **Phase 7**: Extracted Update and Init functions (Commit: 03efd5c)
- **Phase 8**: Extracted model initialization and layout (Commit: 68d5a87)
- **Phase 9**: Split `update.go` (1145 lines) into 3 focused files:
  - `update.go` (111 lines) - dispatcher only
  - `update_keyboard.go` (714 lines) - keyboard handling
  - `update_mouse.go` (383 lines) - mouse handling
- **Final**: `main.go` reduced to 21 lines, all modules under 800 lines

## Benefits of This Architecture

1. **Maintainability**: Easy to locate and modify specific functionality
2. **Readability**: Each file is focused and easier to understand
3. **Collaboration**: Multiple developers can work on different modules
4. **Testing**: Isolated modules are easier to test
5. **Scalability**: New features can be added without cluttering existing code
6. **Navigation**: IDE features work better with smaller, focused files

## Important Reminder

**🚨 When adding new features, always maintain this modular architecture!**

Do NOT add complex logic to `main.go`. Instead:
- Identify which module the feature belongs to
- Add it to that module, or create a new one
- Keep files focused and organized
- Update this document when creating new modules

This architecture took significant effort to establish - let's maintain it! 🏗️

---

## Documentation Management

**Problem:** Documentation files can grow too large, making them hard to read, navigate, and load into AI context. This leads to "documentation bloat" that makes projects unmaintainable.

**Solution:** Strict line limits and archiving rules for all documentation files.

### Core Documentation Files

These files live in the project root and should be kept concise:

| File | Max Lines | Purpose | When to Clean |
|------|-----------|---------|---------------|
| **CLAUDE.md** | 500 | Architecture guide for AI assistants | Archive old sections to `docs/archive/` |
| **README.md** | 400 | Project overview, installation, usage | Split detailed docs to `docs/` |
| **PLAN.md** | 400 | Current roadmap & planned features | Move completed items to CHANGELOG.md |
| **CHANGELOG.md** | 300 | Recent changes & release notes | Archive old versions to `docs/archive/CHANGELOG_YYYY.md` |
| **BACKLOG.md** | 300 | Ideas & future features (brainstorming) | Move refined ideas to PLAN.md or archive |
| **HOTKEYS.md** | 200 | User-facing keyboard shortcuts | Should rarely grow |

### Documentation Workflow

**1. Idea Stage → BACKLOG.md**
- Raw ideas, brainstorming, "nice to have" features
- Things that need more research or aren't prioritized yet
- Parking lot for concepts that don't warrant PLAN.md yet

**2. Planning Stage → PLAN.md**
- Refined ideas with clear requirements
- Prioritized features ready for implementation
- When item is completed → Move to CHANGELOG.md

**3. Implementation Stage → docs/NEXT_SESSION.md**
- Detailed implementation plans for current work
- Session-specific notes and checklists
- After completion → Delete or archive

**4. Completion Stage → CHANGELOG.md**
- Brief description of what was implemented
- Version number, date, key changes
- When file exceeds 300 lines → Archive old versions

**5. Research Notes → docs/**
- Research documents can be large but should be split by topic
- One topic per file (e.g., `RESEARCH_UI_FRAMEWORKS.md`)
- Archive when no longer relevant

### Archiving Rules

**When to archive:**
- CHANGELOG.md exceeds 300 lines → Move entries older than 6 months to `docs/archive/CHANGELOG_2024.md`
- PLAN.md exceeds 400 lines → Move completed items to CHANGELOG.md, defer low-priority items to BACKLOG.md
- BACKLOG.md exceeds 300 lines → Archive old/rejected ideas to `docs/archive/BACKLOG_OLD.md`
- Research docs exceed 1000 lines → Split into multiple focused docs or archive outdated sections

**Archive structure:**
```
docs/
├── archive/
│   ├── CHANGELOG_2024.md
│   ├── BACKLOG_2024.md
│   ├── RESEARCH_OLD.md
│   └── ...
├── NEXT_SESSION.md (current work)
├── RESEARCH_XYZ.md (active research)
└── ...
```

### AI Assistant Reminders

**For Claude Code:**
- If any core doc exceeds its line limit during a session, proactively suggest cleanup
- When adding to PLAN.md, check if it's grown too large
- Suggest moving completed PLAN.md items to CHANGELOG.md
- Keep NEXT_SESSION.md focused on current work only

**Checking file sizes:**
```bash
wc -l *.md docs/*.md
```

### Benefits of This System

✅ **AI Context Efficiency** - Smaller files load faster and fit in context windows
✅ **Human Readability** - Easier to scan and find information
✅ **Project Maintainability** - Clear separation between active and archived info
✅ **Prevents Bloat** - Proactive limits prevent files from becoming unmanageable
✅ **Clear Workflow** - Know exactly where each piece of information belongs

### Current Status (as of 2025-10-16)

Recent line counts:
- CLAUDE.md: 408 lines ✅ (under 500 limit)
- PLAN.md: 339 lines ✅ (under 400 limit - cleaned up Phase 1)
- CHANGELOG.md: 254 lines ✅ (under 300 limit)
- BACKLOG.md: 97 lines ✅ (newly created)
- README.md: 375 lines ✅ (under 400 limit)

**Status:** ✅ All documentation is within limits! Phase 1 completion moved to CHANGELOG.md.
