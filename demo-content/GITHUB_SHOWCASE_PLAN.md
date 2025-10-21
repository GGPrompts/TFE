# GitHub Showcase - Recording & Screenshot Plan

**Goal:** Create hero visuals for the GitHub repository main page

---

## 🎬 Priority Recordings (Pick 2-3)

### 1. **Hero GIF - Complete Workflow** (30-45 seconds)
**Purpose:** First thing visitors see - show TFE's power in under 1 minute

**Script:**
1. Start in TFE project root
2. Navigate files with j/k (show icons/emojis)
3. Enter dual-pane mode (Space)
4. Navigate to CLAUDE.md (orange AI file!)
5. Show live markdown preview updating
6. Open context menu (F2 or right-click)
7. Switch to tree view (3)
8. Expand a folder (→)
9. Open prompts mode (F11)
10. Show TFE-Customization folder
11. Return to browse mode (Esc)

**Highlights:** Icons, dual-pane, markdown, context menu, tree view, prompts

**Save as:** `tfe-hero.gif` (for top of README)

---

### 2. **Prompts Library Workflow** (20-30 seconds) ⭐ UNIQUE FEATURE
**Purpose:** Show off TFE's killer feature that no other file manager has

**Script:**
1. Press F11 (prompts mode)
2. Navigate to TFE-Customization folder
3. Open `add-tui-tool.prompty`
4. Show fillable fields at bottom (if any)
5. Press F5 to copy
6. Show "Copied to clipboard" status message
7. Esc to return

**Highlights:** Prompts library, customization docs, clipboard integration

**Save as:** `prompts-workflow.gif`

---

### 3. **Mobile/Touch Demo** (15-20 seconds) ⭐ UNIQUE FEATURE
**Purpose:** Show Termux/mobile support (very rare for file managers!)

**Script (if you have Termux):**
1. Open TFE on Android/Termux
2. Tap to select files (show touch works)
3. Long-press for context menu
4. Swipe to scroll
5. Open dual-pane mode
6. Navigate with touch

**Highlights:** Touch controls, mobile support

**Save as:** `mobile-demo.gif`

**Alternative:** If no Termux handy, skip this and use screenshot instead

---

### 4. **Context-Aware Help** (15-20 seconds)
**Purpose:** Show intelligent F1 help system

**Script:**
1. In single-pane mode, press F1 (jumps to Navigation section)
2. Esc to close
3. Open dual-pane mode (Space)
4. Press F1 (jumps to Dual-Pane section)
5. Esc to close
6. Enter prompts mode (F11)
7. Press F1 (jumps to Prompts section)
8. Show how it auto-scrolls to relevant help

**Highlights:** Context-aware navigation, intelligent UX

**Save as:** `context-help.gif`

---

## 📸 Priority Screenshots (Static Images)

### Screenshot 1: **Hero Screenshot** - Dual-Pane with Markdown
**Purpose:** Beautiful static image for GitHub social preview

**Setup:**
- Dual-pane mode
- Left: File list showing icons, emojis, orange AI files
- Right: CLAUDE.md with beautiful Glamour markdown rendering
- Full window, nice terminal size (~1200x700)

**How to capture:**
1. Open TFE in dual-pane
2. Navigate to CLAUDE.md (shows as orange)
3. Make sure preview shows nicely formatted markdown
4. Windows: Win+Shift+S (Snipping Tool)
5. Linux: Flameshot or Spectacle
6. Mac: Cmd+Shift+4

**Save as:** `hero-screenshot.png` (for GitHub social card)

---

### Screenshot 2: **Context Menu in Action**
**Purpose:** Show right-click functionality

**Setup:**
- Right-click context menu open on a directory
- Shows all options: Open, Quick CD, New Folder, New File, Copy, Rename, Delete, Favorites
- Maybe with TUI tools visible (if lazygit installed)

**Save as:** `context-menu.png`

---

### Screenshot 3: **Tree View Expanded**
**Purpose:** Show hierarchical navigation

**Setup:**
- Tree view mode (press 3)
- 2-3 folders expanded with →
- Shows nested structure clearly
- Emoji icons visible

**Save as:** `tree-view.png`

---

### Screenshot 4: **Prompts Mode with Fillable Fields** ⭐ UNIQUE
**Purpose:** Show the prompts library UI

**Setup:**
- F11 prompts mode
- Navigate to a .prompty file with fillable fields
- Show the input fields at bottom of screen
- Shows variable highlighting in preview

**Save as:** `prompts-mode.png`

---

### Screenshot 5: **Mobile Screenshot** (if Termux available)
**Purpose:** Show mobile support

**Setup:**
- Take screenshot on Android/Termux
- Shows TFE running with touch-friendly UI
- Maybe with on-screen keyboard showing F-keys

**Save as:** `termux-mobile.png`

---

## 🎨 Conversion Commands

### OBS Recording → GIF (using ffmpeg)

**Option 1: High Quality (2-4 MB)**
```bash
# Install ffmpeg if needed
sudo apt install ffmpeg  # Linux
brew install ffmpeg      # macOS
choco install ffmpeg     # Windows

# Convert MP4 to GIF
ffmpeg -i recording.mp4 -vf "fps=15,scale=1280:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse" -loop 0 output.gif
```

**Option 2: Smaller File (1-2 MB)**
```bash
# Lower quality, smaller file
ffmpeg -i recording.mp4 -vf "fps=10,scale=800:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse" -loop 0 output.gif
```

**Option 3: Using gifski (BEST QUALITY)**
```bash
# Install gifski
cargo install gifski

# Extract frames
ffmpeg -i recording.mp4 frame%04d.png

# Create GIF
gifski -o output.gif --fps 15 frame*.png

# Cleanup
rm frame*.png
```

### Screenshot Optimization

```bash
# Optimize PNG (reduce file size)
optipng hero-screenshot.png

# Or use pngquant for better compression
pngquant --quality=80-95 hero-screenshot.png -o hero-screenshot-optimized.png
```

---

## 📦 File Organization

```
assets/
├── hero-screenshot.png          # Main GitHub social preview
├── tfe-hero.gif                 # Complete workflow demo
├── prompts-workflow.gif         # Prompts library feature
├── context-help.gif             # Context-aware F1
├── context-menu.png             # Right-click menu
├── tree-view.png                # Tree navigation
├── prompts-mode.png             # Prompts with fillable fields
└── termux-mobile.png            # Mobile support (optional)
```

---

## 🎯 Priority Order

**If you only have time for 3 items:**
1. ⭐ Hero Screenshot (dual-pane markdown) - MUST HAVE
2. ⭐ Prompts Workflow GIF - UNIQUE FEATURE
3. ⭐ Hero GIF (complete workflow) - FIRST IMPRESSION

**If you have time for 5 items:**
4. Context Menu Screenshot
5. Context-Aware Help GIF

---

## 📝 Quick Workflow

1. **Set up OBS** (use OBS_SETTINGS.txt)
2. **Practice each recording** without recording first!
3. **Record 2-3 takes** of each (pick best one)
4. **Convert MP4 → GIF** using ffmpeg or gifski
5. **Optimize file sizes** (aim for <3MB per GIF)
6. **Take screenshots** using system tools
7. **Save to assets/** directory
8. **Update README.md** with new visuals if needed

---

## 💡 Tips

- **Go SLOW** - Pause 1-2 seconds between actions
- **Terminal size** - ~1200x700 or 100x30 columns
- **Font** - Use Nerd Font with good emoji support
- **Theme** - Your current dark theme looks great!
- **Practice** - Do a dry run before recording
- **File sizes** - Aim for 1-3MB per GIF (balance quality vs size)
- **Screenshots** - PNG format, optimize with optipng/pngquant

---

**Ready to create amazing visuals!** 📸✨
