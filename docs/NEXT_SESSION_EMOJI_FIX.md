# Next Session - Apply Emoji Width Fix to Markdown and All Rendering

## Context

In the previous session, emoji width issues were fixed in the file tree rendering. The same fix pattern needs to be applied to markdown files and anywhere else emojis are rendered.

## Prompt for Claude Code

```
I previously fixed emoji width alignment issues in the file tree by using consistent width calculation functions. The same issue affects markdown files and potentially other rendering locations.

BACKGROUND:
- Read docs/LESSONS_LEARNED.md - "Visual Width vs Byte Length" section
- Previous fix used m.runeWidth() and m.visualWidthCompensated() instead of len() or runewidth.RuneWidth()
- Terminal-specific emoji handling (WezTerm vs Windows Terminal variation selectors)

TASK 1: AUDIT
Find all locations where emojis might cause width calculation issues:

1. Search for emoji rendering in:
   - Markdown preview (render_preview.go)
   - Prompt file preview (render_preview.go)
   - Any text wrapping or padding functions
   - Status messages, dialog boxes, menu items
   - Any place that uses fmt.Sprintf with %-*s on text that might contain emojis

2. For each location, check if it uses:
   - ❌ len(text) - WRONG
   - ❌ text[:n] - WRONG
   - ❌ runewidth.RuneWidth() directly - WRONG (misses terminal-specific handling)
   - ❌ fmt.Sprintf("%-*s", width, text) - WRONG for emoji text
   - ✅ visualWidth(text) - CORRECT
   - ✅ m.visualWidthCompensated(text) - CORRECT (terminal-aware)
   - ✅ m.runeWidth(r) - CORRECT (terminal-aware, per-rune)
   - ✅ truncateToWidth(text, width) - CORRECT
   - ✅ m.padToVisualWidth(text, width) - CORRECT

TASK 2: FIX MARKDOWN RENDERING
Focus on markdown preview specifically:

1. Check render_preview.go for:
   - Line wrapping logic (probably uses lipgloss/wrap)
   - Width calculations for markdown content
   - Padding/truncation of markdown lines
   - Any hardcoded width values

2. Check file_operations.go for:
   - Markdown content loading (loadPreview for .md files)
   - Width calculations when preparing markdown for Glamour rendering
   - Any text processing before rendering

3. Apply the same pattern as file tree fix:
   - Use m.visualWidthCompensated() for measuring rendered width
   - Use m.runeWidth() when walking through characters
   - Never use len() or direct runewidth.RuneWidth() for display text

TASK 3: TEST CASES
Create test cases to verify the fix:

1. Create test.md with:
   ```markdown
   # Test Emoji Width Alignment

   Regular text without emojis
   Text with emojis: 📝 🌐 ⭐ 🎉 🚀 should align properly
   Mixed: Regular 📝 text 🌐 with 🎉 emojis 🚀 scattered

   | Column 1 📝 | Column 2 🌐 | Column 3 ⭐ |
   |-------------|-------------|-------------|
   | Data        | Data        | Data        |
   ```

2. Test in multiple terminals:
   - WezTerm (emoji + variation selector = 1 cell)
   - Windows Terminal (emoji + variation selector = 2 cells)
   - Termux (1 cell, like WezTerm)

3. Verify:
   - No text wrapping at wrong positions
   - Columns stay aligned
   - Emoji spacing consistent
   - No extra/missing spaces around emojis

TASK 4: APPLY TO OTHER LOCATIONS
Check these other rendering areas:

1. **Dialog boxes** (dialog.go)
   - Input dialogs, confirm dialogs
   - Width calculations for dialog content

2. **Status messages** (view.go, render_preview.go)
   - Status bar rendering
   - Temporary messages

3. **Menu rendering** (menu.go)
   - Menu item text with emojis
   - Menu width calculations

4. **Context menu** (context_menu.go)
   - Menu item alignment
   - Width calculations

5. **Command prompt** (view.go)
   - Command history display
   - Input field width

TASK 5: DOCUMENT THE CHANGES
1. Update CHANGELOG.md with:
   - What was fixed (emoji width in markdown/other areas)
   - Which files were modified
   - Reference to LESSONS_LEARNED.md

2. If you find new patterns or edge cases:
   - Add them to docs/LESSONS_LEARNED.md
   - Document terminal-specific quirks

3. Create summary:
   - List all locations fixed
   - Any remaining locations that might need attention
   - Testing results

EXPECTED OUTPUT:
- All emoji width issues fixed consistently
- Markdown files render with proper alignment
- All text rendering uses visual-width-aware functions
- No len() or direct runewidth.RuneWidth() for display text
- Comprehensive test results
- Updated documentation

REFERENCE FILES:
- docs/LESSONS_LEARNED.md - Visual Width vs Byte Length section
- docs/EMOJI_DEBUG_SESSION.md - Previous emoji fix details (if exists)
- docs/EMOJI_DEBUG_SESSION_2.md - Previous emoji fix details (if exists)
- render_file_list.go - Example of correct emoji width handling (file tree fix)

START BY:
1. Reading LESSONS_LEARNED.md to understand the pattern
2. Auditing all rendering code for emoji width issues
3. Prioritizing markdown preview (highest impact)
4. Testing thoroughly in multiple terminals
5. Documenting all changes
```

## Quick Start Commands

```bash
cd ~/projects/TFE

# Find potential issues
grep -rn "len(.*)" render_*.go file_operations.go | grep -v "len(files)" | grep -v "len(items)"
grep -rn "runewidth.RuneWidth" *.go | grep -v "terminal_graphics.go"
grep -rn 'fmt.Sprintf.*%-\*s' *.go

# Read the lesson
cat docs/LESSONS_LEARNED.md | grep -A 50 "Visual Width"

# Check current markdown rendering
vim render_preview.go
# Search for /markdown or /glamour

# Create test file
cat > test-emoji-width.md << 'EOF'
# Emoji Width Test

Regular text line
Line with emojis: 📝 🌐 ⭐ 🎉 🚀
Mixed text and 📝 emojis 🌐 scattered

| Header 📝 | Header 🌐 | Header ⭐ |
|-----------|-----------|-----------|
| Data 1    | Data 2    | Data 3    |
EOF

# Test current behavior
./tfe test-emoji-width.md
# Press Enter to preview, check for alignment issues
```

## Success Criteria

- ✅ Markdown files with emojis render with correct alignment
- ✅ No text wrapping at wrong positions due to emoji width miscalculation
- ✅ Table columns with emoji headers stay aligned
- ✅ All rendering code uses visual-width-aware functions
- ✅ Works correctly in WezTerm, Windows Terminal, and Termux
- ✅ CHANGELOG.md updated
- ✅ LESSONS_LEARNED.md updated if new patterns discovered

## Estimated Time

- Audit: 15-20 min
- Fix markdown: 30-40 min
- Fix other locations: 20-30 min
- Testing: 15-20 min
- Documentation: 10-15 min

**Total: 90-125 minutes**
