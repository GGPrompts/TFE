# TFE Quick Reference for Project Manager Integration

## What is TFE?

TFE is a **terminal file manager** written in Go using the Bubbletea TUI framework. Think of it as "Midnight Commander meets modern terminal UI." It features:
- Dual-pane file browser with live preview
- Syntax-highlighted code preview
- Built-in app launcher (lazygit, lazydocker, htop, lnav)
- Keyboard shortcuts (F1-F12, vim-style navigation)
- Full mouse support including double-click and right-click menus
- Rich context menu system
- Favorites/bookmarks
- Search and filtering

## Why TFE for PM Integration?

1. **Built for App Launching** - Already launches external TUI apps (lazygit, etc.)
2. **Modular Architecture** - 14 focused Go files, each with single responsibility
3. **Extensible Event System** - Bubbletea message-passing makes adding features clean
4. **Mature Terminal Management** - Handles suspend/resume of external processes correctly
5. **Clear Integration Points** - Multiple ways to add PM features without breaking existing code

## Key Stats

| Metric | Value |
|--------|-------|
| Language | Go 1.24+ |
| Framework | Bubbletea + Lipgloss + Bubbles |
| Total Modules | 14 Go files |
| Main Model Size | ~260 fields (tracking all state) |
| Lines of Code | ~4,500 (excluding tests/docs) |
| View Modes | 3 (Single-pane, Dual-pane, Full preview) |
| Display Modes | 3 (List, Detail, Tree) |
| Keyboard Support | F1-F12, arrows, vim keys, Ctrl+X combos |
| Mouse Support | Left/right/double click, scroll, text selection |

## Integration Options (Ranked by Effort)

### Easiest (2-3 hours)
**Add Context Menu Action to File**
```
Right-click file → "📋 Create Task for this file"
→ Opens PM to create task linked to that file
```
- Edit: `context_menu.go` (add menu item)
- Edit: `context_menu.go` (handle action)
- Launch PM via: `openTUITool("pm", path)`

**Files to modify:** 1 file
**Risk level:** Very low

---

### Medium (4-6 hours)
**Add Keyboard Shortcut for PM Dashboard**
```
Press Ctrl+Shift+P → Shows full PM dashboard in TFE
Press again → Returns to file browser
```
- Edit: `types.go` (add viewMode)
- Edit: `update_keyboard.go` (add handler)
- Edit: `view.go` (add renderer)
- Create: `project_manager.go` (new module)

**Files to modify:** 4 files
**Risk level:** Low (doesn't touch existing logic)

---

### Complex (8-12 hours)
**PM as Side Panel (Like Dual-Pane Mode)**
```
Layout: [Files 40%] [Tasks 60%]
Same as dual-pane but shows projects instead of preview
```
- All medium-level changes PLUS:
- Edit: `model.go` (layout calculations)
- Edit: `styles.go` (add PM styling)
- Edit: `update_keyboard.go` (pane switching)

**Files to modify:** 6-7 files
**Risk level:** Medium (touches layout code)

---

## Architecture Patterns Used by TFE

### 1. Message-Driven Updates
```
Event (keyboard/mouse) 
  → Message type 
  → Update handler 
  → Returns (model, cmd) 
  → View re-renders
```

**For PM:** Just add new message types:
```go
type projectTaskSelectedMsg struct {
    taskID string
}

// Handle in Update()
case projectTaskSelectedMsg:
    m.selectedTask = msg.taskID
    return m, nil
```

### 2. Conditional Rendering by View Mode
```go
func (m model) View() string {
    switch m.viewMode {
    case viewSinglePane:
        return m.renderSinglePane()
    case viewDualPane:
        return m.renderDualPane()
    case viewFullPreview:
        return m.renderFullPreview()
    }
}
```

**For PM:** Add a new case:
```go
case viewProjectManager:
    return m.renderProjectManager()
```

### 3. TUI App Launching Pattern
```go
// Detect if app is available
if editorAvailable("lazygit") {
    // Add menu item
    items = append(items, menuItem{"Git", "lazygit"})
}

// Launch when selected
case "lazygit":
    return m, openTUITool("lazygit", m.currentPath)
```

**For PM:** Works the same for launching your PM:
```go
if editorAvailable("pm") {
    items = append(items, menuItem{"📋 Project Manager", "launch_pm"})
}

case "launch_pm":
    return m, openTUITool("pm", m.currentPath)
```

### 4. Dialog System
```go
// Open input dialog
m.showDialog = true
m.dialog.dialogType = dialogInput
m.dialog.title = "Enter text:"

// User types and presses Enter
// Result: m.dialog.input contains user text
```

**For PM:** Extend with new dialog types:
```go
const (
    dialogProjectSelect  // NEW - Select from projects
    dialogTaskCreate     // NEW - Create task
)
```

## File Navigation Guide

### Core Files (Understand these first)

1. **types.go** (173 lines)
   - Where: Line 32 - `type model struct` - main state
   - Where: Line 33 - `type viewMode int` - view modes
   - Where: Line 62 - `type fileItem struct` - file representation
   - What: All type definitions and enums

2. **view.go** (189 lines)
   - Where: Line 30 - `func (m model) View() string` - main renderer
   - Where: Line 60 - view mode dispatcher switch statement
   - What: How to render different views

3. **context_menu.go** (313 lines)
   - Where: Line 38 - `func getContextMenuItems()` - build menu
   - Where: Line 64 - Conditional tool integration (lazygit check)
   - Where: Line 139 - `func executeContextMenuAction()` - run action
   - What: How to add menu items and execute them

4. **update_keyboard.go** (714 lines)
   - Where: Line 23 - `func handleKeyEvent()` - main keyboard handler
   - Where: Line 75 - View-specific key handling pattern
   - What: How keyboard events work and flow

5. **editor.go** (90 lines)
   - Where: Line 40 - `func openTUITool()` - launch external app
   - Where: Line 41 - `c.Dir = dir` - sets working directory
   - What: How TUI tools are launched and managed

### Reference Files (Read for context)

6. **model.go** - Initialization
7. **command.go** - Command execution (similar pattern to app launching)
8. **update_mouse.go** - Mouse handling
9. **render_file_list.go** - File list rendering
10. **render_preview.go** - Preview rendering

## Quick Implementation Checklist

### Adding Context Menu Item (Easiest)

- [ ] **Step 1:** In `context_menu.go` line ~70, add PM menu item detection:
  ```go
  if editorAvailable("pm") {
      items = append(items, contextMenuItem{"📋 Project Manager", "launch_pm"})
  }
  ```

- [ ] **Step 2:** In `context_menu.go` line ~260, add handler:
  ```go
  case "launch_pm":
      return m, openTUITool("pm", m.currentPath)
  ```

- [ ] **Step 3:** Test: Install PM as `pm` command, right-click in TFE folder

---

### Adding Keyboard Shortcut (Medium)

- [ ] **Step 1:** In `types.go`, add view mode:
  ```go
  type viewMode int
  const (
      viewSinglePane
      viewDualPane
      viewFullPreview
      viewProjectManager  // NEW
  )
  ```

- [ ] **Step 2:** In `update_keyboard.go`, add handler for Ctrl+Shift+P:
  ```go
  case "ctrl+shift+p":
      m.viewMode = viewProjectManager
      return m, nil
  ```

- [ ] **Step 3:** In `view.go`, add renderer call:
  ```go
  case viewProjectManager:
      return m.renderProjectManager()
  ```

- [ ] **Step 4:** Create `project_manager.go` with:
  ```go
  func (m model) renderProjectManager() string {
      // Build your PM UI here
      return "Project Manager Panel"
  }
  ```

---

## Critical Insights

### 1. Don't Modify Main Message Dispatcher
The `Update()` function in `update.go` is the nervous system. Extensions go into handlers, not the dispatcher itself.

### 2. Use the Existing Style System
Lipgloss styles are in `styles.go`. Add new styles there, don't hardcode colors in renderers.

### 3. Follow the View Mode Pattern
Every feature should have a view mode. Makes it easier to toggle on/off.

### 4. Leverage the Dialog System
Instead of building custom input, use the existing dialog system (`dialog.go`). It already handles input, confirmation, messages.

### 5. Terminal State Management is Automatic
Don't worry about terminal cleanup. Bubbletea + TFE handle it. Just call `openTUITool()` or `openEditor()`.

---

## Integration with Other Tools

TFE Already Integrates With:
- **lazygit** - Git management
- **lazydocker** - Docker management  
- **htop** - Process monitoring
- **lnav** - Log viewing
- **viu/timg/chafa** - Image viewing
- **textual-paint** - Image editing
- **micro/nano/vim** - Text editing
- **xclip/pbcopy/clip.exe** - Clipboard

Your PM Can:
- Be launched from TFE context menu (like lazygit)
- Launch other tools itself (git, make, npm, etc.)
- Receive current directory as context
- Integrate with existing file browser workflow

---

## Test Drive

To understand TFE better:

```bash
cd ~/projects/TFE
go run .

# Try these:
1. Navigate folders with arrows
2. Press Space to see dual-pane with preview
3. Press F1 for help
4. Right-click (or F2) for context menu
5. Press F11 for prompts library
6. Press F12 for trash
7. Press Tab to switch between panes (in dual-pane mode)
8. Press Ctrl+P for fuzzy search
```

---

## Summary

TFE is **exceptionally well-designed** for integration. It's not just a file manager—it's a **platform for terminal-based development tools**. The fact that it already integrates lazygit, lazydocker, etc., proves the architecture works for external apps.

Your PM can integrate in multiple ways:
1. **Easiest:** Launch from context menu (2-3 hours)
2. **Better:** Toggle PM view mode (4-6 hours)
3. **Best:** Side panel with dual layout (8-12 hours)

All paths are well-supported by the existing architecture. Start with the easiest, iterate from there.

