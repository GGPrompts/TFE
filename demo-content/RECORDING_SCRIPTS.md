# 🎬 OBS Recording Scripts - Quick Reference

Keep this open in a second monitor while recording!

---

## 📹 SCRIPT 1: Hero GIF (30-45 sec)

**What it shows:** Complete TFE workflow - the ultimate showcase

```
BEFORE RECORDING:
  cd ~/projects/TFE
  Clear terminal (Ctrl+L)
  Make sure terminal is ~1200x700

START RECORDING in OBS

1. ./tfe                           ⏱️ 2 sec
   [Screen: TFE opens showing file list with icons]

2. j j j j j                       ⏱️ 3 sec
   [Screen: Navigate down, show emojis (📁 📄 🐍)]

3. Space or Tab                    ⏱️ 2 sec
   [Screen: Dual-pane mode activates]

4. Navigate to CLAUDE.md           ⏱️ 2 sec
   [Screen: Orange AI file selected, markdown preview appears]

5. j j j (move down files)         ⏱️ 3 sec
   [Screen: Preview updates live as you navigate]

6. Right-click or F2               ⏱️ 2 sec
   [Screen: Context menu appears with options]

7. Press Esc                       ⏱️ 1 sec
   [Screen: Menu closes]

8. Press 3                         ⏱️ 1 sec
   [Screen: Switch to tree view]

9. Press → on a folder             ⏱️ 2 sec
   [Screen: Folder expands showing children]

10. Press F11                      ⏱️ 2 sec
    [Screen: Prompts mode - shows .prompts folder at top]

11. Navigate to TFE-Customization  ⏱️ 3 sec
    [Screen: Show customization prompts folder]

12. Press Esc                      ⏱️ 1 sec
    [Screen: Return to file browser]

13. Press F10                      ⏱️ 1 sec
    [Screen: TFE quits gracefully]

STOP RECORDING
Total: ~25-30 seconds
```

**File name:** `tfe-hero-workflow.mp4`

---

## 📹 SCRIPT 2: Prompts Library Workflow (20-30 sec) ⭐

**What it shows:** TFE's unique prompts feature

```
BEFORE RECORDING:
  cd ~/projects/TFE
  ./tfe
  Navigate to position showing some files

START RECORDING in OBS

1. Press F11                       ⏱️ 1 sec
   [Screen: Prompts mode activates, .prompts shown at top]

2. j j (navigate to .prompts)      ⏱️ 2 sec
   [Screen: .prompts folder visible in pink/orange]

3. Press Enter                     ⏱️ 1 sec
   [Screen: Enter .prompts folder]

4. j j (to TFE-Customization)      ⏱️ 2 sec
   [Screen: Navigate to customization folder]

5. Press Enter                     ⏱️ 1 sec
   [Screen: Show customization prompts]

6. j j (to add-tui-tool.prompty)   ⏱️ 2 sec
   [Screen: Navigate to a prompt file]

7. Press Enter                     ⏱️ 1 sec
   [Screen: Preview shows prompt with code examples]

8. Scroll preview (j j j j)        ⏱️ 3 sec
   [Screen: Show formatted prompt content]

9. Press F5                        ⏱️ 1 sec
   [Screen: Status message "Copied to clipboard"]

10. Wait 2 seconds                 ⏱️ 2 sec
    [Screen: Show status message visible]

11. Press Esc                      ⏱️ 1 sec
    [Screen: Return to file browser]

12. Press F10                      ⏱️ 1 sec
    [Screen: Quit]

STOP RECORDING
Total: ~18-20 seconds
```

**File name:** `prompts-library-workflow.mp4`

---

## 📹 SCRIPT 3: Context-Aware Help (15-20 sec)

**What it shows:** Intelligent F1 help that adapts to your context

```
BEFORE RECORDING:
  cd ~/projects/TFE
  ./tfe

START RECORDING in OBS

1. In single-pane mode             ⏱️ 1 sec
   [Screen: Normal file browser]

2. Press F1                        ⏱️ 2 sec
   [Screen: Help opens, auto-scrolled to "Navigation" section]

3. Wait 2 sec, then Esc            ⏱️ 3 sec
   [Screen: Show help is context-aware, then close]

4. Press Space (dual-pane)         ⏱️ 1 sec
   [Screen: Dual-pane mode activates]

5. Press F1                        ⏱️ 2 sec
   [Screen: Help opens, auto-scrolled to "Dual-Pane Mode" section]

6. Wait 2 sec, then Esc            ⏱️ 3 sec
   [Screen: Different section shown, then close]

7. Press F11 (prompts mode)        ⏱️ 1 sec
   [Screen: Prompts mode activates]

8. Press F1                        ⏱️ 2 sec
   [Screen: Help opens to "Prompts Library" section]

9. Wait 2 sec, then Esc            ⏱️ 3 sec
   [Screen: Prompts help shown, then close]

10. Press F10                      ⏱️ 1 sec
    [Screen: Quit]

STOP RECORDING
Total: ~21 seconds
```

**File name:** `context-aware-help.mp4`

---

## 📸 Screenshot Setup

### Hero Screenshot - Dual-Pane Markdown

```
Setup:
  1. cd ~/projects/TFE
  2. ./tfe
  3. Press Space or Tab (dual-pane)
  4. Navigate to CLAUDE.md (orange AI file)
  5. Make sure preview shows nice markdown formatting
  6. Resize terminal to ~1200x700
  7. Position terminal centered on screen

Take Screenshot:
  Windows: Win + Shift + S
  Linux: Flameshot or Spectacle
  Mac: Cmd + Shift + 4

Save as: hero-screenshot.png
```

### Context Menu Screenshot

```
Setup:
  1. ./tfe
  2. Navigate to a directory
  3. Right-click or press F2
  4. Context menu appears
  5. Make sure menu is fully visible

Take Screenshot (same as above)
Save as: context-menu-screenshot.png
```

---

## 🎬 After Recording - Convert to GIF

### Using FFmpeg (Best Quality)

```bash
# High quality GIF (2-4 MB)
ffmpeg -i recording.mp4 \
  -vf "fps=15,scale=1280:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse" \
  -loop 0 output.gif

# Medium quality (1-2 MB) - smaller file
ffmpeg -i recording.mp4 \
  -vf "fps=10,scale=1000:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse" \
  -loop 0 output.gif

# Low quality (500KB-1MB) - very small
ffmpeg -i recording.mp4 \
  -vf "fps=8,scale=800:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse" \
  -loop 0 output.gif
```

### Using Gifski (HIGHEST QUALITY)

```bash
# Extract frames from MP4
ffmpeg -i recording.mp4 frame%04d.png

# Create GIF with gifski
gifski -o output.gif --fps 15 --quality 90 frame*.png

# Cleanup
rm frame*.png
```

### Quick Convert Script

Save this as `convert-to-gif.sh`:

```bash
#!/bin/bash
# Usage: ./convert-to-gif.sh input.mp4 output.gif

INPUT=$1
OUTPUT=$2

if [ -z "$INPUT" ] || [ -z "$OUTPUT" ]; then
    echo "Usage: ./convert-to-gif.sh input.mp4 output.gif"
    exit 1
fi

echo "Converting $INPUT to $OUTPUT..."
ffmpeg -i "$INPUT" \
  -vf "fps=12,scale=1100:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse" \
  -loop 0 "$OUTPUT"

echo "✅ Done! Saved to $OUTPUT"
ls -lh "$OUTPUT"
```

Make executable: `chmod +x convert-to-gif.sh`

---

## 📦 Organize Files

```bash
# Move MP4 recordings to demo-content
mv ~/Videos/tfe-*.mp4 ~/projects/TFE/demo-content/

# Convert all recordings
cd ~/projects/TFE/demo-content
./convert-to-gif.sh tfe-hero-workflow.mp4 ../assets/tfe-hero.gif
./convert-to-gif.sh prompts-library-workflow.mp4 ../assets/prompts-workflow.gif
./convert-to-gif.sh context-aware-help.mp4 ../assets/context-help.gif

# Check file sizes
ls -lh ../assets/*.gif
```

---

## ✅ Checklist

**Before Recording:**
- [ ] OBS configured (see OBS_SETTINGS.txt)
- [ ] Terminal size ~1200x700
- [ ] Nerd Font with emoji support
- [ ] Terminal cleared (Ctrl+L)
- [ ] cd into correct directory
- [ ] Practice run completed

**While Recording:**
- [ ] Go SLOW - pause 1-2 sec between actions
- [ ] Let viewers see what's happening
- [ ] Show status messages fully
- [ ] Don't rush through menus

**After Recording:**
- [ ] Convert MP4 → GIF
- [ ] Check file size (<3MB ideal)
- [ ] Test GIF plays correctly
- [ ] Move to assets/ directory
- [ ] Delete MP4 if GIF looks good

---

**You've got this!** 🎬✨

Take your time, practice first, and remember: if you mess up, just start over!
The best demos are the ones where you go slow and let viewers see the magic. ✨
