# Debug Prompt for Dual-Pane Alignment Issues

## Problem Description
TFE (Terminal File Explorer) has lag and misaligned panel heights in dual-pane mode, specifically with prompty files. The panels become uneven heights and the app lags when scrolling/navigating.

**Suspected Issue**: The separator line (`────────`) under headers may be wrapping on narrow panels, causing height miscalculations.

## Context
- **Dual-pane mode**: Accordion-style panels where focused pane gets 2/3 width, unfocused gets 1/3
- **Prompty files**: Markdown templates with {{VARIABLES}} that trigger Glamour rendering + input fields
- **Lag occurs**: Specifically in dual-pane with prompty files, NOT in single-pane mode
- **Alignment issue**: Panels become different heights, causing visual distortion

## Code Locations to Analyze

### 1. Dual-Pane Rendering Entry Point
**File**: `render_preview.go`, function `renderDualPane()` starting around line 938

Key areas:
- Line 1171-1178: Header line counting and maxVisible calculation
- Line 1335-1353: Lipgloss pane height settings (`.Height(contentHeight)`)

### 2. Prompty Preview Rendering
**File**: `render_preview.go`, function `renderPreview()` around line 493

Key areas:
- Line 520-585: Header rendering with separator line
- Line 611-648: Glamour caching logic
- Line 663-682: Input fields height calculation
- Line 764-776: Padding to maxVisible lines

### 3. File List Rendering
**File**: `render_file_list.go`
- `renderListView()`: Line 36
- `renderDetailView()`: Line 182
- `renderTreeView()`: Line 867

All have padding logic at the end (search for "Pad with empty lines")

## Specific Questions to Answer

1. **Separator Line Wrapping**:
   - In `renderPreview()` around line 584, there's: `separatorStyle.Render("─────────────────────────────────────")`
   - Could this line wrap in narrow accordion panels (1/3 width)?
   - Is the separator width fixed or dynamic?
   - Does wrapping add an extra line that throws off `maxVisible` calculations?

2. **Header Line Counting**:
   - In `renderDualPane()` line 1172: `headerLines := strings.Count(headerContent, "\n")`
   - Does this correctly count lines when toolbar wraps?
   - Is the toolbar rendering included in `headerContent` before counting?

3. **Prompty-Specific Height Calculation**:
   - Lines 663-682: Input fields section takes 1/3 of `maxVisible`
   - Is `contentHeight` correctly calculated when input fields are active?
   - Could the header (lines 520-585) be variable height and not accounted for?

4. **Glamour Re-rendering**:
   - Lines 623-648: Cache checking logic
   - Cache is valid if `widthDiff <= 10` characters
   - In accordion mode, does switching focus trigger width changes > 10?
   - Could cache misses cause re-rendering every frame?

5. **Lipgloss Height Enforcement**:
   - Line 1337, 1343: `.Height(contentHeight)` sets exact height
   - If `renderPreview()` returns fewer lines than expected, Lipgloss pads
   - If it returns more lines, Lipgloss truncates
   - Could padding logic (lines 764-776) be off-by-one?

## Debug Tasks

Please analyze the code and answer:

1. **Trace the flow**: When viewing a prompty file in dual-pane accordion mode (focused pane = 2/3 width):
   - What is `availableWidth` for the preview pane?
   - What is the width of the separator line?
   - Does it wrap? If so, how many lines does it become?

2. **Count actual lines rendered** in `renderPreview()` for a typical prompty file with input fields active:
   - Header lines (name, description, variables)
   - Separator line
   - Content lines
   - Input fields section
   - Padding lines
   - Total = should equal `maxVisible`

3. **Check for wrapping issues**:
   - Search for all hardcoded strings longer than ~30 characters
   - Could any wrap in a narrow (1/3 width) pane?
   - Are there any fixed-width elements not respecting `availableWidth`?

4. **Verify padding math**:
   - In `renderPreview()` line 764-776
   - In `renderListView()` line 165-178
   - In `renderDetailView()` line 576-590
   - In `renderTreeView()` line 1070-1083
   - Do all correctly calculate `linesRendered` before padding?

5. **Look for off-by-one errors**:
   - Header separator (blank line before/after?)
   - Input fields separator (line 730-738)
   - Any double-counting of newlines?

## Output Format

Please provide:

1. **Root Cause**: What is causing the height mismatch?
2. **Line Numbers**: Exact locations of the bug(s)
3. **Fix Suggestion**: What needs to change (pseudocode is fine)
4. **Verification**: How to confirm the fix works

## Files to Examine
- `render_preview.go` (main focus)
- `render_file_list.go`
- `file_operations.go` (if related to prompt loading)

## Additional Notes
- Recent fixes already applied:
  - Dynamic header line counting (line 1172)
  - Padding logic in all render functions
  - Glamour cache tolerance (±10 chars)
- Issue persists despite these fixes
- Likely a subtle width/wrapping issue in narrow accordion panes
