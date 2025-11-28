# AI Router TUI - Implementation Plan

**Goal**: Build a compact, unified chat interface that routes prompts to multiple AI tools (Claude Code, Copilot CLI, Codex), integrates with tmux, and provides intelligent pattern matching and command queueing.

**Target Environment**: Termux (mobile) + PC with Docker containers for safety

---

## Architecture Overview

```
┌─ AI Router TUI (Bubble Tea) ─────────────┐
│                                          │
│  ┌─ Chat Interface ───────────────────┐ │
│  │ > explain auth.go                  │ │
│  │ Route: Claude Code                 │ │
│  │ [Analyzing...]                     │ │
│  └────────────────────────────────────┘ │
│                                          │
│  ┌─ Command Queue ─────────────────────┐│
│  │ 1. [Worker-1] Run tests           ││
│  │    [Edit] [Send] [Skip]           ││
│  │ 2. [Browser] Open localhost       ││
│  │    [Edit] [Send] [Skip]           ││
│  └────────────────────────────────────┘│
│                                          │
│  ┌─ Active Workers ────────────────────┐│
│  │ [1] Claude-Worker: Editing...      ││
│  │ [2] Copilot: Idle                  ││
│  └────────────────────────────────────┘│
└──────────────────────────────────────────┘
```

---

## Tech Stack (Based on Research)

### Core Framework
- **Bubble Tea** (`github.com/charmbracelet/bubbletea`)
  - Model-Update-View architecture (like Elm/Redux)
  - Perfect for complex TUI apps
  - Active development (2025)
  - Used in production (Glow, Charm, K9s)

### UI Components (Bubbles)
- **Viewport** (`github.com/charmbracelet/bubbles/viewport`)
  - Scrollable chat history
  - Mouse wheel support
  - High-performance rendering

- **Textarea** (`github.com/charmbracelet/bubbles/textarea`)
  - Multiline prompt input (perfect for voice-to-text)
  - Unicode support, paste functionality
  - Line numbers optional

- **List** (`github.com/charmbracelet/bubbles/list`)
  - Command queue display
  - Worker status list
  - Template/pattern selection

### Styling
- **Lipgloss** (`github.com/charmbracelet/lipgloss`)
  - Terminal styling (already used in TFE)
  - Color-coded AI responses
  - Consistent with TFE design

### Markdown Rendering
- **Glamour** (`github.com/charmbracelet/glamour`)
  - Render AI responses with proper markdown
  - Code syntax highlighting
  - Multiple themes (dark/light/auto)
  - Published April 2025 (actively maintained)

### Tmux Integration
- **gotmux** (`github.com/GianlucaP106/gotmux`)
  - Send commands to specific tmux panes
  - Control mode support
  - Type-safe interface
  - Most comprehensive option

### Fuzzy Search
- **go-fuzzyfinder** (`github.com/ktr0731/go-fuzzyfinder`)
  - Pattern matching for suggestions
  - Multi-select support (like fzf -m)
  - 250+ projects using it (April 2025)
  - OR use fzf algorithm directly from `github.com/junegunn/fzf`

### Database (History Storage)
- **modernc.org/sqlite**
  - Pure Go implementation (no CGO required!)
  - Perfect for Termux (no C compiler needed)
  - database/sql compatible
  - OR **github.com/mattn/go-sqlite3** if on PC (CGO version, faster)

---

## Module Architecture (Following TFE Patterns)

```
ai-router/
├── main.go (21 lines)           - Entry point ONLY
├── types.go                     - Type definitions
├── styles.go                    - Lipgloss styles
├── model.go                     - Model initialization
├── update.go                    - Main update dispatcher
├── update_keyboard.go           - Keyboard handling
├── update_mouse.go              - Mouse handling (optional)
├── view.go                      - View rendering
├── render_chat.go               - Chat interface rendering
├── render_queue.go              - Command queue rendering
├── render_workers.go            - Worker status rendering
├── ai_router.go                 - AI tool routing logic
├── tmux_integration.go          - Tmux send-keys, pane control
├── pattern_matcher.go           - Fuzzy search, suggestions
├── command_queue.go             - Queue management
├── history_db.go                - SQLite history storage
├── worker_manager.go            - Worker lifecycle management
├── safety.go                    - Guardrails & capability limits
└── helpers.go                   - Utility functions
```

---

## Phase 1: Core Chat Interface (Day 1)

### Goal: Basic chat UI with single AI (Claude Code)

**Files to create:**
1. `types.go` - Define core types
2. `model.go` - Initialize model
3. `update.go` - Handle messages
4. `view.go` - Render UI
5. `render_chat.go` - Chat rendering

**Features:**
- ✅ Multiline input (textarea component)
- ✅ Send prompt to Claude Code
- ✅ Display response with Glamour markdown
- ✅ Scrollable chat history (viewport)
- ✅ Compact mode (works in narrow pane)

**Key Types:**
```go
type model struct {
    chatHistory   []ChatMessage
    input         textarea.Model
    viewport      viewport.Model
    aiTool        AITool  // claude, copilot, codex
    width, height int
}

type ChatMessage struct {
    Timestamp time.Time
    Tool      AITool
    Role      Role  // user, assistant
    Content   string
}

type AITool int
const (
    Claude AITool = iota
    Copilot
    Codex
)
```

---

## Phase 2: AI Routing & Multi-Tool Support (Day 2)

### Goal: Route prompts to different AI tools

**Files to create:**
- `ai_router.go` - Routing logic

**Features:**
- ✅ Detect tool from prompt prefix (`/claude`, `/copilot`, `/codex`)
- ✅ Auto-suggest tool based on keywords
- ✅ Execute Claude Code commands
- ✅ Execute `gh copilot suggest` and `gh copilot explain`
- ✅ Color-code responses by tool

**Routing Logic:**
```go
type Router struct {
    defaultTool AITool
    patterns    map[string]AITool
}

func (r *Router) Route(prompt string) (AITool, string) {
    // Check explicit prefix
    if strings.HasPrefix(prompt, "/claude") {
        return Claude, strings.TrimPrefix(prompt, "/claude")
    }

    // Pattern matching
    if containsKeywords(prompt, []string{"explain", "refactor", "architecture"}) {
        return Claude, prompt
    }

    if containsKeywords(prompt, []string{"suggest", "command", "how to"}) {
        return Copilot, prompt
    }

    return r.defaultTool, prompt
}
```

---

## Phase 3: Command Queue with Approval (Day 3)

### Goal: Human-in-the-loop command execution

**Files to create:**
- `command_queue.go` - Queue management
- `render_queue.go` - Queue UI rendering

**Features:**
- ✅ Orchestrator AI queues commands
- ✅ User reviews before sending
- ✅ Edit commands inline
- ✅ Skip/approve individual commands
- ✅ Batch approval option
- ✅ Undo recently sent commands

**Key Types:**
```go
type CommandQueue struct {
    items    []QueuedCommand
    selected int
}

type QueuedCommand struct {
    ID        string
    Worker    string
    Command   string
    Priority  Priority
    Status    Status  // Pending, Approved, Sent, Failed
    Editable  bool
    CreatedAt time.Time
}

type Priority int
const (
    Low Priority = iota
    Medium
    High
)
```

**UI:**
```
┌─ Command Queue (3 pending) ──────────┐
│ 🔴 HIGH: Fix auth bug (Worker-1)     │
│    [Edit] [Send] [Skip]              │
│                                      │
│ 🟡 MED: Update tests (Worker-2)      │
│    [Edit] [Send] [Skip]              │
│                                      │
│ 🟢 LOW: Regenerate docs (Worker-3)   │
│    [Edit] [Send] [Skip]              │
│                                      │
│ [S]end All  [C]lear  [R]eview        │
└──────────────────────────────────────┘
```

---

## Phase 4: Tmux Integration (Day 4)

### Goal: Send commands to tmux panes

**Files to create:**
- `tmux_integration.go` - Tmux control

**Features:**
- ✅ List available tmux panes
- ✅ Send command to specific pane
- ✅ Capture pane output
- ✅ Visual pane selector

**Using gotmux library:**
```go
import "github.com/GianlucaP106/gotmux"

type TmuxManager struct {
    server *gotmux.Server
}

func (tm *TmuxManager) SendToPane(paneID string, command string) error {
    pane, err := tm.server.GetPane(paneID)
    if err != nil {
        return err
    }

    return pane.SendKeys(command)
}

func (tm *TmuxManager) ListPanes() ([]Pane, error) {
    sessions, err := tm.server.ListSessions()
    // ... gather all panes from all sessions
    return panes, nil
}
```

**Workflow:**
```
User: "run tests in pane 2"
Router: Queues command for pane 2
User: Approves
System: Sends "npm test" to tmux pane 2
```

---

## Phase 5: Pattern Matching & History (Day 5)

### Goal: Smart suggestions from history

**Files to create:**
- `pattern_matcher.go` - Fuzzy search
- `history_db.go` - SQLite storage

**Features:**
- ✅ Store all prompts/responses in SQLite
- ✅ Fuzzy search history
- ✅ Tag prompts by category
- ✅ Template expansion
- ✅ Suggest completions

**Database Schema:**
```sql
CREATE TABLE chat_history (
    id INTEGER PRIMARY KEY,
    timestamp INTEGER NOT NULL,
    tool TEXT NOT NULL,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    tags TEXT  -- JSON array
);

CREATE TABLE templates (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    prompt TEXT NOT NULL,
    tool TEXT NOT NULL,
    tags TEXT
);

CREATE INDEX idx_tags ON chat_history(tags);
CREATE INDEX idx_timestamp ON chat_history(timestamp);
```

**Pattern Matching with go-fuzzyfinder:**
```go
import "github.com/ktr0731/go-fuzzyfinder"

func (pm *PatternMatcher) SearchHistory(query string) ([]ChatMessage, error) {
    messages, err := pm.db.GetAllHistory()
    if err != nil {
        return nil, err
    }

    // Fuzzy search
    idx, err := fuzzyfinder.FindMulti(
        messages,
        func(i int) string {
            return messages[i].Content
        },
    )

    return selectedMessages, nil
}
```

---

## Phase 6: Worker Management & Safety (Day 6)

### Goal: Multi-worker orchestration with guardrails

**Files to create:**
- `worker_manager.go` - Worker lifecycle
- `safety.go` - Capability limits
- `render_workers.go` - Worker status UI

**Features:**
- ✅ Spawn worker sessions (containerized)
- ✅ Monitor worker status
- ✅ Kill runaway workers
- ✅ Capability limits (read-only, no symlinks, etc.)
- ✅ Semantic analysis (detect "test" bypass)
- ✅ Emergency stop all workers

**Safety Architecture:**
```go
type SafetyConfig struct {
    MaxWorkers         int
    MaxBrowserTabs     int
    MaxFilesPerAction  int
    RequireApprovalFor []ActionType
    BlockedSyscalls    []string
    ReadOnlyPaths      []string
    AllowedPaths       []string
}

type Worker struct {
    ID           string
    Status       WorkerStatus
    Container    *Container  // Docker/podman container
    Capabilities Capabilities
    ActivityLog  []Action
}

func (w *Worker) ValidateCommand(cmd Command) error {
    // Recursive script scanning
    if cmd.IsScript() {
        content := readScript(cmd.Path)
        for _, blocked := range w.Capabilities.BlockedSyscalls {
            if contains(content, blocked) {
                return fmt.Errorf("script contains blocked syscall: %s", blocked)
            }
        }
    }

    // Check against allowed paths
    if !w.Capabilities.CanWrite(cmd.TargetPath) {
        return errors.New("write access denied")
    }

    return nil
}
```

**Worker UI:**
```
┌─ Active Workers ─────────────────────┐
│ [1] Claude-Worker (Backend)          │
│     Status: Editing auth.go          │
│     Actions: 5/20 per min            │
│     Uptime: 2m 15s                   │
│     [Pause] [Kill]                   │
│                                      │
│ [2] Copilot-Worker (Testing)         │
│     Status: Running tests...         │
│     Actions: 2/20 per min            │
│     Uptime: 45s                      │
│     [Pause] [Kill]                   │
│                                      │
│ [Emergency Stop All] [Spawn Worker]  │
└──────────────────────────────────────┘
```

---

## Phase 7: Polish & Integration (Day 7)

### Goal: Complete features and integrate with TFE

**Features:**
- ✅ Voice-to-text friendly (multiline editing)
- ✅ Compact mode for narrow panes
- ✅ Keyboard shortcuts (vim-style)
- ✅ Context file attachment (send current file from TFE)
- ✅ Export chat to markdown
- ✅ Session persistence
- ✅ Config file support

**Keyboard Shortcuts:**
```
Ctrl+S  - Send prompt
Ctrl+Q  - Quit
Ctrl+K  - Kill all workers
Ctrl+P  - Pause/Resume
Ctrl+H  - Search history
Ctrl+T  - Toggle tmux pane selector
Ctrl+E  - Edit queued command
Tab     - Cycle through queue items
Enter   - Approve selected command
Esc     - Cancel/Back
/       - Command mode (routing)
```

**Integration with TFE:**
```bash
# From TFE, press F11 to send current file to AI Router
# AI Router receives: "/claude explain /path/to/file.go"
```

---

## Deployment Strategy

### Termux (Mobile)
```bash
# Install dependencies
pkg install golang git tmux

# Clone and build
git clone <ai-router-repo>
cd ai-router
go build -o ai-router

# Run in tmux pane
tmux split-window -h -p 30  # 30% width pane
./ai-router --compact
```

### PC with Docker (Safe Experimentation)
```dockerfile
# Dockerfile for worker containers
FROM golang:1.23-alpine

# Install Claude Code / Copilot CLI
RUN apk add --no-cache git nodejs npm
RUN npm install -g @anthropic/claude-code
RUN npm install -g @github/copilot

# Isolated config
ENV CLAUDE_CONFIG_DIR=/app/.claude-worker
VOLUME /workspace

# Read-only workspace option
CMD ["claude-code", "--concise"]
```

```yaml
# docker-compose.yml
version: '3.8'
services:
  ai-router:
    build: .
    volumes:
      - ./workspace:/workspace
    environment:
      - AI_ROUTER_MODE=orchestrator

  claude-worker-1:
    build: .
    volumes:
      - ./workspace:/workspace:ro  # Read-only
      - claude-worker-1-config:/app/.claude-worker
    environment:
      - AI_ROUTER_MODE=worker
      - WORKER_ID=claude-1
      - WORKER_ROLE=code-generator

  claude-worker-2:
    build: .
    volumes:
      - ./workspace:/workspace:ro
      - claude-worker-2-config:/app/.claude-worker
    environment:
      - AI_ROUTER_MODE=worker
      - WORKER_ID=claude-2
      - WORKER_ROLE=tester

volumes:
  claude-worker-1-config:
  claude-worker-2-config:
```

---

## Config File Example

```toml
# ~/.config/ai-router/config.toml

[general]
default_tool = "claude"
compact_mode = true
tmux_integration = true

[routing]
# Pattern-based routing
[routing.patterns]
explain = "claude"
refactor = "claude"
suggest = "copilot"
command = "copilot"
generate = "codex"

[safety]
max_workers = 3
max_browser_tabs = 5
max_actions_per_minute = 20
require_approval = true
safe_mode = "paranoid"  # paranoid, strict, normal, permissive

[safety.blocked_syscalls]
syscalls = ["symlink", "mount", "chmod 777"]

[safety.capabilities]
allow_write = ["/workspace/src/**"]
deny_write = ["**/.config/**", "**/.claude/**"]
allow_exec = false
allow_network = false

[history]
database = "~/.config/ai-router/history.db"
max_entries = 10000
auto_archive = true

[tmux]
default_session = "dev"
pane_layout = "main-vertical"

[ui]
theme = "dark"  # dark, light, auto
show_timestamps = true
show_worker_status = true
markdown_rendering = true

[shortcuts]
# Custom keyboard shortcuts
send = "Ctrl+S"
quit = "Ctrl+Q"
kill_all = "Ctrl+K"
history = "Ctrl+H"
```

---

## Dependencies (go.mod)

```go
module github.com/yourusername/ai-router

go 1.23

require (
    github.com/charmbracelet/bubbletea v1.2.4
    github.com/charmbracelet/bubbles v0.20.0
    github.com/charmbracelet/lipgloss v1.0.0
    github.com/charmbracelet/glamour v0.8.0
    github.com/GianlucaP106/gotmux v0.1.0
    github.com/ktr0731/go-fuzzyfinder v0.8.0
    modernc.org/sqlite v1.34.4
    github.com/pelletier/go-toml/v2 v2.2.3
)
```

---

## Testing Strategy

### Unit Tests
```go
// ai_router_test.go
func TestRoutePrompt(t *testing.T) {
    router := NewRouter(Claude)

    tests := []struct {
        prompt   string
        expected AITool
    }{
        {"/claude explain this", Claude},
        {"/copilot suggest command", Copilot},
        {"explain the architecture", Claude},
        {"how to install npm", Copilot},
    }

    for _, tt := range tests {
        tool, _ := router.Route(tt.prompt)
        if tool != tt.expected {
            t.Errorf("Route(%q) = %v, want %v", tt.prompt, tool, tt.expected)
        }
    }
}
```

### Integration Tests
```go
// tmux_integration_test.go
func TestSendToPane(t *testing.T) {
    tm := NewTmuxManager()
    err := tm.SendToPane("test-pane", "echo 'test'")
    if err != nil {
        t.Fatalf("SendToPane failed: %v", err)
    }
}
```

---

## Performance Considerations

### Optimization
- ✅ Use Bubble Tea's high-performance rendering for large chat histories
- ✅ Lazy-load history from SQLite (paginated)
- ✅ Debounce fuzzy search input
- ✅ Cache AI tool availability checks
- ✅ Use viewport for efficient scrolling

### Resource Limits
```go
// Prevent memory bloat
const (
    MaxChatHistoryInMemory = 1000  // messages
    MaxQueueSize          = 100    // commands
    MaxWorkers            = 5      // concurrent workers
    HistoryPageSize       = 50     // DB pagination
)
```

---

## Error Handling

### Graceful Degradation
```go
func (m model) handleAIError(err error) model {
    // Log error
    log.Printf("AI tool error: %v", err)

    // Show user-friendly message
    m.chatHistory = append(m.chatHistory, ChatMessage{
        Role:    Assistant,
        Tool:    m.currentTool,
        Content: fmt.Sprintf("⚠️ Error: %s\n\nFalling back to manual mode.", err),
    })

    // Disable automatic routing
    m.autoRoute = false

    return m
}
```

---

## Future Enhancements (Backlog)

### Phase 8+ Ideas
- [ ] Web UI companion (xterm.js + WebSocket)
- [ ] Plugin system for custom AI tools
- [ ] Session recording/playback
- [ ] Collaborative mode (multiple users)
- [ ] Voice input integration (termux-speech-to-text)
- [ ] Screenshot analysis (Desktop Commander)
- [ ] Code diff visualization
- [ ] Automatic test result parsing
- [ ] GitHub integration (create issues/PRs from chat)
- [ ] Metrics dashboard (token usage, costs)

---

## Success Criteria

**Minimum Viable Product (MVP):**
- ✅ Chat with Claude Code in compact TUI
- ✅ Command queue with manual approval
- ✅ Send commands to tmux panes
- ✅ History search
- ✅ Works in Termux narrow pane

**Full Feature Set:**
- ✅ Multi-tool routing (Claude/Copilot/Codex)
- ✅ Worker management with safety
- ✅ Pattern matching suggestions
- ✅ Capability-based security
- ✅ Emergency controls
- ✅ Session persistence

---

## Development Timeline

**Day 1**: Core chat interface (Phase 1)
**Day 2**: AI routing (Phase 2)
**Day 3**: Command queue (Phase 3)
**Day 4**: Tmux integration (Phase 4)
**Day 5**: Pattern matching & history (Phase 5)
**Day 6**: Worker management & safety (Phase 6)
**Day 7**: Polish & testing (Phase 7)

**Total**: ~1 week for MVP, 2 weeks for full feature set

---

## References

- **Bubble Tea Tutorial**: https://leg100.github.io/en/posts/building-bubbletea-programs/
- **Bubbles Components**: https://github.com/charmbracelet/bubbles
- **Glamour Markdown**: https://github.com/charmbracelet/glamour
- **gotmux Library**: https://github.com/GianlucaP106/gotmux
- **go-fuzzyfinder**: https://github.com/ktr0731/go-fuzzyfinder
- **SQLite Best Practices**: https://jacob.gold/posts/go-sqlite-best-practices/
- **Terminal IRC Client Example**: https://sngeth.com/go/terminal/ui/bubble-tea/2025/08/17/building-terminal-ui-with-bubble-tea/

---

## Getting Started Tomorrow

### Quick Start Checklist
1. ☐ Create new repo: `ai-router`
2. ☐ Initialize Go module: `go mod init github.com/yourusername/ai-router`
3. ☐ Install dependencies: `go get` for all libraries above
4. ☐ Copy TFE's modular architecture (use as template)
5. ☐ Start with Phase 1: Basic chat interface
6. ☐ Test in tmux narrow pane early (validate compact mode)
7. ☐ Commit frequently (learned from Windows disaster!)

### First Commit (Scaffold)
```bash
# Create directory structure
mkdir -p ai-router
cd ai-router

# Initialize
go mod init github.com/yourusername/ai-router
touch main.go types.go styles.go model.go update.go view.go render_chat.go

# Install core deps
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/bubbles
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/glamour

# First commit
git init
git add .
git commit -m "Initial scaffold: AI Router TUI with Bubble Tea"
git push
```

---

## Notes

- **Voice-to-text workflow**: Textarea component supports multiline paste, perfect for voice input
- **Compact mode critical**: Test in narrow pane early (30% tmux split)
- **Safety first**: Implement capability limits from day 1 (learned from test suite bypass!)
- **Modular architecture**: Follow TFE's example (keep main.go minimal)
- **Commit/push constantly**: Protect against config corruption disasters
- **Docker for PC**: Always use containers when experimenting with multi-agent

---

**This plan provides a clear roadmap from basic chat interface to full AI orchestration system with proper guardrails.** Start with Phase 1 tomorrow and iterate! 🚀
