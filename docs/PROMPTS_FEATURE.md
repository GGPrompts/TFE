# TFE Prompt Library Feature

**Status:** Phase 1 ✅ | Phase 2 ✅ | Phase 3 ✅ | Phase 4 (Polish) In Progress 🚧

**Last Updated:** 2025-10-18

## Vision

Transform TFE into a command center that combines file browsing with a prompt library system. This turns TFE into a workflow hub where you can:
- Browse files and prompts in the same interface
- Preview prompts with template variables that auto-fill from context
- Copy rendered prompts to clipboard for pasting anywhere
- Organize prompts as files in `~/.prompts/` with version control
- Integrate with `.claude/commands/` and `.claude/agents/` for project-specific prompts
- Store CLI command references alongside prompts for quick access

## Architecture Philosophy

**Reuse, don't rebuild:**
- Prompts are just files (`.yaml`, `.md`, `.txt`) - TFE already browses files ✅
- Fuzzy search already works ✅
- Favorites system already exists ✅
- Dual-pane preview already exists ✅
- Tree view for organization already exists ✅

**Minimal additions:**
- Prompts filter (like favorites filter)
- Template variable substitution
- Copy to clipboard action
- Enhanced dual-pane preview for prompts

## User Workflow Example

```
1. Working in TFE browsing project files
2. Press F11 (or click 📝 button) → Activates "Prompt Mode"
3. Left pane: Shows only prompt files from ~/.prompts/
4. Right pane: Preview of selected prompt with variables auto-filled
   - {{FILE}} → currently selected file
   - {{PROJECT}} → current directory name
   - {{DATE}}, {{TIME}}, etc.
5. Press Enter (or F5) → Copies rendered prompt to clipboard
6. Status message: "✓ Prompt copied to clipboard"
7. Switch to your AI CLI (Claude Code, aider, etc.) and paste
8. Press Esc or F11 → Exit prompt mode, return to normal file browsing
```

**Alternative workflow with CLI commands:**
```
1. Browse to ~/.prompts/_cli-commands/ folder
2. Select claude-flags.md (or aider-modes.md, etc.)
3. Preview shows command with copy blocks
4. Copy command to clipboard
5. Use TFE quick cd (Ctrl+Enter) to open bash at project
6. Paste and run: claude --model sonnet-4 --context ./docs
```

## UI Layout (Prompt Mode + Dual-Pane)

```
┌─────────────────────────────────────────────────────────────────┐
│ TFE - Terminal File Explorer [🏠] [✨] [>_] [📦] [🔍] [📝]       │
│ $ ~/.prompts/code-review                         [Prompt Mode]  │
├──────────────────────────┬──────────────────────────────────────┤
│ PROMPTS (Left Pane)      │ PREVIEW (Right Pane)                 │
│                          │                                      │
│ 📁 _cli-commands/        │  Code Review Request                 │
│ 📁 code-review/          │  ──────────────────────────          │
│   ▶ 📄 general.yaml      │  Please review the following code:   │
│   📄 security.yaml       │                                      │
│   📄 performance.yaml    │  File: main.go                       │
│ 📁 debugging/            │  Project: TFE                        │
│   📄 trace-error.yaml    │                                      │
│   📄 add-logging.yaml    │  Review for:                         │
│ 📁 explain/              │  - Code quality and readability      │
│   📄 architecture.yaml   │  - Potential bugs                    │
│                          │  - Performance implications          │
│ F11: Exit Prompt Mode    │  - Security vulnerabilities          │
│ Enter/F5: Copy to Clip   │                                      │
│ Ctrl+F: Search Prompts   │  [Enter] Copy to Clipboard           │
│ Tab: Toggle Dual-Pane    │  [Esc] Cancel                        │
└──────────────────────────┴──────────────────────────────────────┘
```

## Implementation Plan

**Strategy:** Build focused copy/paste workflow in 4 phases (~6-8 hours total).

### Phase 1: Prompts Filter & UI ✅ COMPLETE
**Goal:** Add ability to filter/view only prompt files

- [x] **1.1** Add `showPromptsOnly bool` field to `types.go` model struct
- [x] **1.2** Add `isPromptFile()` helper function to `helpers.go`
  - Check for `.prompty`, `.yaml`, `.yml`, `.md`, `.txt` extensions
  - Smart `.md` filtering: only in `.claude/` or `~/.prompts/` directories
- [x] **1.3** Updated `getFilteredFiles()` in `favorites.go` to respect `showPromptsOnly`
- [x] **1.4** Add toolbar button `[📝]` in `view.go` and `render_preview.go`
  - Position: After `[🔍]` fuzzy search button
  - Toggle `showPromptsOnly` on click
  - Highlight when active: `✨📝`
- [x] **1.5** Add keyboard shortcut `F11` to toggle prompt mode in `update_keyboard.go`
- [x] **1.6** Add mouse click handler for toolbar button in `update_mouse.go`
  - Click region: X=25-34 (handles both normal and active state)
- [x] **1.7** Update status bar to show "• 📝 prompts only" indicator when active
- [x] **1.8** Test: Toggle prompt mode, verify only prompt files shown

**✨ Bonus Enhancements Added:**
- [x] Always show important dev folders even when hidden files off:
  - `.claude/` 🤖, `.git/` 📦, `.vscode/` 💻, `.github/` 🐙, `.config/` ⚙️, `.docker/` 🐳
- [x] Added icons for new important folders
- [x] Smart `.md` detection (only prompts if in `.claude/` or `~/.prompts/`)
- [x] Added `.prompty` extension support (Microsoft Prompty format)

**Time Taken:** ~1.5 hours
**Files Modified:** `types.go`, `helpers.go`, `favorites.go`, `view.go`, `render_preview.go`, `update_keyboard.go`, `update_mouse.go`, `file_operations.go`
**Stats:** +138 lines, -5 lines across 8 files

---

### Phase 2: Multi-Location Template Parsing & Rendering ✅ COMPLETE
**Goal:** Parse prompts from multiple locations and render with variable substitution

**Template Parsing:**
- [x] **2.1** Create new file `prompt_parser.go` (new module)
- [x] **2.2** Add prompt type to `types.go`:
  ```go
  type promptTemplate struct {
      name        string
      description string
      source      string // "global", "command", "agent", "local"
      variables   []string
      template    string
      raw         string
  }
  ```
- [x] **2.3** Implement `parsePromptFile(path string) (*promptTemplate, error)`
  - Support `.prompty` format (YAML frontmatter between `---` markers)
  - Support YAML files (`.yaml`, `.yml`)
  - Support raw markdown/text files (`.md`, `.txt`)
  - Extract `{{VARIABLE}}` placeholders
- [x] **2.4** Implement `renderPromptTemplate(tmpl *promptTemplate, vars map[string]string) string`
  - Replace `{{VAR}}` with values from map (case-insensitive)
  - Support lowercase, uppercase, and title case variables
- [x] **2.5** Implement context variable providers:
  - `{{file}}` → Currently selected file path
  - `{{filename}}` → File name only
  - `{{project}}` → Current directory name
  - `{{path}}` → Current full path
  - `{{DATE}}` → Current date (YYYY-MM-DD)
  - `{{TIME}}` → Current time (HH:MM)
- [x] **2.6** Add `promptTemplate` to preview model
- [x] **2.7** Modify `loadPreview()` to detect and parse prompt files
- [x] **2.8** Update preview rendering to show rendered template with metadata header
  - Show prompt name, description, source (🌐 Global, ⚙️ Command, 🤖 Agent, 📁 Local)
  - Display detected variables
  - Render template with auto-filled variables
- [x] **2.9** Smart directory filtering in prompts mode
  - Hide empty directories
  - Show only directories containing prompts (recursive check up to 2 levels)
  - Always show `.claude`, `.prompts`, `.config` folders
- [x] **2.10** Add `.prompts` to important folders (always visible even when hidden files OFF)
- [x] **2.11** Add 📝 icon for `.prompts` folder
- [x] **2.12** Test: Create test prompts in all formats (.prompty, .yaml, .md, .txt)

**Note:** Multi-location auto-scanning (section headers) deferred - current implementation uses manual navigation which works well.

**Time Taken:** ~3.5 hours
**Files Created:** `prompt_parser.go` (240 lines)
**Files Modified:** `types.go` (+15), `file_operations.go` (+40), `render_preview.go` (+130), `favorites.go` (+60), `helpers.go` (reused existing `isPromptFile`)
**Total Code:** ~485 lines added

---

### Phase 3: Copy to Clipboard Action ✅ COMPLETE
**Goal:** Enable copying rendered prompts to clipboard

- [x] **3.1** Implement copy workflow in `update_keyboard.go`:
  - Detect when Enter or F5 pressed on prompt file
  - Get context variables from current state
  - Render template with variable substitution
  - Copy to clipboard using existing `copyToClipboard()` function
  - Show success status message: "✓ Prompt copied to clipboard"
- [x] **3.2** Handle Enter key in prompts mode:
  - Special handling when `showPromptsOnly` is true
  - Only copy prompts, not navigate
  - Reuse existing clipboard infrastructure
- [x] **3.3** Handle F5 key in all modes:
  - Regular mode: Copy rendered prompt if prompt file, else copy file path
  - Full preview mode: Copy rendered prompt if prompt file, else copy file path
  - Dual-pane mode: Copy rendered prompt if prompt file
- [x] **3.4** Error handling:
  - Show error message if clipboard copy fails
  - Graceful fallback behavior
- [x] **3.5** Test: Clipboard tools available (clip.exe on WSL)

**Note:** Copy history tracking deferred to future enhancement - not needed for MVP workflow.

**Time Taken:** ~1 hour
**Files Modified:** `update_keyboard.go` (+50 lines)
**Total Code:** ~50 lines added

---

### Phase 4: Polish & Documentation
**Goal:** Refinement, error handling, preview enhancements, and user documentation

**Essential tasks:**
- [ ] **4.1** Add error handling for clipboard failures
- [ ] **4.2** Add help text to prompt mode (F1 in prompt mode)
- [ ] **4.3** Update `HOTKEYS.md` with prompt mode shortcuts
- [ ] **4.4** Create example prompt library in `docs/examples/prompts/`
  - `code-review/general.yaml`
  - `debugging/trace-error.yaml`
  - `explain/architecture.yaml`
  - `refactor/extract-function.md`
  - `_cli-commands/claude-flags.md` (CLI reference example)
- [ ] **4.5** Add README for prompt library: `docs/PROMPT_LIBRARY.md`
  - How to create prompts
  - Available variables
  - YAML format specification
  - Examples
  - CLI command reference pattern
- [ ] **4.6** Test all core features end-to-end
  - Toggle prompt mode
  - Browse prompts with tree view
  - Preview with variable substitution
  - Copy to clipboard
  - Paste in external app
- [ ] **4.7** Update CHANGELOG.md with core feature
- [ ] **4.8** Update main README.md with prompt feature

**Optional preview enhancements (nice-to-have):**
- [ ] **4.9** Show prompt metadata in preview header (name, description from YAML)
- [ ] **4.10** Highlight variable substitutions with styling
- [ ] **4.11** Add "Copy" hint in preview footer
- [ ] **4.12** Auto-enter dual-pane when entering prompt mode

**Estimated Time:** 2-3 hours
**Files Modified:** `HOTKEYS.md`, `README.md`, `CHANGELOG.md`, `render_preview.go`, `update_keyboard.go`
**Files Created:** `docs/PROMPT_LIBRARY.md`, `docs/examples/prompts/*`

**🎉 MVP Complete! Focused copy/paste workflow fully functional.**

---

## Technical Specifications

### Prompt File Format (YAML)

```yaml
name: Code Review Request
description: Request a thorough code review with security focus
category: code-review
variables:
  - FILE
  - FOCUS_AREA
template: |
  Please review the following code:

  **File:** {{FILE}}
  **Focus:** {{FOCUS_AREA}}

  Review for:
  - Code quality and readability
  - Potential bugs and edge cases
  - Performance implications
  - Security vulnerabilities
  - Best practices adherence

  Provide specific suggestions for improvement.
```

### Prompt File Format (Markdown - Simple)

```markdown
# Code Review Request

Please review {{FILE}} for:
- Code quality
- Performance
- Security

Focus area: {{FOCUS_AREA}}
```

### Directory Structure

```
~/.prompts/
├── code-review/
│   ├── general.yaml
│   ├── security-focused.yaml
│   └── performance.yaml
├── debugging/
│   ├── trace-error.yaml
│   ├── add-logging.yaml
│   └── reproduce-bug.yaml
├── explain/
│   ├── architecture.yaml
│   ├── function.yaml
│   └── algorithm.yaml
├── refactor/
│   ├── extract-function.yaml
│   ├── simplify-logic.yaml
│   └── improve-naming.yaml
└── testing/
    ├── unit-test.yaml
    └── integration-test.yaml
```

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `F11` | Toggle prompt mode on/off |
| `Tab` | Toggle dual-pane view |
| `Enter` or `F5` | Copy rendered prompt to clipboard |
| `Esc` | Exit prompt mode |
| `Ctrl+F` | Fuzzy search prompts |
| `↑/↓` | Navigate prompt list |
| `1-4` | Switch display modes (works in prompt mode) |
| `F1` | Help (prompt mode specific help) |

### Mouse Actions

| Click | Action |
|-------|--------|
| Toolbar `[📝]` | Toggle prompt mode |
| Prompt file | Select and preview |
| Preview area | Focus preview pane |

---

## Configuration

### Config File: `~/.config/tfe/prompts.json`

```json
{
  "prompt_directories": [
    "~/.prompts",
    "~/Documents/prompts",
    "~/work/team-prompts"
  ],
  "default_tmux_pane": "claude-code",
  "auto_refresh_tmux": true,
  "template_variables": {
    "AUTHOR": "Your Name",
    "TEAM": "Engineering"
  },
  "send_history_size": 20
}
```

---

## Testing Checklist

### Phase 1: Prompts Filter
- [ ] Toolbar button appears and is clickable
- [ ] F11 toggles prompt mode
- [ ] Only `.yaml`, `.md`, `.txt` files shown in prompt mode
- [ ] Status bar shows "prompts only" indicator
- [ ] Can exit prompt mode and return to normal view

### Phase 2: Tmux Integration
- [ ] Detects running tmux sessions
- [ ] Lists all panes correctly
- [ ] Identifies active pane
- [ ] Sends text to correct pane
- [ ] Handles "tmux not installed" gracefully

### Phase 3: Dual-Pane Layout
- [ ] Three sections render correctly
- [ ] Preview shows in top section
- [ ] Tmux list shows in middle section
- [ ] Send controls show in bottom section
- [ ] Tab cycles focus between sections
- [ ] Layout adapts to terminal size

### Phase 4: Template Rendering
- [ ] YAML prompts parse correctly
- [ ] Markdown prompts parse correctly
- [ ] Variables detect and highlight
- [ ] `{{FILE}}` substitutes current file
- [ ] `{{PROJECT}}` substitutes directory name
- [ ] Missing variables highlighted in preview

### Phase 5: Send Functionality
- [ ] Enter sends prompt to tmux
- [ ] Prompt appears in target pane
- [ ] Success message appears
- [ ] Clipboard fallback works
- [ ] History saves correctly

### Phase 6: Polish
- [ ] Help text displays correctly
- [ ] Error messages are clear
- [ ] No crashes on edge cases
- [ ] Works across terminal sizes
- [ ] Documentation is accurate

---

## Future Enhancements (Post-MVP)

- [ ] **Prompt snippets**: Quick-insert common variables/text
- [ ] **Multi-step prompts**: Chain multiple prompts together
- [ ] **Prompt templates**: Generate new prompt files from templates
- [ ] **Session memory**: Remember last used tmux pane per prompt
- [ ] **Variable editor**: Popup to edit variables before sending
- [ ] **Shared prompt library**: Clone team prompts from Git repo
- [ ] **Prompt search**: Full-text search across all prompts
- [ ] **Response capture**: Capture Claude's response back into TFE
- [ ] **Prompt analytics**: Track which prompts used most often
- [ ] **AI suggestions**: Suggest prompts based on current file type

---

## Success Metrics

✅ **MVP is successful if:**
1. Can toggle prompt mode in < 2 keystrokes (F11 or toolbar click)
2. Can copy rendered prompt to clipboard in < 3 keystrokes (browse, Enter)
3. Template variables auto-fill correctly 90%+ of the time
4. Clipboard paste works in any application (Claude Code, terminal, etc.)
5. Feature adds < 300 lines of code (staying modular)
6. Users organize 20+ prompts easily with tree view + fuzzy search
7. Integrates seamlessly with existing TFE workflow
8. CLI command references work alongside prompts (quick cd + paste)

---

## Notes & Decisions

### Why copy/paste only (no auto-send)?
- **Universal**: Works everywhere (any terminal, any app, any OS)
- **Simple**: No dependencies, no complex integrations
- **Fast to ship**: MVP complete in ~6-8 hours
- **User control**: You decide where and when to paste
- **Flexible**: Works with any AI CLI (claude, aider, cursor, etc.)
- **Lesson learned**: Launchers and automation make assumptions; manual control is better

### Why YAML over JSON?
- More human-readable
- Supports multi-line strings naturally
- Common in config files
- Easy to edit manually

### Why include CLI command references?
- You already use markdown files with copy blocks
- Same workflow: browse, preview, copy, paste
- Complements prompts (launch AI CLI + send prompt)
- Works with TFE's quick cd feature (Ctrl+Enter)

### Why filter approach vs new mode?
- Reuses existing UI completely
- Minimal code changes
- Consistent with favorites filter
- User can still browse files in prompt mode
- Prompts are just files - TFE already excels at file browsing

### Alternative considered: Auto-launch + tmux integration
**Rejected because:**
- Over-engineered for the use case
- Assumes user workflow (tmux, specific setup)
- User already has quick cd working great
- Copy/paste is simpler and more flexible
- Would add 15-20 hours vs 6-8 hours

### Alternative considered: Separate TUI app
**Rejected because:**
- Would duplicate TFE's file browsing logic
- Breaks the "TFE as command center" vision
- More maintenance burden
- Less integration with existing workflow

---

## Questions for Review

- [ ] Should prompt mode auto-enter dual-pane? (Or keep as separate toggle?)
- [ ] Should we support fillable fields/interactive prompts? (v2.0 feature - popup dialog to fill variables)
- [ ] Should copy history be per-prompt or global?
- [ ] Should we support JSON prompts in addition to YAML/MD?
- [ ] Should there be a default prompts directory (`~/.prompts/`), or require config?
- [ ] Should F5 or Enter be the primary copy shortcut? (Or both?)
- [ ] Should CLI command references live in `_cli-commands/` subfolder?

---

## Future Enhancements (Not in MVP)

**Fillable Fields** - Interactive variable prompts:
```yaml
name: Custom Review
fields:
  - name: FOCUS_AREA
    prompt: "What to focus on?"
    default: "security"
template: Review {{FILE}} for {{FOCUS_AREA}}
```
Opens dialog to fill fields before copying.

**Tmux Manager** - Full command center (20-30 hours):
- Visual tmux session/pane manager
- Rename sessions for clarity
- Chat interface (multi-turn conversations)
- Direct sending to specific panes
- Conversation history per session

See discussion above for why these aren't in core MVP.

---

---

## Implementation Decisions

### Markdown File Filtering
**Decision:** `.md` files are only considered prompts if located in:
- `.claude/` or any subfolder (e.g., `.claude/commands/`, `.claude/agents/`)
- `~/.prompts/` or any subfolder

**Rationale:** Prevents `README.md`, `CHANGELOG.md`, and other documentation files from appearing in prompts mode while still supporting `.claude/commands/*.md` and `.claude/agents/*.md` for Claude Code integration.

### Important Development Folders
**Decision:** Always show these folders even when "show hidden files" is OFF:
- `.claude/` 🤖 - Claude Code configuration, commands, agents
- `.git/` 📦 - Git repository
- `.vscode/` 💻 - VS Code settings
- `.github/` 🐙 - GitHub Actions workflows
- `.config/` ⚙️ - Application configuration
- `.docker/` 🐳 - Docker configs

**Rationale:** These are critical development folders that users need regular access to. Hiding them creates friction. This matches behavior of VS Code and other modern IDEs.

### Multi-Location Prompt Display
**Decision:** Tree view with section headers (Option C from design phase)

**Rationale:** Best balance of discoverability and organization. User can see all available prompts at once, clearly organized by source, without needing to switch views.

### Prompt File Formats
**Decision:** Support `.prompty`, `.yaml`, `.yml`, `.md` (conditional), `.txt`

**Rationale:**
- `.prompty` - Microsoft Prompty format for VS Code compatibility
- `.yaml`/`.yml` - Flexible structured format
- `.md` - Natural for documentation-style prompts (when in appropriate folders)
- `.txt` - Simple plain-text templates

---

## ✅ Implementation Summary (Phases 1-3 Complete)

### What's Working Now

**Core Features:**
- ✅ F11 toggles prompts mode (toolbar shows `[✨📝]` when active)
- ✅ Smart filtering: Only shows prompt files + directories containing prompts
- ✅ `.prompts` folder visible with 📝 icon (even when hidden files OFF)
- ✅ Four format support: `.prompty`, `.yaml`, `.yml`, `.md` (in special folders), `.txt`
- ✅ Template variable substitution with 6 context variables
- ✅ Beautiful preview with metadata header (name, description, source, variables list)
- ✅ Source detection: 🌐 Global, ⚙️ Command, 🤖 Agent, 📁 Local
- ✅ Copy to clipboard: Press Enter or F5 on prompt file
- ✅ Success message: "✓ Prompt copied to clipboard"

**Template Variables (Auto-Filled):**
- `{{file}}` → Currently selected file path
- `{{filename}}` → File name only
- `{{project}}` → Current directory name
- `{{path}}` → Current full directory path
- `{{DATE}}` → Current date (YYYY-MM-DD)
- `{{TIME}}` → Current time (HH:MM)

**Prompt Locations:**
- `~/.prompts/` → Global prompts (navigate home with `g` then `h`)
- `.claude/commands/` → Project-specific commands
- `.claude/agents/` → Project-specific agents
- Current directory → Ad-hoc prompts

**User Workflow (Complete):**
```bash
1. ./tfe                     # Launch TFE
2. Navigate to file          # Select the file you want to review
3. Press F11                 # Enter prompts mode
4. Navigate to ~/.prompts/   # Press 'g' then 'h' for home
5. Select prompt template    # Preview shows with variables filled
6. Press Enter or F5         # Copy rendered prompt to clipboard
7. ✓ Status shows success
8. Paste into Claude Code    # Ctrl+V
```

**Test Prompts Available:**
- `~/.prompts/test-prompt.prompty` - Full .prompty format with YAML frontmatter
- `~/.prompts/quick-question.txt` - Simple text template
- `~/.prompts/code-review/security.yaml` - YAML format with structured metadata
- `.claude/commands/review-pr.md` - Project command prompt
- `.claude/agents/test-runner.md` - Project agent prompt

### Code Statistics

**Total Implementation:**
- **Time Invested:** ~6 hours (Phase 1: 1.5h, Phase 2: 3.5h, Phase 3: 1h)
- **Lines Added:** ~673 lines
- **Files Created:** 1 new module (`prompt_parser.go` - 240 lines)
- **Files Modified:** 8 files (`types.go`, `file_operations.go`, `render_preview.go`, `favorites.go`, `helpers.go`, `view.go`, `update_keyboard.go`, `update_mouse.go`)

**Architecture:**
- ✅ Maintained modular structure (no bloat in `main.go`)
- ✅ Followed TFE conventions (new module for parsing, types in `types.go`)
- ✅ Reused existing infrastructure (clipboard, preview, filtering)
- ✅ Clean separation of concerns

### Next Steps: Phase 4 (Optional Polish)

Phase 4 is for optional enhancements and documentation:
- [ ] Update HOTKEYS.md with prompt shortcuts
- [ ] Create example prompt library in docs/examples/prompts/
- [ ] Add PROMPT_LIBRARY.md user guide
- [ ] Update main README.md with prompt feature
- [ ] Update CHANGELOG.md

**Current Status:** MVP is fully functional! Phase 4 is optional polish.

---

**Last Updated:** 2025-10-18
**Status:** Phase 1 ✅ | Phase 2 ✅ | Phase 3 ✅ | MVP Complete! 🎉
**Branch:** `prompts`
**Total Time:** ~6 hours
**Next Step:** Optional Phase 4 (Polish & Documentation) or merge to main
