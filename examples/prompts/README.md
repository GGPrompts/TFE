# TFE Sample Prompts

This directory contains example AI prompts for common development workflows and Claude Code integration.

## Installation

Copy these prompts to your home directory:

```bash
mkdir -p ~/.prompts
cp examples/prompts/*.prompty ~/.prompts/
```

## Usage

1. Launch TFE: `tfe`
2. Press **F11** to enter Prompts Mode
3. Navigate to `🌐 ~/.prompts/` (shown at top of file list)
4. Select a prompt to see it with variables auto-filled from current context
5. **Tab/Shift+Tab** to navigate between fillable fields
6. Press **F5** to copy the rendered prompt to clipboard
7. Paste into Claude, ChatGPT, or your AI assistant of choice

## Featured Prompts

### 🔍 Context Analyzer (Advanced)
**File:** `context-analyzer.prompty`

Analyze Claude Code's `/context` output to get a comprehensive markdown report on:
- File relevance assessment (High/Medium/Low for each file)
- Suggested navigation paths for understanding the codebase
- CLAUDE.md optimization recommendations (add/remove/clarify)
- .claude folder analysis (agents/commands/skills)
- Token usage optimization (reduce context bloat)

**Workflow:**
1. In Claude Code, run `/context` and copy the output
2. Open TFE, press F11, select `context-analyzer.prompty`
3. Paste the context output into the `{{CONTEXT_PASTE}}` field
4. Press F5 to copy the rendered prompt
5. Paste into a new Claude chat (not your coding session)
6. Get back a detailed markdown report
7. Save as `docs/CONTEXT_ANALYSIS_YYYY-MM-DD.md`

**Perfect for:** Project maintenance, onboarding new developers, context optimization

---

## Standard Prompts

### 📝 Code Review
**File:** `code-review.prompty`

Review code for best practices, potential bugs, performance issues, and security vulnerabilities.

### 🔍 Explain Code
**File:** `explain-code.prompty`

Understand what unfamiliar code does with clear, step-by-step explanations.

### 🧪 Write Tests
**File:** `write-tests.prompty`

Generate comprehensive test cases including edge cases and error conditions.

### 📚 Document Code
**File:** `document-code.prompty`

Create documentation with function descriptions, parameters, usage examples, and notes.

### 🔧 Refactor Suggestions
**File:** `refactor-suggestions.prompty`

Get ideas for improving code structure, removing duplication, and optimizing performance.

### 🐛 Debug Help
**File:** `debug-help.prompty`

Get assistance debugging errors with cause identification and proposed fixes.

### 📝 Git Commit Message
**File:** `git-commit-message.prompty`

Write clear, concise commit messages following conventional commits format.

---

## Customization

Feel free to:
- Edit these prompts to match your workflow
- Add your own variables: `{{VARIABLE_NAME}}`
- Create new prompts for your specific needs
- Organize into subdirectories: `~/.prompts/coding/`, `~/.prompts/writing/`, etc.

## Available Variables

TFE automatically fills these from your current context:
- `{{file}}` - Current file path
- `{{filename}}` - Just the filename
- `{{project}}` - Project/directory name
- `{{path}}` - Full directory path
- `{{DATE}}` - Current date (YYYY-MM-DD)
- `{{TIME}}` - Current time (HH:MM:SS)

Custom variables (like `{{DESCRIPTION}}` or `{{CONTEXT_PASTE}}`) will show as fillable input fields in TFE with:
- **Tab/Shift+Tab** to navigate between fields
- **F3** on file fields (📁 blue) to browse and select files
- **Auto-filled fields** (🕐 green) are pre-populated but editable
- **Text fields** (📝 yellow) for short or long text input

## Prompt Template Formats

TFE supports multiple formats:

### 1. Microsoft Prompty (`.prompty`) - Recommended
```prompty
---
name: Prompt Name
description: What this prompt does
---
Prompt content here with {{variables}}
```

### 2. YAML (`.yaml`, `.yml`) - In .claude/ or ~/.prompts/
```yaml
name: Prompt Name
description: What this prompt does
template: |
  Prompt content here with {{variables}}
```

### 3. Markdown/Text (`.md`, `.txt`) - In .claude/ or ~/.prompts/
```markdown
Prompt content here with {{variables}}
```

**Note:** `.md` and `.txt` files are only recognized as prompts when in `.claude/` or `~/.prompts/` directories.

## More Information

See the main [README.md](../../README.md) for complete Prompts Library documentation and feature details.

## Contributing Your Prompts

Have a useful prompt? Consider sharing it:
1. Test it thoroughly with different projects
2. Add clear description and usage notes
3. Submit a pull request to add it to this collection
4. Help others discover valuable AI workflows!

---

**Pro Tip:** Start with Context Analyzer to understand and optimize your Claude Code setup, then use the other prompts for daily development tasks.
