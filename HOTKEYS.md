# TFE Hotkeys Reference

## F-Keys (Midnight Commander Style)

| Key | Action |
|-----|--------|
| **F1** | Show this help reference |
| **F2** | Open context menu (keyboard alternative to right-click) |
| **F3** | Open images/HTML in browser OR view/preview file OR file picker (in input fields) |
| **F4** | Edit file in external editor |
| **F5** | Copy file path to clipboard (or rendered prompt in fillable fields mode) |
| **F6** | Toggle favorites filter (show only favorites) |
| **F7** | Create new directory (prompts for name) |
| **F8** | Delete file/folder (prompts for confirmation) |
| **F9** | Cycle through display modes (List → Grid → Detail → Tree) |
| **F10** | Quit TFE |

## Navigation

| Key | Action |
|-----|--------|
| **↑** / **k** | Move cursor up |
| **↓** / **j** | Move cursor down |
| **Enter** | Enter directory / Preview file (full-screen) / Toggle tree expansion |
| **Double-Click** | Enter directory / Preview file (full-screen) |
| **←** | Tree: collapse folder OR go to parent / Other modes: go to parent |
| **→** | Tree: expand folder OR enter directory / Other modes: enter directory |
| **h** | Go to parent directory (vim-style) |
| **l** | Enter directory (vim-style) |
| **Esc** | Clear command → Exit dual-pane → Go back a directory level |
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
| **Esc** | Exit full-screen preview / Exit dual-pane mode / Go back a level |
| **↑** / **k** | Scroll preview up (in full-screen or dual-pane right) |
| **↓** / **j** | Scroll preview down (in full-screen or dual-pane right) |
| **PgUp** | Page up in preview |
| **PgDn** | Page down in preview |
| **Mouse Wheel** | Scroll preview (in full-screen or focused right pane) |

## Prompt Templates & Fillable Fields

When viewing a prompt template with `{{VARIABLES}}` placeholders:

| Key | Action |
|-----|--------|
| **Tab** | Navigate to next input field |
| **Shift+Tab** | Navigate to previous input field |
| **Type** | Enter text into focused field |
| **Backspace** | Delete last character from field |
| **Ctrl+U** | Clear entire field |
| **Enter** | Move to next field (wraps around) |
| **F3** | Open file picker to select a file path |
| **F5** | Copy rendered prompt (with filled values) to clipboard |

### File Picker Mode (F3)
When selecting a file for a prompt variable:

| Key | Action |
|-----|--------|
| **↑/↓** or **k/j** | Navigate file list |
| **←/→** or **h/l** | Navigate directories |
| **Enter** (on directory) | Navigate into directory |
| **Enter** (on file) | Select file and return to prompt |
| **Double-Click** (on file) | Select file and return to prompt |
| **Esc** | Cancel file picker and return to prompt |

### Field Types
- **📁 File fields** (blue): For file paths - use F3 to pick files
- **📝 Short fields** (yellow): For single-line text input
- **📝 Long fields** (yellow): For multi-line text (shows truncated with char count)
- **🕐 Auto-filled** (green): Pre-filled with context (DATE, TIME) - editable

## File Operations

| Key | Action |
|-----|--------|
| **F4** | Edit file in external editor (Micro preferred, then nano/vim) |
| **n** / **N** | Edit file in nano specifically |
| **F5** | Copy file path to clipboard (or rendered prompt in F11 mode) |
| **F7** | Create new directory (prompts for name) |
| **F8** | Delete selected file/folder (prompts for confirmation) |

## Context Menu

| Key | Action |
|-----|--------|
| **F2** | Open context menu at cursor position |
| **Right-Click** | Open context menu at mouse position |
| **↑/↓** or **k/j** | Navigate menu items |
| **Enter** | Execute selected menu action |
| **Esc** / **q** | Close context menu |

Context menu actions include:
- 📂 Open / Quick CD (for directories)
- 👁️ Preview file
- 🌐 Open in browser (images/HTML files only)
- ✏️ Edit file
- ▶️ Run Script (executable files: .sh, .bash, .zsh, .fish or chmod +x)
- 📋 Copy path to clipboard
- 📁 New folder (for directories)
- 🗑️ Delete file/folder
- ⭐ Toggle favorite
- 🌿 Git (lazygit) - if available
- 🐋 Docker (lazydocker) - if available
- 📜 Logs (lnav) - if available
- 📊 Processes (htop) - if available

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
| **Esc** | Clear command prompt (then other ESC behaviors) |
| **↑** (with input) | Previous command in history |
| **↓** (with input) | Next command in history |
| **exit** / **quit** | Exit TFE (type and press Enter) |

> **Note:** The command prompt is MC (Midnight Commander) style - you can start typing anywhere without pressing a special key. Your input appears at the top of the screen. ANSI codes are automatically stripped from pasted text.

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
8. **Context Menu:** Press **F2** or right-click for quick access to common actions (including TUI tools like lazygit!)
9. **Favorites:** Use the context menu (F2/right-click) to bookmark files/folders, then **F6** to filter by favorites
10. **Tree Navigation:** In tree view (4), use ← to collapse, → to expand folders (Windows Explorer style)
11. **ESC to Go Back:** Press ESC to navigate back like Windows Explorer's back button
12. **Prompt Templates:** Press **F11** for prompts mode, open a template with `{{VARIABLES}}`, fill fields with Tab navigation, and F5 to copy the rendered result
13. **Run Scripts:** Right-click executable files (.sh, .bash, etc. or chmod +x) and select "▶️ Run Script" to execute them with output - press any key to return to TFE

---

**TFE Version:** Terminal File Explorer
**Built with:** Go + Bubbletea
**View this file:** Press **F1** from anywhere in TFE
