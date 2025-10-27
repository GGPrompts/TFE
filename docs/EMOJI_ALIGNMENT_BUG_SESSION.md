# Emoji Alignment Bug - Debugging Session (2025-10-27)

**Status:** UNRESOLVED - Issue persists after multiple fix attempts

**Environment:** Termux on Android, also affects WezTerm
**Working Environment:** Windows Terminal (perfect alignment)

---

## The Problem

Certain emoji icons render as 1 cell wide in Termux/WezTerm instead of 2 cells, causing misalignment:

**Narrow (1 cell) emojis:**
- ⬆️ (up arrow - parent directory)
- ⚙️ (settings gear - config files .ini, .conf, .cfg)
- 🗜️ (compression - .zip, .gz, .7z)
- 🖼️ (image frame - .png, .jpg, .webp)

**Normal (2 cell) emojis:**
- 📦 (package)
- 📁 (folder)
- 📝 (memo)
- Most other emojis

**Symptom:** Files with narrow emojis have:
- File names starting 1 space earlier than other rows
- All columns shifted left by 1 space
- Box borders misaligned by 1 space to the left
- Entire right edge of preview pane/box shifted left

**Example:**
```
  ⬆️  parent_dir       <-- shifted left
  📦 package.tar      <-- correct alignment
  ⚙️  config.ini       <-- shifted left
  📁 normal_folder    <-- correct alignment
```

---

## What We Learned

### Terminal Rendering Differences

**Test results in Termux:**
```bash
echo "⬆️x"  # x appears further left
echo "📦x"  # x appears in normal position
```
Result: The two 'x' characters do NOT align - confirming ⬆️ renders as 1 cell, 📦 as 2 cells.

**Windows Terminal:** Renders emoji+variation-selector (⬆️) as 2 cells - that's why it works there!

### Library Behavior

**`runewidth.StringWidth()` - CORRECT for emoji units:**
```go
runewidth.StringWidth("⬆️")  // Returns 1 (correct in Termux)
runewidth.StringWidth("📦")  // Returns 2 (correct)
```

**`runewidth.RuneWidth()` - WRONG for variation selectors:**
```go
for _, ch := range "⬆️" {
    runewidth.RuneWidth(ch)  // Returns 2 for base, 2 for VS = 4 total (wrong!)
}
```

**Key insight:** Emoji + Variation Selector must be measured as a UNIT, not character-by-character.

---

## Fixes Attempted (All Failed)

### Fix #1: Terminal Detection Order
**File:** `model.go:187-197`
**Change:** Moved Termux detection BEFORE xterm check
**Reason:** Termux sets `TERM=xterm-256color`, was being detected as xterm
**Result:** ✅ Termux now correctly detected, but alignment still broken

### Fix #2: Icon Padding to 2 Cells
**Files:**
- `file_operations.go:1230-1234` - Added `padIconToWidth()` helper
- `render_file_list.go:157,166,477,568,1182,1195` - Applied padding in all view modes

**Changes:**
```go
// Pad all icons to exactly 2 cells
paddedIcon := m.padIconToWidth(icon)  // "⬆️" becomes "⬆️ " (1 cell + 1 space)
name := fmt.Sprintf("%s%s %s", paddedIcon, favIndicator, displayName)
```

**Theory:** Pad narrow emojis with spaces to match wide emoji width
**Result:** ❌ No effect on alignment

### Fix #3: Git Repos Section Icon Padding
**File:** `render_file_list.go:567-569`
**Issue:** Git repos mode was using raw `icon` instead of `paddedIcon`
**Result:** ❌ Still no effect

### Fix #4: String Replacement Icon References
**File:** `render_file_list.go:665,671,673,680,682`
**Issue:** Color styling replacement was using raw `icon` instead of `paddedIcon`
**Result:** ❌ Still no effect

### Fix #5: ANSI Code Width Calculation
**File:** `file_operations.go:1239-1248`
**Issue:** `padToVisualWidth()` was using `runewidth.StringWidth()` which counts ANSI escape codes as visible characters
**Change:** Use `visualWidthCompensated()` which strips ANSI first
**Result:** ❌ Still no effect

### Fix #6: visualWidthCompensated() ANSI Handling
**File:** `file_operations.go:969-986`
**Issue:** `visualWidthCompensated()` was using `runewidth.StringWidth()` on ANSI-containing strings
**Change:** Use `visualWidth()` which strips ANSI codes first
**Result:** ❌ Still no effect

### Fix #7: visualWidth() Emoji Unit Calculation
**File:** `file_operations.go:933-964`
**Issue:** `visualWidth()` was using `runewidth.RuneWidth()` character-by-character:
- ⬆️ = base(2) + variation-selector(2) = 4 cells (WRONG)
**Change:** Strip ANSI codes, then use `runewidth.StringWidth()` on whole string:
- ⬆️ = 1 cell (CORRECT)
**Result:** ❌ STILL no effect!

---

## Current Code State (after all fixes)

### Terminal Detection
```go
// model.go:187-197
// Check for Termux (Android) - BEFORE xterm check
// Termux sets TERM=xterm-256color, so check PREFIX first
if strings.Contains(os.Getenv("PREFIX"), "com.termux") {
    return terminalTermux
}
```

### Icon Padding
```go
// file_operations.go:1230-1234
func (m model) padIconToWidth(icon string) string {
    return m.padToVisualWidth(icon, 2)
}
```

### Width Calculation
```go
// file_operations.go:936-964
func visualWidth(s string) int {
    // Strip ANSI codes first
    stripped := ""
    inAnsi := false
    for _, ch := range s {
        if ch == '\033' {
            inAnsi = true
            continue
        }
        if inAnsi {
            if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
                inAnsi = false
            }
            continue
        }
        stripped += string(ch)
    }
    // Use StringWidth on whole string (handles emoji+VS as unit)
    return runewidth.StringWidth(stripped)
}
```

### Padding Application
All rendering locations use `paddedIcon`:
- List view: `render_file_list.go:157,166`
- Detail view: `render_file_list.go:477`
- Git repos view: `render_file_list.go:568`
- Tree view: `render_file_list.go:1182,1195`
- Color styling replacements: `render_file_list.go:665,671,673,680,682`

---

## Tests Performed

### Test 1: Terminal Rendering
```bash
echo "⬆️x"
echo "📦x"
```
**Result:** The 'x' characters do NOT align (first is left of second)
**Conclusion:** ⬆️ is 1 cell, 📦 is 2 cells in Termux

### Test 2: runewidth.StringWidth()
```go
runewidth.StringWidth("⬆️")  // 1
runewidth.StringWidth("📦")  // 2
runewidth.StringWidth("⚙️")  // 1
runewidth.StringWidth("🖼️")  // 1
```
**Conclusion:** Library correctly reports emoji widths

### Test 3: visualWidth() Before Fix
```go
visualWidth("⬆️")  // 4 (WRONG - counted char-by-char)
visualWidth("📦")  // 2 (correct)
```

### Test 4: visualWidth() After Fix
```go
visualWidth("⬆️")  // Should be 1 (correct)
```

### Test 5: Padding Logic
```go
padToWidth("⬆️", 2)  // "⬆️ " = 2 cells visually
```
**In isolation this works correctly**

---

## Theories Why Nothing Works

### Theory 1: Lipgloss Style Rendering Interference
Lipgloss may be re-measuring or re-rendering text in a way that strips our padding. When `selectedStyle.Render()` is called, it might:
1. Measure the text width itself
2. Add its own padding
3. Ignore our pre-added spaces

### Theory 2: Terminal Escape Code Issue
The padding spaces might be getting consumed by terminal control codes in ways we don't understand. The "corrupted text with spaces between characters" we saw earlier suggests something is fundamentally wrong with how text is being output.

### Theory 3: Box Border Calculation
The box rendering (`lipgloss.NewStyle().Border()`) might be calculating widths independently and not accounting for our padding. It may be:
1. Measuring content width incorrectly
2. Using different width calculation methods
3. Not respecting our padded icon widths

### Theory 4: Column Width Assumptions
There may be hardcoded assumptions about icon widths elsewhere:
```go
// render_file_list.go:241
maxNameTextLen := nameWidth - 5  // Assumes icon(2) + space(1) + star(2)
```
But this doesn't account for variable-width icons.

### Theory 5: The Padding Isn't Actually Applied
Despite all our changes, maybe:
- The binary isn't being updated properly
- A different code path is being used
- Caching is preventing changes from taking effect

---

## What We Know For Sure

1. ✅ Windows Terminal renders ALL emojis consistently (alignment perfect)
2. ✅ Termux/WezTerm render emoji+VS as 1 cell
3. ✅ `runewidth.StringWidth()` correctly reports widths for Termux
4. ✅ Termux is being detected correctly (terminal type check confirmed)
5. ✅ Icon padding code is in the binary (confirmed with `strings ~/bin/tfe | grep padIconToWidth`)
6. ✅ All rendering code paths use `paddedIcon`
7. ❌ Despite all fixes, alignment is STILL broken

---

## Next Steps to Try

### 1. Debug Output Binary
Add temporary debug logging to see what's actually happening:
```go
// In renderDetailView before building line
fmt.Fprintf(os.Stderr, "DEBUG: icon=[%s] width=%d paddedIcon=[%s] paddedWidth=%d\n",
    icon, m.visualWidthCompensated(icon), paddedIcon, m.visualWidthCompensated(paddedIcon))
```

### 2. Test Without Lipgloss Styling
Try rendering a line WITHOUT any lipgloss styles to see if that's interfering:
```go
line := fmt.Sprintf("%s  %-*s  %-*s  %-*s", paddedName, sizeWidth, size, modifiedWidth, modified, extraWidth, fileType)
s.WriteString(line)  // Skip all styling
```

### 3. Use Fixed-Width Replacement Character
Instead of trying to pad narrow emojis, REPLACE them with normal 2-cell emojis:
```go
func normalizeIconWidth(icon string) string {
    // Replace narrow emojis with similar-looking 2-cell emojis
    replacements := map[string]string{
        "⬆️": "⏫",  // Up arrow → double up
        "⚙️": "🔧",  // Gear → wrench
        "🗜️": "📦",  // Compression → box
        "🖼️": "🎨",  // Frame → palette
    }
    if replacement, ok := replacements[icon]; ok {
        return replacement
    }
    return icon
}
```

### 4. Force All Icons to 2-Cell Emojis
Audit `file_operations.go` and replace ALL 1-cell emoji returns with 2-cell equivalents.

### 5. Box Width Override
Explicitly calculate and set box widths instead of relying on automatic sizing:
```go
boxWidth := calculateExpectedWidth()  // Sum of all column widths + padding
box := lipgloss.NewStyle().
    Width(boxWidth).
    Border(lipgloss.RoundedBorder())
```

### 6. Raw Terminal Output Test
Bypass all TFE rendering and write directly to terminal:
```go
fmt.Print("\033[2J\033[H")  // Clear screen, home cursor
fmt.Printf("  ⬆️  test\n")
fmt.Printf("  📦 test\n")
```
If this aligns correctly, the issue is in TFE's rendering pipeline.

### 7. Check for Lipgloss Issues
Search Lipgloss issues for emoji width handling bugs. There may be known issues or workarounds.

### 8. Terminal Capability Checking
Check if there's a way to query the terminal for how it renders specific characters and adjust accordingly.

### 9. User Configuration Option
Add a config option to manually specify emoji width behavior:
```
export TFE_EMOJI_WIDTH_MODE=force-2cell
```

### 10. Compare Windows Terminal vs Termux Binary Behavior
Run the SAME binary in both terminals and add debug logging to see where behavior diverges.

---

## Files Modified

All changes committed to repository:

1. `model.go` - Terminal detection order
2. `file_operations.go` - Icon padding, width calculation fixes
3. `render_file_list.go` - Icon padding application in all views
4. `tfe-wrapper.sh` - Added `~/bin/tfe` to search path

---

## Summary

Despite 7 different fix attempts covering:
- ✅ Terminal detection
- ✅ Icon padding to consistent width
- ✅ ANSI escape code handling
- ✅ Emoji+variation-selector unit handling
- ✅ All rendering code paths

**The alignment issue persists in Termux/WezTerm.**

The root cause appears to be deeper than width calculations - possibly in how Lipgloss or the terminal itself handles mixed-width content, or an interaction we haven't discovered yet.

---

**Session ended:** 2025-10-27 05:40 (user going to sleep)
**Next session:** Continue with debug output approach or try emoji replacement strategy
