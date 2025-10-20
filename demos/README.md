# TFE Demo Generator

This folder contains VHS tape scripts that automatically generate demo GIFs showcasing TFE features.

**Theme:** All demos use the **CGA (Color Graphics Adapter)** theme - a retro DOS color scheme with bold primary colors on pure black. This matches TFE's development environment and gives a classic, nostalgic computing aesthetic. 🎮

## 🎬 What is VHS?

VHS is a tool that **automatically types commands and records the output**. You don't type anything manually - the `.tape` scripts do everything!

**Install VHS:**
```bash
go install github.com/charmbracelet/vhs@latest
# Or
sudo snap install vhs
```

## 📦 Demo Files

| File | Feature | Output |
|------|---------|--------|
| `01-navigation.tape` | Basic navigation with arrow keys & j/k | `demo-navigation.gif` |
| `02-preview-mode.tape` | Full-screen file preview & scrolling | `demo-preview.gif` |
| `03-dual-pane.tape` | Split-screen browsing with live preview | `demo-dual-pane.gif` |
| `04-tree-view.tape` | Hierarchical folder exploration | `demo-tree-view.gif` |
| `05-view-modes.tape` | Switching between List/Detail/Tree views | `demo-view-modes.gif` |
| `06-context-menu.tape` | F2 context menu navigation | `demo-context-menu.gif` |
| `07-search.tape` | Directory filtering with / search | `demo-search.gif` |
| `08-help-system.tape` | Context-aware F1 help | `demo-help.gif` |
| `09-file-operations.tape` | Creating directories with F7 | `demo-file-ops.gif` |
| `10-complete-workflow.tape` | Full feature showcase | `demo-complete.gif` |

## 🚀 Quick Start

### Generate All Demos
```bash
cd demos
./generate-all.sh
```

This will:
1. ✅ Run all `.tape` scripts
2. 🎬 Generate GIFs in `assets/`
3. 🔧 Optimize file sizes (if gifsicle installed)
4. 📊 Show results summary

### Generate One Demo
```bash
cd demos
vhs 01-navigation.tape
```

Output appears in `../assets/demo-navigation.gif`

## 🎨 Customizing Demos

Each `.tape` file is a simple script. Edit them to change:

**Timing:**
```bash
Sleep 2s        # Wait 2 seconds
Sleep 500ms     # Wait 500 milliseconds
```

**Theme:**
```bash
Set Theme "cga-theme.json"      # Default: CGA (retro DOS colors)
Set Theme "Dracula"             # Alternative: Purple/pink
Set Theme "Catppuccin Mocha"    # Alternative: Warm pastels
Set Theme "Nord"                # Alternative: Cool blues
```

**Size:**
```bash
Set Width 1200    # Smaller = smaller GIF file
Set Height 700
Set FontSize 14
```

**Typing Speed:**
```bash
Set TypingSpeed 100ms     # Faster typing
Type@500ms "slow text"    # Slow typing for this text
```

## 🔧 Optimization

### Install gifsicle (recommended)
```bash
sudo apt install gifsicle
```

The `generate-all.sh` script will automatically optimize GIFs if gifsicle is installed.

### Manual Optimization
```bash
gifsicle -O3 --lossy=80 --colors 256 -o optimized.gif input.gif
```

**Targets:**
- Individual GIFs: **< 1 MB** each
- Total assets: **< 5 MB**

## 📸 Screenshots

For static screenshots, use:

**Flameshot (recommended):**
```bash
sudo apt install flameshot
flameshot gui
```

**scrot (simple):**
```bash
sudo apt install scrot
scrot -s screenshot.png
```

### Hero Screenshot Idea
Create a split terminal screenshot showing:
- **Left side:** TFE in dual-pane mode with CLAUDE.md open
  - Shows orange color for AI context files
  - Shows beautiful markdown rendering
- **Right side:** Claude Code at welcome screen
  - Demonstrates AI workflow integration

This showcases TFE's purpose: easy file access and context management for AI coding workflows.

## 📝 Using Demos in README

```markdown
# TFE - Terminal File Explorer

![TFE Demo](assets/demo-complete.gif)

## Features

### 🎯 Dual-Pane Mode
![Dual-Pane](assets/demo-dual-pane.gif)

### 🌲 Tree View
![Tree View](assets/demo-tree-view.gif)
```

## 🐛 Troubleshooting

**VHS hangs or produces blank GIF:**
- Increase `Sleep` times (terminal needs time to render)
- Check that TFE builds successfully: `go build`
- Run tape manually to see errors: `vhs 01-navigation.tape`

**GIF is too large:**
- Reduce size: `Set Width 1000` and `Set Height 600`
- Reduce FPS: `Set FrameRate 20`
- Optimize: `gifsicle -O3 --lossy=80 -o out.gif in.gif`
- Shorten demo: fewer actions, shorter sleeps

**Colors look wrong:**
- Try different theme: `Set Theme "Dracula"`
- Check terminal color support
- VHS renders in its own terminal, not your actual terminal

## 📚 VHS Documentation

- **GitHub:** https://github.com/charmbracelet/vhs
- **Themes:** https://github.com/charmbracelet/vhs/tree/main/themes
- **Examples:** https://github.com/charmbracelet/vhs/tree/main/examples

## 🎯 Tips

1. **Keep demos short** - 10-15 seconds per feature
2. **Add pauses** - Let viewers see what happened (`Sleep 1s` after actions)
3. **One feature per demo** - Easier to maintain and embed
4. **Test locally first** - Run `vhs tape.tape` before committing
5. **Optimize GIFs** - Use gifsicle to reduce file sizes
6. **Version control** - Commit `.tape` files, GIFs go in `assets/`

## 🔄 Regenerating After Changes

After updating TFE features:
```bash
# Rebuild TFE
go build

# Regenerate affected demos
cd demos
vhs 03-dual-pane.tape  # If dual-pane changed
vhs 08-help-system.tape  # If help changed

# Or regenerate all
./generate-all.sh
```

---

**Questions?** Open an issue: https://github.com/GGPrompts/TFE/issues
