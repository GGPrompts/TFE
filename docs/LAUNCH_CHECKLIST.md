# TFE Public Launch Checklist

**Target:** v1.0 Public Release
**Created:** 2025-10-18
**Status:** Pre-Launch Preparation 🚧

---

## Critical Features for v1.0

### ✅ Already Complete

- [x] **Core Navigation** - All 4 view modes (List/Grid/Detail/Tree)
- [x] **File Preview** - Syntax highlighting, markdown rendering
- [x] **External Editor** - F4 integration (micro/nano/vim)
- [x] **Create Directory** - F7 with dialog
- [x] **Delete Files** - F8 with confirmation
- [x] **Context Menu** - Right-click + F2
- [x] **Favorites System** - Bookmarks with F6 filter
- [x] **Fuzzy Search** - Ctrl+P file finding
- [x] **Directory Search** - `/` key filtering
- [x] **Prompts Library** - F11 mode with multi-format support
- [x] **Fillable Fields** - Interactive prompt variables with F3 file picker ✨
- [x] **Command Prompt** - MC-style shell integration
- [x] **Clipboard** - Multi-platform path copying
- [x] **Mouse Support** - Click, double-click, scroll, column sorting

### 🔴 Critical Missing (Blockers for v1.0)

#### 1. Copy Files ⚠️ HIGH PRIORITY
**Status:** Not implemented
**Why critical:** Most common file operation after navigation
**Implementation:**
- [ ] Add context menu item: "📋 Copy to..."
- [ ] Input dialog for destination path
- [ ] Progress indicator for large files
- [ ] Error handling (permissions, disk space)
- [ ] F5 conflicts with clipboard - use **Shift+F5** or context menu only

**Estimated Time:** 2-3 hours
**Files:** New `file_copy.go` module, `context_menu.go`, `dialog.go`

#### 2. Rename Files ⚠️ HIGH PRIORITY
**Status:** Not implemented
**Why critical:** Extremely common operation, no terminal workaround is convenient
**Implementation:**
- [ ] Add context menu item: "✏️ Rename..."
- [ ] Input dialog pre-filled with current name
- [ ] Validation (no path separators, check exists)
- [ ] Error handling (permissions, conflicts)
- [ ] Support for directories too

**Estimated Time:** 1-2 hours
**Files:** `context_menu.go`, `dialog.go`, `file_operations.go`

#### 3. New File Creation ⚠️ HIGH PRIORITY
**Status:** Not implemented
**Why critical:** `touch file && micro file` is too clunky
**Implementation:**
- [ ] Add context menu item: "📄 New File..."
- [ ] Input dialog for filename
- [ ] Auto-open in external editor after creation
- [ ] Error handling (permissions, exists)

**Estimated Time:** 1 hour
**Files:** `context_menu.go`, `dialog.go`, `file_operations.go`

**Total Estimated Time for Critical Features:** 4-6 hours

---

## Documentation & Marketing

### 🔴 Required for Launch

#### 4. Screenshots & GIFs ⚠️ HIGH PRIORITY
**Status:** Not created
**Why critical:** GitHub visitors need to see the UI immediately
**Required Screenshots:**
- [ ] Single-pane file browser (detail view)
- [ ] Dual-pane mode with preview
- [ ] Tree view with expansion
- [ ] Context menu in action
- [ ] Prompts mode with fillable fields ✨ (UNIQUE FEATURE!)
- [ ] Full-screen preview with syntax highlighting
- [ ] **Termux/mobile screenshot** showing touch controls 📱 (UNIQUE!)

**Required GIFs (10-15 seconds each):**
- [ ] Navigation and file operations (create, delete, rename)
- [ ] Prompts workflow: F11 → select → fill fields → F3 file picker → F5 copy
- [ ] Dual-pane preview in action
- [ ] Fuzzy search (Ctrl+P)
- [ ] **Mobile touch navigation** (tap, long-press context menu) 📱

**Tools:** asciinema + agg (for GIFs), or  termshot
**Estimated Time:** 2 hours
**Update:** README.md with embedded screenshots

#### 5. Installation Documentation
**Status:** Basic instructions exist, needs expansion
**Required:**
- [ ] Platform-specific instructions (Linux, macOS, WSL, native Windows)
- [ ] Dependency documentation (Go 1.24+, Nerd Fonts)
- [ ] Optional dependencies (micro, xclip, wl-clipboard, termux-api)
- [ ] Quick CD wrapper setup (step-by-step)
- [ ] Troubleshooting section

**Estimated Time:** 1 hour
**Update:** README.md

#### 6. Feature Comparison Table
**Status:** Not created
**Why needed:** Show what makes TFE different
**Compare against:**
- ranger (Python)
- nnn (C, minimalist)
- lf (Go, similar)
- yazi (Rust, modern)
- Midnight Commander (classic)

**TFE Unique Advantages:**
- ✨ **Prompts library with fillable fields** (NO OTHER TOOL HAS THIS!)
- 📱 **Mobile/Termux support with full touch controls** (tested throughout development)
- Built for AI workflows (Claude Code integration)
- Modern Go + Bubbletea stack
- F-key shortcuts (MC-compatible)
- Dual-pane preview
- Context menu with TUI tool launcher

**Estimated Time:** 30 minutes
**Add to:** README.md

---

## Technical Polish

### 🟡 Nice to Have (Can ship without)

#### 7. Symlink Indicators
**Status:** Not implemented
**Why useful:** Users can see broken/circular symlinks
**Implementation:**
- [ ] Detect symlinks in `loadFiles()`
- [ ] Different icon (🔗) or color
- [ ] Show target in detail view/status bar

**Estimated Time:** 1 hour
**Priority:** Medium (can ship without)

#### 8. File Permissions Indicator
**Status:** Partial (shown in status bar)
**Enhancement:**
- [ ] Show executable flag in detail view
- [ ] Visual indicator for read-only files
- [ ] Color coding for permission levels

**Estimated Time:** 1 hour
**Priority:** Low (nice to have)

#### 9. Binary File Detection Improvements
**Status:** Basic detection exists
**Enhancement:**
- [ ] Better heuristics (check for NULL bytes)
- [ ] Detect more file types (images, videos, PDFs)
- [ ] Offer to open in system app

**Estimated Time:** 1 hour
**Priority:** Low

---

## Release Process

### 10. GitHub Release Preparation

**Required:**
- [ ] Create GitHub release (v1.0.0)
- [ ] Write release notes (features, usage, installation)
- [ ] Build binaries for multiple platforms:
  - [ ] Linux (amd64, arm64)
  - [ ] macOS (amd64, arm64)
  - [ ] Windows (amd64) - native if possible
- [ ] Attach binaries to release
- [ ] Create installation script (curl | sh pattern)

**Estimated Time:** 2-3 hours
**Tools:** GoReleaser (automates multi-platform builds)

### 11. Marketing & Launch

**Launch Targets:**
- [ ] Post to r/golang (Show HN style)
- [ ] Post to r/commandline
- [ ] Post to r/unixporn (with nice screenshots!)
- [ ] Hacker News (Show HN: TFE - Terminal File Explorer with AI Prompts)
- [ ] lobste.rs
- [ ] Tweet/Mastodon announcement
- [ ] Add to awesome-tui lists

**Marketing Angle:**
> "TFE isn't just another file manager - it's the first terminal file explorer built for the AI era. Browse files like Midnight Commander, manage prompts like a library, fill variables interactively, and copy rendered results with one keystroke. Works beautifully on desktop AND mobile (Termux) with full touch controls. Perfect for developers working with Claude Code and AI tools anywhere."

**Lead with:**
1. Fillable fields + prompts library (unique!)
2. Mobile/Termux support (rare in TUI file managers!)

**Estimated Time:** 1 hour (writing posts)

---

## Quality Assurance

### 12. Testing Checklist

**Pre-launch testing:**
- [ ] Test on clean Linux install (no config)
- [ ] Test on macOS
- [ ] Test on WSL
- [ ] Test with small terminal (80x24)
- [ ] Test with large terminal (200x60)
- [ ] Test all F-keys (F1-F11)
- [ ] Test all context menu actions
- [ ] Test prompts with all variable types
- [ ] Test file picker in prompts mode
- [ ] Test clipboard on all platforms
- [ ] Test with no editor installed
- [ ] Test with no clipboard tool
- [ ] Verify no crashes on permission denied
- [ ] Verify no crashes on disk full

**Estimated Time:** 2 hours

---

## Timeline to Launch

### Minimum Viable v1.0 (4-6 hours of work)
1. ✅ **Copy files** (2-3 hours)
2. ✅ **Rename files** (1-2 hours)
3. ✅ **New file creation** (1 hour)

### Launch Ready (8-12 hours total)
4. ✅ **Screenshots/GIFs** (2 hours)
5. ✅ **Documentation polish** (1.5 hours)
6. ✅ **GitHub release + binaries** (2-3 hours)
7. ✅ **Testing** (2 hours)
8. ✅ **Marketing posts** (1 hour)

### Optional Polish (3-4 hours)
9. 🟡 **Symlink indicators** (1 hour)
10. 🟡 **Permission indicators** (1 hour)
11. 🟡 **Binary detection** (1 hour)

---

## Success Criteria

### v1.0 is ready to ship when:
- [x] All critical file operations work (create dir ✅, delete ✅, copy ❌, rename ❌, new file ❌)
- [ ] Complete README with screenshots
- [ ] Binaries available for Linux/macOS
- [ ] No known crashes or data loss bugs
- [ ] Clipboard works on target platforms
- [ ] Prompts feature fully documented
- [ ] Fillable fields workflow tested end-to-end

### v1.0 will be successful if:
- 100+ GitHub stars in first week
- Featured on Hacker News front page
- Mentioned in "awesome-tui" lists
- Users create their own prompt libraries
- Positive feedback on AI workflow integration

---

## Post-Launch Roadmap (v1.1+)

### High Priority (v1.1)
- Multi-select operations (Space to mark, then bulk copy/delete)
- Archive operations (extract .zip, .tar.gz)
- File permissions editor (GUI for chmod)

### Medium Priority (v1.2)
- Themes/color schemes
- Config file (~/.config/tfe/config.yaml)
- Disk usage analyzer (like ncdu)

### Future (v2.0+)
- Context Visualizer (show Claude Code context, token counts) - **TFE's next killer feature!**
- Plugin system
- Multi-pane mode (3+ panes)
- Tabs support

---

## Notes

**Philosophy for v1.0:**
- Ship a complete, polished file manager
- Highlight the unique prompts feature
- Don't over-engineer
- Get feedback early
- Iterate based on user requests

**Differentiation:**
- TFE's prompts library with fillable fields is genuinely innovative
- No other terminal file manager has this
- Perfect for AI-assisted development workflows
- Lead with this in marketing

**Quality Bar:**
- No data loss bugs
- Graceful error handling
- Clear user feedback
- Works across platforms
- Fast and responsive

---

**Last Updated:** 2025-10-18
**Next Session:** Implement copy/rename/new file operations
**Target Launch Date:** Within 1 week (after 8-12 hours of work)
