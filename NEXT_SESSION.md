# Next Session: Enhanced Full-Screen Viewer

## Context
TFE is a terminal file explorer built with Go + Bubbletea. We've just completed MC-style command prompt implementation where typing always goes to the command prompt (no focus required).

**Current Status:**
- ✅ Phases 1, 1.5, 2 complete (file browsing, view modes, dual-pane, preview)
- ✅ MC-style command prompt (just completed)
- ✅ External editor integration (Micro/nano)
- ✅ Claude context files highlighted in orange

**Decision:** We're skipping the Context Visualizer (Phase 3) since Claude Code already has `/context` for detailed breakdown. Instead, we're focusing on making TFE a great file viewer + quick launcher.

## Goal for Next Session

**Enhance the full-screen file viewer with:**

1. **Line wrapping for all files**
   - Currently lines truncate with "..." which is bad for reading
   - Wrap text naturally at terminal width
   - No more horizontal scrolling needed

2. **Beautiful markdown rendering**
   - Use [Glamour](https://github.com/charmbracelet/glamour) for markdown files
   - Renders headers, bold, italic, lists, code blocks
   - Syntax highlighting in code blocks
   - Hyperlinks work automatically (OSC 8 sequences → terminal handles clicking)
   - Glamour is made by Charm (same as Bubbletea/Lipgloss), so styling will match

3. **Smart file detection**
   - Detect `.md` files → use Glamour
   - All other text files → wrap at terminal width
   - Keep binary/large file detection as-is

4. **Optional: Line numbers**
   - Maybe hide line numbers when wrapping is enabled?
   - Or show line numbers only for unwrapped content (code files)?
   - Decide based on what looks better

## Implementation Steps

### Step 1: Add Glamour dependency
```bash
go get github.com/charmbracelet/glamour
```

### Step 2: Update file_operations.go
Add markdown detection to `loadPreview()`:
```go
// Detect if file is markdown
func isMarkdownFile(path string) bool {
    ext := strings.ToLower(filepath.Ext(path))
    return ext == ".md" || ext == ".markdown"
}
```

### Step 3: Update render_preview.go
Modify `renderPreview()` and `renderFullPreview()`:
- For markdown files: render with Glamour
- For other files: wrap lines at available width
- Remove "..." truncation, use proper wrapping

### Step 4: Handle wrapping in preview
- Calculate available width (terminal width - padding)
- Use a wrapping function for long lines
- May need to adjust scroll calculations (wrapped lines count as multiple display lines)

## Files to Modify

1. **go.mod** - Add glamour dependency
2. **file_operations.go** - Add `isMarkdownFile()` helper, update `loadPreview()` to detect markdown
3. **render_preview.go** - Update rendering logic for wrapping + markdown
4. **types.go** - Maybe add `isMarkdown bool` to `previewModel` struct?

## Key Decisions

- **Keep it simple:** View-only, not editing (press E for Micro)
- **No custom markdown features:** Just use Glamour defaults (they're beautiful)
- **Hyperlinks work for free:** Terminals handle OSC 8 sequences automatically
- **Line wrapping everywhere:** Reading is more important than seeing exact line breaks

## Testing Checklist

After implementation:
- [ ] View PLAN.md - should render beautifully with headers, lists, code blocks
- [ ] View README.md - test hyperlinks (Ctrl+Click should work in Windows Terminal)
- [ ] View a .go file - should wrap long lines
- [ ] View a long text file - verify wrapping works correctly
- [ ] Scroll through wrapped content - ensure scroll calculations work
- [ ] Test in dual-pane preview - wrapping should work there too
- [ ] Test full-screen preview - wrapping should work there too

## Reference Files

Key files in the codebase:
- **render_preview.go:11-76** - `renderPreview()` function (handles line truncation currently)
- **render_preview.go:108-146** - `renderFullPreview()` function
- **file_operations.go:114-165** - `loadPreview()` function (loads file content)
- **types.go:55-63** - `previewModel` struct

## Expected Outcome

After this session:
1. Viewing markdown files in TFE shows beautiful formatted output
2. All files have proper line wrapping (no more "...")
3. TFE becomes a great viewer for documentation and code
4. Quick workflow: browse with TFE, view with TFE, edit with Micro (press E)

## Why This Matters

TFE's value proposition:
- **Fast browsing** - Multiple view modes, dual-pane
- **Beautiful viewing** - Markdown rendering, proper wrapping
- **Quick editing** - One keypress to Micro
- **Simple** - Does one thing well, stays out of your way

This is much more focused than the original plan with Context Visualizer, file operations, etc. Keep it simple and fast!
