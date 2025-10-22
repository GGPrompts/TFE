# TFE Changelog (Version 0.4.0)

This file contains the changelog entry for TFE version 0.4.0.

**See [CHANGELOG.md](CHANGELOG.md) for current versions (v0.5.0+)**
**See [CHANGELOG2.md](CHANGELOG2.md) for older versions (v0.3.0 and earlier)**

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

**For current versions:** See [CHANGELOG.md](CHANGELOG.md)
**For older versions:** See [CHANGELOG2.md](CHANGELOG2.md)
