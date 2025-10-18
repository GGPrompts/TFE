# TFE Prompt Library Feature

## Vision

Transform TFE into a command center that combines file browsing with a prompt library system and tmux session integration. This turns TFE into a multi-monitor workflow hub where you can:
- Browse files and prompts in the same interface
- Preview and send prompts to running tmux sessions (like Claude Code)
- Organize prompts as files in `~/.prompts/` with version control
- Use template variables that auto-fill from context

## Architecture Philosophy

**Reuse, don't rebuild:**
- Prompts are just files (`.yaml`, `.md`, `.txt`) - TFE already browses files ✅
- Fuzzy search already works ✅
- Favorites system already exists ✅
- Dual-pane preview already exists ✅
- Tree view for organization already exists ✅

**Minimal additions:**
- Prompts filter (like favorites filter)
- Tmux integration module
- Enhanced dual-pane mode for prompts
- Template variable substitution

## User Workflow Example

```
1. Working in TFE browsing project files
2. Press F11 (or click 📝 button) → Activates "Prompt Mode"
3. Left pane: Shows only prompt files from ~/.prompts/
4. Right pane (top): Preview of selected prompt with variables filled
5. Right pane (middle): List of active tmux sessions/panes
6. Right pane (bottom): Send button and history
7. Press Enter → Sends prompt to selected tmux pane (e.g., Claude Code)
8. Press Esc → Exit prompt mode, return to normal file browsing
```

## UI Layout (Prompt Mode + Dual-Pane)

```
┌─────────────────────────────────────────────────────────────────┐
│ TFE - Terminal File Explorer [🏠] [✨] [>_] [📦] [🔍] [📝]       │
│ $ ~/.prompts/code-review                                        │
├──────────────────────────┬──────────────────────────────────────┤
│ PROMPTS (Left Pane)      │ PREVIEW (Top Right)                  │
│                          │ ┌────────────────────────────────┐   │
│ 📁 code-review/          │ │ Code Review Request            │   │
│   ▶ 📄 general.yaml      │ │ ──────────────────────────     │   │
│   📄 security.yaml       │ │ Please review main.go:         │   │
│   📄 performance.yaml    │ │ - Code quality                 │   │
│ 📁 debugging/            │ │ - Performance                  │   │
│   📄 trace-error.yaml    │ │ - Security concerns            │   │
│   📄 add-logging.yaml    │ │                                │   │
│ 📁 explain/              │ │ Focus: {{FOCUS_AREA}}          │   │
│   📄 architecture.yaml   │ └────────────────────────────────┘   │
│                          ├──────────────────────────────────────┤
│ F11: Exit Prompt Mode    │ SEND TO (Bottom Right)               │
│ Tab: Focus Pane          │ ┌────────────────────────────────┐   │
│ Ctrl+F: Search Prompts   │ │ > dev-server (tmux:0.1) ●      │   │
│                          │ │   claude-code (tmux:1.0)       │   │
│                          │ │   logs (tmux:1.1)              │   │
│                          │ │                                │   │
│                          │ │ [Enter] Send  [Esc] Cancel     │   │
│                          │ └────────────────────────────────┘   │
└──────────────────────────┴──────────────────────────────────────┘
```

## Implementation Plan

### Phase 1: Prompts Filter & UI
**Goal:** Add ability to filter/view only prompt files

- [ ] **1.1** Add `showPromptsOnly bool` field to `types.go` model struct
- [ ] **1.2** Add `isPromptFile()` helper function to `helpers.go`
  - Check for `.yaml`, `.md`, `.txt` extensions
- [ ] **1.3** Add `getFilteredPromptsFiles()` method (similar to `getFilteredFiles()`)
- [ ] **1.4** Add toolbar button `[📝]` in `view.go` and `render_preview.go`
  - Position: After `[🔍]` fuzzy search button
  - Toggle `showPromptsOnly` on click
  - Highlight when active: `✨📝`
- [ ] **1.5** Add keyboard shortcut `F11` to toggle prompt mode in `update_keyboard.go`
- [ ] **1.6** Add mouse click handler for toolbar button in `update_mouse.go`
  - Click region: X=25-29 (after search button)
- [ ] **1.7** Update status bar to show "• prompts only" indicator when active
- [ ] **1.8** Test: Toggle prompt mode, verify only `.yaml/.md/.txt` files shown

**Estimated Time:** 1-2 hours
**Files Modified:** `types.go`, `helpers.go`, `view.go`, `render_preview.go`, `update_keyboard.go`, `update_mouse.go`

---

### Phase 2: Tmux Integration Module
**Goal:** Detect and interact with tmux sessions

- [ ] **2.1** Create new file `tmux.go` (following TFE modular architecture)
- [ ] **2.2** Add tmux types to `types.go`:
  ```go
  type tmuxPane struct {
      sessionName string
      windowIndex int
      paneIndex   int
      title       string
      active      bool
  }
  ```
- [ ] **2.3** Implement `isTmuxAvailable() bool` - check if tmux is installed
- [ ] **2.4** Implement `listTmuxPanes() []tmuxPane`
  - Run: `tmux list-panes -a -F "#{session_name}:#{window_index}.#{pane_index}|#{pane_title}|#{pane_active}"`
  - Parse output into structs
- [ ] **2.5** Implement `sendToTmuxPane(pane tmuxPane, text string) error`
  - Run: `tmux send-keys -t session:window.pane "text" Enter`
- [ ] **2.6** Add model fields for tmux state:
  ```go
  tmuxPanes       []tmuxPane
  selectedPaneIdx int
  tmuxAvailable   bool
  ```
- [ ] **2.7** Add `refreshTmuxPanes()` method to update pane list
- [ ] **2.8** Test: Run `listTmuxPanes()`, verify detection of active sessions

**Estimated Time:** 2-3 hours
**Files Created:** `tmux.go`
**Files Modified:** `types.go`

---

### Phase 3: Enhanced Dual-Pane for Prompts
**Goal:** Split right pane into preview + tmux selector when in prompt mode

- [ ] **3.1** Create new rendering function `renderPromptDualPane()` in `render_preview.go`
- [ ] **3.2** Modify `View()` dispatcher in `view.go`:
  ```go
  if m.viewMode == viewDualPane {
      if m.showPromptsOnly {
          baseView = m.renderPromptDualPane()
      } else {
          baseView = m.renderDualPane()
      }
  }
  ```
- [ ] **3.3** Design three-section right pane layout:
  - **Top 60%:** Prompt preview (existing `renderPreview()`)
  - **Middle 25%:** Tmux session selector (new)
  - **Bottom 15%:** Send button/status (new)
- [ ] **3.4** Implement `renderTmuxSelector()` helper function
  - List tmux panes with scroll support
  - Highlight selected pane
  - Show active pane with `●` indicator
- [ ] **3.5** Add focus state: `tmuxSelectorFocused bool`
- [ ] **3.6** Add keyboard navigation for tmux selector:
  - Up/Down: Navigate panes
  - Enter: Select target pane
  - Tab: Cycle between preview and selector
- [ ] **3.7** Update `calculateLayout()` to handle three-section split
- [ ] **3.8** Test: Enter prompt mode + dual-pane, verify three sections render

**Estimated Time:** 3-4 hours
**Files Modified:** `render_preview.go`, `view.go`, `types.go`, `update_keyboard.go`, `model.go`

---

### Phase 4: Template Variables & Rendering
**Goal:** Parse and substitute variables in prompt files

- [ ] **4.1** Create new file `prompt_parser.go` (new module)
- [ ] **4.2** Add prompt type to `types.go`:
  ```go
  type promptTemplate struct {
      name        string
      description string
      variables   []string
      template    string
      raw         string
  }
  ```
- [ ] **4.3** Implement `parsePromptFile(path string) (*promptTemplate, error)`
  - Support YAML front matter (name, description, variables)
  - Support raw markdown/text files
  - Extract `{{VARIABLE}}` placeholders
- [ ] **4.4** Implement `renderPromptTemplate(tmpl *promptTemplate, vars map[string]string) string`
  - Replace `{{VAR}}` with values from map
  - Highlight missing variables in preview
- [ ] **4.5** Implement context variable providers:
  - `{{FILE}}` → Currently selected file path
  - `{{FILENAME}}` → File name only
  - `{{PROJECT}}` → Current directory name
  - `{{PATH}}` → Current full path
  - `{{DATE}}` → Current date (YYYY-MM-DD)
  - `{{TIME}}` → Current time (HH:MM)
- [ ] **4.6** Add `promptTemplate` to preview model
- [ ] **4.7** Modify `loadPreview()` to detect and parse prompt files
- [ ] **4.8** Update preview rendering to show rendered template
- [ ] **4.9** Test: Create `test.yaml` with `{{FILE}}`, verify substitution

**Estimated Time:** 2-3 hours
**Files Created:** `prompt_parser.go`
**Files Modified:** `types.go`, `file_operations.go`, `render_preview.go`

---

### Phase 5: Send Action & Integration
**Goal:** Wire up "Send to Tmux" functionality

- [ ] **5.1** Add `sendPromptToTmux()` method in `update_keyboard.go`
- [ ] **5.2** Handle Enter key in prompt mode:
  ```go
  if m.showPromptsOnly && m.viewMode == viewDualPane {
      // Send rendered prompt to selected tmux pane
      m.sendPromptToTmux()
  }
  ```
- [ ] **5.3** Implement send workflow:
  1. Get currently selected prompt file
  2. Parse and render template with variables
  3. Send to selected tmux pane
  4. Show success/error status message
- [ ] **5.4** Add send confirmation dialog (optional)
- [ ] **5.5** Add send history tracking:
  - Store last 10 sent prompts
  - Add `sendHistory []string` to model
  - Save to `~/.config/tfe/prompt_history.json`
- [ ] **5.6** Add clipboard fallback if tmux not available
  - Copy to clipboard with status message
- [ ] **5.7** Test: Send prompt to tmux pane, verify it appears
- [ ] **5.8** Test: Send with no tmux, verify clipboard fallback

**Estimated Time:** 2-3 hours
**Files Modified:** `update_keyboard.go`, `types.go`, `file_operations.go`

---

### Phase 6: Polish & Documentation
**Goal:** Refinement, error handling, and user documentation

- [ ] **6.1** Add error handling for tmux failures
- [ ] **6.2** Add loading spinner when refreshing tmux panes
- [ ] **6.3** Add help text to prompt mode (F1 in prompt mode)
- [ ] **6.4** Update `HOTKEYS.md` with prompt mode shortcuts
- [ ] **6.5** Create example prompt library in `docs/examples/prompts/`
  - `code-review/general.yaml`
  - `debugging/trace-error.yaml`
  - `explain/architecture.yaml`
- [ ] **6.6** Add README for prompt library: `docs/PROMPT_LIBRARY.md`
  - How to create prompts
  - Available variables
  - YAML format specification
  - Examples
- [ ] **6.7** Add demo GIF/screenshot to README
- [ ] **6.8** Test all features end-to-end
- [ ] **6.9** Update CHANGELOG.md
- [ ] **6.10** Update main README.md with prompt feature

**Estimated Time:** 2-3 hours
**Files Modified:** `HOTKEYS.md`, `README.md`, `CHANGELOG.md`
**Files Created:** `docs/PROMPT_LIBRARY.md`, `docs/examples/prompts/*`

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

### Keyboard Shortcuts (Prompt Mode)

| Key | Action |
|-----|--------|
| `F11` | Toggle prompt mode on/off |
| `Tab` | Cycle focus: file list → preview → tmux selector |
| `Enter` | Send prompt to selected tmux pane |
| `Esc` | Exit prompt mode / cancel send |
| `Ctrl+F` | Fuzzy search prompts |
| `↑/↓` | Navigate tmux pane list (when focused) |
| `1-4` | Switch display modes (works in prompt mode) |
| `F1` | Help (prompt mode specific help) |

### Mouse Actions (Prompt Mode)

| Click | Action |
|-------|--------|
| Toolbar `[📝]` | Toggle prompt mode |
| Prompt file | Select and preview |
| Tmux pane | Select target pane |
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

✅ **Feature is successful if:**
1. Can toggle prompt mode in < 2 keystrokes
2. Can send prompt to Claude Code in < 5 keystrokes
3. Template variables auto-fill correctly 90%+ of the time
4. Works smoothly across 2-3 monitor setups
5. Adds < 500 lines of code total (staying modular)
6. Users organize 20+ prompts easily
7. Integrates seamlessly with existing TFE workflow

---

## Notes & Decisions

### Why YAML over JSON?
- More human-readable
- Supports multi-line strings naturally
- Common in config files
- Easy to edit manually

### Why Tmux over other terminals?
- Universal on Linux/Mac
- Well-documented API (send-keys)
- Already common in dev workflows
- Session persistence across connections

### Why filter approach vs new mode?
- Reuses existing UI completely
- Minimal code changes
- Consistent with favorites filter
- User can still browse files in prompt mode

### Alternative considered: Separate TUI app
**Rejected because:**
- Would duplicate TFE's file browsing logic
- Breaks the "TFE as command center" vision
- More maintenance burden
- Less integration with existing workflow

---

## Questions for Review

- [ ] Should prompt mode auto-enter dual-pane? (Or keep as separate toggle?)
- [ ] Should we support custom variable prompts? (e.g., popup to fill `{{CUSTOM_INPUT}}`)
- [ ] Should send history be per-prompt or global?
- [ ] Should we support JSON prompts in addition to YAML/MD?
- [ ] Should there be a default prompts directory, or require config?

---

**Last Updated:** 2025-10-17
**Status:** Planning Phase
**Branch:** `prompts`
