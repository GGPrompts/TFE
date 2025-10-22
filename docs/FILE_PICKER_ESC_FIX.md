# File Picker Esc Key Fix

**Date:** 2025-10-22
**Status:** ✅ FIXED
**Bug:** Pressing Esc in file picker mode kept picker open and exited edit mode

---

## Problem

When using the file picker (F3) from edit mode:

1. User is in edit mode editing a prompt
2. User presses F3 to open file picker for a FILE/PATH variable
3. File picker opens ✅
4. User presses Esc to cancel
5. **BUG:** File picker stays open, but edit mode exits ❌

**Expected:** Esc should close the file picker and return to edit mode.

---

## Root Cause

### Keyboard Event Priority Order (Before Fix)

After adding the universal edit mode handler, the priority order was:

1. Landing page (line 34)
2. Fuzzy search (line 84)
3. Menu navigation (lines 91-207)
4. Preview search (line 209)
5. **EDIT MODE (line 249) ← Caught Esc first!** ❌
6. Preview mode (line 395)
7. Dialogs (line 667)
8. Context menu (line 903)
9. Search mode (line 950)
10. **FILE PICKER (line 1004) ← Never reached!** ❌
11. Command prompt (line 1063)

**Problem:** Edit mode check (line 249) intercepted Esc before file picker check (line 1004) could handle it.

### Code Flow

```go
// Line 249: Edit mode check (higher priority)
if m.promptEditMode && m.preview.isPrompt && m.preview.promptTemplate != nil {
    switch msg.String() {
    case "esc":
        // Exit edit mode ← This was executed!
        m.promptEditMode = false
        return m, nil
    }
}

// Line 1004: File picker check (lower priority, never reached)
if m.filePickerMode {
    switch msg.String() {
    case "esc":
        // Cancel file picker ← This was never reached!
        m.filePickerMode = false
        // ...
        return m, nil
    }
}
```

---

## Solution

**Move file picker mode check to HIGHER priority than edit mode.**

### Rationale

File picker mode is a **more specific state** than edit mode:
- File picker is launched FROM edit mode (F3)
- Esc should close the most specific/recent context first
- After closing picker, user returns to edit mode (still active)

### Implementation

**Before:** File picker check at line 1004 (after edit mode)

**After:** File picker check at line 249 (before edit mode)

```go
// PRIORITY 1: Handle file picker mode (F3 from edit mode)
// File picker has higher priority than edit mode because Esc should close picker first
if m.filePickerMode {
    switch msg.String() {
    case "esc":
        // Cancel file picker and return to preview mode
        m.filePickerMode = false
        m.showPromptsOnly = m.filePickerRestorePrompts
        m.loadFiles()
        m.viewMode = viewFullPreview
        if m.filePickerRestorePath != "" {
            m.loadPreview(m.filePickerRestorePath)
            m.populatePreviewCache()
        }
        m.setStatusMessage("File picker cancelled", false)
        return m, nil

    case "enter":
        // File selection logic...
    }
    // Fall through for navigation keys
}

// PRIORITY 2: Handle prompt edit mode input (works in ALL view modes)
// Must be checked AFTER file picker mode (so Esc closes picker first)
if m.promptEditMode && m.preview.isPrompt && m.preview.promptTemplate != nil {
    switch msg.String() {
    case "esc":
        // Exit prompt edit mode
        m.promptEditMode = false
        // ...
    }
}
```

---

## Files Modified

### `update_keyboard.go`

**Changes:**

1. **Lines 249-307:** Moved file picker mode check to top (PRIORITY 1)
2. **Lines 309-450:** Edit mode check now PRIORITY 2 (after file picker)
3. **Line 1060:** Removed duplicate file picker handling (was at line 1004-1117)

**Result:**
- File picker mode checked first ✅
- Edit mode checked second ✅
- Duplicate code removed ✅

---

## Updated Priority Order

After fix, keyboard event handling priority:

1. Landing page
2. Fuzzy search
3. Menu navigation
4. Preview search
5. **FILE PICKER MODE** ← Now PRIORITY 1! ✅
6. **EDIT MODE** ← Now PRIORITY 2 ✅
7. Preview mode
8. Dialogs
9. Context menu
10. Search mode
11. Command prompt
12. Regular keys

---

## Testing

### Test 1: File Picker Esc (Main Bug)

1. Navigate to a prompt file (e.g., `code-review.prompty`)
2. Press Enter for fullscreen, then Tab to enter edit mode
3. Press F3 to open file picker
4. **Expected:** File picker opens, shows file list ✅
5. Press Esc
6. **Expected:** File picker closes, returns to edit mode ✅
7. **Expected:** Edit mode still active (status message visible) ✅

### Test 2: File Picker Enter (Selection)

1. In edit mode, press F3 to open file picker
2. Navigate to a file
3. Press Enter
4. **Expected:** File picker closes ✅
5. **Expected:** Selected file path inserted into variable ✅
6. **Expected:** Status message shows `"✓ Set VARIABLE = filename"` ✅
7. **Expected:** Edit mode still active ✅

### Test 3: File Picker Navigation (Folders)

1. In edit mode, press F3 to open file picker
2. Navigate to a folder, press Enter
3. **Expected:** Navigate into folder (file picker stays open) ✅
4. Press Esc
5. **Expected:** File picker closes, returns to edit mode ✅

### Test 4: Edit Mode Esc (Without File Picker)

1. In edit mode, press Esc (without opening file picker)
2. **Expected:** Edit mode exits ✅
3. **Expected:** Status message: `"Exited edit mode"` ✅

### Test 5: Nested States (Regression)

1. Enter edit mode
2. Press F3 (file picker)
3. Press Esc (closes picker) ✅
4. Still in edit mode ✅
5. Press Esc again (exits edit mode) ✅

---

## Edge Cases Handled

### Case 1: File Picker Navigation

File picker allows navigation (up/down, left/right) while open. These keys should NOT be intercepted by edit mode:

```go
if m.filePickerMode {
    // ... handle Esc and Enter ...
    // Fall through for navigation keys ✅
}
```

**Result:** Navigation works in file picker (not blocked by edit mode) ✅

### Case 2: State Restoration

When file picker is cancelled (Esc), state is restored:

```go
m.filePickerMode = false
m.showPromptsOnly = m.filePickerRestorePrompts  // Restore filter
m.loadFiles()                                    // Reload with filter
m.viewMode = viewFullPreview                     // Return to preview
// Reload original preview
if m.filePickerRestorePath != "" {
    m.loadPreview(m.filePickerRestorePath)
    m.populatePreviewCache()
}
```

**Result:** Returns to exact state before F3 was pressed ✅

### Case 3: Edit Mode Persistence

Edit mode state (`m.promptEditMode`) is NOT changed when file picker closes:

```go
// File picker Esc handler
case "esc":
    m.filePickerMode = false
    // ... restore state ...
    return m, nil  // ← Does NOT change m.promptEditMode!
```

**Result:** Edit mode remains active after closing file picker ✅

---

## Implementation Notes

### Why Move Instead of Add Check?

**Alternative considered:** Add `!m.filePickerMode` check to edit mode handler:

```go
if m.promptEditMode && !m.filePickerMode && ... {
    // Edit mode handling
}
```

**Rejected because:**
- Clutters edit mode logic
- Harder to understand priority order
- Doesn't follow "check specific state first" principle

**Chosen approach:** Move file picker check to top (before edit mode).

**Benefits:**
- Clear priority order (specific before general)
- Follows natural nesting (file picker launched FROM edit mode)
- Easier to reason about state transitions

### Code Deduplication

Moved file picker handling from line 1004 to line 249, removed duplicate code.

**Before:** 58 lines of duplicate file picker code
**After:** Single file picker handler at top

**Benefit:** DRY principle, easier maintenance ✅

---

## Related Code

### Edit Mode Status

After closing file picker with Esc:
- Edit mode still active ✅
- Status message still shows helper text ✅ (see PERSISTENT_EDIT_MODE_STATUS.md)

### File Picker Launch

File picker is launched from edit mode:

```go
// In edit mode handler (line 359)
case "f3":
    // File picker for focused variable
    if m.focusedVariableIndex >= 0 && ... {
        m.filePickerMode = true
        m.filePickerRestorePath = m.preview.filePath
        m.filePickerRestorePrompts = m.showPromptsOnly
        m.showPromptsOnly = false  // Show all files
        m.viewMode = viewSinglePane
        m.loadFiles()
        m.setStatusMessage("📁 File Picker: Navigate and press Enter...", false)
    }
    return m, nil
```

---

## Build Status

```bash
go build -o tfe
# ✅ Build successful, no compilation errors
```

---

## Commit Message

```
fix: File picker Esc key now closes picker instead of exiting edit mode

Problem:
- In edit mode, pressing F3 opens file picker
- Pressing Esc kept file picker open but exited edit mode
- User had to manually navigate away from picker

Root cause:
- Edit mode check had higher priority than file picker check
- Edit mode intercepted Esc before file picker could handle it
- File picker handler was at line 1004 (after edit mode at line 249)

Solution:
- Moved file picker mode check to PRIORITY 1 (before edit mode)
- Edit mode is now PRIORITY 2 (after file picker)
- Removed duplicate file picker handler at line 1004

Behavior after fix:
- File picker mode (F3) → Esc closes picker, returns to edit mode ✅
- Edit mode still active after closing picker ✅
- Edit mode Esc (without picker) exits edit mode ✅

Implementation:
- Moved file picker handling from line 1004 to line 249
- Updated comments to reflect priority order
- Removed 58 lines of duplicate code

Testing:
- F3 in edit mode → Opens file picker ✅
- Esc in file picker → Closes picker, returns to edit mode ✅
- Edit mode remains active (status visible) ✅
- Esc in edit mode (no picker) → Exits edit mode ✅

Files modified:
- update_keyboard.go (lines 249-307, 309-450, 1060)

Related: PERSISTENT_EDIT_MODE_STATUS.md (edit mode status stays visible)
```

---

**Bug fixed and tested! 🎉**
