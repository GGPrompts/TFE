# TFE Code Patterns - Copy-Paste Ready Examples

This document shows actual patterns from TFE code that you can adapt for PM integration.

## Pattern 1: Detecting and Launching External Apps

**Source:** `editor.go` lines 12-16 and 40-50

### Detection Pattern
```go
// Check if a command-line tool is available
func editorAvailable(cmd string) bool {
    _, err := exec.LookPath(cmd)
    return err == nil
}

// Get first available from list
func getAvailableImageViewer() string {
    viewers := []string{"viu", "timg", "chafa"}
    for _, viewer := range viewers {
        if editorAvailable(viewer) {
            return viewer
        }
    }
    return ""
}
```

### Launching Pattern
```go
// Launch a TUI tool in the specified directory
func openTUITool(tool, dir string) tea.Cmd {
    c := exec.Command(tool)
    c.Dir = dir  // Set working directory - CRITICAL!
    
    return tea.Sequence(
        tea.ClearScreen,
        tea.ExecProcess(c, func(err error) tea.Msg {
            return editorFinishedMsg{err}
        }),
    )
}

// Usage:
case "lazygit":
    if m.contextMenuFile.isDir {
        return m, openTUITool("lazygit", m.contextMenuFile.path)
    }
```

**For your PM:**
```go
// In editor.go, add or modify:
func openProjectManager(dir string) tea.Cmd {
    // Same pattern as openTUITool
    c := exec.Command("pm")  // Assumes "pm" command exists
    c.Dir = dir
    
    return tea.Sequence(
        tea.ClearScreen,
        tea.ExecProcess(c, func(err error) tea.Msg {
            return editorFinishedMsg{err}
        }),
    )
}
```

---

## Pattern 2: Conditionally Adding Menu Items

**Source:** `context_menu.go` lines 54-102

### Detection + Menu Building
```go
func (m model) getContextMenuItems() []contextMenuItem {
    if m.contextMenuFile == nil {
        return []contextMenuItem{}
    }

    items := []contextMenuItem{}

    if m.contextMenuFile.isDir {
        // Directory-specific items
        items = append(items, contextMenuItem{"📂 Open", "open"})
        items = append(items, contextMenuItem{"📂 Quick CD", "quickcd"})
        
        // CONDITIONAL: Only show if tool is available
        hasTools := false
        if editorAvailable("lazygit") {
            if !hasTools {
                items = append(items, contextMenuItem{"─────────", "separator"})
                hasTools = true
            }
            items = append(items, contextMenuItem{"🌿 Git (lazygit)", "lazygit"})
        }
        if editorAvailable("lazydocker") {
            if !hasTools {
                items = append(items, contextMenuItem{"─────────", "separator"})
                hasTools = true
            }
            items = append(items, contextMenuItem{"🐋 Docker (lazydocker)", "lazydocker"})
        }
        
        // Add separator and favorites
        items = append(items, contextMenuItem{"─────────", "separator"})
        items = append(items, contextMenuItem{"📋 Copy to...", "copy"})
    }

    return items
}
```

### Execution Pattern
```go
func (m model) executeContextMenuAction() (tea.Model, tea.Cmd) {
    if m.contextMenuFile == nil {
        m.contextMenuOpen = false
        return m, nil
    }

    items := m.getContextMenuItems()
    if m.contextMenuCursor >= len(items) {
        m.contextMenuOpen = false
        return m, nil
    }

    item := items[m.contextMenuCursor]
    
    switch item.action {
    case "open":
        // Navigate into directory
        m.currentPath = m.contextMenuFile.path
        m.loadFiles()
        m.cursor = 0
        m.contextMenuOpen = false
        return m, nil

    case "lazygit":
        // Launch git tool
        if m.contextMenuFile.isDir {
            return m, openTUITool("lazygit", m.contextMenuFile.path)
        }
        return m, nil
        
    case "copy":
        // Show copy dialog
        m.showDialog = true
        m.dialog.dialogType = dialogInput
        m.dialog.title = "Copy to..."
        m.dialog.input = ""
        m.contextMenuOpen = false
        return m, nil
    }

    return m, nil
}
```

**For your PM:**
```go
// In getContextMenuItems(), add after lazydocker check:
if editorAvailable("pm") {
    if !hasTools {
        items = append(items, contextMenuItem{"─────────", "separator"})
        hasTools = true
    }
    items = append(items, contextMenuItem{"📋 Project Manager", "launch_pm"})
}

// In executeContextMenuAction(), add new case:
case "launch_pm":
    if m.contextMenuFile.isDir {
        m.contextMenuOpen = false
        return m, openTUITool("pm", m.contextMenuFile.path)
    }
    return m, nil
```

---

## Pattern 3: Adding Keyboard Shortcuts

**Source:** `update_keyboard.go` lines 23-100

### Cascade Pattern (Most Important!)
```go
func (m model) handleKeyEvent(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    // 1. FIRST: Handle special modes that consume all input
    if m.fuzzySearchActive { return m, nil }
    if m.inputFieldsActive { /* handle special input */ }
    
    // 2. THEN: Handle view-mode-specific keys
    if m.viewMode == viewFullPreview {
        switch msg.String() {
        case "f4":
            // Edit in external editor
        case "f5":
            // Copy path
        case "esc":
            // Exit full preview
        }
    } else if m.viewMode == viewDualPane {
        switch msg.String() {
        case "tab":
            // Switch between panes
        case "space":
            // Exit dual-pane
        }
    }
    
    // 3. FINALLY: Handle general file browser keys
    switch msg.String() {
    case "f1":
        // Help
    case "f2":
        m.contextMenuOpen = true
    case "space":
        // Toggle dual pane
        if m.viewMode == viewSinglePane {
            m.viewMode = viewDualPane
        } else {
            m.viewMode = viewSinglePane
        }
    // ... more cases ...
    }
}
```

**For your PM:**
```go
// Add new view mode in types.go
type viewMode int
const (
    viewSinglePane
    viewDualPane
    viewFullPreview
    viewProjectManager  // NEW
)

// In update_keyboard.go handleKeyEvent(), add before general keys:
if m.viewMode == viewProjectManager {
    switch msg.String() {
    case "tab":
        // Switch focus within PM (tasks/projects/etc)
        m.pmFocusedPanel++
        if m.pmFocusedPanel >= pmPanelCount {
            m.pmFocusedPanel = 0
        }
        return m, nil
        
    case "esc", "q":
        // Exit PM view
        m.viewMode = viewSinglePane
        return m, nil
        
    case "enter":
        // Select task/project
        if m.pmFocusedPanel == taskPanel {
            // Launch task editor
        }
        return m, nil
    }
}

// Add shortcut to toggle PM view
case "ctrl+shift+p":
    if m.viewMode == viewProjectManager {
        m.viewMode = viewSinglePane
    } else {
        m.viewMode = viewProjectManager
    }
    return m, nil
```

---

## Pattern 4: Adding New View Mode

**Source:** `view.go` lines 30-50 and `types.go` lines 32-40

### Type Definition
```go
// In types.go
type viewMode int

const (
    viewSinglePane viewMode = iota
    viewDualPane
    viewFullPreview
)

func (v viewMode) String() string {
    switch v {
    case viewSinglePane:
        return "Single"
    case viewDualPane:
        return "Dual-Pane"
    case viewFullPreview:
        return "Full Preview"
    default:
        return "Unknown"
    }
}
```

### View Dispatcher
```go
// In view.go - the main rendering function
func (m model) View() string {
    switch m.viewMode {
    case viewSinglePane:
        return m.renderSinglePane()
    case viewDualPane:
        return m.renderDualPane()
    case viewFullPreview:
        return m.renderFullPreview()
    default:
        return "Unknown view mode"
    }
}
```

**For your PM:**
```go
// 1. Add to types.go viewMode enum
const (
    viewSinglePane
    viewDualPane
    viewFullPreview
    viewProjectManager  // NEW
)

// 2. Add String() case
case viewProjectManager:
    return "Projects"

// 3. Add to view.go dispatcher
case viewProjectManager:
    return m.renderProjectManager()

// 4. Create new file: project_manager.go
func (m model) renderProjectManager() string {
    // Build and return PM UI as string
    // Use lipgloss for styling
    
    pmView := lipgloss.NewStyle().
        Width(m.width).
        Height(m.height).
        Render("Project Manager View")
    
    return pmView
}
```

---

## Pattern 5: Using the Dialog System

**Source:** `dialog.go` and usage in `update_keyboard.go`

### Type Definitions
```go
// In types.go
type dialogType int

const (
    dialogNone dialogType = iota
    dialogInput
    dialogConfirm
    dialogMessage
)

type dialogModel struct {
    dialogType dialogType
    title      string
    message    string
    input      string        // For text input dialogs
    confirmed  bool          // User confirmed action
    isError    bool          // For message dialogs (red vs green)
}
```

### Opening a Dialog
```go
// When user triggers an action that needs input
case "newfolder":
    m.showDialog = true
    m.dialog.dialogType = dialogInput
    m.dialog.title = "New Folder Name:"
    m.dialog.input = ""
    return m, nil

case "delete":
    m.showDialog = true
    m.dialog.dialogType = dialogConfirm
    m.dialog.title = "Delete?"
    m.dialog.message = fmt.Sprintf("Delete '%s'?", m.contextMenuFile.name)
    return m, nil
```

### Handling Dialog Results
```go
// When dialog completes (user presses Enter or confirms)
if m.dialog.dialogType == dialogInput && m.dialog.confirmed {
    // User entered text in m.dialog.input
    folderName := m.dialog.input
    
    // Perform action
    err := os.Mkdir(folderName, 0755)
    
    // Show result
    if err != nil {
        m.setStatusMessage(fmt.Sprintf("Error: %v", err), true)
    } else {
        m.setStatusMessage("Folder created", false)
    }
    
    // Clear dialog
    m.showDialog = false
    m.dialog.dialogType = dialogNone
    m.loadFiles()
}
```

**For your PM:**
```go
// Extend dialogType in types.go
const (
    dialogNone
    dialogInput
    dialogConfirm
    dialogMessage
    dialogSelectProject  // NEW
    dialogCreateTask     // NEW
)

// When user wants to create task
case "create_task":
    m.showDialog = true
    m.dialog.dialogType = dialogCreateTask
    m.dialog.title = fmt.Sprintf("Create Task for: %s", m.contextMenuFile.name)
    m.dialog.input = ""
    return m, nil

// Handle in dialog completion
if m.dialog.dialogType == dialogCreateTask && m.dialog.confirmed {
    taskName := m.dialog.input
    filePath := m.contextMenuFile.path
    
    // Call PM backend to create task
    err := createTaskForFile(taskName, filePath)
    
    if err != nil {
        m.setStatusMessage(fmt.Sprintf("Error: %v", err), true)
    } else {
        m.setStatusMessage("Task created", false)
    }
    
    m.showDialog = false
    m.dialog.dialogType = dialogNone
}
```

---

## Pattern 6: File Operations with Status Feedback

**Source:** `file_operations.go` and usage throughout

### The Status Message Pattern
```go
// In model struct (types.go)
type model struct {
    // ... other fields ...
    statusMessage string    // Temporary status message
    statusIsError bool      // Whether status message is an error
    statusTime    time.Time // When status was shown
}

// Set status with auto-clear
func (m *model) setStatusMessage(message string, isError bool) {
    m.statusMessage = message
    m.statusIsError = isError
    m.statusTime = time.Now()
}

// In view rendering - status auto-clears after 3 seconds
func (m model) renderStatus() string {
    elapsed := time.Since(m.statusTime)
    if elapsed > 3*time.Second {
        m.statusMessage = ""
    }
    
    if m.statusMessage == "" {
        return normalStatusBar
    }
    
    if m.statusIsError {
        return errorStatusBar  // Red
    } else {
        return successStatusBar  // Green
    }
}
```

### Using the Status System
```go
// Example: Copy file operation
case "copy":
    srcPath := m.contextMenuFile.path
    destPath := /* user selected destination */
    
    err := copyFileOrDir(srcPath, destPath)
    
    if err != nil {
        // Show error in red, auto-disappears after 3s
        m.setStatusMessage(fmt.Sprintf("Error copying: %v", err), true)
    } else {
        // Show success in green, auto-disappears after 3s
        m.setStatusMessage("File copied successfully", false)
        m.loadFiles()  // Refresh file list
    }
    
    m.contextMenuOpen = false
    return m, nil
```

**For your PM:**
```go
// Use same pattern for PM operations
case "save_task":
    task := m.pmSelectedTask
    
    err := savePMTask(task)
    
    if err != nil {
        m.setStatusMessage(fmt.Sprintf("Error saving task: %v", err), true)
    } else {
        m.setStatusMessage("Task saved", false)
    }
    
    return m, nil
```

---

## Pattern 7: Message Types for Async Operations

**Source:** `types.go` and `command.go` for message definitions and handling

### Message Type
```go
// In types.go
type editorFinishedMsg struct{ err error }
type commandFinishedMsg struct{ err error }
type browserOpenedMsg struct {
    success bool
    err     error
}
```

### Sending a Message from Async Operation
```go
// In command.go - runs in background
func runCommand(command, dir string) tea.Cmd {
    return func() tea.Msg {
        // Execute command...
        err := executeCommand()
        
        // Send message back to main loop
        return commandFinishedMsg{err: err}
    }
}
```

### Handling the Message
```go
// In update.go - main message handler
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case commandFinishedMsg:
        // Async operation completed
        m.loading = false
        
        if msg.err != nil {
            m.setStatusMessage(fmt.Sprintf("Error: %v", msg.err), true)
        } else {
            m.setStatusMessage("Command completed", false)
        }
        
        m.loadFiles()  // Refresh
        return m, nil
        
    case editorFinishedMsg:
        // External editor closed
        m.loadFiles()
        return m, nil
    }
}
```

**For your PM:**
```go
// Define message type
type pmTaskSavedMsg struct {
    taskID string
    err    error
}

// Send from async operation
func savePMTaskAsync(taskID string) tea.Cmd {
    return func() tea.Msg {
        err := savePMTaskBackend(taskID)
        return pmTaskSavedMsg{taskID: taskID, err: err}
    }
}

// Handle in Update()
case pmTaskSavedMsg:
    if msg.err != nil {
        m.setStatusMessage("Error saving task", true)
    } else {
        m.setStatusMessage("Task saved", false)
    }
    return m, nil
```

---

## Pattern 8: Rendering with Lipgloss

**Source:** `styles.go` and all `render_*` files

### Style Definition
```go
// In styles.go
var (
    titleStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FAFAFA")).
        Background(lipgloss.Color("#7D56C4")).
        Bold(true).
        Padding(0, 1)

    statusStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FAFAFA")).
        Background(lipgloss.Color("#3C3C3C"))

    selectedStyle = lipgloss.NewStyle().
        Background(lipgloss.Color("#555555"))

    errorStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FF0000"))

    successStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#00FF00"))
)
```

### Using Styles in Rendering
```go
// In view.go or render_*.go
func (m model) renderHeader() string {
    header := fmt.Sprintf("TFE - %s", m.currentPath)
    return titleStyle.Width(m.width).Render(header)
}

func (m model) renderStatus() string {
    if m.statusIsError {
        return errorStyle.Render(m.statusMessage)
    } else {
        return successStyle.Render(m.statusMessage)
    }
}

// Full view composition
func (m model) View() string {
    header := m.renderHeader()
    fileList := m.renderFileList()
    status := m.renderStatus()
    
    return lipgloss.JoinVertical(lipgloss.Left,
        header,
        fileList,
        status,
    )
}
```

**For your PM:**
```go
// Add new styles in styles.go
var (
    pmPanelStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        Padding(1)

    pmSelectedStyle = lipgloss.NewStyle().
        Background(lipgloss.Color("#444444")).
        Padding(0, 1)

    pmHighPriorityStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FF0000")).
        Bold(true)

    pmLowPriorityStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#00AA00"))
)

// Use in renderProjectManager()
func (m model) renderProjectManager() string {
    taskList := m.renderPMTaskList()
    taskDetails := m.renderPMTaskDetails()
    
    return lipgloss.JoinHorizontal(lipgloss.Top,
        pmPanelStyle.Render(taskList),
        pmPanelStyle.Render(taskDetails),
    )
}
```

---

## Summary: Implementation Roadmap

### To add Context Menu Item (Minimal):
1. Add detection: `if editorAvailable("pm") { ... }`
2. Add menu item: `contextMenuItem{"📋 Project Manager", "launch_pm"}`
3. Add handler: `case "launch_pm": return m, openTUITool("pm", ...)`

### To add PM View Mode (Medium):
1. Add to types: `viewProjectManager` in viewMode enum
2. Add handler: `case "ctrl+shift+p":` toggle to viewProjectManager
3. Add renderer: `case viewProjectManager: return m.renderProjectManager()`
4. Create file: `project_manager.go` with render function

### To add PM as Side Panel (Complex):
All above PLUS:
1. Change layout calculations in `model.go`
2. Add new styles in `styles.go`
3. Handle pane switching in `update_keyboard.go`
4. Update view compositor in `view.go`

---

