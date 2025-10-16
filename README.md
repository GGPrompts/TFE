# TFE - Terminal File Explorer

A simple and clean terminal-based file explorer built with Go and Bubbletea. TFE provides a modern terminal UI with mouse and keyboard navigation, making file browsing efficient and intuitive.

## Features

- **Clean Interface**: Minimalist design focused on usability
- **Dual Navigation**: Both keyboard shortcuts and mouse support
- **F-Key Controls**: Midnight Commander-style F1-F10 hotkeys for common operations
- **Context Menu**: Right-click or F2 for quick access to file operations
- **Quick CD**: Exit TFE and change shell directory to selected folder
- **Dual-Pane Mode**: Split-screen layout with file browser and live preview
- **File Preview**: View file contents with syntax highlighting and line numbers
- **Text Selection**: Mouse text selection enabled in preview mode
- **Markdown Rendering**: Beautiful markdown preview with Glamour
- **External Editor Integration**: Open files in Micro, nano, vim, or vi
- **Command Prompt**: Midnight Commander-style always-active command line
- **Favorites System**: Bookmark files and folders with quick filter (F6)
- **Clipboard Integration**: Copy file paths to system clipboard
- **Multiple Display Modes**: List, Grid, Detail, and Tree views
- **Nerd Font Icons**: Visual file/folder indicators using file type detection
- **Smart Sorting**: Directories first, then files (alphabetically sorted)
- **Scrolling Support**: Handles large directories with auto-scrolling
- **Hidden File Filtering**: Automatically hides dotfiles for cleaner views
- **Double-Click Support**: Double-click to navigate folders or preview files

## Installation

### Prerequisites

- Go 1.24 or higher
- A terminal with Nerd Fonts installed (for proper icon display)
- **For Termux users**: Install `termux-api` for clipboard support: `pkg install termux-api`

### Building from Source

```bash
git clone https://github.com/GGPrompts/tfe.git
cd tfe
go build -o tfe
```

### Setup (Required for Quick CD Feature)

To enable the "Quick CD" feature that lets you exit TFE and automatically change your shell to a selected directory:

1. Add the wrapper to your shell configuration:

```bash
echo 'source ~/tfe/tfe-wrapper.sh' >> ~/.bashrc
```

2. Reload your shell:

```bash
source ~/.bashrc
```

3. Update the wrapper path if you installed TFE in a different location:

```bash
# Edit tfe-wrapper.sh and change the TFE_BIN path to match your installation
```

**Note**: After setup, use `tfe` command instead of `./tfe` to launch the application.

### Running

```bash
tfe    # If wrapper is installed (enables Quick CD)
# or
./tfe # Direct execution (Quick CD won't work)
```

The file explorer will start in your current working directory.

## Usage

### Keyboard Controls

#### F-Keys (Midnight Commander Style)
| Key | Action |
|-----|--------|
| `F1` | Show help (HOTKEYS.md reference) |
| `F2` | Open context menu for current file |
| `F3` | View/Preview file in full-screen |
| `F4` | Edit file in external editor |
| `F5` | Copy file path to clipboard |
| `F6` | Toggle favorites filter |
| `F7` | Create directory *(coming soon)* |
| `F8` | Delete file/folder *(coming soon)* |
| `F9` | Cycle through display modes |
| `F10` | Quit application |

#### Navigation
| Key | Action |
|-----|--------|
| `↑` / `k` | Move cursor up (or scroll preview when right pane focused) |
| `↓` / `j` | Move cursor down (or scroll preview when right pane focused) |
| `h` / `←` | Navigate to parent directory |
| `PageUp` | Scroll preview up one page (when right pane focused) |
| `PageDown` | Scroll preview down one page (when right pane focused) |
| `Enter` | Open folder or preview file |
| `Space` | Toggle dual-pane mode on/off |
| `Tab` | Toggle dual-pane mode / switch between panes |

#### View Modes
| Key | Action |
|-----|--------|
| `F9` | Cycle through display modes |
| `1` | Switch to List view |
| `2` | Switch to Grid view |
| `3` | Switch to Detail view |
| `4` | Switch to Tree view |
| `.` / `Ctrl+h` | Toggle hidden files |

#### Favorites
| Key | Action |
|-----|--------|
| `s` / `S` | Toggle favorite for current file/folder |
| `F6` | Toggle favorites filter (show only favorites) |

#### Other Keys
| Key | Action |
|-----|--------|
| `n` / `N` | Edit file in nano specifically |
| `Esc` | Exit dual-pane/preview mode / close context menu |
| `Ctrl+C` | Force quit application |

### Mouse Controls

- **Left Click**: Select item (or switch pane focus in dual-pane mode)
- **Double Click**: Navigate into folder or preview file
- **Right Click**: Open context menu for file operations (includes Quick CD for folders)
- **Scroll Wheel Up/Down**: Navigate through file list (or scroll context menu when open)
- **Text Selection**: Enabled in preview mode - select and copy text with mouse

### Context Menu Actions

Right-click (or press F2) on any file or folder to access:

**For Folders:**
- 📂 **Open** - Navigate into the directory
- 📂 **Quick CD** - Exit TFE and change shell to this directory (requires wrapper setup)
- 📋 **Copy Path** - Copy full path to clipboard
- ⭐/**☆ Favorite** - Add/remove from favorites

**For Files:**
- 👁 **Preview** - View file in full-screen preview
- ✏ **Edit** - Open in external editor (micro/nano/vim)
- 📋 **Copy Path** - Copy full path to clipboard
- ⭐/**☆ Favorite** - Add/remove from favorites

## Interface

TFE offers three distinct interface modes:

### Single-Pane Mode (Default)

```
┌─────────────────────────────────────────┐
│ TFE - Terminal File Explorer            │
│ /current/path/here                      │
│                                         │
│   ▸ folder1                             │
│   ▸ folder2                             │
│   • file1.txt                           │
│   [GO] file2.go                         │
│                                         │
│ ↑/↓: nav • Tab: dual-pane • q: quit    │
└─────────────────────────────────────────┘
```

### Dual-Pane Mode (Tab or Space)

```
┌────────────────────────────────────────────────────────────┐
│ TFE - Terminal File Explorer [Dual-Pane]                   │
│ /current/path/here                                         │
├───────────────────────┬────────────────────────────────────┤
│                       │ Preview: file2.go                  │
│   ▸ folder1           │ ────────────────────               │
│   ▸ folder2           │     1 │ package main              │
│   • file1.txt         │     2 │                           │
│ ► [GO] file2.go       │     3 │ import "fmt"              │
│                       │     4 │                           │
│                       │     5 │ func main() {             │
│                       │     6 │     fmt.Println("...")    │
│                       │                                    │
├───────────────────────┴────────────────────────────────────┤
│ [LEFT focused] • Tab: switch • Space: exit                 │
└────────────────────────────────────────────────────────────┘
```

### Full-Screen Preview Mode (F or Enter)

```
┌────────────────────────────────────────────────────────────┐
│ Preview: file2.go                                          │
│ Size: 1.2KB | Lines: 42 | Scroll: 1-20                    │
│                                                            │
│     1 │ package main                                       │
│     2 │                                                    │
│     3 │ import "fmt"                                       │
│     4 │                                                    │
│     5 │ func main() {                                      │
│    ... (full screen content)                               │
│                                                            │
│ ↑/↓: scroll • PgUp/PgDown: page • E: edit • Esc: close    │
└────────────────────────────────────────────────────────────┘
```

### Command Prompt (Always Visible)

The command prompt is always visible at the bottom of the screen. Simply start typing any command and it will automatically focus and capture your input. Press Enter to execute commands in the current directory context:

```
┌─────────────────────────────────────────┐
│ TFE - Terminal File Explorer            │
│ /current/path/here                      │
│                                         │
│   ▸ folder1                             │
│   ▸ folder2                             │
│   • file1.txt                           │
│                                         │
│ 3 folders, 12 files • List             │
│ /current/path/here $ ls -la█           │  ← Command prompt
└─────────────────────────────────────────┘
```

**Command Prompt Features:**
- Always visible at the bottom - no need to enter a special mode
- Start typing any character to automatically focus the prompt
- Execute any shell command in the current directory
- TFE suspends while the command runs, then resumes automatically
- File list refreshes automatically after command completes
- Command history with up/down arrows (stores last 100 commands)
- Press `Esc` to unfocus and clear the prompt
- Press `Backspace` to edit command text

**Example Commands:**
- `ls -la` - List files with details
- `touch newfile.txt` - Create a new file
- `mkdir testdir` - Create a new directory
- `git status` - Check git repository status
- `vim file.txt` - Open file in vim and return to TFE

#### Key Interface Elements

1. **Title Bar**: Application name and current mode
2. **Path Display**: Shows current directory path
3. **File List**: Scrollable list of folders and files with type indicators
4. **Preview Pane**: Live file preview with line numbers (dual-pane/full modes)
5. **Status Bar**: File counts, view mode, and selection info
6. **Command Prompt**: Always-visible shell command input at the bottom

## CellBlocksTUI Integration

TFE includes a built-in launcher for **CellBlocksTUI** - a companion terminal app for managing your command library, prompts, and code snippets. Access your card collection without leaving your terminal workflow!

### Quick Launch

Click the **[📦]** button in the TFE toolbar (top-left, after home/favorites/terminal buttons) to launch CellBlocksTUI instantly.

### Setting Up CellBlocksTUI

1. **Clone the repository:**
```bash
cd ~/projects
git clone https://github.com/GGPrompts/CellBlocksTUI.git
cd CellBlocksTUI
```

2. **Build the binary:**
```bash
go build -o cellblocks-tui
```

3. **Install to PATH:**
```bash
# Desktop/Linux
cp cellblocks-tui ~/bin/

# Termux
cp cellblocks-tui $PREFIX/bin/
```

4. **Verify installation:**
```bash
cellblocks-tui --help
# Or test the launch button from TFE
```

### What is CellBlocksTUI?

CellBlocksTUI is a lightweight terminal interface for browsing and copying cards from your CellBlocks library:

- **271 cards** organized in 10 categories (Bash, Prompts, Agents, etc.)
- **Grid and list views** with category-colored borders
- **Full-text search** across titles and content
- **Mouse/touch support** (click to select, double-click to copy)
- **Clipboard integration** (Termux, Linux, macOS, Windows)
- **5MB binary** with ~10MB RAM usage
- **Offline-first** - reads from local JSON file

Perfect for quick command lookups, copying prompts for AI chats, or browsing agent configurations.

### Split-Pane Workflow (Recommended)

For the ultimate terminal productivity setup, run TFE and CellBlocksTUI side-by-side in tmux:

```bash
# Create a split-pane workspace
tmux new -s work \; \
  split-window -v -p 30 \; \
  send-keys -t 0 'tfe' C-m \; \
  send-keys -t 1 'cellblocks-tui' C-m \; \
  select-pane -t 0
```

**Result:**
- **Top pane (70%)**: TFE - Browse and manage files
- **Bottom pane (30%)**: CellBlocksTUI - Quick access to commands and prompts

**Usage:**
- Browse files in TFE, copy commands from CellBlocksTUI
- Paste commands directly into TFE's command prompt
- Switch panes with `Ctrl+b` then arrow keys

### Data Compatibility

Both TFE and CellBlocksTUI can run simultaneously. CellBlocksTUI reads from:
```
~/projects/CellBlocks/data/cellblocks-data.json
```

This is the same data file used by the React CellBlocks app, so all your cards sync automatically.

For more info, see the [CellBlocksTUI README](https://github.com/GGPrompts/CellBlocksTUI).

## Technical Details

### Built With

- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components

### Project Structure

```
tfe/
├── main.go                 # Entry point (21 lines)
├── types.go                # Type definitions (135 lines)
├── styles.go               # Lipgloss styles (36 lines)
├── model.go                # Model initialization & layout (75 lines)
├── update.go               # Event handling (900+ lines)
├── view.go                 # View dispatcher (120 lines)
├── render_file_list.go     # File list rendering (440 lines)
├── render_preview.go       # Preview rendering (442 lines)
├── file_operations.go      # File operations & formatting (465 lines)
├── editor.go               # External editor & clipboard (76 lines)
├── command.go              # Command prompt execution (128 lines)
├── context_menu.go         # Context menu system (205 lines)
├── favorites.go            # Favorites/bookmarks (115 lines)
├── helpers.go              # Helper functions (45 lines)
├── tfe-wrapper.sh          # Shell wrapper for Quick CD
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
├── README.md               # User documentation
├── HOTKEYS.md              # Keyboard shortcuts reference
├── PLAN.md                 # Development roadmap
├── CLAUDE.md               # Architecture & development guide
└── tfe                     # Compiled binary (after build)
```

## Design Philosophy

TFE is designed to be simpler than full-featured file managers like Midnight Commander while maintaining modern terminal capabilities. The focus is on:

- **Simplicity**: Core navigation features without overwhelming options
- **Speed**: Fast startup and responsive navigation
- **Clean UI**: Minimal visual clutter with clear information hierarchy
- **Modern UX**: Mouse support and smooth scrolling for contemporary terminals
- **Modularity**: Well-organized codebase split across focused modules (see CLAUDE.md)

## Development

### Running in Development Mode

```bash
go run .
```

### Building

```bash
go build -o tfe
```

### Dependencies

Install dependencies manually if needed:

```bash
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles
```

### Architecture

TFE follows a modular architecture with 13 focused files:
- See **CLAUDE.md** for complete architecture documentation
- See **HOTKEYS.md** for complete keyboard shortcuts reference
- See **PLAN.md** for development roadmap and future features

## Roadmap

### Completed Features ✅
- ✅ File preview pane (dual-pane and full-screen modes)
- ✅ External editor integration
- ✅ File size and permissions display (Detail view)
- ✅ Multiple display modes (List, Grid, Detail, Tree)
- ✅ Clipboard integration (with Termux support)
- ✅ F-key hotkeys (Midnight Commander style)
- ✅ Context menu (right-click and F2)
- ✅ Quick CD feature (exit and change shell directory)
- ✅ Favorites/bookmarks system
- ✅ Text selection in preview mode
- ✅ Markdown rendering with Glamour
- ✅ Command history (last 100 commands)
- ✅ Bracketed paste support (proper paste handling)
- ✅ Special key filtering (no more literal "end", "home", etc.)

### Planned Features
- File operations (copy, move, delete, rename) - F7/F8 placeholders ready
- File search functionality
- Configurable color schemes and themes
- Custom hidden file patterns
- Syntax highlighting in code preview
- Archive file browsing (.zip, .tar.gz)
- Git status indicators

## License

MIT License - feel free to use and modify as needed.

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## Author

Created by GGPrompts

---

**Note**: This project requires a terminal with Nerd Fonts for proper icon display. Install from [nerdfonts.com](https://www.nerdfonts.com/) if icons don't display correctly.
