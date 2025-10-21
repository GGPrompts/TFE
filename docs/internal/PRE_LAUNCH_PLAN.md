# TFE Pre-Launch Plan

**Generated:** 2025-10-20
**Current Status:** 8.7% test coverage, most critical fixes complete
**Target:** Ready for public release with 40%+ test coverage

---

## Executive Summary

Your TFE project is **almost ready** for public release! The critical security vulnerabilities have been fixed, and the README is comprehensive. However, we need to:

1. ✅ **Fixed (just now):** Failing test in `helpers_test.go`
2. 🔴 **CRITICAL:** Increase test coverage from **8.7% to 40%**
3. 🟡 **HIGH:** Create CONTRIBUTING.md
4. 🟢 **MEDIUM:** Address remaining security items
5. 🟢 **LOW:** Final polish and validation

**Estimated time to launch-ready:** 2-3 days of focused work

---

## ✅ What's Already Done (Great Job!)

### Security Fixes ✅
- ✅ Command injection vulnerabilities fixed
- ✅ Goroutine leaks resolved
- ✅ Unbounded memory consumption fixed
- ✅ Trash/Recycle bin implemented (safe deletion)

### Documentation ✅
- ✅ README.md is comprehensive (375 lines, excellent!)
- ✅ Installation instructions (both quick and full)
- ✅ Usage guide with keyboard/mouse controls
- ✅ Prompts library documentation
- ✅ Mobile/Termux support documented
- ✅ HOTKEYS.md complete
- ✅ CLAUDE.md architecture guide

### Tests ✅
- ✅ Test infrastructure (GitHub Actions, Makefile)
- ✅ favorites_test.go (100% coverage)
- ✅ file_operations_test.go (partial)
- ✅ trash_test.go (partial)
- ✅ helpers_test.go (good coverage)

---

## 🔴 Priority 1: Test Coverage (8.7% → 40%)

**Current:** 8.7% (28 tests)
**Target:** 40% (estimated 100+ tests)
**Estimated time:** 2 days

### Coverage Breakdown (Current)

| Module | Current Coverage | Priority | Target |
|--------|------------------|----------|--------|
| `favorites.go` | 86-100% | ✅ Done | - |
| `helpers.go` | 75-100% | ✅ Good | - |
| `file_operations.go` | 22% (partial) | 🔴 CRITICAL | 70% |
| `trash.go` | 15% (partial) | 🟡 HIGH | 60% |
| `editor.go` | 0% | 🟡 HIGH | 50% |
| `command.go` | 0% | 🟡 HIGH | 50% |
| `context_menu.go` | 0% | 🟢 MEDIUM | 30% |
| `dialog.go` | 0% | 🟢 LOW | 20% |

### Test Writing Plan

#### Phase 1: File Operations (Day 1, AM)
**Target:** Add 15% coverage (8.7% → 23.7%)

Create comprehensive `file_operations_test.go` additions:
```go
// Priority functions to test:
- TestLoadFiles() - directory reading
- TestLoadPreview() - file content loading
- TestGetFileIcon() - icon selection
- TestLoadSubdirFiles() - tree view expansion
- TestIsImageFile() / TestIsHTMLFile() / TestIsBrowserFile()
```

**Estimated:** 3-4 hours, +15 tests

#### Phase 2: Editor & Browser Integration (Day 1, PM)
**Target:** Add 10% coverage (23.7% → 33.7%)

Create `editor_test.go`:
```go
// Core editor functions:
- TestEditorAvailable() - check for nano, vim, etc.
- TestGetAvailableEditor() - priority selection
- TestGetAvailableBrowser() - platform detection
- TestIsWSL() - WSL detection
- TestIsImageFile() - file type detection
- TestGetAvailableImageViewer() - viu/timg/chafa
```

**Estimated:** 2-3 hours, +12 tests

#### Phase 3: Command Execution (Day 2, AM)
**Target:** Add 5% coverage (33.7% → 38.7%)

Create `command_test.go`:
```go
// Command system tests:
- TestShellQuote() - command sanitization
- TestAddToHistory() - command history
- TestGetPreviousCommand() / TestGetNextCommand()
```

**Estimated:** 2 hours, +8 tests

#### Phase 4: Trash & Context Menu (Day 2, PM)
**Target:** Add 3% coverage (38.7% → 41.7%)

Expand `trash_test.go` and create `context_menu_test.go`:
```go
// Trash operations:
- TestMoveToTrash() - trash file movement
- TestRestoreFromTrash() - restoration
- TestEmptyTrash() - permanent deletion
- TestGetTrashPath() - path calculation

// Context menu:
- TestGetContextMenuItems() - menu generation
- TestIsExecutableFile() - file permissions
```

**Estimated:** 2-3 hours, +10 tests

---

## 🟡 Priority 2: Documentation

### CONTRIBUTING.md (30 minutes)
Create a contributor guide with:
- How to report bugs
- How to submit pull requests
- Code style guidelines (refer to CLAUDE.md)
- Testing requirements (40% coverage minimum)
- Development setup

**Template structure:**
```markdown
# Contributing to TFE

## Reporting Bugs
- Use GitHub Issues
- Include terminal type, OS, Go version
- Provide steps to reproduce

## Pull Requests
- Fork the repository
- Create a feature branch
- Write tests (maintain 40%+ coverage)
- Follow architecture in CLAUDE.md
- Update README.md if adding features

## Development Setup
[Quick setup instructions]

## Code Style
- Follow Go conventions
- Keep files under 800 lines
- Add tests for new functions
- Use modules from CLAUDE.md

## Testing
go test ./...
make test-coverage

## Questions?
Open an issue for discussion
```

---

## 🟢 Priority 3: Remaining Security Items

### Items from Audit Report

1. **Symlink Safety Checks** (1 hour)
   - Add `os.Lstat()` instead of `os.Stat()` in file operations
   - Detect symlinks and warn users
   - Test symlink following behavior

2. **Path Traversal Validation** (1 hour)
   - Add `filepath.Clean()` to all path operations
   - Validate paths stay within intended boundaries
   - Test with `../` sequences

3. **Dependency Updates** (15 minutes)
   ```bash
   go get github.com/alecthomas/chroma/v2@latest
   go mod tidy
   ```

4. **Input Validation Helpers** (30 minutes)
   - Create `security.go` module with validation functions
   - Add dangerous pattern detection for commands
   - Centralize security checks

---

## 🟢 Priority 4: Final Polish

### Pre-Launch Checklist

**Testing:**
- [ ] All tests pass: `go test ./...`
- [ ] Coverage ≥40%: `make test-coverage`
- [ ] No race conditions: `make test-race`
- [ ] Benchmark performance: `make test-bench`

**Security:**
- [ ] No command injection vulnerabilities
- [ ] No goroutine leaks
- [ ] File size limits enforced
- [ ] Dependencies updated
- [ ] Symlinks handled safely

**Documentation:**
- [ ] README.md accurate and complete ✅
- [ ] CONTRIBUTING.md exists
- [ ] HOTKEYS.md up to date ✅
- [ ] CLAUDE.md reflects current architecture ✅
- [ ] Comments in code (aim for 25%+ comment coverage)

**Code Quality:**
- [ ] `go vet ./...` passes
- [ ] `go fmt ./...` applied
- [ ] No TODO/FIXME comments left unresolved
- [ ] All compiler warnings addressed

**User Experience:**
- [ ] Test on Linux ✅
- [ ] Test on macOS (if available)
- [ ] Test on Windows (if available)
- [ ] Test in Termux (mobile)
- [ ] Verify all F-keys work
- [ ] Verify mouse/touch controls
- [ ] Test with large directories (1000+ files)

**Release Preparation:**
- [ ] Version number set (v1.0.0)
- [ ] CHANGELOG.md updated
- [ ] Tag release in git
- [ ] Create GitHub release with binaries
- [ ] Update README badges (if any)

---

## Timeline to Launch

### Day 1 (6-8 hours)
**Morning (3-4 hours):**
- ✅ Fix failing test ✅
- Write file_operations tests (15% coverage)

**Afternoon (3-4 hours):**
- Write editor_test.go (10% coverage)
- Create CONTRIBUTING.md
- **Total coverage: ~34%**

### Day 2 (5-6 hours)
**Morning (2-3 hours):**
- Write command_test.go (5% coverage)
- Add symlink safety checks

**Afternoon (3 hours):**
- Write trash/context menu tests (3% coverage)
- Path traversal validation
- Update dependencies
- **Total coverage: ~42%** ✅

### Day 3 (2-3 hours) - Polish
**Morning:**
- Run full test suite
- Fix any issues found
- Code comments pass (add GoDoc comments)

**Afternoon:**
- Final validation
- Create v1.0.0 release
- Build binaries for release
- **LAUNCH! 🚀**

---

## Quick Start Commands

```bash
# Check current status
go test ./... -cover

# Run specific test
go test -v -run TestLoadFiles

# Generate coverage report
make test-coverage

# Run all pre-commit checks
make pre-commit

# Build for release
go build -o tfe

# Create release binaries
GOOS=linux GOARCH=amd64 go build -o tfe-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o tfe-darwin-amd64
GOOS=windows GOARCH=amd64 go build -o tfe-windows-amd64.exe
```

---

## Notes

### Why 40% Coverage?
- Covers all critical paths (file I/O, command execution)
- Industry standard for CLI tools
- Prevents regressions in core functionality
- Security-critical code fully tested

### What's NOT Being Tested?
- UI rendering functions (difficult to test in TUI)
- Full integration tests (can be added later)
- Platform-specific edge cases (covered by manual testing)

### After Launch (v1.1+)
- Increase coverage to 60%
- Add integration tests
- Add fuzzing tests for security
- Performance benchmarks
- Security scanning in CI/CD

---

## Current Blockers

**NONE!** 🎉

You have everything you need to get to 40% coverage and launch. The critical security fixes are done, the documentation is excellent, and the test infrastructure is in place.

**Next Step:** Start writing tests for `file_operations.go` functions. That's the biggest coverage win.

---

## Questions?

If you have questions about:
- **Testing approach** → See test files in repo for examples
- **Security concerns** → See COMPREHENSIVE_AUDIT_REPORT.md
- **Architecture** → See CLAUDE.md
- **Keyboard shortcuts** → See HOTKEYS.md

**Ready to start?** Let me know and I'll help you write the tests! 🚀
