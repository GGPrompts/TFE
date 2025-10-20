# Creating the Hero Screenshot

This guide shows how to create the perfect hero screenshot for TFE's README, showcasing AI workflow integration.

## 🎯 Goal

Create a split terminal screenshot showing:
- **Left terminal:** TFE displaying CLAUDE.md with beautiful markdown rendering
  - Shows **orange color** for AI context files (.claude, CLAUDE.md)
  - Shows dual-pane mode with live preview
  - Demonstrates markdown rendering
- **Right terminal:** Claude Code at the welcome screen
  - Shows the AI assistant interface
  - Demonstrates real-world AI coding workflow

## 📋 Setup

### 1. Install Screenshot Tool

**Flameshot (recommended):**
```bash
sudo apt install flameshot

# Launch
flameshot gui
```

**Alternative: scrot**
```bash
sudo apt install scrot

# Full screen
scrot screenshot.png

# Select area
scrot -s screenshot.png
```

### 2. Prepare Terminals

**Terminal Layout:**
```
┌─────────────────────────┬─────────────────────────┐
│                         │                         │
│   TFE                   │   Claude Code           │
│   (dual-pane mode)      │   (welcome screen)      │
│   CLAUDE.md visible     │                         │
│                         │                         │
└─────────────────────────┴─────────────────────────┘
```

### 3. Launch TFE in Left Terminal

```bash
cd /path/to/your/project  # Project with CLAUDE.md
./tfe

# Once in TFE:
# 1. Navigate to CLAUDE.md (arrow keys or j/k)
# 2. Press Space or Tab to enter dual-pane mode
# 3. Position cursor so CLAUDE.md is visible with preview
```

**What to show in TFE:**
- File list on left showing:
  - `CLAUDE.md` (orange icon 🤖)
  - `.claude/` folder (orange icon)
  - Other project files
- Preview pane on right showing:
  - Rendered markdown from CLAUDE.md
  - Nice formatting (headers, bullets, code blocks)

### 4. Launch Claude Code in Right Terminal

```bash
# In a second terminal
claude-code

# Or just show the welcome screen
# Make sure it's visible and looks good
```

## 📸 Taking the Screenshot

### With Flameshot:
1. Run: `flameshot gui`
2. Select the entire terminal window (both panes)
3. Optional: Add annotations:
   - Arrow pointing to orange AI file icon
   - Text: "AI Context Files"
   - Text: "Beautiful Markdown"
   - Text: "Claude Code Integration"
4. Save as `assets/hero-screenshot.png`

### With scrot:
```bash
# Select area manually
scrot -s assets/hero-screenshot.png

# Or delay and position windows
scrot -d 5 assets/hero-screenshot.png
```

## 🎨 Screenshot Checklist

Before taking the screenshot, verify:

- ✅ TFE is in dual-pane mode
- ✅ CLAUDE.md is selected/visible (orange icon visible)
- ✅ Preview shows nicely formatted markdown
- ✅ Claude Code welcome screen is visible
- ✅ Terminal colors look good (not washed out)
- ✅ Font size is readable (14-16pt recommended)
- ✅ No sensitive information visible (paths, tokens, etc.)
- ✅ Both terminals fit in frame
- ✅ No UI glitches or artifacts

## 🖼️ Example Layout

```
┌─────────────────────────────────────────────────────────────────┐
│ TFE - Terminal File Explorer                                    │
│                                                                   │
│ Files               │  Preview: CLAUDE.md                        │
│ ──────────────────  │  ────────────────────────────             │
│ 📁 .claude/         │  # TFE Architecture Guide                  │
│ 🤖 CLAUDE.md ◄──────┼─ This document describes...               │
│ 📄 README.md        │                                            │
│ 📄 main.go          │  ## Core Principle                         │
│ 📄 types.go         │  When adding features, maintain            │
│ 📁 docs/            │  this modular architecture...              │
│                     │                                            │
└─────────────────────┴────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ Claude Code                                                      │
│                                                                   │
│ Welcome to Claude Code!                                          │
│                                                                   │
│ I'm Claude, your AI pair programmer.                             │
│ How can I help you today?                                        │
│                                                                   │
│ >                                                                │
└─────────────────────────────────────────────────────────────────┘
```

## 🔧 Optimization

After taking the screenshot:

```bash
# Optimize PNG
sudo apt install optipng
optipng -o7 assets/hero-screenshot.png

# Or use pngcrush
sudo apt install pngcrush
pngcrush -reduce assets/hero-screenshot.png assets/hero-screenshot-opt.png
mv assets/hero-screenshot-opt.png assets/hero-screenshot.png
```

**Target size:** < 500 KB for fast loading

## 📝 Using in README

```markdown
# TFE - Terminal File Explorer

<p align="center">
  <img src="assets/hero-screenshot.png" alt="TFE with Claude Code" width="900">
</p>

*TFE integrates seamlessly with AI coding workflows, making it easy to manage context files and navigate projects while pair programming with Claude Code.*
```

## 💡 Tips

1. **Use a nice color scheme** - Dark themes (Dracula, Nord) look professional
2. **Clean up terminal** - Close unnecessary tabs/panes
3. **Readable font size** - 14-16pt for screenshots
4. **Good contrast** - Make sure text is legible
5. **Highlight key features** - Orange AI file icon should be visible
6. **Show real content** - Use actual CLAUDE.md from TFE project
7. **Professional look** - No terminal clutter, clean UI

## 🎯 Alternative: Animated GIF

Instead of a static screenshot, create an animated GIF showing:
1. Opening TFE
2. Navigating to CLAUDE.md (orange icon visible)
3. Entering dual-pane mode
4. Scrolling the preview
5. Switching to Claude Code

Use the VHS tape format:
```bash
# Save as: demos/11-hero-demo.tape
Output ../assets/hero-demo.gif
Set Width 1600
Set Height 900
# ... (VHS commands)
```

## 📚 Examples to Look At

Before creating yours, check out other CLI tool screenshots:
- **lazygit** - Split view with clean UI
- **btop** - Professional terminal aesthetics
- **ranger** - Dual-pane file manager look

Look for:
- Professional color scheme
- Clear visual hierarchy
- Key features highlighted
- No clutter

---

**Questions?** Open an issue: https://github.com/GGPrompts/TFE/issues
