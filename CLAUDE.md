# TFE Architecture & Development Guide

This document describes the architecture of the TFE (Terminal File Explorer) project and provides guidelines for maintaining and extending the codebase.

## Architecture Overview

TFE follows a **modular architecture** where each file has a single, clear responsibility. This organization was established through a comprehensive refactoring that reduced `main.go` from 1668 lines to just 21 lines, distributing functionality across 11 focused modules.

### Core Principle

**When adding new features, always maintain this modular architecture by creating new files or extending existing modules rather than adding everything to `main.go`.**

### ⚠️ Important: Read LESSONS_LEARNED.md First

Before modifying UI rendering code (especially anything involving width calculations, scrolling, or alignment), **read `docs/LESSONS_LEARNED.md`**. It contains critical lessons about:
- Visual width vs byte length (never use `len()` for display text)
- Terminal-specific rendering differences (WezTerm vs Windows Terminal)
- ANSI escape code handling
- Header vs data row alignment
- Common pitfalls and how to avoid them

**Ignoring these lessons WILL cause bugs.** Many hours have been spent debugging width calculation issues, text wrapping, and scrolling bugs. Learn from these mistakes!

## File Structure

```
tfe/
├── main.go - Entry point ONLY
├── types.go - Type definitions & enums
├── styles.go - Lipgloss style definitions
├── model.go - Model initialization & layout calculations
├── update.go - Main update dispatcher & initialization
├── update_keyboard.go - Keyboard event handling
├── update_mouse.go - Mouse event handling
├── view.go - View dispatcher & single-pane rendering
├── menu.go - Menu bar rendering
├── render_preview.go - Preview rendering (full & dual-pane)
├── render_file_list.go - File list views (List/Detail/Tree)
├── file_operations.go - File operations & formatting
├── editor.go - External editor integration
├── command.go - Command execution system
├── git_operations.go - Git repository operations (pull, push, sync, fetch)
├── dialog.go - Dialog system (input/confirm)
├── context_menu.go - Right-click context menu
├── favorites.go - Favorites/bookmarks system
├── trash.go - Trash/recycle bin system
├── prompt_parser.go - Prompt template variable parsing
├── fuzzy_search.go - Fuzzy file search (Ctrl+P)
├── terminal_graphics.go - HD image preview via terminal protocols
└── helpers.go - Helper functions for model
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
- `renderPromptPreview()` - prompt file preview with metadata header

**Critical: Preventing Height Overflow**

To keep dual-pane boxes vertically aligned, ALL content must fit exactly within `maxVisible` lines:

1. **Use `visualWidth()` not `len()`**: When deciding whether to wrap text, ALWAYS use `visualWidth()` instead of byte length. Emojis (📝, 🌐) and Unicode characters have visual width ≠ byte length, causing premature/missed wrapping.

2. **Wrap ALL content except pre-rendered markdown**: Never skip wrapping for ANSI-styled text (colored variables, edit mode). Long lines will terminal-wrap and add extra visual rows. Only skip wrapping for Glamour-rendered markdown (already wrapped).

3. **Use `truncateToWidth()` for force-breaks**: When breaking long words, use `truncateToWidth(word, width)` not `word[:width]`. Byte-position slicing breaks mid-ANSI-code (e.g., `\033[38;5;220m`), corrupting output.

4. **Truncate after padding**: If adding padding (e.g., `"  " + line`), truncate the final result to `boxContentWidth` to prevent exceeding box bounds.

5. **Accumulate then join once**: Build all output in `renderedLines` slice, then join with newlines (no trailing newline). Clamp to exactly `maxVisible` lines with padding loop.

**When to extend**: Modify preview rendering logic, add preview features (syntax highlighting, etc.).

### 8. `render_file_list.go` - File List Rendering
**Purpose**: File list rendering in all display modes
**Contents**:
- `renderListView()` - simple list view
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

### 11. `command.go` - Command Execution & History
**Purpose**: Command prompt system with persistent history
**Contents**:
- `executeCommand()` - runs shell commands in current directory
- `addToHistory()` - adds commands to history (max 100)
- `getPreviousCommand()` / `getNextCommand()` - history navigation
- `loadCommandHistory()` - loads history from ~/.config/tfe/command_history.json
- `saveCommandHistory()` - persists history to disk (JSON format)
- Command cursor editing (position tracking, insertion, deletion)

**When to extend**: Add command completion, aliases, or advanced command features.

**Recent additions**: Full cursor editing with word jumping, persistent history across restarts.

### 12. `git_operations.go` - Git Repository Operations
**Purpose**: Git workspace management with visual triage and quick operations
**Contents**:
- **Git operations:**
  - `gitPull()` - Execute git pull with feedback
  - `gitPush()` - Execute git push with error handling
  - `gitSync()` - Smart pull + push workflow
  - `gitFetch()` - Update remote tracking branches
- **Message types:**
  - `gitOperationFinishedMsg` - Operation completion notification
- **Integration:**
  - Context menu integration for git repositories
  - Auto-refresh after operations complete
  - Status message display for success/failure

**When to extend**: Add more git operations (stash, branch, merge, etc.), implement conflict resolution UI.

**Recent additions**: Full git workspace management system with visual status indicators in git repos view.

### 13. `favorites.go` - Favorites System
**Purpose**: Bookmarking files and directories
**Contents**:
- `loadFavorites()` / `saveFavorites()` - persistence to ~/.config/tfe/favorites.json
- `toggleFavorite()` - add/remove favorites
- `getFilteredFiles()` - filter by favorites

**When to extend**: Add favorite management features (import/export, categories, etc.).

### 14. `helpers.go` - Helper Functions
**Purpose**: Utility functions for model operations
**Contents**:
- `getCurrentFile()` - gets selected file (handles tree view expansion)
- `getMaxCursor()` - calculates cursor bounds for current display mode

**When to extend**: Add reusable helper functions that don't fit other modules.

### 15. `trash.go` - Trash/Recycle Bin System
**Purpose**: Move files to trash instead of permanent deletion
**Contents**: `moveToTrash()`, `restoreFromTrash()`, trash metadata (JSON), cross-platform directory detection

### 16. `prompt_parser.go` - Prompt Template Parsing
**Purpose**: Parse and render prompt templates with {{VARIABLE}} substitution
**Contents**: `parsePromptVariables()`, `classifyVariableType()`, `substituteVariables()`, auto-fill for DATE/TIME/FILE/DIRECTORY

### 17. `fuzzy_search.go` - Fuzzy File Search
**Purpose**: Ctrl+P fuzzy search using external fzf + fd/find
**Contents**:
- `getFileFinder()` - Auto-detects best file finder (fd > fdfind > find)
- `launchFuzzySearch()` - Launches fzf with file list pipeline
- `navigateToFuzzyResult()` - Navigates to selected file

**Dependencies**: Requires `fzf` (external), uses `fd`/`fdfind`/`find` for file discovery

**Performance**: Instant search with no lag, searches entire directory tree recursively

**When to extend**: Add fzf options, customize preview, add file type filtering

### 18. `terminal_graphics.go` - Terminal Graphics Protocol Support
**Purpose**: HD image preview rendering using terminal graphics protocols
**Contents**:
- **Protocol detection:**
  - `detectTerminalProtocol()` - Auto-detects Kitty/iTerm2/Sixel support
  - `getProtocolName()` - Returns human-readable protocol name
- **Image rendering:**
  - `renderImageWithProtocol()` - Main entry point for HD image rendering
  - `loadImageFile()` - Loads PNG/JPG/GIF/WebP images
  - `scaleImage()` - Scales images to fit preview pane dimensions
- **Protocol encoders:**
  - `encodeKittyImage()` - Kitty graphics protocol (WezTerm, Kitty)
  - `encodeITerm2Image()` - iTerm2 inline images protocol
  - `encodeSixelImage()` - Sixel protocol (xterm, mlterm, foot)

**Dependencies**: Uses `github.com/BourgeoisBear/rasterm` for protocol encoding

**When to extend**: Add new terminal protocols, improve scaling algorithms, add image format support

**Supported terminals**: WezTerm (Kitty), Kitty (native), iTerm2 (macOS), xterm/mlterm/foot (Sixel)

### 19. `menu.go` - Menu Bar Rendering
**Purpose**: Renders the top menu bar with clickable emoji buttons
**Contents**: `renderMenuBar()`, button definitions, width-aware rendering for narrow terminals

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

### Emoji Usage Rules

**⚠️ CRITICAL: Never use emoji variation selectors (U+FE0F / U+FE0E) in the codebase!**

**Why this matters:**
- go-runewidth has a bug (#76) where variation selectors are incorrectly counted as width=1 instead of width=0
- This causes misalignment in terminal width calculations
- Different terminals (WezTerm, Termux, Windows Terminal, etc.) render emoji+VS inconsistently
- Adding workaround code for each terminal type creates maintenance burden

**Rule:**
- ✅ **Always use base emoji characters without variation selectors**
- ❌ **Never use:** `"🗑️"` (U+1F5D1 + U+FE0F)
- ✅ **Instead use:** `"🗑"` (U+1F5D1 alone)

**Examples:**
```go
// ❌ WRONG - Has variation selector
icon := "⚙️"  // U+2699 + U+FE0F
trash := "🗑️" // U+1F5D1 + U+FE0F

// ✅ CORRECT - Base emoji only
icon := "⚙"   // U+2699
trash := "🗑"  // U+1F5D1
```

**How to check:**
- If you copy an emoji from a website, it may include variation selectors
- Use a Unicode inspector or run: `echo -n "🗑️" | xxd` to check for U+FE0F bytes
- Most emojis look identical with or without variation selectors

**Visual impact:** Minimal to none! Base emojis render the same in 99% of cases.

**Benefits:**
- Universal compatibility across all terminals
- No terminal-specific workaround code needed
- Simpler, more maintainable codebase
- Faster rendering (no string replacement overhead)

### Testing Strategy

When adding tests (future):
- Create corresponding `*_test.go` files alongside each module
- Test files should mirror the structure: `file_operations_test.go`, `render_preview_test.go`, etc.
- Keep test files focused on their corresponding module

## Common Patterns

### Modifying the Header/Title Bar

**⚠️ IMPORTANT: Headers exist in TWO locations!**

When modifying the header/title bar (GitHub link, menu bar, mode indicators), you must update BOTH:
- **Single-Pane:** `view.go` → `renderSinglePane()` (~line 64)
- **Dual-Pane:** `render_preview.go` → `renderDualPane()` (~line 816)

**Note:** `renderFullPreview()` has a different header intentionally (shows filename, not menu bar).

**Why this matters:** Forgetting to update both locations leads to inconsistent UI between view modes. Extract shared header rendering to fix this duplication (see PLAN.md issue #14).

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

This modular architecture was achieved through a systematic refactoring process that reduced `main.go` from 1668 lines to just 21 lines.

**See [docs/REFACTORING_HISTORY.md](docs/REFACTORING_HISTORY.md) for the complete timeline and lessons learned.**

## Benefits of This Architecture

**Maintainability** - Easy to locate/modify specific functionality | **Readability** - Focused files | **Collaboration** - Multiple devs, no conflicts | **Testing** - Isolated modules | **Scalability** - Add features without clutter | **Navigation** - Better IDE support

## Important Reminder

**🚨 When adding new features, always maintain this modular architecture!**

Do NOT add complex logic to `main.go`. Instead:
- Identify which module the feature belongs to
- Add it to that module, or create a new one
- Keep files focused and organized
- Update this document when creating new modules
- When this document gains new architectural context, mirror the contributor-facing highlights in `AGENTS.md`

This architecture took significant effort to establish - let's maintain it! 🏗️

---

## Security & Threat Model

**TFE is a local terminal file manager where the user is the operator, not an attacker.**

### Why Common "Security Issues" Don't Apply

**"Command Injection" in command prompt:**
- ✅ **Not a vulnerability** - User is already in a terminal with full shell access
- Users can run ANY command they want directly (e.g., `rm -rf ~`)
- Sanitizing commands would just add friction with no security benefit
- This is like saying `bash` has a "command injection vulnerability"

**"Path Traversal" in file navigation:**
- ✅ **Not a vulnerability** - That's the entire point of a file browser
- Users can navigate to any path they have permissions for (e.g., `../../etc/passwd`)
- Blocking this would make TFE completely useless
- The OS enforces file permissions, not TFE

**History/Favorites file permissions (0644 vs 0600):**
- ✅ **Minor privacy consideration** - Not a security vulnerability
- No privilege escalation, no data theft from other users
- Users on shared systems can manually `chmod 600` if desired

### Actual Bug Fixed

**File handle leak (2025-10-24):**
- ✅ **Fixed** - All `os.Open()` calls now have `defer file.Close()`
- This was a resource leak bug, not a security issue
- Files: `file_operations.go:119,647,2147` and `trash.go:347,354`

**Threat Model:** TFE assumes the user has legitimate access to their system. It's not designed to defend against a malicious user attacking themselves on their own machine.

---

## Documentation Management

**Problem:** Documentation files can grow too large, making them hard to read, navigate, and load into AI context. This leads to "documentation bloat" that makes projects unmaintainable.

**Solution:** Strict line limits and archiving rules for all documentation files.

### Core Documentation Files

These files live in the project root and should be kept concise:

| File | Max Lines | Purpose | When to Clean |
|------|-----------|---------|---------------|
| **CLAUDE.md** | 500 | Architecture guide for AI assistants | Archive old sections to `docs/archive/` |
| **README.md** | 600 | Project overview, installation, usage | Split detailed docs to `docs/` (user-facing, can be longer) |
| **PLAN.md** | 400 | Current roadmap & planned features | Move completed items to CHANGELOG.md |
| **CHANGELOG.md** | 350 | Recent changes & release notes | Create CHANGELOG2.md when exceeds limit |
| **BACKLOG.md** | 300 | Ideas & future features (brainstorming) | Move refined ideas to PLAN.md or archive |
| **HOTKEYS.md** | - | User-facing keyboard shortcuts | Keep the list comprehensive and current |

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
- When file exceeds 350 lines → Create CHANGELOG2.md (see below)

**5. Research Notes → docs/**
- Research documents can be large but should be split by topic
- One topic per file (e.g., `RESEARCH_UI_FRAMEWORKS.md`)
- Archive when no longer relevant

### Managing File Growth

**CHANGELOG Approach (Keep History Visible):**
- When CHANGELOG.md exceeds 350 lines, create CHANGELOG2.md
- Move older entries (v0.1.x, v0.2.x, etc.) to CHANGELOG2.md
- Keep recent versions (latest 3-4) in CHANGELOG.md
- Continue pattern: CHANGELOG3.md, CHANGELOG4.md, etc. as needed
- All files remain in project root for easy access
- Link between files: "See CHANGELOG2.md for older versions"

**Example structure when splitting:**
```
CHANGELOG.md      → v0.5.0, v0.4.0, v0.3.0 (current + recent)
CHANGELOG2.md     → v0.2.0, v0.1.5, v0.1.0 (older versions)
```

**Other files:**
- PLAN.md exceeds 400 lines → Move completed items to CHANGELOG.md, defer low-priority items to BACKLOG.md
- BACKLOG.md exceeds 300 lines → Archive old/rejected ideas to `docs/archive/BACKLOG_OLD.md`
- Research docs exceed 1000 lines → Split into multiple focused docs or archive outdated sections

### AI Assistant Reminders

**For Claude Code:**
- If any core doc exceeds its line limit during a session, proactively suggest cleanup
- When CHANGELOG.md exceeds 350 lines, create CHANGELOG2.md and move older entries
- When adding to PLAN.md, check if it's grown too large
- Suggest moving completed PLAN.md items to CHANGELOG.md
- Keep NEXT_SESSION.md focused on current work only
- Check file sizes: `wc -l *.md docs/*.md`

### Benefits of This System

**AI Context Efficiency** - Smaller files load faster | **Human Readability** - Easier to scan | **Project Maintainability** - Clear separation | **Prevents Bloat** - Proactive limits | **Clear Workflow** - Know where info belongs

### Current Status (as of 2025-10-24)

Documentation health:
- CLAUDE.md: 520 lines ⚠️ (104% of 500 limit - added threat model section)
- PLAN.md: 234 lines ✅ (59% of 400 limit - removed false security items)
- CHANGELOG.md: 316 lines ✅ (90% of 350 limit)
- BACKLOG.md: 97 lines ✅ (32% of 300 limit)

**Status:** ⚠️ CLAUDE.md slightly over limit (added important threat model clarification). Consider archiving older sections if adding more content.
