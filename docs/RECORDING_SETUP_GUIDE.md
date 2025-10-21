# Quick Recording Setup Guide

## Step 1: Set Up Windows Terminal Profile

1. **Open Windows Terminal Settings:**
   - Press `Ctrl + ,` in Windows Terminal
   - Or click dropdown arrow → Settings

2. **Open settings.json:**
   - Click "Open JSON file" button in bottom-left corner
   - Or edit via the UI (Settings → Add a new profile)

3. **Add the TFE Recording Profile:**
   - Copy the profile from `WINDOWS_TERMINAL_RECORDING_PROFILE.json`
   - Paste it into the `"list"` array under `"profiles"`
   - **Important:** Change the GUID to something unique (or remove it to auto-generate)

4. **Save and test:**
   - Save settings.json
   - Open new tab dropdown → Select "TFE Recording"
   - Should open at 120x30 with your theme

## Step 2: Optimal Recording Dimensions

**Recommended terminal sizes for recording:**

| Use Case | Columns | Rows | Aspect Ratio | File Size |
|----------|---------|------|--------------|-----------|
| **Quick demo** | 100 | 25 | Compact | Smaller GIF |
| **Standard demo** | 120 | 30 | Balanced | Medium GIF |
| **Detailed demo** | 140 | 35 | Spacious | Larger GIF |

**Current profile:** 120x30 (recommended for most demos)

**To adjust:**
- Edit `initialCols` and `initialRows` in the profile
- Or manually resize terminal window during recording

## Step 3: Pre-Recording Checklist

Before you hit record:

```bash
# 1. Navigate to demo content
cd ~/projects/TFE

# 2. Clear terminal
clear

# 3. Quick test run
./tfe
# (Practice your demo flow, then quit with q)

# 4. Ready to record!
```

## Step 4: OBS Settings for This Profile

**Video Settings:**
- Base Resolution: 1920x1080
- Output Resolution: 1280x720 (or 1920x1080 for HD)
- FPS: 30

**Window Capture:**
- Select: "Windows Terminal"
- Crop to terminal window (no title bar if you want)

**Output:**
- Format: MP4
- Quality: High Quality (Medium works too, smaller files)

## Step 5: Recording Workflow

```
1. Open Windows Terminal → "TFE Recording" profile
2. cd ~/projects/TFE
3. clear
4. Start OBS recording (F12 or button)
5. ./tfe
6. [Do your demo following the script]
7. q (quit TFE)
8. Stop OBS recording
9. Check recording in ~/Videos
```

## Step 6: Convert to GIF (Optional)

```bash
# High quality (larger file):
ffmpeg -i ~/Videos/recording.mp4 \
  -vf "fps=15,scale=1200:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse" \
  -loop 0 assets/demo-new.gif

# Medium quality (recommended):
ffmpeg -i ~/Videos/recording.mp4 \
  -vf "fps=12,scale=1000:-1:flags=lanczos" \
  assets/demo-new.gif

# Small file (quick demos):
ffmpeg -i ~/Videos/recording.mp4 \
  -vf "fps=10,scale=800:-1" \
  assets/demo-new.gif
```

## Tips for Great Recordings

### Terminal Appearance:
✅ Hide scrollbar (profile has `scrollbarState: hidden`)
✅ Clean background (no transparency/acrylic)
✅ Bright cursor (white, bar shape)
✅ Consistent dimensions (120x30)

### During Recording:
🐢 **Go slower than normal** - viewers need time to process
⏸️ **Pause 1-2 seconds** after each action
🎯 **Focus on one feature** per recording
📦 **Show file icons** - navigate through varied file types
🤖 **Highlight unique features** - context-aware help, dual-pane, etc.

### Mistakes:
- Don't worry! Just stop recording and start again
- OBS recordings are quick - you can do multiple takes
- Practice runs help identify issues before recording

## Demo Script Templates

### Quick Feature Demo (20-30 seconds):
```
1. clear
2. tfe
3. Navigate down a few files (arrow keys)
4. Press F4 (preview mode)
5. Show preview content
6. ESC (exit preview)
7. q (quit)
```

### Dual-Pane Demo (30-45 seconds):
```
1. clear
2. tfe
3. Space (enter dual-pane)
4. Navigate files (preview updates automatically)
5. Tab (switch to preview pane)
6. Scroll preview (arrow keys)
7. Tab (back to file list)
8. Navigate more
9. ESC (exit dual-pane)
10. q (quit)
```

### Tree View Demo (25 seconds):
```
1. clear
2. tfe
3. Press 3 (tree view)
4. Navigate to directory
5. Press → (expand folder)
6. Show nested structure
7. Press ← (collapse)
8. q (quit)
```

### Context-Aware Help Demo (20 seconds):
```
1. clear
2. tfe
3. Press F1 (help - shows Navigation section)
4. ESC
5. Press F4 (preview mode)
6. Press F1 (help jumps to Preview section!)
7. ESC ESC
8. q (quit)
```

## File Size Optimization

If GIF is too large (>5MB):

```bash
# Option 1: Reduce dimensions
ffmpeg -i input.gif -vf "scale=600:-1" output-smaller.gif

# Option 2: Reduce FPS
ffmpeg -i input.gif -vf "fps=8" output-slower.gif

# Option 3: Optimize with gifsicle
gifsicle -O3 --lossy=80 --colors=128 -o output-opt.gif input.gif

# Option 4: Use WebM instead of GIF (much smaller!)
ffmpeg -i recording.mp4 -c:v libvpx-vp9 -b:v 0 -crf 30 demo.webm
# Note: GitHub supports WebM in markdown!
```

## Troubleshooting

### Terminal window is too small/large in OBS:
- Right-click source → Transform → Edit Transform
- Crop: Adjust left/top/right/bottom values
- Scale: Adjust to fit canvas

### Colors look wrong:
- Verify colorScheme is set to "CGA" (or your preferred scheme)
- Check OBS isn't applying filters

### Cursor not visible:
- Increase cursorHeight in profile
- Change cursorColor to bright color (#FFFFFF)
- Try cursorShape: "block" for maximum visibility

### GIF shows artifacts/dithering:
- Use the high-quality ffmpeg command (with palettegen)
- Increase scale (larger resolution)
- Use WebM format instead

---

**Ready to record?** 🎬

1. Open "TFE Recording" profile in Windows Terminal
2. Follow the checklist above
3. Record your demo with OBS
4. Convert to GIF (optional)
5. Update README.md with new demos!

Good luck! 🚀
