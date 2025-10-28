# TFE (Terminal File Explorer) - Comprehensive Architecture & Integration Analysis

## Executive Summary

TFE is a sophisticated terminal-based file manager built with Go and Bubbletea (a Charm Bracelet TUI framework). It's designed as a **modern alternative to Midnight Commander** with support for dual-pane browsing, file preview, external tool integration, and extensive keyboard/mouse controls. Critically for your use case, TFE has **built-in hooks for launching external TUI applications** from within its context menu.

**Key Finding:** TFE is architected with clear extension points that make it ideal for integration with a project manager TUI.

---

## Architecture Overview

### 1. Modular Design Pattern

TFE follows a **single-responsibility principle** across 14 focused Go modules:

```
core/
├── main.go              (21 lines)  - Minimal entry point
├── types.go             (173 lines) - Type definitions
├── model.go             (78 lines)  - Model initialization
├── styles.go            (35 lines)  - Lipgloss styling

event-handling/
├── update.go            (111 lines) - Message dispatcher
├── update_keyboard.go   (714 lines) - Keyboard events
├── update_mouse.go      (383 lines) - Mouse events

ui-rendering/
├── view.go              (189 lines) - View dispatcher
├── render_file_list.go  (447 lines) - File list rendering
├── render_preview.go    (468 lines) - Preview rendering

features/
├── file_operations.go   (657 lines) - File management
├── context_menu.go      (313 lines) - Right-click menu
├── editor.go            (90 lines)  - Editor/app launching
├── command.go           (127 lines) - Command execution
├── favorites.go         (150 lines) - Bookmarking system
├── dialog.go            (141 lines) - UI dialogs
└── helpers.go           (69 lines)  - Utilities
```

**Why This Matters:** Each module has a clear purpose, making it easy to add features without monolithic file growth.

---

## Key Features (Highly Relevant for PM Integration)

### 1. TUI Application Launcher (Context Menu)

**File:** `context_menu.go` + `editor.go`

TFE has **built-in support for launching TUI applications** from the context menu. Currently integrated:

- **lazygit** - Git management (if installed)
- **lazydocker** - Docker management (if installed)
- **lnav** - Log file viewer (if installed)
- **htop** - Process monitor (if installed)
- **viu/timg/chafa** - Image viewers
- **textual-paint** - Image editor
- **External editors** - micro, nano, vim, vi

**How It Works:**

```go
// From context_menu.go (lines 64-90)
if editorAvailable("lazygit") {
    items = append(items, contextMenuItem{"🌿 Git (lazygit)", "lazygit"})
}

// Execution (lines 253-256)
case "lazygit":
    return m, openTUITool("lazygit", m.contextMenuFile.path)

// From editor.go (lines 40-50)
func openTUITool(tool, dir string) tea.Cmd {
    c := exec.Command(tool)
    c.Dir = dir  // Set working directory
    return tea.Sequence(
        tea.ClearScreen,
        tea.ExecProcess(c, func(err error) tea.Msg {
            return editorFinishedMsg{err}
        }),
    )
}
```

**Key Points:**
- TUI tools launch with proper directory context (`c.Dir = dir`)
- Terminal is cleared before launch
- UI resumes after tool exits
- Error handling via messages

### 2. Application Lifecycle Management

**Pattern:** Suspend → Execute → Resume

TFE uses Bubbletea's `tea.ExecProcess()` to:
1. Save terminal state
2. Clear screen
3. Launch external process with stdin/stdout/stderr passthrough
4. Restore terminal state after process completes
5. Refresh file list automatically

**This is Critical:** A project manager could use this same pattern to launch and manage child processes.

### 3. Command Execution System

**File:** `command.go`

Features:
- Always-visible command prompt (Midnight Commander style)
- Arbitrary shell command execution
- Command history (last 100 commands)
- Output display with "Press any key" pause
- Auto-refresh of file list after execution

```go
// Lines 24-60 show the pattern:
func runCommand(command, dir string) tea.Cmd {
    return func() tea.Msg {
        script := fmt.Sprintf(`
echo "$ %s"
cd %s || exit 1
%s
echo ""
echo "Press any key to continue..."
read -n 1 -s -r
`, shellQuote(command), shellQuote(dir), command)
        
        c := exec.Command("bash", "-c", script)
        c.Stdin = os.Stdin
        c.Stdout = os.Stdout
        c.Stderr = os.Stderr
        
        return tea.Sequence(
            tea.ClearScreen,
            tea.ExecProcess(c, func(err error) tea.Msg {
                return commandFinishedMsg{err: err}
            }),
        )()
    }
}
```

---

## Integration Architecture

### 1. View Modes

TFE supports multiple view modes (in `types.go`):

```go
type viewMode int

const (
    viewSinglePane   // File list only
    viewDualPane     // File list + preview
    viewFullPreview  // Full-screen preview
)

type displayMode int

const (
    modeList   // Simple list view
    modeDetail // Detailed with columns
    modeTree   // Hierarchical tree view
)
```

**For PM Integration:** Could add a new view mode like:
```go
viewProjectManager // Full PM dashboard mode
```

### 2. Event Flow

**Pattern:** Messages → Update → View

```
Keyboard/Mouse Event
    ↓
handleKeyEvent() / handleMouseEvent()
    ↓
Returns (model, cmd)
    ↓
View renders model.viewMode
```

To add a new TUI panel:
1. Add new view mode to `types.go`
2. Handle input in `update_keyboard.go`
3. Add rendering to `view.go`

### 3. State Management

The `model` struct (lines 189-254 in `types.go`) tracks:

```go
type model struct {
    // Navigation
    currentPath  string
    files        []fileItem
    cursor       int
    
    // View state
    viewMode     viewMode
    displayMode  displayMode
    
    // UI state
    height       int
    width        int
    focusedPane  paneType
    
    // Context
    contextMenuOpen  bool
    contextMenuFile  *fileItem
    
    // Features
    favorites        map[string]bool
    showFavoritesOnly bool
    promptInputFields []promptInputField
    // ... more fields
}
```

**For PM Integration:** You could extend this with:
```go
type model struct {
    // ... existing fields ...
    
    // Project manager state (NEW)
    pmActive       bool
    selectedProject *project
    tasks          []task
}
```

---

## Keyboard/Mouse Input System

### 1. Keyboard Handling

**File:** `update_keyboard.go` (714 lines)

Supports:
- F-keys (F1-F12)
- Navigation keys (arrows, vim keys h/j/k/l)
- Special keys (Tab, Enter, Escape, Ctrl+X)
- Text input in dialogs and command prompt
- Context-aware behavior (different modes)

**Key Pattern for Extension:**

```go
func (m model) handleKeyEvent(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    // Early exits for special modes
    if m.fuzzySearchActive { return m, nil }
    if m.inputFieldsActive { /* handle input fields */ }
    
    // Then cascade through view modes
    if m.viewMode == viewFullPreview {
        // Preview-specific keys
    } else if m.viewMode == viewDualPane {
        // Dual-pane-specific keys
    }
    
    // Fall through to general file browser keys
    switch msg.String() {
    case "f1":
        // Help
    case "f2":
        // Context menu
    // ... more cases ...
    }
}
```

### 2. Mouse Handling

**File:** `update_mouse.go` (383 lines)

Features:
- Left click (select/navigate)
- Right click (context menu)
- Double click (open folder/file)
- Scroll wheel (navigation)
- Click detection with coordinate mapping

**Critical for PM:** Position detection is reliable for placing new UI elements.

---

## File Operations & Context Menu

### 1. Context Menu System

**File:** `context_menu.go` (313 lines)

Dynamically builds menu items based on:
- File type (directory vs file)
- Special file types (images, HTML, executables)
- View mode (trash vs normal)
- Available tools (lazygit if installed, etc.)

```go
func (m model) getContextMenuItems() []contextMenuItem {
    items := []contextMenuItem{}
    
    // Folder items
    if m.contextMenuFile.isDir {
        items = append(items, contextMenuItem{"📂 Open", "open"})
        items = append(items, contextMenuItem{"📂 Quick CD", "quickcd"})
        
        // Conditional TUI tool integration
        if editorAvailable("lazygit") {
            items = append(items, contextMenuItem{"🌿 Git (lazygit)", "lazygit"})
        }
        // ... more items
    }
    
    return items
}
```

**For PM Integration:** You could dynamically add PM actions:

```go
// NEW for PM mode
if m.pmActive && m.selectedProject != nil {
    items = append(items, contextMenuItem{"📋 Create Task", "create_task"})
    items = append(items, contextMenuItem{"🎯 Assign to Project", "assign_project"})
}
```

### 2. File Operations Pattern

Three types of operations:

**A. Simple (editor launch):**
```go
openEditor(editor, path)   // F4
openImageViewer(path)       // Right-click image
```

**B. With Interaction (copy/rename):**
```go
m.showDialog = true
m.dialog.dialogType = dialogInput
m.dialog.title = "Rename..."
```

**C. Command-based (Quick CD):**
```go
writeCDTarget(path)  // Write to ~/.tfe_cd_target
// Shell wrapper reads this and cds
```

---

## Extension Points for Project Manager Integration

### 1. Add New View Mode

**Minimal Change:**

In `types.go`:
```go
type viewMode int
const (
    // ... existing ...
    viewProjectManager  // NEW
)
```

In `view.go`:
```go
func (m model) View() string {
    switch m.viewMode {
    case viewProjectManager:
        return m.renderProjectManager()  // NEW function
    // ... existing cases ...
    }
}
```

### 2. Add Context Menu Actions

In `context_menu.go`:
```go
if m.pmActive {
    items = append(items, contextMenuItem{
        "📋 Add to Project",
        "add_to_project",
    })
}
```

In `executeContextMenuAction()`:
```go
case "add_to_project":
    m.pmActive = true
    m.selectedFile = m.contextMenuFile
    // Show project selection UI
    return m, nil
```

### 3. Add Keyboard Shortcut

In `update_keyboard.go`:
```go
case "ctrl+p":  // Already used for fuzzy search
    // Could repurpose or use different key
    
case "shift+p":  // NEW for Project Manager
    m.viewMode = viewProjectManager
    return m, nil
```

### 4. Extend Dialog System

The existing dialog system (`dialog.go`) handles input/confirm. Could add:

```go
const (
    dialogNone       = iota
    dialogInput
    dialogConfirm
    dialogMessage
    dialogProjectSelect  // NEW
    dialogTaskCreate     // NEW
)
```

### 5. Launch Project Manager as External App

**Alternative approach:** Instead of embedding, launch as separate process:

```go
// In context_menu.go
case "launch_pm":
    return m, openTUITool("pm", m.currentPath)  // Launch PM in dir
```

This is the **lowest friction integration** - your PM runs as a completely separate process but is launched contextually from TFE.

---

## Feature Examples for PM Integration

### 1. **Project-Aware File Browser**

```
TFE with PM mode enabled:
┌─────────────────────────────────────────┐
│ TFE - Project: MyApp                    │
├──────────────┬──────────────────────────┤
│ Files        │ Project Info             │
│              │                          │
│ ▸ src/       │ Tasks: 8 open            │
│ ▸ tests/     │ Contributors: 3          │
│ • main.go    │ Due: Oct 27              │
│              │ Status: On track         │
├──────────────┴──────────────────────────┤
│ [Add to Project] [View Tasks] [Assign]   │
└─────────────────────────────────────────┘
```

### 2. **Contextual Task Creation**

```
Right-click on file → "Create Task for this file"
→ Dialog: "Task: Fix bug in main.go"
→ Assigned to selected file
→ Added to project tasks
```

### 3. **Project Launcher Panel**

```
Add to toolbar (existing toolbar has 🏠 ⭐ 📝 🗑️):
Add 📋 (projects) button that toggles project view
```

### 4. **Seamless Git + Task Integration**

```
With lazygit already working:
1. User commits file
2. Commit message contains "Closes #123"
3. TFE could detect this and auto-link to task
4. Or show task status in commit message
```

---

## Dependencies & Tech Stack

From `go.mod`:

```go
require (
    github.com/charmbracelet/bubbletea v1.3.10    // TUI framework
    github.com/charmbracelet/lipgloss v1.1.1      // Styling
    github.com/charmbracelet/bubbles v0.21.0      // UI components
    github.com/charmbracelet/glamour v0.10.0      // Markdown rendering
    github.com/alecthomas/chroma/v2 v2.14.0       // Syntax highlighting
    github.com/koki-develop/go-fzf v0.15.0        // Fuzzy search
    gopkg.in/yaml.v3 v3.0.1                       // Config files
)
```

**For PM Integration:** You'd use the same tech stack:
- Bubbletea for TUI
- Lipgloss for styling
- YAML or JSON for project configs

---

## Critical Insights

### 1. TFE is Built for Extension

The modular architecture isn't accidental - it's designed with future features in mind. Adding a PM mode is conceptually straightforward.

### 2. App Launching is First-Class

TFE treats external tool launching as core functionality (lazygit, lazydocker, etc.). This means the infrastructure for subprocess management is mature and tested.

### 3. Event Model is Extensible

The Bubbletea message-passing architecture means you can add new event types without breaking existing code:

```go
type projectTaskSelectedMsg struct {
    taskID string
}

// Then handle in Update()
case projectTaskSelectedMsg:
    m.selectedTask = msg.taskID
    return m, nil
```

### 4. Terminal State Management is Robust

TFE handles terminal state transitions cleanly:
- Saves state before launching external apps
- Restores state after
- File list auto-refreshes
- No terminal corruption artifacts

This is critical for a PM that might launch multiple child processes.

### 5. Single Source of Truth Pattern

Each component has one responsibility:
- `context_menu.go` - Menu generation and execution
- `editor.go` - External app launching
- `command.go` - Shell command execution
- `update_keyboard.go` - Input handling

For PM: You'd add `project_manager.go` for all PM logic, keeping concerns separate.

---

## Natural Integration Points

### Top Priority:
1. **Add context menu action** to existing files
   - Zero impact on other features
   - Quick to implement
   - "Create task for this file"

### Medium Priority:
2. **Add new keyboard shortcut** (e.g., Ctrl+Shift+P)
   - Shows project dashboard
   - Doesn't break existing functionality
   - Can toggle on/off

### Higher Effort:
3. **Add PM toolbar button**
   - Modify existing toolbar rendering
   - Add PM toggle mode
   - More UI integration needed

### Most Complex:
4. **Embed PM as side panel**
   - New viewMode (viewDualPaneWithPM)
   - Layout calculations change
   - But architectural pattern already exists (dual-pane mode)

---

## Recommendations for PM/TFE Integration

### Option 1: **Standalone PM Launcher**
- **Best for:** Minimal coupling
- **How:** PM runs as `pm` command, TFE launches it like `lazygit`
- **Pros:** Both tools stay independent, composable
- **Cons:** Can't share state between TFE and PM

### Option 2: **Extend TFE with PM Mode**
- **Best for:** Unified experience
- **How:** Add `viewProjectManager` mode, add PM logic to `update_keyboard.go`
- **Pros:** Single unified TUI, shared file context
- **Cons:** Couples PM logic into TFE

### Option 3: **PM as Sidebar**
- **Best for:** Context-aware project work
- **How:** New viewMode combining file browser + project dashboard
- **Pros:** See files and projects simultaneously
- **Cons:** Complex layout management (but pattern exists in dual-pane)

### Option 4: **Context Menu Integration**
- **Best for:** Gradual integration
- **How:** Add "📋 Project Tasks" to context menu, launches PM subprocess
- **Pros:** Non-intrusive, low risk
- **Cons:** Not as integrated as other options

---

## Workflow Example: Project-Aware Development

**Current TFE workflow:**
```
1. Open TFE in project directory
2. Navigate to src/main.go
3. Press F4 to edit in vim
4. After saving, press F2 → "🌿 Git (lazygit)"
5. Commit changes
6. Esc to return to TFE
```

**With PM Integration (Option 2):**
```
1. Open TFE in project directory
2. Navigate to src/main.go
3. Press Ctrl+Shift+P to show project sidebar
4. See tasks: "Fix bug #42: Incorrect parsing" (RED - In Progress)
5. Press F4 to edit in vim
6. After saving, return to TFE
7. Press F2 → "📋 Link to Task" → Select task #42
8. File now linked to task in sidebar
9. Commit with F2 → "🌿 Git (lazygit)"
10. Commit message auto-includes "Closes #42"
11. Task status updates in sidebar to "Ready for Review"
```

---

## Summary Table

| Aspect | Current TFE | Integration Point | PM Opportunity |
|--------|------------|-------------------|-----------------|
| **View Modes** | Single, Dual-pane, Full preview | `types.go` + `view.go` | Add PM mode |
| **Context Menu** | File operations + lazygit | `context_menu.go` | Add task creation |
| **Keyboard** | F1-F12 + custom | `update_keyboard.go` | Add PM shortcuts |
| **External Apps** | Mature system | `editor.go` | Launch PM subprocess |
| **State** | File browser focused | `model` struct | Add PM fields |
| **Dialogs** | Input/confirm/message | `dialog.go` | Extend for project select |

---

## Files to Study for Implementation

### Essential:
1. **types.go** - Understand model struct and enums
2. **context_menu.go** - How actions are added and executed
3. **update_keyboard.go** - Event handling pattern
4. **editor.go** - Process launching pattern
5. **view.go** - View rendering dispatcher

### Reference:
6. **CLAUDE.md** - Architecture philosophy
7. **PLAN.md** - Roadmap and design decisions

---

