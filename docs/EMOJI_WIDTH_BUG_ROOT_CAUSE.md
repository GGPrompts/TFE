# TFE Emoji Width Bug - Root Cause Analysis

**⚠️ STATUS: UNVERIFIED THEORY - BUG NOT FIXED**

**Source:** Codex GPT-5 (2025-10-27) - Deep debugging session (3 hours)
**Problem:** Rows with emojis ⬆️ ⚙️ 🗜️ 🖼️ shift left by 1 space in WezTerm only
**Resolution:** **Not fixed - abandoned as not worth the effort**

## Why This Bug Was Abandoned

1. **Don't use WezTerm** - Primary terminal is Windows Terminal (works fine)
2. **WezTerm was only for testing HD image previews**
3. **HD image previews don't work in WSL→Windows anyway** (terminal graphics protocols can't cross the WSL boundary)
   - WezTerm DOES support Kitty protocol (works on native Linux/macOS/Windows)
   - The issue is WSL→Windows Terminal boundary doesn't pass through escape sequences
4. **Even Yazi has the same WSL limitation** (not unique to TFE)
5. **3 hours of debugging with no resolution** (diminishing returns)

**If you DO use WezTerm and want to fix this, the theory below might help. Otherwise, ignore it.**

---

## Codex's Theory (Unverified)

## Root Cause: TWO Different Width Engines

You have **two incompatible width calculation systems** running simultaneously:

### 1. `padToVisualWidth()` / `visualWidthCompensated()` (file_operations.go)
- Uses `runewidth.StringWidth()`
- Returns **1 cell** for `⚙️` in WezTerm ✅ (correct)
- Returns **2 cells** for `⚙️` in Windows Terminal ✅ (correct)

### 2. `m.runeWidth()` (render_file_list.go:892)
- Uses **hard-coded Unicode ranges**:
  ```go
  if r >= 0x2600 && r <= 0x26FF { // Misc symbols (many emojis)
      return 2
  }
  ```
- Returns **2 cells** for `⚙️` (U+2699) in **ALL terminals** ❌ (wrong for WezTerm)
- Used by `truncateToVisualWidth()` and `extractVisiblePortion()`

---

## The Mismatch

```
⚙️ (U+2699 + FE0F) in WezTerm:

runewidth.StringWidth: 1 cell  ← padToVisualWidth uses this
m.runeWidth:           2 cells ← truncateToVisualWidth uses this

Result: 1-column disagreement = visual shift left by 1 space
```

**Windows Terminal:** Masked because both functions agree on 2 cells (coincidentally correct)

---

## Where It Breaks

1. **`padToVisualWidth()`** pads name to `nameWidth` assuming emoji is **1 cell**
2. **`truncateToVisualWidth()`** / **`extractVisiblePortion()`** measure using `m.runeWidth()` thinking emoji is **2 cells**
3. Later rendering code believes content is **1 column longer** than it actually is
4. Subsequent columns appear **1 column earlier** → "border shifted left by 1"

---

## Additional Issues

### Problem 2: `fmt.Sprintf` with Width Specifiers

`fmt`'s `%-*s` uses **rune count**, not display width:
- `"⚙️"` = 2 runes (U+2699 + U+FE0F)
- `fmt.Sprintf("%-5s", "⚙️")` pads for 5 runes, not 5 visual cells

**Current code:** Correctly avoids `%-*s` for name column, but uses it for other columns

### Problem 3: Byte-Length Truncation

Several places use `len()` and `s[:n]` for truncation:
- `displayName`, `repoDisplayName`, `fileType`, `location`
- Can split mid-rune or be visually incorrect

---

## The Fix

### 1. Unify on One Width Model

**Replace `m.runeWidth()` with delegating to `runewidth`:**

```go
// render_file_list.go:892
func (m model) runeWidth(r rune) int {
    // Delegate to runewidth library (correct for all terminals)
    w := runewidth.RuneWidth(r)

    // Only compensate for variation selector in Windows Terminal
    if m.terminalType == terminalWindowsTerminal && r == '\uFE0F' {
        return 1  // Add +1 for VS in Windows Terminal
    }

    return w
}
```

**Remove hard-coded Unicode ranges** (0x2600-0x26FF, etc.) - they don't match actual terminal behavior.

### 2. Update All Width Calculations

Make sure `truncateToVisualWidth()` and `extractVisiblePortion()` use the unified `m.runeWidth()` (which now delegates to runewidth).

### 3. Stop Byte-Length Truncation

Replace all `len()/s[:n]` truncations with `truncateToVisualWidth()`:

**Bad:**
```go
if len(displayName) > maxLen {
    displayName = displayName[:maxLen-2] + ".."
}
```

**Good:**
```go
if m.visualWidthCompensated(displayName) > maxLen {
    displayName = m.truncateToVisualWidth(displayName, maxLen-2) + ".."
}
```

### 4. Avoid `fmt` Width for Emoji Fields

- Keep using `padToVisualWidth()` for name column ✅
- Never use `%-*s` on emoji-bearing strings
- Headers (ASCII only) can use `fmt` width specifiers

### 5. Verify Icon Width Constant

```go
// renderDetailView
maxNameTextLen := nameWidth - 5  // icon + star + spacing
```

Consider computing dynamically:
```go
iconWidth := m.visualWidthCompensated(icon)
starWidth := m.visualWidthCompensated(favIndicator)
maxNameTextLen := nameWidth - iconWidth - starWidth - 1  // spacing
```

---

## Quick Verification

**Minimal repro:**
1. Print single row: `name = "⚙️ Test"`
2. Pad with `padToVisualWidth(name, 20)`
3. Add two ASCII columns via `fmt.Sprintf`
4. Compare WezTerm vs Windows Terminal

You should see:
- **Before fix:** 1-column drift in WezTerm when m.runeWidth path is involved
- **After fix:** Perfect alignment in both terminals

---

## Files to Modify

| File | Function | Line | Fix |
|------|----------|------|-----|
| render_file_list.go | `runeWidth()` | 892 | Delegate to runewidth, remove hard-coded ranges |
| render_file_list.go | `truncateToVisualWidth()` | 842 | Uses unified m.runeWidth (now correct) |
| render_file_list.go | `extractVisiblePortion()` | 760 | Uses unified m.runeWidth (now correct) |
| render_file_list.go | Various | Multiple | Replace byte-length truncations with visual truncations |

---

## Why This Worked in Windows Terminal

Windows Terminal renders these emojis as **2 cells**, which **accidentally matches** the hard-coded `m.runeWidth()` assumption.

The bug was always there, just masked in Windows Terminal.

---

**Generated:** 2025-10-27
**Model:** Codex GPT-5 with high reasoning effort
**Status:** Root cause identified, fix ready to implement
