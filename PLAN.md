# TFE Development Plan

## Post-v1.0 Features & Improvements

### Known Issues to Address

#### xterm.js Emoji Spacing (CellBlocks compatibility)
**Problem:** TFE works perfectly in native terminals (Termux, WSL, etc.) but emoji toolbar buttons have incorrect spacing in web-based terminals that use xterm.js (CellBlocks, ttyd, wetty).

**Cause:** xterm.js renders emoji widths inconsistently compared to native terminals. TFE calculates button positions assuming emojis are 2 chars wide, but xterm.js sometimes renders them differently.

**Impact:**
- ✅ Works perfectly: Termux (Android), WSL Terminal, GNOME Terminal, iTerm2
- ❌ Broken spacing: CellBlocks (xterm.js), ttyd, wetty, web SSH clients
- ~95% of users unaffected (native terminal users)

**Solutions (v1.1):**
1. **Option A - Auto-detect ASCII mode:**
   ```go
   // Detect xterm.js environment
   if os.Getenv("TERM_PROGRAM") == "xterm.js" {
       // Use ASCII buttons: [H][*][E][P][>][?][#][G][T]
   }
   ```

2. **Option B - CLI flag:**
   ```bash
   tfe --ascii-mode  # For web terminals
   ```

3. **Option C - Smart width calculation:**
   ```go
   import "github.com/mattn/go-runewidth"
   // Calculate actual rendered width dynamically
   ```

**Recommended:** Option C (most robust) or Option B (quickest)

**Priority:** Medium (affects remote access via CellBlocks but not primary use cases)

---

## v1.0 Launch Checklist

- [x] Dual-pane accordion layout fixes
- [x] Vertical split for detail view
- [x] Mouse accuracy in all modes
- [x] Space bar command mode fix
- [x] Documentation complete
- [ ] Final testing in Termux
- [ ] Final testing in WSL
- [ ] Launch announcement
- [ ] Hacker News post

---

## Future Feature Ideas

### Mobile Optimizations
- Haptic feedback on touch events (Termux)
- Swipe gestures for navigation
- Mobile-specific keybindings

### Container Integration (Post-orchestration experiments)
- F12: Container mode showing running containers
- Context menu: "Open in Container"
- Safety indicators (green=host, red=container)
- Docker volume mounting for file access

### Advanced Prompts Library
- Global + project prompts merging
- Prompt templates with variables
- Team-shared prompt collections
- Version control for prompts

---

*Last updated: 2025-10-21 04:00*
