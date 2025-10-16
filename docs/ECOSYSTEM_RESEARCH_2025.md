# Charm/Bubbletea Ecosystem Research - Comprehensive Library Survey 2025

**Research Date:** October 16, 2025
**Purpose:** Identify libraries and tools from the Charm/Bubbletea ecosystem to enhance the TFE terminal file explorer

## Current Stack

TFE currently uses:
- **Bubbletea** v1.3.10 - TUI framework
- **Lipgloss** v1.1.1 - Styling
- **Bubbles** v0.21.0 - Components
- **Glamour** v0.10.0 - Markdown rendering
- **Chroma** v2.14.0 - Syntax highlighting
- **termenv** v0.16.0 - Terminal detection (indirect dependency)

---

## Official Charm Libraries

### 1. **Huh** - Forms and Prompts
- **GitHub:** https://github.com/charmbracelet/huh
- **Package:** `github.com/charmbracelet/huh`
- **Latest:** v2.x available
- **Status:** Production-ready

**Description:**
A simple, powerful library for building interactive forms and prompts in the terminal. Works standalone or integrates seamlessly with Bubbletea applications.

**Features:**
- Multiple field types: Select, MultiSelect, Input, Text, Confirm
- Groups/pages system for complex forms
- Accessible mode for screen readers (`.WithAccessible(true)`)
- Sensible defaults with customizable styling
- Native `tea.Model` implementation

**Use Cases for TFE:**
- ✅ **File operation dialogs** - Create, rename, move dialogs with proper forms
- ✅ **Search filters** - Multi-field search with file type, date range, size filters
- ✅ **Settings configuration** - User preferences and configuration screens
- ✅ **Batch operations** - Select multiple operations with confirmation prompts
- ✅ **Advanced filters** - Complex filter building for file views

**Priority:** HIGH - Would significantly improve user interaction quality

---

### 2. **Harmonica** - Physics-based Animation
- **GitHub:** https://github.com/charmbracelet/harmonica
- **Package:** `github.com/charmbracelet/harmonica`
- **Status:** Stable

**Description:**
A simple, efficient spring animation library for smooth, natural motion. Framework-agnostic and works well in 2D/3D contexts.

**Features:**
- Physics-based spring animations
- Smooth, natural motion curves
- Works with Bubbletea's tick system
- Examples for particles and spring effects

**Use Cases for TFE:**
- ✅ **Smooth scrolling** - Natural scroll animations for file lists
- ✅ **Panel transitions** - Smooth pane switching and resizing
- ✅ **Progress indicators** - Smooth progress bar animations
- ✅ **Hover effects** - Subtle animations for selected items
- ✅ **Loading states** - Natural spinner movements

**Priority:** MEDIUM - Nice-to-have for polish, not critical functionality

---

### 3. **Log** - Logging Library
- **GitHub:** https://github.com/charmbracelet/log
- **Package:** `github.com/charmbracelet/log`
- **Status:** Stable

**Description:**
A happy, powerful little logger designed for terminal applications. Styled output that matches Charm aesthetics.

**Use Cases for TFE:**
- ✅ **Debug mode** - User-friendly debug logs
- ✅ **Error reporting** - Formatted error messages
- ✅ **Operation logging** - Track file operations for debugging
- ✅ **Development** - Better development experience

**Priority:** LOW - Current error handling works, but this would be nice

---

### 4. **VHS** - Terminal GIF Recorder
- **GitHub:** https://github.com/charmbracelet/vhs
- **Package:** Tool (not a library)
- **Status:** Production-ready

**Description:**
Create terminal GIFs with code. Perfect for documentation, demos, and testing.

**Use Cases for TFE:**
- ✅ **Documentation** - Create animated README demos
- ✅ **Tutorial videos** - Show feature walkthroughs
- ✅ **Bug reports** - Reproduce issues with recordings
- ✅ **Marketing** - Show off features visually

**Priority:** MEDIUM - Great for project visibility and documentation

---

### 5. **Gum** - Glamorous Shell Scripts
- **GitHub:** https://github.com/charmbracelet/gum
- **Package:** CLI tool (not a library)
- **Status:** Production-ready

**Description:**
A tool for glamorous shell scripts using Bubbles and Lipgloss without writing Go.

**Use Cases for TFE:**
- ✅ **Installation scripts** - Pretty install/setup scripts
- ✅ **Build scripts** - Enhanced build/release scripts
- ✅ **CI/CD** - Better pipeline output

**Priority:** LOW - Useful for project tooling, not core features

---

### 6. **Wish** - SSH Apps Framework
- **GitHub:** https://github.com/charmbracelet/wish
- **Package:** `github.com/charmbracelet/wish`
- **Status:** Production-ready

**Description:**
Make SSH apps easily. Includes Bubbletea middleware for serving TUI apps over SSH.

**Features:**
- Built on gliderlabs/ssh (no OpenSSH needed)
- Bubbletea middleware for serving TUIs over SSH
- Self-contained SSH server
- Multiple middleware options

**Use Cases for TFE:**
- ✅ **Remote file management** - Browse files over SSH with TFE interface
- ✅ **Server administration** - File explorer for remote servers
- ✅ **Cloud integration** - Browse cloud storage with SSH tunnels
- ✅ **Multi-user support** - Shared file browsing sessions

**Priority:** MEDIUM - Interesting for remote scenarios, but not core use case

---

### 7. **Soft Serve** - Git Server with TUI
- **GitHub:** https://github.com/charmbracelet/soft-serve
- **Package:** Application (reference implementation)
- **Status:** Production-ready

**Description:**
Self-hostable Git server with first-class TUI. Great reference for advanced Bubbletea apps.

**Use Cases for TFE:**
- ✅ **Architecture reference** - Learn advanced patterns
- ✅ **Git integration ideas** - See how they handle git operations
- ✅ **Multi-pane layouts** - Study their layout system

**Priority:** LOW - Reference material, not direct integration

---

### 8. **Melt** - SSH Key Backup
- **GitHub:** https://github.com/charmbracelet/melt
- **Package:** CLI tool
- **Status:** Stable

**Description:**
Backup and restore Ed25519 SSH keys with seed words.

**Use Cases for TFE:**
- ❌ Not relevant for file explorer use case

**Priority:** N/A

---

### 9. **Skate** - Key-Value Store
- **GitHub:** https://github.com/charmbracelet/skate
- **Package:** `github.com/charmbracelet/skate`
- **Status:** Stable (v1.0.0+ local-only)

**Description:**
A personal key-value store that operates locally.

**Use Cases for TFE:**
- ✅ **Persistent state** - Store user preferences
- ✅ **Recent files** - Track recently accessed files
- ✅ **Bookmarks** - Alternative to current favorites system
- ✅ **Cache** - Cache file metadata

**Priority:** LOW - Current JSON-based config works fine

---

### 10. **Freeze** - Code Screenshots
- **GitHub:** https://github.com/charmbracelet/freeze
- **Package:** CLI tool
- **Status:** Production-ready

**Description:**
Generate code screenshots with syntax highlighting.

**Use Cases for TFE:**
- ✅ **Preview feature** - Add "freeze" integration for sharing code snippets
- ✅ **Documentation** - Beautiful code screenshots for docs

**Priority:** LOW - Nice feature but not essential

---

### 11. **charmbracelet/x** - Experimental Libraries
- **GitHub:** https://github.com/charmbracelet/x
- **Package:** `github.com/charmbracelet/x/...`
- **Status:** Experimental (no backward compatibility guarantee)

**Notable Modules:**
- `x/ansi` - ANSI parsing (already used in TFE)
- `x/term` - Terminal utilities (already used)
- `x/exp/teatest` - Testing library for Bubbletea apps
- `x/cellbuf` - Cell buffer utilities (already used)

**Use Cases for TFE:**
- ✅ **teatest** - Add proper testing for TUI components
- ✅ **Experimental features** - Try new functionality early

**Priority:** MEDIUM - teatest would be valuable for testing

---

## Community Libraries

### 12. **bubble-table** by Evertras
- **GitHub:** https://github.com/Evertras/bubble-table
- **Package:** `github.com/evertras/bubble-table/table`
- **Status:** Actively maintained

**Description:**
A customizable, interactive table component for Bubbletea with extensive features.

**Features:**
- Header, rows, footer with customizable borders
- Fixed-width and flexible columns
- Horizontal scrolling with frozen columns
- Built-in pagination with automatic footer
- Built-in filtering (keybind: `/`)
- Sorting (ascending/descending, numeric-aware)
- Events for user interactions
- Global and per-cell styling
- Column/row/cell-level customization

**Use Cases for TFE:**
- ✅ **Enhanced detail view** - Replace current detail view with interactive table
- ✅ **Sortable columns** - Click to sort by name, size, date
- ✅ **Filterable content** - Quick filter with `/` key
- ✅ **Frozen columns** - Keep filename visible while scrolling
- ✅ **Column resize** - User-adjustable column widths
- ✅ **Pagination** - Handle large directories efficiently

**Priority:** HIGH - Would dramatically improve detail view mode

**Examples in repo:**
- `/examples/features` - Full feature showcase
- `/examples/pokemon` - Pokemon table demo
- `/examples/pagination` - Pagination example
- `/examples/sorting` - Sorting demo

---

### 13. **promptkit** by erikgeiser
- **GitHub:** https://github.com/erikgeiser/promptkit
- **Package:** `github.com/erikgeiser/promptkit`
- **Status:** Active (API not stable pre-v1.0.0)

**Description:**
A Go library for common CLI prompts with sensible defaults and re-mappable keybindings.

**Features:**
- Selection with filter and pagination
- Text input prompts
- Confirmation prompts
- Can be used as standalone prompts or Bubbletea widgets
- Customizable styling and keybindings

**Note:** Windows not explicitly supported due to Bubbletea input bug

**Use Cases for TFE:**
- ✅ **Quick dialogs** - Simpler alternative to Huh for basic prompts
- ✅ **Yes/No confirmations** - Delete, overwrite confirmations
- ✅ **Selection lists** - Quick file selection from filtered lists

**Priority:** MEDIUM - Overlaps with Huh, but lighter weight

---

### 14. **bubblezone** by lrstanley
- **GitHub:** https://github.com/lrstanley/bubblezone
- **Package:** `github.com/lrstanley/bubblezone`
- **Status:** Stable

**Description:**
Helper utility for Bubbletea allowing easy mouse event tracking with zero-printable-width markers.

**How it works:**
1. Wrap components with `zone.Mark(id, content)`
2. Wrap view in `zone.Scan()` to process markers
3. Check mouse events with `zone.Get(id).InBounds(mouseEvent)`

**Features:**
- Zero-width markers (doesn't affect lipgloss.Width())
- Automatic offset tracking
- Simple mouse event checking
- Works with any component

**Use Cases for TFE:**
- ✅ **Clickable buttons** - Make UI buttons actually clickable
- ✅ **Clickable file icons** - Click icon to preview/open
- ✅ **Click-to-focus** - Click any pane to focus it (already implemented, but could improve)
- ✅ **Drag and drop** - Track mouse zones for drag operations
- ✅ **Hover states** - Detect when mouse hovers over elements
- ✅ **Clickable links** - Make path breadcrumbs clickable

**Priority:** HIGH - Would significantly improve mouse UX

---

### 15. **treeview** by Digital-Shane
- **GitHub:** https://github.com/Digital-Shane/treeview
- **Package:** `github.com/Digital-Shane/treeview`
- **Status:** Active

**Description:**
Feature-rich Go module for displaying and navigating hierarchical structures in the terminal.

**Features:**
- Full TUI support via Bubbletea
- Complete terminal file browser example
- 12 progressive examples from basic to complex
- Tree navigation with expand/collapse

**Use Cases for TFE:**
- ✅ **Replace current tree view** - More robust tree implementation
- ✅ **Better tree navigation** - Tested tree navigation patterns
- ✅ **Reference implementation** - Study their file browser example

**Priority:** MEDIUM - Current tree view works, but this might be better

---

### 16. **tree-bubble** by savannahostrowski
- **GitHub:** https://github.com/savannahostrowski/tree-bubble
- **Package:** `github.com/savannahostrowski/tree-bubble`
- **Status:** Active

**Description:**
Another TUI tree view for Bubbletea framework. Examples in `/example` directory.

**Use Cases for TFE:**
- ✅ **Alternative to Digital-Shane/treeview**
- ✅ **Compare tree implementations**

**Priority:** LOW - Evaluate if better than Digital-Shane's or current implementation

---

### 17. **go-fzf** by koki-develop
- **GitHub:** https://github.com/koki-develop/go-fzf
- **Package:** `github.com/koki-develop/go-fzf`
- **Status:** Active

**Description:**
Fuzzy finder built using Bubbletea, inspired by fzf command-line tool.

**Features:**
- `New()` creates fuzzy finder, returns `tea.Program`
- `Find()` launches finder and returns selected item indexes
- Preview window support
- Customizable item prefixes
- Fuzzy matching algorithm

**Use Cases for TFE:**
- ✅ **Fuzzy file search** - Quick file finder (Ctrl+P style)
- ✅ **Command palette** - Search commands/actions
- ✅ **Quick navigation** - Jump to any file by typing
- ✅ **Tag search** - Find files by tags/metadata
- ✅ **History search** - Search recent files

**Priority:** HIGH - Fuzzy finding is a killer feature for file explorers

**Implementation note:** Use stderr for output to avoid conflicts with normal rendering

---

### 18. **additional-bubbles** (Community Repository)
- **GitHub:** https://github.com/charm-and-friends/additional-bubbles
- **Package:** Various
- **Status:** Community catalog

**Description:**
Catalog of community-maintained bubbles for Bubbletea.

**Includes:**
- bubble-datepicker - jQuery-inspired datepicker
- promptkit (listed separately above)
- bubble-table (listed separately above)
- And more community components

**Use Cases for TFE:**
- ✅ **Discover components** - Find community solutions
- ✅ **Datepicker** - Filter files by date ranges
- ✅ **Browse patterns** - Learn from community examples

**Priority:** LOW - Reference catalog, evaluate specific components as needed

---

## Related TUI Libraries (Go)

### 19. **tview** by rivo
- **GitHub:** https://github.com/rivo/tview
- **Package:** `github.com/rivo/tview`
- **Status:** Very mature, actively maintained

**Description:**
Comprehensive TUI library with rich components: tables, forms, lists, modals, etc.

**Note:** Uses tcell directly, not Bubbletea-compatible. Would require rewrite.

**Use Cases for TFE:**
- ❌ **Not compatible** - Different framework
- ✅ **Reference** - Study their component designs
- ✅ **Patterns** - Learn layout and interaction patterns

**Priority:** LOW - Reference only, not for integration

---

### 20. **termui** by gizak
- **GitHub:** https://github.com/gizak/termui
- **Package:** `github.com/gizak/termui`
- **Status:** Mature

**Description:**
Specialized for displaying information graphically: bar charts, line diagrams, spark lines, gauges.

**Use Cases for TFE:**
- ✅ **Disk usage visualization** - Bar charts for space usage
- ✅ **File statistics** - Graphs showing file type distribution
- ✅ **Timeline view** - File modification timeline
- ✅ **Size comparison** - Visual size comparisons

**Priority:** LOW - Nice visualizations but not Bubbletea-compatible

---

### 21. **lf** (List Files) File Manager
- **GitHub:** https://github.com/gokcehan/lf
- **Language:** Go
- **Status:** Production-ready, actively maintained

**Description:**
Terminal file manager written in Go, heavily inspired by ranger.

**Features:**
- Single binary, no runtime dependencies
- Fast startup, low memory footprint
- Cross-platform (Linux, OSX, BSDs, partial Windows)
- Server/client architecture for shared selections
- Extensive customization

**Use Cases for TFE:**
- ✅ **Architecture reference** - Study Go file manager patterns
- ✅ **Feature inspiration** - See what power users expect
- ✅ **Configuration ideas** - Learn from their config system
- ✅ **Performance patterns** - Fast file loading techniques

**Priority:** MEDIUM - Excellent reference implementation

---

## Utility Libraries

### 22. **termenv** by muesli
- **GitHub:** https://github.com/muesli/termenv
- **Package:** `github.com/muesli/termenv` (already in use)
- **Current version:** v0.16.0
- **Status:** Stable

**Description:**
Advanced ANSI style & color support. Already an indirect dependency via Lipgloss.

**Features:**
- Color profile detection (Ascii, ANSI, ANSI256, TrueColor)
- Query terminal foreground/background colors
- Dark mode detection
- NO_COLOR and CLICOLOR environment variable support
- Automatic color degradation

**Current Status:**
✅ Already integrated as indirect dependency

---

### 23. **colorprofile** by charmbracelet
- **Package:** `github.com/charmbracelet/colorprofile` (already in use)
- **Current version:** v0.3.2
- **Status:** Stable

**Description:**
Terminal color profile detection library.

**Current Status:**
✅ Already integrated as indirect dependency

---

### 24. **rasterm** by BourgeoisBear
- **GitHub:** https://github.com/BourgeoisBear/rasterm
- **Package:** `github.com/BourgeoisBear/rasterm`
- **Status:** Active

**Description:**
Encode images to iTerm / Kitty / SIXEL terminal inline graphics protocols.

**Use Cases for TFE:**
- ✅ **Image preview** - Show actual images in preview pane
- ✅ **File icons** - Use actual image thumbnails
- ✅ **PDF previews** - Render PDF pages as images
- ✅ **Video thumbnails** - Show video first frames

**Priority:** MEDIUM - Great for image-heavy workflows

**Compatibility:** Works with iTerm2, Kitty, and terminals supporting Sixel

---

### 25. **timg** by hzeller
- **GitHub:** https://github.com/hzeller/timg
- **Language:** C/C++ (CLI tool)
- **Status:** Production-ready

**Description:**
Terminal image and video viewer supporting Sixel, Kitty, and iTerm2 protocols.

**Use Cases for TFE:**
- ✅ **External viewer integration** - Launch timg for image preview
- ✅ **Inline preview** - Embed timg output in preview pane
- ✅ **Video preview** - Play videos in terminal

**Priority:** LOW - External tool, could integrate via command.go

---

## Official Bubbletea Examples

The official Bubbletea repository includes excellent examples:

**Repository:** https://github.com/charmbracelet/bubbletea/tree/main/examples

### Key Examples:

1. **file-picker** (`examples/file-picker/main.go`)
   - File system navigation
   - Directory browsing
   - File selection with filtering

2. **progress-static** (`examples/progress-static/main.go`)
   - Static progress bar
   - Step-by-step incrementation

3. **progress-animated** (`examples/progress-animated/main.go`)
   - Animated progress bar
   - Smooth transitions

4. **progress-download** (`examples/progress-download/main.go`)
   - Download progress with file size
   - Network I/O progress tracking

5. **views** (`examples/views/main.go`)
   - Multiple view management
   - View switching patterns

**Use Cases for TFE:**
- ✅ **Study patterns** - Learn official patterns
- ✅ **Copy operations** - Use progress bars for file copy/move
- ✅ **View management** - Better multi-pane handling

**Priority:** HIGH - Official examples are authoritative

---

## Recommended Implementation Priorities

### Immediate High Priority (Should implement soon)

1. **bubble-table** - Transform detail view into powerful interactive table
2. **go-fzf** - Add fuzzy file finding (game changer for UX)
3. **bubblezone** - Improve mouse interaction throughout app
4. **Huh** - Replace current dialog system with proper forms

### Medium Priority (Nice to have)

5. **Official examples** - Study and adopt patterns (progress bars, etc.)
6. **VHS** - Create demo videos for documentation
7. **lf source code** - Study for architecture improvements
8. **rasterm** - Add image preview support
9. **Harmonica** - Polish animations and transitions
10. **x/exp/teatest** - Add proper TUI testing

### Low Priority (Consider later)

11. **Wish** - SSH support for remote file management
12. **Log** - Better logging/debugging
13. **treeview** - Evaluate vs. current tree implementation
14. **tview/termui** - Study for patterns only

### Not Recommended

- **Skate** - Current JSON config works fine
- **Melt** - Not relevant
- **Freeze** - Nice but not core functionality
- **Gum** - Build tooling, not app features

---

## Integration Roadmap

### Phase 1: Enhanced Interactivity (Week 1-2)
```
1. Install bubble-table
2. Refactor detail view to use bubble-table
3. Add sorting, filtering, column customization
4. Test with large directories
```

### Phase 2: Fuzzy Finding (Week 2-3)
```
1. Install go-fzf
2. Add Ctrl+P fuzzy file finder
3. Integrate with file history
4. Add command palette with fuzzy search
```

### Phase 3: Better Mouse UX (Week 3-4)
```
1. Install bubblezone
2. Add zone markers to UI elements
3. Make buttons, icons, breadcrumbs clickable
4. Implement hover states
5. Test mouse interactions
```

### Phase 4: Forms & Dialogs (Week 4-5)
```
1. Install Huh
2. Replace dialog.go with Huh forms
3. Add multi-field search form
4. Add settings configuration screen
5. Improve file operation dialogs
```

### Phase 5: Polish & Testing (Week 5-6)
```
1. Install Harmonica for animations
2. Add smooth transitions
3. Install teatest and write tests
4. Create VHS demo videos
5. Document new features
```

### Phase 6: Advanced Features (Future)
```
1. Evaluate rasterm for image previews
2. Consider Wish for SSH support
3. Add Log for debug mode
4. Explore additional-bubbles catalog
```

---

## Code Examples

### Example 1: Adding bubble-table to Detail View

```go
// In types.go
import "github.com/evertras/bubble-table/table"

type model struct {
    // ... existing fields
    fileTable table.Model
}

// In render_file_list.go
func (m model) renderDetailView(maxVisible int) string {
    if m.fileTable == nil {
        // Initialize table
        columns := []table.Column{
            table.NewColumn("icon", "Icon", 3),
            table.NewColumn("name", "Name", 40).WithFiltered(true),
            table.NewColumn("size", "Size", 10).WithFiltered(true),
            table.NewColumn("modified", "Modified", 20).WithFiltered(true),
        }

        rows := []table.Row{}
        for _, f := range m.files {
            rows = append(rows, table.NewRow(table.RowData{
                "icon": getFileIcon(f),
                "name": f.name,
                "size": formatFileSize(f.size),
                "modified": formatModTime(f.modTime),
            }))
        }

        m.fileTable = table.New(columns).
            WithRows(rows).
            WithPageSize(maxVisible).
            Focused(m.focusPane == paneFiles).
            SortByAsc("name")
    }

    return m.fileTable.View()
}
```

### Example 2: Adding go-fzf Fuzzy Finder

```go
// In update_keyboard.go
case "ctrl+p":
    // Launch fuzzy finder
    items := []string{}
    for _, f := range m.files {
        items = append(items, f.name)
    }

    fzf, err := fzf.New(
        fzf.WithItems(items),
        fzf.WithPreview(),
    )
    if err != nil {
        return m, nil
    }

    indexes, err := fzf.Find()
    if err != nil {
        return m, nil
    }

    if len(indexes) > 0 {
        m.cursor = indexes[0]
        m = m.loadPreview()
    }
    return m, nil
```

### Example 3: Adding bubblezone for Clickable Elements

```go
// In view.go
import "github.com/lrstanley/bubblezone"

var zone = bubblezone.New()

func (m model) View() string {
    // Mark home button as clickable
    homeButton := zone.Mark("home-button", "[🏠 Home]")

    // Wrap final view in zone scanner
    view := lipgloss.JoinVertical(
        lipgloss.Left,
        homeButton,
        m.renderBody(),
    )

    return zone.Scan(view)
}

// In update_mouse.go
func (m model) handleMouseEvent(msg tea.MouseMsg) (model, tea.Cmd) {
    if msg.Type == tea.MouseLeft {
        if zone.Get("home-button").InBounds(msg) {
            // Home button clicked
            m.currentPath = homeDir
            m = m.loadFiles()
        }
    }
    return m, nil
}
```

### Example 4: Adding Huh Forms for File Operations

```go
// In file_operations.go
import "github.com/charmbracelet/huh"

func (m *model) showCreateFileDialog() tea.Cmd {
    var filename string
    var fileType string

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewInput().
                Title("File Name").
                Value(&filename).
                Validate(func(s string) error {
                    if s == "" {
                        return errors.New("filename required")
                    }
                    return nil
                }),

            huh.NewSelect[string]().
                Title("Type").
                Options(
                    huh.NewOption("File", "file"),
                    huh.NewOption("Directory", "dir"),
                ).
                Value(&fileType),

            huh.NewConfirm().
                Title("Create?").
                Affirmative("Yes").
                Negative("No"),
        ),
    )

    return func() tea.Msg {
        err := form.Run()
        if err != nil {
            return errMsg{err}
        }

        // Create file or directory
        if fileType == "dir" {
            return createDirMsg{filename}
        }
        return createFileMsg{filename}
    }
}
```

---

## Dependencies Summary

### Ready to Install (No Conflicts)
- `github.com/charmbracelet/huh` (forms)
- `github.com/charmbracelet/harmonica` (animations)
- `github.com/charmbracelet/log` (logging)
- `github.com/evertras/bubble-table` (tables)
- `github.com/erikgeiser/promptkit` (prompts)
- `github.com/lrstanley/bubblezone` (mouse zones)
- `github.com/koki-develop/go-fzf` (fuzzy finder)
- `github.com/BourgeoisBear/rasterm` (images)
- `github.com/charmbracelet/x/exp/teatest` (testing)

### Already Installed (Indirect)
- `github.com/muesli/termenv` ✅
- `github.com/charmbracelet/colorprofile` ✅
- `github.com/charmbracelet/x/ansi` ✅
- `github.com/charmbracelet/x/term` ✅

### Not Compatible / Different Frameworks
- `github.com/rivo/tview` (uses tcell, not Bubbletea)
- `github.com/gizak/termui` (different framework)

---

## Performance Considerations

### bubble-table
- **Pros:** Efficient with pagination, virtual scrolling
- **Cons:** Memory usage for very large directories (10k+ files)
- **Recommendation:** Use pagination, lazy loading for huge dirs

### go-fzf
- **Pros:** Fast fuzzy matching algorithm
- **Cons:** Needs to index all files upfront
- **Recommendation:** Index in background, cache results

### bubblezone
- **Pros:** Minimal overhead, zero-width markers
- **Cons:** Full screen scan on every render
- **Recommendation:** Fine for typical use, be mindful of nested zones

### Harmonica
- **Pros:** Lightweight physics calculations
- **Cons:** Tick-based, uses CPU during animation
- **Recommendation:** Use sparingly, disable on slow terminals

---

## Testing Strategy with teatest

```go
// file_operations_test.go
package main

import (
    "testing"
    "github.com/charmbracelet/x/exp/teatest"
)

func TestFileNavigation(t *testing.T) {
    tm := teatest.NewTestModel(
        t, initialModel(),
        teatest.WithInitialTermSize(80, 24),
    )

    // Test navigation down
    tm.Send(tea.KeyMsg{Type: tea.KeyDown})
    tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))

    // Assert cursor moved
    finalModel := tm.FinalModel(t).(model)
    if finalModel.cursor != 1 {
        t.Errorf("expected cursor at 1, got %d", finalModel.cursor)
    }
}

func TestFuzzyFinder(t *testing.T) {
    tm := teatest.NewTestModel(t, initialModel())

    // Open fuzzy finder
    tm.Send(tea.KeyMsg{Type: tea.KeyCtrlP})

    // Type search
    tm.Type("readme")

    // Select result
    tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

    // Check file selected
    finalModel := tm.FinalModel(t).(model)
    if finalModel.files[finalModel.cursor].name != "README.md" {
        t.Error("fuzzy finder did not select README.md")
    }
}
```

---

## Community Resources

### Discord
- **Charm Community Discord:** https://charm.sh/discord
- Active community for questions and support

### GitHub Discussions
- **Bubbletea Discussions:** https://github.com/charmbracelet/bubbletea/discussions
- **Bubbles Discussions:** https://github.com/charmbracelet/bubbles/discussions

### Learning Resources
- **Official Tutorials:** https://github.com/charmbracelet/bubbletea/tree/main/tutorials
- **Community Blog Posts:** Search "Bubbletea tutorial" for community guides
- **Example Projects:** Search GitHub for "topic:bubbletea" to find real-world examples

---

## Conclusion

The Charm/Bubbletea ecosystem is mature and feature-rich. The four highest-priority libraries for TFE are:

1. **bubble-table** - Dramatically improves file list UX
2. **go-fzf** - Adds essential fuzzy finding capability
3. **bubblezone** - Makes mouse interactions more intuitive
4. **Huh** - Provides proper forms for complex operations

Combined with the existing stack, these additions would make TFE a best-in-class terminal file explorer with modern UX patterns that users expect from GUI file managers.

The ecosystem is well-documented, actively maintained, and has a helpful community. All recommended libraries are production-ready and widely used in real applications.

**Next Steps:**
1. Review this report with team
2. Prioritize which features to implement first
3. Start with bubble-table integration (highest ROI)
4. Add go-fzf (killer feature for power users)
5. Polish with bubblezone and Huh

---

**Research completed:** October 16, 2025
**Researcher:** Claude Code (Sonnet 4.5)
**Document version:** 1.0
