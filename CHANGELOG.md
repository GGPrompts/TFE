# TFE Changelog

All notable changes to the Terminal File Explorer (TFE) project.

## [Unreleased]

### Added (2025-10-19 Session)
- **Context-Aware F1 Help**
  - F1 now automatically jumps to the most relevant help section based on current mode
  - Smart context detection:
    - Input fields active → "Prompt Templates & Fillable Fields" section
    - Context menu open → "Context Menu" section
    - Command mode → "Command Prompt" section
    - Full preview mode → "Preview & Full-Screen Mode" section
    - Dual-pane mode → "Dual-Pane Mode" section
    - Single-pane mode → "Navigation" section (default)
  - Improves discoverability of relevant shortcuts
  - Users can still manually scroll to other sections if needed
  - Files modified: `helpers.go` (added getHelpSectionName() and findSectionLine()), `update_keyboard.go`
- **Image Viewing & Editing Integration**
  - New context menu options for image files (.png, .jpg, .gif, .bmp, .svg, .webp, .ico, .tiff):
    - **🖼️ View Image**: Opens images in terminal viewer (viu > timg > chafa)
    - **🎨 Edit Image**: Opens images in terminal paint program (textual-paint > durdraw)
  - Smart tool detection with graceful fallbacks
  - Works best in Kitty, iTerm2, WezTerm (graphics protocol support)
  - Fallback to ASCII art in other terminals
  - Files modified: `editor.go`, `context_menu.go`, `HOTKEYS.md`

- **File Operations (v1.0 Critical Features)**
  - **Rename Files/Folders**: Context menu → "✏️ Rename..."
    - Pre-fills current name for easy editing
    - Validation (no empty names, no "/" characters)
    - Cursor automatically moves to renamed item
    - Works for both files and directories
  - **Copy Files/Folders**: Context menu → "📋 Copy to..."
    - Supports absolute and relative destination paths
    - Recursive directory copying with permission preservation
    - Progress feedback via status messages
    - Handles errors gracefully
  - Files modified: `context_menu.go`, `update_keyboard.go`, `file_operations.go`

- **Preview Mode Enhancements**
  - **Mouse Toggle (m key)**: Press 'm' in full-screen preview to toggle mouse
    - Mouse ON: Beautiful border, wheel scrolling, wonky text selection
    - Mouse OFF: Border removed, clean text selection, keyboard scrolling
    - Visual feedback: border disappears to show mode is active
    - Status messages: "🖱️ Mouse ON" / "⌨️ Mouse OFF"
  - **Ctrl-F Search**: Search within file previews
    - Type query for incremental search (case-insensitive)
    - Press 'n' or Enter for next match
    - Press Shift+N for previous match
    - Shows match counter (e.g., "Match 3/15")
    - Auto-scrolls to matches
    - ESC to exit search mode
  - Files modified: `types.go`, `model.go`, `update_keyboard.go`, `render_preview.go`, `helpers.go`, `HOTKEYS.md`

### Fixed (2025-10-19 Session)
- **Browser Opening on WSL**
  - Fixed `cmd.exe /c start` treating first argument as window title
  - Now uses: `cmd.exe /c start "" <path>` (empty title)
  - Browser opening now works correctly for images and HTML files
  - Files modified: `editor.go`

### Added
- **Fillable Fields for Prompt Templates (Phase 5 Complete!)**
  - Automatic detection of `{{VARIABLE}}` placeholders in prompt templates
  - Interactive input fields with smart type classification:
    - **File fields** (blue): For file/path variables - supports F3 file picker
    - **Long fields** (yellow): For multi-line content (code, text, body)
    - **Short fields** (yellow): For single-line input (priority, name, etc.)
    - **Auto-filled fields** (green): Pre-filled with context (DATE, TIME, FILE, DIRECTORY)
  - Tab/Shift+Tab navigation between fields
  - Real-time preview highlighting shows where variables will be inserted
  - Character count display for long content (e.g., "2.5k chars")
  - Ctrl+U to clear field content
  - **F3 File Picker Mode:**
    - Browse and select files to populate input fields
    - Navigate directories with full file browser features
    - Enter to select file, Esc to cancel
    - Double-click support for quick file selection
    - Automatically disables prompts filter to show all files
    - Restores preview and field state when returning
  - F5 copies fully rendered prompt with all variables substituted
  - Paste detection and handling (shows "✓ Pasted X characters")
  - Files modified: `types.go`, `prompt_parser.go`, `render_preview.go`, `update_keyboard.go`, `update_mouse.go`, `view.go`
- **Glamour Markdown Rendering for Prompts**
  - Prompt templates (.md files) now render with beautiful Glamour formatting
  - Full markdown support: headers, lists, code blocks, emphasis, links
  - Variables get substituted first, then Glamour renders the result
  - Smart mode switching: plain text when editing fields, formatted when viewing
  - Graceful fallback to plain text if Glamour fails
  - Files modified: `render_preview.go`
- **Run Script Feature**
  - New context menu option "▶️ Run Script" for executable files
  - Auto-detects executables by extension (.sh, .bash, .zsh, .fish)
  - Auto-detects files with execute permission (chmod +x)
  - Runs script in its directory with full output display
  - "Press any key to continue" prompt to review output
  - Reuses existing `runCommand()` infrastructure
  - Files modified: `context_menu.go`
- **Enhanced Mobile/Termux Documentation**
  - Added Termux to platform badges in README
  - Created full "Mobile & Termux Support" section
  - Documented touch controls and mobile usage tips
  - Added Termux installation guide
  - Updated features list with mobile support

### Fixed
- **Silent Error Handling (Phase 2.1 Complete)**
  - Editor availability: Now shows "No editor available (tried micro, nano, vim, vi)" when F4 pressed without editors
    - Fixed in preview mode (update_keyboard.go:47)
    - Fixed in file browser (update_keyboard.go:695)
    - Fixed in context menu (context_menu.go:173)
  - Clipboard operations: Now shows success/error messages for F5 and context menu copy
    - "Path copied to clipboard" on success
    - "Failed to copy to clipboard: [error]" on failure
    - Fixed in preview mode (update_keyboard.go:65)
    - Fixed in file browser (update_keyboard.go:717)
    - Fixed in context menu (context_menu.go:185)
  - Quick CD: Now shows "Failed to save directory for quick CD: [error]" on write failure
    - Fixed in context menu (context_menu.go:145)
  - All operations now use the existing status message system (auto-dismiss after 3 seconds)
  - No more silent failures - users always get feedback
- **Enter Key Consistency in Prompts Mode**
  - Fixed inconsistent Enter behavior when viewing prompt templates
  - Enter now always previews files (consistent with rest of TFE)
  - F5 copies rendered prompts (clear and obvious)
  - Allows users to see prompts before copying
  - Files modified: `update_keyboard.go`
- **UI Polish and Alignment Fixes**
  - Removed CellBlocks emoji button from dual-pane mode header
  - Fixed preview pane alignment in tree view split-pane mode
  - Fixed mouse click offset issue in dual-pane tree view (clicks now accurate throughout entire file tree)
  - All file list rendering functions now trim trailing newlines for consistent box heights
  - Mouse click calculations now correctly account for dual-pane vs single-pane header differences
  - Files modified: `render_preview.go`, `render_file_list.go`, `update_mouse.go`

### Added
- **Directory Search Feature (Phase 2.2 Complete)**
  - Press `/` to enter search mode and filter files by name
  - Incremental filtering as you type (case-insensitive substring match)
  - ESC to clear search and show all files
  - Enter to accept search and exit input mode (filter remains active)
  - Backspace to delete characters from search query
  - Status bar shows: "Search: [query]█ (X matches)" while typing
  - Status bar shows: "Filtered: [query] (X matches)" when search is accepted
  - Parent directory ".." always visible regardless of search
  - Files modified: `types.go`, `file_operations.go`, `update_keyboard.go`, `favorites.go`, `view.go`
- **Grid View Mouse Click Fix (Phase 2.3 Complete)**
  - Fixed variable-width cell issue in grid view causing click misalignment
  - Problem: Favorite stars (⭐) made some cells 2 chars wider, breaking click detection
  - Solution: Always reserve 2 characters for favorite indicator (space or star)
  - Consistent cell width: icon(2) + fav_indicator(2) + name(12) + padding(2) = 18 chars
  - Updated both left-click and right-click handlers to match rendering
  - Mouse clicks now accurately select grid items with mixed favorite/non-favorite files
  - Files modified: `render_file_list.go`, `update_mouse.go`
- **Fixed Border Rendering in All View Modes**
  - Fixed Lipgloss border rendering issues (right/bottom borders now visible)
  - Switched from MaxWidth/MaxHeight to fixed Width/Height for consistent borders
  - Corrected width calculations to prevent content overflow
  - Borders now stay fixed-size regardless of file content size
  - Applied fixes to single-pane, dual-pane, and full-preview modes
  - Files modified: `render_preview.go`, `view.go`
- **Optimized Preview Content Width**
  - Increased usable preview width by 7 characters in dual-pane mode
  - Markdown files get 6 additional characters (no line numbers/scrollbar)
  - Better text wrapping and reduced horizontal overflow
  - Width now calculated based on actual box dimensions, not estimates
  - Files modified: `render_preview.go`
- **Fixed Dual-Pane Alignment**
  - Both panes now perfectly aligned at top and bottom
  - Removed redundant preview title header (filename already in status bar)
  - Fixed mouse click coordinate offset issues
  - Cleaner UI with less visual clutter
  - Files modified: `render_preview.go`
- **Fuzzy File Search with go-fzf**
  - Ctrl+P or click 🔍 button to launch fuzzy search
  - Search across current directory and subdirectories (depth=1)
  - Keyboard-driven interface (type to filter, arrow keys to navigate)
  - Auto-navigates to selected file/folder on selection
  - Performance optimized: 200 item limit, 8 visible results
  - Returns to TFE with proper terminal state restoration
  - Files modified: `fuzzy_search.go` (new), `types.go`, `update.go`, `update_keyboard.go`, `update_mouse.go`, `view.go`
- **Enhanced UI Borders and Separators**
  - Rounded borders on all panes (single-pane, dual-pane, full-preview)
  - Horizontal separator lines above status bar (connects with pane borders)
  - Adaptive border colors (blue/cyan) matching terminal theme
  - Focus indicators in dual-pane mode (bright blue for active pane)
  - Professional, polished appearance across all view modes
  - Files modified: `view.go`, `render_preview.go`
- **Clickable Column Headers for Sorting (Detail View)**
  - Click column headers (Name, Size, Modified, Type) to sort files
  - Visual indicators: ↑ (ascending) or ↓ (descending) arrows show active sort
  - Click same column again to reverse sort order
  - Smart sorting behavior:
    - **Name sort:** Folders grouped first, then files (traditional behavior)
    - **Other sorts:** Files and folders mixed by sort criteria
  - ".." parent directory always stays at top
  - Cursor maintains position on same file after sorting
  - Sort persists across directory navigation
  - Works in both single-pane and dual-pane modes
  - Files modified: `update_mouse.go`, `file_operations.go`, `render_file_list.go`
- **Browser Support for Images and HTML Files**
  - F3 automatically opens images in default browser (PNG, JPG, GIF, SVG, WebP, BMP, ICO, TIFF)
  - F3 automatically opens HTML files in default browser
  - Context menu "🌐 Open in Browser" option for images and HTML files
  - Platform-aware detection (wslview, cmd.exe, xdg-open, open)
  - Falls back to text preview for non-browser files
  - Cross-platform support (WSL, Linux, macOS)
- **Syntax Highlighting for Code Files**
  - Powered by Chroma v2.14.0 (100+ languages supported)
  - Automatic language detection from file extension
  - Color-coded keywords, strings, comments, functions
  - Works in all preview modes (single-pane, dual-pane, full-screen)
  - Monokai color scheme optimized for dark terminals
  - Fallback to plain text for unknown file types
  - Zero configuration needed
- **Adaptive Colors for Light/Dark Terminals**
  - Automatic adaptation to terminal background color
  - Professional appearance without manual configuration
  - Better readability across different terminal themes
  - Adaptive colors for:
    - Title bar (dark blue/bright cyan)
    - Selected items (high contrast in both modes)
    - Folders (blue tones)
    - Files (black/white)
    - Status bar (gray tones)
    - Claude context files (orange)
- **Rounded Borders for Modern UI**
  - Modern rounded corners for preview pane borders
  - Adaptive border colors for light/dark terminals
  - Enhanced visual polish in dual-pane mode
  - Better visual hierarchy and focus indicators

### Changed
- **Tree View Now Default Display Mode**
  - Changed from Detail view to Tree view as default
  - Reason: Tree/List views work better on narrow terminals
  - Grid/Detail views can have formatting issues with limited width
  - Users can still switch modes with F9 or number keys (1-4)
  - File modified: `model.go`
- **Dynamic Filename Truncation in Tree View**
  - Filenames now use available screen width instead of fixed 25-char limit
  - Calculates width dynamically based on terminal size and view mode
  - Accounts for indentation, tree characters, icons, and favorites
  - Min: 20 chars, Max: 100 chars for optimal readability
  - Works correctly in both single-pane and dual-pane modes
  - File modified: `render_file_list.go`

### Fixed
- **Mouse Click Accuracy with Borders**
  - Fixed mouse clicks offset by border dimensions
  - Y-axis: Account for top border (+1 line) in both single-pane and dual-pane
  - X-axis: Account for left border (-2 chars) in both modes
  - Applied to all click handlers: file selection, column headers, context menu
  - Clicks now register accurately on the intended item/location
  - Files modified: `update_mouse.go`
- **Terminal State After External TUI Apps**
  - Fixed mouse input not working after exiting external TUI applications
  - Issue: Terminal state (including mouse mode) not properly restored
  - Solution: Use `tea.Sequence(tea.ClearScreen, tea.ExecProcess(...))` pattern
  - Ensures proper cleanup and state restoration
  - File modified: `update_mouse.go`
- **Fuzzy Search UI Interference**
  - Fixed background UI scrolling/updating during fuzzy search
  - Fixed typing lag and missing keystrokes in fuzzy search
  - Fixed filter results flickering through background UI
  - Solution: Block all keyboard/mouse events when `fuzzySearchActive` is true
  - Files modified: `update_keyboard.go`, `update_mouse.go`, `view.go`
- **Command Line Paste Bug:** Fixed brackets appearing around pasted text
  - Root cause: Using `msg.String()` which wraps paste events in brackets by design
  - Solution: Use `msg.Runes` to get raw text (Bubble Tea handles escape sequences)
  - Removed unnecessary helper functions (`cleanBracketedPaste`, `isBracketedPasteMarker`)
  - Fixed in: command prompt input, dialog input, and command continuation
  - Credit: Analysis by OpenAI Codex
- **Command Line Input:** Removed 's' key hotkey to allow typing 's' in command prompt
  - 's' key was intercepting command input before reaching the prompt
  - To toggle favorites, use F2 (context menu) or right-click → "☆ Add Favorite"
  - Prioritizes command typing over single-letter shortcuts

### To Be Implemented
- Search/filter functionality within directories
- Multi-select and bulk operations
- Menu bar with dropdowns (File/View/Tools/Help)
- Splash screen on launch

---

## [0.4.0] - 2025-10-16

### Added - File Operations (Phase 1 Complete!)
- **F7: Create Directory**
  - Input dialog with validation (rejects invalid characters)
  - Auto-moves cursor to newly created directory
  - Success/error status messages
  - Available from F7 key or context menu "New Folder"
- **F8: Delete File/Folder**
  - Confirmation dialog (prevents accidents)
  - Safety checks (won't delete parent "..", warns on non-empty dirs)
  - Permission checks (read-only protection)
  - Success/error status messages
  - Available from F8 key or context menu "Delete"
- **Dialog System** (new module: `dialog.go`)
  - Input dialogs for text entry
  - Confirmation dialogs for yes/no prompts
  - Centered on screen with proper positioning
  - Styled with lipgloss (blue for input, red for warnings)
- **Status Message System**
  - Auto-dismissing messages (3 seconds)
  - Green for success, red for errors
  - Shows in status bar (replaces normal status temporarily)
- **Context Menu Enhancements**
  - Added "📁 New Folder..." to directory menus
  - Added "🗑️  Delete" to all file/folder menus
  - Integrates seamlessly with dialog system
- **Documentation Management System**
  - Created BACKLOG.md for brainstorming/ideas
  - Added documentation rules to CLAUDE.md
  - Line limits for all core .md files (prevents bloat)
  - Archiving workflow for old documentation

### Changed
- TFE is now a true file *manager*, not just a viewer
- F7/F8 are no longer placeholders

---

## [0.3.0] - 2025-10-16

### Added - Context Menu & Favorites
- **Context Menu System** (right-click or F2)
  - Quick access to common file operations
  - Different menus for files vs directories
  - Keyboard navigation (up/down/enter/esc)
  - Mouse wheel scrolling support
- **Favorites/Bookmarks System**
  - Toggle favorite with 's' key
  - F6 to filter by favorites only
  - Persistent storage in ~/.config/tfe/favorites.json
  - Visual indicators (star emoji) for favorited items
- **TUI Tool Launcher**
  - Launch lazygit, lazydocker, lnav, htop from context menu
  - Auto-detection of installed tools
  - Smart directory-specific options
- **Quick CD Feature**
  - Exit TFE and change shell to selected directory
  - Requires bash wrapper (tfe-wrapper.sh)
  - Accessible via context menu for directories

### Enhanced
- Text selection enabled in preview mode
- Markdown rendering improvements with Glamour
- Command history now stores last 100 commands
- Bracketed paste support for proper terminal paste handling
- Special key filtering (prevents literal "end", "home" text)

### Fixed
- Mouse coordinate calculations in dual-pane mode
- Context menu positioning and overflow handling
- Command prompt input handling edge cases

---

## [0.2.0] - 2025-10-15

### Added - Dual-Pane & Preview System
- **Dual-Pane Mode**
  - Split-screen layout (40/60 split)
  - Toggle with Space or Tab
  - Click to switch pane focus
  - Visual focus indicators (bright blue borders)
  - Independent scrolling per pane
- **File Preview System**
  - Live preview in right pane (dual-pane mode)
  - Full-screen preview mode (F3 or Enter)
  - Line numbers for text files
  - Markdown rendering with Glamour styling
  - Smart line wrapping at terminal width
  - Binary file detection and warnings
  - Large file detection (>1MB limit)
  - Preview scrolling (arrow keys, PageUp/PageDown)
- **External Editor Integration**
  - F4 to open in preferred editor
  - Priority: Micro → Nano → Vim → Vi
  - 'n' key for nano specifically
  - Auto-detect available editors
  - TFE suspends while editor runs
  - File list auto-refreshes after editing
- **Clipboard Integration**
  - F5 to copy file path to clipboard
  - Multi-platform support (xclip, xsel, pbcopy, clip.exe)
  - Termux support (termux-clipboard-set)
- **F-Key Hotkeys** (Midnight Commander style)
  - F1: Help (HOTKEYS.md reference)
  - F3: Full-screen preview
  - F4: Edit in external editor
  - F5: Copy path to clipboard
  - F9: Cycle display modes
  - F10: Quit application
  - F7/F8: Placeholders for future features

### Added - Command Prompt
- **MC-Style Command Prompt**
  - Always-visible at bottom of screen
  - Start typing to auto-focus
  - Execute any shell command in current directory
  - TFE suspends during command execution
  - File list auto-refreshes after command
  - Command history with up/down arrows
  - ESC to unfocus and clear prompt

### Fixed
- Preview scrolling calculations (consistent height - 7)
- Large file rendering (line truncation prevents overflow)
- Mouse click accuracy with proper header offsets
- go.mod and long-line file rendering issues
- Pane boundary detection for clicks

---

## [0.1.5] - 2025-10-14

### Added - View Modes
- **List View** (default)
  - One file per line, vertical scrolling
  - Icon/marker + filename display
- **Grid View**
  - Multi-column responsive layout
  - Icon-focused display
  - Adapts to terminal width
- **Detail View**
  - Columns: Name, Size, Modified, Type
  - File metadata display (size, date, permissions)
  - Relative time formatting ("5m ago", "2d ago")
  - Directory item counts
- **Tree View**
  - Hierarchical directory structure
  - Expandable/collapsible folders
  - Visual tree indicators
  - Recursive subdirectory loading

### Added - Navigation Enhancements
- Mouse support
  - Single click to select
  - Double-click to open/navigate
  - Scroll wheel support
  - Grid view click detection
- Keyboard shortcuts
  - Number keys (1-4) for direct view mode switching
  - F9 to cycle through view modes
  - Arrow keys and vim keys (hjkl)
- Double-click timing threshold (500ms)

---

## [0.1.0] - 2025-10-13

### Added - Foundation
- **Core File Browser**
  - Directory navigation (up/down arrows, Enter)
  - Parent directory navigation (h or left arrow)
  - File and folder listing with icons
  - Smart sorting (directories first, alphabetical)
- **File Metadata Display**
  - File size (formatted: KB, MB, GB, TB)
  - Modification time (relative format)
  - File permissions in status bar
- **File Type Icons**
  - Extension-based icon mapping
  - 50+ file type indicators (emoji-based)
  - Special icons for folders and parent directory
  - Claude context file highlighting (orange color)
  - Categories: code, configs, images, archives, docs
- **Hidden Files Toggle**
  - '.' or Ctrl+H to toggle
  - Dynamic filtering
  - Status bar indicator when showing hidden files
- **Status Bar**
  - File/folder counts
  - Selected file info (name, size, time)
  - Current view mode indicator
  - Hidden files status
- **Window Resize Handling**
  - Responsive to terminal size changes
  - Proper layout recalculation
  - Works across all view modes
- **Modular Architecture**
  - 13 focused Go files
  - Single responsibility per module
  - main.go reduced to 21 lines
  - Clear separation of concerns

### Technical
- Built with Go 1.24+
- Bubbletea TUI framework
- Lipgloss for styling
- Bubbles components
- Nerd Font icon support

---

## Project Milestones

### ✅ Milestone 1: Usable File Manager (Complete)
- Phase 1: Enhanced single-pane
- Phase 1.5: View modes (List, Grid, Detail, Tree)
- Phase 2: Dual-pane preview + editor integration
- **Achievement:** Fully functional file manager with preview and editing

### 🎯 Next: File Operations & Polish
- F7/F8 implementation (Create/Delete)
- Dialog system
- Error feedback system
- Search functionality

---

## Version History Summary

- **v0.3.0** - Context menu, favorites, TUI tool launcher
- **v0.2.0** - Dual-pane, preview, editor integration, F-keys, command prompt
- **v0.1.5** - View modes (List/Grid/Detail/Tree), mouse support
- **v0.1.0** - Initial release, core file browser, metadata display

---

**Project Started:** October 2025
**Current Status:** Feature-complete file viewer/browser, ready for file operation features
