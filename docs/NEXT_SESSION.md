# Next Session: Remove Grid View Display Mode

## Context

Grid view (F2) doesn't display well on any monitor size and isn't useful. Remove it completely and simplify the display mode system to just:
- **List View** (F1) - Simple vertical list
- **Detail View** (F4) - Table with columns (default) ✨ Now with zebra striping!
- **Tree View** (F5) - Hierarchical tree

## Tasks

### 1. Remove Grid View Enum and Logic

**File: `types.go`**
```go
// Find this:
const (
    modeList displayMode = iota
    modeGrid    // ← DELETE THIS LINE
    modeDetail
    modeTree
)

// Replace with:
const (
    modeList displayMode = iota
    modeDetail
    modeTree
)
```

- Remove `modeGrid` from the `displayMode` enum
- Update any comments referencing 4 display modes → 3 modes
- Remove `gridColumns int` field from model struct (if it exists)

**File: `render_file_list.go`**
- Delete the entire `renderGridView()` function (~150 lines)
- Remove grid-related imports if they become unused

**File: `model.go`**
- Delete `calculateGridLayout()` function if it exists (~30 lines)

### 2. Update Display Mode Cycling

**File: `update_keyboard.go`**

Find and update the F1/F2 key handlers:

```go
case "f1":
    // OLD: Cycles through 4 modes
    m.displayMode = (m.displayMode + 1) % 4

    // NEW: Cycles through 3 modes
    m.displayMode = (m.displayMode + 1) % 3

case "f2":
    // REMOVE: No longer cycles to grid view
    // OPTION 1: Remove F2 handler entirely
    // OPTION 2: Reassign to toggle hidden files
    // OPTION 3: Quick toggle between List ↔ Detail
```

**Suggested F2 Reassignment:**
```go
case "f2":
    // Quick toggle between List and Detail (most common views)
    if m.displayMode == modeList {
        m.displayMode = modeDetail
    } else {
        m.displayMode = modeList
    }
```

### 3. Update View Rendering Switch

**File: `view.go` or dispatch location**

Find the switch statement that calls view renderers:

```go
switch m.displayMode {
case modeList:
    return m.renderListView(maxVisible)
case modeGrid:    // ← DELETE THIS CASE
    return m.renderGridView(maxVisible)
case modeDetail:
    return m.renderDetailView(maxVisible)
case modeTree:
    return m.renderTreeView(maxVisible)
default:
    return m.renderDetailView(maxVisible)  // Ensure safe default
}
```

### 4. Verify Dual-Pane Compatibility

**File: `helpers.go`**

Check `isDualPaneCompatible()` - should still work correctly:
```go
func (m model) isDualPaneCompatible() bool {
    return m.displayMode == modeList || m.displayMode == modeTree
    // modeDetail incompatible (needs full width for columns)
}
```

No changes needed here - just verify it still works after grid removal.

### 5. Update Documentation

**File: `HOTKEYS.md`**
- Remove F2 grid view reference from F-keys table
- Update display mode section:
  ```markdown
  ## Display Modes (3 total)

  - F1: List View - Simple vertical list
  - F2: (Available) or Quick toggle List ↔ Detail
  - F4: Detail View - Table with columns (default)
  - F5: Tree View - Hierarchical with folders
  ```

**File: `README.md`**
- Update feature list: "4 display modes" → "3 display modes"
- Remove grid view from any screenshots/descriptions
- Update Quick Start if it mentions F2 grid view

**File: `CLAUDE.md`**
- Update Module Responsibilities if it mentions grid view
- Update any examples showing display mode counts

### 6. Clean Up Tests

**File: `helpers_test.go`** (check if exists)
- Remove any `modeGrid` test cases
- Update `TestIsDualPaneCompatible` if it tests grid mode
- Update any tests that enumerate all display modes (4 → 3)

### 7. Search for All References

Before starting, run these commands to find all grid references:

```bash
# Find all grid code references
grep -rn "modeGrid" --include="*.go"
grep -rn "renderGridView" --include="*.go"
grep -rn "calculateGridLayout" --include="*.go"
grep -rn "gridColumns" --include="*.go"

# Find F2 key handler
grep -rn '"f2"' --include="*.go"

# Find display mode switches
grep -rn "switch.*displayMode" --include="*.go"
grep -rn "case modeGrid" --include="*.go"

# Find documentation references
grep -rn "grid" --include="*.md" -i
grep -rn "F2" --include="*.md"
```

## Testing Checklist

After removal:
- [ ] Build succeeds: `go build -o tfe .`
- [ ] All tests pass: `make test`
- [ ] All 169 tests still passing
- [ ] F1 cycles: List → Detail → Tree → List (no grid)
- [ ] F4 goes directly to Detail view
- [ ] F5 goes directly to Tree view
- [ ] Detail view is default on startup
- [ ] Detail view shows zebra striping ✨
- [ ] Dual-pane mode works with List and Tree only
- [ ] No grid references in UI or error messages
- [ ] HOTKEYS.md updated
- [ ] README.md updated

## Expected Impact

**Lines Removed:** ~200-300 lines total
- `renderGridView()`: ~150 lines
- `calculateGridLayout()`: ~30 lines
- Enum/constants/tests: ~20-50 lines
- Documentation cleanup: ~10-20 lines

**Files Modified:** 6-8 files
1. `types.go` - Remove enum value, gridColumns field
2. `render_file_list.go` - Delete renderGridView()
3. `update_keyboard.go` - Update F1/F2 handlers
4. `view.go` - Remove case from switch
5. `model.go` - Delete calculateGridLayout() if exists
6. `HOTKEYS.md` - Update F-keys, remove grid references
7. `README.md` - Update mode count, remove grid mentions
8. `CLAUDE.md` - Update architecture docs

**Benefits:**
- ✅ Cleaner codebase (fewer modes to maintain)
- ✅ Simpler UX (3 clear, useful view options)
- ✅ Less confusion about when to use grid
- ✅ More keyboard shortcuts available (F2 freed up)
- ✅ Easier to explain and document
- ✅ Focus on the modes that actually work well

## Optional: F2 Key Reassignment Ideas

**Option 1: Quick Toggle List ↔ Detail** (Recommended)
```go
case "f2":
    // Toggle between two most common views
    if m.displayMode == modeList {
        m.displayMode = modeDetail
    } else {
        m.displayMode = modeList
    }
```
Benefits: Fast switching between compact (List) and detailed (Detail)

**Option 2: Toggle Hidden Files**
```go
case "f2":
    m.showHidden = !m.showHidden
    m.loadFiles()
```
Benefits: Currently no hotkey for this common operation

**Option 3: Leave Unassigned**
Keep F2 available for future features. Document as "Available for future use"

## Git Commit Message Template

```
feat: Remove grid view display mode

Grid view didn't display well on any monitor size and wasn't useful
in practice. Simplified display modes to three focused options:

- List View (F1) - Compact vertical list
- Detail View (F4) - Table with columns & zebra striping (default)
- Tree View (F5) - Hierarchical folder navigation

Changes:
- Removed modeGrid enum and renderGridView() function
- Removed calculateGridLayout() helper function
- Removed gridColumns field from model struct
- Updated display mode cycling to skip grid (mod 3 instead of 4)
- Updated documentation (HOTKEYS.md, README.md, CLAUDE.md)
- F2 key reassigned to [quick toggle List↔Detail / hidden files / available]

This reduces code complexity by ~250 lines and improves UX clarity.
Users now have 3 distinct, well-working view modes instead of 4 with
one problematic option.
```

## Success Criteria

✅ Code compiles without errors
✅ All 169 tests pass
✅ No references to "grid" in code or docs (except git history)
✅ F1 cycles through exactly 3 modes smoothly
✅ F2 does something useful or is clearly unassigned
✅ Detail view with zebra striping works perfectly
✅ Dual-pane compatibility unaffected

---

**Priority:** Medium (code cleanup, improves maintainability)
**Estimated Time:** 30-45 minutes
**Difficulty:** Easy (straightforward removal, minimal refactoring)
**Risk:** Low (grid view rarely used, well-isolated code)

## Notes

- Keep List, Detail, and Tree views - they're all useful and distinct
- Detail view is the default and most feature-rich (now with zebra striping!)
- Tree view is unique for hierarchical browsing with expansion
- List view is great for compact, fast navigation
- This cleanup makes TFE easier to explain and maintain
- F2 freed up for a more useful feature

---

**Created:** 2025-10-19
**Status:** Ready to implement
**Session Type:** Quick cleanup task (< 1 hour)
