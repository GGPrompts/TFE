# Fix F3 Browser Opening from Image Preview

## Issue

Helper text in image preview says "Press F3 to open in browser", but F3 doesn't work when viewing the preview pane.

**Current behavior:**
- F3 works from FILE LIST to open browser (update_keyboard.go:1679)
- F3 does NOT work from PREVIEW PANE (viewFullPreview mode)
- V key works in preview to open external image viewer (update_keyboard.go:652)
- User sees "Press F3 to open in browser" but F3 does nothing

**Root cause:**
- F3 handler at line 1679 only works in file list mode: `if currentFile := m.getCurrentFile()`
- Preview mode key handler (line 530-680) has V key but NO F3 key
- Helper text added in file_operations.go:2104 suggests F3 works

## Solution Options

### Option A: Add F3 to Preview Mode (Recommended)

Add F3 handler to preview mode to open current file in browser:

**Location:** `update_keyboard.go` around line 657 (after V key handler)

```go
case "v", "V":
    // View image in terminal viewer (for binary image files)
    if m.preview.loaded && m.preview.isBinary && isImageFile(m.preview.filePath) {
        return m, openImageViewer(m.preview.filePath)
    }

case "f3":
    // F3: Open in browser (images/HTML/PDF)
    if m.preview.loaded && m.preview.filePath != "" {
        if isBrowserFile(m.preview.filePath) {
            return m, openInBrowser(m.preview.filePath)
        }
    }
```

**Pro:** Makes F3 consistent (works in both file list and preview)
**Pro:** Helper text becomes accurate
**Con:** None

### Option B: Fix Helper Text Only

Change helper text to reflect current behavior:

**Location:** `file_operations.go:2104`

```go
// BEFORE:
footer = append(footer, "   Press F3 to open in browser")

// AFTER:
footer = append(footer, "   Press Esc, then F3 to open in browser")
// OR
footer = append(footer, "   Press F4 to open in external viewer")
```

**Pro:** Quick fix
**Con:** Requires extra step (exit preview first)
**Con:** Less intuitive UX

### Option C: Use Different Keys

Move browser opening to a key that works everywhere:

- F3 = Full preview (current behavior in file list)
- F4 = Open appropriate viewer (current in preview, line 568)
- B = Open in browser (new, works everywhere)

**Pro:** Consistent key for "browser"
**Con:** Adds new hotkey
**Con:** More changes needed

## Recommended Implementation

**Go with Option A** - Add F3 to preview mode.

**Files to modify:**
1. `update_keyboard.go` (~line 657) - Add F3 case in preview mode handler
2. Test in preview mode with image/HTML/PDF files

**Testing:**
```bash
# 1. View image in preview
./tfe
# Navigate to .png file
# Press Enter (full preview)
# Press F3 → Should open in browser ✅

# 2. Verify from file list still works
# Press Esc (back to file list)
# Press F3 → Should open in browser ✅

# 3. Test with HTML
# Navigate to .html file
# Press Enter, then F3 → Should open in browser ✅

# 4. Test V key still works
# Press Enter on image
# Press V → Should open in terminal viewer (chafa/timg) ✅
```

## Quick Prompt for Claude

```
Add F3 browser opening to preview mode.

ISSUE: F3 opens browser from file list, but not from preview pane. Helper text says "Press F3" but it doesn't work.

FIX: In update_keyboard.go around line 657 (after V key handler in preview mode):

Add:
case "f3":
    // F3: Open in browser (images/HTML/PDF)
    if m.preview.loaded && m.preview.filePath != "" {
        if isBrowserFile(m.preview.filePath) {
            return m, openInBrowser(m.preview.filePath)
        }
    }

TEST:
1. View image/HTML in preview (Enter key)
2. Press F3 → Should open in browser
3. Verify V key still works for external viewer
4. Verify F3 from file list still works

UPDATE CHANGELOG.md
```

## Reference

- **Existing F3 handler** (file list): `update_keyboard.go:1679`
- **Preview mode handlers**: `update_keyboard.go:530-680`
- **V key (works)**: `update_keyboard.go:652`
- **Helper text**: `file_operations.go:2104`
- **isBrowserFile()**: `editor.go:154` (images/HTML/PDF)
