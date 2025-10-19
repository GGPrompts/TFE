# Implementation Guide: Mouse Toggle in Preview Mode

## Goal

Add a keyboard toggle (`m` key) to enable/disable mouse in preview mode, making text selection easier while preserving scroll functionality.

---

## Current Issue

When viewing files in preview mode (F3/Enter):
- Shift+Click works for text selection, but it's wonky (selection tracks screen position, not text)
- Scrolling changes content but selection stays at same row
- Users need a better way to select/copy text from previews

---

## Solution

Add `m` key to toggle mouse on/off in preview modes:
- **Mouse ON** (default): Wheel scrolling works, text selection is wonky
- **Mouse OFF**: Terminal handles selection properly, no scrolling

---

## Implementation Steps

### Step 1: Add Mouse State to Model

**File:** `types.go`

Find the `model` struct and add a new field:

```go
type model struct {
    // ... existing fields ...

    // Mouse state for preview mode
    previewMouseEnabled bool  // Whether mouse is enabled in preview mode (default: true)

    // ... rest of fields ...
}
```

**File:** `model.go`

In `initialModel()`, initialize the field:

```go
func initialModel() model {
    // ... existing initialization ...

    m := model{
        // ... existing fields ...
        previewMouseEnabled: true,  // Mouse enabled by default
        // ... rest of fields ...
    }

    return m
}
```

---

### Step 2: Add Keyboard Handler for 'm' Key

**File:** `update_keyboard.go`

Find the keyboard handler for full preview mode (`if m.viewMode == viewFullPreview {`).

Add this case **before** the scrolling keys (around line 120-140):

```go
case "m", "M":
    // Toggle mouse in preview mode
    m.previewMouseEnabled = !m.previewMouseEnabled

    if m.previewMouseEnabled {
        m.setStatusMessage("🖱️  Mouse enabled (scrolling works, text selection wonky)", false)
        return m, tea.EnableMouseCellMotion
    } else {
        m.setStatusMessage("⌨️  Mouse disabled (text selection works, use arrows to scroll)", false)
        return m, tea.DisableMouse
    }
```

**Also handle in dual-pane mode** (find `if m.viewMode == viewDualPane {`):

Add similar case in the dual-pane keyboard handler:

```go
case "m", "M":
    // Toggle mouse in dual-pane preview
    m.previewMouseEnabled = !m.previewMouseEnabled

    if m.previewMouseEnabled {
        m.setStatusMessage("🖱️  Mouse enabled (scrolling works, text selection wonky)", false)
        return m, tea.EnableMouseCellMotion
    } else {
        m.setStatusMessage("⌨️  Mouse disabled (text selection works, use arrows to scroll)", false)
        return m, tea.DisableMouse
    }
```

---

### Step 3: Update Preview Help Text

**File:** `render_preview.go`

Find the `renderFullPreview()` function (around line 630).

Look for the help text section at the bottom (around line 690-698):

```go
// Help text
s.WriteString("\n")
helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).PaddingLeft(2)

// Show different F5 text if viewing a prompt (with or without fillable fields)
f5Text := "copy path"
if m.preview.isPrompt || (m.inputFieldsActive && len(m.promptInputFields) > 0) {
    f5Text = "copy rendered prompt"
}
helpText := fmt.Sprintf("↑/↓: scroll • PgUp/PgDown: page • F4: edit • F5: %s • Esc: close • F10: quit", f5Text)
s.WriteString(helpStyle.Render(helpText))
```

**Replace with:**

```go
// Help text
s.WriteString("\n")
helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).PaddingLeft(2)

// Show different F5 text if viewing a prompt (with or without fillable fields)
f5Text := "copy path"
if m.preview.isPrompt || (m.inputFieldsActive && len(m.promptInputFields) > 0) {
    f5Text = "copy rendered prompt"
}

// Mouse toggle indicator
mouseStatus := "🖱️ ON"
if !m.previewMouseEnabled {
    mouseStatus = "⌨️  OFF"
}

helpText := fmt.Sprintf("↑/↓: scroll • m: mouse %s • F4: edit • F5: %s • Esc: close", mouseStatus, f5Text)
s.WriteString(helpStyle.Render(helpText))
```

---

### Step 4: Update Dual-Pane Help Text

**File:** `render_preview.go`

Find the `renderDualPane()` function (around line 705).

Look for the status bar section at the bottom (search for "Status bar").

The current status shows focus state. Update it to also show mouse toggle:

Find this section (around line 940-970):

```go
// Status bar
var statusLeft, statusRight string

if m.focusPane == leftPane {
    statusLeft = selectedStyle.Render("● File List")
    statusRight = "○ Preview"
} else {
    statusLeft = "○ File List"
    statusRight = selectedStyle.Render("● Preview")
}
```

**Add mouse indicator after the status variables:**

```go
// Status bar
var statusLeft, statusRight string

if m.focusPane == leftPane {
    statusLeft = selectedStyle.Render("● File List")
    statusRight = "○ Preview"
} else {
    statusLeft = "○ File List"
    statusRight = selectedStyle.Render("● Preview")
}

// Mouse toggle indicator
mouseStatus := "🖱️ ON"
if !m.previewMouseEnabled {
    mouseStatus = "⌨️  OFF"
}
```

Then find where the status bar is rendered and add the mouse indicator:

Look for `statusText :=` or similar (around line 970-990).

**Update the status text to include mouse toggle:**

```go
// Build status text
statusText := fmt.Sprintf("%s | %s | Tab: switch • Space: toggle • m: mouse %s",
    statusLeft, statusRight, mouseStatus)
```

---

### Step 5: Fix Mouse Re-enabling After Exit

**File:** `update_keyboard.go`

Find where preview mode exits (search for `m.viewMode = viewSinglePane`).

Make sure mouse is re-enabled when exiting preview:

Around line 126-134:

```go
case "f10", "ctrl+c", "esc":
    // Exit preview mode (F10 replaces q)
    m.viewMode = viewSinglePane
    m.calculateLayout()
    m.populatePreviewCache() // Refresh cache with new width
    // Clear any stray command input that might have captured terminal responses
    m.commandInput = ""
    m.commandFocused = false
    // Re-enable mouse when exiting preview
    m.previewMouseEnabled = true  // ← ADD THIS LINE
    return m, tea.EnableMouseCellMotion
```

---

### Step 6: Update HOTKEYS.md Documentation

**File:** `HOTKEYS.md`

Find the "Preview & Full-Screen Mode" section (around line 44).

Add the mouse toggle key:

```markdown
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
| **m** / **M** | Toggle mouse on/off (for better text selection) |
| **Mouse Wheel** | Scroll preview (when mouse is enabled) |
```

Also add a tip at the bottom of the Tips section:

```markdown
18. **Text Selection in Preview:** Press **m** to disable mouse for better text selection, or use **Shift+Click** (wonky but works), or press **F4** to edit in Micro for easier copying
```

---

## Testing Checklist

After implementation:

### Full Preview Mode (F3/Enter on file)
- [ ] Mouse enabled by default (can scroll with wheel)
- [ ] Press `m` → Mouse disabled → Status message shows "Mouse disabled"
- [ ] Help text shows "m: mouse ⌨️  OFF"
- [ ] Can select text normally with mouse (no wonky behavior)
- [ ] Arrow keys still work for scrolling
- [ ] Press `m` again → Mouse re-enabled → Status message shows "Mouse enabled"
- [ ] Help text shows "m: mouse 🖱️ ON"
- [ ] Mouse wheel scrolling works again
- [ ] Press Esc → Returns to file list → Mouse is enabled

### Dual-Pane Mode (Tab)
- [ ] Mouse enabled by default
- [ ] Press `m` → Mouse disabled → Status bar shows "m: mouse ⌨️  OFF"
- [ ] Can select text from preview pane
- [ ] Press `m` again → Mouse enabled → Status bar shows "m: mouse 🖱️ ON"
- [ ] Mouse wheel scrolling works in preview pane
- [ ] Press Space → Exit dual-pane → Mouse still enabled

### Edge Cases
- [ ] Toggle works in markdown preview
- [ ] Toggle works in text file preview
- [ ] Toggle persists when scrolling
- [ ] Status messages are clear and helpful

---

## Expected Behavior

**Before:**
- User has to use Shift+Click which is wonky
- Scrolling changes what's selected (screen position)
- Frustrating text selection experience

**After:**
- User presses `m` to disable mouse
- Full terminal text selection works perfectly
- Arrow keys ↑/↓ to scroll while selecting
- Press `m` again to re-enable scrolling
- Help text shows current state
- Clear status messages guide the user

---

## Build and Test

```bash
# Build TFE
go build -o tfe

# Test the feature
./tfe

# Navigate to a file
# Press Enter or F3 to preview
# Press 'm' to toggle mouse
# Try selecting text with mouse
# Try scrolling with arrow keys
# Press 'm' again to re-enable mouse wheel
```

---

## Files Modified

1. ✅ `types.go` - Add previewMouseEnabled field
2. ✅ `model.go` - Initialize previewMouseEnabled to true
3. ✅ `update_keyboard.go` - Add 'm' key handler (full preview + dual-pane)
4. ✅ `update_keyboard.go` - Reset mouse when exiting preview
5. ✅ `render_preview.go` - Update help text in renderFullPreview()
6. ✅ `render_preview.go` - Update status bar in renderDualPane()
7. ✅ `HOTKEYS.md` - Document the new 'm' key

---

## Time Estimate

- Implementation: 15-20 minutes
- Testing: 5-10 minutes
- **Total: ~30 minutes**

---

## After This Feature: Proceed to v1.0 Launch Tasks

Once mouse toggle is complete and tested, move on to the **v1.0 critical path:**

### Next Steps (in order):

1. **Implement Copy Files** (2-3 hours)
   - See `docs/NEXT_SESSION.md` for full implementation guide
   - Context menu → "Copy to..." → Dialog → Copy operation
   - Handle files and directories (recursive)

2. **Implement Rename Files** (1-2 hours)
   - See `docs/NEXT_SESSION.md` for full implementation guide
   - Context menu → "Rename..." → Dialog → Rename operation
   - Pre-fill current name, validate input

3. **Testing & Polish** (2-3 hours)
   - Test all features thoroughly
   - Fix any bugs discovered
   - Ensure smooth UX

4. **Documentation for Launch** (2-3 hours)
   - Update README.md with all features
   - Create GIF demos
   - Write installation instructions
   - Feature comparison table

5. **Build & Release** (2-3 hours)
   - Build binaries for Linux/macOS
   - Create GitHub release v1.0.0
   - Write release notes

6. **Marketing** (1-2 hours)
   - Reddit posts (r/linux, r/commandline, r/ClaudeAI)
   - Hacker News submission
   - Tweet/announcement

**Total time to v1.0 launch:** ~10-16 hours after mouse toggle is done

---

**Priority:** Medium (nice QoL feature, but not blocking v1.0)
**Difficulty:** Easy
**Time:** 30 minutes
**Status:** Ready to implement
