# TFE Security Testing Plan
## Pre-Launch Manual Security Verification

This document outlines manual security testing procedures to verify TFE's security before public launch.

---

## Overview

While automated subagent reviews flagged several security concerns, most were **false positives** due to incomplete code analysis. This manual testing plan verifies that security measures are working correctly.

### Already Verified (Code Review)
✅ Command injection prevention (shellQuote() function in command.go)
✅ Nil pointer checks (getCurrentFile() usage)
✅ Repository URLs correct (github.com/GGPrompts/tfe)
✅ F1 help implemented with context awareness

---

## Test Categories

### 1. Command Execution Security
### 2. File System Operations
### 3. Input Sanitization
### 4. External Tool Integration
### 5. File Preview Safety
### 6. Path Traversal Prevention

---

## 1. Command Execution Security

**Goal:** Verify command prompt sanitizes input and prevents injection attacks.

### Test 1.1: Basic Command Injection Attempts

**Setup:**
1. Launch TFE: `./tfe`
2. Press `:` to activate command prompt

**Test Cases:**

| Input Command | Expected Result | Pass/Fail |
|---------------|-----------------|-----------|
| `ls` | Execute ls normally | ⬜ |
| `ls; whoami` | Execute only "ls; whoami" as literal (semicolon quoted) | ⬜ |
| `echo test` | Echo "test" | ⬜ |
| `echo $(whoami)` | Echo literal string "$(whoami)", NOT execute whoami | ⬜ |
| `ls \| grep test` | Execute "ls | grep test" safely (pipe quoted) | ⬜ |
| `rm -rf /tmp/test && echo done` | Execute safely (both commands as single quoted string) | ⬜ |
| `ls 'test file'` | Handle quoted filenames correctly | ⬜ |

**Verification:**
- Check that special characters (`;`, `|`, `$()`, backticks) are properly quoted
- Verify no unintended command execution
- Review command.go:158-162 (shellQuote function) during testing

**Implementation Check:**
```bash
# Verify shellQuote() escapes single quotes
grep -A5 "func shellQuote" command.go
```

---

### Test 1.2: File Path Command Injection

**Setup:**
1. Create test files with malicious names:
```bash
mkdir /tmp/tfe_security_test
cd /tmp/tfe_security_test
touch "normal.txt"
touch "file;rm -rf ~.txt"  # semicolon in filename
touch "test\$(whoami).txt"  # command substitution in filename
```

**Test Cases:**

| Action | Expected Result | Pass/Fail |
|--------|-----------------|-----------|
| Navigate to /tmp/tfe_security_test | Display files with special chars | ⬜ |
| Select "file;rm -rf ~.txt", press Enter for preview | Preview file safely, no command execution | ⬜ |
| Select malicious file, press F4 (editor) | Open in editor with path properly quoted | ⬜ |
| Right-click malicious file, copy path | Copy full path safely | ⬜ |

**Verification:**
- No commands should execute from filenames
- Paths should be properly escaped in all contexts
- Home directory (~) should remain intact after all operations

**Cleanup:**
```bash
rm -rf /tmp/tfe_security_test
```

---

## 2. File System Operations

**Goal:** Verify file operations handle edge cases safely.

### Test 2.1: Path Traversal Prevention

**Setup:**
```bash
mkdir -p /tmp/tfe_path_test/{safe,restricted}
echo "Safe file" > /tmp/tfe_path_test/safe/file.txt
echo "Restricted" > /tmp/tfe_path_test/restricted/secret.txt
chmod 700 /tmp/tfe_path_test/restricted
```

**Test Cases:**

| Action | Expected Result | Pass/Fail |
|--------|-----------------|-----------|
| Navigate to /tmp/tfe_path_test | Display both directories | ⬜ |
| Try to navigate to restricted/ | Show permission error gracefully | ⬜ |
| Navigate to safe/, then type `cd ../restricted` in command | Permission denied or safe handling | ⬜ |
| Try to preview /etc/passwd via any mechanism | Only works if user has permission | ⬜ |

**Verification:**
- TFE respects file system permissions
- No unauthorized access to restricted directories
- Error messages are clear but don't leak sensitive info

**Cleanup:**
```bash
chmod 755 /tmp/tfe_path_test/restricted
rm -rf /tmp/tfe_path_test
```

---

### Test 2.2: Symbolic Link Handling

**Setup:**
```bash
mkdir -p /tmp/tfe_symlink_test
cd /tmp/tfe_symlink_test
echo "Target file" > target.txt
ln -s target.txt link.txt
ln -s /etc/hosts hosts_link
ln -s nonexistent.txt broken_link
ln -s ../.. escape_attempt
```

**Test Cases:**

| Action | Expected Result | Pass/Fail |
|--------|-----------------|-----------|
| Navigate to /tmp/tfe_symlink_test | Display all links | ⬜ |
| Preview link.txt | Show "Target file" content | ⬜ |
| Preview hosts_link | Show /etc/hosts if readable | ⬜ |
| Preview broken_link | Show error gracefully (file not found) | ⬜ |
| Navigate into escape_attempt | Follow link safely (no crash) | ⬜ |

**Verification:**
- Symlinks are followed but handled safely
- Broken symlinks don't crash TFE
- No infinite loops with circular symlinks

**Cleanup:**
```bash
rm -rf /tmp/tfe_symlink_test
```

---

### Test 2.3: Large File Handling

**Setup:**
```bash
mkdir /tmp/tfe_large_test
dd if=/dev/zero of=/tmp/tfe_large_test/5mb.bin bs=1M count=5
dd if=/dev/zero of=/tmp/tfe_large_test/15mb.bin bs=1M count=15
dd if=/dev/urandom of=/tmp/tfe_large_test/100mb.bin bs=1M count=100
```

**Test Cases:**

| File | Action | Expected Result | Pass/Fail |
|------|--------|-----------------|-----------|
| 5mb.bin | Preview (F4 or Enter) | Preview with size limit message or first portion | ⬜ |
| 15mb.bin | Preview | Show "too large" message (>10MB limit) | ⬜ |
| 100mb.bin | Preview | Show "too large" message, no memory spike | ⬜ |
| 5mb.bin | Open in editor (F4) | Editor opens successfully | ⬜ |

**Verification:**
- Monitor memory usage during preview (use `htop` or `top`)
- TFE should not load >10MB files into memory
- No crashes or freezes
- Review file_operations.go:136-161 for preview size limits

**Memory Monitoring:**
```bash
# In another terminal while testing:
watch -n 1 'ps aux | grep "[t]fe"'
```

**Cleanup:**
```bash
rm -rf /tmp/tfe_large_test
```

---

## 3. Input Sanitization

**Goal:** Verify special characters in inputs don't break UI or cause issues.

### Test 3.1: Special Characters in Filenames

**Setup:**
```bash
mkdir /tmp/tfe_special_chars
cd /tmp/tfe_special_chars
touch "normal.txt"
touch "spaces in name.txt"
touch "tab	char.txt"
touch "quote'.txt"
touch 'double"quote.txt'
touch "newline
in-name.txt" 2>/dev/null || echo "Skipped newline test"
touch "emoji_😀_file.txt"
touch "unicode_Ω_file.txt"
```

**Test Cases:**

| File | Action | Expected Result | Pass/Fail |
|------|--------|-----------------|-----------|
| All files | List in TFE | Display without UI corruption | ⬜ |
| spaces in name.txt | Preview | Works correctly | ⬜ |
| spaces in name.txt | Copy path | Path properly quoted | ⬜ |
| quote'.txt | All operations | Handle single quote safely | ⬜ |
| double"quote.txt | All operations | Handle double quote safely | ⬜ |
| emoji_😀_file.txt | Display & preview | Emoji renders or shows placeholder | ⬜ |

**Verification:**
- No UI corruption or broken layouts
- All operations work with special characters
- Paths are properly escaped in all contexts

**Cleanup:**
```bash
rm -rf /tmp/tfe_special_chars
```

---

### Test 3.2: Command Prompt Input Edge Cases

**Setup:**
1. Launch TFE
2. Press `:` for command prompt

**Test Cases:**

| Input | Expected Result | Pass/Fail |
|-------|-----------------|-----------|
| Very long command (500+ chars) | Handle gracefully, no buffer overflow | ⬜ |
| Unicode characters: `echo Ω` | Process correctly | ⬜ |
| Emoji in command: `echo 😀` | Process correctly or show error | ⬜ |
| Empty input (just Enter) | Do nothing gracefully | ⬜ |
| Spaces only: `     ` | Trim and ignore | ⬜ |
| Up arrow (history) on empty history | No crash | ⬜ |

**Verification:**
- No crashes or UI corruption
- Input validation works correctly
- History navigation is safe

---

## 4. External Tool Integration

**Goal:** Verify editor, browser, and clipboard integrations are secure.

### Test 4.1: Editor Opening

**Setup:**
```bash
mkdir /tmp/tfe_editor_test
cd /tmp/tfe_editor_test
echo "Normal file" > normal.txt
touch "file with spaces.txt"
echo "Content" > "file with spaces.txt"
touch "special'chars.txt"
echo "Content" > "special'chars.txt"
```

**Test Cases:**

| File | Action | Expected Result | Pass/Fail |
|------|--------|-----------------|-----------|
| normal.txt | Press F4 (editor) | Opens in micro/nano/vim | ⬜ |
| file with spaces.txt | Press F4 | Opens correctly (path quoted) | ⬜ |
| special'chars.txt | Press F4 | Opens correctly (quote escaped) | ⬜ |

**Verification:**
- Editor opens with correct file in all cases
- No command injection via filename
- Review editor.go:55-78 for path escaping

**Cleanup:**
```bash
rm -rf /tmp/tfe_editor_test
```

---

### Test 4.2: Browser Opening (WSL/Linux)

**Setup:**
```bash
mkdir /tmp/tfe_browser_test
cd /tmp/tfe_browser_test
echo "<html><body>Test</body></html>" > test.html
echo "<html><body>Spaces</body></html>" > "test file.html"
```

**Test Cases:**

| File | Action | Expected Result | Pass/Fail |
|------|--------|-----------------|-----------|
| test.html | Context menu → Open in Browser | Opens in default browser | ⬜ |
| test file.html | Context menu → Open in Browser | Opens correctly (path quoted) | ⬜ |

**Verification:**
- Browser opens correct file
- WSL uses proper `cmd.exe /c start ""` syntax
- Review editor.go:97-103 for WSL fix

**Cleanup:**
```bash
rm -rf /tmp/tfe_browser_test
```

---

### Test 4.3: Clipboard Integration

**Setup:**
```bash
mkdir /tmp/tfe_clipboard_test
cd /tmp/tfe_clipboard_test
touch "normal.txt"
touch "file with spaces.txt"
touch "special'quote.txt"
```

**Test Cases:**

| File | Action | Expected Result | Pass/Fail |
|------|--------|-----------------|-----------|
| normal.txt | Context menu → Copy Path | Path copied to clipboard | ⬜ |
| file with spaces.txt | Copy Path | Full path with spaces copied | ⬜ |
| special'quote.txt | Copy Path | Path with quote copied safely | ⬜ |

**Verification:**
- Paste clipboard content in another terminal
- Verify paths are correct and usable
- No terminal escape sequences in clipboard
- Review editor.go:142-195 for clipboard handling

**Cleanup:**
```bash
rm -rf /tmp/tfe_clipboard_test
```

---

## 5. File Preview Safety

**Goal:** Verify preview doesn't execute code or malicious content.

### Test 5.1: Binary File Detection

**Setup:**
```bash
mkdir /tmp/tfe_binary_test
echo "Plain text file" > /tmp/tfe_binary_test/text.txt
cp /bin/ls /tmp/tfe_binary_test/binary_file
dd if=/dev/urandom of=/tmp/tfe_binary_test/random.dat bs=1K count=10
```

**Test Cases:**

| File | Action | Expected Result | Pass/Fail |
|------|--------|-----------------|-----------|
| text.txt | Preview | Show text content | ⬜ |
| binary_file | Preview | Show "Binary file" warning | ⬜ |
| random.dat | Preview | Detect as binary, show warning | ⬜ |

**Verification:**
- Binary files don't render garbage in preview
- Binary detection is accurate
- Review file_operations.go:208-221 for binary detection

**Cleanup:**
```bash
rm -rf /tmp/tfe_binary_test
```

---

### Test 5.2: Terminal Escape Sequences

**Setup:**
```bash
mkdir /tmp/tfe_escape_test
cd /tmp/tfe_escape_test

# Create file with ANSI escape codes
echo -e "\033[31mRed text\033[0m" > ansi.txt

# Create file with terminal title escape
echo -e "\033]0;Malicious Title\007" > title_escape.txt

# Create file with cursor movement
echo -e "Normal text\033[2J\033[H" > cursor_tricks.txt
```

**Test Cases:**

| File | Action | Expected Result | Pass/Fail |
|------|--------|-----------------|-----------|
| ansi.txt | Preview | Shows escape codes as text OR renders safely | ⬜ |
| title_escape.txt | Preview | No terminal title change | ⬜ |
| cursor_tricks.txt | Preview | No screen clearing or cursor movement | ⬜ |

**Verification:**
- Terminal state remains stable
- No unintended terminal control
- Preview shows content safely

**Cleanup:**
```bash
rm -rf /tmp/tfe_escape_test
```

---

## 6. Path Traversal Prevention

**Goal:** Verify TFE can't be tricked into accessing unauthorized paths.

### Test 6.1: Relative Path Navigation

**Setup:**
```bash
mkdir -p /tmp/tfe_traversal/subdir
echo "Secret" > /tmp/tfe_traversal/secret.txt
echo "Public" > /tmp/tfe_traversal/subdir/public.txt
```

**Test Cases:**

| Starting Point | Action | Expected Result | Pass/Fail |
|----------------|--------|-----------------|-----------|
| /tmp/tfe_traversal/subdir | Navigate to "../" via .. entry | Go to parent (/tmp/tfe_traversal) | ⬜ |
| /tmp/tfe_traversal | Try command: `cd ../../etc` | Navigate to /etc (if allowed) or error | ⬜ |
| Any directory | Try to type `../../../etc/passwd` in command | Handled safely by shell, not TFE bug | ⬜ |

**Verification:**
- Navigation is predictable and safe
- Parent (..) directory always works
- No crashes or unexpected behavior
- TFE respects filesystem permissions

**Cleanup:**
```bash
rm -rf /tmp/tfe_traversal
```

---

## Performance & Stress Testing

### Test 7.1: Large Directory Handling

**Setup:**
```bash
mkdir /tmp/tfe_large_dir
cd /tmp/tfe_large_dir
for i in {1..1000}; do touch "file_$i.txt"; done
```

**Test Cases:**

| Action | Expected Result | Pass/Fail |
|--------|-----------------|-----------|
| Navigate to /tmp/tfe_large_dir | Load and display all 1000 files | ⬜ |
| Scroll through list | Smooth scrolling, no lag | ⬜ |
| Preview files | Quick preview loading | ⬜ |
| Switch display modes | No crashes, reasonable performance | ⬜ |

**Performance Check:**
- Monitor CPU/memory while testing
- Note any slowdowns or freezes
- Check time to load directory: `time ls -la /tmp/tfe_large_dir`

**Cleanup:**
```bash
rm -rf /tmp/tfe_large_dir
```

---

### Test 7.2: Tree View Depth

**Setup:**
```bash
mkdir -p /tmp/tfe_deep/{a/b/c/d/e/f/g/h/i/j,test1,test2}
touch /tmp/tfe_deep/a/b/c/d/e/f/g/h/i/j/deep_file.txt
for i in {1..50}; do mkdir /tmp/tfe_deep/dir_$i; done
```

**Test Cases:**

| Action | Expected Result | Pass/Fail |
|--------|-----------------|-----------|
| Navigate to /tmp/tfe_deep in tree mode | Display directory structure | ⬜ |
| Expand a/b/c/d/e/f/g/h/i/j | Show deeply nested structure | ⬜ |
| Collapse and re-expand | State maintained correctly | ⬜ |
| Navigate with 50+ directories | No performance issues | ⬜ |

**Verification:**
- Tree view handles depth gracefully
- No infinite loops or crashes
- Expansion state is logical

**Cleanup:**
```bash
rm -rf /tmp/tfe_deep
```

---

## Edge Cases & Error Handling

### Test 8.1: Empty Directories

**Setup:**
```bash
mkdir /tmp/tfe_empty_test
```

**Test Cases:**

| Action | Expected Result | Pass/Fail |
|--------|-----------------|-----------|
| Navigate to /tmp/tfe_empty_test | Show empty directory message | ⬜ |
| Press Up/Down | No crashes | ⬜ |
| Press Enter | No crash, no preview | ⬜ |
| All display modes (F1/F2/F3) | Work correctly when empty | ⬜ |

**Cleanup:**
```bash
rmdir /tmp/tfe_empty_test
```

---

### Test 8.2: Terminal Resize During Operations

**Test Cases:**

| During Action | Resize Terminal | Expected Result | Pass/Fail |
|---------------|-----------------|-----------------|-----------|
| File list view | Shrink to 80x24 | UI adjusts correctly | ⬜ |
| File list view | Expand to 200x60 | UI adjusts correctly | ⬜ |
| Preview mode | Resize | Preview reflows properly | ⬜ |
| Dual-pane mode | Resize | Panes adjust proportionally | ⬜ |
| Very small (40x10) | Any action | Graceful degradation or warning | ⬜ |

**Verification:**
- No crashes or visual corruption
- Layout recalculates correctly
- Cursor remains in valid position

---

## Test Results Summary

After completing all tests, fill out this summary:

### Overall Results

| Category | Tests Passed | Tests Failed | Notes |
|----------|--------------|--------------|-------|
| Command Execution | __/__ | __/__ | |
| File System Ops | __/__ | __/__ | |
| Input Sanitization | __/__ | __/__ | |
| External Tools | __/__ | __/__ | |
| Preview Safety | __/__ | __/__ | |
| Path Traversal | __/__ | __/__ | |
| Performance | __/__ | __/__ | |
| Edge Cases | __/__ | __/__ | |

### Critical Issues Found

(List any critical security issues discovered during testing)

1. _______________________________________________
2. _______________________________________________
3. _______________________________________________

### Recommendations

(List recommended fixes or improvements)

1. _______________________________________________
2. _______________________________________________
3. _______________________________________________

---

## Sign-Off

**Tester:** ____________________
**Date:** ______________________
**TFE Version Tested:** v1.0.0
**Platform:** ___________________
**Terminal:** ___________________

**Ready for Launch?** ☐ Yes ☐ No ☐ With Fixes

**Notes:**
_______________________________________________________
_______________________________________________________
_______________________________________________________

---

## Quick Reference: Key Security Features

✅ **Command execution** - Uses shellQuote() to escape arguments (command.go:158)
✅ **Nil checks** - getCurrentFile() usage verified (update_keyboard.go)
✅ **File size limits** - Check file_operations.go:136-161
✅ **Binary detection** - isBinaryFile() in file_operations.go:208
✅ **Path handling** - Verify filepath.Clean() usage throughout
✅ **External tool escaping** - editor.go properly quotes paths

For questions, see SECURITY.md or contact security@ggprompts.com
