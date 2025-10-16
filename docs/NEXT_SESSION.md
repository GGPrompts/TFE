# Next Session: Quick Wins - Syntax Highlighting & UI Polish

## Objective

Implement three high-impact, low-effort improvements to TFE:
1. **Syntax highlighting** for code files (Chroma v2.14.0 already installed!)
2. **Adaptive colors** for light/dark terminal compatibility
3. **Rounded borders** for modern UI polish

**Total Estimated Time:** 5-7 hours
**Impact:** Major visual upgrade with minimal code changes

---

## Quick Start Prompt

```
Hi! I need to implement three quick wins for TFE based on recent research:

1. SYNTAX HIGHLIGHTING - Chroma v2.14.0 is already installed in go.mod but not being used!
   Add syntax highlighting to file previews for code files (100+ languages supported).

2. ADAPTIVE COLORS - Replace hardcoded colors with lipgloss.AdaptiveColor to work
   in both light and dark terminals.

3. ROUNDED BORDERS - Update preview pane to use lipgloss.RoundedBorder() for a
   modern appearance.

Please implement these in order (syntax highlighting first, it has the biggest impact).

The relevant files are:
- file_operations.go (add syntax highlighting)
- render_preview.go (integrate highlighted code)
- types.go (add isSyntaxHighlighted field)
- styles.go (adaptive colors)
- view.go, render_preview.go (rounded borders)

Implementation details are in docs/NEXT_SESSION.md sections below.
```

---

## Task 1: Syntax Highlighting with Chroma

### Status
✅ **Chroma v2.14.0 already installed** (see go.mod line 8)
❌ Not currently being used
🎯 Just needs to be integrated into file preview

### Implementation Steps

**Step 1: Add syntax highlighting function to `file_operations.go`**

Add this import:
```go
import (
    "bytes"
    "github.com/alecthomas/chroma/v2/quick"
    "github.com/alecthomas/chroma/v2/formatters"
    "github.com/alecthomas/chroma/v2/lexers"
    "github.com/alecthomas/chroma/v2/styles"
)
```

Add this function after `isBinaryFile()`:
```go
// highlightCode applies syntax highlighting to code files using Chroma
// Returns highlighted content and success status
func highlightCode(content, filepath string) (string, bool) {
    var buf bytes.Buffer

    // Try to determine lexer from filename
    lexer := lexers.Match(filepath)
    if lexer == nil {
        // Fallback: analyze content
        lexer = lexers.Analyse(content)
    }
    if lexer == nil {
        // Still nothing, use fallback plain text
        return "", false
    }

    // Configure lexer
    lexer = chroma.Coalesce(lexer)

    // Use terminal256 formatter for better color support
    formatter := formatters.Get("terminal256")
    if formatter == nil {
        formatter = formatters.Fallback
    }

    // Use monokai style (works well in dark terminals)
    // Alternative styles: dracula, vim, github, solarized-dark
    style := styles.Get("monokai")
    if style == nil {
        style = styles.Fallback
    }

    // Tokenize and format
    iterator, err := lexer.Tokenise(nil, content)
    if err != nil {
        return "", false
    }

    err = formatter.Format(&buf, style, iterator)
    if err != nil {
        return "", false
    }

    return buf.String(), true
}
```

**Step 2: Integrate into `loadPreview()` function**

Find the section in `loadPreview()` that loads text files (around line 725-737), and modify:

```go
// Current code (around line 725):
// Split into lines for regular text files
lines := strings.Split(string(content), "\n")

// Replace with:
// Try syntax highlighting for code files
highlighted, ok := highlightCode(string(content), path)
var lines []string

if ok {
    // Syntax highlighting succeeded
    lines = strings.Split(highlighted, "\n")
    m.preview.isSyntaxHighlighted = true
} else {
    // Fallback to plain text
    lines = strings.Split(string(content), "\n")
    m.preview.isSyntaxHighlighted = false
}
```

**Step 3: Add field to `previewModel` in `types.go`**

In the `previewModel` struct (around line 76-93), add:
```go
type previewModel struct {
    filePath   string
    fileName   string
    content    []string
    scrollPos  int
    maxPreview int
    loaded     bool
    isBinary   bool
    tooLarge   bool
    fileSize   int64
    isMarkdown bool
    isSyntaxHighlighted bool  // ← ADD THIS
    // ... rest of fields
}
```

**Step 4: Reset flag when loading new file**

In `loadPreview()`, around line 655-661, add:
```go
m.preview.scrollPos = 0
m.preview.loaded = false
m.preview.isBinary = false
m.preview.tooLarge = false
m.preview.isMarkdown = false
m.preview.isSyntaxHighlighted = false  // ← ADD THIS
```

### Testing Checklist

After implementation:
- [ ] Open a .go file - should have syntax highlighting
- [ ] Open a .py file - should have syntax highlighting
- [ ] Open a .js file - should have syntax highlighting
- [ ] Open a .json file - should have syntax highlighting
- [ ] Open a .md file - should use Glamour (markdown rendering)
- [ ] Open a .txt file - should show plain text
- [ ] Open a binary file - should show binary warning
- [ ] Test in detail view (F3)
- [ ] Test in dual-pane mode (Space/Tab)
- [ ] Test in full-screen preview (Enter/F3)
- [ ] Verify no performance regression

### Expected Outcome

Beautiful syntax-highlighted code previews with:
- Color-coded keywords, strings, comments, functions
- Language-specific highlighting for 100+ languages
- Automatic fallback to plain text for unknown formats
- Works in all preview modes

---

## Task 2: Adaptive Colors

### Current Issue
TFE uses hardcoded colors that may look poor in light terminals.

### Implementation

**Update `styles.go` (entire file):**

Replace hardcoded colors with adaptive colors:

```go
package main

import (
	"github.com/charmbracelet/lipgloss"
)

// Adaptive color definitions - work in both light and dark terminals
var (
	// Title bar styling
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{
			Light: "#0087d7", // Dark blue for light backgrounds
			Dark:  "#5fd7ff", // Bright cyan for dark backgrounds
		})

	// Path display styling
	pathStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{
			Light: "#666666", // Medium gray for light
			Dark:  "#999999", // Light gray for dark
		})

	// Status bar styling
	statusStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{
			Light: "#444444",
			Dark:  "#AAAAAA",
		})

	// Selected item styling
	selectedStyle = lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.AdaptiveColor{
			Light: "#0087d7", // Dark blue background for light
			Dark:  "#00d7ff", // Bright cyan background for dark
		}).
		Foreground(lipgloss.AdaptiveColor{
			Light: "#FFFFFF", // White text on dark blue
			Dark:  "#000000", // Black text on bright cyan
		})

	// Folder styling
	folderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{
			Light: "#005faf", // Dark blue for light
			Dark:  "#5fd7ff", // Bright cyan for dark
		})

	// File styling
	fileStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{
			Light: "#000000", // Black for light
			Dark:  "#FFFFFF", // White for dark
		})

	// Claude context file styling (orange)
	claudeContextStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{
			Light: "#D75F00", // Darker orange for light
			Dark:  "#FF8700", // Bright orange for dark
		})
)
```

### Testing Checklist

- [ ] Test in dark terminal (should look like before or better)
- [ ] Test in light terminal (should be readable!)
- [ ] Selected items have good contrast
- [ ] Folders are clearly distinct from files
- [ ] Claude context files stand out
- [ ] Status bar is readable
- [ ] Build succeeds with no warnings

### Expected Outcome

TFE looks professional in both light and dark terminals with no configuration needed.

---

## Task 3: Rounded Borders

### Current State
Preview pane uses default borders or no borders.

### Implementation

**Update `render_preview.go`:**

Find the preview pane rendering (around line 96-116 in `renderDualPane()`):

```go
// Current code:
previewStyle := lipgloss.NewStyle().
    Width(m.rightWidth).
    Height(visibleLines)

// Replace with:
previewStyle := lipgloss.NewStyle().
    Width(m.rightWidth).
    Height(visibleLines).
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.AdaptiveColor{
        Light: "#CCCCCC", // Light gray border for light terminals
        Dark:  "#444444", // Dark gray border for dark terminals
    }).
    Padding(0, 1) // Add horizontal padding inside border
```

**Optional: Update full-screen preview borders**

In `renderFullPreview()` (if borders are used):
```go
previewContainer := lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.AdaptiveColor{
        Light: "#999999",
        Dark:  "#666666",
    }).
    Padding(1)
```

### Testing Checklist

- [ ] Preview pane has rounded corners
- [ ] Borders visible in dual-pane mode
- [ ] Borders don't break layout
- [ ] Full-screen preview looks good
- [ ] Works with syntax highlighting
- [ ] Works with markdown rendering

### Expected Outcome

Modern, polished UI with rounded corners instead of sharp 90-degree corners.

---

## Build & Test

After implementing all three tasks:

```bash
# Build
go build

# Run and test
./tfe

# Test various file types:
# - Navigate to a .go file (syntax highlighting)
# - Navigate to a .py file (syntax highlighting)
# - Navigate to a .md file (markdown rendering)
# - Toggle dual-pane (Space/Tab)
# - Enter full-screen preview (Enter/F3)

# Test in different terminals:
# - Dark terminal (your current setup)
# - Light terminal (if available)
```

---

## Documentation Updates

After successful implementation:

1. **Update CHANGELOG.md** (Unreleased section):
```markdown
### Added
- **Syntax Highlighting for Code Files**
  - Powered by Chroma v2.14.0 (100+ languages supported)
  - Automatic language detection from file extension
  - Color-coded keywords, strings, comments, functions
  - Works in all preview modes (single-pane, dual-pane, full-screen)
  - Fallback to plain text for unknown file types
- **Adaptive Colors**
  - Automatic adaptation to light and dark terminals
  - Professional appearance without configuration
  - Better readability across different terminal themes
- **Rounded Borders**
  - Modern UI with rounded corners for preview pane
  - Adaptive border colors for light/dark terminals
```

2. **Update PLAN.md** (mark Phase 2.2 complete if it mentions syntax highlighting)

---

## Troubleshooting

### Chroma Import Errors
If you get import errors:
```bash
go mod tidy
go build
```

### Syntax Highlighting Not Working
- Check file extension is recognized: add debug log showing lexer name
- Verify Chroma can tokenize: check error returns
- Test with known file type like .go first

### Colors Look Wrong
- Verify adaptive color syntax matches lipgloss v1.1.1 format
- Try different color values if needed
- Test in actual light/dark terminals

### Border Layout Issues
- Adjust padding values if content is cut off
- Check width calculations account for border (2 chars)
- Verify border doesn't exceed terminal width

---

## Expected Results Summary

**Before:**
- Plain text previews only
- Hardcoded colors (may look bad in light terminals)
- Sharp corners or no borders

**After:**
- ✨ Beautiful syntax highlighting for code files
- 🎨 Colors adapt to terminal background automatically
- 🎯 Modern, polished UI with rounded borders
- 📈 Professional appearance with minimal code changes

**Total Code Changes:**
- `file_operations.go`: +50 lines (syntax highlighting function)
- `types.go`: +1 line (isSyntaxHighlighted field)
- `styles.go`: Complete rewrite with adaptive colors (~35 lines)
- `render_preview.go`: +10 lines (rounded borders)
- **Total: ~100 lines of code for massive UX improvement**

---

## Previously Completed

### ✅ Markdown Scrolling Fixed
**Issue:** ANSI sequences bleeding into command prompt causing formatting issues
**Solution:** Made command prompt a selectable panel
**Status:** Resolved - no longer an issue

### ✅ Clickable Column Headers with Sorting
**Features:** Click headers to sort, visual indicators (↑↓), smart folder grouping
**Files Modified:** `update_mouse.go`, `file_operations.go`, `render_file_list.go`
**Commit:** e374f5f

### ✅ Mouse Click Detection Fix
**Issue:** Clicks misaligned when favorites filter active
**Solution:** Use getFilteredFiles() in click detection
**Commit:** c11ea41

---

## Current File Status

```
main.go: 21 lines ✅
styles.go: 35 lines → Will be rewritten with adaptive colors
helpers.go: 69 lines ✅
model.go: 78 lines ✅
update.go: 104 lines ✅
command.go: 127 lines ✅
dialog.go: 141 lines ✅
favorites.go: 150 lines ✅
editor.go: 156 lines ✅
types.go: 173 lines → +1 line for syntax highlight flag
view.go: 198 lines ✅
context_menu.go: 318 lines ✅
render_file_list.go: 477 lines ✅
render_preview.go: 498 lines → +10 lines for rounded borders
update_keyboard.go: 730 lines ✅
update_mouse.go: 502 lines ✅
file_operations.go: 885 lines → +50 lines for syntax highlighting
```

**Architecture Status:** ✅ All modules under control, modular architecture maintained

---

## Next Steps After This Session

Once these three quick wins are complete:

**Priority 1: More Quick Wins**
- Loading spinners for slow operations (spinner already in model!)
- Better Glamour markdown integration (auto-style, emoji support)

**Priority 2: Medium Enhancements**
- Huh forms for rename/search dialogs
- Bubbles list component migration
- Enhanced mouse support

**Priority 3: Advanced Features**
- Context visualizer (show Claude Code context)
- Plugin system
- Multiple panes/tabs

---

**Last Updated:** 2025-10-16
**Ready for Implementation:** ✅ All details provided above
