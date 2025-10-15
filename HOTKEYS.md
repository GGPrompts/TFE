# TFE Hotkeys Reference

## Navigation

| Key | Action |
|-----|--------|
| **↑** / **k** | Move cursor up |
| **↓** / **j** | Move cursor down |
| **Enter** | Enter directory / Preview file (full-screen) |
| **Double-Click** | Enter directory / Preview file (full-screen) |
| **h** / **←** | Go to parent directory |
| **Tab** | Toggle dual-pane mode / Switch focus (left ↔ right) |
| **Space** | Toggle dual-pane mode on/off |

## View Modes

| Key | Action |
|-----|--------|
| **v** | Cycle through display modes (List → Grid → Detail → Tree) |
| **1** | Switch to List view |
| **2** | Switch to Grid view |
| **3** | Switch to Detail view |
| **4** | Switch to Tree view |
| **.** / **Ctrl+h** | Toggle hidden files visibility |

## Preview & Full-Screen Mode

| Key | Action |
|-----|--------|
| **f** | Force full-screen preview of current file |
| **Enter** | Open full-screen preview (when on a file) |
| **Esc** | Exit full-screen preview / Exit dual-pane mode |
| **↑** / **k** | Scroll preview up (in full-screen or dual-pane right) |
| **↓** / **j** | Scroll preview down (in full-screen or dual-pane right) |
| **PgUp** | Page up in preview |
| **PgDn** | Page down in preview |
| **Mouse Wheel** | Scroll preview (in full-screen or focused right pane) |

## File Operations

| Key | Action |
|-----|--------|
| **e** / **E** | Edit file in external editor (Micro preferred, then nano/vim) |
| **n** / **N** | Edit file in nano specifically |
| **y** | Copy file path to clipboard (vim-style "yank") |
| **c** | Copy file path to clipboard |

## Command Prompt (MC-Style - Always Active)

| Key | Action |
|-----|--------|
| **Any letter/number** | Type into command prompt |
| **Backspace** | Delete last character from command |
| **Enter** | Execute command (or navigate if empty) |
| **Esc** | Clear command prompt |
| **↑** (with input) | Previous command in history |
| **↓** (with input) | Next command in history |
| **exit** / **quit** | Exit TFE (type and press Enter) |

> **Note:** The command prompt is MC (Midnight Commander) style - you can start typing anywhere without pressing a special key. Your input appears at the top of the screen.

## Dual-Pane Mode

| Key | Action |
|-----|--------|
| **Tab** | Switch focus between left pane (file list) and right pane (preview) |
| **Space** | Toggle dual-pane mode on/off |
| **↑/↓** or **k/j** | Navigate file list (left focus) or scroll preview (right focus) |
| **PgUp/PgDn** | Page up/down in preview (when right pane focused) |
| **Mouse Click** | Click on pane to switch focus |

## Quitting

| Key | Action |
|-----|--------|
| **q** | Quit TFE |
| **Ctrl+C** | Force quit TFE |
| **exit** or **quit** | Exit TFE (type in command prompt + Enter) |

## Help

| Key | Action |
|-----|--------|
| **?** | Show this hotkeys reference |

## File Type Indicators

TFE uses emoji icons to indicate file types:

- 📁 Folder
- 🐹 Go files (.go)
- 🐍 Python files (.py)
- 🟨 JavaScript (.js)
- 🔷 TypeScript (.ts)
- ⚛️  React (.jsx, .tsx)
- 📝 Markdown (.md)
- 📄 Text files (.txt)
- 🤖 Claude config files (CLAUDE.md, .claude/)
- ...and many more!

## Preview Features

### Markdown Files
- Beautiful rendering with **Glamour**
- Styled headers, lists, code blocks
- Syntax highlighting in code blocks
- Clickable hyperlinks (in supported terminals)
- No line numbers (cleaner reading)

### Text Files
- Line numbers shown
- Smart line wrapping at terminal width
- No horizontal scrolling
- Scrollbar indicator

### Binary/Large Files
- Detection and warning message
- Press **E** to open in external editor

## Tips

1. **Quick Preview:** Press **Tab** to enter dual-pane mode and see file previews as you navigate
2. **Full-Screen Reading:** Press **Enter** or **f** on a file for distraction-free viewing
3. **Command Execution:** Type any shell command and press Enter - TFE pauses, runs it, and returns
4. **Fast Editing:** Press **E** on any file to jump straight into Micro/nano editor
5. **Copy Paths:** Press **y** or **c** to copy file paths for pasting elsewhere
6. **Command History:** Use ↑/↓ arrows when typing to recall previous commands

---

**TFE Version:** Terminal File Explorer
**Built with:** Go + Bubbletea
**View this file:** Press **?** from anywhere in TFE
