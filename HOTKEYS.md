# TFE Hotkeys Reference

## F-Keys (Midnight Commander Style)

| Key | Action |
|-----|--------|
| **F1** | Show this help reference |
| **F2** | Open context menu (keyboard alternative to right-click) |
| **F3** | Open images/HTML in browser OR view/preview file in full-screen |
| **F4** | Edit file in external editor |
| **F5** | Copy file path to clipboard |
| **F6** | Toggle favorites filter (show only favorites) |
| **F7** | Create directory *(placeholder for future)* |
| **F8** | Delete file/folder *(placeholder for future)* |
| **F9** | Cycle through display modes (List → Grid → Detail → Tree) |
| **F10** | Quit TFE |

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
| **F9** | Cycle through display modes (List → Grid → Detail → Tree) |
| **1** | Switch to List view |
| **2** | Switch to Grid view |
| **3** | Switch to Detail view |
| **4** | Switch to Tree view |
| **.** / **Ctrl+h** | Toggle hidden files visibility |

## Preview & Full-Screen Mode

| Key | Action |
|-----|--------|
| **F3** | Open images/HTML in default browser OR full-screen preview of current file |
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
| **F4** | Edit file in external editor (Micro preferred, then nano/vim) |
| **n** / **N** | Edit file in nano specifically |
| **F5** | Copy file path to clipboard |

## Context Menu

| Key | Action |
|-----|--------|
| **F2** | Open context menu at cursor position |
| **Right-Click** | Open context menu at mouse position |
| **↑/↓** or **k/j** | Navigate menu items |
| **Enter** | Execute selected menu action |
| **Esc** / **q** | Close context menu |

Context menu actions include:
- 👁️ Preview file
- 🌐 Open in browser (images/HTML files only)
- ✏️ Edit file
- 📋 Copy path to clipboard
- ⭐ Toggle favorite
- 📂 Quick CD (for directories)
- 🗑️ Delete file/folder

## Favorites

| Key | Action |
|-----|--------|
| **F6** | Toggle favorites filter (show only favorites) |
| **F2** or **Right-Click** | Open context menu to add/remove favorites |

To add or remove favorites, use the context menu (F2 or right-click) and select "☆ Add Favorite" or "⭐ Unfavorite".

When in favorites mode, press Enter on a favorite to navigate to its location.

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
| **F10** | Quit TFE |
| **Ctrl+C** | Force quit TFE |
| **exit** or **quit** | Exit TFE (type in command prompt + Enter) |

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
- Press **F4** to open in external editor

## Tips

1. **Quick Preview:** Press **Tab** to enter dual-pane mode and see file previews as you navigate
2. **Full-Screen Reading:** Press **Enter** or **F3** on a file for distraction-free viewing
3. **Browser Support:** Press **F3** on images (.png, .jpg, .gif, .svg, etc.) or HTML files to open them in your default browser
4. **Command Execution:** Type any shell command and press Enter - TFE pauses, runs it, and returns
5. **Fast Editing:** Press **F4** on any file to jump straight into Micro/nano editor
6. **Copy Paths:** Press **F5** to copy file paths for pasting elsewhere
7. **Command History:** Use ↑/↓ arrows when typing to recall previous commands
8. **Context Menu:** Press **F2** or right-click for quick access to common actions
9. **Favorites:** Press **s** to bookmark files/folders, then **F6** to filter by favorites

---

**TFE Version:** Terminal File Explorer
**Built with:** Go + Bubbletea
**View this file:** Press **F1** from anywhere in TFE
