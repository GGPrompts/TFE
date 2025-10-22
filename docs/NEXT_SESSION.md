# Next Session: Pre-Launch Testing & Polish

**Goal:** Test critical security fixes and complete high-priority polish before v1.0 release.

**Status:** ✅ CRITICAL FIXES COMPLETE - Ready for testing & screenshots (2-3 hours to launch)

**Last Updated:** 2025-10-22

---

## ✅ COMPLETED (Session 2025-10-22)

### Security Fixes - ALL COMPLETE

1. **Command Injection in `command.go`** ✅
   - **What:** Implemented command allowlist (30+ safe commands: ls, cat, grep, git, etc.)
   - **File:** `command.go:35-129`
   - **Note:** `!` prefix allows unrestricted access (power-user feature documented)
   - **Test:** Try `:ls; echo INJECTED` → Should show "command not allowed" error

2. **Path Traversal in `loadFiles()`** ✅
   - **What:** Added path validation restricting navigation to home/working directories
   - **File:** `file_operations.go:870-909`
   - **Note:** Blocks system directories (/etc, /root, /boot, /sys, /proc)
   - **Test:** Navigate to `../../../../../../etc` → Should show "access denied"

3. **Filename Injection in `openEditor()`** ✅
   - **What:** Validates filenames, blocks those starting with `-`
   - **File:** `editor.go:30-59`
   - **Test:** `touch -- --dangerous-flag.txt; tfe` then F4 → Should error

4. **Cross-Device Trash Move** ✅
   - **What:** Detects EXDEV errors, falls back to copy+delete
   - **Files:** `trash.go:137-162` (main), `trash.go:300-362` (helpers)
   - **Test:** `cp /etc/hosts /tmp/test.txt; tfe /tmp` then delete → Should work

### Documentation - COMPLETE

5. **FAQ.md Created** ✅
   - 38 Q&A entries covering all common user questions
   - Sections: Installation, Terminal Compatibility, Features, Performance, Termux, Commands, Prompts, Navigation, Troubleshooting, Advanced

6. **CONTRIBUTING.md Enhanced** ✅
   - Added decision tree for where to add code
   - Added security considerations section
   - Added common pitfalls (header duplication, mouse coords, file handles)
   - Updated documentation line limits

### Verified

7. **Build Passes** ✅
   - `go build` completes without errors
   - All imports properly added (syscall, errors, io, filepath)

---

## 📋 NEXT SESSION TASKS (2-3 hours)

### Priority 1: Testing (30 minutes)

Run these security tests to verify fixes:

```bash
# 1. Command injection test
#    Launch TFE, press :, type: ls; echo INJECTED
#    Expected: "command not allowed" error with helpful message

# 2. Path traversal test
#    Navigate to ../../../../../../etc
#    Expected: "access denied" or prevented navigation

# 3. Filename injection test
touch -- --dangerous-flag.txt
#    Launch TFE, select file, press F4
#    Expected: "invalid filename: cannot start with '-'" error

# 4. Cross-device trash test
cp /etc/hosts /tmp/test.txt
#    Launch TFE in /tmp, delete test.txt
#    Expected: Should move to trash without EXDEV error

# 5. File handle leak test (optional - requires monitoring)
for i in {1..2000}; do echo "test" > /tmp/test$i.txt; done
#    Preview all files rapidly in TFE
#    Check: lsof | grep tfe | wc -l
#    Expected: File handles should be properly closed
```

### Priority 2: Screenshots for README.md (30 minutes)

Take 3-5 screenshots to add to README.md:

1. **Tree view with dual-pane preview** - Show the main interface
2. **Detail view with file metadata** - Show rich file information
3. **Prompt template with fillable fields** - Demonstrate prompt feature
4. **Context menu (right-click)** - Show available operations
5. **Termux on Android** (optional) - Show mobile usage

Save to `examples/` directory and reference in README.md.

### Priority 3: High-Priority Polish (Optional - 1.5 hours)

From PLAN.md, consider these quick wins before v1.0:

**Quick Fixes (<30 min each):**

1. **Empty Directory Message** [15 min]
   - Issue: Blank screen when entering empty directory
   - Fix: Show "📂 Empty directory - Press ← to go back"

2. **Auto-Switch Dual-Pane View** [20 min]
   - Issue: Error message when entering dual-pane from Detail view
   - Fix: Auto-switch to List view when entering dual-pane
   - File: `update_keyboard.go:1381, 1400`

3. **File Permissions for History/Favorites** [30 min]
   - Issue: Stored with 0644 (world-readable) instead of 0600
   - Fix: Change WriteFile permissions to 0600
   - Files: `command.go:249`, `favorites.go` (save functions)

**Larger Improvements (can defer to v1.1):**
- Glamour markdown timeout (2 seconds instead of infinite)
- Detail view dynamic width for narrow terminals
- Complete HOTKEYS.md documentation
- Unbounded data structures (LRU cache)

---

## 🚀 LAUNCH CHECKLIST

### Critical (MUST DO)
- [x] All 4 security fixes complete
- [x] FAQ.md created
- [x] CONTRIBUTING.md enhanced
- [x] Code builds successfully
- [ ] Security tests pass (see "Testing" above)
- [ ] README.md screenshots added (3-5 images)

### Recommended (SHOULD DO)
- [ ] Empty directory message fix
- [ ] File permissions fix (0600)
- [ ] Final Termux testing (Android)
- [ ] Final WSL2 testing

### Optional (NICE TO HAVE)
- [ ] Auto-switch dual-pane view
- [ ] Complete HOTKEYS.md
- [ ] Glamour timeout fix

### Release
- [ ] Update CHANGELOG.md with all fixes
- [ ] Tag v1.0 release: `git tag -a v1.0 -m "Initial stable release"`
- [ ] Push tag: `git push origin v1.0`
- [ ] Create GitHub release with notes
- [ ] Optional: Post to Hacker News

---

## 📝 NOTES FROM LAST SESSION

### Issues Not Actually Bugs

During the session, we found that some audit items were already fixed or false positives:

- **File Handle Leak:** All `os.Open()` calls already have `defer file.Close()`
- **Circular Symlinks:** Go's `os.Stat()` naturally detects and errors on circular references; navigation is user-driven (not recursive) so no infinite loop risk

### Audit Accuracy

The pre-launch audit had some inaccuracies:
- Some line numbers were outdated
- File handle leak was already fixed in current code
- Circular symlink issue was not a real vulnerability

**Lesson:** Always verify audit findings against current code before implementing fixes.

### Command Allowlist Design

The command allowlist balances security and usability:
- **Default (`:command`)**: Safe read-only commands only
- **Unrestricted (`!command`)**: Full shell access for power users
- **Error messages**: Helpful, explain how to use `!` prefix

This design respects user choice while protecting against accidental destructive commands.

---

## 🎯 RECOMMENDED NEXT STEPS

**Option A - Quick Launch (30 min):**
1. Run security tests
2. Add 3 README screenshots
3. Update CHANGELOG.md
4. Tag v1.0 and release

**Option B - Polished Launch (2-3 hours):**
1. Run security tests
2. Add 5 README screenshots
3. Fix empty directory message
4. Fix file permissions (0600)
5. Test on Termux + WSL2
6. Update CHANGELOG.md
7. Tag v1.0 and release

**Option C - Perfect Launch (defer polish to v1.1):**
1. Run security tests
2. Add README screenshots
3. Update CHANGELOG.md with security fixes
4. Tag v1.0 and release
5. Create GitHub issues for v1.1 polish items

---

**Time Investment:**
- Critical security fixes: ✅ DONE (2-3 hours)
- Testing + screenshots: 1 hour
- High-priority polish: 1.5 hours (optional)
- **Total to launch:** 1-2.5 hours remaining

**Recommendation:** Run tests and add screenshots (1 hour), then release v1.0. Polish items can go in v1.1.
