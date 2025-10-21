# Next Session Tasks

## 🐛 TODO: Bug Fixes for Preview & Layout

### 1. Prompty File Preview Causing Height Mismatch

**Problem:**
When viewing prompty files (prompt templates with fillable fields) in a narrow window, the file tree/preview pane heights become mismatched and sometimes the app headers get hidden. This suggests incorrect width calculations or text wrapping issues when rendering fillable input fields.

**Where to Look:**
- `render_preview.go` - Prompty rendering with fillable fields
- `prompt_parser.go` - Template parsing and field rendering
- Width calculations for input fields in narrow windows
- Height calculations that might be getting thrown off by wrapped text

**Expected Behavior:**
- File tree and preview pane should always have matching heights
- App headers should never be hidden
- Input fields should wrap properly in narrow windows
- Layout should remain stable regardless of window width

**Test Case:**
1. Open a `.prompty` file with multiple `{{VARIABLES}}`
2. Resize terminal to narrow width (e.g., 80 columns)
3. Observe if heights mismatch or headers disappear
4. Check if input fields are causing extra height/wrapping issues

---

### 2. Preview Banner Not Hidden in Text Selection Mode (M key)

**Problem:**
When previewing a file in full-screen mode and pressing **M** to enable text selection (which removes the border), the bright blue banner at the top that displays "Preview: FILENAME [Filetype]" does NOT get hidden. This means when selecting/copying multiple pages of text, the banner appears in the middle of the copied content.

**Current Behavior:**
- Press **Enter** or **F3** to preview a file (full-screen)
- Press **M** to toggle mouse/border for text selection
- Border disappears (correct) ✅
- Banner stays visible (incorrect) ❌

**Expected Behavior:**
- Pressing **M** should hide BOTH:
  1. The decorative border ✅ (already works)
  2. The blue "Preview: FILENAME [Filetype]" banner ❌ (needs fix)
- This allows clean text selection without the banner interrupting

**Where to Look:**
- `render_preview.go` - Full-screen preview rendering
- `renderFullPreview()` function
- The banner is rendered separately from the border
- Need to check the same toggle that hides border (`m.preview.mouseEnabled`?) and also hide the banner

**Test Case:**
1. Open a multi-page text file (e.g., CLAUDE.md)
2. Press **Enter** to view in full-screen
3. Press **M** to enable text selection mode
4. Verify banner at top is hidden
5. Try selecting text from top of file - banner should not be included

**Implementation Hint:**
The banner rendering likely looks something like:
```go
if m.viewMode == viewFullPreview {
    // Render banner with filename and filetype
    bannerStyle := lipgloss.NewStyle().Background(lipgloss.Color("39"))...
    s.WriteString(bannerStyle.Render("Preview: " + filename + " [" + filetype + "]"))
}
```

Need to add a condition:
```go
if m.viewMode == viewFullPreview && m.preview.mouseEnabled {
    // Only show banner when mouse/border is enabled
    // Hide it in text selection mode (M pressed)
}
```

---

**Files to Check:**

**For Issue #1 (Prompty Height Mismatch):**
- `render_preview.go` - Preview pane rendering
- `prompt_parser.go` - Prompty template parsing
- `model.go` - Layout calculations (`calculateLayout()`)
- `view.go` - Height calculations

**For Issue #2 (Banner Visibility):**
- `render_preview.go` - `renderFullPreview()` function
- Look for banner rendering code (bright blue background)
- Check where `m.preview.mouseEnabled` is used

**Success Criteria:**

✅ **Issue #1 Fixed:**
- Prompty files render correctly in narrow windows
- File tree and preview pane heights always match
- App headers never get hidden
- Input fields wrap properly without breaking layout

✅ **Issue #2 Fixed:**
- Pressing **M** in full-screen preview hides both border AND banner
- Clean text selection without banner interrupting
- Banner returns when **M** is pressed again (toggle)

**Priority:** High - Both are UX issues that affect core functionality
**Expected Time:** 30-45 minutes

---

## ✅ COMPLETED (Session 2025-10-21): Persistent Command History

### Feature Summary
Implemented persistent command history that saves to disk and survives TFE restarts, making it easy to recall and reuse complex commands across sessions.

### What Was Implemented
✅ **Persistent Storage** - History saved to `~/.config/tfe/command_history.json`
✅ **Load on Startup** - History automatically loaded from disk when TFE starts
✅ **Auto-Save** - History saved when adding commands and when quitting TFE
✅ **Keyboard Navigation** - Up/Down arrows navigate history (already worked, improved)
✅ **Mouse Wheel Support** - Scroll wheel navigates command history when prompt is focused
✅ **Navigation Lock** - Arrow keys, left/right, pageup/pagedown don't navigate file tree when command prompt is focused
✅ **Visual Feedback** - Red `!` prefix for run-and-exit commands (both single-pane and dual-pane modes)

### Implementation Details

**Files Modified:**
1. **command.go** - Added `loadCommandHistory()` and `saveCommandHistory()` functions with JSON persistence
2. **model.go** - Load history from disk in `initialModel()`
3. **update_keyboard.go** - Save history on quit (F10, Ctrl+C, exit/quit commands), added commandFocused checks for navigation keys
4. **update_mouse.go** - Added mouse wheel scrolling for command history navigation
5. **view.go** - Added red color-coding for `!` prefix in single-pane mode
6. **render_preview.go** - Added red color-coding for `!` prefix in dual-pane mode

**Storage Location:**
- `~/.config/tfe/command_history.json`

**Format:**
```json
{
  "commands": ["ls -la", "!claude --yolo", "htop"],
  "maxSize": 100
}
```

### Bug Fixes Applied

**Issue 1: File tree selection visible when command focused**
- Problem: File tree showed selection highlight even when command prompt was focused
- Fix: Added `!m.commandFocused` check to all selection rendering in `render_file_list.go`
- Result: File tree loses highlight completely when `:` is pressed

**Issue 2: Arrow keys still navigated file tree when command focused**
- Problem: Up/Down/k/j keys moved cursor in file tree even with command prompt focused
- Fix: Changed condition from checking `commandFocused && len(history) > 0` to just `commandFocused`
- Result: All navigation keys blocked when command mode active, even with empty history

**Issue 3: Mouse wheel still scrolled file tree when command focused**
- Problem: Mouse wheel scrolling navigated file tree instead of command history
- Fix: Updated mouse wheel handler to block file navigation when `commandFocused`, regardless of history length
- Result: Mouse wheel only works for history navigation (or does nothing if no history)

### Command Line Editing Features Added

**New Navigation:**
- **Left/Right arrows** - Move cursor within command text
- **Home/Ctrl+A** - Jump to beginning of line
- **End/Ctrl+E** - Jump to end of line
- **Ctrl+Left/Alt+Left/Alt+B** - Jump one word left
- **Ctrl+Right/Alt+Right/Alt+F** - Jump one word right

**New Editing:**
- **Delete** - Forward delete at cursor position
- **Ctrl+K** - Delete from cursor to end of line
- **Ctrl+U** - Delete from cursor to beginning of line
- **Backspace** - Now deletes at cursor (not just end)
- **Text insertion** - Inserts at cursor position (not just end)

**Visual:**
- Cursor `█` now renders at actual cursor position in text
- History navigation (↑/↓/mouse wheel) moves cursor to end

**Commit:** (pending)
**Branch:** headerdropdowns
**Status:** ✅ COMPLETE - All features tested and confirmed working

---

## 📋 Original Implementation Plan (Completed)

#### **1. Persistent Storage**
Save command history to: `~/.config/tfe/command_history.json`

**File format:**
```json
{
  "commands": [
    "ls -la",
    "!claude --dangerously-skip-permissions",
    "htop",
    "!vim myfile.txt"
  ],
  "maxSize": 100
}
```

#### **2. Load/Save Functions**

**In `command.go`:**
```go
// loadCommandHistory reads command history from disk
func loadCommandHistory() []string {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return []string{}
    }

    historyPath := filepath.Join(homeDir, ".config", "tfe", "command_history.json")
    data, err := os.ReadFile(historyPath)
    if err != nil {
        return []string{} // File doesn't exist yet, start fresh
    }

    var history struct {
        Commands []string `json:"commands"`
    }

    if err := json.Unmarshal(data, &history); err != nil {
        return []string{}
    }

    return history.Commands
}

// saveCommandHistory writes command history to disk
func (m *model) saveCommandHistory() error {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return err
    }

    configDir := filepath.Join(homeDir, ".config", "tfe")
    os.MkdirAll(configDir, 0755) // Create directory if it doesn't exist

    historyPath := filepath.Join(configDir, "command_history.json")

    history := struct {
        Commands []string `json:"commands"`
        MaxSize  int      `json:"maxSize"`
    }{
        Commands: m.commandHistory,
        MaxSize:  100,
    }

    data, err := json.MarshalIndent(history, "", "  ")
    if err != nil {
        return err
    }

    return os.WriteFile(historyPath, data, 0644)
}
```

#### **3. Integration Points**

**In `model.go` - initialModel():**
```go
m := model{
    // ... existing fields ...
    commandHistory: loadCommandHistory(), // Load from disk on startup
    historyPos:     0,
}
```

**In `update.go` - when quitting TFE:**
```go
case tea.KeyMsg:
    if msg.String() == "f10" || msg.String() == "ctrl+c" {
        m.saveCommandHistory() // Save before quitting
        return m, tea.Quit
    }
```

**In `command.go` - addToHistory():**
```go
func (m *model) addToHistory(command string) {
    // ... existing logic ...

    // Save to disk after adding
    m.saveCommandHistory()
}
```

#### **4. Visual Enhancement: Color-Code ! Prefix**

When displaying command history with ↑/↓ arrows, highlight the `!` prefix in red.

**In `view.go` and `render_preview.go` - command prompt rendering:**
```go
// Show the command with colored ! prefix
if m.commandFocused && m.commandInput != "" {
    // Check if command starts with !
    if strings.HasPrefix(m.commandInput, "!") {
        // Red ! prefix
        prefixStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
        s.WriteString(prefixStyle.Render("!"))
        // Normal text for rest of command
        s.WriteString(inputStyle.Render(m.commandInput[1:]))
    } else {
        s.WriteString(inputStyle.Render(m.commandInput))
    }
} else if m.commandFocused && m.commandInput == "" {
    // ... existing helper text ...
}
```

#### **5. Fix: Disable Arrow Key Navigation When Command Prompt is Focused**

**Problem:** Currently, when command prompt is focused (`:` pressed), arrow keys still navigate the file tree/preview pane. This conflicts with command history navigation (↑/↓).

**Solution:** In `update_keyboard.go`, check `m.commandFocused` before handling arrow key navigation.

**Current behavior:**
```go
case "up", "k":
    // Navigate file list up
    if m.cursor > 0 {
        m.cursor--
    }
```

**Fixed behavior:**
```go
case "up", "k":
    // If command prompt is focused, arrow keys are for history (already handled above)
    if m.commandFocused {
        return m, nil // Don't navigate file list
    }

    // Navigate file list up
    if m.cursor > 0 {
        m.cursor--
    }
```

**Apply same fix for:**
- `"down"`, `"j"` - Down navigation
- `"left"`, `"h"` - Left navigation (if used in file browser)
- `"right"`, `"l"` - Right navigation (if used in file browser)
- `"pageup"`, `"pagedown"` - Page navigation
- `"home"`, `"end"` - Jump to start/end

**Note:** Make sure this check is added EARLY in the switch statement, before arrow key handling for file navigation.

### Files to Modify

1. **command.go** - Add `loadCommandHistory()` and `saveCommandHistory()` functions
2. **model.go** - Load history from disk in `initialModel()`
3. **update.go** - Save history when quitting TFE
4. **update_keyboard.go** - Add `if m.commandFocused { return m, nil }` check before arrow key file navigation
5. **view.go** - Color-code `!` prefix in red when rendering command input
6. **render_preview.go** - Same `!` prefix coloring for dual-pane mode

### Testing Checklist

After implementation:
- [ ] Run TFE, execute a few commands (`:ls`, `:htop`, `:!vim test.txt`)
- [ ] Press `:` and use ↑/↓ to verify history works
- [ ] Quit TFE (F10)
- [ ] Check `~/.config/tfe/command_history.json` exists and contains commands
- [ ] Restart TFE
- [ ] Press `:` and ↑ - verify previous commands are loaded
- [ ] Verify `!` prefix appears in **red** when scrolling through history
- [ ] Press `:` to focus command prompt, then press arrow keys
- [ ] Verify arrow keys navigate history, NOT the file tree (file tree should not move)
- [ ] Press Esc to unfocus command prompt
- [ ] Verify arrow keys now navigate file tree normally

### Priority
**High** - Persistent history is a fundamental usability improvement that makes the command prompt much more practical for complex commands.

### Expected Time
**30-45 minutes** - Straightforward file I/O + small rendering changes + arrow key fix

---

## ✅ COMPLETED (Session 2025-10-21): Keyboard Navigation & Command Improvements

### Keyboard Navigation for Menus ✅
- Alt/F9 enters menu bar navigation mode
- Left/Right or Tab/Shift+Tab to navigate between menus
- Down/Enter to open dropdown, Up/Down to navigate items
- Left/Right in dropdown smoothly switches to adjacent menus
- Esc to close dropdown or exit menu mode
- Visual feedback with gray highlight for focused menu
- View menu reordered to match 1-2-3 hotkeys

### Command Execution Enhancement ✅
- `!` prefix support: `:!command` exits TFE after running command
- `:command` suspends TFE and returns after execution
- Smart helper text shows "! prefix to run & exit" when focused
- Works in both single-pane and dual-pane modes
- Perfect for launching Claude Code: `:!claude --dangerously-skip-permissions`

**Commit:** `28eced1` - feat: Add keyboard navigation for menu bar and ! prefix for command execution
**Branch:** headerdropdowns
**Status:** ✅ COMPLETE

---

## ✅ FIXED: Dropdown and Context Menu Alignment Issues

All dropdown and context menu alignment issues have been fixed:

✅ **Dropdown menus** - Simplified overlay with empty space padding (no ANSI bleeding)
✅ **Context menus** - Proper emoji width handling with go-runewidth
✅ **Menu width calculations** - Using lipgloss.Width() for accurate emoji/unicode width
✅ **Checkmark width** - Using actual visual width instead of hardcoded +2
✅ **Favorited files** - Context menu aligns correctly on files with ⭐ emoji
✅ **Empty space areas** - Context menu alignment consistent below file tree
✅ **Dynamic positioning** - Context menus reposition upward when near terminal bottom

**Commit:** `0674c8d` - fix: Resolve dropdown and context menu alignment issues
**Branch:** headerdropdowns
**Status:** ✅ COMPLETE

---

## ✅ FIXED: Menu Performance Lag (Dropdown + Context Menus)

### Root Cause Identified
Both the **dropdown menus** and **context menus** (right-click/F2) had **repeated filesystem lookups**:

**Dropdown menus** (`menu.go` - `getMenus()`):
- `editorAvailable("lazygit")` → `exec.LookPath("lazygit")`
- `editorAvailable("lazydocker")` → `exec.LookPath("lazydocker")`
- `editorAvailable("lnav")` → `exec.LookPath("lnav")`
- `editorAvailable("htop")` → `exec.LookPath("htop")`
- `editorAvailable("bottom")` → `exec.LookPath("bottom")`
- **Impact**: 5 filesystem lookups × 10+ renders/second = **50+ lookups/second** ⚠️

**Context menus** (`context_menu.go` - `getContextMenuItems()`):
- Same 5 tool checks PLUS `editorAvailable("micro")` check
- **Impact**: 6 filesystem lookups every time you navigate context menu with arrows or mouse ⚠️

### Solution Implemented
**Cache tool availability at startup** instead of checking on every render.

**Performance Improvement:**
- **Before**: 5-6 filesystem lookups per render = **50-60+ lookups/second** during navigation ⚠️
- **After**: 6 filesystem lookups total (at startup only) = **instant menus** ✅

**Status**: ✅ FIXED - Ready for testing

---

## 🐛 Bug to Fix: GIF Preview Mode

**Problem:** When previewing a GIF file, it shows "file too big to display" with text "press V to open in image viewer", but terminal image viewers can't show animated GIFs.

**Solution:** Add browser open option in preview mode for GIF files
- Add "B" key binding in preview mode to open in browser (like context menu does)
- Update help text to show: "V: view image • B: open in browser" for GIF files
- Reuses existing `openInBrowser()` function (already fixed with PowerShell!)

**Files to modify:**
- `update_keyboard.go` - Add "b" case in preview mode (around line 164-294)
- `render_preview.go` - Update help text for GIF files (around line 751-755)

**Implementation:**
```go
// In update_keyboard.go, preview mode section:
case "b", "B":
    // Open GIF in browser (for animated playback)
    if m.preview.loaded && m.preview.filePath != "" && isImageFile(m.preview.filePath) {
        return m, openInBrowser(m.preview.filePath)
    }
```

**Priority:** Medium (nice to have, works around limitation)

---

**Branch:** headerdropdowns
**Status:** Ready for persistent command history implementation
