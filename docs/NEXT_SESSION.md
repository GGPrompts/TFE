# Pre-Launch Comprehensive Review

**Goal:** Conduct a thorough parallel review of TFE before public release, identifying issues across security, performance, UX, code quality, and mobile optimization.

---

## 🚀 IMPORTANT: Run All Agents in PARALLEL

Launch all 6 review agents **in a single message with multiple Task tool calls** for maximum efficiency.

---

## Review Agent Instructions

### 1. Security Review Agent
**Agent Type:** `code-reviewer`
**Task:**
```
Perform a comprehensive security audit of the TFE (Terminal File Explorer) codebase.

Focus areas:
- File operation safety (path traversal, symlink attacks, permission checks)
- Command execution vulnerabilities (command injection in executeCommand, runCommand)
- Input validation (dialog inputs, command prompt, file picker, search queries)
- Secrets exposure risks (.env files, credentials in favorites/history)
- Terminal escape sequence injection risks
- Race conditions in file operations
- Directory traversal with .. handling
- Favorites/history file permissions and storage security

Search these files:
- file_operations.go (loadFiles, loadPreview, file I/O)
- command.go (executeCommand, runCommand, history storage)
- editor.go (external tool launching)
- update_keyboard.go (input handling)
- dialog.go (user input validation)

Provide specific findings with:
- File path and line number
- Severity: Critical/High/Medium/Low
- Description of vulnerability
- Recommended fix
- Example exploit if applicable

Output format:
## Critical Issues
- [file:line] Description - Recommended fix

## High Priority Issues
- [file:line] Description - Recommended fix

## Medium/Low Priority
- [file:line] Description - Recommended fix
```

---

### 2. Mobile/Narrow Terminal UX Agent
**Agent Type:** `Explore` (thoroughness: **very thorough**)
**Task:**
```
Review mobile and narrow terminal support (width < 100 columns, primarily Termux on Android).

Recent changes to validate:
- Width calculation fixes in model.go, render_file_list.go, render_preview.go
- Emoji/wide character handling with runewidth library
- Horizontal scrolling in detail view
- Vertical split (top/bottom) panes on narrow terminals
- Mouse coordinate fixes in update_mouse.go
- Prompt template header wrapping

Check for:
- Inconsistent use of m.isNarrowTerminal() checks
- Hardcoded widths that should be dynamic
- Text wrapping issues (backgrounds, selections, borders)
- Visual width vs byte length vs rune count confusion
- Mouse click coordinate mismatches
- Pane height mismatches in vertical split
- Header/data alignment issues in detail view
- Preview content touching borders without padding

Search these files:
- render_file_list.go (list/detail/tree views)
- render_preview.go (preview pane, markdown rendering)
- update_mouse.go (click detection)
- model.go (calculateLayout, width calculations)
- file_operations.go (preview width calculations)

Provide:
- Areas where narrow terminal support is incomplete
- Specific width calculation bugs
- Any remaining hardcoded values
- User experience issues on phones

Output format:
## Critical UX Issues (Blocks mobile usage)
- [file:line] Description

## Improvements Needed
- [file:line] Description

## Well-Implemented Areas
- What's working correctly
```

---

### 3. Performance & Resource Agent
**Agent Type:** `debugger`
**Task:**
```
Analyze TFE for performance bottlenecks, resource leaks, and optimization opportunities.

Focus areas:
- Memory leaks (unclosed files, unbounded slices/maps)
- Expensive operations in rendering loops
- File loading limits and large directory handling
- Glamour markdown rendering timeouts
- Cache invalidation/population logic
- Blocking operations that could hang UI
- Redundant file reads or operations
- Tree view expansion performance with deep hierarchies

Search these files:
- file_operations.go (loadFiles, loadPreview, caching)
- render_*.go (rendering loops)
- update.go, update_keyboard.go, update_mouse.go (event handling)
- command.go (command execution, history storage)

Look for:
- Functions with defer file.Close() - are all files closed?
- Unbounded appends to slices (history, expandedDirs, etc.)
- Expensive operations inside for loops
- Synchronous file I/O in UI thread
- Missing limits on preview file size, directory entries
- Cache invalidation bugs causing excessive re-renders

Provide:
- Specific performance bottlenecks with measurements if possible
- Memory leak risks
- Optimization suggestions with priority

Output format:
## Critical Performance Issues
- [file:line] Description - Impact - Fix

## Optimization Opportunities
- [file:line] Description - Expected improvement

## Resource Management
- Areas handling resources correctly
- Areas needing improvement
```

---

### 4. Code Quality & Architecture Agent
**Agent Type:** `code-reviewer`
**Task:**
```
Review code quality and adherence to CLAUDE.md architectural principles.

Check CLAUDE.md first to understand:
- Modular architecture (11 focused files)
- Single responsibility per file
- Module size limits (target <800 lines)
- Documentation requirements

Then review:
- Functions >100 lines that should be split
- Duplicate code needing extraction
- Error handling consistency
- TODO/FIXME comments indicating incomplete work
- Naming consistency
- Adherence to Go best practices
- Dead code or unused functions

Search these files:
- All .go files in root directory
- Check against CLAUDE.md architecture guidelines

Provide:
- Functions violating size/complexity guidelines
- Duplicate code opportunities
- Architectural violations
- Missing error handling
- Code smell patterns

Output format:
## Architectural Issues
- [file] Violation - Recommendation

## Code Quality Issues
- [file:line] Issue - Suggested refactoring

## Well-Architected Areas
- What's following best practices
```

---

### 5. User Experience & Edge Cases Agent
**Agent Type:** `general-purpose`
**Task:**
```
Review UX and test edge case handling across TFE functionality.

Test scenarios:
- Empty directories, permission-denied folders
- Very long filenames (>255 chars), paths, symlink chains
- Circular symlinks, broken symlinks
- Files with special characters in names
- Binary files, very large files (>1GB)
- Rapid navigation, quick key presses
- All keyboard shortcuts for conflicts
- Dialog inputs with invalid characters
- Context menu on every file type
- Favorites with deleted/moved files
- Command history with very long commands
- Prompts with many variables (>20)

Check:
- Status messages for clarity
- Error messages for user-friendliness
- Help text (HOTKEYS.md) accuracy
- Graceful degradation when features unavailable
- Recovery from errors without crashing

Search these files:
- All view/render files for user-facing text
- HOTKEYS.md for documentation accuracy
- update_keyboard.go for shortcut conflicts
- dialog.go for input validation
- file_operations.go for edge case handling

Provide:
- Missing error messages
- Confusing UX patterns
- Keyboard shortcut conflicts
- Edge cases causing crashes or hangs

Output format:
## UX Issues
- Description - Recommended improvement

## Missing Error Handling
- [file:line] Scenario - What should happen

## Documentation Gaps
- What's missing or inaccurate
```

---

### 6. Documentation & Launch Readiness Agent
**Agent Type:** `documentation`
**Task:**
```
Review all documentation for completeness and launch readiness.

Check:
- README.md (installation, features, screenshots, usage)
- HOTKEYS.md (all shortcuts documented and accurate)
- CHANGELOG.md (recent versions complete, under 350 line limit per CLAUDE.md)
- CLAUDE.md (architecture doc current, under 500 line limit)
- PLAN.md (under 400 line limit)
- GitHub repo metadata (description, topics, license, about section)
- Error messages user-facing (helpful, not technical)
- Missing docs (contributing guide, FAQ, troubleshooting)

Compare HOTKEYS.md against actual keybindings in:
- update_keyboard.go
- update_mouse.go

Verify:
- All F-keys documented
- All context menu options documented
- All special modes documented (prompts, favorites, trash)
- Mobile-specific guidance included

Provide:
- Missing documentation sections
- Outdated or inaccurate information
- Documentation exceeding line limits
- Launch checklist items

Output format:
## Critical Documentation Gaps
- What's missing for launch

## Documentation Updates Needed
- What's inaccurate or incomplete

## Launch Readiness Checklist
- [ ] Task 1
- [ ] Task 2

## Well-Documented Areas
- What's complete and accurate
```

---

## After All Agents Complete

Once all 6 agents have reported their findings, please:

1. **Read all agent reports carefully**
2. **Synthesize and categorize all findings**
3. **Identify patterns** - Are there systemic issues?
4. **Prioritize by severity** - Critical → High → Medium → Low
5. **Assess launch readiness** - Can we launch or must we fix first?
6. **Create actionable task list** with estimated effort

---

## Final Summary Format

```markdown
# TFE Pre-Launch Review Summary

## Executive Summary
[2-3 paragraphs: Overall state, major findings, launch recommendation]

---

## Critical Issues (MUST Fix Before Launch)
1. **[Category]** - [file:line] - Description
   - **Impact:** Why this blocks launch
   - **Fix:** What needs to happen
   - **Effort:** Estimated time

2. **[Category]** - [file:line] - Description
   ...

---

## High Priority (Should Fix Before Launch)
1. **[Category]** - [file:line] - Description
   - **Impact:** Why this matters
   - **Fix:** Recommended solution
   - **Effort:** Estimated time

---

## Medium Priority (Fix Soon After Launch)
[List with brief descriptions]

---

## Low Priority / Future Enhancements
[List with brief descriptions]

---

## Positive Findings
### Security
- [What's secure and well-implemented]

### Mobile Optimization
- [What's working well on narrow terminals]

### Performance
- [What's efficient and fast]

### Code Quality
- [Well-architected areas]

### Documentation
- [Complete and accurate docs]

---

## Systemic Patterns Identified
- Pattern 1: [Description across multiple files]
- Pattern 2: [Description of recurring issue]

---

## Launch Readiness Assessment

**Overall Status:** ✅ Ready / ⚠️ Ready with Caveats / ❌ Not Ready

**Reasoning:**
[Detailed explanation of why this status was chosen]

**Blocking Issues:** [Count]
**Must-Fix Issues:** [Count]
**Nice-to-Have Issues:** [Count]

---

## Recommended Action Plan

### Immediate (Before Launch)
1. [ ] Fix [Critical Issue 1] - [File] - [Estimated: X hours]
2. [ ] Fix [Critical Issue 2] - [File] - [Estimated: X hours]

### Short-Term (Week 1 After Launch)
1. [ ] Address [High Priority 1]
2. [ ] Address [High Priority 2]

### Medium-Term (Month 1)
1. [ ] Implement [Medium Priority improvements]
2. [ ] Add [Missing features]

### Long-Term (Future Versions)
1. [ ] Enhance [Low Priority features]
2. [ ] Optimize [Performance improvements]

---

## Testing Recommendations
- [ ] Test on Android/Termux with width < 100
- [ ] Test with very large directories (10,000+ files)
- [ ] Test all keyboard shortcuts for conflicts
- [ ] Test symlink handling (circular, broken, nested)
- [ ] Test command execution with special characters
- [ ] Fuzz test file picker and dialogs

---

## Conclusion
[Final recommendation: Launch now / Fix X issues first / Major work needed]
```

---

## Context for Agents

**Project:** TFE (Terminal File Explorer)
**Target Users:** Mobile developers using Termux on Android, WSL users, TUI enthusiasts
**Primary Use Case:** File management on phones via Termux
**Key Features:** Tree/list/detail views, dual-pane, markdown preview, prompts, favorites, trash, command execution, symlinks
**Recent Work:** Massive mobile optimization (width calcs, emoji handling, scrolling, mouse coords, prompt templates)
**Tech Stack:** Go, Bubbletea, Lipgloss, Glamour, go-runewidth

**Ready for public launch if no critical issues found.**

---

**REMEMBER:** Run all 6 agents **in parallel** using a single message with 6 Task tool calls!
