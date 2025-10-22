# Fix: Edit Mode Typing Conflict in Dual-Pane

**Created:** 2025-10-22
**Priority:** High (blocks core prompt editing feature)

---

## Problem Description

When editing prompt variables in dual-pane mode, keyboard input is intercepted by app hotkeys instead of being sent to variable input.

### Current Behavior

| Mode | Edit Mode Works? | Details |
|------|------------------|---------|
| **Fullscreen (F10)** | ✅ YES | Typing goes to variable input seamlessly |
| **Dual-pane** | ❌ NO | App hotkeys take precedence over typing |

**Bug reproduction:**
1. Launch TFE in dual-pane mode
2. Enable prompts filter (F6 or 📝 button)
3. Select `code-review.prompty` or `context-analyzer.prompty`
4. Focus right pane (Tab or click)
5. Press `E` to enter edit mode
6. Try typing "variable" → Instead triggers: V (view), A (no-op), R (no-op), I (no-op), etc.
7. Try typing "description" → D triggers display mode change

**Expected:** All typing should go to variable input when edit mode is active.

---

## Root Cause Hypothesis

`update_keyboard.go` likely processes hotkeys BEFORE checking `m.promptEditMode` in dual-pane, but checks edit mode FIRST in fullscreen.

**Likely issue:**
```go
// WRONG order (suspected current code)
func (m model) handleKeyEvent(msg tea.KeyMsg) (model, tea.Cmd) {
    // Process hotkeys first
    switch msg.String() {
    case "v":
        // Toggle view mode
    case "d":
        // Toggle display mode
    // ... more hotkeys
    }

    // Check edit mode AFTER hotkeys (too late!)
    if m.promptEditMode {
        // Handle typing
    }
}
```

**Correct order:**
```go
// RIGHT order (needed fix)
func (m model) handleKeyEvent(msg tea.KeyMsg) (model, tea.Cmd) {
    // Check edit mode FIRST (highest priority)
    if m.promptEditMode {
        return handlePromptEditMode(msg)  // Isolated handling
    }

    // Then process hotkeys (only if NOT in edit mode)
    switch msg.String() {
    case "v", "d", etc.
    }
}
```

---

## Investigation Plan

### 1. Read `update_keyboard.go`

Search for these key sections:

```bash
# Find where promptEditMode is checked
grep -n "promptEditMode" update_keyboard.go

# Find where hotkeys are processed
grep -n "case \"v\":" update_keyboard.go
grep -n "case \"d\":" update_keyboard.go

# Find view mode routing
grep -n "viewFullPreview" update_keyboard.go
grep -n "viewDualPane" update_keyboard.go
```

**Questions to answer:**
- Where in the control flow is `m.promptEditMode` checked?
- Is there a difference between fullscreen and dual-pane keyboard handling?
- Which hotkeys are processed before the edit mode check?
- Is there a `handlePromptEditMode()` function or is it inline?

### 2. Compare Fullscreen vs Dual-Pane

Find the divergence point:

```go
// Look for branches like this
if m.viewMode == viewFullPreview {
    // Does this check promptEditMode early?
} else {
    // Does this NOT check promptEditMode early?
}
```

### 3. Identify All Conflicting Hotkeys

Keys that currently conflict when edit mode is active:
- `V` - Toggle view mode
- `D` - Toggle display mode
- `E` - Toggle edit mode (should exit, not toggle)
- `/` - Search mode
- `:` - Command mode
- `T` - Toggle tree expand
- `F` - Toggle favorites
- Any single-letter keys users might need to type

---

## Required Implementation

### Solution 1: Early Guard (Recommended)

Add an early return at the TOP of `handleKeyEvent()`:

```go
func (m model) handleKeyEvent(msg tea.KeyMsg) (model, tea.Cmd) {
    // PRIORITY 1: Edit mode (works in ALL view modes)
    if m.promptEditMode {
        return m.handlePromptEditKeys(msg)
    }

    // Continue with normal hotkey processing...
}

func (m model) handlePromptEditKeys(msg tea.KeyMsg) (model, tea.Cmd) {
    switch msg.String() {
    case "f3":
        // File picker for FILE/PATH variables
        return m.launchFilePicker()

    case "tab":
        // Next variable
        m.focusedVariableIndex++
        if m.focusedVariableIndex >= len(m.preview.promptTemplate.variables) {
            m.focusedVariableIndex = 0
        }

    case "shift+tab":
        // Previous variable
        m.focusedVariableIndex--
        if m.focusedVariableIndex < 0 {
            m.focusedVariableIndex = len(m.preview.promptTemplate.variables) - 1
        }

    case "enter":
        // Save and exit edit mode
        m.promptEditMode = false
        m.statusMessage = "Variables saved"
        m.statusTime = time.Now()

    case "esc":
        // Exit without saving
        m.promptEditMode = false
        m.statusMessage = "Edit mode cancelled"
        m.statusTime = time.Now()

    default:
        // ALL OTHER KEYS: typing into current variable
        runes := []rune(msg.String())
        if len(runes) == 1 {
            // Single character typed
            varName := m.preview.promptTemplate.variables[m.focusedVariableIndex]
            currentValue := m.filledVariables[varName]
            m.filledVariables[varName] = currentValue + string(runes[0])
        }
    }

    return m, nil
}
```

### Solution 2: View Mode Check (Alternative)

If the code already has separate paths for fullscreen vs dual-pane, ensure both check `m.promptEditMode` at the same priority level.

---

## Files to Modify

Based on CLAUDE.md architecture:

### Primary File
- **`update_keyboard.go`** (lines ~1-800)
  - `handleKeyEvent()` - Main entry point
  - Add `handlePromptEditKeys()` helper (new function)
  - Ensure edit mode check happens BEFORE all hotkey processing

### Reference Files (read-only)
- **`render_preview.go`** - To understand edit mode rendering
- **`types.go`** - To verify model fields:
  - `promptEditMode bool`
  - `focusedVariableIndex int`
  - `filledVariables map[string]string`

---

## Testing Checklist

After implementing the fix:

### Basic Typing Test
1. Launch TFE in dual-pane mode
2. Enable prompts filter (F6)
3. Navigate to `code-review.prompty`
4. Focus right pane (Tab)
5. Press `E` to enter edit mode
6. Type: `"description"` → Should type all letters, not trigger hotkeys ✅
7. Type: `"variable_name"` → Should type all letters ✅
8. Type: `"test/path/file.txt"` → Should type all characters ✅

### Special Keys Test
9. Press Tab → Should move to next variable ✅
10. Press Shift+Tab → Should move to previous variable ✅
11. Press F3 → Should open file picker (for FILE/PATH variables) ✅
12. Press Enter → Should save and exit edit mode ✅
13. Press Esc → Should cancel and exit edit mode ✅

### View Modes Test
14. Repeat steps 1-8 in fullscreen mode (F10) → Should still work ✅
15. Repeat steps 1-8 in single-pane mode → Should work ✅

### Regression Test
16. Exit edit mode (Esc)
17. Press `V` → Should toggle view mode (normal hotkey behavior) ✅
18. Press `D` → Should toggle display mode ✅
19. Press `/` → Should enter search mode ✅

---

## Test Files

Use these prompt files from `~/.prompts/`:
- **`code-review.prompty`** - Simple, 2 variables: `{{file}}`, `{{project}}`
- **`context-analyzer.prompty`** - Complex, 3 variables: `{{CONTEXT_PASTE}}`, `{{project}}`, `{{DATE}}`

---

## Success Criteria

- ✅ Can type ANY characters in edit mode (dual-pane)
- ✅ No hotkey interference when typing variable values
- ✅ F3 file picker still works for FILE/PATH variables
- ✅ Tab/Shift+Tab navigation between variables works
- ✅ Enter saves, Esc cancels
- ✅ Fullscreen mode continues to work (no regression)
- ✅ Normal hotkeys work when NOT in edit mode (no regression)

---

## Related Context

**Why this matters:**
- Edit mode is a flagship feature (inline variable editing)
- Users need to fill custom variables: `{{author}}`, `{{description}}`, `{{task}}`, etc.
- Currently unusable in the primary interface mode (dual-pane)

**Related commits:**
- `cbd2c92` - feat: Refactor prompts to use inline variable editing
- `df61efd` - fix: Prevent prompt preview height overflow in dual-pane mode

**Architecture (CLAUDE.md):**
> **Module 5a: `update_keyboard.go` - Keyboard Event Handling**
> Purpose: All keyboard input processing
> When to extend: Add new keyboard shortcuts or key bindings here

**Edit mode documentation:**
- `docs/INLINE_EDITING_REFACTOR.md` - Full feature specification

---

## Prompt for Claude Code

```
I need to fix a keyboard handling bug in prompt edit mode.

PROBLEM:
- In fullscreen mode (F10): Edit mode works perfectly - typing goes to variable input ✅
- In dual-pane mode: Edit mode activated but typing triggers app hotkeys instead ❌

When I press E to enter edit mode on a prompt preview in dual-pane, I can't type.
Keys like V, D, /, etc. trigger view changes instead of typing characters.

GOAL:
When m.promptEditMode == true (in ANY view mode), disable all app hotkeys except F3
and route all typing to variable input.

INVESTIGATION:
1. Read update_keyboard.go:handleKeyEvent() - find where promptEditMode is checked
2. Compare keyboard handling between viewFullPreview and viewDualPane modes
3. Identify where hotkeys (V, D, /, E) are processed relative to edit mode check

FIX:
Add early guard at top of handleKeyEvent() to give edit mode highest priority:
- Check m.promptEditMode FIRST (before all hotkey processing)
- Route to isolated handlePromptEditKeys() function
- Allow only: F3 (file picker), Tab/Shift+Tab (navigate), Enter (save), Esc (cancel)
- All other keys: typing input

EXCEPTION:
F3 file picker must still work for FILE/PATH variable types.

TEST:
1. Dual-pane mode, prompts filter on
2. Select code-review.prompty or context-analyzer.prompty
3. Focus right pane, press E to enter edit mode
4. Type "description" - should type all letters, not trigger D (display mode) hotkey
5. Type "variable" - should type all letters, not trigger V (view mode) hotkey
6. Press F3 - should open file picker
7. Press Esc - should exit edit mode

FILES:
- update_keyboard.go (primary - add early guard)
- render_preview.go (reference - understand edit mode)
- types.go (reference - model state)

See docs/EDIT_MODE_TYPING_FIX.md for full analysis and test plan.
```
