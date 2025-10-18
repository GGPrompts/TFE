# TFE Prompt Library Feature

## Vision

Transform TFE into a command center that combines file browsing with a prompt library system. This turns TFE into a workflow hub where you can:
- Browse files and prompts in the same interface
- Preview prompts with template variables that auto-fill from context
- Copy rendered prompts to clipboard for pasting anywhere
- Organize prompts as files in `~/.prompts/` with version control
- **(Optional)** Send prompts directly to tmux sessions (like Claude Code)

## Architecture Philosophy

**Reuse, don't rebuild:**
- Prompts are just files (`.yaml`, `.md`, `.txt`) - TFE already browses files ✅
- Fuzzy search already works ✅
- Favorites system already exists ✅
- Dual-pane preview already exists ✅
- Tree view for organization already exists ✅

**Minimal additions (Core MVP):**
- Prompts filter (like favorites filter)
- Template variable substitution
- Copy to clipboard action
- Enhanced dual-pane preview for prompts

**Optional enhancement:**
- Tmux integration module for direct sending

## User Workflow Example (Core MVP)

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
7. Paste anywhere (Claude Code, terminal, editor, etc.)
8. Press Esc or F11 → Exit prompt mode, return to normal file browsing
```

## User Workflow Example (With Tmux Enhancement)

```
1-4. Same as above (browse prompts, preview with variables)
5. Right pane shows tmux session selector (optional bottom panel)
6. Select target tmux pane (e.g., "claude-code")
7. Press Ctrl+Enter → Sends prompt directly to tmux pane
8. OR press Enter → Copies to clipboard (default)
```

## UI Layout (Core MVP - Prompt Mode + Dual-Pane)

```
┌─────────────────────────────────────────────────────────────────┐
│ TFE - Terminal File Explorer [🏠] [✨] [>_] [📦] [🔍] [📝]       │
│ $ ~/.prompts/code-review                         [Prompt Mode]  │
├──────────────────────────┬──────────────────────────────────────┤
│ PROMPTS (Left Pane)      │ PREVIEW (Right Pane)                 │
│                          │                                      │
│ 📁 code-review/          │  Code Review Request                 │
│   ▶ 📄 general.yaml      │  ──────────────────────────          │
│   📄 security.yaml       │  Please review the following code:   │
│   📄 performance.yaml    │                                      │
│ 📁 debugging/            │  File: main.go                       │
│   📄 trace-error.yaml    │  Project: TFE                        │
│   📄 add-logging.yaml    │                                      │
│ 📁 explain/              │  Review for:                         │
│   📄 architecture.yaml   │  - Code quality and readability      │
│                          │  - Potential bugs                    │
│ F11: Exit Prompt Mode    │  - Performance implications          │
│ Enter/F5: Copy to Clip   │  - Security vulnerabilities          │
│ Ctrl+F: Search Prompts   │                                      │
│ Tab: Toggle Dual-Pane    │  [Enter] Copy to Clipboard           │
│                          │  [Esc] Cancel                        │
└──────────────────────────┴──────────────────────────────────────┘
```

## UI Layout (With Tmux Enhancement - Optional)

```
┌─────────────────────────────────────────────────────────────────┐
│ TFE - Terminal File Explorer [🏠] [✨] [>_] [📦] [🔍] [📝]       │
│ $ ~/.prompts/code-review                         [Prompt Mode]  │
├──────────────────────────┬──────────────────────────────────────┤
│ PROMPTS (Left Pane)      │ PREVIEW (Top Right)                  │
│                          │ ┌────────────────────────────────┐   │
│ 📁 code-review/          │ │ Code Review Request            │   │
│   ▶ 📄 general.yaml      │ │ ──────────────────────────     │   │
│   📄 security.yaml       │ │ Please review main.go:         │   │
│   📄 performance.yaml    │ │ - Code quality                 │   │
│ 📁 debugging/            │ │ - Performance                  │   │
│   📄 trace-error.yaml    │ │ - Security concerns            │   │
│                          │ └────────────────────────────────┘   │
│                          ├──────────────────────────────────────┤
│ F11: Exit Prompt Mode    │ SEND TO (Bottom Right)               │
│ Enter: Copy to Clipboard │ ┌────────────────────────────────┐   │
│ Ctrl+Enter: Send to Tmux │ │ > claude-code (tmux:1.0) ●     │   │
│ Ctrl+F: Search Prompts   │ │   dev-server (tmux:0.1)        │   │
│                          │ │   logs (tmux:1.1)              │   │
│                          │ │                                │   │
│                          │ │ [↑/↓] Select  [Ctrl+Enter] Send│   │
│                          │ └────────────────────────────────┘   │
└──────────────────────────┴──────────────────────────────────────┘
```

## Implementation Plan

**Strategy:** Build core copy/paste workflow first (Phases 1-5), add tmux integration as optional enhancement (Phase 6).

### Phase 1: Prompts Filter & UI (Core MVP)
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

### Phase 2: Template Variables & Rendering (Core MVP)
**Goal:** Parse and substitute variables in prompt files

- [ ] **2.1** Create new file `prompt_parser.go` (new module)
- [ ] **2.2** Add prompt type to `types.go`:
  ```go
  type promptTemplate struct {
      name        string
      description string
      variables   []string
      template    string
      raw         string
  }
  ```
- [ ] **2.3** Implement `parsePromptFile(path string) (*promptTemplate, error)`
  - Support YAML front matter (name, description, variables)
  - Support raw markdown/text files
  - Extract `{{VARIABLE}}` placeholders
- [ ] **2.4** Implement `renderPromptTemplate(tmpl *promptTemplate, vars map[string]string) string`
  - Replace `{{VAR}}` with values from map
  - Highlight missing variables in preview
- [ ] **2.5** Implement context variable providers:
  - `{{FILE}}` → Currently selected file path
  - `{{FILENAME}}` → File name only
  - `{{PROJECT}}` → Current directory name
  - `{{PATH}}` → Current full path
  - `{{DATE}}` → Current date (YYYY-MM-DD)
  - `{{TIME}}` → Current time (HH:MM)
- [ ] **2.6** Add `promptTemplate` to preview model
- [ ] **2.7** Modify `loadPreview()` to detect and parse prompt files
- [ ] **2.8** Update preview rendering to show rendered template
- [ ] **2.9** Test: Create `test.yaml` with `{{FILE}}`, verify substitution

**Estimated Time:** 2-3 hours
**Files Created:** `prompt_parser.go`
**Files Modified:** `types.go`, `file_operations.go`, `render_preview.go`

---

### Phase 3: Copy to Clipboard Action (Core MVP)
**Goal:** Enable copying rendered prompts to clipboard

- [ ] **3.1** Add `copyPromptToClipboard()` method in `update_keyboard.go`
- [ ] **3.2** Handle Enter key in prompt mode:
  ```go
  if m.showPromptsOnly {
      // Copy rendered prompt to clipboard
      m.copyPromptToClipboard()
  }
  ```
- [ ] **3.3** Implement copy workflow:
  1. Get currently selected prompt file
  2. Parse and render template with variables
  3. Copy to clipboard using existing `copyToClipboard()` function
  4. Show success status message: "✓ Prompt copied to clipboard"
- [ ] **3.4** Add F5 as alternative copy shortcut (consistent with path copy)
- [ ] **3.5** Add copy history tracking:
  - Store last 10 copied prompts
  - Add `promptHistory []string` to model
  - Save to `~/.config/tfe/prompt_history.json`
- [ ] **3.6** Test: Press Enter on prompt, verify clipboard contains rendered text
- [ ] **3.7** Test: Paste in external app, verify variables were substituted

**Estimated Time:** 1-2 hours
**Files Modified:** `update_keyboard.go`, `types.go`, `file_operations.go`

---

### Phase 4: Enhanced Dual-Pane for Prompts (Core MVP)
**Goal:** Improve prompt preview display in dual-pane mode

- [ ] **4.1** Add special handling for prompt files in dual-pane preview
- [ ] **4.2** Show prompt metadata in preview header:
  - Prompt name (from YAML front matter)
  - Description
  - Required variables list
- [ ] **4.3** Highlight variable substitutions in preview with styling
  - Show `{{VAR}}` in one color if missing value
  - Show substituted text in another color
- [ ] **4.4** Add "Copy" button hint in preview footer
- [ ] **4.5** Auto-enter dual-pane when entering prompt mode (optional UX enhancement)
- [ ] **4.6** Test: Open prompt in dual-pane, verify enhanced preview
- [ ] **4.7** Test: Variables highlighted correctly

**Estimated Time:** 1-2 hours
**Files Modified:** `render_preview.go`, `update_keyboard.go`

---

### Phase 5: Polish & Documentation (Core MVP Complete)
**Goal:** Refinement, error handling, and user documentation for core feature

- [ ] **5.1** Add error handling for clipboard failures
- [ ] **5.2** Add help text to prompt mode (F1 in prompt mode)
- [ ] **5.3** Update `HOTKEYS.md` with prompt mode shortcuts
- [ ] **5.4** Create example prompt library in `docs/examples/prompts/`
  - `code-review/general.yaml`
  - `debugging/trace-error.yaml`
  - `explain/architecture.yaml`
  - `refactor/extract-function.md`
- [ ] **5.5** Add README for prompt library: `docs/PROMPT_LIBRARY.md`
  - How to create prompts
  - Available variables
  - YAML format specification
  - Examples
- [ ] **5.6** Add demo GIF/screenshot to README
- [ ] **5.7** Test all core features end-to-end
  - Toggle prompt mode
  - Browse prompts with tree view
  - Preview with variable substitution
  - Copy to clipboard
  - Paste in external app
- [ ] **5.8** Update CHANGELOG.md with core feature
- [ ] **5.9** Update main README.md with prompt feature

**Estimated Time:** 2-3 hours
**Files Modified:** `HOTKEYS.md`, `README.md`, `CHANGELOG.md`
**Files Created:** `docs/PROMPT_LIBRARY.md`, `docs/examples/prompts/*`

**🎉 Core MVP Complete! Copy/paste workflow fully functional.**

---

### Phase 6: Tmux Integration (OPTIONAL Enhancement)
**Goal:** Add direct sending to tmux sessions for advanced users

- [ ] **6.1** Create new file `tmux.go` (following TFE modular architecture)
- [ ] **6.2** Add tmux types to `types.go`:
  ```go
  type tmuxPane struct {
      sessionName string
      windowIndex int
      paneIndex   int
      title       string
      active      bool
  }
  ```
- [ ] **6.3** Implement `isTmuxAvailable() bool` - check if tmux is installed
- [ ] **6.4** Implement `listTmuxPanes() []tmuxPane`
  - Run: `tmux list-panes -a -F "#{session_name}:#{window_index}.#{pane_index}|#{pane_title}|#{pane_active}"`
  - Parse output into structs
- [ ] **6.5** Implement `sendToTmuxPane(pane tmuxPane, text string) error`
  - Run: `tmux send-keys -t session:window.pane "text" Enter`
- [ ] **6.6** Add model fields for tmux state
- [ ] **6.7** Create `renderPromptDualPaneWithTmux()` - three-section layout
  - Top 60%: Prompt preview
  - Middle 25%: Tmux session selector
  - Bottom 15%: Send controls
- [ ] **6.8** Add keyboard shortcuts:
  - `Ctrl+Enter`: Send to selected tmux pane
  - `Enter`: Copy to clipboard (default, unchanged)
  - Up/Down in tmux selector: Navigate panes
- [ ] **6.9** Add tmux pane selection UI with highlight
- [ ] **6.10** Test: Send prompt to tmux, verify it appears
- [ ] **6.11** Update documentation for tmux feature

**Estimated Time:** 3-4 hours
**Files Created:** `tmux.go`
**Files Modified:** `types.go`, `render_preview.go`, `update_keyboard.go`, `view.go`

**Note:** This phase is completely optional. Core feature works perfectly without tmux.

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

### Keyboard Shortcuts (Core MVP)

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

### Keyboard Shortcuts (With Tmux Enhancement)

| Key | Action |
|-----|--------|
| `Enter` or `F5` | Copy to clipboard (default) |
| `Ctrl+Enter` | Send to selected tmux pane |
| `↑/↓` | Navigate tmux pane list (when tmux selector focused) |
| `Tab` | Cycle focus: file list → preview → tmux selector |

### Mouse Actions (Core MVP)

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

✅ **Core MVP is successful if:**
1. Can toggle prompt mode in < 2 keystrokes (F11 or toolbar click)
2. Can copy rendered prompt to clipboard in < 3 keystrokes (browse, Enter)
3. Template variables auto-fill correctly 90%+ of the time
4. Clipboard paste works in any application
5. Core feature adds < 300 lines of code (staying modular)
6. Users organize 20+ prompts easily with tree view
7. Integrates seamlessly with existing TFE workflow

✅ **Tmux enhancement is successful if:**
1. Detects all running tmux sessions accurately
2. Sends prompts to correct pane 100% of the time
3. Adds < 200 additional lines of code
4. Works smoothly across 2-3 monitor setups
5. Gracefully falls back to clipboard if tmux unavailable

---

## Notes & Decisions

### Why prioritize copy/paste over tmux?
- **Universal**: Works everywhere (any terminal, any app, any OS)
- **Simple**: No dependencies, no tmux requirement
- **Fast to implement**: Core feature complete in ~8-10 hours
- **User control**: Paste where and when needed
- **Tmux optional**: Power users can enable later if desired

### Why YAML over JSON?
- More human-readable
- Supports multi-line strings naturally
- Common in config files
- Easy to edit manually

### Why Tmux as enhancement (not core)?
- Not all users run tmux
- Adds complexity (session detection, pane selection)
- Copy/paste covers 90% of use cases
- Can be added later without breaking core feature

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

## Questions for Review (Core MVP)

- [ ] Should prompt mode auto-enter dual-pane? (Or keep as separate toggle?)
- [ ] Should we support custom variable prompts? (e.g., popup to fill `{{CUSTOM_INPUT}}`)
- [ ] Should copy history be per-prompt or global?
- [ ] Should we support JSON prompts in addition to YAML/MD?
- [ ] Should there be a default prompts directory (`~/.prompts/`), or require config?
- [ ] Should F5 or Enter be the primary copy shortcut? (Or both?)

## Questions for Review (Tmux Enhancement)

- [ ] Should tmux selector be always visible or toggle-able?
- [ ] Should there be a config to set default tmux pane per prompt?
- [ ] Should Ctrl+Enter be the send shortcut? (Or different combo?)

---

**Last Updated:** 2025-10-17 (Reorganized to prioritize copy/paste workflow)
**Status:** Planning Phase
**Branch:** `prompts`
**Implementation Order:** Phases 1-5 (Core MVP), Phase 6 (Optional Tmux)
