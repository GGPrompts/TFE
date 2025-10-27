# Emoji Alignment Bug - Debug Session 2 (2025-10-27)

**Branch:** `emojidebug`
**Status:** Ready for testing in Termux/WezTerm

---

## Summary of Changes

Based on comprehensive research into Lipgloss, go-runewidth, terminal rendering, and other TUI applications, identified two unexplored root causes and implemented fixes.

---

## Root Causes Identified

### 1. **go-runewidth Bug #76 (Open since Feb 2024)**

**Issue:** Variation Selectors (U+FE0F, U+FE0E) incorrectly report width = 1 instead of width = 0

**Impact on TFE:**
- `visualWidth("⬆️")` was calculating: base emoji (1) + VS (1) = 2 cells
- Should be: base emoji (1) + VS (0) = 1 cell
- This caused `padIconToWidth()` to think icons were already 2 cells wide
- No padding was added, but terminal rendered emoji as 1 cell
- Result: Misalignment by 1 space

### 2. **Terminal Rendering Inconsistency**

**Research finding:** Different terminals handle emoji + variation selector differently:

| Terminal | Emoji+VS Rendering | What They See |
|----------|-------------------|---------------|
| **Windows Terminal** | 2 cells | ⬆️ = 2 cells (wide) |
| **WezTerm** | 1 cell | ⬆️ = 1 cell (narrow) |
| **Termux** | 1 cell | ⬆️ = 1 cell (narrow) |
| **Kitty** | 2 cells | ⬆️ = 2 cells (wide) |

The variation selector (U+FE0F) requests "emoji presentation" (colorful, typically 2 cells), but terminals disagree on whether this affects layout width.

---

## Fixes Implemented

### Fix 1: Strip Variation Selectors in visualWidth()

**File:** `file_operations.go:961-964`

**Change:**
```go
// Strip variation selectors to work around go-runewidth bug #76
// VS incorrectly reports width=1 instead of width=0, causing padding miscalculations
stripped = strings.ReplaceAll(stripped, "\uFE0F", "") // VS-16 (emoji presentation)
stripped = strings.ReplaceAll(stripped, "\uFE0E", "") // VS-15 (text presentation)
```

**Rationale:** By removing VS before width calculation, we measure the base emoji character only, which gives accurate width for padding calculations.

### Fix 2: Strip VS from Icons in WezTerm/Termux

**File:** `file_operations.go:1237-1243`

**Change:**
```go
// Strip variation selectors for terminals that render emoji+VS as 1 cell
// This prevents width calculation mismatches
if m.terminalType == terminalWezTerm || m.terminalType == terminalTermux {
    icon = strings.ReplaceAll(icon, "\uFE0F", "") // VS-16 (emoji presentation)
    icon = strings.ReplaceAll(icon, "\uFE0E", "") // VS-15 (text presentation)
}
```

**Rationale:** Remove VS from the actual displayed string in terminals that don't handle it well. This ensures:
1. The emoji renders without VS (e.g., "⬆" instead of "⬆️")
2. Width calculation matches actual terminal rendering
3. Padding is applied correctly

**Flow after fix:**
1. Icon "⬆️" enters `padIconToWidth()`
2. For WezTerm/Termux: Strip VS → "⬆"
3. Calculate width: `visualWidth("⬆")` = 1 cell (base emoji without VS)
4. Pad to 2 cells: "⬆" + " " = "⬆ " (2 cells)
5. Terminal renders: ⬆ (1 cell) + space (1 cell) = 2 cells total ✓

---

## What to Test

### Test 1: Basic Alignment
Navigate through directories with mixed file types and check if columns align properly:
- Files with ⬆️ (parent dir)
- Files with ⚙️ (config files)
- Files with 🗜️ (compressed files)
- Files with 🖼️ (images)
- Files with 📦 (packages) - should still align with others

### Test 2: All View Modes
Test alignment in:
- **List view** (F1)
- **Detail view** (F2) - default
- **Tree view** (F3)

### Test 3: All Display Contexts
- Single-pane mode (F10)
- Dual-pane horizontal split (wide terminal)
- Dual-pane vertical split (narrow terminal or Detail mode)
- Favorites view (F7)
- Git repos view (F8)

### Test 4: Visual Appearance
Check if emojis lost their colorful presentation:
- Do emojis still appear colorful or are they now monochrome?
- Are some emojis now missing entirely?
- Do emoji appear as boxes/question marks?

### Expected Results

**Alignment:** ✅ All file names should start at the same column, regardless of emoji type

**Emoji appearance:**
- **May look slightly different** (variation selector removed)
- Some emojis might be monochrome instead of colorful
- This is an acceptable trade-off for proper alignment

---

## Research Findings Summary

### Other Projects Have Similar Issues

- **lazygit**: Issue #3514 - emoji alignment problems (still open)
- **k9s**: Provides `noIcons` config option to disable emoji entirely
- **fzf**: Fixed by switching from `RuneWidth()` to `StringWidth()`
- **Lipgloss**: PR #563 (still open) trying to improve emoji width handling

### Key Insights

1. **No perfect solution exists** - Different terminals render emoji differently, and there's no way to detect this at runtime without complex cursor position queries

2. **go-runewidth limitations:**
   - Issue #76 (VS width bug) is OPEN and unfixed since Feb 2024
   - Issue #59 ("first non-zero width" heuristic fails for some scripts)
   - Issue #28 (Flag emoji/Regional Indicators measured incorrectly)

3. **Unicode Standard Gap:** Unicode only defines width at the codepoint level, not grapheme level. There's no authoritative answer for complex emoji sequences.

4. **TFE's approach matches industry best practices:**
   - Use `StringWidth()` not `RuneWidth()`
   - Strip ANSI before calculation
   - Apply terminal-specific compensation
   - All modern TUI apps do something similar

---

## Build Instructions

```bash
# Branch is already checked out
git status  # Should show: On branch emojidebug

# Build binary
go build -o ~/bin/tfe

# Or build with different name to test alongside main version
go build -o ~/bin/tfe-emojidebug
```

---

## Alternative Approaches Not Implemented

These were considered but not implemented in this session (can try if VS stripping doesn't work):

### Priority 2 Options:
1. **Emoji replacement map** - Replace problematic emoji with always-wide alternatives
2. **Ideographic space padding** - Use U+3000 instead of ASCII space
3. **Runtime width detection** - Query terminal with cursor position (complex)

### Not Recommended:
- ❌ Zero-width joiners (ZWJ) - Makes problems worse
- ❌ Unicode normalization (NFC/NFD) - Doesn't affect VS
- ❌ Adding more VS - Compounds the issue

---

## Commit Message Template

```
fix: Strip variation selectors to fix emoji alignment in WezTerm/Termux

- Strip U+FE0F and U+FE0E from icons in WezTerm/Termux before display
- Strip VS in visualWidth() to work around go-runewidth bug #76
- Ensures width calculation matches actual terminal rendering

Addresses emoji alignment issues documented in:
- docs/EMOJI_ALIGNMENT_BUG_SESSION.md
- CHANGELOG.md (entry from 2025-10-24)

Root cause: go-runewidth bug #76 (VS reports width=1 instead of 0)
+ terminal-specific emoji+VS rendering differences

Files modified:
- file_operations.go (padIconToWidth, visualWidth)

Testing needed in:
- Termux (Android)
- WezTerm (WSL/Linux/macOS)
```

---

## Next Steps

1. **Test in Termux:** Deploy `tfe-emojidebug` binary and verify alignment
2. **Test in WezTerm:** Check both WSL and native environments
3. **Document results:** Update this file with test outcomes
4. **If successful:** Merge to main with commit message above
5. **If unsuccessful:** Try Priority 2 alternatives (emoji replacement map)

---

**Session Started:** 2025-10-27
**Status:** Awaiting testing in Termux/WezTerm
**Claude Model:** Sonnet 4.5 (claude-sonnet-4-5-20250929)
