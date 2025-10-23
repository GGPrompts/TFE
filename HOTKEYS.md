# TFE Hotkeys Reference

## F-Keys (Midnight Commander Style)

| Key | Action |
|-----|--------|
| **F1** | Show this help reference |
| **F2** | Open context menu (keyboard alternative to right-click) |
| **F3** | Open images/HTML in browser OR view/preview file OR file picker (in input fields) |
| **F4** | Edit file in external editor |
| **F5** | Copy file path to clipboard (or rendered prompt in prompts mode) |
| **F6** | Toggle favorites filter (show only favorites) |
| **F7** | Create new directory (prompts for name) |
| **F8** | Delete file/folder (moves to trash - use F12 to view/restore) |
| **F9** / **Alt** | Enter menu bar navigation mode (keyboard access to File/Edit/View/Tools/Help menus) |
| **F10** | Quit TFE |
| **F11** | Toggle prompts filter (show only .yaml, .md, .txt files + ~/.prompts & ~/.claude) |
| **F12** | Toggle trash/recycle bin view (restore/permanently delete items) |

## Navigation

| Key | Action |
|-----|--------|
| **↑** / **k** | Move cursor up |
| **↓** / **j** | Move cursor down |
| **Enter** | Enter directory / Preview file (full-screen) / Toggle tree expansion |
| **Double-Click** | Enter directory / Preview file (full-screen) |
| **←** | Go to parent directory (in all modes) / In Tree view: collapse expanded folder |
| **→** | Enter directory (in all modes) / In Tree view: expand collapsed folder |
| **h** | Go to parent directory (vim-style) |
| **l** | Enter directory (vim-style) |
| **Esc** | Clear command → Exit dual-pane → Go back a directory level |
| **Tab** | Toggle dual-pane mode / Switch focus (left ↔ right) |
| **Space** | Toggle dual-pane mode on/off |

**Tree View Navigation (when in tree mode - press 3):**
- Use **↑/↓** or **k/j** to move between files and folders
- Use **→** (right arrow) to expand a collapsed folder - shows its contents
- Use **←** (left arrow) to collapse an expanded folder - hides its contents
- Use **Enter** to toggle folder expansion (expand if collapsed, collapse if expanded)
- Use **Ctrl+W** to collapse all expanded folders at once (reset tree view)

## View Modes

| Key | Action |
|-----|--------|
| **1** | Switch to List view |
| **2** | Switch to Detail view |
| **3** | Switch to Tree view |
| **.** / **Ctrl+h** | Toggle hidden files visibility |

## Menu Bar Navigation

| Key | Action |
|-----|--------|
| **Alt** / **F9** | Enter menu bar navigation mode (highlights first menu) |
| **←** / **Shift+Tab** | Navigate to previous menu (when in menu bar mode) |
| **→** / **Tab** | Navigate to next menu (when in menu bar mode) |
| **↓** / **Enter** | Open highlighted menu dropdown |
| **↑/↓** | Navigate menu items (when dropdown is open) |
| **←/→** | Switch to adjacent menu (when dropdown is open) |
| **Enter** | Execute selected menu item |
| **Esc** | Close dropdown (returns to menu bar) / Exit menu mode |

**How it works:**
1. Press **Alt** or **F9** to enter menu mode - the first menu (File) will be highlighted
2. Use **←/→** or **Tab/Shift+Tab** to navigate between menus (File → Edit → View → Tools → Help)
3. Press **↓** or **Enter** to open a dropdown
4. Use **↑/↓** to select menu items, **Enter** to execute
5. Press **←/→** while in a dropdown to smoothly switch to adjacent menus
6. Press **Esc** to close dropdown (stays in menu mode) or press **Esc** again to exit menu mode

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
| **m** / **M** | Toggle mouse & border (FULL PREVIEW ONLY - removes border for clean text selection) |
| **Ctrl+F** | Search within file preview |
| **n** | Next search match (when searching) |
| **Shift+N** | Previous search match (when searching) |
| **Mouse Wheel** | Scroll preview (when mouse is enabled) |

**Note:** To copy text from files, the best method is to press **F4** to open the file in Micro editor, where you can select and copy text normally. The **m** key (mouse toggle) works in full-screen preview mode only - when you press **m**, the decorative border disappears and mouse is disabled, giving you clean terminal text selection. Press **m** again to restore the border and mouse scrolling.

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
- 🖼️ View Image (images only - viu/timg/chafa)
- 🎨 Edit Image (images only - textual-paint/durdraw)
- 🌐 Open in browser (images/HTML files only)
- ✏️ Edit file
- ▶️ Run Script (executable files: .sh, .bash, .zsh, .fish or chmod +x)
- 📋 Copy path to clipboard
- 📋 Copy to... (copy files/folders)
- ✏️ Rename... (rename files/folders)
- 📁 New folder (for directories)
- 📄 New file (for directories)
- 🗑️ Delete file/folder
- ⭐ Toggle favorite
- 🌿 Git (lazygit) - if available
- 🐋 Docker (lazydocker) - if available
- 📜 Logs (lnav) - if available
- 📊 Processes (htop) - if available

**Image files** get special menu options:
- **🖼️ View Image** - Opens in terminal image viewer (requires viu, timg, or chafa)
- **🎨 Edit Image** - Opens in terminal paint program (requires textual-paint or durdraw)
- Works best in Kitty, iTerm2, or WezTerm terminals (fallback to ASCII art in others)

## Favorites

| Key | Action |
|-----|--------|
| **F6** | Toggle favorites filter (show only favorites) |
| **F2** or **Right-Click** | Open context menu to add/remove favorites |

To add or remove favorites, use the context menu (F2 or right-click) and select "☆ Add Favorite" or "⭐ Unfavorite".

When in favorites mode, press Enter on a favorite to navigate to its location.

## Prompts Mode (F11)

| Key | Action |
|-----|--------|
| **F11** | Toggle prompts filter on/off |

When prompts filter is active:
- Shows only `.yaml`, `.md`, and `.txt` files (prompt templates)
- Auto-displays **🌐 ~/.prompts/** folder at the top (global prompts library)
- Auto-displays **🌐 ~/.claude/** folder (slash commands, agents, skills)
- Shows local `.claude/` and `.prompts/` folders if they exist
- Folders containing prompt files are always shown
- Navigate to virtual folders (🌐 ~/.prompts/) to browse global prompts

**Fillable Fields:**
When viewing a prompt with `{{VARIABLES}}`:
- Input fields appear automatically
- Press **Tab** to navigate between fields
- Press **F3** in a file field to open file picker
- Press **F5** to copy rendered prompt to clipboard
- See "Prompt Templates & Fillable Fields" section above for full details

## Trash/Recycle Bin (F12)

| Key | Action |
|-----|--------|
| **F12** | Toggle trash view on/off |
| **F8** (in normal mode) | Move file/folder to trash (safe deletion) |

When in trash view:
- Shows all deleted items with deletion timestamps
- Right-click or press **F2** for trash context menu:
  - ♻️ **Restore** - Move item back to original location
  - 🗑️ **Delete Permanently** - Cannot be undone!
  - 🧹 **Empty Trash** - Permanently delete all items in trash

**Trash location:** `~/.config/tfe/trash/`

**Safety features:**
- F8 moves to trash instead of permanent deletion
- Original paths are tracked for restoration
- Trash can be browsed like a normal directory
- Empty trash requires confirmation

## Command Prompt (Vim-Style)

### Basic Commands

| Key | Action |
|-----|--------|
| **:** | Enter command mode (focus command prompt) |
| **Esc** | Exit command mode and clear prompt |
| **Enter** | Execute command (or navigate if empty) |
| **!** prefix | Run command and exit TFE (e.g., `:!claude --yolo`) |
| **exit** / **quit** | Exit TFE (type and press Enter) |

### Cursor Movement

| Key | Action |
|-----|--------|
| **←** / **→** | Move cursor left/right in command text |
| **Home** / **Ctrl+A** | Jump to beginning of line |
| **End** / **Ctrl+E** | Jump to end of line |
| **Ctrl+Left** / **Alt+Left** / **Alt+B** | Jump one word left |
| **Ctrl+Right** / **Alt+Right** / **Alt+F** | Jump one word right |

### Editing

| Key | Action |
|-----|--------|
| **Backspace** | Delete character before cursor |
| **Delete** | Delete character at cursor (forward delete) |
| **Ctrl+K** | Delete from cursor to end of line |
| **Ctrl+U** | Delete from cursor to beginning of line |
| **Type** | Insert text at cursor position |

### History Navigation

| Key | Action |
|-----|--------|
| **↑** (in command mode) | Previous command in history |
| **↓** (in command mode) | Next command in history |
| **Mouse Wheel** | Scroll through command history (when command focused) |

### Features

- **Persistent History**: Command history saved to `~/.config/tfe/command_history.json` and survives restarts
- **Cursor Editing**: Full cursor control - edit anywhere in the command line
- **Word Jumping**: Fast navigation with Ctrl/Alt + arrows
- **Visual Feedback**: Cursor `█` shows current position, `!` prefix appears in red
- **Smart Navigation**: Arrow keys and mouse wheel navigate history when command is focused, don't affect file tree

> **Note:** Press **:** (colon) to enter command mode - your input appears at the top of the screen with a cursor. Use arrow keys to move the cursor and edit anywhere in the text. Press **Esc** to exit command mode. Command history persists across TFE restarts.

## Dual-Pane Mode

| Key | Action |
|-----|--------|
| **Tab** | Switch focus between left pane (file list) and right pane (preview) |
| **Space** | Toggle dual-pane mode on/off |
| **↑/↓** or **k/j** | Navigate file list (left focus) or scroll preview (right focus) |
| **PgUp/PgDn** | Page up/down in preview (when right pane focused) |
| **Mouse Click** | Click on pane to switch focus |

## Background Processes & Shell Access

| Key | Action |
|-----|--------|
| **Ctrl+Z** | Suspend TFE and drop to shell (type `fg` to resume) |

When you run scripts that start background processes (like servers, tmux sessions, etc.), you can:
1. Press **Ctrl+Z** to suspend TFE
2. Check on background processes, view logs, run commands
3. Type `fg` to resume TFE exactly where you left off

## Quitting

| Key | Action |
|-----|--------|
| **F10** | Quit TFE |
| **Ctrl+C** | Force quit TFE |
| **Ctrl+Z** | Suspend TFE (drop to shell - type `fg` to resume) |
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
4. **Command Execution:** Press **:** to enter command mode, type any shell command, and press Enter - TFE pauses, runs it, and returns
5. **Fast Editing:** Press **F4** on any file to jump straight into Micro/nano editor
6. **Copy Paths:** Press **F5** to copy file paths for pasting elsewhere
7. **Command History:** Press **:** to enter command mode, then use ↑/↓ arrows to recall previous commands
8. **Context Menu:** Press **F2** or right-click for quick access to common actions (including TUI tools like lazygit!)
9. **Favorites:** Use the context menu (F2/right-click) to bookmark files/folders, then **F6** to filter by favorites
10. **Tree Navigation:** In tree view (press 3), use **→** to expand folders, **←** to collapse, **↑/↓** to navigate (Windows Explorer style)
11. **ESC to Go Back:** Press ESC to navigate back like Windows Explorer's back button
12. **Prompt Templates:** Press **F11** for prompts mode, open a template with `{{VARIABLES}}`, fill fields with Tab navigation, and F5 to copy the rendered result
13. **Run Scripts:** Right-click executable files (.sh, .bash, etc. or chmod +x) and select "▶️ Run Script" to execute them with output - press any key to return to TFE
14. **Background Processes:** Run a script that starts servers/background processes, press **Ctrl+Z** to suspend TFE and check on them, then `fg` to resume
15. **Safe Deletion:** Press **F8** to move files to trash (not permanent!), press **F12** to view trash and restore or permanently delete
16. **Global Prompts:** Press **F11** to see your ~/.prompts and ~/.claude folders from anywhere - perfect for AI-assisted development
17. **Command Mode:** Press **:** to focus the command line (see gray hint text), type any shell command, press Enter to execute
18. **Copying Text from Files:** Press **F4** to open in Micro editor - this is the easiest way to select and copy text. In full-screen preview (F3/Enter), you can also press **m** to remove the border and disable mouse, enabling clean terminal text selection. The border disappears as visual feedback
19. **Search in Preview:** Press **Ctrl+F** while viewing a file to search, type your query, press **n** for next match, **Shift+N** for previous, **Esc** to exit search
20. **Viewing Images:** Right-click on image files (.png, .jpg, .gif, etc.) and select "🖼️ View Image" to see them in your terminal! Requires viu, timg, or chafa. For editing, select "🎨 Edit Image" to use textual-paint (MS Paint in terminal!)
21. **Hidden Files:** Press **.** (period) or **Ctrl+H** to toggle hidden files. Note: Important AI/development folders (.claude, .codex, .copilot, .devcontainer, .gemini, .opencode, .git, .vscode, .github, .config, .docker, .prompts), secrets files (.env, credentials, etc.), ignore files (.gitignore, .dockerignore, etc.), and all symlinks are always shown for security and project awareness
22. **Open in File Explorer:** Press **Ctrl+O** to open the current directory in your system file explorer (Windows Explorer in WSL, Finder on macOS, or default file manager on Linux)

---

**TFE Version:** Terminal File Explorer
**Built with:** Go + Bubbletea
**View this file:** Press **F1** from anywhere in TFE
