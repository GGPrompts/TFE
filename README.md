# TFE - Terminal File Explorer

A simple and clean terminal-based file explorer built with Go and Bubbletea. TFE provides a modern terminal UI with mouse and keyboard navigation, making file browsing efficient and intuitive.

## Features

- **Clean Interface**: Minimalist design focused on usability
- **Dual Navigation**: Both keyboard shortcuts and mouse support
- **Nerd Font Icons**: Visual file/folder indicators using Nerd Fonts
- **Smart Sorting**: Directories first, then files (alphabetically sorted)
- **Scrolling Support**: Handles large directories with auto-scrolling
- **Hidden File Filtering**: Automatically hides dotfiles for cleaner views

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

| Key | Action |
|-----|--------|
| `↑` / `k` | Move cursor up |
| `↓` / `j` | Move cursor down |
| `Enter` / `Space` | Open selected folder |
| `h` / `←` | Navigate to parent directory |
| `q` / `Esc` / `Ctrl+C` | Quit application |

### Mouse Controls

- **Left Click**: Select and open item
- **Scroll Wheel Up/Down**: Navigate through list

## Interface

The TFE interface consists of three main sections:

```
┌─────────────────────────────────────────┐
│ TFE - Terminal File Explorer            │
│ /current/path/here                      │
│                                         │
│   📁 folder1                            │
│   📁 folder2                            │
│   📄 file1.txt                          │
│   📄 file2.go                           │
│                                         │
│ ↑/↓: navigate • enter: open • h: back  │
└─────────────────────────────────────────┘
```

1. **Title Bar**: Application name
2. **Path Display**: Shows current directory path
3. **File List**: Scrollable list of folders and files
4. **Help Bar**: Quick reference for keyboard shortcuts

## Technical Details

### Built With

- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components

### Project Structure

```
tfe/
├── main.go       # Main application code
├── go.mod        # Go module definition
├── go.sum        # Dependency checksums
├── README.md     # This file
└── tfe           # Compiled binary (after build)
```

## Design Philosophy

TFE is designed to be simpler than full-featured file managers like Midnight Commander while maintaining modern terminal capabilities. The focus is on:

- **Simplicity**: Core navigation features without overwhelming options
- **Speed**: Fast startup and responsive navigation
- **Clean UI**: Minimal visual clutter with clear information hierarchy
- **Modern UX**: Mouse support and smooth scrolling for contemporary terminals

## Development

### Running in Development Mode

```bash
go run main.go
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

## Roadmap

Potential future features:

- File operations (copy, move, delete)
- File search functionality
- Configurable color schemes
- File preview pane
- Bookmarks/favorites
- Custom hidden file patterns
- File size and permissions display

## License

MIT License - feel free to use and modify as needed.

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## Author

Created by GGPrompts

---

**Note**: This project requires a terminal with Nerd Fonts for proper icon display. Install from [nerdfonts.com](https://www.nerdfonts.com/) if icons don't display correctly.
