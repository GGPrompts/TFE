# Next Session Tasks

## 🎯 PRIORITY: Add Keyboard Navigation for Header Dropdowns

### Feature Request
Add keyboard shortcuts to navigate and interact with the header dropdown menus (File, Edit, View, Tools, Help) without using the mouse.

### Why This is Important
✅ **Accessibility** - Full keyboard navigation for users who prefer/need keyboard-only interaction
✅ **Power User Efficiency** - Faster navigation than mouse (Alt → Arrow → Enter is quicker than mouse)
✅ **Professional UX** - Matches traditional desktop application behavior (Windows, Linux, macOS apps)
✅ **Consistency** - TFE already has excellent keyboard shortcuts (F1-F12, Vim bindings) - this completes the picture
✅ **Discoverability** - Users who press Alt or F9 will discover the menu system

### Proposed Implementation

#### **Hotkey to Enter Menu Mode:**
- **Alt** or **F9** - Enter "menu mode" and highlight the first menu (File)
  - Alt is traditional (Windows/Linux) but may not work in all terminals
  - F9 is the safe fallback (F10 is already "Exit", so F9 makes sense)
  - Either key should work

#### **Navigation Keys:**
When in menu mode (menu bar is focused):
- **Left/Right arrows** - Move between menus (File ↔ Edit ↔ View ↔ Tools ↔ Help)
- **Down arrow / Enter** - Open the currently highlighted menu dropdown
- **Escape** - Exit menu mode, return to file browser
- **Tab** - Alternative to Right arrow (some users prefer Tab for menu navigation)

When dropdown is open (already implemented ✓):
- **Up/Down arrows** - Navigate menu items (already works!)
- **Enter** - Execute selected menu item (already works!)
- **Escape** - Close dropdown and return to menu bar focus
- **Left/Right arrows** - Close current menu and open adjacent menu (smooth horizontal navigation)

#### **Visual Feedback:**
- Menu bar item should show highlight when in menu mode:
  ```
  Normal:    File  Edit  View  Tools  Help
  Focused:  [File] Edit  View  Tools  Help
  Active:   [File] Edit  View  Tools  Help (with dropdown open below)
  ```
- Use existing `selectedStyle` or similar styling for consistency

### Implementation Details

#### **New State Variables (types.go):**
```go
menuBarFocused   bool   // True when user is navigating the menu bar
highlightedMenu  string // Which menu is highlighted ("file", "edit", etc.)
```

#### **Keyboard Handler (update_keyboard.go):**

**Entry:**
```go
case "alt", tea.KeyF9:
    if !m.menuBarFocused {
        m.menuBarFocused = true
        m.highlightedMenu = "file" // Start with first menu
        m.menuOpen = false
    }
```

**Menu bar navigation:**
```go
if m.menuBarFocused && !m.menuOpen {
    switch msg.String() {
    case "left", "shift+tab":
        // Move to previous menu
        m.highlightedMenu = getPreviousMenu(m.highlightedMenu)

    case "right", "tab":
        // Move to next menu
        m.highlightedMenu = getNextMenu(m.highlightedMenu)

    case "down", "enter":
        // Open the highlighted menu
        m.menuOpen = true
        m.activeMenu = m.highlightedMenu
        m.selectedMenuItem = getFirstSelectableMenuItem(m.activeMenu)

    case "esc":
        // Exit menu mode
        m.menuBarFocused = false
        m.highlightedMenu = ""
    }
}
```

**Dropdown navigation (enhance existing):**
```go
if m.menuOpen && m.activeMenu != "" {
    switch msg.String() {
    case "left":
        // Close current menu, open previous menu
        m.activeMenu = getPreviousMenu(m.activeMenu)
        m.selectedMenuItem = getFirstSelectableMenuItem(m.activeMenu)

    case "right":
        // Close current menu, open next menu
        m.activeMenu = getNextMenu(m.activeMenu)
        m.selectedMenuItem = getFirstSelectableMenuItem(m.activeMenu)

    // Existing up/down/enter/esc logic stays the same
    }
}
```

#### **Helper Functions (menu.go):**
```go
// getPreviousMenu returns the menu key to the left of the current menu
func getPreviousMenu(current string) string {
    order := getMenuOrder()
    for i, key := range order {
        if key == current {
            if i == 0 {
                return order[len(order)-1] // Wrap to last menu
            }
            return order[i-1]
        }
    }
    return order[0]
}

// getNextMenu returns the menu key to the right of the current menu
func getNextMenu(current string) string {
    order := getMenuOrder()
    for i, key := range order {
        if key == current {
            if i == len(order)-1 {
                return order[0] // Wrap to first menu
            }
            return order[i+1]
        }
    }
    return order[0]
}
```

#### **Rendering (menu.go - renderMenuBar):**
```go
// When rendering each menu label, check if it's highlighted
for _, menuKey := range menuOrder {
    menu := menus[menuKey]

    var style lipgloss.Style
    if m.activeMenu == menuKey && m.menuOpen {
        style = menuActiveStyle // Already exists
    } else if m.highlightedMenu == menuKey && m.menuBarFocused {
        style = menuHighlightedStyle // New: show focus without opening
    } else {
        style = menuInactiveStyle
    }

    renderedMenu := style.Render(menu.Label)
    renderedMenus = append(renderedMenus, renderedMenu)
}
```

**New style for highlighted (but not open) menu:**
```go
menuHighlightedStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("0")).
    Background(lipgloss.Color("240")).  // Different from active (39)
    Bold(true).
    Padding(0, 1)
```

### Files to Modify

1. **types.go** - Add `menuBarFocused` and `highlightedMenu` fields to model
2. **model.go** - Initialize new fields to `false` and `""`
3. **update_keyboard.go** - Add Alt/F9 handler and menu navigation logic
4. **menu.go** - Add `getPreviousMenu()`, `getNextMenu()`, update `renderMenuBar()` with highlight style
5. **HOTKEYS.md** - Document the new keyboard shortcuts

### Testing Checklist

After implementation:
- [ ] Press Alt or F9 → First menu (File) is highlighted
- [ ] Left/Right arrows → Navigate between menus
- [ ] Down arrow or Enter → Open highlighted menu
- [ ] Up/Down arrows → Navigate within dropdown (already works)
- [ ] Left/Right arrows in dropdown → Switch to adjacent menu smoothly
- [ ] Enter in dropdown → Execute action, close menu
- [ ] Escape in dropdown → Return to menu bar focus (stay in menu mode)
- [ ] Escape in menu bar → Exit menu mode, return to file browser
- [ ] Tab key → Works as alternative to Right arrow
- [ ] Shift+Tab → Works as alternative to Left arrow
- [ ] Menu wrapping → Right on Help goes to File, Left on File goes to Help

### Benefits

✅ **Complete keyboard workflow** - Every menu item accessible without mouse
✅ **Discoverability** - Users explore menus with keyboard (press Alt to discover)
✅ **Power user efficiency** - `Alt → Right → Right → Enter` faster than mouse
✅ **Accessibility compliance** - Keyboard-only users can access all features
✅ **Professional polish** - Matches expectations from desktop apps

### Priority
**High** - This is a quality-of-life improvement that elevates TFE's UX significantly

### Expected Time
**45-60 minutes** - Straightforward implementation, mostly keyboard handling logic

---

## ✅ FIXED: Dropdown and Context Menu Alignment Issues

### Issues Resolved (Session 2025-10-21)
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

## ✅ Completed
- Context-aware F1 help system (implemented and working!)
- VHS demo system created (10 tape files - good for docs but no emojis)
- Fixed browser opening for images/GIFs (PowerShell instead of cmd.exe)
- Tried VHS: emojis show as boxes ❌
- Tried asciinema: emojis show as boxes ❌
- Dropdown menu alignment issues (emoji width, ANSI bleeding)
- Context menu alignment issues (favorited files, empty space)

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

## 🎬 Goal for This Session
Record with **OBS** to capture TFE's **actual terminal appearance** with:
- ✅ Proper emoji rendering (file icons, folders, AI context files)
- ✅ CGA theme colors
- ✅ FiraCode Nerd Font
- ✅ Bright visible cursor
- ✅ Mouse interaction visible
- ✅ Real project files

## Why OBS?
OBS captures your actual screen, so everything looks exactly like your terminal - emojis, colors, fonts, cursor - PERFECT! 🎥

---

## 🎥 Quick Start: Record with OBS

### Step 1: Prep Your Terminal
```bash
# Navigate to demo content
cd ~/projects/TFE/demo-content

# Open cheat sheet in second window (optional)
cat DEMO_CHEATSHEET.txt

# Size your Windows Terminal to reasonable size
# Not too small, not full screen
# Recommended: ~1200x700 or similar
```

### Step 2: Setup OBS
1. Open OBS Studio
2. **Add Source:** "Window Capture"
   - Select: Windows Terminal
   - Or: "Display Capture" for fullscreen
3. **Crop:** Right-click source → Transform → Edit Transform
   - Crop to just show terminal window (no taskbar/desktop)
4. **Settings → Output:**
   - Output Mode: Simple
   - Recording Quality: High Quality
   - Recording Format: MP4
5. **Settings → Video:**
   - Base Resolution: 1920x1080 (or your screen res)
   - Output Resolution: 1280x720 (smaller = smaller file)
   - FPS: 30

### Step 3: Record Demo
1. **Start Recording** in OBS (or hotkey)
2. **Launch TFE:**
   ```bash
   tfe
   ```
3. **Follow the demo script** (see cheat sheet or below)
4. **Stop Recording** when done
5. **File saved to:** Videos folder (default)

### Step 4: Convert MP4 to GIF (if needed)
```bash
# Install ffmpeg (if not already)
sudo apt install ffmpeg

# Convert (adjust path to your recording)
ffmpeg -i ~/path/to/recording.mp4 \
  -vf "fps=15,scale=1200:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse" \
  -loop 0 assets/tfe-showcase.gif

# Or simpler (lower quality but smaller):
ffmpeg -i recording.mp4 -vf "fps=15,scale=800:-1" assets/tfe-demo.gif
```

### Step 5: View Result
```bash
# MP4
explorer.exe ~/Videos  # or wherever OBS saved it

# GIF (if converted)
cd ~/projects/TFE/assets
explorer.exe .
```

---

## 📋 Demo Script Ideas

### Demo 1: Complete Feature Tour (45 seconds)
**Filename:** `tfe-complete-tour.cast`

**Script:**
1. Launch TFE in demo-content
2. Navigate down 3-4 files (shows icons)
3. Toggle to tree view (press 3)
4. Expand a folder (→)
5. Enter dual-pane mode (Space)
6. Navigate files (preview updates)
7. Switch to preview pane (Tab)
8. Scroll preview
9. Exit (Esc, F10)

### Demo 2: AI Context Files (20 seconds)
**Filename:** `tfe-ai-context.cast`

**Script:**
1. Launch TFE in project root (~/projects/TFE)
2. Navigate to CLAUDE.md (orange icon 🤖)
3. Preview it (Enter)
4. Show beautiful markdown rendering
5. Navigate to .claude/ folder
6. Show prompt files
7. Exit

### Demo 3: Dual-Pane Workflow (30 seconds)
**Filename:** `tfe-dual-pane.cast`

**Script:**
1. Launch TFE
2. Enter dual-pane immediately (Space)
3. Navigate 4-5 files (preview updates automatically)
4. Switch to preview pane (Tab)
5. Scroll preview
6. Switch back to file list (Tab)
7. Navigate more files
8. Exit dual-pane (Esc)

---

## 🎨 Tips for Great Recordings

### Before Recording:
- ✅ Clear your terminal (Ctrl+L)
- ✅ Make sure window is full size
- ✅ Test the workflow first (practice run)
- ✅ Know your ending point (plan last action)

### During Recording:
- 🐢 Go SLOWER than normal (viewers need time to see)
- ⏸️ Pause 1-2 seconds after each action
- 🎯 Focus on one feature at a time
- 📦 Show varied file types (icons!)
- 🤖 Highlight AI context files (.claude/, CLAUDE.md)

### After Recording:
- 🎬 Review the .cast file (play it with `asciinema play file.cast`)
- ✂️ If you mess up, just record again (quick!)
- 🔧 Convert to GIF and check file size

---

## 🛠️ Advanced: Edit Recordings

If you want to trim or speed up parts:

```bash
# Play recording to review
asciinema play tfe-showcase.cast

# Edit timing (if needed) - opens in editor
# You can manually adjust timestamps
nano tfe-showcase.cast

# Or use asciinema tools to cut/trim
# (see: https://docs.asciinema.org/manual/cli/editing/)
```

---

## 📦 Final Output Goals

### For README.md Hero Section:
- **1 showcase GIF** (45-60 seconds) showing complete feature tour
- Shows emoji icons, beautiful rendering, smooth navigation
- Demonstrates why TFE is awesome
- File: `assets/tfe-showcase.gif`
- Target size: < 3 MB

### For Features Section:
- **1-2 feature-specific GIFs** (20-30 seconds each)
- Dual-pane workflow
- AI context file management
- Files: `assets/demo-dual-pane-real.gif`, `assets/demo-ai-context.gif`
- Target size: < 2 MB each

---

## 🔧 Troubleshooting

### GIF is too large (> 5 MB)
```bash
# Option 1: Lower quality with agg
agg input.cast output.gif --cols 100 --rows 30

# Option 2: Optimize with gifsicle
sudo apt install gifsicle
gifsicle -O3 --lossy=80 --colors 128 -o output-opt.gif output.gif

# Option 3: Record shorter demo (< 45 seconds)
```

### OBS not capturing terminal properly
- Make sure "Window Capture" is selected (not Display Capture)
- Select the correct window: "Windows Terminal"
- Right-click source → Transform → Edit Transform to crop
- Check Settings → Video → Output Resolution (1280x720 is good)

### MP4 file is huge (> 50 MB)
- Lower output resolution in OBS (Settings → Video → 1280x720)
- Lower FPS (30 is fine, 24 is smaller)
- Convert to GIF (much smaller): `ffmpeg -i video.mp4 -vf "scale=800:-1" out.gif`

### Cursor not visible in recording
- Your bright cursor should show! If not:
- In OBS, make sure you're capturing the window (not display)
- Check Windows Terminal cursor settings

---

## ✅ Session Checklist

- [ ] Setup OBS (Window Capture → Windows Terminal)
- [ ] Record complete feature tour with OBS (~45 seconds)
- [ ] Check MP4 recording looks good (emojis visible!)
- [ ] Convert MP4 to GIF with ffmpeg (optional)
- [ ] Verify file size (< 5 MB for MP4, < 3 MB for GIF)
- [ ] Optional: Record dual-pane demo
- [ ] Optional: Record AI context files demo
- [ ] Update README.md with new demos

---

## 📚 Quick Reference

```bash
# OBS Recording
# 1. Open OBS
# 2. Window Capture → Windows Terminal
# 3. Start Recording
# 4. Do demo
# 5. Stop Recording

# Convert MP4 to GIF
ffmpeg -i recording.mp4 -vf "fps=15,scale=800:-1" output.gif

# Optimize GIF (if needed)
gifsicle -O3 --lossy=80 -o optimized.gif input.gif

# Open folders
explorer.exe ~/Videos        # OBS recordings
explorer.exe assets          # GIF output
```

---

## 🎯 Success Criteria

✅ At least 1 beautiful GIF showing TFE's real appearance
✅ Emojis render properly (not boxes!)
✅ Shows key features: navigation, dual-pane, tree view
✅ File size under 3 MB (loads fast on GitHub)
✅ Looks professional and engaging

---

**Branch:** headerdropdowns
**Status:** Ready for keyboard navigation implementation
