# VHS Quick Reference for TFE

## 📝 Common VHS Commands

### Keyboard Input
```bash
Type "text"              # Types text (simulates typing)
Type@500ms "slow text"   # Types slowly (500ms per char)
Enter                    # Presses Enter
Escape                   # Presses Escape
Backspace                # Presses Backspace
Tab                      # Tab key
Space                    # Space bar
```

### Navigation Keys
```bash
Up / Down / Left / Right   # Arrow keys
PageUp / PageDown          # Page navigation
Home / End                 # Home/End keys
```

### Function Keys
```bash
F1 F2 F3 F4 F5 F6 F7 F8 F9 F10 F11 F12
```

### Control Keys
```bash
Ctrl+C    # Interrupt
Ctrl+D    # EOF
Ctrl+Z    # Suspend
Ctrl+A    # Ctrl+A (any letter works)
```

### Timing
```bash
Sleep 2s      # Wait 2 seconds
Sleep 500ms   # Wait 500 milliseconds
```

### Screen Control
```bash
Hide          # Hide next commands from output
Show          # Show commands again
Screenshot demo.png   # Take screenshot
```

## 🎨 Settings

```bash
# Output
Output demo.gif           # GIF output
Output demo.mp4           # Video output
Output demo.webm          # WebM output

# Appearance
Set FontSize 14           # Font size
Set Width 1200            # Terminal width
Set Height 700            # Terminal height
Set Theme "Dracula"       # Color theme
Set TypingSpeed 100ms     # Default typing speed
Set FrameRate 30          # FPS (lower = smaller file)

# Padding
Set Padding 10            # Window padding
Set Margin 20             # Window margin
Set MarginFill "#000000"  # Margin color
```

## 🎭 Themes

Popular themes:
- `Dracula` - Dark purple/pink
- `Nord` - Cool blue tones
- `Catppuccin Mocha` - Warm pastels
- `Monokai` - Classic dark theme
- `Solarized Dark` - Blue/cyan tones
- `GitHub Dark` - GitHub's dark theme
- `One Dark` - Atom's dark theme
- `Tokyo Night` - Blue/purple night theme

Full list: https://github.com/charmbracelet/vhs/tree/main/themes

## 📐 Size Guidelines

### For GitHub README
```bash
Set Width 1200
Set Height 700
Set FontSize 14
Set FrameRate 30
```

### For Smaller Embeds
```bash
Set Width 1000
Set Height 600
Set FontSize 12
Set FrameRate 20   # Smaller file size
```

### For Hero/Banner
```bash
Set Width 1400
Set Height 800
Set FontSize 14
Set FrameRate 30
```

## 🎯 File Size Tips

To reduce GIF size:
1. **Lower resolution:** `Set Width 1000`
2. **Lower framerate:** `Set FrameRate 20`
3. **Shorter duration:** Fewer actions, shorter sleeps
4. **Optimize after:** `gifsicle -O3 --lossy=80`

Target sizes:
- Small demo: **< 500 KB**
- Medium demo: **< 1 MB**
- Large/complete demo: **< 2 MB**

## 🔄 Example Template

```bash
# Description of what this demo shows
Output ../assets/demo-name.gif

# Settings
Set FontSize 14
Set Width 1200
Set Height 700
Set Theme "Dracula"
Set TypingSpeed 100ms

# Launch app
Type "./tfe"
Enter
Sleep 2s

# Do something
Type "jjj"
Sleep 1s

# Do more things
Enter
Sleep 2s

# Clean exit
Type "F10"
Sleep 500ms
```

## ⚙️ Running VHS

```bash
# Generate one demo
vhs demo.tape

# Generate with custom ttyd
vhs --ttyd /path/to/ttyd demo.tape

# Test without output (faster)
vhs --dry-run demo.tape
```

## 🐛 Debugging

### VHS hangs
- Increase sleep times
- Check app actually runs: `./tfe`
- Add explicit quits: `Ctrl+C`

### Output is blank
- App might need more startup time: `Sleep 3s`
- Check terminal size is adequate
- Try simpler commands first

### Timing is off
- Add more `Sleep` commands
- Increase sleep durations
- Terminal rendering takes time

### Colors wrong
- Try different theme
- Check VHS version: `vhs --version`
- Update VHS: `go install github.com/charmbracelet/vhs@latest`

## 📚 Resources

- **VHS GitHub:** https://github.com/charmbracelet/vhs
- **Examples:** https://github.com/charmbracelet/vhs/tree/main/examples
- **Themes:** https://github.com/charmbracelet/vhs/tree/main/themes
- **Documentation:** https://github.com/charmbracelet/vhs#readme

## 🎬 TFE-Specific Tips

### Show multiple features quickly
```bash
# Quick tour
Type "jjj"    # Navigate
Sleep 500ms
Type "3"      # Tree view
Sleep 1s
Type " "      # Dual-pane
Sleep 1s
Escape        # Exit
```

### Highlight AI context files
```bash
# Navigate to show orange .claude/ folder
Type "j"      # Find CLAUDE.md
Sleep 1s
Enter         # Preview (shows markdown)
Sleep 2s
```

### Demonstrate smooth workflow
```bash
# Realistic usage
Type "jjj"              # Browse
Sleep 1s
Type "3"                # Tree mode
Sleep 1s
Right                   # Expand folder
Sleep 1s
Type "jj"               # Navigate inside
Sleep 1s
Enter                   # Preview file
Sleep 2s
Escape                  # Back to list
Sleep 1s
```

---

**Need help?** Check `demos/README.md` or open an issue!
