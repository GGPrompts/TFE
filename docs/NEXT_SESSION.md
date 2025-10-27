# Next Session - Horizontal Scrolling Bug in Detail View

**Priority:** Medium
**Added:** After implementing horizontal scrolling for Termux narrow screen support
**Status:** Bug - Display corruption during horizontal scroll

---

## The Issue

When horizontally scrolling in Detail view on narrow screens (Termux), the display becomes corrupted:

### Symptoms:
1. **Start scrolling right** → Works well initially
2. **Middle of scroll range** → Text expands and corrupts the display
   - Characters appear to have extra spaces between them
   - Columns misalign
   - Visual corruption/garbled text
3. **Near end of scroll** → Display looks normal again
4. **Extra empty space** → Can scroll through large amounts of empty space after the last column

### Environment:
- **Termux** (narrow terminal width)
- **Detail view** (F2)
- **Horizontal scroll** (Left/Right arrow keys when content exceeds screen width)

---

## Investigation Steps

### 1. Find the Horizontal Scroll Implementation

**Files to check:**
- `render_file_list.go` - Detail view rendering (look for horizontal scroll logic)
- `update_keyboard.go` - Left/Right arrow key handling in Detail view
- `types.go` - Check for horizontal scroll offset variable (e.g., `horizontalScrollOffset`)

**Questions:**
- When was horizontal scrolling added to Detail view?
- Is there a `git log` commit related to "horizontal scroll" or "Termux narrow"?
- How is the scroll offset calculated and applied?

### 2. Identify the Corruption Cause

**Likely culprits:**

#### A. **ANSI Code Splitting**
- Is the horizontal offset being applied to byte positions instead of visual positions?
- Are ANSI escape codes (colors) being split mid-sequence?
- Does `truncateToWidth()` or similar functions handle ANSI properly?

Example bug:
```go
// WRONG - Splits ANSI codes
line := coloredText[scrollOffset:]

// RIGHT - Use visual width functions
line := truncateToVisualWidth(coloredText, scrollOffset, visibleWidth)
```

#### B. **Emoji/Wide Character Handling**
- Is the offset calculation accounting for emoji width (1 or 2 cells)?
- Does scrolling use `visualWidth()` or raw string length?
- Are emojis being split (showing half an emoji)?

#### C. **Box Border Calculation**
- Is the available width calculation wrong when scrolling?
- Does Lipgloss box width change during scroll?
- Are padding/margins being recalculated incorrectly?

#### D. **Column Separator Spacing**
- Are column separators (e.g., "  ") being counted inconsistently?
- Does the scroll offset skip separators or include them?

### 3. Reproduce and Debug

**Test cases:**
```bash
# Narrow terminal (Termux or resize WezTerm to ~60 cols)
cd /some/directory/with/files
./tfe

# Press F2 (Detail view)
# Press Right arrow repeatedly
# Observe at which column/offset corruption starts
# Note when it returns to normal
```

**Add debug output:**
```go
// In render_file_list.go detail view rendering
fmt.Fprintf(os.Stderr, "DEBUG: scrollOffset=%d visibleWidth=%d lineWidth=%d\n",
    m.horizontalScrollOffset, visibleWidth, visualWidth(line))
```

### 4. Check the Extra Space Issue

**Why does empty space exist after last column?**
- Is `maxScrollOffset` calculated correctly?
- Should scroll be limited to: `max(0, totalContentWidth - visibleWidth)`?
- Is padding being added unnecessarily during scroll?

**Code to check:**
```go
// Look for horizontal scroll bounds checking
if m.horizontalScrollOffset > maxOffset {
    m.horizontalScrollOffset = maxOffset
}
```

---

## Potential Fixes

### Option A: Fix the Scrolling Logic ✅ Recommended

**If fixable, implement:**

1. **Use visual width functions everywhere:**
   ```go
   // Don't use: line[scrollOffset:]
   // Use: visualSlice(line, scrollOffset, visibleWidth)
   ```

2. **Strip ANSI before slicing, re-apply after:**
   ```go
   stripped := stripANSI(line)
   sliced := stripped[scrollOffset : scrollOffset+visibleWidth]
   recolored := reapplyColors(sliced, originalLine)
   ```

3. **Limit scroll offset to actual content:**
   ```go
   maxOffset := max(0, visualWidth(line) - visibleWidth)
   if m.horizontalScrollOffset > maxOffset {
       m.horizontalScrollOffset = maxOffset
   }
   ```

4. **Handle emoji boundaries:**
   - Don't split emoji in the middle
   - Round scroll offset to nearest safe character boundary

### Option B: Remove Horizontal Scrolling ❌ Last Resort

**If unfixable or too complex:**

1. Revert the horizontal scrolling commits for Detail view
2. Document limitation: "Detail view requires minimum terminal width"
3. Suggest alternatives:
   - Use List view (F1) for narrow terminals
   - Use Tree view (F3) for narrow terminals
   - Rotate phone to landscape in Termux

**Trade-off:** Lose Termux narrow-screen support for Detail view, but maintain stability.

---

## Decision Criteria

### Keep & Fix If:
- ✅ Fix is straightforward (visual width functions)
- ✅ Corruption is limited to specific edge cases
- ✅ Detail view is frequently used in Termux
- ✅ Horizontal scroll adds significant value

### Remove If:
- ❌ Fix requires extensive refactoring
- ❌ Corruption affects core functionality
- ❌ Detail view rarely used in Termux anyway
- ❌ Alternative views (List/Tree) work well on narrow screens

---

## Questions to Answer

1. **When was horizontal scrolling added?**
   ```bash
   git log --all --grep="horizontal" --oneline
   git log --all --grep="scroll" --oneline
   git log --all --grep="Termux" --oneline
   ```

2. **What's the current implementation?**
   - Where is `horizontalScrollOffset` defined?
   - How is it applied in rendering?
   - What are the Left/Right arrow key handlers doing?

3. **Can it be fixed easily?**
   - Does TFE already have visual width slicing functions?
   - Can we reuse `visualWidth()` and `truncateToWidth()`?
   - How much code needs to change?

4. **How important is this feature?**
   - Do users actually use Detail view in Termux?
   - Are List/Tree views sufficient for narrow screens?
   - Is the complexity worth the benefit?

---

## Recommended Approach

**Session Plan:**

1. **Investigate (15 min)** - Find the code, understand the implementation
2. **Reproduce (10 min)** - Confirm the bug in narrow terminal
3. **Debug (20 min)** - Add logging, identify root cause
4. **Decide (5 min)** - Fix or remove?
5. **Implement (30 min)** - Either fix the visual width handling OR revert the feature
6. **Test (10 min)** - Verify fix works or removal doesn't break anything

**Total time estimate:** 90 minutes

---

## Files to Focus On

**Primary:**
- `render_file_list.go` - Detail view rendering with horizontal scroll
- `update_keyboard.go` - Left/Right arrow key handling
- `types.go` - Horizontal scroll offset variable

**Supporting:**
- `file_operations.go` - Visual width functions (`visualWidth`, `truncateToWidth`)
- `helpers.go` - Any scroll-related helper functions

**Reference:**
- `docs/EMOJI_DEBUG_SESSION_2.md` - Similar visual width debugging session
- `CHANGELOG.md` - Look for when horizontal scroll was added

---

## Success Criteria

### If Fixing:
- ✅ No text corruption when scrolling right
- ✅ Smooth scrolling throughout entire range
- ✅ No extra empty space after last column
- ✅ Emoji and ANSI codes remain intact
- ✅ Works in both Termux and narrow WezTerm

### If Removing:
- ✅ Detail view rendering restored to pre-scroll version
- ✅ No horizontal scroll keys active in Detail view
- ✅ Clear documentation of narrow-screen alternatives
- ✅ No regression in List/Tree views

---

## Related Issues to Check

- Any open issues about "horizontal scroll" or "text corruption"?
- Any TODOs in code comments related to scrolling?
- Any PLAN.md or BACKLOG.md items about this feature?

---

**Session Started:** [To be filled in next session]
**Branch:** Create `fix-horizontal-scroll` or `remove-horizontal-scroll` branch
**Status:** Investigation needed

---

## Quick Start Commands

```bash
# Start investigation
cd ~/projects/TFE
git log --all --grep="horizontal" --oneline
git log --all --grep="Detail.*scroll" --oneline

# Find the code
grep -rn "horizontalScrollOffset" .
grep -rn "horizontal.*scroll" render_file_list.go

# Create debug branch
git checkout -b fix-horizontal-scroll

# Test in narrow terminal
# (Resize WezTerm to ~60 columns or use Termux)
./tfe
# Press F2, then Right arrow repeatedly
```
