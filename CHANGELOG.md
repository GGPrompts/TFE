# TFE Changelog

All notable changes to the Terminal File Explorer (TFE) project.

## [Unreleased]

### Added
- **Browser Support for Images and HTML Files**
  - F3 automatically opens images in default browser (PNG, JPG, GIF, SVG, WebP, BMP, ICO, TIFF)
  - F3 automatically opens HTML files in default browser
  - Context menu "🌐 Open in Browser" option for images and HTML files
  - Platform-aware detection (wslview, cmd.exe, xdg-open, open)
  - Falls back to text preview for non-browser files
  - Cross-platform support (WSL, Linux, macOS)

### Fixed
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
