# TFE Hotkeys & Commands

Quick reference for Terminal File Explorer keyboard shortcuts and workflows.

## 🎯 Navigation

### Basic Movement
```
↑ / k - Move up
↓ / j - Move down
← / h - Go to parent directory
→ / l - Enter directory / Open file
Enter - Enter directory / Open file
```

### Quick Jumps
```
g - Jump to top of list
G - Jump to bottom of list
Home - Jump to top (alternative)
End - Jump to bottom (alternative)
Page Up - Scroll up one page
Page Down - Scroll down one page
```

### Directory Navigation
```
Backspace - Go to parent directory
~ - Go to home directory
/ - Go to root directory
. - Refresh current directory
```

## 🔍 Search & Filter

### Search
```
Ctrl+P - Fuzzy search (fzf + fd/find)
/ - Search/filter current directory
n - Next search result
N - Previous search result
Esc - Clear search filter
```

### Quick Access
```
Ctrl+H - Toggle hidden files
. - Show/hide hidden files (alternative)
```

## 📋 File Operations

### Copy & Move
```
y - Yank (copy) file path to clipboard
Y - Yank absolute path
c - Copy file/folder (opens file picker)
m - Move file/folder
```

### Create & Delete
```
n - New file/folder
d - Delete file/folder (confirmation required)
r - Rename file/folder
```

### Trash & Recycle
```
F8 - Move to trash
F12 - Navigate to trash directory
Delete - Move to trash (alternative)
```

## 📁 View Modes

### Display Modes
```
1 - List view (compact)
2 - Detail view (with metadata)
3 - Tree view (hierarchical)
```

### Panel Modes
```
D - Toggle dual-pane mode
p - Toggle preview pane
Tab - Switch between panes (in dual mode)
```

## 🎨 Preview & Editing

### Preview Controls
```
Space - Toggle preview for selected file
p - Toggle preview pane
Ctrl+F - Search within preview
n - Next search result in preview
N - Previous search result in preview
```

### Preview Scrolling
```
↑ / k - Scroll preview up
↓ / j - Scroll preview down
Page Up - Scroll preview page up
Page Down - Scroll preview page down
m - Toggle mouse mode (for text selection)
```

### Edit Files
```
e - Edit in Micro editor
E - Edit in $EDITOR (vim/nano/etc)
F4 - Edit file (Midnight Commander style)
```

## 🖱️ Mouse Controls

### Click Actions
```
Single click - Select file/folder
Double click - Enter folder / Open file
Right click - Context menu (F2)
```

### Scroll Actions
```
Mouse wheel - Scroll file list
Shift+wheel - Horizontal scroll (if needed)
```

### Touch Support (Mobile/Termux)
```
Tap - Select
Double tap - Enter/Open
Long press - Context menu
Swipe - Scroll
```

## 🎯 Function Keys (F-keys)

### Primary Functions
```
F1 - Help (context-aware)
F2 - Context menu
F3 - File picker (in prompts mode)
F4 - Edit file
F5 - Refresh directory
F6 - Toggle favorites filter
F7 - Create new folder
F8 - Move to trash
F9 - Menu
F10 - Quit
F11 - Prompts library
F12 - Navigate to trash
```

## 📚 Prompts Library (F11)

### Navigate Prompts
```
F11 - Open prompts library
↑↓ / jk - Navigate prompts
Enter - Use selected prompt
Esc - Exit prompts mode
```

### Prompt Actions
```
Tab - Next input field
Shift+Tab - Previous input field
Ctrl+C - Copy prompt to clipboard
F3 - Open file picker (for file/folder fields)
n - Create new prompt
e - Edit prompt template
d - Delete prompt
```

### File Picker (F3 in Prompts)
```
F3 - Open file picker for current field
↑↓ - Navigate files
Enter - Select file/folder
Esc - Cancel picker
```

## ⭐ Favorites

### Manage Favorites
```
F6 - Toggle favorites filter
f - Add current file/folder to favorites
u - Remove from favorites (unfavorite)
```

### Navigate Favorites
```
F6 - Show only favorites
↑↓ - Navigate favorites list
F6 - Show all files (toggle off)
```

## 🐙 Git Operations (in Git repos)

### Git Status View
```
g s - Show git status
g d - Git diff
g l - Git log
```

### Git Actions (Context Menu)
```
Right-click → Git →
  - Pull
  - Push
  - Sync (pull + push)
  - Fetch
  - Status
```

### Git Indicators
```
⚡ - Dirty (uncommitted changes)
↑ - Ahead of remote
↓ - Behind remote
✓ - Clean (synced)
```

## 🎮 Games Integration (Optional)

### Launch Games (if TUIClassics installed)
```
Tools menu →
  - Snake
  - Minesweeper
  - Solitaire
  - 2048
```

## 🖼️ Image Viewing

### Preview Images
```
Space - Preview image (inline or via external tool)
i - View with viu (colored blocks)
I - View with timg (24-bit color)
```

### Image Tools
```
Right-click image →
  - View (viu/timg/chafa)
  - Edit (textual-paint - MS Paint in terminal!)
```

## 📝 Common Workflows

### Quick File Navigation
```bash
Ctrl+P           # Fuzzy search
# Type filename
Enter            # Open file
```

### Copy Files to Another Location
```bash
c                # Start copy
# Navigate to destination
Enter            # Complete copy
```

### Create & Edit New File
```bash
n                # New file
# Type filename
e                # Edit in Micro
```

### Use AI Prompt Template
```bash
F11              # Open prompts
↓↓               # Navigate to prompt
Tab              # Fill input fields
F3               # Pick file (if needed)
Ctrl+C           # Copy to clipboard
# Paste in Claude Code or other AI
```

### Quick CD (Exit to Folder)
```bash
Right-click folder
# Select "Quick CD"
# TFE exits, shell is now in that folder!
```

### Git Workspace Triage
```bash
# Navigate to folder with git repos
# Status indicators show repo state
Right-click repo with ⚡
# Select "Git → Sync"
# Repo auto-pulls and pushes
```

### Trash Management
```bash
F8               # Move file to trash
F12              # Navigate to trash
↓                # Select item
Right-click
# "Restore" or "Permanently Delete"
```

## 🎨 Context Menu (F2)

### File Context Menu
```
Open
Edit
Copy
Move
Rename
Delete
Move to Trash
---
Copy Path
Copy Absolute Path
---
Add to Favorites
Quick CD
---
Git (if in repo)
Tools (if available)
```

### Folder Context Menu
```
Enter
Open in New Pane
---
Launch Claude Code
Launch TUI Tools (auto-detected)
---
Copy Path
Add to Favorites
Quick CD
---
Git (if repo)
```

## ⚙️ Configuration

### Theme Toggle
```
t - Toggle theme (light/dark)
```

### View Settings
```
1/2/3 - View modes
D - Dual pane
p - Preview pane
```

## 🛠️ Troubleshooting

### Refresh Display
```
F5 - Refresh current directory
Ctrl+L - Redraw screen
```

### Exit Modes
```
Esc - Exit current mode
q - Quit TFE
Ctrl+C - Force quit
F10 - Quit (alternative)
```

## 🚀 Power User Combos

### Fast File Finding
```bash
Ctrl+P           # Fuzzy search
# Type: "readme"
Enter            # Open README.md
```

### Clipboard Workflow
```bash
y                # Yank path
# In another terminal:
cd $(pbpaste)    # Jump to that path
```

### Multi-Pane File Management
```bash
D                # Enable dual pane
c                # Copy from left
Tab              # Switch to right
# Navigate to destination
Enter            # Paste
```

### Git Triage Workflow
```bash
# In folder with multiple repos
g s              # Show git status
↓↓               # Navigate to dirty repo
Right-click      # Context menu
# Git → Sync
# Repeat for other repos
```

### Prompt-to-AI Workflow
```bash
F11              # Prompts library
↓                # Select "Analyze code"
Tab Tab          # Fill fields
F3               # Pick file
Ctrl+C           # Copy prompt
# Alt+Tab to terminal
# Paste in Claude Code
```

## 📊 Visual Indicators

### File Type Icons
```
📁 - Folder
📄 - File
🧠 - Obsidian vault
📦 - Package/Archive
🖼️ - Image
📝 - Text/Code
🔧 - Config
```

### Git Status Icons
```
⚡ - Dirty
↑ - Ahead
↓ - Behind
✓ - Clean
```

## ⌨️ Quick Reference Card

```
Navigation:     hjkl       Vim-style
                ←↑↓→       Arrow keys
                g/G        Top/bottom
                Ctrl+P     Fuzzy search

Actions:        Enter      Open
                e          Edit
                c/m        Copy/Move
                y          Yank path

Views:          1/2/3      List/Detail/Tree
                D          Dual pane
                p          Preview

Functions:      F1         Help
                F2         Context menu
                F11        Prompts
                F12        Trash

Quick:          Space      Preview
                .          Refresh
                q          Quit
```

---

**Version**: TFE v1.0+
**Last Updated**: 2024-11-02
**Platform**: Linux | macOS | WSL | Termux
