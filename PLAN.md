# TFE Development Plan

## Pre-Launch Fixes (Before v1.0)

### Critical Issues (MUST Fix - Estimated: 6 hours)

#### Security Vulnerabilities
1. **Command Injection in executeCommand()** [2hrs]
   - File: `command.go:27-47`
   - Issue: User input passed to `/bin/sh -c` without sanitization
   - Fix: Implement command allowlist or add confirmation for dangerous patterns
   - Exploit: `ls; rm -rf ~` or `cat /etc/passwd`

2. **Path Traversal in loadFiles()** [1hr]
   - File: `file_operations.go:41-57`
   - Issue: No validation against `../../etc` style traversal
   - Fix: Validate paths with `filepath.Clean()`, check against allowed boundaries

3. **Command Injection in openEditor()** [30min]
   - File: `editor.go:68-91`
   - Issue: Filenames starting with `-` can inject editor arguments
   - Fix: Validate filename doesn't start with `-`, use absolute paths

#### Performance Critical
4. **File Handle Leak in loadPreview()** [5min] ⚡ ONE LINE FIX
   - File: `file_operations.go:298-299`
   - Issue: Missing `defer file.Close()` - crashes after ~1000 previews
   - Fix: Add `defer file.Close()` after `os.Open(path)`

#### Edge Cases
5. **Circular Symlink Detection** [1hr]
   - File: `file_operations.go:948-959`
   - Issue: Infinite loop with circular symlinks (dir1 → dir2 → dir1)
   - Fix: Track visited paths, limit symlink depth to 40

6. **Cross-Device Trash Move** [1hr]
   - File: `trash.go:135-137`
   - Issue: Fails across mount points with cryptic error
   - Fix: Detect `syscall.EXDEV`, fallback to copy+delete

#### Documentation
7. **Split CHANGELOG.md** [30min]
   - File: `CHANGELOG.md` (344 lines → approaching 350 limit)
   - Fix: Move v0.4.0 to CHANGELOG3.md, keep v0.5.0 + Unreleased in main

8. **Create FAQ.md + CONTRIBUTING.md** [3hrs]
   - Missing: User troubleshooting guide and contributor documentation
   - Fix: Create FAQ with 10 common issues, CONTRIBUTING with dev setup
   - Add: README screenshots (3-5 images)

---

### High Priority (Should Fix - Estimated: 8 hours)

#### Security Hardening
9. **File Permissions for History/Favorites** [1hr]
   - Files: `command.go:60-80`, `favorites.go:16-36`
   - Issue: Stored with `0644` (world-readable) instead of `0600`
   - Fix: Change permissions, filter sensitive patterns from history

#### Performance Optimization
10. **Unbounded Data Structures** [2hrs]
    - Files: `types.go:22` (expandedDirs), command history, preview cache
    - Issue: Memory leaks in long-running sessions
    - Fix: Implement LRU cache (max 500 entries), rotate command history

11. **Glamour Markdown Timeout** [30min]
    - File: `file_operations.go:314-323`
    - Issue: Malformed markdown can hang UI indefinitely
    - Fix: Add 2-second timeout context to Glamour rendering

#### Mobile UX (Narrow Terminals)
12. **Detail View Width on Narrow Terminals** [2hrs]
    - File: `render_file_list.go:185`
    - Issue: Forces 120-column width on 80-column Termux screens
    - Fix: Dynamic column widths (name 25%, size 15%, modified 20%, type 40%)

13. **Header Scrolling Misalignment** [1.5hrs]
    - File: `render_file_list.go:302-330`
    - Issue: Headers don't align with data rows in horizontal scroll
    - Fix: Unify scrolling logic between headers and data rows

#### Code Quality
14. **Header Duplication (DRY Violation)** [45min]
    - Files: `view.go:64`, `render_preview.go:816`
    - Issue: Menu bar logic duplicated in two locations
    - Fix: Extract `renderHeader()` shared function

#### UX Polish
15. **Auto-Switch Dual-Pane View** [20min]
    - File: `update_keyboard.go:1381, 1400`
    - Issue: Error message instead of auto-switching from Detail to List
    - Fix: Auto-switch to List view when entering dual-pane

16. **Empty Directory Message** [15min]
    - Issue: Blank screen when entering empty directory
    - Fix: Add "📂 Empty directory - Press ← to go back" message

17. **Complete HOTKEYS.md** [45min]
    - Missing: Ctrl+F, Ctrl+P, /, Ctrl+W, Alt+D, mouse guide
    - Fix: Add missing shortcuts and mouse interaction section

---

### Medium Priority (Fix After Launch - Estimated: 10 hours)

18. **Filename ANSI Escape Sanitization** [30min]
    - Issue: Files named with `\x1b[31mRED\x1b[0m` corrupt terminal colors
    - Fix: Strip ANSI codes from filenames before display

19. **Prompt Variable Limit** [45min]
    - Issue: Prompts with >20 variables overflow UI
    - Fix: Limit to 20 variables, show warning for excess

20. **Directory Entry Limit** [30min]
    - Issue: No limit on entries, could OOM on `/proc` or huge dirs
    - Fix: Cap at 10,000 entries with warning message

21. **Preview Cache Mtime Invalidation** [1hr]
    - File: `file_operations.go:262-264`
    - Issue: Stale previews when files are modified externally
    - Fix: Check mtime before returning cached preview

22. **Extract Keyboard Sub-Handlers** [2hrs]
    - File: `update_keyboard.go:398-662` (264-line mega-function)
    - Fix: Split into `handleNavigationKeys()`, `handleFileOperationKeys()`, `handleDisplayModeKeys()`

23. **Magic Numbers to Constants** [1hr]
    - Files: Multiple (header heights, padding values scattered)
    - Fix: Define in `types.go`: `headerHeight`, `previewPadding`, etc.

24. **Status Message Persistence** [30min]
    - Issue: Error messages auto-dismiss after 3 seconds (too fast)
    - Fix: Keep errors visible until next action

25. **Search Mode Visual Indicator** [1hr]
    - Issue: Search mode only shows in status bar (easy to miss)
    - Fix: Add visual search box overlay like dialog system

26. **README Expansion** [1hr]
    - Missing: Prerequisites, installation verification, quick start tutorial
    - Fix: Add "Getting Started" section with 5-step walkthrough

27. **Troubleshooting Guide** [1hr]
    - Missing: Common problems and solutions
    - Fix: Create TROUBLESHOOTING.md with diagnostics

---

### Low Priority / Future Enhancements

28. UTF-16/UTF-32 BOM detection for text files
29. Tree view depth limit documentation
30. Command history rotation/archiving (keep last 1000)
31. Performance metrics/telemetry
32. Icon lookup map optimization (switch → map)
33. Video demo (asciinema recording)
34. Package manager submissions (Homebrew, AUR, Nix)
35. Comparison guide (TFE vs ranger vs nnn vs mc)

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

**Pre-Launch (Critical):**
- [ ] Fix 6 critical security/performance issues (6 hours)
- [ ] Split CHANGELOG.md → CHANGELOG3.md (30 min)
- [ ] Create FAQ.md (1.5 hours)
- [ ] Create CONTRIBUTING.md (1 hour)
- [ ] Add README screenshots (30 min)

**Testing:**
- [ ] Final testing in Termux (Android mobile)
- [ ] Final testing in WSL2
- [ ] Test command injection scenarios
- [ ] Test circular symlink handling
- [ ] Test cross-device trash moves
- [ ] Test all keyboard shortcuts

**Launch:**
- [ ] Tag v1.0 release
- [ ] GitHub release notes
- [ ] Launch announcement
- [ ] Hacker News post (optional)

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

*Plan updated: 2025-10-22 (Post pre-launch review)*
*Line count: Target <400 lines (currently ~250)*
