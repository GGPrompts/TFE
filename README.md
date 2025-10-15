# TFE - Terminal File Explorer

A simple and clean terminal-based file explorer built with Go and Bubbletea. TFE provides a modern terminal UI with mouse and keyboard navigation, making file browsing efficient and intuitive.

## Features

- **Clean Interface**: Minimalist design focused on usability
- **Dual Navigation**: Both keyboard shortcuts and mouse support
- **Dual-Pane Mode**: Split-screen layout with file browser and live preview
- **File Preview**: View file contents with syntax highlighting and line numbers
- **External Editor Integration**: Open files in Micro, nano, vim, or vi
- **Command Prompt**: Execute shell commands directly from TFE
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

### Building from Source

```bash
git clone https://github.com/GGPrompts/tfe.git
cd tfe
go build -o tfe
```

### Running

```bash
./tfe
```

The file explorer will start in your current working directory.

## Usage

### Keyboard Controls

#### Navigation
| Key | Action |
|-----|--------|
| `↑` / `k` | Move cursor up (or scroll preview when right pane focused) |
| `↓` / `j` | Move cursor down (or scroll preview when right pane focused) |
| `h` / `←` | Navigate to parent directory |
| `PageUp` | Scroll preview up one page (when right pane focused) |
| `PageDown` | Scroll preview down one page (when right pane focused) |

#### File Operations
| Key | Action |
|-----|--------|
| `Enter` | Open folder or preview file |
| `Space` | Toggle dual-pane mode on/off |
| `Tab` | Toggle dual-pane mode / switch between panes |
| `f` | Full-screen preview of selected file |
| `E` | Edit file in external editor (Micro preferred) |
| `N` | Edit file in nano |
| `y` / `c` | Copy file path to clipboard |

#### View Modes
| Key | Action |
|-----|--------|
| `v` | Cycle through display modes |
| `1` | Switch to List view |
| `2` | Switch to Grid view |
| `3` | Switch to Detail view |
| `4` | Switch to Tree view |
| `.` / `Ctrl+h` | Toggle hidden files |

#### Exit
| Key | Action |
|-----|--------|
| `q` / `Ctrl+C` | Quit application |
| `Esc` | Exit dual-pane/preview mode (or quit from single-pane) |

### Mouse Controls

- **Left Click**: Select and open item (or switch pane focus in dual-pane mode)
- **Double Click**: Navigate into folder or preview file
- **Scroll Wheel Up/Down**: Navigate through file list

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

## Technical Details

### Built With

- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components

### Project Structure

```
tfe/
├── main.go                 # Entry point (21 lines)
├── types.go                # Type definitions (116 lines)
├── styles.go               # Lipgloss styles (36 lines)
├── model.go                # Model initialization & layout (64 lines)
├── update.go               # Event handling (520 lines)
├── view.go                 # View dispatcher (120 lines)
├── render_file_list.go     # File list rendering (284 lines)
├── render_preview.go       # Preview rendering (266 lines)
├── file_operations.go      # File operations & formatting (329 lines)
├── editor.go               # External editor integration (72 lines)
├── command.go              # Command prompt & execution (96 lines)
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
├── README.md               # User documentation
├── PLAN.md                 # Development roadmap
├── CLAUDE.md               # Architecture & development guide
├── docs/                   # Additional documentation
│   ├── REFACTOR_PLAN.md    # Refactoring history
│   └── RESEARCH.md         # Background research
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

TFE follows a modular architecture with 11 focused files:
- See **CLAUDE.md** for complete architecture documentation
- See **docs/REFACTOR_PLAN.md** for refactoring history
- See **PLAN.md** for development roadmap and future features

## Roadmap

### Completed Features ✅
- ✅ File preview pane (dual-pane and full-screen modes)
- ✅ External editor integration
- ✅ File size and permissions display (Detail view)
- ✅ Multiple display modes (List, Grid, Detail, Tree)
- ✅ Clipboard integration

### Planned Features
- File operations (copy, move, delete, rename)
- File search functionality
- Configurable color schemes and themes
- Bookmarks/favorites system
- Custom hidden file patterns
- Syntax highlighting in preview
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
