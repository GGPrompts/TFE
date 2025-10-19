# TFE - Terminal File Explorer

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://go.dev/)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows%20%7C%20Termux-lightgrey)](https://github.com/GGPrompts/TFE)

A powerful and clean terminal-based file explorer built with Go and Bubbletea. TFE combines traditional file management with modern features like dual-pane preview, syntax highlighting, and an integrated AI prompts library. **Works beautifully on desktop and mobile (Termux) with full touch support.**

## Features

- **Clean Interface**: Minimalist design focused on usability
- **Dual Navigation**: Both keyboard shortcuts and mouse/touch support
- **Mobile Ready**: Full touch controls and optimized single-pane modes for Termux/Android
- **F-Key Controls**: Midnight Commander-style F1-F10 hotkeys for common operations
- **Fuzzy Search**: Fast file search with go-fzf (Ctrl+P or click 🔍)
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
- **Multiple Display Modes**: List, Detail, and Tree views
- **Nerd Font Icons**: Visual file/folder indicators using file type detection
- **Smart Sorting**: Directories first, then files (alphabetically sorted)
- **Scrolling Support**: Handles large directories with auto-scrolling
- **Hidden File Filtering**: Automatically hides dotfiles for cleaner views
- **Double-Click Support**: Double-click to navigate folders or preview files
- **Prompts Library**: F11 mode for AI prompt templates with fillable input fields, file picker (F3), and clipboard copy
- **Trash/Recycle Bin**: F12 to view deleted items, restore or permanently delete (F8 moves to trash)
- **Image Support**: View images with viu/timg/chafa and edit with textual-paint (MS Paint in terminal!)
- **File Operations**: Copy files/folders (📋 Copy to...), rename (✏️ Rename...) via context menu
- **Preview Search**: Ctrl-F to search within file previews, 'n' for next match, Shift-N for previous
- **Mouse Toggle**: Press 'm' in full preview to remove border for clean text selection

## Installation

### Prerequisites

- Go 1.24 or higher
- A terminal with Nerd Fonts installed (for proper icon display)
  - **Windows Terminal**: Requires manual font selection in Settings → Appearance → Font face (e.g., "CaskaydiaCove Nerd Font")
  - **Termux**: Works out of the box, no configuration needed
  - **macOS/Linux**: Depends on your terminal emulator (iTerm2, Alacritty, etc.)
- **For Termux users**: Install `termux-api` for clipboard support: `pkg install termux-api`

### Optional Dependencies

TFE works great without these, but install them for additional features:

**For Image Support:**
- **viu** (recommended) - View images in terminal with best quality
  ```bash
  # Install via cargo (Rust package manager)
  curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
  source $HOME/.cargo/env
  cargo install viu
  ```
  ```bash
  # Termux
  pkg install rust
  cargo install viu
  ```
- **timg** (alternative) - Another excellent image viewer
  ```bash
  # Ubuntu/Debian
  sudo apt install timg
  ```
- **chafa** (fallback) - ASCII art image viewer (works everywhere)
  ```bash
  # Ubuntu/Debian
  sudo apt install chafa

  # Termux
  pkg install chafa
  ```

**For Image Editing:**
- **textual-paint** - MS Paint in your terminal! (Python-based)
  ```bash
  # Ubuntu/Debian (requires system dependencies)
  sudo apt install -y python3-dev libjpeg-dev zlib1g-dev libtiff-dev libfreetype6-dev liblcms2-dev libwebp-dev
  pipx install textual-paint

  # Or via pip
  pip3 install --user textual-paint

  # Termux
  pkg install python-pillow
  pip install textual-paint
  ```

**Notes:**
- TFE checks for these tools and only shows image options if they're installed
- If none are installed, you can still open images in your browser (F3 or context menu)
- Priority order: viu > timg > chafa for viewing, textual-paint for editing

> 💡 **Need help installing?** Ask Claude or your AI assistant: *"Help me install TFE from https://github.com/GGPrompts/TFE on [your OS]"*

### Option 1: Quick Install (Recommended for Most Users)

**Install globally using Go:**

```bash
go install github.com/GGPrompts/TFE@latest
```

This installs the `tfe` binary to `~/go/bin/` (or `$GOPATH/bin`). Make sure this directory is in your PATH:

```bash
# Add to ~/.bashrc or ~/.zshrc if not already present
export PATH=$PATH:~/go/bin
```

**Usage:**
```bash
tfe    # Launch from any directory
```

✅ **What you get:**
- Global `tfe` command - launch from anywhere
- Clean installation via Go's package manager
- Easy updates with `go install`

❌ **What's missing:**
- Quick CD feature (see Option 2 if you want this)

---

### Option 2: Full Installation with Quick CD Feature

**For users who want the "Quick CD" feature** that lets you exit TFE and automatically change your shell to a selected directory:

1. **Clone and build:**

```bash
git clone https://github.com/GGPrompts/tfe.git
cd tfe
go build -o tfe
```

2. **Set up the shell wrapper:**

```bash
# Add wrapper to your shell config
echo 'source ~/tfe/tfe-wrapper.sh' >> ~/.bashrc

# For zsh users:
echo 'source ~/tfe/tfe-wrapper.sh' >> ~/.zshrc

# Reload your shell
source ~/.bashrc  # or source ~/.zshrc
```

3. **Update the path if you moved TFE** (the wrapper auto-detects by default):

The wrapper automatically finds the binary in the same directory. If you move it, edit `tfe-wrapper.sh` and adjust the path.

**Usage:**
```bash
tfe    # Launch from any directory with Quick CD support
```

✅ **What you get:**
- Global `tfe` command - launch from anywhere
- **Quick CD feature** - right-click folder → "📂 Quick CD" → exits TFE and changes your shell to that directory
- Direct access to source code for customization

---

### Which Option Should I Choose?

**Choose Option 1** if you:
- Just want a great terminal file manager
- Prefer standard Go package installation
- Don't need the Quick CD shell integration

**Choose Option 2** if you:
- Want the Quick CD feature (exit TFE and auto-cd to selected folder)
- Like having the source code locally
- Want to customize or contribute to TFE

**Note:** You can always start with Option 1 and add the wrapper later if you decide you want Quick CD!

## Usage

### Keyboard Controls

#### F-Keys (Midnight Commander Style)
| Key | Action |
|-----|--------|
| `F1` | Show help (HOTKEYS.md reference) |
| `F2` | Open context menu for current file |
| `F3` | View/Preview file in full-screen |
| `F4` | Edit file in external editor |
| `F5` | Copy file path to clipboard (or prompt in F11 mode) |
| `F6` | Toggle favorites filter |
| `F7` | Create directory |
| `F8` | Delete file/folder |
| `F9` | Cycle through display modes |
| `F10` | Quit application |
| `F11` | Toggle Prompts Library mode |
| `F12` | Toggle Trash/Recycle Bin view |

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
| `2` | Switch to Detail view |
| `3` | Switch to Tree view |
| `.` / `Ctrl+h` | Toggle hidden files |

#### Favorites
| Key | Action |
|-----|--------|
| `s` / `S` | Toggle favorite for current file/folder |
| `F6` | Toggle favorites filter (show only favorites) |

#### Other Keys
| Key | Action |
|-----|--------|
| `Ctrl+P` | Launch fuzzy file search |
| `Ctrl+F` | Search within file preview (n: next, Shift-N: previous, Esc: exit) |
| `m` / `M` | Toggle mouse & border in full preview mode (for clean text selection) |
| `n` / `N` | Edit file in nano specifically |
| `Esc` | Exit dual-pane/preview mode / close context menu |
| `Ctrl+C` | Force quit application |

### Mouse Controls

- **Toolbar Buttons**: Click [🏠] home, [⭐] favorites, [>_] command mode, [🔍] fuzzy search
- **Left Click**: Select item (or switch pane focus in dual-pane mode)
- **Double Click**: Navigate into folder or preview file
- **Right Click**: Open context menu for file operations (includes Quick CD for folders)
- **Scroll Wheel Up/Down**: Navigate through file list (or scroll context menu when open)
- **Text Selection**: Enabled in preview mode - select and copy text with mouse
- **Column Headers** (Detail view): Click to sort by Name, Size, Modified, or Type

### Mobile & Termux Support

TFE has been **extensively tested on Termux/Android** throughout development and works beautifully with touch controls:

**Touch Controls:**
- **Tap**: Select file/folder (same as left click)
- **Double Tap**: Navigate into folder or preview file
- **Long Press**: Open context menu (same as right click)
- **Swipe Up/Down**: Scroll through file list
- **Pinch/Spread**: Not needed - use keyboard for view switching

**Optimized for Mobile:**
- **Single-pane modes**: List, Detail, and Tree views all work excellently on small screens
- **Toolbar buttons**: Large touch targets for easy tapping
- **Context menu**: Touch-friendly menu system
- **Full preview mode**: Distraction-free reading on mobile
- **F-key access**: Use on-screen keyboard or external keyboard (many Termux keyboards have F-keys)

**Termux Installation:**
```bash
# Install required packages
pkg install golang-1.21 git termux-api

# Clone and build
git clone https://github.com/GGPrompts/tfe.git
cd tfe
go build -o tfe

# Run
./tfe
```

**Tips for Mobile:**
- Use **List** or **Detail** view for best readability on small screens
- **Tree view** works great for hierarchical navigation
- Access **context menu** with long press instead of F2
- **Prompts library** (F11) is perfect for managing AI prompts on mobile
- Install `termux-api` package for clipboard support

### Context Menu Actions

Right-click (or press F2) on any file or folder to access:

**For Folders:**
- 📂 **Open** - Navigate into the directory
- 📂 **Quick CD** - Exit TFE and change shell to this directory (requires wrapper setup)
- 📁 **New Folder...** - Create a new subdirectory
- 📄 **New File...** - Create a new file (auto-opens in editor)
- 📋 **Copy Path** - Copy full path to clipboard
- 📋 **Copy to...** - Copy directory recursively to destination
- ✏️ **Rename...** - Rename directory
- ⭐/**☆ Favorite** - Add/remove from favorites

**For Files:**
- 👁 **Preview** - View file in full-screen preview
- ✏ **Edit** - Open in external editor (micro/nano/vim)
- 📋 **Copy Path** - Copy full path to clipboard
- 📋 **Copy to...** - Copy file to destination
- ✏️ **Rename...** - Rename file
- ⭐/**☆ Favorite** - Add/remove from favorites

**For Images (PNG, JPG, GIF, SVG, etc.):**
- 🖼️ **View Image** - Display in terminal (requires viu, timg, or chafa)
- 🎨 **Edit Image** - Edit in terminal paint program (requires textual-paint)
- 🌐 **Open in Browser** - Open with default browser

**For HTML Files:**
- 🌐 **Open in Browser** - Open with default browser

### Prompts Library (F11)

TFE includes a built-in **Prompts Library** system for managing AI prompt templates across multiple locations. Press **F11** to enter Prompts Mode and access your prompt collection.

**Key Features:**
- **Multi-location support**: Access prompts from `~/.prompts/` (global), `.claude/commands/`, `.claude/agents/`, and local project folders
- **Global prompts section**: Quick access to `~/.prompts/` from any directory (shown at top of file list)
- **Template parsing**: Supports `.prompty` (Microsoft Prompty format), `.yaml`, `.md`, and `.txt` files
- **Variable substitution**: Auto-fills `{{file}}`, `{{filename}}`, `{{project}}`, `{{path}}`, `{{DATE}}`, `{{TIME}}` from current context
- **Fillable Fields**: Interactive input fields for custom `{{VARIABLES}}` with smart type detection
  - **Tab/Shift+Tab** navigation between fields
  - **File fields** (📁 blue): Use **F3** to browse and select files
  - **Auto-filled fields** (🕐 green): Pre-populated with context (editable)
  - **Text fields** (📝 yellow): Short or long text input with paste support
  - **F5** copies fully rendered prompt with all filled values
- **Clipboard copy**: Press **Enter** or **F5** to copy rendered prompt to clipboard
- **Smart filtering**: Only shows prompt files and folders containing prompts
- **Preview rendering**: View prompts with metadata (name, description, source, variables)

**Supported Formats:**

1. **Microsoft Prompty** (`.prompty`) - YAML frontmatter between `---` markers:
```prompty
---
name: Code Review
description: Review code changes
---
Please review {{file}} for best practices and potential issues.
```

2. **YAML** (`.yaml`, `.yml`) - Simple YAML with `template` field (only in `.claude/` or `~/.prompts/`):
```yaml
name: Bug Fix
description: Create a bug fix
template: |
  Fix the bug in {{file}}.
  Project: {{project}}
```

3. **Markdown/Text** (`.md`, `.txt`) - Plain text with `{{variables}}` (only in `.claude/` or `~/.prompts/`):
```markdown
Analyze {{file}} and suggest improvements for the {{project}} project.
```

**Setting Up Your Prompts Library:**

To use the global prompts feature, create a `~/.prompts/` directory in your home folder:

```bash
# Create the global prompts directory
mkdir -p ~/.prompts

# Create an example prompt (Microsoft Prompty format)
cat > ~/.prompts/code-review.prompty << 'EOF'
---
name: Code Review
description: Review code for best practices and potential issues
---
Please review the following code for:
- Best practices and code quality
- Potential bugs or edge cases
- Performance considerations
- Security vulnerabilities

File: {{file}}
Project: {{project}}

Provide specific, actionable feedback.
EOF
```

Now when you press **F11** in TFE, you'll see `🌐 ~/.prompts/ (Global Prompts)` at the top of the file list, accessible from any directory.

**Tips:**
- Organize prompts into subdirectories: `~/.prompts/coding/`, `~/.prompts/writing/`, etc.
- Use `.prompty` format for the best metadata support
- Prompts are accessible from **any directory** when F11 mode is active
- The `~/.prompts/` folder auto-expands when you enable F11 mode

**Quick Start:**
1. Press **F11** to enable Prompts Mode
2. Navigate to `🌐 ~/.prompts/ (Global Prompts)` or `.claude/` folders
3. Select a prompt file to preview it with auto-filled variables
4. **If the prompt has `{{VARIABLES}}`:**
   - Use **Tab/Shift+Tab** to navigate between input fields
   - Type to fill in text fields, or press **F3** to pick files
   - Press **Ctrl+U** to clear a field
5. Press **F5** to copy the fully rendered prompt to clipboard
6. Paste into your AI chat (Claude, ChatGPT, etc.)

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

The command prompt is always visible at the top of the screen (3rd row, below the toolbar). Simply start typing any command and it will automatically focus and capture your input. Press Enter to execute commands in the current directory context:

```
┌─────────────────────────────────────────┐
│ TFE - Terminal File Explorer            │  ← Title bar
│ 🏠 ⭐ 📝 🗑️                              │  ← Toolbar (clickable buttons)
│ $ ls -la█                               │  ← Command prompt (3rd row)
│                                         │
│   ▸ folder1                             │  ← File list
│   ▸ folder2                             │
│   • file1.txt                           │
│                                         │
│ /current/path • 3 folders, 12 files    │  ← Status bar
└─────────────────────────────────────────┘
```

**Command Prompt Features:**
- Always visible at the top (3rd row) - no need to enter a special mode
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
2. **Toolbar**: Clickable emoji buttons (🏠 Home, ⭐ Favorites, 📝 Prompts, 🗑️ Trash)
3. **Command Prompt**: Always-visible shell command input (3rd row)
4. **File List**: Scrollable list of folders and files with type indicators
5. **Preview Pane**: Live file preview with line numbers (dual-pane/full modes)
6. **Status Bar**: Current path, file counts, view mode, and selection info

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
- ✅ Multiple display modes (List, Detail, Tree)
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
- ✅ Fuzzy file search with go-fzf (Ctrl+P or click 🔍)
- ✅ Clickable toolbar buttons (home, favorites, search, etc.)
- ✅ Column header sorting in Detail view (click to sort)
- ✅ Rounded borders and polished UI
- ✅ Syntax highlighting for code files (Chroma)
- ✅ Prompts Library with template parsing and variable substitution (F11)
- ✅ File operations: Create (F7), Delete to Trash (F8), Copy, Rename
- ✅ Trash/Recycle Bin (F12) - restore or permanently delete items
- ✅ Image viewing (viu/timg/chafa) and editing (textual-paint)
- ✅ Preview search (Ctrl-F) with match navigation (n/Shift-N)
- ✅ Mouse toggle in preview ('m' key) for clean text selection

### Planned Features (v1.1+)
- Configurable color schemes and themes
- Custom hidden file patterns
- Archive file browsing (.zip, .tar.gz)
- Git status indicators
- Multi-select and bulk operations
- Context Visualizer - show Claude Code context and token counts

## License

MIT License - feel free to use and modify as needed.

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## Author

Created by GGPrompts

---

**Note**: This project requires a terminal with Nerd Fonts for proper icon display. Install from [nerdfonts.com](https://www.nerdfonts.com/) if icons don't display correctly.
