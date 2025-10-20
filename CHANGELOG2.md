# TFE Changelog (Older Versions)

This file contains changelog entries for TFE versions 0.2.0 and earlier.
**See [CHANGELOG.md](CHANGELOG.md) for current versions (v0.3.0+)**

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

---

## Version History Summary

- **v0.2.0** - Dual-pane, preview, editor integration, F-keys, command prompt
- **v0.1.5** - View modes (List/Grid/Detail/Tree), mouse support
- **v0.1.0** - Initial release, core file browser, metadata display

---

**Project Started:** October 2025
**For current versions:** See [CHANGELOG.md](CHANGELOG.md)
