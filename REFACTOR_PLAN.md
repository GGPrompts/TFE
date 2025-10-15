# TFE Refactoring Plan

## Current Status

As of the last commit, `main.go` has grown to **1648 lines**. While this is manageable for a single-file project, splitting it into multiple files will improve:

1. **Maintainability**: Easier to navigate and understand specific components
2. **Testability**: Separate files make unit testing easier
3. **Collaboration**: Multiple developers can work on different files simultaneously
4. **Code Organization**: Logical grouping of related functionality

## Proposed File Structure

```
tfe/
├── main.go                 # Entry point, program initialization (50-75 lines)
├── model.go                # Model struct and initialization (100-150 lines)
├── update.go               # Update function and message handling (400-500 lines)
├── view.go                 # View dispatching and main rendering logic (50-75 lines)
├── render_file_list.go     # File list rendering (List, Grid, Detail, Tree) (350-400 lines)
├── render_preview.go       # Preview rendering (dual-pane, full-screen) (300-350 lines)
├── file_operations.go      # File loading, icon mapping, formatting (200-250 lines)
├── editor.go               # External editor integration and clipboard (100-120 lines)
├── styles.go               # Lipgloss styles and color definitions (50-75 lines)
├── types.go                # Type definitions (displayMode, viewMode, etc.) (150-180 lines)
└── utils.go                # Utility functions (ANSI stripping, etc.) (50-75 lines)
```

## Refactoring Strategy

### Phase 1: Extract Types and Constants
**Goal**: Move all type definitions, constants, and enums to `types.go`

**Files to create**:
- `types.go` - All struct definitions, enums (displayMode, viewMode, paneType)

**Benefits**: Clean separation of data structures from logic

### Phase 2: Extract Styles
**Goal**: Move all Lipgloss style definitions to `styles.go`

**Files to create**:
- `styles.go` - All `lipgloss.NewStyle()` definitions

**Benefits**: Centralized theming, easier to customize colors

### Phase 3: Extract File Operations
**Goal**: Move file system operations and utilities to dedicated files

**Files to create**:
- `file_operations.go` - loadFiles(), getFileIcon(), formatFileSize(), formatModTime(), isBinaryFile()
- `editor.go` - openEditor(), editorAvailable(), getAvailableEditor(), copyToClipboard()

**Benefits**: Clear separation of concerns, easier to add new file operations

### Phase 4: Extract Rendering Logic
**Goal**: Split rendering functions into logical groups

**Files to create**:
- `render_file_list.go` - renderListView(), renderGridView(), renderDetailView(), renderTreeView()
- `render_preview.go` - renderPreview(), renderFullPreview(), renderDualPane()
- `view.go` - View() dispatcher and renderSinglePane()

**Benefits**: Each rendering mode is self-contained and testable

### Phase 5: Extract Update Logic
**Goal**: Move Update() function to its own file

**Files to create**:
- `update.go` - Update() function and all message handling

**Benefits**: Isolate event handling logic, easier to add new keyboard/mouse commands

### Phase 6: Extract Model and Utilities
**Goal**: Organize model initialization and utility functions

**Files to create**:
- `model.go` - initialModel(), calculateGridLayout(), calculateLayout(), loadPreview()
- `utils.go` - stripAnsi(), visibleWidth(), truncateOrPad()

**Benefits**: Model initialization and layout calculations are grouped together

### Phase 7: Simplify main.go
**Goal**: Reduce main.go to just program entry point

**Resulting main.go**:
```go
package main

import (
	"fmt"
	"os"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
```

## Testing Strategy

After each refactoring phase:

1. **Build**: Ensure `go build` succeeds
2. **Run**: Test basic functionality (navigate, preview, mouse, keyboard)
3. **Compare**: Verify behavior matches pre-refactor version
4. **Commit**: Create a commit for each phase

## Important Notes

- **Keep all code in the same package** (`package main`) to avoid export issues
- **Maintain exact same functionality** - this is pure refactoring, no feature changes
- **Test thoroughly** after each phase to catch any regressions early
- **Use git** to track changes and enable easy rollback if needed

## Success Criteria

✅ All code split into logical files (no single file > 500 lines)
✅ Program builds without errors
✅ All functionality works identically to before refactoring
✅ Code is more readable and maintainable
✅ Each file has a clear, single responsibility

## Next Steps

When ready to begin refactoring:

1. Create a new branch: `git checkout -b refactor-split-files`
2. Start with Phase 1 (Extract Types and Constants)
3. Test after each phase
4. Commit frequently with descriptive messages
5. Merge back to main when complete

## Estimated Time

- **Phase 1-2**: 30-45 minutes (types and styles)
- **Phase 3**: 30-45 minutes (file operations and editor)
- **Phase 4**: 60-90 minutes (rendering logic)
- **Phase 5**: 45-60 minutes (update logic)
- **Phase 6-7**: 30-45 minutes (model and main.go)

**Total**: 3.5-5 hours of focused refactoring work

---

*This refactoring plan was generated to prepare TFE for easier maintenance and future feature development.*
