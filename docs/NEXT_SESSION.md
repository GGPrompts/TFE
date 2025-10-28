# TFE Next Session - Quick Wins & Testing

**Created:** Oct 28, 2025, 14:35
**Priority Order:** Quick Win First (30 min), Then Testing (20+ hours)

---

## Prompt 1: Quick Win - Extract Shared Header Function (30-60 minutes) ⚡

```
I need to implement a quick win from the TFE audit: extract the duplicated header
rendering code into a shared function.

**Current Issue:**
The toolbar row (emoji buttons: 🏠 ⭐ V ⬜ >_ 🔍 📝 🔀 🗑) is duplicated in two locations:
- view.go: renderSinglePane() (~line 231-235 for trash icon)
- render_preview.go: renderDualPane() (~line 1279-1281 for trash icon)

This duplication means:
- Changes must be made in two places (error-prone)
- Recently had to update trash icon (🗑/♻) in both locations
- ~50 lines of duplicated code

**Goal:**
Create a shared function that renders the toolbar row and replace both duplicates.

**Implementation Steps:**

1. Extract to helpers.go:
```go
// renderToolbarRow renders the emoji button toolbar row
// Shows: [🏠] [⭐/✨] [V] [⬜/⬌] [>_] [🔍] [📝] [🔀] [🗑/♻]
func renderToolbarRow(m model) string {
    var s strings.Builder

    // Home button [🏠]
    homeButtonStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
    s.WriteString(homeButtonStyle.Render("[🏠]"))
    s.WriteString(" ")

    // Star button [⭐/✨] - shows ✨ when favorites filter is active
    starIcon := "⭐"
    if m.showFavoritesOnly {
        starIcon = "✨"
    }
    s.WriteString(homeButtonStyle.Render("[" + starIcon + "]"))
    s.WriteString(" ")

    // View mode button - shows different emoji for each mode
    viewIcon := "📊" // Detail view (default)
    switch m.displayMode {
    case modeList:
        viewIcon = "📄" // Document icon for simple list view
    case modeDetail:
        viewIcon = "📊" // Bar chart icon for detailed columns
    case modeTree:
        viewIcon = "🌲" // Tree icon for hierarchical view
    }
    s.WriteString(homeButtonStyle.Render("[" + viewIcon + "]"))
    s.WriteString(" ")

    // ... (continue with all buttons through trash icon)

    // Trash/Recycle bin button
    trashIcon := "🗑"
    if m.showTrashOnly {
        trashIcon = "♻" // Recycle icon when viewing trash
    }
    s.WriteString(homeButtonStyle.Render("[" + trashIcon + "]"))

    return s.String()
}
```

2. Update view.go (renderSinglePane):
   - Find toolbar rendering code (~lines 195-235)
   - Replace with: `s.WriteString(renderToolbarRow(m))`

3. Update render_preview.go (renderDualPane):
   - Find toolbar rendering code (~lines 1245-1285)
   - Replace with: `s.WriteString(renderToolbarRow(m))`

4. Test both modes work correctly:
   - Single-pane view
   - Dual-pane view
   - All emoji buttons appear correctly
   - Trash icon switches between 🗑 and ♻

5. Build and verify:
   ```bash
   go build
   ./tfe  # Test single-pane
   # Press Tab to test dual-pane
   # Press F12 to verify trash icon changes
   ```

**Files to Modify:**
- helpers.go (add renderToolbarRow function)
- view.go (replace toolbar code in renderSinglePane)
- render_preview.go (replace toolbar code in renderDualPane)

**Success Criteria:**
- ✅ Toolbar renders identically in both modes
- ✅ All emoji buttons work (home, favorites, view mode, etc.)
- ✅ Trash icon switches correctly (🗑 → ♻)
- ✅ Code reduced by ~50 lines
- ✅ Future changes only need one location

**Audit Reference:** AUDIT_REPORT_2025.md, Section 1.2 "Quick Wins"
**Estimated Time:** 30-60 minutes
```

---

## Prompt 2: Fix Broken Test Suite & Expand Coverage (20-32 hours) 🔥

```
I need to fix TFE's broken test suite and expand test coverage. Here's the context:

**Immediate Priority (MUST DO FIRST):**
1. Fix broken tests in favorites_test.go (undefined: directoryContainsPrompts)
2. Fix broken tests in file_operations_test.go (undefined: renderMarkdownWithTimeout)
3. Fix scripts/ build issue (duplicate main functions)
4. Verify all tests pass: go test ./...

**Current Test Failures:**

Run this to see failures:
```bash
cd ~/projects/TFE
go test ./...
```

Expected output:
```
./favorites_test.go:218:6: undefined: directoryContainsPrompts
./favorites_test.go:228:5: undefined: directoryContainsPrompts
./favorites_test.go:233:5: undefined: directoryContainsPrompts
./favorites_test.go:254:5: undefined: directoryContainsPrompts
./favorites_test.go:268:6: undefined: directoryContainsPrompts
./favorites_test.go:289:5: undefined: directoryContainsPrompts
./favorites_test.go:302:6: undefined: directoryContainsPrompts
./file_operations_test.go:544:21: undefined: renderMarkdownWithTimeout
./file_operations_test.go:565:12: undefined: renderMarkdownWithTimeout
scripts/emoji_audit_targeted.go:14:6: main redeclared in this block
```

**Investigation Steps:**

1. Search for missing functions in git history:
```bash
git log --all --full-history --source -- "*directoryContainsPrompts*"
git log --all --full-history --source -- "*renderMarkdownWithTimeout*"
```

2. Search for current implementations:
```bash
grep -r "ContainsPrompts" .
grep -r "renderMarkdown" .
```

3. Determine fix approach:
   - Option A: Remove tests that reference deleted functions
   - Option B: Re-implement functions if still needed
   - Option C: Update tests to use new function names if renamed

**After Tests Pass:**
4. Expand test coverage for recently refactored trash navigation feature
5. Add tests for helpers.go: navigateToPath(), visualWidth(), truncateToWidth()
6. Create tests for menu.go trash auto-exit behavior
7. Add tests for update_keyboard.go key state transitions
8. Target 50% overall test coverage on critical paths

**Testing Best Practices:**

Use table-driven tests:
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

**Reference Documents:**
- Full audit report: AUDIT_REPORT_2025.md
- Detailed testing plan: docs/NEXT_SESSION_TESTING.md
- Architecture guide: CLAUDE.md
- Module details: docs/MODULE_DETAILS.md

**Success Criteria - Phase 1 (Fix Tests):**
- ✅ All `go test ./...` pass without errors
- ✅ No undefined function references
- ✅ Scripts build issue resolved

**Success Criteria - Phase 2 (Expand Coverage):**
- ✅ helpers.go coverage expanded (navigateToPath, visualWidth, truncateToWidth)
- ✅ file_loading tests created
- ✅ menu.go tests created (trash auto-exit)
- ✅ update_keyboard.go tests created (key state transitions)
- ✅ types.go tests created
- ✅ Overall coverage reaches 50%

**Useful Commands:**
```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific test
go test -v -run TestNavigateToPath

# Check for race conditions
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run benchmarks
go test -bench=. ./...
```

**Goal:** Get test suite passing, then expand to 50% coverage on critical paths.

**Estimated Time:**
- Phase 1 (Fix tests): 1-2 hours
- Phase 2 (Expand coverage): 16-24 hours
- Phase 3 (CI/CD setup): 2-4 hours
- Total: 20-32 hours (spread across multiple sessions)

Please start by investigating the missing functions and recommend the best fix approach.
```

---

## Additional Context

### Why Testing is Priority #1

From the audit (AUDIT_REPORT_2025.md):
- **Overall Code Health:** 8.5/10 ⭐⭐⭐⭐⭐⭐⭐⭐⭐
- **Testing Rating:** 4/10 ⭐⭐⭐⭐ (pulls down overall score)
- **Top Issue:** "Broken Test Suite - Tests exist but currently fail"

**Impact of No Tests:**
- Cannot verify code correctness
- Refactoring is risky without tests
- Regression detection is impossible
- CI/CD pipeline cannot validate changes
- Recent trash refactor had zero test protection

### Current Test Files
- ✅ command_test.go (6.9KB) - Command execution
- ✅ editor_test.go (11KB) - External editor integration
- ❌ favorites_test.go (12KB) - BROKEN: directoryContainsPrompts
- ❌ file_operations_test.go (24KB) - BROKEN: renderMarkdownWithTimeout
- ✅ helpers_test.go (12KB) - Helper functions
- ✅ trash_test.go (20KB) - Trash/recycle bin

### Session Order Recommendation

**Session 1: Quick Win (30-60 min)**
- Extract shared header function
- Simple, low-risk refactor
- Immediate code quality improvement
- Good warm-up before diving into testing

**Session 2+: Testing (20-32 hours)**
- Fix broken tests (1-2 hours) - CRITICAL
- Expand coverage (16-24 hours) - HIGH PRIORITY
- Add CI/CD (2-4 hours) - AFTER TESTS PASS
- Performance testing (4-8 hours) - LONG-TERM

---

**Last Updated:** Oct 28, 2025, 14:35
**Auto-compact at:** ~4% remaining context
