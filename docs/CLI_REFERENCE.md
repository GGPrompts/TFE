# CLI Quick Reference

> Compact command reference for Claude Code, Codex, and TUI tools

## Claude Code

### Basic Commands
```bash
claude                          # Launch interactive session
claude "your query"             # Start with prompt
claude -p "query"               # Print mode (non-interactive)
claude -c / --continue          # Resume last conversation
claude -r "<session-id>" "msg"  # Resume specific session
claude update                   # Update to latest version
claude mcp                      # Configure MCP servers
```

### Essential Flags

**Execution:**
- `--dangerously-skip-permissions` - YOLO mode (bypass all prompts)
- `--continue` / `-c` - Load most recent conversation
- `--print` / `-p` - Non-interactive, print and exit

**Models:**
- `--model sonnet` - Sonnet (default)
- `--model opus` - Most capable
- `--model haiku` - Fastest/cheapest

**Output:**
- `--output-format text` - Plain text (default)
- `--output-format json` - Structured JSON
- `--output-format stream-json` - Streaming JSON (for CI/CD)

**Customization:**
- `--append-system-prompt "text"` - Add custom system instructions
- `--max-turns N` - Limit agentic iterations
- `--verbose` - Detailed debug logging
- `--add-dir /path` - Additional working directories

**Tool Control:**
- `--allowedTools "Read,Write,Bash"` - Auto-approve these tools
- `--disallowedTools "WebFetch"` - Block specific tools
- `--permission-mode <mode>` - Set initial permission level

### Common Workflows
```bash
# CI/CD lint fix
claude -p "Fix lint errors" --output-format json

# Unattended task in Docker
claude --dangerously-skip-permissions -p "task"

# Continue with custom context
claude -c --append-system-prompt "Use Python 3.12"

# Pipe input
cat logs.txt | claude -p "Explain these errors"
```

---

## Codex (OpenAI)

### Basic Commands
```bash
codex                    # Launch interactive terminal agent
codex "your query"       # Start with prompt
codex /model             # Switch models (default: GPT-5)
codex /help              # Show help
```

### Approval Modes
- **Read-only** - Explicit approval for all actions
- **Auto** - Auto-approve workspace access, prompt for external
- **Full** - Read files anywhere, run commands with network

### Features
- Multimodal: Accept text, screenshots, diagrams
- Local execution: File operations stay on your machine
- Built in Rust: Fast and efficient
- Open source: GitHub.com/openai/codex

### Installation
```bash
# macOS/Linux (recommended)
brew install openai-codex  # or download from openai.com

# Windows (experimental - use WSL)
```

### Authentication
- ChatGPT account (Plus, Pro, Team, Edu, Enterprise)
- API key (alternative)

---

## TUI Tools

### Development
```bash
lazygit          # Git interface (staging, commits, branching)
lazydocker       # Docker container management
```

### System Monitoring
```bash
htop             # Process viewer (interactive top)
bottom           # System monitor (CPU, RAM, network)
btm              # bottom alias
```

### File & Log Viewing
```bash
lnav             # Log navigator (auto-format, search)
viu image.png    # Image viewer (terminal)
timg image.jpg   # Alternative image viewer
chafa pic.gif    # ASCII/Unicode image renderer
```

### Editors
```bash
micro file.txt   # Modern nano-like editor
nano file.txt    # Simple text editor
vim file.txt     # Vi improved
```

### Image Editing
```bash
textual-paint    # MS Paint for terminal (TFE uses this)
durdraw          # ASCII/ANSI art editor
```

### Media & Entertainment
```bash
pyradio          # Terminal internet radio player
```

### File Managers
```bash
mc               # Midnight Commander (dual-pane file manager)
```

---

## TUI Launch Commands (Detailed)

> Detailed launch commands with useful CLI flags for TUI applications

### micro (Text Editor)

**Basic Usage:**
```bash
micro file.txt                    # Open/create file
micro file1.txt file2.txt         # Open multiple files (tabs)
micro +10 file.txt                # Open at line 10
micro +10:5 file.txt              # Open at line 10, column 5
```

**Common Flags:**
```bash
micro -version                    # Show version
micro -clean                      # Fresh start (ignore config)
micro -config-dir ~/.config/micro # Custom config directory
micro -readonly file.txt          # Read-only mode
```

**Installation:**
```bash
# Linux
curl https://getmic.ro | bash
sudo mv micro /usr/local/bin/

# Snap
snap install micro --classic

# From source
git clone https://github.com/zyedidia/micro && cd micro && make build
```

### mc (Midnight Commander)

**Basic Usage:**
```bash
mc                                # Launch in current directory
mc /path/to/dir                   # Launch at specific directory
mc dir1 dir2                      # Open with dir1 (left) and dir2 (right)
```

**Common Flags:**
```bash
mc -v file.txt                    # View file (read-only)
mc -e file.txt                    # Edit file
mc -P file.txt                    # Print last working directory to file
mc -u                             # Disable mouse support
mc -c                             # Force color mode
mc -b                             # Force black & white mode
```

**Advanced Usage:**
```bash
mc -P ~/last_dir.txt              # Save last directory for quick cd
cd "$(cat ~/last_dir.txt)"        # Jump to last MC directory

# In TFE context menu, Quick CD uses this approach
```

**Key Shortcuts (once inside):**
```
F1  - Help            F7  - Create directory
F2  - Menu            F8  - Delete
F3  - View file       F9  - Top menu
F4  - Edit file       F10 - Quit
F5  - Copy            Tab - Switch panes
F6  - Move/Rename     Ctrl+O - Show/hide panels
```

### htop (Process Viewer)

**Basic Usage:**
```bash
htop                              # Launch interactive process viewer
```

**Common Flags:**
```bash
htop -d 10                        # Update delay (tenths of seconds, default 15)
htop -C                           # Monochrome mode
htop -u username                  # Show only processes of specific user
htop -p 1234,5678                 # Show only specific PIDs
htop -s PERCENT_CPU               # Sort by column (e.g., PERCENT_CPU, PERCENT_MEM)
htop -t                           # Tree view (show process hierarchy)
```

**Key Shortcuts (once inside):**
```
F1  - Help            u   - Filter by user
F2  - Setup           t   - Tree view
F3  - Search          H   - Hide/show threads
F4  - Filter          K   - Hide kernel threads
F5  - Tree view       Space - Tag process
F6  - Sort by         c   - Tag process and children
F9  - Kill process    F   - Follow process
F10 - Quit            /   - Search
```

**Installation:**
```bash
sudo apt install htop             # Debian/Ubuntu
sudo yum install htop             # RHEL/CentOS
brew install htop                 # macOS
```

### bottom (btm) - System Monitor

**Basic Usage:**
```bash
bottom                            # Launch system monitor
btm                               # Short alias
```

**Common Flags:**
```bash
bottom -b                         # Basic mode (less features, better compatibility)
bottom -t 1000                    # Temperature type (Celsius=1, Kelvin=2, Fahrenheit=3)
bottom -d                         # Show disk usage
bottom -n                         # Show network usage
bottom -r 1000                    # Refresh rate in milliseconds (default 1000)
bottom -u 1000                    # Update rate in milliseconds
bottom --hide_time                # Hide time scale
bottom --battery                  # Show battery info (laptops)
```

**Advanced Flags:**
```bash
bottom -C config.toml             # Use custom config file
bottom --default_time_value 30000 # Default time range (ms)
bottom --default_widget_type cpu  # Starting widget (cpu, mem, net, disk, temp, proc)
bottom --expanded                 # Start with expanded widgets
bottom --mem_as_value             # Show mem as values not percentage
```

**Key Shortcuts (once inside):**
```
?   - Help            m   - Memory widget
q   - Quit            n   - Network widget
/   - Search          d   - Disk widget
dd  - Kill process    t   - Temperature widget
c   - CPU widget      p   - Process widget
e   - Expand widget   +/- - Zoom time scale
```

**Installation:**
```bash
cargo install bottom              # From Rust
sudo apt install bottom           # Debian/Ubuntu (may need newer repos)
brew install bottom               # macOS
snap install bottom               # Snap

# Download binary from GitHub
# https://github.com/ClementTsang/bottom/releases
```

### lazygit (Git TUI)

**Basic Usage:**
```bash
lazygit                           # Launch in current git repo
lazygit -p /path/to/repo          # Launch in specific repo
```

**Common Flags:**
```bash
lazygit -v                        # Version info
lazygit -d                        # Debug mode
lazygit -c                        # Use custom config
lazygit -w /path/to/worktree      # Open specific git worktree
```

**Key Shortcuts (once inside):**
```
?   - Help            p   - Pull
1-5 - Switch panels   P   - Push
x   - Open menu       c   - Commit
Space - Stage/unstage f   - Fetch
a   - Stage all       n   - New branch
```

---

## Common Shell Commands

### Directory Navigation
```bash
cd -             # Previous directory
cd ~             # Home directory
z pattern        # Jump to directory (zoxide)
```

### Quick Operations
```bash
!!               # Repeat last command
sudo !!          # Repeat last command as root
history          # Command history
fc               # Fix/edit last command
```

### File Operations
```bash
ls -lah          # List all files with human sizes
tree -L 2        # Directory tree (2 levels)
du -sh *         # Disk usage summary
df -h            # Disk free space
```

---

## Platform-Specific

### WSL (Windows)
```bash
wslview file.html         # Open in Windows browser
explorer.exe .            # Open current dir in Windows Explorer
cmd.exe /c clip           # Access Windows clipboard
```

### Termux (Android)
```bash
termux-clipboard-set      # Copy to clipboard
termux-clipboard-get      # Paste from clipboard
termux-open file.html     # Open with Android app
pkg install tool          # Install packages
```

---

## Aliases & Shortcuts

### Recommended Claude Code Aliases
```bash
alias cc="claude --dangerously-skip-permissions"
alias ccc="claude --continue"
alias ccp="claude -p"
```

### TFE Integration
```bash
# Quick CD is built-in via context menu
# Right-click folder → 📂 Quick CD
```

---

**Last Updated:** 2025-10-20
**Sources:** Claude Code CLI Docs, OpenAI Codex Docs, Tool Manpages
