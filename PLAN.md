# TFE Development Plan

**Project:** TFE - Terminal File Explorer
**Language:** Go
**Framework:** Bubbletea + Lipgloss
**Target Users:** Windows→WSL developers, Claude Code users, terminal power users
**Updated:** 2025-10-15

---

## Project Vision

TFE is a modern, beginner-friendly terminal file explorer with a **unique Context Visualizer** that shows what Claude Code sees when launched from any directory. Unlike traditional file managers (Midnight Commander, ranger, yazi), TFE bridges Windows and Linux concepts while providing deep integration with AI coding workflows.

### Unique Value Propositions

1. **Context Visualizer** - The only tool that shows Claude Code's complete context hierarchy
   - Which files are in context
   - Token usage per file
   - CLAUDE.md hierarchy chain
   - Settings precedence visualization
   - Optimization suggestions

2. **Windows→Linux Bridge** - Explains Linux concepts in Windows terms
   - "Shortcut (symlink)" not just "symlink"
   - Visual permissions editor
   - Plain English command translation

3. **Hybrid Approach** - Like Midnight Commander
   - Native dual-pane for fast browsing/preview
   - External editor (Micro/nano) for full-featured editing
   - Best of both worlds

---

## Architecture Decisions

### Core Technology Stack

- **Language:** Go 1.24+
- **TUI Framework:** Bubbletea (proven, excellent for building complex TUIs)
- **Styling:** Lipgloss (for layout and styling)
- **Components:** Bubbles (textarea for preview)
- **Editor Integration:** External (Micro/nano via shell execution)

### Layout Strategy: Hybrid Approach (CONFIRMED)

**Native Dual-Pane for Preview:**
```
┌─────────────────────────────────────────────────────────────┐
│ TFE - Terminal File Explorer                                │
├────────────────────────┬────────────────────────────────────┤
│ 📁 Left Pane (40%)    │ 📄 Right Pane (60%)                │
│ File Tree              │ Preview / Context View             │
│                        │                                    │
│ Navigate files         │ Quick preview                      │
│ Browse directories     │ Context analysis                   │
│ Toggle views           │ Token breakdown                    │
└────────────────────────┴────────────────────────────────────┘
```

**External Editor for Editing:**
- Press `E` → Suspend TFE → Launch Micro → Resume TFE
- Full-featured editing without reinventing the wheel
- Fallback to nano if Micro not installed

**Why This Approach:**
- ✅ Works without tmux (portable)
- ✅ Fast integrated preview
- ✅ Professional editing experience
- ✅ Supports Context Visualizer (needs native panes)
- ✅ Proven pattern (Midnight Commander uses this)

---

## Development Roadmap

### Phase 1: Enhanced Single-Pane ✅ COMPLETE

**Goal:** Improve foundation before adding dual-pane

**Status:** All features implemented and tested

**Features Implemented:**
1. File metadata display
   - Size (formatted: 2.3KB, 1.5MB)
   - Modified time (relative: "5m ago", "2h ago")
   - Permissions in status bar
   - Add fields to `fileItem` struct

2. Better file type icons ✅
   - Extension-based icon mapping
   - Categories: code, configs, images, archives, docs
   - ASCII markers: `[GO]`, `[JS]`, `[MD]`, `[JSON]`, etc.
   - Special markers for folders: `▸` and parent: `↑`
   - Claude Code context files highlighted in orange (CLAUDE.md, .claude/)

3. Toggle hidden files ✅
   - Keybinding: `.` or `Ctrl+H`
   - Dynamic toggle with `showHidden bool` in model
   - Status bar indicates when showing hidden files

4. Status/Info bar ✅
   - Format: `3 folders, 12 files • showing hidden | Selected: config.js (2.3KB, 5m ago)`
   - Shows file count, selected file info with size and relative time
   - Hidden files indicator

5. Window resize handling ✅
   - Track terminal width and height
   - Respond to `tea.WindowSizeMsg`
   - Ready for dual-pane layout calculations

**Implementation Notes:**
```go
// Enhanced fileItem struct
type fileItem struct {
    name     string
    path     string
    isDir    bool
    size     int64      // NEW
    modTime  time.Time  // NEW
    mode     os.FileMode // NEW
}

// Model additions
type model struct {
    // ... existing fields
    width      int  // NEW - terminal width
    showHidden bool // NEW - toggle dotfiles
}
```

**Phase 1 Achievements:**
- Fully functional file browser with metadata
- Portable ASCII icons (work in any terminal)
- Smart file type detection
- Claude context file highlighting
- Smooth navigation and window resize handling

---

### Phase 1.5: View Modes (Optional Enhancement - 1 week)

**Goal:** Add multiple view modes inspired by Windows Explorer

**Motivation:** Windows Explorer offers different ways to view files (list, icons, details, tree). Adding view modes would make TFE more versatile and familiar to Windows users.

**Features:**
1. **List View** (current default)
   - One file per line
   - Vertical scrolling
   - Shows icon/marker, filename

2. **Grid/Icon View**
   - Multiple columns (responsive to terminal width)
   - Icon-focused display
   - Like Windows Explorer "Medium/Large icons" view
   - Example layout:
   ```
   ┌──────────┬──────────┬──────────┬──────────┐
   │ ▸ docs   │ ▸ src    │ [GO]     │ [MD]     │
   │          │          │ main.go  │ README   │
   ├──────────┼──────────┼──────────┼──────────┤
   │ [JSON]   │ [MD]     │ • file   │          │
   │ go.mod   │ PLAN.md  │          │          │
   └──────────┴──────────┴──────────┴──────────┘
   ```

3. **Detail View**
   - Columns: Name | Size | Modified | Type
   - Sortable by column
   - Like Windows Explorer "Details" view
   - Example:
   ```
   Name          Size    Modified    Type
   ▸ docs        -       1h ago      Folder
   ▸ src         -       2d ago      Folder
   [GO] main.go  2.3KB   5m ago      Go Source
   [MD] README   1.8KB   1d ago      Markdown
   ```

4. **Tree View**
   - Hierarchical folder structure
   - Expandable/collapsible folders
   - Like Windows Explorer left sidebar
   - Example:
   ```
   ▾ home
     ▾ matt
       ▾ projects
         ▾ TFE
           ▸ docs
           ▸ src
           • main.go
           • README.md
   ```

**View Mode Keybindings:**
- `v` or `Tab` - Cycle through view modes
- Or use numbers: `1` = list, `2` = grid, `3` = detail, `4` = tree
- Current mode shown in status bar

**Implementation Notes:**
```go
type displayMode int
const (
    modeList displayMode = iota
    modeGrid
    modeDetail
    modeTree
)

type model struct {
    // ... existing fields
    displayMode displayMode
    gridColumns int  // Calculated based on terminal width
    sortBy      string // For detail view: "name", "size", "modified"
    sortAsc     bool
}

// Grid layout calculation
func (m model) calculateGridLayout() (columns int) {
    itemWidth := 12 // Estimated width per item
    return max(1, m.width / itemWidth)
}
```

**Why This is Valuable:**
- Makes TFE feel familiar to Windows users
- Different views optimize for different tasks:
  - List: Fast navigation
  - Grid: Visual browsing (lots of files)
  - Detail: Finding files by size/date
  - Tree: Understanding folder structure
- Differentiates TFE from other terminal file managers
- Can combine with dual-pane (Phase 2): tree on left, grid/list on right

---

### Phase 2: Native Dual-Pane Preview (2-3 weeks)

**Goal:** Split-pane layout with integrated file preview

**Features:**
6. Split-pane layout
   - Toggle with `Space` or `Tab`
   - Responsive 40/60 split (configurable)
   - Left: File tree, Right: Preview/Context
   - Use Lipgloss for horizontal layout

7. File preview pane
   - Use `bubbles/textarea` (read-only mode)
   - Line numbers
   - Syntax highlighting (basic - by extension)
   - Scroll support
   - File metadata in header

8. External editor integration
   - `E` → Edit in Micro
   - `N` → Edit in nano (fallback)
   - Suspend TFE, launch editor, resume on exit
   - Auto-detect available editors
   - Show message if neither installed

9. Quick viewer
   - `V` → View with `bat` or `less`
   - Good for logs, large files
   - Fallback to `cat` if others unavailable

**Implementation Notes:**
```go
type viewMode int
const (
    modeSinglePane viewMode = iota
    modeDualPane
    modeContextView
)

type previewModel struct {
    textarea   textarea.Model
    filePath   string
    content    string
    scrollPos  int
    lineCount  int
}

type model struct {
    // ... existing fields
    viewMode   viewMode
    leftWidth  int
    rightWidth int
    preview    previewModel
}

// Layout calculation
func (m model) calculateLayout() (leftWidth, rightWidth int) {
    if m.viewMode == modeSinglePane {
        return m.width, 0
    }
    leftWidth = m.width * 40 / 100
    rightWidth = m.width - leftWidth
    return
}

// External editor launch
func (m model) launchEditor(path string) tea.Cmd {
    return tea.ExecProcess(exec.Command("micro", path), func(err error) tea.Msg {
        // Refresh file list on return
        return refreshMsg{}
    })
}
```

**Key Bindings:**
- `Space` / `Tab` - Toggle dual-pane mode
- `Enter` - Preview file (dual-pane) / Navigate directory
- `E` - Edit in Micro
- `N` - Edit in nano
- `V` - Quick view with bat/less
- `Esc` - Close preview / Back to file tree
- `Tab` (in dual-pane) - Switch focus left/right

---

### Phase 3: Context Visualizer (2-3 weeks) ⭐ KILLER FEATURE

**Goal:** Show what Claude Code sees from current directory

**Features:**
10. Context file analyzer
    - Press `C` to toggle Context View mode
    - Show all project files with token counts
    - Visual status indicators:
      - ✅ Included in context
      - ❌ Excluded (.gitignore, .claudeignore)
      - ⚠️ Too large (preview only)
      - 🔲 Optional (available but not auto-loaded)
    - Display in right pane (dual-pane layout)
    - Integrates with `/context` command data

11. Config hierarchy viewer
    - Press `Shift+C` for hierarchy view
    - Walk up directory tree from current path
    - Find all CLAUDE.md and CLAUDE.local.md files
    - Show active settings files with precedence:
      1. Enterprise managed-settings.json
      2. Local project .claude/settings.local.json
      3. Shared project .claude/settings.json
      4. User global ~/.claude/settings.json
    - Display as tree with token counts

12. Token counter & optimizer
    - Estimate tokens per file (~4 chars = 1 token)
    - Recursive directory token counting
    - Show total: `Token usage: 45K / 200K (22%)`
    - Suggest files/folders to exclude
    - Calculate token savings
    - Generate .claudeignore entries
    - "Add to .claudeignore" action

**Implementation Notes:**
```go
// Context analysis types
type contextStatus int
const (
    statusIncluded contextStatus = iota
    statusExcluded
    statusTooLarge
    statusOptional
)

type contextFile struct {
    file          fileItem
    status        contextStatus
    tokens        int
    excludeReason string // e.g., ".gitignore", "binary", "too large"
}

type contextModel struct {
    files         []contextFile
    totalTokens   int
    maxTokens     int // Claude's limit: 200K
    suggestions   []string
}

// Hierarchy types
type memoryFile struct {
    path   string
    tokens int
    level  int // Distance from current dir
}

type settingsFile struct {
    path       string
    precedence int // 1=highest, 5=lowest
    active     bool
}

type hierarchyModel struct {
    memoryFiles   []memoryFile
    settingsFiles []settingsFile
    currentPath   string
}

// Token estimation
func estimateTokens(content string) int {
    // Rough estimate: ~4 characters = 1 token
    return len(content) / 4
}

// Directory walker (upward)
func walkUpForClaudeFiles(startPath string) []memoryFile {
    var files []memoryFile
    current := startPath
    level := 0

    for current != "/" {
        // Check for CLAUDE.md
        claudePath := filepath.Join(current, "CLAUDE.md")
        if _, err := os.Stat(claudePath); err == nil {
            content, _ := os.ReadFile(claudePath)
            files = append(files, memoryFile{
                path:   claudePath,
                tokens: estimateTokens(string(content)),
                level:  level,
            })
        }

        // Check for CLAUDE.local.md
        localPath := filepath.Join(current, "CLAUDE.local.md")
        if _, err := os.Stat(localPath); err == nil {
            content, _ := os.ReadFile(localPath)
            files = append(files, memoryFile{
                path:   localPath,
                tokens: estimateTokens(string(content)),
                level:  level,
            })
        }

        current = filepath.Dir(current)
        level++
    }

    return files
}

// Ignore file parser
func parseGitignore(path string) ([]string, error) {
    // Parse .gitignore patterns
    // Return list of glob patterns to exclude
}

func parseClaudeIgnore(path string) ([]string, error) {
    // Parse .claudeignore patterns
    // Return list of glob patterns to exclude
}

func isFileExcluded(filePath string, patterns []string) bool {
    // Check if file matches any ignore pattern
    // Use filepath.Match for glob pattern matching
}
```

**Context View Display:**
```
┌─ Context Analysis ───────────────────────────────────────────┐
│ Token Usage: 45.2K / 200K (22.6%)                            │
├──────────────────────────────────────────────────────────────┤
│ File                          Size    Tokens   Status         │
│ ✅ main.go                    2.1KB   2,100    Included       │
│ ✅ README.md                  1.8KB   1,800    Included       │
│ ✅ docs/PLANNING.md           8.5KB   8,500    Included       │
│ ✅ go.mod                     0.3KB     300    Included       │
│ ❌ go.sum                     2.1KB     -      .gitignore     │
│ ❌ .git/                      -        -       Hidden         │
│ ⚠️  docs/RESEARCH.md          25KB    25,000   Preview only   │
│                                                               │
│ 💡 Optimization Suggestions:                                  │
│   • Exclude docs/ to save ~33K tokens                        │
│   • Add *.sum to .claudeignore to exclude checksums          │
│                                                               │
│ [I] Add to .claudeignore  [Enter] Preview  [Esc] Close       │
└──────────────────────────────────────────────────────────────┘
```

**Hierarchy View Display:**
```
┌─ Claude Code Context Hierarchy ──────────────────────────────┐
│                                                               │
│ 📋 Memory Chain (CLAUDE.md files loaded)                     │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ / (root)                                   [not loaded] │ │
│ │ └─ /home                                   [no file]    │ │
│ │    └─ /home/matt                                        │ │
│ │       ✅ CLAUDE.md                   1.2K tokens        │ │
│ │       └─ /home/matt/projects          [no file]        │ │
│ │          └─ /home/matt/projects/TFE  ← You are here    │ │
│ │             ✅ CLAUDE.md              0.8K tokens       │ │
│ │             ├─ .claude/                                 │ │
│ │             │  ⚙️  settings.json      (active)         │ │
│ │             │  ⚙️  settings.local.json (active)        │ │
│ │             └─ docs/                                    │ │
│ │                ✅ CLAUDE.md           0.5K tokens       │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                               │
│ ⚙️  Active Settings (precedence order)                       │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ 1. [none] Enterprise managed-settings.json              │ │
│ │ 2. ✅ .claude/settings.local.json (personal)            │ │
│ │ 3. ✅ .claude/settings.json (team)                      │ │
│ │ 4. ✅ ~/.claude/settings.json (global)                  │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                               │
│ 📊 Total memory: 2.5K tokens                                 │
│ [Enter] Preview file  [Esc] Close                            │
└───────────────────────────────────────────────────────────────┘
```

---

### Phase 4: Integrated Editor (Optional - 2-3 weeks)

**Goal:** Built-in editing if we want more control

**Note:** This phase is optional. The hybrid approach (external editor) may be sufficient. Only implement if:
- Users request integrated editing
- Want more control over editor features
- Need custom Claude Code integration in editor

**Features:**
13. Built-in editor mode
    - Switch from preview to edit mode
    - Use `bubbles/textarea` for editing
    - `Ctrl+S` to save
    - Track modifications
    - Confirm before closing unsaved

14. Editor enhancements
    - Undo/redo support
    - Find/replace
    - Basic syntax highlighting (via chroma library)
    - Auto-save option
    - Status indicators

---

### Phase 5: File Operations (2 weeks)

**Goal:** Make TFE a functional file manager

**Features:**
15. File operations menu
    - Press `D` for operations menu
    - Copy, Move, Rename, Delete
    - Confirmation dialogs before destructive operations
    - Progress indicators for long operations

16. Safe delete with trash
    - Move to trash instead of permanent delete
    - Option to permanently delete if needed
    - Undo capability
    - Trash management view

17. Create new file/folder
    - Press `N` for new menu
    - Input dialogs with Bubbles text input
    - Validation (no overwrite without confirm)
    - Templates for common file types

18. Batch operations
    - Mark multiple files with `Space`
    - Visual indication of marked files
    - Apply operation to all marked files
    - Confirm with file count

**Implementation Notes:**
```go
type operation int
const (
    opCopy operation = iota
    opMove
    opDelete
    opRename
)

type operationModel struct {
    op          operation
    sources     []string
    destination string
    inProgress  bool
    error       error
}

// Trash directory
var trashDir = filepath.Join(os.Getenv("HOME"), ".local/share/Trash/files")
```

---

### Phase 6: Windows-Friendly Features (1-2 weeks)

**Goal:** Bridge Windows and Linux concepts

**Features:**
19. Dual terminology mode
    - Toggle to show Windows-equivalent terms
    - "Shortcut (symlink)" not just "symlink"
    - "Properties (chmod)" not just "permissions"
    - Tooltip explanations
    - Help overlay explaining concepts

20. Visual permissions editor
    - Open with `P` on selected file
    - Checkbox UI for Owner/Group/Others
    - Read/Write/Execute options
    - Show octal notation (644, 755)
    - Show Windows equivalent description
    - Apply changes with confirmation

21. Plain English command helper
    - Show equivalent terminal commands
    - "This will run: chmod +x filename"
    - Learn while using
    - Command history/examples

**Visual Permissions Editor:**
```
┌─ Properties: main.go ─────────────────────────────────────┐
│                                                            │
│ Owner (you):  [x] Read  [x] Write  [ ] Execute           │
│ Group:        [x] Read  [ ] Write  [ ] Execute           │
│ Others:       [x] Read  [ ] Write  [ ] Execute           │
│                                                            │
│ Linux notation:   rw-r--r-- (644)                         │
│ Windows equiv:    Read-only for others                    │
│                                                            │
│ This is a regular file, not executable.                   │
│ To make executable: chmod +x main.go                      │
│                                                            │
│ [Apply] [Cancel]                                          │
└────────────────────────────────────────────────────────────┘
```

---

### Phase 7: Advanced Navigation (1-2 weeks)

**Goal:** Fast navigation and discovery

**Features:**
22. Quick filter
    - Press `/` to filter current directory
    - Type to filter (fuzzy or exact)
    - Incremental matching
    - Clear filter with `Esc`

23. Recursive file search
    - Press `Ctrl+F` for recursive search
    - Search by name in subdirectories
    - Jump to found file
    - Navigate through results

24. Bookmarks/Favorites
    - Press `B` to bookmark current directory
    - Press `Shift+B` to view bookmarks
    - Quick jump to favorites
    - Edit/delete bookmarks
    - Store in config file

25. Command palette
    - Press `:` for command mode
    - Type commands: `goto`, `search`, `bookmark`, etc.
    - Auto-complete suggestions
    - Command history

---

### Phase 8: AI Integration (Future)

**Goal:** AI-powered assistance

**Features:**
26. Ask Claude/Copilot
    - Right-click or `Ctrl+A` → Ask AI menu
    - "Explain this file"
    - "What does this directory contain?"
    - "Suggest improvements"
    - Shell out to `copilot -p` or Claude API

27. AI Scout mode
    - Pre-filter with Copilot before launching Claude
    - "Scout this directory"
    - Get quick analysis
    - Integration with context visualizer

28. Launch Claude with context
    - From Context View, press `Enter` to launch `claude`
    - Pre-optimized context based on analysis
    - Show command that will be run
    - Option to edit before launch

---

## Key Bindings Reference

### Navigation
- `↑` / `k` - Move cursor up
- `↓` / `j` - Move cursor down
- `Enter` - Preview file / Navigate into directory
- `h` / `←` - Go to parent directory
- `q` / `Esc` / `Ctrl+C` - Quit application

### View Modes
- `Space` / `Tab` - Toggle dual-pane mode
- `C` - Context analyzer view
- `Shift+C` - Context hierarchy view
- `.` / `Ctrl+H` - Toggle hidden files

### File Operations
- `E` - Edit in Micro
- `N` - Edit in nano
- `V` - Quick view (bat/less)
- `D` - File operations menu
- `P` - Properties/Permissions editor
- `Space` (in file list) - Mark file for batch operation

### Search & Navigation
- `/` - Quick filter
- `Ctrl+F` - Recursive search
- `B` - Bookmark current directory
- `Shift+B` - View bookmarks
- `:` - Command palette

### Context View (when active)
- `I` - Add file to .claudeignore
- `O` - Show optimization suggestions
- `Enter` - Preview file content
- `Esc` - Return to file browser

---

## Technical Implementation Details

### Project Structure

```
TFE/
├── main.go                  # Entry point, main model
├── go.mod                   # Dependencies
├── go.sum                   # Checksums
├── README.md                # User documentation
├── PLAN.md                  # This file
├── .claude/
│   ├── settings.json        # TFE project settings
│   └── settings.local.json  # Local overrides
├── docs/
│   └── RESEARCH.md          # Background research notes
├── internal/               # Internal packages
│   ├── fileops/           # File operations
│   ├── context/           # Context analyzer
│   ├── preview/           # Preview pane
│   ├── editor/            # Editor integration
│   ├── icons/             # Icon mapping
│   └── layout/            # Layout management
└── pkg/                    # Public packages (if any)
```

### Module Breakdown

**main.go** - Core application
- Main model and update loop
- Key binding handlers
- View rendering coordination
- Mode switching logic

**internal/layout/** - Layout management
- Calculate pane widths
- Responsive layout
- Split-pane rendering

**internal/preview/** - Preview pane
- Load file content
- Syntax detection
- Scroll handling
- Line numbers

**internal/context/** - Context analyzer
- Token counting
- Ignore file parsing
- Hierarchy walker
- Optimization suggestions

**internal/fileops/** - File operations
- Copy, move, delete, rename
- Trash integration
- Batch operations
- Progress tracking

**internal/editor/** - Editor integration
- Launch external editors
- Detect available editors
- Process management

**internal/icons/** - Icon mapping
- Extension to icon mapping
- File type detection
- Nerd Font icon database

---

## Dependencies

### Current (from go.mod)
```go
require (
    github.com/charmbracelet/bubbletea v1.3.10
    github.com/charmbracelet/lipgloss v1.1.0
    github.com/charmbracelet/bubbles v0.21.0
)
```

### Additional Dependencies Needed

**Phase 2 (Preview):**
```go
// Already included in bubbles
github.com/charmbracelet/bubbles/textarea
```

**Phase 3 (Context Visualizer):**
```go
// For .gitignore parsing
github.com/go-git/go-git/v5
// OR use filepath.Match for simpler implementation
```

**Phase 4 (Integrated Editor - Optional):**
```go
// Syntax highlighting
github.com/alecthomas/chroma/v2
```

**Phase 5 (File Operations):**
```go
// For safe file operations
github.com/otiai10/copy  // recursive copy
```

---

## Configuration

### User Configuration File
**Location:** `~/.claude/tfe-settings.json`

```json
{
  "tfe": {
    "editor": "micro",
    "fallback_editor": "nano",
    "viewer": "bat",
    "fallback_viewer": "less",
    "show_hidden_by_default": false,
    "dual_pane_by_default": false,
    "left_pane_width_percent": 40,
    "show_line_numbers": true,
    "theme": "default",
    "nerd_fonts_enabled": true,
    "use_tmux_for_edit": false,
    "max_preview_size_kb": 1024,
    "token_estimate_divisor": 4,
    "bookmarks": [
      "/home/matt/projects",
      "/home/matt/Documents"
    ]
  }
}
```

### Project Configuration
**Location:** `.claude/tfe-settings.json` (in project root)

```json
{
  "tfe": {
    "exclude_patterns": [
      "node_modules/",
      "*.pyc",
      "__pycache__/"
    ],
    "context_optimization_enabled": true,
    "auto_generate_claudeignore": false
  }
}
```

---

## Testing Strategy

### Manual Testing Checklist

**Phase 1 (Enhanced Single-Pane):** ✅ ALL COMPLETE
- [x] File metadata displays correctly
- [x] Icons match file types (ASCII markers)
- [x] Hidden files toggle works (`.` key)
- [x] Status bar updates accurately
- [x] Window resize handled properly
- [x] Claude context files highlighted in orange

**Phase 2 (Dual-Pane):**
- [ ] Toggle between single/dual pane
- [ ] Preview shows file content
- [ ] External editor launches and returns
- [ ] Fallback editor works if primary missing
- [ ] Quick viewer works

**Phase 3 (Context Visualizer):**
- [ ] Token counts are reasonable
- [ ] .gitignore patterns respected
- [ ] .claudeignore patterns respected
- [ ] Hierarchy walks up correctly
- [ ] Settings precedence shown correctly
- [ ] Optimization suggestions helpful

**Phase 5 (File Operations):**
- [ ] Copy operation works
- [ ] Move operation works
- [ ] Delete moves to trash
- [ ] Rename updates correctly
- [ ] Batch operations process all files

### Edge Cases to Test
- Very long file paths
- Files with special characters in names
- Very large files (preview)
- Very deep directory hierarchies (context)
- Permissions errors
- No Micro/nano installed
- Terminal resize during operation

---

## Performance Considerations

### Optimization Priorities

**Phase 1-2:**
- Fast file listing (cache directory reads)
- Responsive UI (never block on file I/O)
- Efficient rendering (only redraw changed panes)

**Phase 3:**
- Lazy token counting (calculate on demand)
- Cache token estimates (file hash + estimate)
- Limit hierarchy walk depth (configurable max)
- Background processing for large directories

**Phase 5:**
- Progress indicators for slow operations
- Cancel capability for long-running ops
- Batch operation chunking

### Memory Management
- Stream large files for preview (don't load entirely)
- Limit preview to first N lines for huge files
- Clean up resources when switching modes

---

## Success Metrics

### Phase 1-2 (Usable File Manager)
- ✅ Can browse files faster than `ls`
- ✅ Preview files without leaving TFE
- ✅ Edit files comfortably
- ✅ Use it daily instead of `cd` + `ls` + `micro`

### Phase 3 (Unique Value)
- ✅ Shows context that `/context` command doesn't
- ✅ Saves time debugging context issues
- ✅ Helps optimize token usage
- ✅ Understand Claude Code's view instantly

### Phase 5+ (Full-Featured)
- ✅ Replace `cp`, `mv`, `rm` commands with TFE
- ✅ Daily driver for file management
- ✅ Windows users understand Linux concepts better
- ✅ Portfolio-worthy project

---

## Milestones & Timeline

### Milestone 1: Usable File Manager ✅ IN PROGRESS
- ✅ Phase 1: Enhanced single-pane (COMPLETE)
- 🔄 Phase 1.5: View modes (OPTIONAL - under consideration)
- ⏭️ Phase 2: Dual-pane preview + editor integration (NEXT)
- **Goal:** Better than `ls`, can view/edit files

### Milestone 2: Unique Value Proposition (5-7 weeks)
- Add Phase 3: Context Visualizer
- **Goal:** The only tool showing Claude context breakdown

### Milestone 3: Full-Featured (7-9 weeks)
- Add Phase 5: File Operations
- **Goal:** Daily driver file manager

### Milestone 4: Polish & Differentiation (10-13 weeks)
- Phase 6: Windows-friendly features
- Phase 7: Advanced navigation
- **Goal:** Unique, polished, portfolio-worthy

---

## Open Questions & Decisions

### Resolved
- ✅ Hybrid approach confirmed (native dual-pane + external editor)
- ✅ Context Visualizer as killer feature
- ✅ Go + Bubbletea as tech stack
- ✅ Not relying on tmux (portable)

### To Decide
- [ ] Default keybindings (vim-like vs arrow keys priority)
- [ ] Theme support priority (Phase 6 or later?)
- [ ] Syntax highlighting in preview (basic or advanced?)
- [ ] Configuration format (JSON vs TOML vs YAML)
- [ ] Build/release process (binaries, package managers)

---

## Future Enhancements (Post-MVP)

### Potential Features
- Multiple tabs (manage multiple directories)
- Git integration (show git status in file list)
- Archive browsing (browse .zip, .tar.gz contents)
- FTP/SFTP support (browse remote files)
- Plugin system (extend with custom features)
- Themes and color schemes
- Mouse support enhancements
- File diff viewer
- Hex viewer for binary files

### Community Feedback
- Gather user feedback after Phase 3
- Prioritize based on actual usage
- Consider Windows-specific features
- Evaluate AI integration depth

---

## Resources & References

### Documentation
- [Bubbletea Tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
- [Lipgloss Examples](https://github.com/charmbracelet/lipgloss/tree/master/examples)
- [Bubbles Components](https://github.com/charmbracelet/bubbles)
- [Claude Code Docs](https://docs.claude.com/en/docs/claude-code)

### Inspiration (Existing File Managers)
- [Midnight Commander](https://github.com/MidnightCommander/mc) - Classic dual-pane
- [Yazi](https://github.com/sxyazi/yazi) - Modern Rust TUI
- [Ranger](https://github.com/ranger/ranger) - Python, powerful
- [Felix](https://github.com/kyoheiu/felix) - Simple Rust
- [Superfile](https://github.com/yorukot/superfile) - Eye-candy UI

### Tools We Integrate With
- [Micro](https://github.com/zyedidia/micro) - Modern terminal editor
- [Bat](https://github.com/sharkdp/bat) - Enhanced cat with syntax highlighting
- [Claude Code](https://github.com/anthropics/claude-code) - AI coding assistant

---

## Notes

- Focus on shipping Phase 1-2 first (usable file manager)
- Context Visualizer (Phase 3) is the differentiator - perfect it
- Don't over-engineer early phases
- Get user feedback after each milestone
- Keep it fast and responsive
- Windows-friendly features make it unique vs ranger/yazi
- Documentation as important as features

---

**Last Updated:** 2025-10-15
**Status:** Phase 1 Complete ✅
**Next Step:** Begin Phase 2 (Dual-Pane Preview) or implement view modes
