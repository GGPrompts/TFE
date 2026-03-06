# TFE Architecture & Development Guide

This document is a **concise index** to TFE's architecture and development documentation. For detailed information, follow the links to specific guides.

---

## Quick Start: When to Read What

**Before starting ANY task, check this decision tree:**

- 🎨 **Working on UI/rendering/width calculations?** → Read [`docs/LESSONS_LEARNED.md`](docs/LESSONS_LEARNED.md) FIRST (critical!)
- 📦 **Adding a new feature?** → Read ["Where Does My Code Go?"](#where-does-my-code-go) below
- 🔧 **Need to know what a module does?** → See [Module Quick Reference](#module-quick-reference) below or [`docs/MODULE_DETAILS.md`](docs/MODULE_DETAILS.md)
- 📝 **Adding keyboard shortcuts or display modes?** → Read [`docs/DEVELOPMENT_PATTERNS.md`](docs/DEVELOPMENT_PATTERNS.md)
- 🔒 **Questions about "security issues"?** → Read [`docs/THREAT_MODEL.md`](docs/THREAT_MODEL.md)
- 📚 **Managing documentation?** → Read [`docs/DOCUMENTATION_GUIDE.md`](docs/DOCUMENTATION_GUIDE.md)
- 🏗️ **Want architectural history?** → Read [`docs/REFACTORING_HISTORY.md`](docs/REFACTORING_HISTORY.md)

**After making code changes:**
- ⚠️ **ALWAYS run:** `go test ./...` to verify tests pass (see [Testing Strategy](#testing-strategy))
- ⚠️ **ALWAYS run:** `./build.sh` OR manually update both `~/bin/tfe` and `~/.local/bin/tfe` (see [Building and Installing](#building-and-installing))

---

## Architecture Overview

TFE follows a **modular architecture** where each file has a single, clear responsibility. This organization was established through a comprehensive refactoring that reduced `main.go` from 1668 lines to just 21 lines, distributing functionality across 19 focused modules.

### Core Principle

**When adding new features, always maintain this modular architecture by creating new files or extending existing modules rather than adding everything to `main.go`.**

---

## File Structure

```
tfe/
├── main.go                    # Entry point ONLY (21 lines)
├── types.go                   # Type definitions & enums
├── styles.go                  # Lipgloss style definitions
├── model.go                   # Model initialization & layout
├── update.go                  # Main update dispatcher
├── update_keyboard.go         # Keyboard event handling
├── update_mouse.go            # Mouse event handling
├── view.go                    # View dispatcher
├── menu.go                    # Menu bar rendering
├── render_preview.go          # Preview rendering
├── render_file_list.go        # File list views
├── file_operations.go         # File operations (load, preview, CRUD)
├── file_icons.go              # File type detection & icons
├── text_wrapping.go           # Width calculations & text wrapping
├── editor.go                  # External tool integration
├── command.go                 # Command execution
├── git_operations.go          # Git operations & status queries
├── dialog.go                  # Dialog system
├── context_menu.go            # Context menu
├── favorites.go               # Favorites/bookmarks
├── trash.go                   # Trash/recycle bin
├── prompt_parser.go           # Prompt template parsing
├── fuzzy_search.go            # Fuzzy file search
├── terminal_graphics.go       # HD image preview
└── helpers.go                 # Helper functions
```

---

## Module Quick Reference

**For full details on any module, see [`docs/MODULE_DETAILS.md`](docs/MODULE_DETAILS.md)**

### Core Modules
- **`main.go`** - Application entry point only
- **`types.go`** - All type definitions, structs, enums
- **`styles.go`** - Lipgloss style definitions
- **`model.go`** - Model initialization & layout calculations

### Event Handling
- **`update.go`** - Main update dispatcher
- **`update_keyboard.go`** - All keyboard input processing
- **`update_mouse.go`** - All mouse input processing

### Rendering
- **`view.go`** - Top-level view dispatching
- **`render_preview.go`** - File preview rendering
- **`render_file_list.go`** - File list in all display modes
- **`menu.go`** - Menu bar rendering

### File System
- **`file_operations.go`** - File loading, preview, CRUD operations
- **`file_icons.go`** - File type detection, icons, metadata formatting
- **`favorites.go`** - Bookmarks system
- **`trash.go`** - Trash/recycle bin

### Text & Layout
- **`text_wrapping.go`** - Width calculations, text wrapping, truncation

### External Integration
- **`editor.go`** - Editors, browsers, clipboard
- **`command.go`** - Command execution & history
- **`git_operations.go`** - Git operations & status queries

### UI Components
- **`dialog.go`** - Input/confirmation dialogs
- **`context_menu.go`** - Right-click context menu

### Search & Advanced
- **`fuzzy_search.go`** - Ctrl+P fuzzy search (fzf)
- **`prompt_parser.go`** - Prompt template variables
- **`terminal_graphics.go`** - HD image preview

### Utilities
- **`helpers.go`** - Reusable helper functions

---

## Where Does My Code Go?

Follow this decision tree when adding new features:

1. **Is it a new type or data structure?** → `types.go`
2. **Is it a visual style?** → `styles.go`
3. **Is it event handling (keyboard/mouse)?** → `update_keyboard.go` or `update_mouse.go`
4. **Is it a rendering function?** → `view.go`, `render_preview.go`, or `render_file_list.go`
5. **Is it a file operation?** → `file_operations.go`
6. **Is it external tool integration?** → `editor.go` or create new module
7. **Is it complex enough to need its own module?** → Create a new file

**For detailed examples and patterns, see [`docs/DEVELOPMENT_PATTERNS.md`](docs/DEVELOPMENT_PATTERNS.md)**

---

## Essential Development Rules

### 1. Emoji Usage Rule (CRITICAL!)

**⚠️ NEVER use emoji variation selectors (U+FE0F / U+FE0E) in the codebase!**

**Rule:**
- ✅ **Always use base emoji characters without variation selectors**
- ❌ **Never use:** `"🗑️"` (U+1F5D1 + U+FE0F)
- ✅ **Instead use:** `"🗑"` (U+1F5D1 alone)

**Why this matters:**
- go-runewidth has a bug (#76) where variation selectors are counted as width=1 instead of width=0
- This causes misalignment in terminal width calculations
- Different terminals render emoji+VS inconsistently

**How to check:**
```bash
echo -n "🗑️" | xxd  # Look for efb88f bytes (U+FE0F)
```

**Examples:**
```go
// ❌ WRONG - Has variation selector
icon := "⚙️"  // U+2699 + U+FE0F
trash := "🗑️" // U+1F5D1 + U+FE0F

// ✅ CORRECT - Base emoji only
icon := "⚙"   // U+2699
trash := "🗑"  // U+1F5D1
```

---

### 2. Width Calculations Rule (CRITICAL!)

**Before modifying UI code, read [`docs/LESSONS_LEARNED.md`](docs/LESSONS_LEARNED.md)**

Quick rules:
- **Use `visualWidth()` not `len()`** for display text
- **Use `truncateToWidth()` not byte slicing** for breaking text
- **Wrap ALL content** except pre-rendered markdown
- **Truncate after padding** to prevent overflow

**Ignoring these rules WILL cause bugs.** Many hours have been spent fixing width calculation issues.

---

### 3. Keep main.go Minimal

**NEVER add business logic to `main.go`**

`main.go` should ONLY contain:
```go
func main() {
    p := tea.NewProgram(initialModel(), tea.WithAltScreen())
    if err := p.Start(); err != nil {
        log.Fatal(err)
    }
}
```

---

### 4. Code Organization Principles

1. **Single Responsibility** - Each file has one clear purpose
2. **DRY** - Extract common logic to helpers
3. **Separate Concerns** - UI rendering ≠ business logic
4. **Clear Naming** - File names should be self-explanatory
5. **Group Related Functions** - Keep related code together

**For detailed patterns, see [`docs/DEVELOPMENT_PATTERNS.md`](docs/DEVELOPMENT_PATTERNS.md)**

---

## Building and Installing

**⚠️ IMPORTANT: After rebuilding TFE, always update the installed binaries!**

```bash
# Option 1: Use build script (RECOMMENDED - updates both locations automatically)
./build.sh

# Option 2: Manual build and install (must update BOTH locations)
go build
cp ./tfe /home/matt/.local/bin/tfe
cp ./tfe /home/matt/bin/tfe
```

**Why this matters:**
- User has TFE installed in **TWO** locations: `~/bin/tfe` and `~/.local/bin/tfe`
- Both must be updated or user will run old version
- The `build.sh` script now updates both automatically
- Forgetting to update means user won't have latest fixes

**When to do this:**
- After any `go build` command
- After fixing bugs that need testing
- After implementing new features
- Before asking user to test installed version

**Best practice workflow:**
```bash
# 1. Run tests to verify changes
go test ./...

# 2. Build and install if tests pass
./build.sh
```

---

## Creating New Modules

If a feature is substantial enough (200+ lines, self-contained), create a new module:

1. Create new `.go` file with descriptive name
2. Keep it in `package main`
3. Document module's purpose at the top
4. Add to [`docs/MODULE_DETAILS.md`](docs/MODULE_DETAILS.md)
5. Update this document's [Module Quick Reference](#module-quick-reference)

**Template:**
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

**For detailed guidance, see [`docs/DEVELOPMENT_PATTERNS.md`](docs/DEVELOPMENT_PATTERNS.md)**

---

## Common Development Tasks

**For detailed examples, see [`docs/DEVELOPMENT_PATTERNS.md`](docs/DEVELOPMENT_PATTERNS.md)**

Quick links to common tasks:
- [Adding a keyboard shortcut](docs/DEVELOPMENT_PATTERNS.md#pattern-2-adding-a-new-keyboard-shortcut)
- [Adding a display mode](docs/DEVELOPMENT_PATTERNS.md#pattern-3-adding-a-new-display-mode)
- [Adding a file operation](docs/DEVELOPMENT_PATTERNS.md#pattern-4-adding-a-new-file-operation)
- [Adding a context menu item](docs/DEVELOPMENT_PATTERNS.md#pattern-5-adding-a-context-menu-item)
- [Adding a message type](docs/DEVELOPMENT_PATTERNS.md#pattern-6-adding-a-new-message-type)

---

## Documentation Index

**Core Documentation:**
- **[`CLAUDE.md`](CLAUDE.md)** (this file) - Architecture index
- **[`README.md`](README.md)** - Project overview & installation
- **[`PLAN.md`](PLAN.md)** - Current roadmap
- **[`CHANGELOG.md`](CHANGELOG.md)** - Recent changes
- **[`BACKLOG.md`](BACKLOG.md)** - Future ideas
- **[`HOTKEYS.md`](HOTKEYS.md)** - Keyboard shortcuts

**Detailed Guides:**
- **[`docs/MODULE_DETAILS.md`](docs/MODULE_DETAILS.md)** - Full module descriptions
- **[`docs/DEVELOPMENT_PATTERNS.md`](docs/DEVELOPMENT_PATTERNS.md)** - Detailed examples & patterns
- **[`docs/LESSONS_LEARNED.md`](docs/LESSONS_LEARNED.md)** - Critical UI/rendering lessons
- **[`docs/THREAT_MODEL.md`](docs/THREAT_MODEL.md)** - Security philosophy
- **[`docs/DOCUMENTATION_GUIDE.md`](docs/DOCUMENTATION_GUIDE.md)** - How docs are managed
- **[`docs/REFACTORING_HISTORY.md`](docs/REFACTORING_HISTORY.md)** - Architectural history

---

## Testing Strategy

### Running Tests

**⚠️ ALWAYS run tests after making code changes!**

```bash
# Run all tests (recommended)
go test ./...

# Run with verbose output (shows all test names)
go test -v ./...

# Run specific test
go test -v -run TestAddToHistory

# Run tests with coverage report
go test -cover ./...

# Quick validation: test + build
go test ./... && go build
```

### When to Run Tests

**Claude should run tests:**
1. ✅ **After fixing bugs** - Verify the fix didn't break anything
2. ✅ **After refactoring** - Confirm behavior hasn't changed
3. ✅ **After adding features** - Check integration with existing code
4. ✅ **Before completing a task** - Ensure all tests pass before reporting success

**Do NOT run tests:**
- ❌ For simple documentation changes
- ❌ For non-code file edits (README, CHANGELOG, etc.)

### Writing Tests

When adding tests:
- Create corresponding `*_test.go` files alongside each module
- Test files should mirror structure: `file_operations_test.go`, etc.
- Keep test files focused on their corresponding module
- Initialize model fields properly (see existing tests for patterns)

---

## Important Reminders

**🚨 When adding new features, always maintain this modular architecture!**

Do NOT add complex logic to `main.go`. Instead:
1. Identify which module the feature belongs to (use [decision tree](#where-does-my-code-go))
2. Add it to that module, or create a new one
3. Keep files focused and organized
4. Update [`docs/MODULE_DETAILS.md`](docs/MODULE_DETAILS.md) when creating new modules
5. Read [`docs/LESSONS_LEARNED.md`](docs/LESSONS_LEARNED.md) before touching UI code
6. **Always run tests after changes:** `go test ./...`
7. **Always rebuild and install:** `./build.sh` (or manually copy to both locations)

---

## Documentation Health Check

Check documentation sizes periodically:

```bash
wc -l CLAUDE.md README.md PLAN.md CHANGELOG.md BACKLOG.md docs/*.md
```

**Target sizes:**
- `CLAUDE.md`: < 500 lines
- `PLAN.md`: < 400 lines
- `CHANGELOG.md`: < 350 lines
- `BACKLOG.md`: < 300 lines

See [`docs/DOCUMENTATION_GUIDE.md`](docs/DOCUMENTATION_GUIDE.md) for workflow details.

---

## Summary

**TFE uses a modular architecture with strict separation of concerns.**

- Each file has ONE clear responsibility
- `main.go` is just the entry point (21 lines)
- UI rendering is separate from business logic
- Documentation is split into focused guides

**This architecture took significant effort to establish - let's maintain it!** 🏗️

For any development task:
1. Check ["When to Read What"](#quick-start-when-to-read-what) at the top
2. Follow the decision tree to find the right file
3. Read the relevant detailed documentation
4. Maintain the modular structure
5. Update docs when adding new modules

**Quick reminder:** Always read [`docs/LESSONS_LEARNED.md`](docs/LESSONS_LEARNED.md) before modifying UI code!
