# Pre-Launch Security & Performance Fixes

**Goal:** Fix critical issues identified in the comprehensive pre-launch review before v1.0 release.

**Status:** ⚠️ Ready with Caveats - 8 critical issues block safe public launch (6 hours estimated)

**See [PLAN.md](../PLAN.md) for complete issue list with 35 items across all priorities.**

---

## 🔥 CRITICAL FIXES (MUST Complete - 6 hours)

These issues represent security vulnerabilities and performance bugs that could cause system compromise or crashes. Fix in priority order:

### 1. File Handle Leak [5 minutes] ⚡ ONE LINE FIX

**File:** `file_operations.go:298-299`

**Issue:** Missing `defer file.Close()` in `loadPreview()` causes file descriptor leak. Application crashes after ~1000 large file previews (ulimit exhaustion).

**Fix:**
```go
// Line 298-299
file, err := os.Open(path)
if err != nil {
    return "", err
}
defer file.Close()  // ← ADD THIS LINE

stat, err := file.Stat()
// ... rest of function
```

**Test:** Preview 2000+ large files rapidly, check `lsof | grep tfe` for leaked handles.

---

### 2. Command Injection in executeCommand() [2 hours]

**File:** `command.go:27-47`

**Issue:** User input passed directly to `/bin/sh -c` without sanitization. Allows arbitrary command execution with `;`, `&&`, `||`, piping.

**Exploit Example:** User types `ls; rm -rf ~` in command prompt.

**Fix Option A - Command Allowlist (Recommended):**
```go
func (m *model) executeCommand(cmdStr string) tea.Cmd {
    parts := strings.Fields(cmdStr)
    if len(parts) == 0 {
        return nil
    }

    // Allowlist of safe commands
    safeCommands := map[string]bool{
        "ls": true, "cat": true, "grep": true, "find": true,
        "head": true, "tail": true, "wc": true, "file": true,
        "git": true, "tree": true, "du": true, "df": true,
    }

    executable := parts[0]
    if !safeCommands[executable] {
        return func() tea.Msg {
            return editorFinishedMsg{
                err: fmt.Errorf("command not allowed: %s (see HOTKEYS.md for safe commands)", executable),
            }
        }
    }

    // Don't use shell, execute directly
    return func() tea.Msg {
        cmd := exec.Command(executable, parts[1:]...)
        cmd.Dir = m.currentPath
        output, err := cmd.CombinedOutput()

        return editorFinishedMsg{
            output: string(output),
            err:    err,
        }
    }
}
```

**Fix Option B - Confirmation Dialog (Alternative):**
```go
// Detect dangerous patterns
dangerousPatterns := []string{"rm -rf", "> /dev/", "dd if=", "mkfs", ":(){ :|:& };:"}
for _, pattern := range dangerousPatterns {
    if strings.Contains(cmdStr, pattern) {
        // Show confirmation dialog before executing
        m.dialogType = "confirm"
        m.dialogTitle = "⚠️  Dangerous Command Detected"
        m.dialogMessage = fmt.Sprintf("Command contains '%s'. Execute anyway?", pattern)
        // ... set up callback to execute if confirmed
        return nil
    }
}
```

**Test:** Try `ls; echo INJECTED`, `cat /etc/passwd`, `rm -rf test/`

---

### 3. Path Traversal in loadFiles() [1 hour]

**File:** `file_operations.go:41-57`

**Issue:** No validation against `../../etc` style directory traversal. Users can navigate to sensitive system directories.

**Fix:**
```go
func (m *model) loadFiles(path string) tea.Cmd {
    return func() tea.Msg {
        // Validate and clean path
        absPath, err := filepath.Abs(path)
        if err != nil {
            return editorFinishedMsg{err: err}
        }

        cleanPath := filepath.Clean(absPath)

        // Optional: Restrict to home directory or working directory tree
        homeDir, _ := os.UserHomeDir()
        if homeDir != "" && !strings.HasPrefix(cleanPath, homeDir) {
            // Check if within original working directory
            wd, _ := os.Getwd()
            if !strings.HasPrefix(cleanPath, wd) {
                return editorFinishedMsg{
                    err: fmt.Errorf("access denied: path outside allowed directories"),
                }
            }
        }

        entries, err := os.ReadDir(cleanPath)
        if err != nil {
            return editorFinishedMsg{err: err}
        }

        // ... rest of function
    }
}
```

**Note:** Consider making path restrictions configurable via flag `--allow-full-access` for power users.

**Test:** Navigate to `../../../../../../etc`, verify prevention.

---

### 4. Command Injection in openEditor() [30 minutes]

**File:** `editor.go:68-91`

**Issue:** Filenames starting with `-` can inject editor arguments (e.g., file named `--dangerous-flag`).

**Fix:**
```go
func openEditor(editor string, path string) tea.Cmd {
    // Validate filename doesn't start with dangerous characters
    cleanPath := filepath.Clean(path)
    filename := filepath.Base(cleanPath)

    if strings.HasPrefix(filename, "-") {
        return func() tea.Msg {
            return editorFinishedMsg{
                err: fmt.Errorf("invalid filename: cannot start with '-'"),
            }
        }
    }

    // Use absolute path to avoid ambiguity
    absPath, err := filepath.Abs(cleanPath)
    if err != nil {
        return func() tea.Msg {
            return editorFinishedMsg{err: err}
        }
    }

    c := exec.Command(editor, absPath)
    return tea.ExecProcess(c, func(err error) tea.Msg {
        return editorFinishedMsg{err: err}
    })
}
```

**Test:** Create file named `--dangerous-flag.txt`, try opening with F4.

---

### 5. Circular Symlink Detection [1 hour]

**File:** `file_operations.go:948-959`

**Issue:** No detection of circular symlinks. Causes infinite loop when navigating `dir1 → dir2 → dir1`.

**Fix:**
```go
// Add to types.go
type model struct {
    // ... existing fields
    visitedPaths map[string]bool // Track visited symlink paths
    symlinkDepth int              // Current symlink follow depth
}

// In file_operations.go
const maxSymlinkDepth = 40 // Linux kernel limit

func (m *model) followSymlink(path string) (string, error) {
    // Reset depth when starting fresh navigation
    if m.symlinkDepth == 0 {
        m.visitedPaths = make(map[string]bool)
    }

    // Check for circular reference
    absPath, _ := filepath.Abs(path)
    if m.visitedPaths[absPath] {
        return "", fmt.Errorf("circular symlink detected: %s", path)
    }

    // Check depth limit
    if m.symlinkDepth >= maxSymlinkDepth {
        return "", fmt.Errorf("symlink depth limit exceeded (max %d)", maxSymlinkDepth)
    }

    m.visitedPaths[absPath] = true
    m.symlinkDepth++

    // Resolve symlink
    target, err := os.Readlink(path)
    if err != nil {
        return "", err
    }

    return target, nil
}
```

**Test:** Create circular symlink `ln -s dir1 dir2/link; ln -s dir2 dir1/link`, navigate into it.

---

### 6. Cross-Device Trash Move [1 hour]

**File:** `trash.go:135-137`

**Issue:** `os.Rename()` fails when moving files across different mount points (e.g., `/tmp` → `~/.trash`). Returns cryptic `syscall.EXDEV` error.

**Fix:**
```go
func moveToTrash(path string) error {
    trashDir := getTrashDir()
    trashFile := filepath.Join(trashDir, "files", filepath.Base(path))

    // Try rename first (fast, atomic)
    err := os.Rename(path, trashFile)
    if err == nil {
        return nil // Success
    }

    // Check if cross-device error
    if errors.Is(err, syscall.EXDEV) {
        // Fallback to copy+delete for cross-device moves
        return copyAndDelete(path, trashFile)
    }

    return err
}

func copyAndDelete(src, dst string) error {
    // Copy file/directory recursively
    if err := copyRecursive(src, dst); err != nil {
        return fmt.Errorf("copy failed: %w", err)
    }

    // Delete original
    if err := os.RemoveAll(src); err != nil {
        return fmt.Errorf("delete failed after copy: %w", err)
    }

    return nil
}

func copyRecursive(src, dst string) error {
    info, err := os.Stat(src)
    if err != nil {
        return err
    }

    if info.IsDir() {
        return copyDir(src, dst, info)
    }
    return copyFile(src, dst, info)
}

// Implement copyDir and copyFile with permission preservation
```

**Test:** Move file from `/tmp` to trash, verify fallback works.

---

### 7. Split CHANGELOG.md → CHANGELOG3.md [COMPLETED ✅]

**Status:** Done in commit `260f369`

---

### 8. Create FAQ.md + CONTRIBUTING.md + Screenshots [3 hours]

#### A. FAQ.md [1.5 hours]

Create `FAQ.md` with common troubleshooting:

```markdown
# TFE Frequently Asked Questions

## Installation & Setup

### Q: TFE won't start, shows "command not found"
**A:** Ensure Go is installed and `~/go/bin` is in your PATH:
\`\`\`bash
echo 'export PATH=$PATH:~/go/bin' >> ~/.bashrc
source ~/.bashrc
\`\`\`

### Q: Permission denied errors when accessing directories
**A:** TFE respects file system permissions. Use `chmod` to grant access or run from directories you own.

## Terminal Compatibility

### Q: Emoji buttons have weird spacing (CellBlocks, ttyd, wetty)
**A:** Web-based terminals using xterm.js have emoji rendering issues. This affects ~5% of users.
**Workaround:** Use native terminals (Termux, WSL, iTerm2, GNOME Terminal).
**Future:** v1.1 will add `--ascii-mode` flag.

### Q: Mouse clicks don't work
**A:** Ensure your terminal supports mouse events. Most modern terminals do. If using tmux, add `set -g mouse on` to `.tmux.conf`.

## Feature Questions

### Q: How do I copy files?
**A:** Right-click → "📋 Copy to..." or use context menu (F2).

### Q: Clipboard copy (F5) doesn't work
**A:** Install clipboard utility:
- Linux: `sudo apt install xclip`
- macOS: Built-in (pbcopy)
- Termux: `pkg install termux-api`

## Performance

### Q: TFE is slow with large directories (10,000+ files)
**A:** Use tree view (press 3) for better performance. Detail view renders all metadata.

### Q: Preview is slow for large markdown files
**A:** Glamour rendering has 5-second timeout. Files >1MB show size warning without preview.

## Termux / Mobile

### Q: Text is too small on phone
**A:** Adjust Termux font size: Long-press → Style → Font size

### Q: Scrolling doesn't work on phone
**A:** Use arrow keys or vim keys (j/k). Mouse wheel scrolling works if Termux touch mode is enabled.
```

#### B. CONTRIBUTING.md [1 hour]

Create `CONTRIBUTING.md` with development setup:

```markdown
# Contributing to TFE

## Development Setup

1. **Install Go 1.21+**
2. **Clone repo:**
   \`\`\`bash
   git clone https://github.com/GGPrompts/TFE
   cd TFE
   \`\`\`
3. **Install dependencies:**
   \`\`\`bash
   go mod download
   \`\`\`
4. **Run locally:**
   \`\`\`bash
   go run .
   \`\`\`

## Code Architecture

See [CLAUDE.md](CLAUDE.md) for complete architecture guide.

**Key Principles:**
- Keep `main.go` minimal (entry point only)
- One responsibility per file (target <500 lines, max 800)
- Follow decision tree for where new code belongs
- Update CLAUDE.md when adding modules

## Making Changes

1. **Create feature branch:** `git checkout -b feature/your-feature`
2. **Follow Go conventions:** Use `gofmt`, add comments
3. **Test manually:** Verify on Termux (Android) and desktop
4. **Update docs:** HOTKEYS.md for new shortcuts, CHANGELOG.md for changes
5. **Submit PR:** Include description, testing notes, screenshots

## Testing

Currently manual testing. Automated tests welcome! See `*_test.go` files for examples.

**Test checklist:**
- [ ] Works on narrow terminals (80 columns)
- [ ] Works on wide terminals (200+ columns)
- [ ] No panics or crashes
- [ ] Keyboard shortcuts don't conflict
- [ ] Mouse interactions accurate

## Questions?

Open an issue or discussion on GitHub!
```

#### C. README.md Screenshots [30 minutes]

Take 3-5 screenshots and add to README.md:

1. **Tree view with dual-pane preview**
2. **Detail view with file metadata**
3. **Prompt template with fillable fields**
4. **Context menu demonstration**
5. **Termux on Android (mobile usage)**

Use `screenshot` utility or take actual photos of Termux on phone.

---

## Testing Before Launch

After fixing all critical issues, run these tests:

```bash
# Security tests
# 1. Try command injection
#    Launch TFE, press :, type: ls; echo INJECTED
#    Expected: Either "command not allowed" or confirmation dialog

# 2. Try path traversal
#    Navigate to ../../../../../../etc
#    Expected: "access denied" or similar

# 3. Test filename injection
#    touch -- --dangerous-flag.txt
#    Press F4 to edit
#    Expected: "invalid filename" error

# Performance tests
# 4. Test file handle leak
#    for i in {1..2000}; do echo "test" > /tmp/test$i.txt; done
#    Preview all files rapidly in TFE
#    Check: lsof | grep tfe | wc -l
#    Expected: Should not grow indefinitely

# Edge case tests
# 5. Circular symlinks
#    mkdir -p test/dir1 test/dir2
#    ln -s ../dir1 test/dir2/link
#    ln -s ../dir2 test/dir1/link
#    Navigate into symlink loop
#    Expected: "circular symlink detected" error

# 6. Cross-device trash
#    cp /etc/hosts /tmp/test.txt
#    Move to trash from /tmp
#    Expected: Should work (fallback to copy+delete)
```

---

## After Critical Fixes Complete

Move to **High Priority** fixes in PLAN.md (8 hours estimated):

1. File permissions for history/favorites (0600 instead of 0644)
2. Unbounded data structures (LRU cache for expandedDirs)
3. Glamour markdown timeout (2 seconds)
4. Detail view width on narrow terminals (dynamic columns)
5. Header scrolling misalignment in detail view
6. Extract header rendering function (DRY fix)
7. Auto-switch to List view for dual-pane
8. Empty directory message
9. Complete HOTKEYS.md (Ctrl+F, Ctrl+P, mouse guide)

---

## Launch Checklist

- [ ] All 6 critical code fixes complete
- [ ] FAQ.md created
- [ ] CONTRIBUTING.md created
- [ ] README.md screenshots added
- [ ] All tests pass
- [ ] Final Termux testing (Android)
- [ ] Final WSL2 testing
- [ ] Tag v1.0 release
- [ ] GitHub release notes
- [ ] Optional: Hacker News post

---

**Estimated Total Time:** 6 hours (critical fixes) + 3 hours (docs) = **9 hours to launch-ready**

**Current Status:** Documentation complete (PLAN.md, CHANGELOG splits, CLAUDE.md optimized)

**Next Step:** Start with the 5-minute file handle leak fix in `file_operations.go:299` ⚡
