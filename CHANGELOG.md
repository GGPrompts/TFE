# TFE Changelog

All notable changes to the Terminal File Explorer (TFE) project.

## [Unreleased]

### Added
- **CLI File/Directory Arguments**
  - Open TFE to a specific directory: `tfe ~/projects`
  - Open with a file pre-selected: `tfe ~/projects/main.go`
  - Supports `~` expansion and validates paths exist
  - Shows helpful error messages for invalid paths
  - Files modified: `main.go`, `model.go`

- **--preview Flag for Auto-Opening Preview Pane**
  - New `--preview` or `-p` flag opens preview pane automatically
  - Preview pane is focused (60% width) when using this flag
  - File content is loaded immediately
  - Perfect for integration with other tools: `tmux split-window "tfe --preview file.go"`
  - Files modified: `main.go`, `model.go`

- **Cross-Platform Compatibility Improvements**
  - **setsid fallback**: Uses `nohup` on macOS/Termux where `setsid` isn't available (tmux quad feature)
  - **Dynamic Termux paths**: Uses PREFIX env var instead of hardcoded `/data/data/com.termux/...` paths
  - **wl-copy clipboard support**: Added Wayland clipboard support (checked before xclip/xsel)
  - **Better HOME detection**: Uses `os.UserHomeDir()` with fallback to HOME env var
  - **Improved Termux detection**: Added TERMUX_VERSION and TERMUX_APP_PID checks
  - **Better wslpath handling**: Validates converted Windows paths, helpful error messages
  - **Multi-drive WezTerm detection**: Checks C:, D:, E: drives for WezTerm in WSL
  - Files modified: `editor.go`, `helpers.go`, `model.go`, `terminal_graphics.go`

### Fixed
- **Terminal Resize Ghost Content**
  - Added `tea.ClearScreen` to WindowSizeMsg handler
  - Prevents duplicate footer text when resizing terminal
  - Files modified: `update.go`

- **Git Operations Display**
  - Fixed git operations (pull, push, sync, fetch) rendering underneath TFE
  - Now properly exits alt screen like other terminal operations
  - Uses `tea.ExecProcess` instead of raw `exec.Command().Run()`
  - Files modified: `git_operations.go`

### Changed
- **Trash Bin Behavior - Navigation Instead of Toggle**
  - F12 now navigates TO trash view (like visiting a special location) instead of toggling a filter
  - Your current path is saved and automatically restored when you exit trash
  - **Auto-exit trash**: Any navigation action automatically exits trash mode:
    - Navigating into directories (Enter, Right arrow, 'l')
    - Going to parent (Backspace, Left arrow)
    - Using cd command
    - Clicking navigation buttons (Home, Favorites, Prompts, Git Repos)
    - Using filter toggles (F6, F11)
  - **Manual exit**: Press F12 again to explicitly exit and restore previous location
  - Prevents confusion from accidentally staying in trash mode when navigating
  - Files modified: `types.go`, `helpers.go`, `update_keyboard.go`, `update_mouse.go`, `menu.go`

- **Automatic View Mode Selection Based on Terminal Width**
  - Narrow terminals (width < 100) now default to single-pane mode on startup
  - Wide terminals (width >= 100) default to dual-pane mode
  - Improves usability on mobile devices (Termux) and small terminal windows
  - Users can still toggle between modes using the pane button ([⬜]/[⬌])
  - Files modified: `model.go` (lines 32, 86-92)

### Fixed
- **Menu Bar Highlight Persistence**
  - Fixed menu highlight staying visible after using letter hotkeys (F/E/V/T/H) to open menus
  - Menu highlight now clears when navigating to files with up/down/j/k keys
  - Mouse clicks already cleared properly, now keyboard navigation does too
  - Files modified: `update_keyboard.go` (lines 1437-1441, 1483-1487)

- **Image Preview in WezTerm on WSL**
  - Fixed false positive: WezTerm in WSL no longer tries to use Kitty protocol (doesn't work on Windows)
  - Now shows fallback viewer options (V key) instead of blank "HD Preview via Kitty"
  - Added helper text footer to HD preview: "Press V to view in [viewer]" as fallback
  - Helps users in edge cases where terminal protocol detection might fail
  - Files modified: `terminal_graphics.go` (line 72), `file_operations.go` (lines 2097-2105)

- **Detail View Horizontal Scrolling on Narrow Terminals**
  - Fixed text corruption and misalignment when scrolling right in Detail view
  - Fixed column headers not staying aligned with data rows during scroll
  - Fixed text wrapping instead of truncating in single-pane mode
  - Added terminal-specific box width calculation for WezTerm/Termux vs Windows Terminal
  - Fixed infinite scrolling through empty space after last column
  - Fixed headers being scrolled twice (manual + automatic) causing wrong position
  - Fixed inconsistent width calculation between headers (runewidth.RuneWidth) and data rows (m.runeWidth)
  - Fixed missing "  " prefix on headers causing 2-column offset
  - Fixed headers using byte-based padding instead of visual-width-aware padding
  - Horizontal scrolling now works perfectly on narrow terminals (<100 cols) in both single and dual-pane modes
  - Files modified: `render_file_list.go` (lines 204-208, 276, 295, 336, 357, 362, 364-367, 378), `update_keyboard.go` (lines 1807-1815, 1816-1829)

### Added
- **HD Image Previews via Terminal Graphics Protocols**
  - Added `terminal_graphics.go` module for rendering images at full resolution
  - Supports Kitty protocol (WezTerm, Kitty terminal), iTerm2, and Sixel
  - Auto-detects terminal capabilities and uses best available protocol
  - Images render inline in preview pane (no more F5 exit → view → return)
  - Supports PNG, JPG, GIF, WebP formats with automatic scaling
  - Falls back to helpful message in unsupported terminals
  - Dependencies: `github.com/BourgeoisBear/rasterm`, `golang.org/x/image/webp`
  - Files added: `terminal_graphics.go`
  - Files modified: `file_operations.go`, `go.mod`, `go.sum`, `CLAUDE.md`

- **Blazing Fast Fuzzy Search (Ctrl+P)**
  - Replaced slow `go-fzf` library with external `fzf` + `fd`/`find`
  - Zero typing lag, instant results from entire directory tree
  - Auto-detects best file finder: `fd` > `fdfind` > `find`
  - Live file preview disabled by default (press `?` to toggle)
  - No file count limits, searches recursively without depth restrictions
  - Requires: `fzf` (external), optional: `fd`/`fdfind` for faster scans
  - Files modified: `fuzzy_search.go`, `CLAUDE.md`

- **Scroll Position Indicators**
  - Full preview mode (F5): Shows `Line 45/120 (37%)` in info bar
  - Dual-pane mode: Compact footer shows ` 45/120 (37%) ` at bottom
  - Works for all file types: markdown, text, code, prompts
  - Updates dynamically as you scroll with arrow keys or Page Up/Down
  - Files modified: `render_preview.go`

- **Markdown Files Visual Scrollbar**
  - Added scrollbar to markdown and prompt files (previously only in code files)
  - Appears on left edge with `│` track and `█` thumb indicator
  - Thumb size represents visible portion of file
  - Consistent with code file scrollbars (line number position)
  - Files modified: `render_preview.go`

- **Glamour Markdown Renderer Performance Optimization**
  - Caches Glamour renderer instance (like yazi does)
  - Uses fixed "dark" style instead of slow auto-detection
  - Eliminates terminal probing lag in WezTerm/Termux
  - First render has slight delay, all subsequent scrolls are instant
  - 10-50x faster markdown scrolling on slow terminals
  - Files modified: `file_operations.go`, `render_preview.go`, `types.go`

- **Update Notification System**
  - Auto-checks GitHub Releases API for new versions (once per 24 hours)
  - Shows `🎉 Update Available: vX.X.X (click for details)` in header during first 5 seconds
  - Clickable notification opens formatted changelog with update commands
  - Silent fail on network issues, respects GitHub rate limits
  - Cache stored in `~/.config/tfe/update_check.json`
  - Files modified: `types.go`, `update.go`, `view.go`, `render_preview.go`, `update_mouse.go`

- **Git Repos Toolbar Button**
  - Replaced 🎮 games button with 🔀 git repos toggle in toolbar (position 8)
  - Auto-switches to Detail view when enabled (shows git status columns)
  - Active state shows gray background (consistent with ⭐ and 📝 toggles)
  - Games launcher still accessible via Tools → Games Launcher menu
  - Toolbar layout: `[🏠] [⭐] [V] [⬌] [>_] [🔍] [📝] [🔀] [🗑️]`
  - Files modified: `view.go`, `render_preview.go`, `update_mouse.go`, `menu.go`

- **Prompt Variable Visual Feedback**
  - Variables in prompt header turn green when filled with custom text
  - Variables stay gray when empty/unfilled
  - Occurrence counts display: `variable_name (3)` shows duplicates
  - Real-time updates as you type, delete, or paste
  - Works in edit mode with live cache invalidation
  - Files modified: `prompt_parser.go`, `render_preview.go`, `update_keyboard.go`, `update_mouse.go`, `helpers.go`

- **Project-Specific Prompts Support**
  - Added `.prompts/` to `.gitignore` for personal prompts
  - `examples/.prompts/` remains tracked and shared with users
  - Created release preparation prompts in `.prompts/` (not tracked)
  - Files modified: `.gitignore`

- **Git Workspace Management System**
  - **Enhanced Git Repos Detail View**: New columns when viewing git repositories (🔀 filter)
    - Name (with relative path from scan root)
    - Branch (current branch name)
    - Status (⚡ Dirty, ↑3 Ahead, ↓2 Behind, ↑1↓2 Diverged, ✓ Clean)
    - Last Commit (relative time: "2 hours ago", "3 days ago")
    - Replaces old Size/Modified columns (not useful for repos)
  - **Git Operations via Context Menu**: Right-click any git repository to access:
    - ↓ Pull - Execute `git pull` with output feedback
    - ↑ Push - Execute `git push` with error handling
    - 🔄 Sync - Smart `git pull && git push` with step-by-step feedback
    - 🔍 Fetch - Update remote tracking branches without merging
  - **Auto-Refresh After Operations**: Status indicators update automatically after git commands
  - **Visual Triage Workflow**: See all repos with pending changes/unpushed commits at a glance
  - **Comprehensive Git Status Detection**:
    - `getGitStatus()` - Reads git refs to determine ahead/behind counts
    - `hasUncommittedChanges()` - Detects dirty working directory
    - `formatGitStatus()` - Emoji-based status display
    - `formatLastCommitTime()` - Human-readable relative time
  - New module: `git_operations.go` - Git command execution with feedback
  - Files modified: `file_operations.go`, `types.go`, `render_file_list.go`, `context_menu.go`, `update.go`

- **Parent Directory Name Display**
  - Parent directory (`..`) now shows actual folder name for better context
  - Format: `⬆️ .. (projects)` shows which folder you're navigating to
  - Handles edge cases: root directory shows "(root)", handles WSL paths correctly
  - Improves navigation context especially useful in deep folder hierarchies
  - Files modified: `render_file_list.go`

### Fixed
- **Terminal-Aware Emoji Width Compensation**
  - Fixed alignment issues with variation selector emojis (⬆️, ⚙️, 🗜️, 🖼️)
  - Different terminals render these emojis as different widths (WezTerm: 1 cell, Windows Terminal: 2 cells)
  - Added terminal detection system (`detectTerminalType()`) to identify WezTerm vs Windows Terminal
  - Width compensation now applied only where needed (Windows Terminal gets +1 per variation selector)
  - Perfect alignment achieved in both Windows Terminal and WezTerm at all window sizes
  - Files modified: `types.go`, `model.go`, `file_operations.go`, `render_file_list.go`

- **Git Dirty Status Detection**: Fixed false positives in uncommitted changes detection
  - Old heuristic (file mtime comparison) showed all repos as dirty
  - Now uses `git status --porcelain` for accurate detection
  - Only shows ⚡ Dirty when there are actual uncommitted changes
  - Files modified: `file_operations.go`

- **Wrapper Binary Detection**: Improved `tfe-wrapper.sh` to prevent recursion
  - Now checks specific paths first (`~/go/bin/tfe`, `~/.local/bin/tfe`) before using `command -v`
  - Prevents infinite loop when wrapper finds itself in PATH
  - Maintains cross-platform compatibility (macOS, Linux, WSL)
  - Files modified: `tfe-wrapper.sh`

## [1.0.0] - 2025-10-23

**First Public Release** 🎉

TFE is now ready for public use with comprehensive documentation, screenshots, and video demo.

### Documentation & Release Preparation
- Added 6 feature screenshots (main interface, light theme, dual-pane, tree view, context menu, prompts, fuzzy search)
- Embedded YouTube demo video showcasing all features
- Clarified Unicode emoji icons (Nerd Fonts not required)
- Cleaned up project structure (moved docs to `docs/`, archived old files)
- Removed unused code (landing_page.go)
- Updated installation instructions and prerequisites

### Added (2025-10-23)
- **Icon Differentiation for Prompt Templates**
  - Prompt files now show different icons based on whether they have fillable fields
  - 📝 (memo with pencil) = Editable template with `{{variables}}`
  - 📄 (plain document) = Plain prompt without fillable fields
  - **Performance-optimized with caching**: Variable check happens once on directory load (not every frame)
  - Cached in `fileItem.hasVariables` field to prevent lag with many prompt files
  - Only checks when viewing prompts (F11) to avoid overhead in regular file browsing
  - Helper functions: `hasPromptVariables()`, `isInPromptsDirectory()`
  - Files modified: `file_operations.go`, `types.go`

- **Smart File Opening (F4 Enhancement)**
  - CSV/TSV files: Opens in VisiData (interactive spreadsheet viewer)
  - Video files (mp4, mkv, avi, mov, webm): Opens in mpv media player
  - Audio files (mp3, wav, flac, ogg, m4a): Opens in mpv
  - PDF files: Opens in timg (terminal image viewer) or browser fallback
  - SQLite databases (.db, .sqlite, .sqlite3): Opens in harlequin
  - Archive files (.zip, .tar, .gz, .7z, .rar): Detection added
  - Binary files: Opens in hexyl (hex viewer) with helpful install instructions if missing
  - Graceful fallbacks when specialized viewers aren't installed
  - New file type detection functions: `isCSVFile()`, `isPDFFile()`, `isVideoFile()`, `isAudioFile()`, `isDatabaseFile()`, `isArchiveFile()`
  - New viewer functions: `getAvailableCSVViewer()`, `openCSVViewer()`, `getAvailableVideoPlayer()`, `openVideoPlayer()`, `getAvailableAudioPlayer()`, etc.
  - CSV preview: Shows helpful hints about VisiData with F4 shortcut
  - Files modified: `editor.go`, `file_operations.go`, `update_keyboard.go`, `HOTKEYS.md`

- **More AI Directory Exceptions for Hidden Files**
  - Added `.codex`, `.copilot`, `.devcontainer`, `.gemini`, and `.opencode` to always-visible list
  - These AI/IDE config folders now show even when "Show Hidden Files" is disabled
  - Joins existing exceptions: `.claude`, `.git`, `.vscode`, `.github`, `.config`, `.docker`, `.prompts`
  - Custom icons: 🤖 for AI folders (.codex, .copilot, .gemini, .opencode), 🐳 for .devcontainer (Docker-based)
  - Files modified: `file_operations.go`, `HOTKEYS.md`

- **One-Line Install Script (Like Midnight Commander)**
  - New automated install script: `curl -sSL https://raw.githubusercontent.com/GGPrompts/TFE/main/install.sh | bash`
  - Auto-installs TFE binary via `go install`
  - Downloads wrapper to `~/.config/tfe/tfe-wrapper.sh`
  - Auto-detects shell (bash/zsh) and configures wrapper
  - Enables Quick CD feature automatically (no manual setup needed)
  - Companion uninstall script for clean removal
  - Makes TFE installation feel "built-in" like MC
  - Files added: `install.sh`, `uninstall.sh`
  - Files modified: `README.md` (reorganized installation options)

- **New Prompt Template Creation**
  - Added **"📝 New Prompt..."** menu item in File menu
  - Creates timestamped `.prompty` files with proper YAML frontmatter template
  - Automatically opens in user's default editor (micro, nano, vim, or vi)
  - Template includes:
    - YAML frontmatter with `---` markers
    - Sample name, description, and input variable definitions
    - Structured sections (System Prompt, User Request, Instructions)
    - Example `{{variable}}` placeholders
  - Files modified: `menu.go`

- **Interactive File Picker for Copy Operations**
  - Replaced manual text input with visual file browser when copying files
  - Same UX as F3 file picker in prompts mode
  - Navigate folders with arrow keys, press Enter to select destination
  - Shows persistent helper text: "📁 Select destination for: filename (Enter = select folder, Esc = cancel)"
  - Header indicator: `[📋 Copy Mode - Select Destination]`
  - Files modified: `types.go`, `context_menu.go`, `update_keyboard.go`, `view.go`, `render_preview.go`

### Fixed (2025-10-23)
- **Prompt Template Tab Navigation Bugs**
  - Fixed Field 2 scroll position bug where Tab navigation scrolled to wrong location
  - Root cause: Search pattern matched first occurrence of variable name in text (e.g., "project" in "this project") instead of the actual `{{project}}` placeholder
  - Solution: Search for ANSI-styled variable (`\033[48;5;235m\033[38;5;220m{varName}\033[0m`) to match only the focused field
  - Fixed scroll calculation to use actual header height (not estimated) and match `renderPromptPreview()` logic exactly
  - Tab order now follows document order (first variable to last) instead of random order
  - Root cause: `extractVariables()` used Go map (unordered) to store variables
  - Solution: Preserve insertion order using separate slice while deduplicating with map
  - Updated file picker helper text: "Arrows/double-click to navigate, Enter to select file, Esc to cancel" (more explicit than "Navigate and press Enter")
  - Files modified: `helpers.go`, `prompt_parser.go`, `update_keyboard.go`

- **File Picker Status Message Persistence**
  - Status messages now persist during file picker navigation (no 3-second timeout)
  - Helper text stays visible until destination is selected or picker is cancelled
  - Applies to all view modes (single-pane, dual-pane, fullscreen)
  - Files modified: `view.go`, `render_preview.go`

- **Copy Path Logic**
  - Fixed destination path construction to include filename
  - Now properly creates `/destination/filename.ext` instead of failing
  - Handles both file and directory destinations correctly
  - Files modified: `update_keyboard.go`

- **Prompt Detection Overly Broad**
  - Regular `.md` files in `.claude/` directory no longer treated as prompts
  - Now only treats files as prompts if they have:
    - YAML frontmatter (starts with `---`) OR
    - `{{VARIABLES}}` in content
  - Documentation files like `CLAUDE.md` render as normal markdown
  - Files modified: `prompt_parser.go`

### Removed (2025-10-23)
- **Unused Slash Commands**
  - Deleted `.claude/commands/prompt-engineer.md` (redundant with global agent)
  - Deleted `.claude/commands/sync-projects.md` (project-specific, not generally useful)

### Fixed (2025-10-21)
- **Dual-Pane Accordion Layout**
  - Fixed accordion-style distribution (2/3 focused, 1/3 unfocused)
  - Improved width calculations for focused vs unfocused panes
  - Files modified: `model.go`, `render_preview.go`
- **Vertical Split for Detail View**
  - Added top/bottom pane split for narrow terminals
  - Automatic mode switching based on terminal width
  - Files modified: `model.go`, `render_preview.go`
- **Mouse Accuracy in All Modes**
  - Fixed click coordinate calculations across all view modes
  - Corrected header offsets for single-pane vs dual-pane
  - Files modified: `update_mouse.go`
- **Space Bar Command Mode**
  - Fixed space bar properly entering command mode when not in dual-pane
  - Files modified: `update_keyboard.go`

### Added (2025-10-20)
- **Context-Aware F1 Help Navigation**
  - F1 now intelligently jumps to the most relevant help section based on your current context
  - Smart detection for: input fields, context menu, command mode, preview modes, dual-pane, etc.
  - Greatly improves help discoverability - see exactly what you need, when you need it
  - Users can still manually scroll to other sections
  - Implementation: `helpers.go` (getHelpSectionName(), findSectionLine()), `update_keyboard.go`

- **Demo Recording Approach**
  - Created 10 VHS tape files but emojis/icons render as boxes (unusable for TFE)
  - Switched to OBS Studio recording to capture actual terminal appearance
  - OBS perfectly captures emojis, Nerd Font icons, colors, and cursor
  - See `docs/NEXT_SESSION.md` for OBS recording workflow

### Fixed (2025-10-20)
- **Browser Opening for Images/GIFs on WSL**
  - Fixed `cmd.exe /c start` treating first argument as window title instead of file path
  - Now uses correct syntax: `cmd.exe /c start "" <filepath>` (empty title parameter)
  - Images and HTML files now open correctly in default Windows browser from WSL2
  - Implementation: `editor.go:97-103`

---


---

**For older versions (v0.5.0 and earlier), see [CHANGELOG2.md](CHANGELOG2.md)**
