# Lessons Learned - TFE Development

This document captures hard-won lessons from debugging complex issues in TFE. Use this as a reference when adding new features or debugging similar problems.

---

## Visual Width vs Byte Length

### The Problem
**Never use `len()` or byte-based string slicing for text that will be displayed in a terminal.**

Characters have different **visual widths** (columns they occupy on screen) vs **byte lengths**:

| Character | Byte Length | Visual Width | Why |
|-----------|-------------|--------------|-----|
| `a` | 1 | 1 | ASCII character |
| `⭐` | 3 | 2 | Wide emoji (most emojis) |
| `📝` | 4 | 2 | Wide emoji with variation selector |
| `\033[38;5;220m` | 11 | 0 | ANSI escape code (color) |
| `\uFE0F` | 3 | 0 | Variation selector (invisible) |

### The Solution
**Always use visual-width-aware functions:**

```go
// ❌ WRONG - Uses byte length
if len(text) > maxWidth {
    text = text[:maxWidth]  // Splits ANSI codes, emojis
}

// ✅ RIGHT - Uses visual width
if visualWidth(text) > maxWidth {
    text = truncateToWidth(text, maxWidth)  // Preserves ANSI, emojis
}

// ❌ WRONG - Byte-based padding
line = fmt.Sprintf("%-*s", width, name)

// ✅ RIGHT - Visual-width padding
paddedName := m.padToVisualWidth(name, width)
line = fmt.Sprintf("%s", paddedName)
```

### Functions to Use
- **`visualWidth(s)`** - Get visual width (strips ANSI, counts emoji correctly)
- **`m.visualWidthCompensated(s)`** - Terminal-aware visual width (handles Windows Terminal emoji quirks)
- **`truncateToWidth(s, width)`** - Truncate to visual width (preserves ANSI codes)
- **`m.truncateToVisualWidth(s, width)`** - Terminal-aware truncation
- **`m.padToVisualWidth(s, width)`** - Pad string to exact visual width
- **`m.runeWidth(r)`** - Get visual width of single rune (terminal-aware)

### When This Matters
- Column alignment (headers vs data rows)
- Box rendering (content must fit exactly)
- Horizontal scrolling (extracting visible portions)
- Truncating long file names or paths
- Any fmt.Sprintf with width specifiers

---

## Terminal-Specific Rendering Differences

### The Problem
**Different terminals interpret lipgloss `Width()` and emoji rendering differently.**

| Terminal | Lipgloss Width() | Emoji + Variation Selector | Notes |
|----------|------------------|---------------------------|-------|
| **Windows Terminal** | Content width (borders added) | 2 cells | Width() excludes borders |
| **WezTerm** | Total width (borders included) | 1 cell | Width() includes borders |
| **Termux** | Total width (borders included) | 1 cell | Same behavior as WezTerm |
| **Kitty** | Total width (borders included) | 1 cell | Same behavior as WezTerm |
| **xterm** | Total width (borders included) | 1 cell | Native xterm, narrow emoji rendering |
| **xterm.js (default)** | Total width (borders included) | Inconsistent (1-2 cells) | ⚠️ Requires Unicode11 addon! |
| **xterm.js + Unicode11** | Total width (borders included) | 2 cells | ✅ Works like Windows Terminal |

### The Solution
**Use `m.terminalType` to apply terminal-specific adjustments:**

```go
// Box width calculation
availableWidth := m.width
if m.viewMode == viewDualPane {
    availableWidth = m.leftWidth - 6
} else {
    // Single-pane: account for lipgloss Width() differences
    if m.terminalType == terminalWezTerm {
        availableWidth = m.width - 8  // WezTerm/Termux: borders included
    } else {
        availableWidth = m.width - 6  // Windows Terminal: borders added
    }
}
```

### Terminal Detection
TFE automatically detects terminal type in `model.go`:
- Checks `$TERM_PROGRAM` for "WezTerm", "iTerm.app", etc.
- Checks `$WT_SESSION` for Windows Terminal
- Uses this for both emoji width AND box width calculations

### When to Check Terminal Type
- Lipgloss box width calculations
- Emoji width calculations (variation selectors)
- Any hardcoded width/padding values
- Terminal graphics protocol selection

### xterm.js Emoji Alignment

**Special Case: Web-Based Terminals Using xterm.js**

If you're embedding TFE in applications using xterm.js (VS Code terminal, web IDEs, custom terminal apps), emoji alignment requires special attention.

#### The Problem

xterm.js **without Unicode11 addon** renders emojis inconsistently:
- Base emojis (📦 🖼 🐹): ~2 cells (varies by emoji)
- Symbol emojis (⚙ ⬆): ~1 cell
- Result: Alignment off by 1 space per emoji

**Symptoms:**
- File list box borders misaligned
- Emojis stick out of menu bar brackets `[🔍]`
- Everything after emojis shifts 1 space left
- Favorite stars (⭐) overlap file icons

#### The Solution

**Install the Unicode11 addon** in your xterm.js application:

```typescript
import { Unicode11Addon } from '@xterm/addon-unicode11';

const unicode11Addon = new Unicode11Addon();
term.loadAddon(unicode11Addon);
term.unicode.activeVersion = '11';
```

**Why This Works:**
- Unicode11 addon makes ALL emojis render consistently as 2 cells
- Matches Windows Terminal, WezTerm, Termux behavior
- TFE's `runewidth` calculations align perfectly
- No special TFE configuration needed

**Result:**
- ✅ TFE detects as Windows Terminal (via `WT_SESSION`)
- ✅ Expects 2-cell emoji widths
- ✅ xterm.js renders 2-cell emoji widths
- ✅ Perfect alignment!

#### Alternative (Not Recommended)

If you **cannot** install Unicode11 addon, you can:
1. Filter `WT_SESSION` from PTY environment
2. TFE will detect as `xterm` instead
3. TFE will apply narrow emoji compensation

However, this approach is:
- Less reliable (xterm.js emoji width still inconsistent)
- Requires maintaining terminal-specific workarounds
- May break in future xterm.js versions

**Always prefer the Unicode11 addon approach for production use.**

#### Testing

After adding Unicode11 addon:
1. Restart your terminal application
2. Spawn a new terminal
3. Run `tfe`
4. Title bar should show `[Windows Terminal]`
5. Check emoji alignment in file list
6. Verify menu bar brackets properly contain emojis

#### Resources
- [xterm.js Unicode11 Addon](https://github.com/xtermjs/xterm.js/tree/master/addons/addon-unicode11)
- [Unicode East Asian Width](https://www.unicode.org/reports/tr11/)
- Real-world fix: [Opustrator project investigation](https://github.com/GGPrompts/TFE/issues/)

---

## ANSI Escape Code Handling

### The Problem
**ANSI escape codes (colors, styles) are invisible but have byte length.**

Slicing a string with ANSI codes can:
- Split an escape sequence mid-code → terminal corruption
- Count escape codes toward visual width → misalignment

### The Solution
**Use ANSI-aware functions that skip escape sequences:**

```go
// ❌ WRONG - Splits ANSI codes
visiblePart := styledLine[scrollOffset:scrollOffset+width]

// ✅ RIGHT - Preserves ANSI codes
visiblePart := m.extractVisiblePortion(styledLine, width)
```

### How extractVisiblePortion() Works
```go
func (m model) extractVisiblePortion(line string, viewWidth int) string {
    // 1. Walk through runes one by one
    // 2. Detect ANSI escape sequences (\x1b[...m)
    // 3. Don't count ANSI codes toward visual width
    // 4. Only add characters that fit in viewWidth
    // 5. Prepend active ANSI codes to result (preserve styling)
    // 6. Append \033[0m reset at end
}
```

### ANSI Code Patterns
- **Color codes**: `\033[38;5;220m` (foreground), `\033[48;5;235m` (background)
- **Style codes**: `\033[1m` (bold), `\033[3m` (italic)
- **Reset code**: `\033[0m` (clear all styling)
- **Detection**: Start with `\x1b` or `\033`, end with letter (usually `m`)

### When This Matters
- Horizontal scrolling (extracting visible portions)
- Truncating styled text
- Padding styled text to exact width
- Any string slicing on colored/styled text

---

## Consistent Scrolling Logic

### The Problem
**Headers and data rows must use IDENTICAL scrolling logic.**

When we had different scrolling implementations:
- Headers scrolled character-by-character ("Si" → "Siz" → "Size")
- Data rows scrolled as complete lines
- Headers appeared at wrong positions during scroll
- Column alignment broke progressively

### The Solution
**Use the same scrolling function for everything:**

```go
// ❌ WRONG - Different scrolling for headers
if m.isNarrowTerminal() {
    // Custom manual scrolling for header
    for _, r := range headerRunes {
        if visualCol >= scrollOffset && visualCol < scrollOffset+width {
            visibleHeader.WriteRune(r)
        }
    }
}

// ✅ RIGHT - Same scrolling for headers and data
// Don't scroll headers manually - let applyHorizontalScroll() handle it
headerLine := headerStyle.Render(header)
s.WriteString(headerLine)  // Added to result
// Later: applyHorizontalScroll() processes entire result (headers + data)
```

### The Single-Pass Principle
1. Build complete content (headers + data rows)
2. Apply styling/formatting
3. Apply horizontal scrolling **once** to entire result
4. Never scroll the same content twice

### When to Apply This
- Any scrollable view (Detail, List, Tree)
- Dual-pane preview scrolling
- Full-screen preview scrolling
- Menu scrolling (if implemented)

---

## Box Content Width Calculation

### The Problem
**Content width must EXACTLY match the lipgloss box internal width.**

If content is wider than box:
- Text wraps to next line
- Layout breaks (double-height rows)
- Vertical alignment fails

If content is narrower than box:
- Works, but wastes space

### The Solution
**Calculate availableWidth to match box internal width:**

```go
// In view.go - Box definition
fileListStyle := lipgloss.NewStyle().
    Width(m.width - 6).  // Box total width
    Border(lipgloss.RoundedBorder())  // Adds 2-char borders (terminal-dependent)

// In render_file_list.go - Content width
availableWidth := m.width
if m.viewMode == viewDualPane {
    availableWidth = m.leftWidth - 6
} else {
    // Must match box internal width
    if m.terminalType == terminalWezTerm {
        availableWidth = m.width - 8  // Width includes borders
    } else {
        availableWidth = m.width - 6  // Width excludes borders
    }
}
```

### Box Width Anatomy
```
Terminal width: 60 chars
├─ Margins: -6 chars (3 left + 3 right)
├─ Box total: 54 chars
   ├─ Left border: -1 char (if included in Width())
   ├─ Right border: -1 char (if included in Width())
   └─ Content area: 52 or 54 chars (terminal-dependent)
```

### When to Recalculate
- Window resize (update.go - window size msg)
- View mode change (single ↔ dual-pane)
- Display mode change (List/Detail/Tree)
- Any time `m.width` or `m.leftWidth` changes

---

## Scroll Bounds Checking

### The Problem
**Without bounds checking, users can scroll infinitely through empty space.**

This happens when:
- No maximum scroll offset calculated
- Scroll offset incremented without limits
- Content width not measured

### The Solution
**Always calculate and enforce scroll bounds:**

```go
// Calculate maximum scroll offset
maxScroll := renderWidth - availableWidth
if maxScroll < 0 {
    maxScroll = 0  // No scrolling needed
}

// Increment scroll offset
m.detailScrollX += 4

// Clamp to bounds
if m.detailScrollX > maxScroll {
    m.detailScrollX = maxScroll
}
if m.detailScrollX < 0 {
    m.detailScrollX = 0
}
```

### Scroll Bounds Formula
```
maxScroll = max(0, contentWidth - viewportWidth)

Where:
- contentWidth = renderWidth (width of all content)
- viewportWidth = availableWidth (visible area)
- maxScroll = 0 if content fits entirely in viewport
```

### When to Apply Bounds
- Every scroll increment (arrow keys, mouse wheel)
- After window resize (content may fit now)
- After display mode change (different widths)
- Page Up/Down scrolling

---

## Header vs Data Row Alignment

### The Problem
**Headers and data rows must be built with IDENTICAL formatting.**

Misalignment happens when:
- Different padding methods (%-*s vs padToVisualWidth)
- Different width calculations (len vs visualWidth)
- Different prefixes (headers missing "  " prefix)
- Different emoji handling (runeWidth vs RuneWidth)

### The Solution Checklist

**✅ Use same padding method:**
```go
// Headers
paddedNameHeader := m.padToVisualWidth(nameHeader, nameWidth)

// Data rows
paddedName := m.padToVisualWidth(name, nameWidth)
```

**✅ Use same width calculation:**
```go
// Headers
charWidth := m.runeWidth(r)

// Data rows
charWidth := m.runeWidth(r)
```

**✅ Use same prefix:**
```go
// Headers
header = "  " + header

// Data rows
s.WriteString("  ")
s.WriteString(line)
```

**✅ Use same scrolling logic:**
```go
// Don't scroll headers separately
// Let applyHorizontalScroll() handle everything
```

### Debugging Header Alignment
If headers don't align with columns:

1. **Check padding method** - Both using `padToVisualWidth()`?
2. **Check width calculation** - Both using `m.runeWidth()`?
3. **Check prefix** - Both have "  " prefix?
4. **Check scrolling** - Single-pass scrolling only?
5. **Check terminal type** - Same `availableWidth` calculation?

---

## Debugging Strategies

### Visual Width Issues
**Symptom:** Columns misaligned, text wraps unexpectedly, emoji positioning wrong

**Debug steps:**
1. Add debug output showing byte length vs visual width:
   ```go
   fmt.Fprintf(os.Stderr, "DEBUG: byteLen=%d visualWidth=%d text=%q\n",
       len(line), visualWidth(line), line)
   ```
2. Check for `len()` or `[:N]` slicing in rendering code
3. Look for `fmt.Sprintf("%-*s", ...)` on text with emojis
4. Verify using `m.runeWidth()` not `runewidth.RuneWidth()`

### Box Wrapping Issues
**Symptom:** Text wraps to next line, double-height rows, layout breaks

**Debug steps:**
1. Compare content width to box width:
   ```go
   fmt.Fprintf(os.Stderr, "DEBUG: contentWidth=%d boxWidth=%d terminal=%s\n",
       visualWidth(line), availableWidth, m.terminalType)
   ```
2. Check if `availableWidth` accounts for terminal type
3. Test in both WezTerm AND Windows Terminal
4. Verify box width calculation matches content calculation

### Scrolling Issues
**Symptom:** Headers offset, text corruption, empty space scrolling

**Debug steps:**
1. Check if content is scrolled multiple times
2. Verify headers use same scrolling as data rows
3. Add scroll bounds checking
4. Test with terminal width < 100 (narrow terminal)

### General Debugging
```bash
# Test in narrow terminal
export COLUMNS=60
./tfe

# Test in WezTerm specifically
# (Already auto-detects, but good to verify)

# Check for ANSI code corruption
./tfe 2>&1 | cat -A  # Shows all control characters

# Monitor stderr debug output
./tfe 2>debug.log
tail -f debug.log
```

---

## Architecture Principles

### 1. **Modular Design**
Don't add complex UI logic to `main.go`. Create dedicated modules:
- `render_*.go` - Rendering functions
- `update_*.go` - Event handlers
- `helpers.go` - Utility functions

### 2. **Single Responsibility**
Each rendering function should:
- Build content ONCE
- Apply styling ONCE
- Apply scrolling ONCE (via `applyHorizontalScroll`)

### 3. **Consistent Patterns**
If headers and data rows must align:
- Use same padding functions
- Use same width calculations
- Use same scrolling logic
- Use same prefix/spacing

### 4. **Terminal Awareness**
Always use `m.terminalType` for:
- Emoji width calculations
- Box width calculations
- Any hardcoded dimensions

### 5. **Test in Multiple Terminals**
Before marking a UI feature complete, test in:
- WezTerm (most common dev terminal)
- Windows Terminal (Windows users)
- Termux (mobile users)
- At narrow width (<100 cols)
- At wide width (>120 cols)

---

## Common Pitfalls to Avoid

### ❌ DON'T: Use byte-based string operations
```go
len(text)                  // ❌ Byte length
text[:n]                   // ❌ Byte slicing
fmt.Sprintf("%-*s", w, s)  // ❌ Byte padding
```

### ✅ DO: Use visual-width functions
```go
visualWidth(text)               // ✅ Visual width
truncateToWidth(text, n)        // ✅ Visual slicing
m.padToVisualWidth(s, w)        // ✅ Visual padding
```

### ❌ DON'T: Scroll content multiple times
```go
// Scroll headers manually
visibleHeader := extractPortion(header)
s.WriteString(visibleHeader)

// Scroll data rows manually
visibleData := extractPortion(data)
s.WriteString(visibleData)

// THEN scroll entire result again
result = applyHorizontalScroll(result)  // ❌ Double scroll!
```

### ✅ DO: Single-pass scrolling
```go
// Build content without scrolling
s.WriteString(header)
s.WriteString(data)

// Scroll once at the end
result := s.String()
result = applyHorizontalScroll(result)  // ✅ Single pass
```

### ❌ DON'T: Ignore terminal differences
```go
availableWidth := m.width - 6  // ❌ Same for all terminals
```

### ✅ DO: Use terminal-specific calculations
```go
if m.terminalType == terminalWezTerm {
    availableWidth = m.width - 8  // ✅ WezTerm/Termux
} else {
    availableWidth = m.width - 6  // ✅ Windows Terminal
}
```

### ❌ DON'T: Forget scroll bounds
```go
m.scrollX += 4  // ❌ Infinite scrolling
```

### ✅ DO: Clamp scroll offset
```go
m.scrollX += 4
maxScroll := contentWidth - viewWidth
if m.scrollX > maxScroll {
    m.scrollX = maxScroll  // ✅ Bounded
}
```

---

## Quick Reference

### Width Functions Priority
1. **`m.runeWidth(r)`** - For character-by-character walking (terminal-aware)
2. **`m.visualWidthCompensated(s)`** - For measuring final rendered width
3. **`visualWidth(s)`** - For measuring plain text (strips ANSI)
4. **`truncateToWidth(s, w)`** - For truncating with ANSI preservation
5. **`m.padToVisualWidth(s, w)`** - For exact-width padding

### Terminal Type Checks
```go
if m.terminalType == terminalWezTerm {
    // WezTerm, Termux, Kitty, most Unix terminals
}
if m.terminalType == terminalWindowsTerminal {
    // Windows Terminal only
}
```

### Scroll Bounds Pattern
```go
maxScroll := max(0, contentWidth - viewportWidth)
m.scrollOffset += increment
m.scrollOffset = clamp(m.scrollOffset, 0, maxScroll)
```

### Box Width Pattern
```go
// In view.go
boxStyle := lipgloss.NewStyle().Width(m.width - 6).Border(...)

// In render function
if m.terminalType == terminalWezTerm {
    availableWidth = m.width - 8  // Borders included in Width()
} else {
    availableWidth = m.width - 6  // Borders added to Width()
}
```

---

## Related Issues

- **Emoji Debug Session #1** - `docs/EMOJI_DEBUG_SESSION.md` - Initial emoji width investigation
- **Emoji Debug Session #2** - `docs/EMOJI_DEBUG_SESSION_2.md` - Comprehensive emoji width fix
- **Horizontal Scroll Investigation** - `docs/NEXT_SESSION.md` - This debugging session

---

## Mobile Terminal UX

### The Problem
**Dual-pane layouts don't work well on narrow terminals, even with vertical stacking.**

On mobile devices (Termux, phones):
- Screen width is limited (~60 columns)
- Vertical space is precious (keyboard takes 40% of screen)
- Vertical-split dual-pane means:
  - File list cramped in top half
  - Preview cramped in bottom half
  - Keyboard eats even more vertical space
  - Can't see enough of either pane to be useful

### The Solution
**Default to single-pane mode on narrow terminals (width < 100).**

Single-pane workflow on mobile:
1. Full screen shows file list
2. Press Enter on file → Full screen preview
3. Press Esc → Back to full screen file list

This makes better use of limited screen real estate by dedicating the full screen to one thing at a time.

### Implementation
```go
// In model.go - initialModel()
if m.width >= 100 {
    m.viewMode = viewDualPane  // Wide terminals get side-by-side
}
// else: viewSinglePane (narrow terminals get full-screen toggle)
```

### When This Matters
- Default view mode on startup
- Responsive layout changes on window resize
- Any feature that assumes screen real estate
- Mobile-specific UX decisions

### Testing on Mobile
Always test new UI features in Termux:
- Portrait mode (~60 cols × 30 rows)
- With keyboard up (~60 cols × 15 rows)
- Landscape mode (~120 cols × 20 rows)

---

## Future Considerations

When adding new features, ask:

1. **Does this involve text rendering?** → Use visual width functions
2. **Does this involve boxes/borders?** → Check terminal type
3. **Does this involve scrolling?** → Single-pass scrolling, bounds checking
4. **Does this involve alignment?** → Same padding/width for all columns
5. **Does this involve emojis?** → Use `m.runeWidth()`, test in multiple terminals

---

**Remember:** Terminal UI is tricky. Visual width ≠ byte length. ANSI codes are invisible. Terminals are inconsistent. Test early, test often, test in multiple terminals.
