# Task: Add View Toggle Emoji Buttons to Header

## Goal
Add two new clickable emoji buttons to the toolbar for quick view switching.

## New Buttons to Add

### 1. **👁️ View Mode Toggle** (Eye Emoji)
- **Function**: Cycle through the 3 display modes
- **Click behavior**: List → Detail → Tree → List (cycle)
- **Current keyboard shortcut**: F9 or 1/2/3
- **Position**: Add after existing buttons (before search icon)

### 2. **⬜ / ⬌ Pane Toggle** (Rectangle/Split Pane Emoji)
- **Function**: Toggle between single-pane and dual-pane view
- **Click behavior**: Single ↔ Dual-pane
- **Current keyboard shortcuts**: Tab or Space
- **Position**: Add after view mode toggle
- **Icon options**:
  - `⬜` (single pane) / `⬌` (dual pane) - show current state
  - Or use single icon like `◫` or `▦` that toggles

## ⚠️ CRITICAL: Update BOTH Headers!

**IMPORTANT REMINDER**: TFE has **TWO different headers** that MUST **BOTH** be updated:

1. **Single-pane/Full preview header** (`view.go` - `renderSinglePane()` and `renderFullPreview()`)
2. **Dual-pane header** (`render_preview.go` - `renderDualPane()`)

### Common mistake to avoid:
- ❌ Updating only one header and forgetting the other
- ❌ Testing in single-pane mode only, missing dual-pane bugs
- ✅ **Search for ALL toolbar rendering locations** and update each one
- ✅ Test in BOTH single-pane and dual-pane modes before committing

### Files to Check:
- `view.go` - Single-pane mode toolbar
- `render_preview.go` - Dual-pane mode toolbar (around line ~730-750)
- Also check `renderFullPreview()` if it has its own toolbar

**Pro tip:** Search for existing emoji patterns to find all locations:
```bash
grep -n "🏠\|⭐\|📝\|🔍" view.go render_preview.go
```

## Implementation Steps

### Step 1: Find Current Toolbar Code
Search for existing emoji buttons to find all toolbar locations:
```bash
grep -n "🏠" view.go render_preview.go
```

Expected to find **at least 2 locations** (single-pane and dual-pane headers).

### Step 2: Add Emoji Buttons to Toolbar Rendering

**In each header location**, add the new buttons:

```go
// Example toolbar (adjust spacing as needed):
toolbar := "[🏠] [⭐] [👁️] [⬜] [🔍] [📝]"
```

**Spacing considerations:**
- Each emoji button: `[ + emoji(2 chars) + ]` = 4 characters width
- Space between buttons: 1 character
- Update all X-position calculations when adding new buttons

### Step 3: Add Mouse Click Handlers (`update_mouse.go`)

Find the existing toolbar click handlers and add new cases:

**View Mode Toggle (👁️):**
```go
// View mode toggle button [👁️] (X=XX-YY: [ + emoji(2) + ])
if msg.X >= XX && msg.X <= YY {
    // Cycle through display modes
    if m.displayMode == modeList {
        m.displayMode = modeDetail
    } else if m.displayMode == modeDetail {
        m.displayMode = modeTree
    } else {
        m.displayMode = modeList
    }
    return m, nil
}
```

**Pane Toggle (⬜/⬌):**
```go
// Pane toggle button [⬜] (X=XX-YY: [ + emoji(2) + ])
if msg.X >= XX && msg.X <= YY {
    // Toggle between single and dual-pane
    if m.viewMode == viewDualPane {
        m.viewMode = viewSinglePane
    } else {
        m.viewMode = viewDualPane
    }
    m.calculateLayout()
    m.populatePreviewCache() // Refresh cache with new layout
    return m, nil
}
```

### Step 4: Calculate Correct X Positions

**Current toolbar (before changes):**
```
[🏠] [⭐] [📝] [🔍] [🗑️]
 0-4  5-9  10-14 15-19 20-24
```

**New toolbar (after adding buttons):**
```
[🏠] [⭐] [👁️] [⬜] [🔍] [📝] [🗑️]
 0-4  5-9  10-14 15-19 20-24 25-29 30-34
```

**Action items:**
1. Add the new buttons in the correct position
2. Update ALL X-position ranges for existing buttons
3. Update comments to reflect new positions

### Step 5: Test in BOTH Modes

**Testing checklist:**
- [ ] Single-pane mode: Toolbar shows all buttons
- [ ] Single-pane mode: All buttons clickable at correct X positions
- [ ] Dual-pane mode: Toolbar shows all buttons
- [ ] Dual-pane mode: All buttons clickable at correct X positions
- [ ] Full preview mode: Toolbar correct (if applicable)
- [ ] 👁️ Eye button cycles List → Detail → Tree correctly
- [ ] ⬜ Pane button toggles single ↔ dual-pane correctly
- [ ] No visual overlap or spacing issues
- [ ] All existing buttons still work with updated X positions

## Reference Files

**Existing toolbar code locations:**
- `view.go` - Lines with emoji rendering (search for `🏠`)
- `render_preview.go` - Dual-pane header (around line 730-750)
- `update_mouse.go` - All mouse click handlers for toolbar buttons

**Existing examples to copy from:**
- Home button click handler (navigate to home)
- Favorites button click handler (toggle favorites)
- Search button click handler (launch fuzzy search)

## Success Criteria

- [ ] 👁️ Eye button appears in **ALL** headers (single/dual/full)
- [ ] 👁️ Eye button cycles List → Detail → Tree when clicked
- [ ] ⬜ Pane button appears in **ALL** headers
- [ ] ⬜ Pane button toggles single ↔ dual-pane when clicked
- [ ] Tested in single-pane mode: ✅
- [ ] Tested in dual-pane mode: ✅
- [ ] All existing buttons still work with updated X positions
- [ ] No visual overlap or spacing issues

## Common Pitfalls

1. **Forgetting dual-pane header** - ALWAYS update both!
2. **Wrong X positions** - Recalculate after adding buttons
3. **Not testing dual-pane** - Must test in both modes
4. **Emoji width assumptions** - Emojis are 2 chars, `[` and `]` are 1 char each
5. **Missing calculateLayout()** - Call when toggling pane mode

## Tips

- Use `grep` to find ALL instances of toolbar rendering
- Copy the pattern from existing emoji buttons for consistency
- Remember to update X-position ranges when adding new buttons
- Test thoroughly in both single and dual-pane modes before committing
- Add helpful status messages when toggling (optional but nice UX)

## Optional Enhancements

If you want to go the extra mile:

1. **Visual feedback**: Show status message when toggling
   ```go
   m.setStatusMessage("View mode: Detail", false)
   m.setStatusMessage("Dual-pane mode enabled", false)
   ```

2. **Dynamic emoji**: Show different emoji based on current state
   - Single pane: `⬜`
   - Dual pane: `⬌`

3. **Hover text**: Add tooltip-style help when cursor near button (advanced)

---

**Note**: This task adds quality-of-life improvements for mouse users who prefer clicking over keyboard shortcuts!

**Estimated time**: 30-45 minutes

**Files to modify**:
- `view.go` (toolbar rendering)
- `render_preview.go` (dual-pane toolbar)
- `update_mouse.go` (click handlers)
