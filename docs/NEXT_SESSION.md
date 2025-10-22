# Next Session Tasks - Landing Page, Games Launcher & Auto-Scroll

**Created:** 2025-10-22 (Session end)
**Estimated Time:** 45-60 minutes (3 tasks)

---

## Task 1: Remove Landing Page Completely

**Goal:** Remove the 90s Windows-style landing page with star background that appears when launching TFE. Launch directly into the file browser.

### What to Remove

The landing page currently shows:
- Starfield animation background
- Menu with options: Browse Files, Prompts, Favorites, Trash, Settings, Exit
- Nostalgic 90s Windows aesthetic

### Files to Modify

1. **`model.go`** - Remove landing page initialization
   - Line ~69: Change `showLandingPage: true` to `showLandingPage: false`
   - Or remove the landing page initialization entirely

2. **`types.go`** - (Optional) Clean up landing page fields if removing completely
   - Lines ~296-297: `showLandingPage` and `landingPage` fields
   - Can leave these for now or remove if you want complete cleanup

3. **`update_keyboard.go`** - Remove landing page keyboard handling
   - Lines ~34-82: Entire landing page input handling section
   - This intercepts keys before file browser gets them

4. **`view.go`** - Remove landing page rendering
   - Find where landing page is rendered (check for `if m.showLandingPage`)
   - Remove the conditional and always render file browser

### Search Commands

```bash
# Find all landing page references
grep -rn "showLandingPage" .
grep -rn "landingPage" .
grep -rn "LandingPage" .

# Find the landing page type definition
grep -rn "type LandingPage" .
```

### Expected Behavior After Fix

- Launch TFE: `./tfe`
- **Expected:** Goes directly to file browser (current directory)
- **No more:** Landing page with menu options

---

## Task 2: Fix Games Launcher Path

**Goal:** Update games launcher to point to `~/projects/TUIClassics` instead of current (wrong) path.

### Current Issues

The games launcher is accessible from:
1. **Emoji menu bar** - 🎮 Games button (clickable)
2. **Tools dropdown menu** - Has a "Games" option

Both currently point to the wrong location.

### Files to Modify

1. **`menu.go`** - Menu bar emoji buttons and Tools dropdown menu

### What to Find

Search for games launcher references:

```bash
# Find games launcher code
grep -rn "games" menu.go
grep -rn "Games" menu.go
grep -rn "🎮" menu.go

# Find tools menu definition
grep -rn "tools" menu.go
grep -rn "Tools" menu.go
```

### Expected Locations

Look for:
- Menu bar button definitions (emoji buttons array)
- Tools menu items array
- Action handlers for games button/menu item

### What to Change

**Current (wrong):** Probably points to old path or incorrect launcher

**New (correct):**
```go
// Path to games launcher
gamesPath := filepath.Join(os.Getenv("HOME"), "projects", "TUIClassics", "launcher")
// Or: ~/projects/TUIClassics/launcher
```

### Expected Behavior After Fix

1. Click 🎮 Games button in menu bar
   - **Expected:** Launches `~/projects/TUIClassics/launcher`

2. Open Tools menu (T or F9), select "Games"
   - **Expected:** Launches `~/projects/TUIClassics/launcher`

---

## Task 3: Auto-Scroll to Focused Variable in Edit Mode

**Goal:** When navigating between variables with Tab/Shift+Tab in edit mode, automatically scroll the preview to show the focused variable if it's off-screen.

### Current Issue

In prompt edit mode:
1. User presses Tab to navigate to next variable
2. If variable is several screens down, it's off-screen
3. User has to manually scroll down (j/k keys) to find focused variable
4. No visual feedback showing where the focus went

### Expected Behavior

When pressing Tab or Shift+Tab:
1. Focus moves to next/previous variable ✅ (already works)
2. **NEW:** Preview automatically scrolls to show the focused variable
3. Focused variable should be centered in view (or near top)

### Files to Modify

1. **`update_keyboard.go`** - Tab/Shift+Tab handling in edit mode
   - Lines ~261-278: Tab navigation code
   - After changing `m.focusedVariableIndex`, calculate scroll position

2. **`render_preview.go`** - Preview rendering with line numbers
   - May need helper function to find line number of focused variable

### Implementation Approach

**Step 1:** Find the line number where the focused variable appears in the preview

```go
// In edit mode Tab handler (update_keyboard.go ~261)
case "tab":
    // Navigate to next variable
    if len(m.preview.promptTemplate.variables) > 0 {
        m.focusedVariableIndex++
        if m.focusedVariableIndex >= len(m.preview.promptTemplate.variables) {
            m.focusedVariableIndex = 0 // Wrap around
        }

        // NEW: Auto-scroll to focused variable
        m.scrollToFocusedVariable()
    }
    return m, nil
```

**Step 2:** Add helper function to calculate scroll position

```go
// In helpers.go or update_keyboard.go
func (m *model) scrollToFocusedVariable() {
    if m.focusedVariableIndex < 0 || m.preview.promptTemplate == nil {
        return
    }

    // Get focused variable name
    varName := m.preview.promptTemplate.variables[m.focusedVariableIndex]

    // Find line number where this variable appears
    // Search preview.content for "{{" + varName + "}}"
    targetLine := -1
    searchPattern := "{{" + varName + "}}"
    for i, line := range m.preview.content {
        if strings.Contains(line, searchPattern) {
            targetLine = i
            break
        }
    }

    if targetLine >= 0 {
        // Calculate scroll position to center the variable
        visibleLines := m.height - 6  // Adjust for header/footer
        centerOffset := visibleLines / 2

        newScrollPos := targetLine - centerOffset
        if newScrollPos < 0 {
            newScrollPos = 0
        }

        m.preview.scrollPos = newScrollPos
    }
}
```

### Testing

**Test 1:** Long prompt with off-screen variables
1. Open a prompt with 5+ variables spread across 100+ lines
2. Press Tab to enter edit mode
3. Press Tab repeatedly
4. **Expected:** Preview auto-scrolls to show each focused variable ✅
5. Variable appears near center or top of screen ✅

**Test 2:** Shift+Tab (previous variable)
1. In edit mode, Tab through several variables
2. Press Shift+Tab to go backwards
3. **Expected:** Preview scrolls up to show previous variables ✅

**Test 3:** Wrap-around behavior
1. Tab to last variable (bottom of file)
2. Press Tab again (wraps to first variable)
3. **Expected:** Preview scrolls back to top ✅

**Test 4:** Short prompts (no scrolling needed)
1. Open prompt with 2 variables, all visible on one screen
2. Press Tab to navigate
3. **Expected:** No scrolling (all variables already visible) ✅

### Edge Cases

- **Variable not found in content:** If variable doesn't appear in preview.content (shouldn't happen), don't scroll
- **Dual-pane vs fullscreen:** Auto-scroll should work in both modes
- **First Tab press (entering edit mode):** Should scroll to first variable
- **Multiple instances of same variable:** Scroll to first occurrence

---

## Testing Checklist

### Test 1: Landing Page Removal
- [ ] Build: `go build -o tfe`
- [ ] Run: `./tfe`
- [ ] **Expected:** File browser appears immediately (no landing page)
- [ ] Navigate normally with arrow keys
- [ ] All features work (prompts, favorites, etc.)

### Test 2: Games Launcher (Menu Bar)
- [ ] Launch TFE
- [ ] Click 🎮 Games button in menu bar
- [ ] **Expected:** TUIClassics launcher starts
- [ ] Games menu shows available games
- [ ] Can launch a game (test with one game)

### Test 3: Games Launcher (Tools Menu)
- [ ] Launch TFE
- [ ] Press T or F9 to open Tools menu
- [ ] Navigate to "Games" option
- [ ] Press Enter
- [ ] **Expected:** TUIClassics launcher starts

### Test 4: Regression (Other Menu Items)
- [ ] Other emoji buttons still work (📝, 🌐, 🔧, ❓)
- [ ] Other Tools menu items still work (lazygit, htop, etc.)

### Test 5: Auto-Scroll in Edit Mode
- [ ] Open long prompt (100+ lines with 5+ variables)
- [ ] Press Tab to enter edit mode
- [ ] Press Tab repeatedly
- [ ] **Expected:** Preview auto-scrolls to show focused variable
- [ ] Press Shift+Tab to go backwards
- [ ] **Expected:** Preview scrolls up to previous variables
- [ ] Tab through all variables (wrap around to first)
- [ ] **Expected:** Smooth scrolling throughout

---

## Implementation Tips

### Finding the Code

1. **Landing page:**
   ```bash
   # Start with model.go
   grep -A 5 "showLandingPage" model.go

   # Check view.go for rendering
   grep -A 10 "showLandingPage" view.go

   # Check keyboard handling
   grep -A 20 "Handle landing page" update_keyboard.go
   ```

2. **Games launcher:**
   ```bash
   # Check menu.go for button definitions
   grep -B 5 -A 5 "🎮" menu.go

   # Check for Tools menu items
   grep -B 5 -A 5 "Games" menu.go
   ```

### Code Pattern to Look For

**Landing page initialization (model.go):**
```go
m := model{
    // ...
    showLandingPage: true,  // ← Change to false or remove
    landingPage: nil,       // ← Can remove
}
```

**Games launcher (menu.go):**
```go
// Look for something like:
{
    label: "🎮 Games",
    action: func() {
        // Wrong path here - fix it!
        launchGames("/old/path")
    },
}
```

**Correct games path:**
```go
gamesLauncher := filepath.Join(os.Getenv("HOME"), "projects", "TUIClassics", "launcher")
// Or use os.UserHomeDir():
homeDir, _ := os.UserHomeDir()
gamesLauncher := filepath.Join(homeDir, "projects", "TUIClassics", "launcher")
```

---

## Rollback Plan (If Issues)

If anything breaks:

1. **Landing page issues:**
   ```bash
   # Restore showLandingPage: true in model.go
   git diff model.go  # See what changed
   ```

2. **Games launcher issues:**
   ```bash
   # Check what path was used before
   git diff menu.go
   # Verify TUIClassics launcher exists
   ls -la ~/projects/TUIClassics/launcher
   ```

---

## Files Summary

**Must modify:**
- `model.go` - Disable landing page
- `update_keyboard.go` - Remove landing page keyboard handling, add auto-scroll for Tab/Shift+Tab
- `view.go` - Remove landing page rendering
- `menu.go` - Fix games launcher path
- `helpers.go` - Add scrollToFocusedVariable() function

**Optional cleanup:**
- `types.go` - Remove landing page type fields
- `landing_page.go` (if exists) - Can delete entire file

---

## Architecture Notes (from CLAUDE.md)

**Landing page:**
- Purpose: "90s Windows nostalgic intro"
- Added as feature, not core to file browser
- Safe to remove without affecting core functionality

**Menu system:**
- `menu.go` - Menu bar rendering & button definitions
- Emoji buttons are clickable shortcuts
- Tools dropdown has TUI tool integrations

---

## Quick Start Prompt for Next Session

Copy and paste this into your next chat:

```
I need to make three changes to TFE:

1. REMOVE LANDING PAGE COMPLETELY:
   - The app currently shows a landing page with star background when launched
   - I want to go directly to the file browser instead
   - Need to modify model.go, update_keyboard.go, view.go, and possibly types.go
   - Look for "showLandingPage" and "landingPage" references

2. FIX GAMES LAUNCHER PATH:
   - The 🎮 Games button in menu bar and "Games" in Tools dropdown menu
   - Currently points to wrong path
   - Should point to: ~/projects/TUIClassics/launcher
   - Need to modify menu.go

3. AUTO-SCROLL TO FOCUSED VARIABLE IN EDIT MODE:
   - When I press Tab/Shift+Tab to navigate between variables in edit mode
   - If the focused variable is off-screen (several screens down)
   - The preview should automatically scroll to show the focused variable
   - Need to add scrollToFocusedVariable() helper function
   - Need to modify Tab/Shift+Tab handlers in update_keyboard.go

INVESTIGATION:
- Find all references to landing page and remove/disable
- Find games launcher code in menu.go and update path
- Find Tab/Shift+Tab handling in edit mode (update_keyboard.go ~261-278)
- Add auto-scroll logic after changing focusedVariableIndex

TESTING:
- Test that TFE launches directly to file browser (no landing page)
- Test that games launcher works from menu bar and Tools menu
- Test that Tab navigation auto-scrolls to show focused variables

See docs/NEXT_SESSION.md for full details, implementation approach, and testing checklist.
```

---

**Ready for next session! 🚀**
