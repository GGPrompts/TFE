# TFE Development Patterns

This document provides detailed examples and patterns for common development tasks in TFE.

## Table of Contents
- [Adding New Features](#adding-new-features)
- [Creating New Modules](#creating-new-modules)
- [Common Patterns](#common-patterns)
- [Code Organization Principles](#code-organization-principles)
- [Building and Installing](#building-and-installing)

---

## Adding New Features

When adding new features, follow this decision tree to determine where the code should live:

### Decision Tree

1. **Is it a new type or data structure?** → Add to `types.go`
2. **Is it a visual style?** → Add to `styles.go`
3. **Is it event handling (keyboard/mouse)?** → Add to `update_keyboard.go` or `update_mouse.go`
4. **Is it a rendering function?** → Add to `view.go`, `render_preview.go`, or `render_file_list.go`
5. **Is it a file operation?** → Add to `file_operations.go`
6. **Is it external tool integration?** → Add to `editor.go` or create new module
7. **Is it complex enough to need its own module?** → Create a new file

### Examples

**Example 1: Adding a new file type icon**
- Decision: File type detection → `file_operations.go`
- Add extension to `getIconForExtension()` function

**Example 2: Adding a keyboard shortcut for splitting panes**
- Decision: Keyboard event handling → `update_keyboard.go`
- Add case to the keyboard handler switch statement

**Example 3: Adding a new theme/color scheme**
- Decision: Visual styling → `styles.go`
- Define new Lipgloss styles for the theme

---

## Creating New Modules

If a feature is substantial enough to warrant its own module, follow these steps:

### Steps to Create a New Module

1. **Create a new `.go` file** with a descriptive name (e.g., `search.go`, `bookmarks.go`)
2. **Keep it in `package main`** (all files share the same package)
3. **Document the module's purpose** at the top of the file
4. **Add it to `docs/MODULE_DETAILS.md`** under the appropriate section
5. **Update the module quick reference in `CLAUDE.md`**

### Module Template

```go
package main

// Module: search.go
// Purpose: File and content search functionality
// Responsibilities:
// - Search indexing
// - Pattern matching
// - Search result filtering

import (
    // ... imports
)

// ... implementation
```

### When to Create a New Module

Create a new module when:
- The feature requires 200+ lines of code
- The functionality is self-contained and independent
- Multiple related functions share a common purpose
- The feature might need separate testing

Don't create a new module when:
- The feature is a simple helper function (add to `helpers.go`)
- It's tightly coupled to an existing module (extend that module)
- It's less than 100 lines of code

---

## Common Patterns

### Pattern 1: Modifying the Header/Title Bar

**⚠️ IMPORTANT: Headers exist in TWO locations!**

When modifying the header/title bar (GitHub link, menu bar, mode indicators), you must update BOTH:
- **Single-Pane:** `view.go` → `renderSinglePane()` (~line 64)
- **Dual-Pane:** `render_preview.go` → `renderDualPane()` (~line 816)

**Note:** `renderFullPreview()` has a different header intentionally (shows filename, not menu bar).

**Why this matters:** Forgetting to update both locations leads to inconsistent UI between view modes. Extract shared header rendering to fix this duplication (see PLAN.md issue #14).

---

### Pattern 2: Adding a New Keyboard Shortcut

**Steps:**
1. Go to `update_keyboard.go`
2. Find the appropriate switch statement (preview mode vs regular mode)
3. Add a new case for your key
4. Implement the logic or call a function from another module

**Example:**
```go
case "s":
    // Save bookmark
    m.saveBookmark(m.files[m.cursor].path)
```

**Full Example with Context:**
```go
// In handleKeyEvent() function, around line 200
switch key.String() {
case "s":
    // Toggle favorite for current file
    file := m.getCurrentFile()
    if file != nil {
        m.toggleFavorite(file.path)
        m.setStatusMessage("Toggled favorite", false)
    }
    return m, statusTimeoutCmd()
}
```

---

### Pattern 3: Adding a New Display Mode

**Steps:**

1. **Add the enum to `types.go`:**
```go
const (
    modeList displayMode = iota
    modeDetail
    modeTree
    modeYourNewMode  // Add here
)
```

2. **Add rendering function to `render_file_list.go`:**
```go
func (m model) renderYourNewMode(maxVisible int) string {
    var output strings.Builder

    // Build your custom view
    for i, file := range visibleFiles {
        // Custom rendering logic
        output.WriteString(file.name)
        output.WriteString("\n")
    }

    return output.String()
}
```

3. **Update the switch in `render_file_list.go` or `view.go`:**
```go
switch m.displayMode {
case modeList:
    return m.renderListView(maxVisible)
case modeDetail:
    return m.renderDetailView(maxVisible)
case modeTree:
    return m.renderTreeView(maxVisible)
case modeYourNewMode:
    return m.renderYourNewMode(maxVisible)
}
```

4. **Add keyboard shortcut in `update_keyboard.go`:**
```go
case "f4":  // Or whatever key you want
    m.displayMode = modeYourNewMode
    m.setStatusMessage("Switched to Your New Mode", false)
    return m, statusTimeoutCmd()
```

---

### Pattern 4: Adding a New File Operation

**Steps:**

1. **Add the function to `file_operations.go`:**
```go
func (m *model) yourNewOperation(path string) error {
    // Implementation
    file, err := os.Open(path)
    if err != nil {
        return fmt.Errorf("failed to open file: %w", err)
    }
    defer file.Close()

    // Do operation...

    return nil
}
```

2. **Add keyboard shortcut in `update_keyboard.go` to call it:**
```go
case "x":
    file := m.getCurrentFile()
    if file == nil {
        return m, nil
    }

    if err := m.yourNewOperation(file.path); err != nil {
        m.setStatusMessage(fmt.Sprintf("Operation failed: %v", err), true)
    } else {
        m.setStatusMessage("Operation completed successfully", false)
    }
    return m, statusTimeoutCmd()
```

---

### Pattern 5: Adding a Context Menu Item

**Steps:**

1. **Add the menu item in `context_menu.go`:**
```go
items := []contextMenuItem{
    {label: "Open", action: "open"},
    {label: "Edit", action: "edit"},
    {label: "Your New Action", action: "your_action"},  // Add here
    // ... more items
}
```

2. **Handle the action in the context menu handler:**
```go
case "your_action":
    // Implement your action
    m.yourNewOperation(m.getCurrentFile().path)
    m.contextMenuOpen = false
    return m, nil
```

---

### Pattern 6: Adding a New Message Type

For asynchronous operations (background tasks), create a custom message type:

**1. Add message type to `types.go`:**
```go
type yourOperationFinishedMsg struct {
    success bool
    result  string
    err     error
}
```

**2. Create a command function:**
```go
func doYourOperation(path string) tea.Cmd {
    return func() tea.Msg {
        // Do long-running operation
        result, err := performOperation(path)

        return yourOperationFinishedMsg{
            success: err == nil,
            result:  result,
            err:     err,
        }
    }
}
```

**3. Handle the message in `update.go`:**
```go
case yourOperationFinishedMsg:
    if msg.success {
        m.setStatusMessage("Operation completed: " + msg.result, false)
    } else {
        m.setStatusMessage("Operation failed: " + msg.err.Error(), true)
    }
    return m, statusTimeoutCmd()
```

---

## Code Organization Principles

### 1. Single Responsibility
Each file should have one clear purpose. Don't mix concerns.

**Good:**
- `file_operations.go` - Only file system operations
- `render_preview.go` - Only preview rendering

**Bad:**
- Mixing file operations and rendering in one file
- Adding networking code to UI rendering

---

### 2. Keep main.go Minimal
`main.go` should ONLY contain the program entry point.

**Good:**
```go
func main() {
    p := tea.NewProgram(initialModel(), tea.WithAltScreen())
    if err := p.Start(); err != nil {
        log.Fatal(err)
    }
}
```

**Bad:**
```go
func main() {
    // 500 lines of business logic
    // UI rendering code
    // File operations
    // ...
}
```

---

### 3. Group Related Functions
Keep related functions together in the same file.

**Good:**
```go
// In file_operations.go
func loadFiles() { }
func loadSubdirFiles() { }
func loadPreview() { }
```

**Bad:**
```go
// Scattered across multiple files
// file_operations.go: loadFiles()
// helpers.go: loadSubdirFiles()
// model.go: loadPreview()
```

---

### 4. Separate Concerns
UI rendering should be separate from business logic.

**Good:**
- `file_operations.go` - Business logic (file reading)
- `render_file_list.go` - UI rendering (display files)

**Bad:**
- Mixing file reading and rendering in the same function

---

### 5. DRY (Don't Repeat Yourself)
Extract common logic into helper functions.

**Good:**
```go
// In helpers.go
func (m *model) getCurrentFile() *fileItem {
    // Common logic used everywhere
}

// Used in multiple places
file := m.getCurrentFile()
```

**Bad:**
```go
// Repeated in 10 different files
if m.cursor < len(m.files) {
    file := m.files[m.cursor]
    // ...
}
```

---

### 6. Clear Naming
File names should immediately convey their purpose.

**Good naming:**
- `git_operations.go` - Obviously handles git operations
- `fuzzy_search.go` - Clearly about fuzzy search
- `terminal_graphics.go` - Terminal graphics protocols

**Bad naming:**
- `utils.go` - Too generic
- `stuff.go` - Meaningless
- `helpers2.go` - Why is there a 2?

---

## Building and Installing

### Development Workflow

**⚠️ IMPORTANT: After rebuilding TFE, always update the installed binary!**

The user has TFE installed at `/home/matt/.local/bin/tfe` and frequently rebuilds during development.

**Standard build and install:**
```bash
# Build the binary
go build

# Copy to installation directory
cp ./tfe /home/matt/.local/bin/tfe
```

**Why this matters:**
- The user tests with `./tfe` in the project folder during development
- But uses `tfe` (from PATH) in normal usage
- Forgetting to update the installed binary means the user won't have the latest fixes
- This is especially critical after bug fixes or feature additions

**When to do this:**
- After any `go build` command
- After fixing bugs that need testing
- After implementing new features
- Before asking the user to test the installed version

---

## Testing Strategy

When adding tests (future):
- Create corresponding `*_test.go` files alongside each module
- Test files should mirror the structure: `file_operations_test.go`, `render_preview_test.go`, etc.
- Keep test files focused on their corresponding module

**Example test file structure:**
```go
// file_operations_test.go
package main

import "testing"

func TestLoadFiles(t *testing.T) {
    // Test loadFiles() function
}

func TestFormatFileSize(t *testing.T) {
    // Test formatFileSize() function
}
```

---

## Important Reminders

**🚨 When adding new features, always maintain the modular architecture!**

Do NOT add complex logic to `main.go`. Instead:
- Identify which module the feature belongs to
- Add it to that module, or create a new one
- Keep files focused and organized
- Update `docs/MODULE_DETAILS.md` when creating new modules
- Update the module quick reference in `CLAUDE.md`

This architecture took significant effort to establish - let's maintain it!
