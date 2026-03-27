# TFE Module Details

This document provides detailed descriptions of all TFE modules, their responsibilities, and when to extend them.

## Overview

TFE follows a modular architecture with 27+ specialized modules, each handling a specific aspect of the application. This document serves as the comprehensive reference for understanding what each module does.

---

## Core Modules

### 1. `main.go` - Application Entry Point
**Purpose**: ONLY contains the main() function

**Contents**:
- Creates the Bubbletea program
- Configures terminal options (alt screen, mouse support)
- Runs the application loop
- Handles top-level errors

**Rule**: Never add business logic to this file. It should remain minimal.

---

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

---

### 3. `styles.go` - Visual Styling
**Purpose**: All Lipgloss style definitions

**Contents**:
- `titleStyle` - application title
- `pathStyle` - path display
- `statusStyle` - status bar
- `selectedStyle` - selected item
- `folderStyle`, `fileStyle` - item styling

**When to extend**: Add new styles here when introducing new visual components or changing color schemes.

---

### 4. `model.go` - Model Management
**Purpose**: Model initialization and layout calculations

**Contents**:
- `initialModel()` - creates the initial application state
- `calculateLayout()` - computes dual-pane widths

**When to extend**: Add new initialization logic or layout calculation functions here.

---

## Update & Event Handling Modules

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

---

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

---

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

---

## Rendering Modules

### 6. `view.go` - View Rendering
**Purpose**: Top-level view dispatching and single-pane rendering

**Contents**:
- `View()` - main render dispatcher
- `renderSinglePane()` - single-pane mode rendering

**When to extend**: Add new view modes here or modify single-pane layout.

---

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

**Important**: Many of these rendering lessons are also documented in `docs/LESSONS_LEARNED.md` - read that file before modifying UI code!

---

### 8. `render_file_list.go` - File List Rendering
**Purpose**: File list rendering in all display modes

**Contents**:
- `renderListView()` - simple list view
- `renderDetailView()` - detailed view with metadata (default)
- `renderTreeView()` - expandable hierarchical tree view
- `buildTreeItems()` - recursively builds tree with expanded folders

**When to extend**: Add new display modes or modify existing view layouts.

---

### 19. `menu.go` - Menu Bar Rendering
**Purpose**: Renders the top menu bar with clickable emoji buttons

**Contents**: `renderMenuBar()`, button definitions, width-aware rendering for narrow terminals

---

## File System & Operations Modules

### 9. `file_operations.go` - File Operations
**Purpose**: File loading, preview, and CRUD operations

**Contents**:
- `loadFiles()` - reads directory contents
- `loadSubdirFiles()` - loads subdirectory for tree view expansion
- `loadPreview()` - loads file for preview
- File classification helpers (`isClaudeContextFile()`, `isInPromptsDirectory()`)
- File sorting and filtering

**When to extend**:
- Add new file operations (copy, move, delete) here
- Add new file loading or preview logic

---

### 9a. `file_icons.go` - File Type Detection & Icons
**Purpose**: File type classification, icons, and metadata utilities

**Contents**:
- `isPromptFile()` - prompt file detection
- `isObsidianVault()` - vault detection
- File type checks (`isTextFile()`, `isImageFile()`, `isVideoFile()`, `isAudioFile()`, etc.)
- `getFileIcon()` - icon assignment based on file type/extension
- `getFileType()` - file type string based on extension
- `formatFileSize()`, `formatModTime()` - metadata formatting
- `isDirEmpty()`, `getDirItemCount()` - directory utilities
- `isMarkdownFile()`, `isBinaryFile()` - format detection
- `highlightCode()` - syntax highlighting support

**When to extend**:
- Add new file type detection or classification
- Add new icon mappings
- Add file metadata formatting utilities

---

### 9b. `text_wrapping.go` - Width Calculations & Text Wrapping
**Purpose**: Visual width calculations, text wrapping, and truncation

**Contents**:
- `visualWidth()` - accurate visual width using go-runewidth
- `visualWidthCompensated()` - width with emoji compensation
- `truncateToWidth()` - safe text truncation preserving ANSI codes
- `truncateToWidthCompensated()` - truncation with compensation
- `padIconToWidth()`, `padToVisualWidth()` - padding utilities
- `wrapLine()`, `getWrappedLineCount()` - line wrapping

**When to extend**:
- Add new width-aware text utilities
- Modify wrapping or truncation behavior

---

### 13. `favorites.go` - Favorites System
**Purpose**: Bookmarking files and directories

**Contents**:
- `loadFavorites()` / `saveFavorites()` - persistence to ~/.config/tfe/favorites.json
- `toggleFavorite()` - add/remove favorites
- `getFilteredFiles()` - filter by favorites

**When to extend**: Add favorite management features (import/export, categories, etc.).

---

### 15. `trash.go` - Trash/Recycle Bin System
**Purpose**: Move files to trash instead of permanent deletion

**Contents**: `moveToTrash()`, `restoreFromTrash()`, trash metadata (JSON), cross-platform directory detection

---

## External Integration Modules

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

---

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

---

### 12. `git_operations.go` - Git Repository Operations
**Purpose**: Git status queries and workspace management operations

**Contents**:
- **Git status queries:**
  - `isGitRepo()`, `getGitBranch()`, `hasUncommittedChanges()`
  - `getGitStatus()`, `getAheadBehindCounts()`, `checkPackedRefs()`
  - `getLastCommitInfo()`, `formatGitStatus()`, `formatLastCommitTime()`
  - `scanGitReposRecursive()`, `getGitStatusSortValue()`, `sortGitReposList()`
- **Git operations:**
  - `gitPull()` - Execute git pull with feedback
  - `gitPush()` - Execute git push with error handling
  - `gitSync()` - Smart pull + push workflow
  - `gitFetch()` - Update remote tracking branches
- **Message types:**
  - `gitOperationFinishedMsg` - Operation completion notification

**When to extend**: Add more git operations (stash, branch, merge, etc.), implement conflict resolution UI.

---

## UI Components Modules

### 18. `dialog.go` - Dialog System
**Purpose**: Input and confirmation dialogs

**Contents**: Dialog types, input handling, confirmation prompts

---

### 20. `context_menu.go` - Right-Click Context Menu
**Purpose**: Context-sensitive right-click menu

**Contents**: Context menu rendering, action handling, mouse/keyboard navigation

---

## Search & Navigation Modules

### 17. `fuzzy_search.go` - Fuzzy File Search
**Purpose**: Ctrl+P fuzzy search using external fzf + fd/find

**Contents**:
- `getFileFinder()` - Auto-detects best file finder (fd > fdfind > find)
- `launchFuzzySearch()` - Launches fzf with file list pipeline
- `navigateToFuzzyResult()` - Navigates to selected file

**Dependencies**: Requires `fzf` (external), uses `fd`/`fdfind`/`find` for file discovery

**Performance**: Instant search with no lag, searches entire directory tree recursively

**When to extend**: Add fzf options, customize preview, add file type filtering

---

## Advanced Features Modules

### 16. `prompt_parser.go` - Prompt Template Parsing
**Purpose**: Parse and render prompt templates with {{VARIABLE}} substitution

**Contents**: `parsePromptVariables()`, `classifyVariableType()`, `substituteVariables()`, auto-fill for DATE/TIME/FILE/DIRECTORY

---

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

---

## Utility Modules

### 14. `helpers.go` - Helper Functions
**Purpose**: Utility functions for model operations

**Contents**:
- `getCurrentFile()` - gets selected file (handles tree view expansion)
- `getMaxCursor()` - calculates cursor bounds for current display mode

**When to extend**: Add reusable helper functions that don't fit other modules.

---

### 22. `file_watcher.go` - Live File System Watching
**Purpose**: fsnotify-based file watching with debounce pipeline

**Contents**:
- `initWatcher()` / `startWatcher()` / `stopWatcher()` / `closeWatcher()` - watcher lifecycle
- `runWatcherBridge()` - bridge goroutine with 3-layer debounce (per-file dedup, timer batching, max delay cap)
- `waitForWatcherEvent()` - idiomatic Bubbletea blocking command pattern
- Atomic write detection (editors delete+recreate files within 100ms)

**When to extend**: Add new debounce strategies, watch additional paths (e.g., config files), or add filtering for specific file types.

---

### 23. `theme.go` - Configurable Theme System
**Purpose**: TOML-based theme loading and style initialization

**Contents**:
- `Theme` struct with 15 color fields (defined in types.go)
- `defaultTheme()` - returns original hardcoded colors as fallback
- `loadTheme()` - reads `~/.config/tfe/theme.toml`
- `initTheme()` / `initStyles()` - loads theme and rebuilds all lipgloss styles

**When to extend**: Add new color fields for new UI elements, add theme hot-reload, add theme switching at runtime.

---

### 24. `agent_awareness.go` - AI Agent Session Detection
**Purpose**: Detect and display which AI agent modified files

**Contents**:
- `AgentSession` struct matching `/tmp/claude-code-state/*.json` schema
- `getAgentSessions()` - reads and parses agent state files (graceful fallback)
- `buildAgentFileMap()` - pre-computes path-to-agent-label lookup
- `checkAgentCompletions()` - detects agent active→idle transitions for auto-open
- `agentLabel()` - returns short labels ("CC", "CC:Explore", etc.)

**When to extend**: Add support for other agent state file formats, add agent activity timeline, add per-agent change grouping.

---

### 25. `render_layout.go` - Pane Layout & Tab Bar Rendering
**Purpose**: Layout calculations for dual-pane mode, tab bar rendering

**Contents**:
- Pane border rendering with theme colors
- `renderTabBar()` - styled tab bar with git status indicators and overflow handling
- Dual-pane horizontal/vertical split layout logic

**When to extend**: Add new layout modes, customize tab bar appearance, add split-pane resizing.

---

### 26. `render_prompts.go` - Prompt Template Rendering
**Purpose**: Rendering prompt templates in the preview pane

**When to extend**: Add new prompt template formats or rendering options.

---

### 27. `tmux.go` - Tmux Integration
**Purpose**: Tmux session management and pane splitting

**When to extend**: Add new tmux commands, improve pane layout strategies.

---

## Module Dependencies

**Core chain**: `main.go` → `model.go` → `update.go` → `view.go`

**Event handling**: `update.go` dispatches to `update_keyboard.go` and `update_mouse.go`

**Rendering chain**: `view.go` → `render_preview.go`, `render_file_list.go`, `render_layout.go`, `menu.go`

**Data operations**: Modules use `file_operations.go` for file loading, `file_icons.go` for type detection, `text_wrapping.go` for width calculations

**File watching**: `file_watcher.go` → `update.go` (fileChangedMsg) → `file_operations.go` (loadFiles)

**Agent review pipeline**: `file_watcher.go` → `agent_awareness.go` → `git_operations.go` (getChangedFiles/getFileDiff) → `render_preview.go` (diff rendering)

**Theming**: `theme.go` → `styles.go` (all styles rebuilt from theme)

**Cross-cutting**: `types.go` and `styles.go` are used by all modules
