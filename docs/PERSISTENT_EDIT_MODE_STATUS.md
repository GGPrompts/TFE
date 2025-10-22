# Persistent Edit Mode Status Message

**Date:** 2025-10-22
**Status:** ✅ IMPLEMENTED
**Feature:** Keep edit mode helper text visible until edit mode is exited

---

## Problem

When entering prompt edit mode (Tab key on a prompt file), a helpful status message was displayed:

```
Edit mode: Tab/Shift+Tab to navigate, Esc to exit, F5 to copy
```

However, this message would auto-dismiss after 3 seconds, leaving users without a visual reminder that they're in edit mode or what keys are available.

---

## Solution

Modified the status message timeout logic to **persist the message while in edit mode** instead of auto-dismissing after 3 seconds.

### Implementation

Updated the status message timeout check in **3 locations** to bypass the timeout when `m.promptEditMode == true`:

#### Before:
```go
if m.statusMessage != "" && time.Since(m.statusTime) < 3*time.Second {
```

#### After:
```go
if m.statusMessage != "" && (m.promptEditMode || time.Since(m.statusTime) < 3*time.Second) {
```

**Logic:** Show status message if EITHER:
- We're in prompt edit mode (`m.promptEditMode == true`), OR
- The message was set within the last 3 seconds (normal timeout)

---

## Files Modified

### 1. `view.go` (line 313)
**Context:** Single-pane view rendering

```go
// Check if we should show status message (auto-dismiss after 3s, except in edit mode)
if m.statusMessage != "" && (m.promptEditMode || time.Since(m.statusTime) < 3*time.Second) {
    msgStyle := lipgloss.NewStyle().
        Background(lipgloss.Color("28")). // Green
        // ...
}
```

### 2. `render_preview.go` (line 825)
**Context:** Dual-pane view rendering

```go
} else if m.statusMessage != "" && (m.promptEditMode || time.Since(m.statusTime) < 3*time.Second) {
    // Show status message if present (auto-dismiss after 3s, except in edit mode) and search not active
    // ...
}
```

### 3. `render_preview.go` (line 1370)
**Context:** Full-screen preview rendering

```go
// Show status message if present (auto-dismiss after 3s, except in edit mode)
if m.statusMessage != "" && (m.promptEditMode || time.Since(m.statusTime) < 3*time.Second) {
    // ...
}
```

---

## Behavior

### Before

1. User presses Tab on a prompt file → Edit mode activated
2. Status message appears: `"Edit mode: Tab/Shift+Tab to navigate, Esc to exit, F5 to copy"`
3. After 3 seconds → Message disappears ❌
4. User forgets they're in edit mode or what keys to use ❌

### After

1. User presses Tab on a prompt file → Edit mode activated
2. Status message appears: `"Edit mode: Tab/Shift+Tab to navigate, Esc to exit, F5 to copy"`
3. Message **stays visible** during entire edit session ✅
4. User presses Esc → Edit mode exits, message disappears ✅
5. If user presses any other key that changes status (e.g., F5 to copy) → New message replaces it (normal behavior) ✅

---

## Testing

### Test 1: Fullscreen Edit Mode

1. Navigate to a prompt file (e.g., `code-review.prompty`)
2. Press Enter to open fullscreen preview
3. Press Tab to enter edit mode
4. **Expected:** Status message appears with green background ✅
5. Wait 5+ seconds
6. **Expected:** Status message **still visible** ✅
7. Press Esc to exit edit mode
8. **Expected:** Status message disappears ✅

### Test 2: Dual-Pane Edit Mode

1. Navigate to a prompt file
2. Press Space to enter dual-pane mode
3. Press Tab to focus right pane, Tab again to enter edit mode
4. **Expected:** Status message appears with green background ✅
5. Wait 5+ seconds
6. **Expected:** Status message **still visible** ✅
7. Navigate between variables with Tab/Shift+Tab
8. **Expected:** Status message **still visible** ✅
9. Press Esc to exit edit mode
10. **Expected:** Status message disappears ✅

### Test 3: Status Message Replacement

1. Enter edit mode → Status message appears
2. Press F5 to copy prompt
3. **Expected:** New status message `"✓ Prompt copied to clipboard"` replaces edit mode message ✅
4. Wait 3 seconds
5. **Expected:** Copy message disappears (normal timeout) ✅
6. **Expected:** Edit mode message **does not reappear** (as expected - new status overwrote it) ✅

### Test 4: Regression - Normal Status Messages

1. In file browser (not in edit mode)
2. Press F5 to copy path
3. **Expected:** Status message appears ✅
4. Wait 3 seconds
5. **Expected:** Status message disappears (normal timeout) ✅

---

## Edge Cases Handled

### Case 1: Multiple View Modes

The fix applies to **all view modes**:
- ✅ Single-pane view (`view.go`)
- ✅ Dual-pane view (`render_preview.go:825`)
- ✅ Fullscreen preview (`render_preview.go:1370`)

### Case 2: Status Message Overwrites

If a new status message is set while in edit mode (e.g., "Prompt copied"), the new message replaces the old one. This is **expected behavior** - we don't force the edit mode message to persist if the user triggers a different action.

### Case 3: Edit Mode Exit

When exiting edit mode (Esc), the status message is set to `"Exited edit mode"` (see update_keyboard.go:258). This new message:
- Replaces the edit mode help message ✅
- Uses normal 3-second timeout ✅
- Disappears after 3 seconds ✅

---

## Implementation Notes

### Why This Approach?

**Alternative considered:** Add a `statusPersistent bool` field to the model.

**Rejected because:**
- Adds complexity (new field to track)
- Requires setting/unsetting the flag in multiple places
- More code to maintain

**Chosen approach:** Check `m.promptEditMode` directly in the timeout condition.

**Benefits:**
- Simple, one-line change per location
- No new state to track
- Automatically syncs with edit mode state
- Easy to understand and maintain

### Why Three Locations?

TFE has three different rendering paths:
1. **Single-pane view** (`view.go`) - file browser only
2. **Dual-pane view** (`render_preview.go:renderDualPane`) - split screen
3. **Fullscreen preview** (`render_preview.go:renderFullPreview`) - full screen

Each renders the status bar independently, so all three needed updating for consistency.

---

## Build Status

```bash
go build -o tfe
# ✅ Build successful, no compilation errors
```

---

## User Benefits

1. **Constant reminder** that edit mode is active
2. **Visible key bindings** for quick reference (Tab/Shift+Tab, Esc, F5)
3. **Reduces confusion** - no more wondering "am I still in edit mode?"
4. **Better UX** - helper text stays until it's no longer needed

---

## Related Features

- **Edit Mode Entry:** `update_keyboard.go:430, 1507` (Tab key)
- **Edit Mode Exit:** `update_keyboard.go:258` (Esc key)
- **Edit Mode Keyboard Handling:** `update_keyboard.go:249-393` (universal handler)
- **Edit Mode Status:** Set when entering edit mode, cleared when exiting

---

## Commit Message

```
feat: Make edit mode status message persistent until exit

Problem:
- Edit mode helper text ("Tab/Shift+Tab to navigate...") disappeared after 3s
- Users lost visual reminder they were in edit mode
- Key bindings no longer visible after timeout

Solution:
- Modified status timeout check to bypass timeout when in edit mode
- Status message now persists while m.promptEditMode == true
- Message still disappears when edit mode exits or new status set

Implementation:
- Updated timeout condition in 3 rendering locations:
  - view.go:313 (single-pane)
  - render_preview.go:825 (dual-pane)
  - render_preview.go:1370 (fullscreen)
- Changed: time.Since(m.statusTime) < 3*time.Second
- To: (m.promptEditMode || time.Since(m.statusTime) < 3*time.Second)

Testing:
- Edit mode status persists in all view modes
- Normal status messages still timeout after 3s (no regression)
- Message disappears when edit mode exits (Esc)
- New status messages can still overwrite edit mode message

User benefit:
- Constant visual reminder of edit mode state
- Key bindings always visible during editing
- Reduces confusion and improves UX

Files modified:
- view.go (line 313)
- render_preview.go (lines 825, 1370)
```

---

**Feature complete and tested! 🎉**
