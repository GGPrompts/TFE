# TFE Testing Improvement Session - Next Steps

## Context

Last night (Oct 28, 2025 at 02:31), a comprehensive audit of TFE was completed with an overall code health rating of **8.5/10**. The #1 priority issue identified was the **broken test suite**.

### Current Test Status
- **6 test files exist** but all are **currently failing**
- **Test Rating: 4/10** ⭐⭐⭐⭐
- Cannot run CI/CD without passing tests
- No regression protection currently available

---

## Session Goal

**Fix all failing tests and expand test coverage to 50% on critical paths.**

### Why This Matters
1. **Regression Protection** - Recent refactoring (trash UX, emoji fixes) had no test protection
2. **Safe Refactoring** - Can't confidently split large files without tests
3. **CI/CD Enablement** - Need passing tests to add GitHub Actions
4. **Code Quality** - Tests document expected behavior and catch edge cases

---

## Part 1: Fix Broken Tests (HIGH PRIORITY) 🔥

### Current Test Failures

```bash
# Run tests to see failures:
cd ~/projects/TFE
go test ./...
```

**Expected Failures:**

1. **favorites_test.go** - 7 errors:
   ```
   undefined: directoryContainsPrompts
   ```
   - Lines: 218, 228, 233, 254, 268, 289, 302
   - **Root Cause**: Function was likely removed/renamed during refactoring
   - **Fix Options:**
     - Option A: Remove tests that reference this function
     - Option B: Re-implement the function if still needed
     - Option C: Update tests to use new function name if renamed

2. **file_operations_test.go** - 2 errors:
   ```
   undefined: renderMarkdownWithTimeout
   ```
   - Lines: 544, 565
   - **Root Cause**: Function was likely removed/renamed during markdown rendering refactor
   - **Fix Options:**
     - Option A: Remove tests that reference this function
     - Option B: Re-implement the function if still needed
     - Option C: Update tests to call the actual markdown rendering function

3. **scripts/** - 1 error:
   ```
   main redeclared in this block
   ```
   - **Root Cause**: Multiple main() functions in scripts/
   - **Fix**: Add `// +build ignore` at top of script files OR move to separate directories

### Step-by-Step Fix Process

1. **Investigate Missing Functions**
   ```bash
   # Search for old function names in git history
   git log --all --full-history --source -- "*directoryContainsPrompts*"
   git log --all --full-history --source -- "*renderMarkdownWithTimeout*"

   # Search for current implementations
   grep -r "ContainsPrompts" .
   grep -r "renderMarkdown" .
   ```

2. **Fix Test Files**
   - Start with `favorites_test.go`
   - Then `file_operations_test.go`
   - Finally fix `scripts/` build issue

3. **Verify All Tests Pass**
   ```bash
   go test ./... -v
   go test -race ./...  # Check for race conditions
   ```

4. **Document What Was Fixed**
   - Update CHANGELOG.md with test fixes
   - Note any functionality that was removed

**Expected Time: 1-2 hours**

---

## Part 2: Expand Test Coverage (SHORT-TERM)

### Current Test Files
- ✅ `command_test.go` (6.9KB) - Command execution tests
- ✅ `editor_test.go` (11KB) - External editor integration
- ✅ `favorites_test.go` (12KB) - Favorites/bookmarks (NEEDS FIX)
- ✅ `file_operations_test.go` (24KB) - File operations (NEEDS FIX)
- ✅ `helpers_test.go` (12KB) - Helper functions
- ✅ `trash_test.go` (20KB) - Trash/recycle bin

### Missing Test Coverage (Priority Order)

#### 1. **helpers.go** - EXPAND EXISTING ✅
**Why**: Core utilities used everywhere, already has tests
**Current Coverage**: Partial (12KB test file exists)
**Add Tests For:**
- `navigateToPath()` - NEW function from trash refactor
- `visualWidth()` - Critical for emoji alignment
- `truncateToWidth()` - Used in all rendering
- Edge cases: empty strings, very long strings, emoji edge cases

**Example Test:**
```go
func TestNavigateToPath(t *testing.T) {
    tests := []struct {
        name           string
        currentPath    string
        newPath        string
        inTrash        bool
        expectTrashExit bool
    }{
        {"navigate while in trash", "/home/user", "/tmp", true, true},
        {"navigate when not in trash", "/home/user", "/tmp", false, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

#### 2. **file_loading.go** - CREATE NEW 📝
**Why**: Core functionality, used constantly
**Test Cases:**
- Loading empty directory
- Loading directory with 100 files
- Loading directory with 10,000 files (performance)
- Hidden file filtering
- Permission errors
- Non-existent paths
- Symbolic links

#### 3. **menu.go** - CREATE NEW 📝
**Why**: Complex state machine, recently modified for trash
**Test Cases:**
- Menu action dispatch (toggle-favorites, toggle-prompts, etc.)
- Trash auto-exit on menu actions
- Menu state transitions

#### 4. **update_keyboard.go** - CREATE NEW (PARTIAL) 📝
**Why**: 109 case statements, hard to test but critical
**Test Strategy**: Test state transitions, not UI rendering
**Test Cases:**
- Trash auto-exit on navigation keys
- Mode switching (F1/F2/F3)
- Filter toggles (F6/F11/F12)

#### 5. **types.go** - CREATE NEW 📝
**Why**: Data structures should have validation tests
**Test Cases:**
- `fileItem` creation and validation
- `model` initialization
- Field constraints (e.g., cursor bounds)

### Go Testing Best Practices

```go
// Use table-driven tests for multiple scenarios
func TestFormatFileSize(t *testing.T) {
    tests := []struct {
        name     string
        bytes    int64
        expected string
    }{
        {"zero", 0, "0 B"},
        {"bytes", 500, "500 B"},
        {"kilobytes", 1024, "1.0 KB"},
        {"megabytes", 1048576, "1.0 MB"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := formatFileSize(tt.bytes)
            if result != tt.expected {
                t.Errorf("formatFileSize(%d) = %s; want %s",
                    tt.bytes, result, tt.expected)
            }
        })
    }
}

// Use subtests for organization
func TestTrashOperations(t *testing.T) {
    t.Run("delete file", func(t *testing.T) { /* ... */ })
    t.Run("restore file", func(t *testing.T) { /* ... */ })
    t.Run("empty trash", func(t *testing.T) { /* ... */ })
}

// Use test helpers for setup/teardown
func setupTestDir(t *testing.T) string {
    t.Helper()
    dir, err := os.MkdirTemp("", "tfe-test-*")
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { os.RemoveAll(dir) })
    return dir
}
```

**Expected Time: 16-24 hours for 50% coverage**

---

## Part 3: Add CI/CD Pipeline (AFTER TESTS PASS)

### GitHub Actions Workflow

Create `.github/workflows/test.yml`:

```yaml
name: Tests

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest]
        go: ['1.24', '1.25']

    runs-on: ${{ matrix.os }}

    steps:
    - uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go }}

    - name: Run tests
      run: go test -v -race -coverprofile=coverage.txt ./...

    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        files: ./coverage.txt
```

**Expected Time: 2-4 hours**

---

## Part 4: Performance Testing (AFTER COVERAGE)

### Large Directory Test

```bash
# Create test directory with 10,000 files
mkdir -p ~/tfe_perf_test
cd ~/tfe_perf_test
for i in {1..10000}; do touch "file_$i.txt"; done

# Launch TFE and measure:
# 1. Initial load time
# 2. Scroll performance
# 3. Search speed
# 4. Memory usage
time tfe
```

### Benchmark Tests

```go
func BenchmarkLoadFiles(b *testing.B) {
    // Setup: Create directory with 1000 files
    dir := setupLargeTestDir(b, 1000)
    m := initialModel()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        m.currentPath = dir
        m.loadFiles()
    }
}
```

**Expected Time: 4-8 hours**

---

## Quick Reference: Go Test Commands

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific test file
go test -v ./helpers_test.go

# Run specific test function
go test -v -run TestNavigateToPath

# Run with race detector
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run benchmarks
go test -bench=. ./...

# Run tests continuously (watch mode)
# Install: go install github.com/cespare/reflex@latest
reflex -r '\.go$' -s -- go test ./...
```

---

## Success Criteria

### Phase 1: Fix Tests (IMMEDIATE)
- ✅ All `go test ./...` pass without errors
- ✅ No undefined function references
- ✅ Scripts build issue resolved

### Phase 2: Expand Coverage (SHORT-TERM)
- ✅ helpers.go coverage expanded (navigateToPath, visualWidth, truncateToWidth)
- ✅ file_loading tests created
- ✅ menu.go tests created (trash auto-exit)
- ✅ update_keyboard.go tests created (key state transitions)
- ✅ types.go tests created
- ✅ Overall coverage reaches 50%

### Phase 3: CI/CD (AFTER TESTS PASS)
- ✅ GitHub Actions workflow added
- ✅ Tests run on every PR
- ✅ Coverage tracking enabled

### Phase 4: Performance (LONG-TERM)
- ✅ Benchmark tests added
- ✅ Large directory test (10,000+ files) passes
- ✅ Performance baseline documented

---

## Audit Summary (Full Report: AUDIT_REPORT_2025.md)

### Overall Rating: 8.5/10 ⭐⭐⭐⭐⭐⭐⭐⭐⭐

**Strengths:**
- ✅ Outstanding documentation (CLAUDE.md, LESSONS_LEARNED.md)
- ✅ Clean modular architecture (main.go: 70 lines)
- ✅ Zero technical debt markers (no TODO/FIXME/HACK)
- ✅ Excellent error handling (no panic/log.Fatal)
- ✅ Mobile-aware design (Termux support)

**Critical Issues:**
- ⚠️ **Broken test suite** (MUST FIX FIRST)
- ⚠️ Large files need splitting (file_operations.go: 72KB)
- ⚠️ Performance testing needed (10,000+ files)

**Quick Wins:**
1. Fix tests (1-2 hours) ← **START HERE**
2. Add CI/CD (2-4 hours)
3. Extract magic numbers (30 min)
4. Add pre-commit hook (15 min)

---

## Resources

- **Audit Report**: `AUDIT_REPORT_2025.md` (Oct 28, 02:31)
- **Architecture Guide**: `CLAUDE.md`
- **Lessons Learned**: `docs/LESSONS_LEARNED.md`
- **Module Details**: `docs/MODULE_DETAILS.md`
- **Go Testing Guide**: https://go.dev/doc/tutorial/add-a-test
- **Table-Driven Tests**: https://dave.cheney.net/2019/05/07/prefer-table-driven-tests

---

## Prompt for Next Session

```
I need to fix TFE's broken test suite and expand test coverage. Here's the context:

**Immediate Priority:**
1. Fix broken tests in favorites_test.go (undefined: directoryContainsPrompts)
2. Fix broken tests in file_operations_test.go (undefined: renderMarkdownWithTimeout)
3. Fix scripts/ build issue (duplicate main functions)
4. Verify all tests pass: go test ./...

**After Tests Pass:**
5. Expand test coverage for recently refactored trash navigation feature
6. Add tests for helpers.go: navigateToPath(), visualWidth(), truncateToWidth()
7. Create tests for menu.go trash auto-exit behavior
8. Target 50% overall test coverage

**Reference Documents:**
- Full audit: AUDIT_REPORT_2025.md
- Testing plan: docs/NEXT_SESSION_TESTING.md
- Architecture: CLAUDE.md

**Goal:** Get test suite passing, then expand to 50% coverage on critical paths.

Please start by investigating the missing functions (directoryContainsPrompts,
renderMarkdownWithTimeout) and recommend the best fix approach.
```

---

**Created:** Oct 28, 2025, 14:30
**For Use In:** Next development session
**Estimated Total Time:** 20-32 hours (spread across multiple sessions)
