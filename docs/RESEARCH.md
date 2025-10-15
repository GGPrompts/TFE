# GitHub Copilot CLI & Claude Code Integration Research

**Date:** 2025-10-15
**Topic:** VS Code extensions, MCP servers, GitHub Copilot CLI capabilities, and optimization strategies

---

## Overview

Research into whether VS Code extensions exist to give Claude Code access to tools similar to GitHub Copilot (terminal logs, browser console errors, etc.), and exploration of GitHub Copilot CLI capabilities.

---

## Key Findings

### 1. VS Code Extensions & MCP Servers Already Exist

Multiple MCP (Model Context Protocol) servers provide the capabilities needed:

#### **vscode-mcp-server** (by juehang)
- Terminal/shell command execution
- Diagnostics (errors & warnings)
- File operations
- Symbol search
- GitHub: https://github.com/juehang/vscode-mcp-server

#### **mcp-server-vscode** (by malvex)
- 25 tools total
- Language intelligence (go to definition, find references, symbol search)
- Full debugging support (breakpoints, variable inspection, call stacks)
- Diagnostics and code navigation
- Refactoring capabilities
- GitHub: https://github.com/malvex/mcp-server-vscode

#### **vscode-as-mcp-server** (by acomagu)
- Real-time diagnostics
- Terminal output retrieval via `get_terminal_output`
- Code checking via `code_checker`
- Command execution in integrated terminal
- GitHub: https://github.com/acomagu/vscode-as-mcp-server

### 2. Browser Console Access

#### **VS Code Simple Browser MCP Server**
- Executes JavaScript in VS Code's Simple Browser
- Real-time console log monitoring
- `get_console_logs` tool with filtering

#### **Chrome DevTools MCP** (by ChromeDevTools)
- Browser console messages via `list_console_messages`
- Network request monitoring (`list_network_requests`)
- JavaScript evaluation in browser context
- Screenshots and performance tracing
- GitHub: https://github.com/ChromeDevTools/chrome-devtools-mcp

### 3. How to Add MCP Servers to Claude Code

```bash
claude mcp add-json "vscode" '{"command":"mcp-proxy","args":["http://127.0.0.1:6010/sse"]}'
```

---

## GitHub Copilot Capabilities

### What GitHub Copilot's Agent Mode Can Do (2025)

- Read terminal output and auto-correct based on errors
- Access editor diagnostics (compile/lint errors)
- Run terminal commands
- Monitor test output
- Iterate autonomously until tasks complete
- Multi-step coding tasks across multiple files

---

## GitHub Copilot CLI Commands

### Two CLI Tools (Transition Period)

#### **NEW: `copilot` CLI** (Recommended - Standard after Oct 25, 2025)

**Interactive Mode:**
```bash
copilot
```

**Programmatic Mode:**
```bash
copilot -p "your prompt here"
copilot --prompt "explain how to set up docker"
```

**Slash Commands (inside interactive mode):**
- `/model` - Switch between AI models
- `/usage` - Check token usage and premium requests
- `/cwd /path` - Change working directory
- `/add-dir /path` - Add trusted directory
- `/login` - Authenticate
- `/feedback` - Submit feedback

**Special Syntax:**
- `@filename` - Reference specific files
- `!command` - Run direct shell commands

#### **OLD: `gh copilot` Commands** (Being deprecated Oct 25, 2025)

```bash
gh copilot suggest "Install git"
gh copilot explain "docker run -it ubuntu"
gh copilot config
gh copilot alias -- bash  # Creates shell aliases
```

**After setting up aliases:**
```bash
ghcs "find large files"  # gh copilot suggest
ghce "rm -rf /"          # gh copilot explain
```

---

## Model Usage & Limits

### Unlimited Usage Models (Included)
- GPT-5 mini
- GPT-4.1
- GPT-4o

### Premium Models (Limited Monthly Requests)
- Claude Sonnet 4.5 (newest!)
- Other advanced models

**Request Limits by Plan:**
- Pro: 300 premium requests/month
- Pro+: 1,500 premium requests/month
- Business: 300 per user/month
- Enterprise: 1,000 per user/month

**Key Point:** Unlimited usage of base models. When premium requests run out, fall back to unlimited base model usage.

---

## Programmatic Mode for Tasks

### Can Execute Tasks with Proper Flags

```bash
# Allow all tools
copilot -p "fix all TypeScript errors" --allow-all-tools

# Allow specific tools
copilot -p "create a new user model" --allow-tool write_to_file

# Multiple specific tools
copilot -p "setup linting config" --allow-tool write_to_file --allow-tool run_command
```

### Key Flags for Tasks
- `--allow-all-tools` - Autonomous execution without approval prompts
- `--allow-tool <tool_name>` - Allow specific tool (write_to_file, run_command, etc.)
- `--deny-tool <tool_name>` - Block specific tool

### Mode Comparison

| Feature | Interactive Mode | Programmatic Mode (`-p`) |
|---------|-----------------|-------------------------|
| Task execution | ✅ | ✅ |
| File editing | ✅ | ✅ |
| Multi-step tasks | ✅ | ✅ |
| Approval prompts | Shows by default | Blocked unless `--allow-*` flags |
| Use case | Exploratory, iterative | Scripting, automation, CI/CD |
| Context retention | Across conversation | Single-shot only |

---

## Comparison to Claude Code Subagents

### GitHub Copilot CLI `-p --allow-all-tools`

**Similar to Claude subagents:**
- ✅ Autonomous task execution
- ✅ Multi-step problem solving
- ✅ Can use tools without asking permission
- ✅ Handles complex tasks independently

**Different from Claude subagents:**
- ❌ Not background/async (synchronous - you wait)
- ❌ No parallel execution of multiple tasks
- ❌ Single-shot (doesn't maintain state)

### GitHub Copilot `gh agent-task` (Background Agent)

**More similar to Claude subagents:**
```bash
gh agent-task create "refactor the auth module"
gh agent-task list
gh agent-task status <task-id>
```

- ✅ Asynchronous execution
- ✅ Can keep working while it runs
- ✅ Opens draft PRs when done
- ✅ True delegation model

### Comparison Table

| Feature | Claude Subagents | Copilot `-p --allow-all-tools` | Copilot `gh agent-task` |
|---------|-----------------|-------------------------------|------------------------|
| Autonomous | ✅ | ✅ | ✅ |
| Multi-step | ✅ | ✅ | ✅ |
| Background/async | ✅ | ❌ | ✅ |
| Parallel tasks | ✅ | ❌ | ✅ |
| Returns results | ✅ | ✅ (stdout) | ✅ (draft PR) |

---

## Optimization Strategy: Using Copilot as Scout

### The Idea

Use GitHub Copilot's unlimited GPT-5 mini to scout codebases before Claude Code executes tasks, saving context tokens.

### Why This is Smart

**Context Savings:**
- GPT-5 mini does broad codebase search (unlimited, fast)
- Returns curated list of relevant files/patterns
- Claude only reads important files instead of exploring blindly
- Saves 10-100K tokens potentially

**Speed:**
- GPT-5 mini is very fast for reconnaissance
- Can run in parallel
- Preprocessing happens almost instantly

**Cost Efficiency:**
- Unlimited usage = no premium request consumption
- Claude's context window is expensive
- Copilot's preprocessing is "free" with subscription

### Example Slash Command

`.claude/commands/scout.md`:
````markdown
---
description: Use GitHub Copilot to scout codebase before task
---

First, run this command to gather context:

```bash
copilot -p "Analyze the codebase for: {{prompt}}

Find and list:
1. All relevant files (with brief descriptions)
2. Key patterns and conventions used
3. Dependencies and imports involved
4. Potential gotchas or complexity areas
5. Recommended files to examine first

Be concise but thorough." --allow-tool read_file --allow-tool list_files
```

Then use that context to: {{prompt}}
````

### Usage Example

```bash
/scout refactor authentication to use JWT tokens

# Behind the scenes:
# 1. Copilot scouts → returns focused file list + patterns
# 2. Claude receives pre-filtered context
# 3. Claude works on actual refactoring with targeted knowledge
```

### Hybrid Approach: Parallel Execution

```markdown
Run these in parallel:
1. GitHub Copilot: Scout file locations and patterns
2. Claude subagent: Analyze architecture and design patterns

Combine both results for comprehensive context.
```

### Scout Comparison

| Aspect | Claude general-purpose subagent | Copilot `-p` scout |
|--------|-------------------------------|-------------------|
| Speed | Fast | **Very fast** |
| Cost | Uses Claude context/tokens | **Unlimited/free** |
| File access | Via Claude tools | Via Copilot tools |
| Quality | High reasoning | Fast reconnaissance |
| Best for | Deep analysis | Quick mapping |

**Sweet Spot:** Use Copilot for fast initial mapping, then Claude does deep work with focused context.

---

## Installation

### GitHub Copilot CLI
```bash
npm install -g @github/copilot@latest
```

### Requires
- GitHub CLI (`gh`)
- Active GitHub Copilot subscription
- Authentication via `gh auth login`

---

## Additional Resources

- [GitHub Copilot CLI Docs](https://docs.github.com/en/copilot/concepts/agents/about-copilot-cli)
- [Using GitHub Copilot CLI](https://docs.github.com/en/copilot/how-tos/use-copilot-agents/use-copilot-cli)
- [Claude Code VS Code Integration](https://docs.claude.com/en/docs/claude-code/vs-code)
- [MCP Developer Guide](https://code.visualstudio.com/api/extension-guides/ai/mcp)
- [VS Code Extension API](https://code.visualstudio.com/api/references/vscode-api)

---

## Notes

- MCP architecture makes Claude Code modular - can install existing MCP servers for terminal, diagnostics, browser console access
- GitHub Copilot CLI is in public preview (as of October 2025)
- Claude Sonnet 4.5 now available in GitHub Copilot CLI
- Important transition: Old `gh copilot` commands being deprecated October 25, 2025

---

## Project Ideas: Terminal Tools Strategy

### Philosophy: Build Reusable Tools, Not Monolithic Apps

**Why terminal tools are better for growing codebases:**
- ✅ **Portable** - Use them in any project, any language
- ✅ **Composable** - Pipe them together with other tools
- ✅ **Small scope** - Easier to finish and maintain
- ✅ **Immediately useful** - You use them daily
- ✅ **Great for learning** - Each one teaches something new
- ✅ **Compound value** - Each tool makes you more productive forever

Instead of building one massive project that gets abandoned, build 5-10 small tools that you'll use forever.

---

## Main Project Idea: AI-Enhanced File Explorer for Windows→WSL Users

### The Problem

Windows developers transitioning to WSL/Linux struggle with:
- Terminal file navigation
- Understanding symlinks, chmod, file permissions
- Unfamiliar terminology (what's a "symlink"? It's a shortcut!)
- Command-line file operations that were clicks in Windows Explorer
- No visual feedback
- Fear of breaking things with wrong commands

**Existing tools (Midnight Commander, ranger, nnn, lf, yazi) are built BY Linux users FOR Linux users.**

### The Opportunity

Build a **beginner-friendly, AI-enhanced TUI file manager** that bridges Windows and Linux concepts.

### Midnight Commander Background

- **Open Source**: GNU GPL v3+ license
- **Repository**: https://github.com/MidnightCommander/mc
- **Can fork/modify**: Yes, as long as you keep it GPL-licensed
- **Written in**: C (traditional)
- **Features**: Dual-pane, FTP, archive browsing, built-in viewer/editor

### Modern TUI File Managers (Proof of Concept)

**Yazi** (2023 - Most Modern)
- Written in Rust
- Async I/O (never freezes)
- Image previews in terminal
- Mouse support
- Icons with Nerd Fonts
- File previews with syntax highlighting
- Blazing fast

**Superfile**
- Modern, eye-candy UI
- Multi-panel workflow
- Marketed as "Explorer alternative"

**felix**
- Vim-like keybindings
- Fast and configurable
- Image previews

**nnn**
- Lightweight, fast (C)
- Plugin system
- Minimal but powerful

**ranger**
- Python-based
- Highly configurable
- Preview pane

### What Makes This Project Unique

**1. Windows-Friendly Terminology**
- Show both Windows AND Linux terms
- "Properties (chmod)" not just "chmod"
- "Shortcut (symlink)" not just "symlink"
- "Read-only" instead of "444 permissions"

**2. AI Integration** (Revolutionary Feature)
```
┌─────────────────────────────────────────────┐
│ Name          Modified      Size     Type   │
│ 📁 Documents  2h ago        -        Folder │
│ 📄 config.js  5m ago        2.3KB    JS     │
│ 🔗 link.txt   1d ago        →app/    Link   │
│                                             │
│ [💡 Ask Claude: "What is this link for?"]  │
│ [🤖 Ask Copilot: "Explain this config"]    │
└─────────────────────────────────────────────┘
```

**3. Visual chmod/Permissions**
```
┌─ Properties: config.js ───────────────┐
│                                       │
│ Owner:  [x] Read  [x] Write  [ ] Exec │
│ Group:  [x] Read  [ ] Write  [ ] Exec │
│ Others: [x] Read  [ ] Write  [ ] Exec │
│                                       │
│ Linux: -rw-r--r-- (644)              │
│ Windows: Read-only [ ]                │
└───────────────────────────────────────┘
```

**4. Context Visualizer** (KILLER FEATURE)

Show what Claude Code would see if launched from this directory!

```
┌─────────────────────────────────────────────────────────────┐
│ 📁 /home/user/my-project          Context: 45K / 200K tokens│
├─────────────────────────────────────────────────────────────┤
│ Name              Size    Tokens   Status                   │
│ ✅ src/           -       25K      Included                 │
│   ✅ app.js       3KB     800      In context              │
│   ✅ utils.js     2KB     500      In context              │
│   ⚠️  large.js    50KB    12K      Too large (preview only)│
│ ❌ node_modules/  -       -        .gitignore              │
│ ❌ .env           1KB     -        .gitignore              │
│ ✅ README.md      2KB     400      In context              │
│ 🔲 tests/         -       8K       Not in default context  │
│                                                             │
│ 💡 Tip: Add .claudeignore to exclude tests/                │
└─────────────────────────────────────────────────────────────┘
```

**What the Context Visualizer Shows:**

**Per File:**
- ✅ **Included** - Will be in Claude's context
- ❌ **Excluded** - .gitignore, .claudeignore, binary files
- ⚠️ **Too Large** - Claude will only see preview/summary
- 🔲 **Optional** - Available but not auto-loaded
- 📊 **Token estimate** - How much context each file uses

**Summary Stats:**
- Total token usage if Claude launched from here
- Number of files included
- Warning if over token budget
- Which ignore files are active (.gitignore, .claudeignore)
- Suggestions to optimize context

**Interactive Context Building:**
```
Commands:
- Press 'c' to toggle file in/out of context
- Press 'i' to add file/folder to .claudeignore
- Press 's' to see Claude's summary of this file
- Press 'v' to view exactly what Claude sees
- Press 'Enter' to launch Claude Code with this context
```

**Context Optimization Suggestions:**
```
💡 You're at 180K/200K tokens. Suggestions:
  - Exclude build/ folder (saves 45K tokens)
  - Add *.test.js to .claudeignore (saves 30K)
  - Summarize docs/ instead of full content (saves 20K)
  - node_modules/ already excluded by .gitignore ✓
```

**Why This is Valuable:**

For you:
- Know what Claude sees BEFORE wasting a session
- Optimize context before starting
- Debug "why doesn't Claude see my file?" issues
- Maximize useful context, minimize noise

For Windows users learning WSL:
- Visual understanding of .gitignore
- See how hidden files work
- Understand project structure hierarchy

For everyone:
- Better prompt engineering
- Token budget management
- Avoid context window overflow

### Core Features

**1. Beginner-Friendly Navigation**
- Breadcrumb path: `/home/user/projects/myapp`
- Quick Access sidebar (Favorites, Recent, Downloads)
- Search bar (not grep - actual search box)
- Right-click context menus
- Mouse support for clicks
- Keyboard shortcuts with tooltips

**2. Safe Operations**
- Confirm before destructive operations
- "Undo" for recent operations (with trash)
- Preview mode before executing
- Plain English explanations

**3. Plain English Commands**
Type what you want:
- "make this executable" → `chmod +x`
- "create a shortcut to this" → `ln -s`
- "show hidden files" → toggle dotfiles
- "search for config files" → smart search

**4. AI Integration**
- Right-click → "Ask Claude what this does"
- Right-click → "Ask Copilot to explain"
- "AI Scout" - use Copilot to analyze directory before Claude launch
- Inline explanations from AI

**5. Translation Layer**
Show concepts in both Windows and Linux terms:
```
File Properties:
├─ Name: config.json
├─ Type: JSON Configuration File
├─ Size: 2.3 KB
├─ Modified: 5 minutes ago
├─ Permissions:
│  ├─ Linux: -rw-r--r-- (644)
│  └─ Windows equivalent: Read-only for others
├─ Is Symlink: No
│  └─ (Symlink = Windows Shortcut)
└─ Owner: user (you)
```

**6. Context Visualizer Integration**
- Tab 1: File Browser
- Tab 2: Context View
- See what Claude/AI will see from this directory
- Optimize before launching
- Token usage preview

**7. Help & Learning**
- Tooltips everywhere
- "What's a symlink?" button → explanation
- Built-in tutorials
- Hover over icons for explanations
- Command explainer: shows what the equivalent terminal command would be

### Technical Implementation

**Tech Stack Options:**

**Option 1: Modern & Fast (Recommended)**
- **Language**: Rust
- **TUI Framework**: `ratatui` (what Yazi uses)
- **Why**: Best performance, async I/O, modern
- **Con**: Steeper learning curve

**Option 2: Fast & Simple**
- **Language**: Go
- **TUI Framework**: `bubbletea`
- **Why**: Fast, easier than Rust, good TUI support
- **Con**: Not as fast as Rust

**Option 3: Rapid Development**
- **Language**: Python
- **TUI Framework**: `textual`
- **Why**: Easiest, fastest to prototype
- **Con**: Slower than Rust/Go

**Recommended**: Start with **Python + Textual** for rapid prototyping, rewrite in Rust later if needed.

### Core Components to Build

**1. File Browser Engine**
- Directory traversal
- File listing with metadata
- Sorting, filtering
- Icon mapping (file type → icon)

**2. Ignore File Parser**
- Parse .gitignore patterns
- Parse .claudeignore patterns
- Apply exclusion rules
- Show what's excluded and why

**3. Token Counter**
- Estimate tokens per file (~4 chars = 1 token)
- Calculate directory totals recursively
- Track against Claude's limits (200K tokens)

**4. Context Simulator**
- Simulate Claude Code's file reading hierarchy
- Show which files get full read vs summary
- Display exactly what would be in context

**5. AI Integration Layer**
- Call `copilot -p` for quick questions
- Integration with Claude Code MCP servers
- Optional: Direct Anthropic API integration

**6. Permission Visualizer**
- Parse Unix permissions (chmod)
- Display in checkbox UI
- Two-way translation (visual ↔ octal)

**7. Operation Queue**
- Queue file operations
- Allow undo/redo
- Trash integration
- Confirm destructive ops

### Integration with Other Tools

**File Explorer + Context Viz + AI Scout Workflow:**

1. **Browse** with TUI file manager
2. **Preview context** before launching Claude (see token usage)
3. **Scout** with `copilot -p` to pre-filter important files
4. **Optimize** by adjusting .claudeignore
5. **Launch** Claude Code with optimized context
6. **Ask AI** directly from file manager during work

### Why This Would Be Revolutionary

**No one has built:**
- A file manager specifically for Windows→WSL learners
- A Claude Code context visualizer
- An AI-integrated file browser
- A file manager that explains Linux concepts in Windows terms

**This solves real pain points:**
- Windows devs scared of terminal
- Wasted Claude sessions due to poor context
- Not knowing what files Claude sees
- Token budget management
- Understanding Linux file concepts

### Potential Project Names

- "Explorer TUI" / "ExTUI"
- "Context Commander"
- "AI Navigator"
- "Smart Commander"
- "Friendly Files"
- "Bridge Explorer" (Windows→Linux bridge)
- "Claude Navigator"

---

## Additional Terminal Tool Ideas

Build these as separate small tools to complement your workflow:

### 1. Git Status Dashboard
- Show all your repos' status at once
- Build on your existing git status script
- Interactive TUI version
- Quick commit/push from dashboard

### 2. Claude Code Session Manager
- Save/restore Claude sessions per project
- Quick switch between project contexts
- Store common prompts/slash commands per project
- Session history and notes

### 3. Project Scaffolder
- Your custom templates for new projects
- Stop "starting from scratch" fatigue
- Ask Claude to generate boilerplate
- Language/framework templates

### 4. Dependency Checker
- Check which projects have outdated deps
- Works across all your repos
- Run updates in batch
- Alert on security issues

### 5. Code Search Across Projects
- Search ALL your projects at once
- "Where did I implement that pattern?"
- Could integrate with Copilot scout idea
- Regex support, context preview

### 6. Smart Symlink Manager
- Manage symlinks visually
- Share configs between projects
- Explains what links to what
- Windows user friendly

### 7. Terminal Snippet Manager
- Save commands you always forget
- Searchable, categorized
- Better than `.bash_history` scrolling
- Can ask AI to explain snippets

### 8. Token Counter CLI
- Standalone tool: estimate tokens in files/directories
- Check before starting Claude session
- Find token-heavy files
- Optimize your codebase for AI context

---

## Implementation Notes: Context Visualizer

### How to Build the Context Feature

**1. Parse Ignore Files**
```python
# Parse .gitignore patterns
# Parse .claudeignore patterns
# Apply glob pattern matching
# Determine which files are excluded
```

**2. Estimate Tokens**
```python
def estimate_tokens(file_path):
    """Rough estimate: ~4 characters = 1 token"""
    try:
        with open(file_path, 'r') as f:
            content = f.read()
            return len(content) // 4
    except:
        return 0  # Binary or unreadable
```

**3. Simulate Claude's View**
```python
def get_claude_context(directory):
    """
    Returns what Claude Code would see:
    - File tree structure
    - Full content of small files
    - Summaries of large files
    - Excluded files (with reasons)
    """
    context = {
        'included': [],
        'excluded': [],
        'too_large': [],
        'total_tokens': 0
    }
    # Walk directory, apply rules
    return context
```

**4. Default Exclusion Patterns**
Research Claude Code's defaults:
- `node_modules/`, `__pycache__/`, `.git/`
- Binary files, images (unless specified)
- Files over certain size threshold
- Standard ignore patterns

**5. Interactive Optimization**
```python
# Allow toggling files in/out
# Auto-suggest optimizations
# Show before/after token counts
# Generate .claudeignore suggestions
```

### Data to Track

**Per File:**
- Path
- Size (bytes)
- Estimated tokens
- Status (included/excluded/too_large)
- Exclusion reason (if applicable)
- File type
- Last modified

**Per Directory:**
- Total size
- Total tokens
- Number of files included/excluded
- Active ignore files
- Optimization suggestions

**Summary:**
- Total context size (tokens)
- Percentage of Claude's limit used
- Number of files
- Warnings/suggestions

---

## Next Steps

### For the File Explorer Project

**Phase 1: Basic File Browser**
1. Choose tech stack (recommend Python + Textual for speed)
2. Build basic file listing
3. Add navigation (arrow keys, mouse)
4. Implement file operations (copy, move, delete with confirmation)

**Phase 2: Context Visualizer**
1. Build ignore file parser (.gitignore, .claudeignore)
2. Implement token counter
3. Create context simulation
4. Build visualization UI
5. Add optimization suggestions

**Phase 3: AI Integration**
1. Add "Ask Claude" feature
2. Add "Ask Copilot" feature
3. Integrate AI Scout (Copilot pre-filtering)
4. Add inline explanations

**Phase 4: Windows→Linux Bridge**
1. Add translation layer for terminology
2. Visual chmod/permissions editor
3. Symlink explainer and creator
4. Plain English command interface

**Phase 5: Polish**
1. Add themes
2. Built-in tutorials
3. Tooltips and help system
4. Configuration options
5. Keyboard shortcuts

### Success Metrics

**You'll know it's successful when:**
- You use it daily in your own workflow
- Other Windows→WSL users find it helpful
- It saves you time and tokens on Claude sessions
- You can visualize context before launching Claude
- It becomes your go-to file manager in terminal

### Portfolio Value

**This project shows:**
- TUI development skills
- Understanding of AI context management
- User-focused design (Windows→Linux bridge)
- Innovation (context visualizer is unique)
- Practical tool-building
- Cross-platform thinking

---

## Resources for Building TUI File Manager

### Learning TUI Development

**Python + Textual:**
- Docs: https://textual.textualize.io/
- Examples: Built-in examples in docs
- Similar projects: Browse Textual showcase

**Rust + Ratatui:**
- Docs: https://ratatui.rs/
- Examples: Yazi source code
- Tutorial: Ratatui book

**Go + Bubbletea:**
- Docs: https://github.com/charmbracelet/bubbletea
- Examples: Charm.sh examples
- Community: Active Discord

### Studying Existing File Managers

**For inspiration:**
- Yazi: https://github.com/sxyazi/yazi (modern, async)
- Ranger: https://github.com/ranger/ranger (Python, mature)
- Midnight Commander: https://github.com/MidnightCommander/mc (classic)
- Felix: https://github.com/kyoheiu/felix (simple, Rust)

### Token Counting
- tiktoken (OpenAI's tokenizer): https://github.com/openai/tiktoken
- Anthropic's tokenizer: Check Claude docs
- Or use simple heuristic: ~4 chars = 1 token

### Ignore File Parsing
- pathspec (Python): https://github.com/cpburnz/python-pathspec
- gitignore (Rust): https://crates.io/crates/gitignore
- Study .gitignore spec: Git documentation

---

## Final Thoughts

**This isn't just a file manager - it's a bridge between Windows and Linux, a context optimizer for AI coding, and a learning tool all in one.**

No one has built anything like this yet. The context visualizer alone would be incredibly valuable to anyone using Claude Code seriously.

Start small, iterate, and use it yourself first. If it solves YOUR problems, it'll solve others' problems too.
