# Next Session Tasks

## 🚨 PRIORITY: Fix Context Menu Alignment Issues

### Problem Summary
The context menu (right-click/F2) has alignment issues that cause the box borders to be misaligned and sometimes overflow off-screen:

1. **Emoji width inconsistencies**: Some file types (gzip files) have emojis that add extra spacing
2. **Favorited file stars**: The ⭐ star icon on favorited files adds extra spacing that throws off alignment
3. **Bottom border collision**: When the context menu is opened near the bottom of the terminal, the menu box art overlaps with the file tree's bottom border
4. **Off-screen rendering**: Context menus opened toward the bottom can extend beyond the terminal height
5. **Tree view expanded folders**: Alignment breaks when 2+ folders are expanded in tree view (tree characters ├─, └─, │ and indentation may affect positioning)

### Specific Issues

**Issue 1: Variable emoji widths**
- Some emojis render wider than others (e.g., gzip emoji vs folder emoji)
- This causes menu item text to misalign
- The box borders don't line up properly when items have different icon widths

**Issue 2: Favorited file stars**
- Files marked as favorites show a ⭐ icon
- This adds visual width but the menu width calculation doesn't account for it
- Results in ragged right edge or text overflow

**Issue 3: Border overlap**
- The file tree is wrapped in a bordered box (lipgloss.RoundedBorder)
- When context menu opens near bottom, it overlaps the file tree's bottom border
- Creates visual artifacts: `└─────┘` (file tree border) colliding with `┌──────┐` (menu border)

**Issue 4: Off-screen overflow**
- `m.contextMenuY` can be set too close to `m.height`
- Menu extends beyond visible terminal area
- User can't see all menu items

**Issue 5: Tree view with expanded folders**
- When multiple folders are expanded in tree view (press 3, then → to expand)
- Tree characters (├─, └─, │) add visual complexity and indentation
- Context menu alignment appears to break with 2+ expanded folders
- The tree characters may have different widths or the indentation calculation may be off

### Files to Review

**Primary file:**
- `context_menu.go` - Lines 399-471 (`renderContextMenu` function)
  - Calculates menu width based on item labels
  - Renders the bordered box with menu items
  - Does NOT currently account for emoji/star width variations

**Secondary files:**
- `update_mouse.go` - Lines 756-784 (context menu positioning on right-click)
  - Sets `m.contextMenuX` and `m.contextMenuY`
  - Has bounds checking but might not account for menu height properly
- `view.go` - Lines 395-478 (`overlayContextMenu` function)
  - Overlays the context menu on the base view
  - Ensures menu stays on screen (lines 400-412)
  - May need better bottom-edge detection

### Current Implementation (Potentially Problematic)

```go
// context_menu.go:410-417
// Calculate menu dimensions - find the longest item
maxWidth := 0
for _, item := range items {
    // Count runes, not bytes (better emoji support)
    width := len([]rune(item.label))  // ❌ Doesn't account for emoji visual width!
    if width > maxWidth {
        maxWidth = width
    }
}
```

**Problem**: `len([]rune(item.label))` counts Unicode code points, but emojis like 🗑️ can be 2 runes wide while displaying as 2 visual columns. Meanwhile, favorited files have `⭐ filename` which adds width not captured in the original label.

### Suggested Fixes

**1. Use `lipgloss.Width()` for accurate width calculation**
```go
// Instead of:
width := len([]rune(item.label))

// Use:
width := lipgloss.Width(item.label)
```
This handles emoji rendering width properly.

**2. Account for favorite stars in width calculation**
```go
// In renderContextMenu, check if file is favorited:
actualLabel := item.label
if m.contextMenuFile != nil && m.isFavorite(m.contextMenuFile.path) {
    actualLabel = "⭐ " + actualLabel  // Account for star width
}
width := lipgloss.Width(actualLabel)
```

**3. Improve bottom-edge collision detection**
```go
// In overlayContextMenu or when setting contextMenuY:
menuHeight := len(menuItems) + 2  // +2 for borders
maxY := m.height - menuHeight - 3  // -3 to avoid file tree bottom border
if m.contextMenuY > maxY {
    m.contextMenuY = maxY
}
```

**4. Consider moving menu up if too close to bottom**
```go
// After calculating menu height:
if m.contextMenuY + menuHeight >= m.height - 2 {
    // Move menu above the cursor position instead
    m.contextMenuY = m.contextMenuY - menuHeight
    if m.contextMenuY < 1 {
        m.contextMenuY = 1  // Don't go off top
    }
}
```

### Testing Checklist

After implementing fixes:
- [ ] Open context menu on a regular file (no star) → borders aligned properly
- [ ] Open context menu on a favorited file (⭐) → borders aligned properly
- [ ] Open context menu on gzip file (has wider emoji) → borders aligned properly
- [ ] Open context menu near bottom of terminal → doesn't overlap file tree border
- [ ] Open context menu at very bottom → menu repositions upward or fits on screen
- [ ] Scroll file tree and test context menu at various positions → always renders correctly
- [ ] **Tree View with expanded folders**: Switch to tree view (press 3), expand 2-3 folders (press →), then test context menu
  - Tree characters (├─, └─, │) add visual complexity
  - Multiple indentation levels may affect positioning
  - Check if borders align properly when menu opens on deeply nested items

### Reference Files for Emoji/Icon Handling

Check these for examples of proper width handling:
- `render_file_list.go` - Detail view already handles emoji alignment (might have good patterns)
- `file_operations.go` - `getFileIcon()` function returns different emojis

### Expected Outcome

**Before:**
```
┌─────────────────────────┐
│ 📁 src/                 │
│ 📄 main.go              │
│ 🗜️  archive.gz         │  ← gzip emoji wider, text misaligned
│ ⭐ config.json          │  ← star adds width, border doesn't match
└──────────────────────┘
  ┌────────────────┐         ← context menu border misaligned
  │ 📂 Open      │         ← items don't line up
  │ 📋 Copy Path  │
  └──────────────┘
```

**After:**
```
┌─────────────────────────┐
│ 📁 src/                 │
│ 📄 main.go              │
│ 🗜️  archive.gz          │  ← properly aligned
│ ⭐ config.json          │  ← properly aligned
│   ┌──────────────────┐  │
│   │ 📂 Open         │  │  ← clean borders
│   │ 📋 Copy Path    │  │  ← properly aligned
│   └──────────────────┘  │
└─────────────────────────┘
```

### Prompt to Use

Copy-paste this into your next session:

---

**PROMPT START:**

I need help fixing alignment issues with the context menu (right-click/F2) in my TFE file explorer. The context menu has several rendering problems:

1. **Emoji width inconsistencies**: Different file type emojis (like gzip 🗜️) have varying visual widths, causing menu items to misalign
2. **Favorited files**: Files with ⭐ stars have extra width not accounted for in the menu width calculation
3. **Bottom border collision**: When opened near the bottom, the context menu overlaps the file tree's bottom border
4. **Off-screen overflow**: Context menus can extend beyond terminal height when opened at the bottom

The main issue is in `context_menu.go` lines 410-417 where `len([]rune(item.label))` is used instead of `lipgloss.Width()` for width calculation. This doesn't properly account for emoji visual width.

Please:
1. Read `context_menu.go` (especially `renderContextMenu` function)
2. Fix the width calculation to use `lipgloss.Width()` instead of `len([]rune())`
3. Account for favorite stars (⭐) when calculating menu width
4. Add better bottom-edge detection to prevent border collisions
5. Implement menu repositioning (move up) when too close to terminal bottom

Test cases should include: regular files, favorited files, gzip files, opening menus at various vertical positions (top, middle, bottom of terminal), and **especially in tree view with 2-3 expanded folders** (the tree characters ├─, └─, │ and indentation may affect alignment).

**PROMPT END:**

---

**Branch**: headerdropdowns
**Priority**: 🔥 Medium (UX polish)
**Expected Time**: 30-45 minutes

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

**Files modified:**
1. **types.go** (lines 288-290): Added caching fields to model
   ```go
   // Menu caching (performance optimization)
   cachedMenus    map[string]Menu  // Cached menu structure (built once)
   toolsAvailable map[string]bool // Cached tool availability (lazygit, htop, etc.)
   ```

2. **model.go** (lines 48-56): Check tool availability once at initialization
   ```go
   // Menu caching - check tool availability once at startup (performance optimization)
   toolsAvailable: map[string]bool{
       "lazygit":     editorAvailable("lazygit"),
       "lazydocker":  editorAvailable("lazydocker"),
       "lnav":        editorAvailable("lnav"),
       "htop":        editorAvailable("htop"),
       "bottom":      editorAvailable("bottom"),
       "micro":       editorAvailable("micro"), // Used in context menu edit action
   },
   ```

3. **menu.go** (lines 85-118): Use cached availability in dropdown menus
   ```go
   // Use cached tool availability instead of filesystem lookups (performance optimization)
   if m.toolsAvailable["lazygit"] {
       // ... add lazygit menu item ...
   }
   ```

4. **context_menu.go** (lines 64-98, 227): Use cached availability in context menus
   ```go
   // Use cached tool availability (performance optimization)
   if m.toolsAvailable["lazygit"] {
       items = append(items, contextMenuItem{"🌿 Git (lazygit)", "lazygit"})
   }
   // ... same for other tools ...

   // In edit action:
   if m.toolsAvailable["micro"] {
       editor = "micro"
   }
   ```

### Performance Improvement
- **Before**: 5-6 filesystem lookups per render = **50-60+ lookups/second** during navigation ⚠️
- **After**: 6 filesystem lookups total (at startup only) = **instant menus** ✅

### Testing Results
Build successful! Binary created: `tfe` (16MB)

**Expected behavior:**
- ✅ No lag when opening dropdown menus (File, Edit, View, Tools, Help)
- ✅ No lag when opening context menus (right-click or F2 on files/folders)
- ✅ Arrow key navigation is instant in both menu types
- ✅ Mouse interaction is smooth
- ✅ Works in both single-pane and dual-pane modes

**Branch**: headerdropdowns
**Status**: ✅ FIXED - Ready for testing

---

## ✅ Completed
- Context-aware F1 help system (implemented and working!)
- VHS demo system created (10 tape files - good for docs but no emojis)
- Fixed browser opening for images/GIFs (PowerShell instead of cmd.exe)
- Tried VHS: emojis show as boxes ❌
- Tried asciinema: emojis show as boxes ❌

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

**Good luck! Sleep well! 😴**

When you're ready tomorrow with ☕:
1. Open OBS Studio
2. Window Capture → Windows Terminal
3. `cd ~/projects/TFE/demo-content`
4. Start OBS Recording
5. Launch `tfe` and show off features (follow DEMO_CHEATSHEET.txt)
6. Stop OBS Recording
7. Optional: Convert MP4 to GIF with ffmpeg

You got this! OBS will capture everything perfectly! 🚀🎥
