# TFE Development Plan

## Pre-Launch Fixes (Before v1.0)

### Note on "Security Vulnerabilities"

Earlier reviews flagged "command injection" and "path traversal" as security issues. **These are not vulnerabilities** in TFE's threat model:

- **Command execution**: User is already in a terminal with full shell access. Sanitizing would just be annoying.
- **Path traversal**: That's the entire point of a file browser. Users can navigate anywhere they have permissions.
- TFE is a **local tool** where the user is the operator, not an attacker.

See CLAUDE.md "Security & Threat Model" section for full explanation.

### Completed ✅

1. **File Handle Leak** (2025-10-24)
   - All `os.Open()` calls now have `defer file.Close()`
   - Files: `file_operations.go:119,647,2147` and `trash.go:347,354`
   - This was a resource leak, not a security issue

2. **Context Menu Overlay Fix** (2025-10-24)
   - Context menus now preserve content on both sides (not just left)
   - Applied same ANSI-aware overlay logic from dropdown menus
   - File: `view.go:560-583`

### Critical Issues (MUST Fix - Estimated: 4 hours)

#### Edge Cases
3. **Circular Symlink Detection** [1hr]
   - File: `file_operations.go:948-959`
   - Issue: Infinite loop with circular symlinks (dir1 → dir2 → dir1)
   - Fix: Track visited paths, limit symlink depth to 40

4. **Cross-Device Trash Move** [1hr]
   - File: `trash.go:135-137`
   - Issue: Fails across mount points with cryptic error
   - Fix: Detect `syscall.EXDEV`, fallback to copy+delete

#### Documentation
5. **Split CHANGELOG.md** [30min]
   - File: `CHANGELOG.md` (344 lines → approaching 350 limit)
   - Fix: Move v0.4.0 to CHANGELOG3.md, keep v0.5.0 + Unreleased in main

6. **Create FAQ.md + CONTRIBUTING.md** [3hrs]
   - Missing: User troubleshooting guide and contributor documentation
   - Fix: Create FAQ with 10 common issues, CONTRIBUTING with dev setup
   - Add: README screenshots (3-5 images)

---

### High Priority (Should Fix - Estimated: 7 hours)

#### Performance Optimization
7. **Unbounded Data Structures** [2hrs]
   - Files: `types.go:22` (expandedDirs), command history, preview cache
   - Issue: Memory leaks in long-running sessions
   - Fix: Implement LRU cache (max 500 entries), rotate command history

8. **Glamour Markdown Timeout** [30min]
   - File: `file_operations.go:314-323`
   - Issue: Malformed markdown can hang UI indefinitely
   - Fix: Add 2-second timeout context to Glamour rendering

#### Mobile UX (Narrow Terminals)
9. **Detail View Width on Narrow Terminals** [2hrs]
   - File: `render_file_list.go:185`
   - Issue: Forces 120-column width on 80-column Termux screens
   - Fix: Dynamic column widths (name 25%, size 15%, modified 20%, type 40%)

10. **Header Scrolling Misalignment** [1.5hrs]
    - File: `render_file_list.go:302-330`
    - Issue: Headers don't align with data rows in horizontal scroll
    - Fix: Unify scrolling logic between headers and data rows

#### Code Quality
11. **Header Duplication (DRY Violation)** [45min]
    - Files: `view.go:64`, `render_preview.go:816`
    - Issue: Menu bar logic duplicated in two locations
    - Fix: Extract `renderHeader()` shared function

#### UX Polish
12. **Auto-Switch Dual-Pane View** [20min]
    - File: `update_keyboard.go:1381, 1400`
    - Issue: Error message instead of auto-switching from Detail to List
    - Fix: Auto-switch to List view when entering dual-pane

13. **Empty Directory Message** [15min]
    - Issue: Blank screen when entering empty directory
    - Fix: Add "📂 Empty directory - Press ← to go back" message

14. **Complete HOTKEYS.md** [45min]
    - Missing: Ctrl+F, Ctrl+P, /, Ctrl+W, Alt+D, mouse guide
    - Fix: Add missing shortcuts and mouse interaction section

---

### Medium Priority (Fix After Launch - Estimated: 10 hours)

15. **Filename ANSI Escape Sanitization** [30min]
    - Issue: Files named with `\x1b[31mRED\x1b[0m` corrupt terminal colors
    - Fix: Strip ANSI codes from filenames before display

16. **Prompt Variable Limit** [45min]
    - Issue: Prompts with >20 variables overflow UI
    - Fix: Limit to 20 variables, show warning for excess

17. **Directory Entry Limit** [30min]
    - Issue: No limit on entries, could OOM on `/proc` or huge dirs
    - Fix: Cap at 10,000 entries with warning message

18. **Preview Cache Mtime Invalidation** [1hr]
    - File: `file_operations.go:262-264`
    - Issue: Stale previews when files are modified externally
    - Fix: Check mtime before returning cached preview

19. **Extract Keyboard Sub-Handlers** [2hrs]
    - File: `update_keyboard.go:398-662` (264-line mega-function)
    - Fix: Split into `handleNavigationKeys()`, `handleFileOperationKeys()`, `handleDisplayModeKeys()`

20. **Magic Numbers to Constants** [1hr]
    - Files: Multiple (header heights, padding values scattered)
    - Fix: Define in `types.go`: `headerHeight`, `previewPadding`, etc.

21. **Status Message Persistence** [30min]
    - Issue: Error messages auto-dismiss after 3 seconds (too fast)
    - Fix: Keep errors visible until next action

22. **Search Mode Visual Indicator** [1hr]
    - Issue: Search mode only shows in status bar (easy to miss)
    - Fix: Add visual search box overlay like dialog system

23. **README Expansion** [1hr]
    - Missing: Prerequisites, installation verification, quick start tutorial
    - Fix: Add "Getting Started" section with 5-step walkthrough

24. **Troubleshooting Guide** [1hr]
    - Missing: Common problems and solutions
    - Fix: Create TROUBLESHOOTING.md with diagnostics

---

### Low Priority / Future Enhancements

25. UTF-16/UTF-32 BOM detection for text files
26. Tree view depth limit documentation
27. Command history rotation/archiving (keep last 1000)
28. Performance metrics/telemetry
29. Icon lookup map optimization (switch → map)
30. Video demo (high-quality OBS recording at 1080p)
31. Package manager submissions (Homebrew, AUR, Nix)
32. Comparison guide (TFE vs ranger vs nnn vs mc)

---

## Known Issues (Post-v1.0)

### xterm.js Emoji Spacing (CellBlocks compatibility)
**Problem:** TFE works perfectly in native terminals (Termux, WSL, etc.) but emoji toolbar buttons have incorrect spacing in web-based terminals using xterm.js (CellBlocks, ttyd, wetty).

**Cause:** xterm.js renders emoji widths inconsistently. TFE assumes emojis are 2 chars wide, but xterm.js sometimes differs.

**Impact:**
- ✅ Works perfectly: Termux (Android), WSL Terminal, GNOME Terminal, iTerm2
- ❌ Broken spacing: CellBlocks (xterm.js), ttyd, wetty, web SSH clients
- ~95% of users unaffected (native terminal users)

**Solutions (v1.1):**
1. **Option A - Auto-detect xterm.js:** Check `$TERM_PROGRAM` env var
2. **Option B - CLI flag:** `tfe --ascii-mode` for web terminals
3. **Option C - Smart width calc:** Use `go-runewidth` dynamically

**Recommended:** Option C (most robust) or Option B (quickest)
**Priority:** Medium (affects remote access but not primary use cases)

---

## v1.0 Launch Checklist

**Completed ✅**
- [x] Fix file handle leak (resource management)
- [x] Fix context menu overlay (preserve content on both sides)
- [x] Create CONTRIBUTING.md
- [x] Update documentation (remove false security concerns)

**Pre-Launch (Critical):**
- [ ] Split CHANGELOG.md → CHANGELOG3.md (30 min)
- [ ] Create FAQ.md (1.5 hours)
- [ ] Add README screenshots (30 min)
- [ ] Fix circular symlink detection (1 hour)
- [ ] Fix cross-device trash move (1 hour)
- [ ] Record high-quality demo video at 1080p (OBS, proper settings)

**Testing:**
- [ ] Final testing in Termux (Android mobile)
- [ ] Final testing in WSL2
- [ ] Test circular symlink handling
- [ ] Test cross-device trash moves
- [ ] Test all keyboard shortcuts
- [ ] Verify context menu overlays work correctly

**Launch (Monday Morning):**
- [ ] Tag v1.0 release
- [ ] GitHub release notes
- [ ] Upload demo video to YouTube
- [ ] Create optimized GIF for Imgur
- [ ] Post to Hacker News (Show HN) - 8-10 AM EST
- [ ] Post to r/commandline, r/golang, r/linux

---

## Future Feature Ideas (v1.1+)

### Mobile Optimizations
- Haptic feedback on touch events (Termux)
- Swipe gestures for navigation
- Mobile-specific keybindings

### Container Integration
- F12: Container mode showing running containers
- Context menu: "Open in Container"
- Safety indicators (green=host, red=container)
- Docker volume mounting for file access

### Advanced Prompts Library
- Global + project prompts merging
- Prompt templates with variables (✅ done!)
- Team-shared prompt collections
- Version control for prompts

---

*Plan updated: 2025-10-24 (Removed false security concerns, updated completed items)*
*Line count: Target <400 lines (currently 234 - 59% capacity)*
