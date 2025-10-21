# Next Session Tasks

## 🚨 PRIORITY: Fix Dropdown Menu Performance Lag

### Problem
The dropdown menus are causing noticeable lag in TFE, likely due to overlay rendering issues. The current implementation may have ASCII bleed-through or inefficient overlay compositing causing performance degradation.

### Investigation Required

**Reference implementation**: `~/projects/TUITemplate/examples/tui-showcase/`

Compare the dropdown overlay logic between:
- **TFE**: `menu.go` and `view.go` (overlayDropdown function)
- **TUITemplate**: Check how dropdowns are rendered without causing lag

**Key files to review:**
```bash
# TFE (current - laggy)
/home/matt/projects/TFE/menu.go - renderActiveDropdown()
/home/matt/projects/TFE/view.go - overlayDropdown() (around line 481-517)

# TUITemplate (reference - should be performant)
~/projects/TUITemplate/examples/tui-showcase/menu.go
~/projects/TUITemplate/examples/tui-showcase/view.go
```

### Specific Issues to Check

1. **ASCII Bleed-Through:**
   - Are ANSI codes being properly stripped/handled?
   - Is the overlay clearing the background properly?
   - Are there remnants of previous frames showing through?

2. **Overlay Rendering:**
   - How does TUITemplate composite the dropdown over base content?
   - Should we use Lipgloss `Place()` instead of manual line replacement?
   - Are we reconstructing the entire view on every render?

3. **Performance:**
   - Is the overlay being rendered every frame even when not visible?
   - Are we caching the dropdown content or regenerating it constantly?

### Current TFE Implementation (Potentially Problematic)

```go
// view.go:481-517
func (m model) overlayDropdown(baseView, dropdown string, x, y int) string {
    // Simple approach: splits base view into lines, replaces lines with dropdown
    // This might be inefficient or causing ASCII issues
    baseLines := strings.Split(baseView, "\n")
    dropdownLines := strings.Split(dropdown, "\n")

    // ... creates new lines with padding + dropdown
    newLine := strings.Repeat(" ", x) + dropdownLine
}
```

### Potential Fixes

Based on TUITemplate comparison, likely need to:

1. **Use Lipgloss Place() instead of manual overlay**
2. **Strip ANSI codes from base view before overlaying**
3. **Only render dropdown when menu is open**
4. **Cache dropdown content**

### Testing

**Before fixing:**
```bash
cd /home/matt/projects/TFE
./tfe
# Wait 5+ seconds for menu to appear
# Click "File" menu - notice lag
# Navigate with arrow keys - check responsiveness
```

**After fixing:**
- [ ] No lag when opening dropdown menus
- [ ] Arrow key navigation is instant
- [ ] No visual artifacts or ASCII bleed-through
- [ ] Works in both single-pane and dual-pane modes

### Commands for Investigation

```bash
# Compare implementations
cd ~/projects/TUITemplate/examples/tui-showcase
grep -A 50 "overlay\|renderMenu" *.go

cd /home/matt/projects/TFE
grep -A 50 "overlayDropdown\|renderActiveDropdown" view.go menu.go

# Check for ANSI stripping utilities
grep -r "stripANSI\|cleanANSI" ~/projects/TUITemplate/
grep -r "stripANSI\|cleanANSI" /home/matt/projects/TFE/
```

**Branch**: headerdropdowns
**Priority**: 🔥 High (UX blocker)
**Expected Time**: 1-2 hours

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
