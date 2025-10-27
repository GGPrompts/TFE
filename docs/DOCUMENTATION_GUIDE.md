# TFE Documentation Management Guide

This document describes how documentation is organized and maintained in the TFE project.

## Problem Statement

Documentation files can grow too large, making them hard to read, navigate, and load into AI context. This leads to "documentation bloat" that makes projects unmaintainable and slows down AI assistants.

## Solution

Strict line limits and archiving rules for all documentation files, with a clear workflow for moving information through different stages.

---

## Core Documentation Files

These files live in the project root and should be kept concise:

| File | Max Lines | Purpose | When to Clean |
|------|-----------|---------|---------------|
| **CLAUDE.md** | 500 | Architecture guide for AI assistants | Archive old sections to `docs/archive/` |
| **README.md** | 600 | Project overview, installation, usage | Split detailed docs to `docs/` (user-facing, can be longer) |
| **PLAN.md** | 400 | Current roadmap & planned features | Move completed items to CHANGELOG.md |
| **CHANGELOG.md** | 350 | Recent changes & release notes | Create CHANGELOG2.md when exceeds limit |
| **BACKLOG.md** | 300 | Ideas & future features (brainstorming) | Move refined ideas to PLAN.md or archive |
| **HOTKEYS.md** | - | User-facing keyboard shortcuts | Keep the list comprehensive and current |

---

## Detailed Documentation Files (docs/ directory)

These files can be longer but should still be focused:

| File | Max Lines | Purpose |
|------|-----------|---------|
| **docs/MODULE_DETAILS.md** | 800 | Full module descriptions and responsibilities |
| **docs/DEVELOPMENT_PATTERNS.md** | 600 | Detailed examples for adding features |
| **docs/LESSONS_LEARNED.md** | 800 | Critical lessons from debugging sessions |
| **docs/THREAT_MODEL.md** | 300 | Security philosophy and threat model |
| **docs/REFACTORING_HISTORY.md** | - | Historical refactoring timeline |
| **docs/NEXT_SESSION.md** | 200 | Current work session notes (delete when done) |

---

## Documentation Workflow

Information flows through different files as it matures:

### Stage 1: Idea → BACKLOG.md
**Purpose**: Brainstorming and idea parking lot

**Contents:**
- Raw ideas and "nice to have" features
- Things that need more research
- Concepts that aren't prioritized yet
- Low-priority feature requests

**When to move forward**: Idea is refined, has clear requirements, and becomes a priority

**Example:**
```markdown
## Future Ideas
- [ ] Add SSH file browser integration (needs research on libraries)
- [ ] Implement file tagging system (low priority)
- [ ] Add plugin system for custom file viewers (complex, future)
```

---

### Stage 2: Planning → PLAN.md
**Purpose**: Active roadmap and prioritized features

**Contents:**
- Refined ideas with clear requirements
- Prioritized features ready for implementation
- Current sprint/milestone goals
- Features actively being designed

**When to move forward**: Implementation is complete

**Example:**
```markdown
## Current Sprint

### High Priority
- [ ] Fix emoji alignment in Detail view
- [ ] Add image preview with terminal graphics protocols
- [ ] Implement git operations (pull, push, sync)

### Medium Priority
- [ ] Add fuzzy search with Ctrl+P
- [ ] Implement favorites/bookmarks system
```

---

### Stage 3: Implementation → docs/NEXT_SESSION.md
**Purpose**: Session-specific implementation notes

**Contents:**
- Detailed implementation plans for current work
- Step-by-step checklists
- Debugging notes and findings
- "What to do next" instructions

**When to move forward**: Work is completed or session ends

**After completion:**
- Delete if no longer relevant
- Archive to `docs/archive/` if contains valuable lessons
- Extract lessons to `docs/LESSONS_LEARNED.md` if broadly applicable

**Example:**
```markdown
# Next Session: Fix Emoji Alignment

## Context
Users reported emoji misalignment in WezTerm and Termux terminals.

## Steps
1. [ ] Audit all emoji usage with `xxd` to find variation selectors
2. [ ] Replace emoji+VS with base emoji characters
3. [ ] Remove terminal-specific workaround code
4. [ ] Test in WezTerm, Termux, Windows Terminal

## Notes
- go-runewidth bug #76 causes VS to count as width=1
- See docs/LESSONS_LEARNED.md for visual width calculation rules
```

---

### Stage 4: Completion → CHANGELOG.md
**Purpose**: Record of what was implemented

**Contents:**
- Brief description of implemented features
- Version number and date
- Key changes and bug fixes
- Breaking changes

**Format:**
```markdown
## v0.3.0 - 2025-10-27

### Added
- Image preview with terminal graphics protocols (Kitty, iTerm2, Sixel)
- Fuzzy file search with Ctrl+P (using fzf + fd)
- Git operations: pull, push, sync, fetch

### Fixed
- Emoji alignment issues across all terminals
- Image viewer text bleed-through after viewing multiple images

### Changed
- Removed ~100 lines of variation selector workaround code
```

---

### Stage 5: Research → docs/
**Purpose**: Detailed research and reference documents

**Contents:**
- Research documents can be large but should be split by topic
- One topic per file (e.g., `RESEARCH_UI_FRAMEWORKS.md`)
- Archive when no longer relevant

**Example files:**
- `docs/TERMINAL_PROTOCOLS.md` - Research on graphics protocols
- `docs/LIBRARY_COMPARISON.md` - Comparison of different libraries
- `docs/PERFORMANCE_ANALYSIS.md` - Performance benchmarks and optimizations

---

## Managing File Growth

### CHANGELOG Approach (Keep History Visible)

When CHANGELOG.md exceeds 350 lines:
1. Create CHANGELOG2.md
2. Move older entries (v0.1.x, v0.2.x) to CHANGELOG2.md
3. Keep recent versions (latest 3-4) in CHANGELOG.md
4. Add link: "See CHANGELOG2.md for older versions"
5. Continue pattern: CHANGELOG3.md, CHANGELOG4.md as needed

**Example structure:**
```
CHANGELOG.md      → v0.5.0, v0.4.0, v0.3.0 (current + recent)
CHANGELOG2.md     → v0.2.0, v0.1.5, v0.1.0 (older versions)
```

### Other Files

**PLAN.md exceeds 400 lines:**
- Move completed items to CHANGELOG.md
- Defer low-priority items to BACKLOG.md
- Archive cancelled/rejected items

**BACKLOG.md exceeds 300 lines:**
- Archive old/rejected ideas to `docs/archive/BACKLOG_OLD.md`
- Remove truly obsolete ideas
- Promote refined ideas to PLAN.md

**CLAUDE.md exceeds 500 lines:**
- Archive old sections to `docs/archive/`
- Move detailed content to docs/ files
- Keep CLAUDE.md as an index with pointers

**Research docs exceed 1000 lines:**
- Split into multiple focused documents
- Archive outdated sections
- Create index file if multiple related docs exist

---

## AI Assistant Reminders

**For Claude Code:**

When working on TFE, Claude should:

1. **Check doc sizes proactively:**
   ```bash
   wc -l *.md docs/*.md
   ```

2. **Suggest cleanup when limits exceeded:**
   - If any core doc exceeds its line limit, proactively suggest cleanup
   - When CHANGELOG.md exceeds 350 lines, create CHANGELOG2.md
   - When adding to PLAN.md, check if it's grown too large

3. **Move completed work:**
   - Suggest moving completed PLAN.md items to CHANGELOG.md
   - Keep NEXT_SESSION.md focused on current work only
   - Archive or delete NEXT_SESSION.md when work is complete

4. **Extract lessons:**
   - When debugging reveals important insights, add to `docs/LESSONS_LEARNED.md`
   - When creating new patterns, add to `docs/DEVELOPMENT_PATTERNS.md`

---

## Benefits of This System

| Benefit | Description |
|---------|-------------|
| **AI Context Efficiency** | Smaller files load faster into AI context windows |
| **Human Readability** | Easier to scan and find information |
| **Project Maintainability** | Clear separation prevents documentation sprawl |
| **Prevents Bloat** | Proactive limits catch growth early |
| **Clear Workflow** | Always know where information belongs |
| **Historical Context** | CHANGELOG2.md, CHANGELOG3.md preserve history |

---

## Archive Strategy

### When to Archive

Archive documentation when:
- Information is no longer actively referenced
- Feature has been deprecated or removed
- Research is outdated or superseded
- Session notes from completed work contain valuable lessons

### Archive Location

```
docs/archive/
├── BACKLOG_OLD.md         # Old ideas from 2024
├── REFACTORING_2024.md    # Historical refactoring notes
├── SESSION_EMOJI_FIX.md   # Completed session notes
└── RESEARCH_OLD_LIBS.md   # Deprecated library research
```

### Archive Format

Add header explaining why archived:

```markdown
# [ARCHIVED] Old Backlog Ideas

**Archived Date:** 2025-10-27
**Reason:** These ideas were superseded by new architecture decisions in v0.3.0

**Status:** Reference only - do not implement

---

## Original Content
...
```

---

## Documentation Health Check

Run this command to check documentation sizes:

```bash
wc -l CLAUDE.md README.md PLAN.md CHANGELOG.md BACKLOG.md HOTKEYS.md docs/*.md
```

**Healthy status:**
- CLAUDE.md: < 500 lines ✅
- PLAN.md: < 400 lines ✅
- CHANGELOG.md: < 350 lines ✅
- BACKLOG.md: < 300 lines ✅
- docs/* files: Focused and organized ✅

**Needs attention:**
- Any core file > 110% of limit ⚠️
- Multiple files exceeding limits 🚨
- NEXT_SESSION.md present after work completed ⚠️

---

## Summary

**Key Principle**: Documentation should flow through stages as it matures, with strict limits preventing bloat.

**Workflow**:
1. Idea → BACKLOG.md
2. Planning → PLAN.md
3. Implementation → docs/NEXT_SESSION.md
4. Completion → CHANGELOG.md
5. Reference → docs/*.md

**Maintenance**:
- Check sizes regularly
- Move content between files proactively
- Archive outdated information
- Keep core files under limits

This system ensures documentation remains useful for both humans and AI assistants!
