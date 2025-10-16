# Charm/Bubbletea Ecosystem - Quick Reference

**Last Updated:** October 16, 2025

This is a condensed reference guide. For detailed information, see `ECOSYSTEM_RESEARCH_2025.md`.

---

## Top 4 Recommended Libraries for TFE

### 1. bubble-table ⭐⭐⭐
**GitHub:** https://github.com/Evertras/bubble-table
**Package:** `github.com/evertras/bubble-table/table`

Interactive table component with sorting, filtering, pagination, frozen columns.

**Use Case:** Replace detail view with powerful interactive table
**Priority:** HIGH

---

### 2. go-fzf ⭐⭐⭐
**GitHub:** https://github.com/koki-develop/go-fzf
**Package:** `github.com/koki-develop/go-fzf`

Fuzzy finder built with Bubbletea (like fzf command-line tool).

**Use Case:** Add Ctrl+P fuzzy file search, command palette
**Priority:** HIGH

---

### 3. bubblezone ⭐⭐⭐
**GitHub:** https://github.com/lrstanley/bubblezone
**Package:** `github.com/lrstanley/bubblezone`

Mouse event tracking with zero-width markers.

**Use Case:** Make buttons, icons, breadcrumbs clickable; hover states
**Priority:** HIGH

---

### 4. Huh ⭐⭐⭐
**GitHub:** https://github.com/charmbracelet/huh
**Package:** `github.com/charmbracelet/huh`

Forms and prompts library (official Charm library).

**Use Case:** Replace dialog system with proper multi-field forms
**Priority:** HIGH

---

## Official Charm Libraries

| Library | Purpose | Priority |
|---------|---------|----------|
| **Huh** | Forms & prompts | HIGH |
| **Harmonica** | Physics animations | MEDIUM |
| **Log** | Styled logging | LOW |
| **VHS** | Terminal GIF recorder | MEDIUM |
| **Wish** | SSH apps framework | MEDIUM |
| **x/exp/teatest** | TUI testing | MEDIUM |

---

## Community Libraries

| Library | Purpose | Priority |
|---------|---------|----------|
| **bubble-table** | Interactive tables | HIGH |
| **go-fzf** | Fuzzy finder | HIGH |
| **bubblezone** | Mouse zones | HIGH |
| **promptkit** | Simple prompts | MEDIUM |
| **treeview** | Tree navigation | MEDIUM |
| **rasterm** | Image display | MEDIUM |

---

## Reference Projects

| Project | Language | Purpose |
|---------|----------|---------|
| **lf** | Go | File manager reference |
| **Soft Serve** | Go | Advanced Bubbletea patterns |
| **tview** | Go | Component ideas (different framework) |

---

## Quick Installation Commands

```bash
# Top 4 recommended
go get github.com/evertras/bubble-table
go get github.com/koki-develop/go-fzf
go get github.com/lrstanley/bubblezone
go get github.com/charmbracelet/huh

# Additional useful libraries
go get github.com/charmbracelet/harmonica
go get github.com/charmbracelet/log
go get github.com/charmbracelet/x/exp/teatest
go get github.com/BourgeoisBear/rasterm
```

---

## Implementation Order

1. **Week 1-2:** bubble-table (detail view enhancement)
2. **Week 2-3:** go-fzf (fuzzy finding)
3. **Week 3-4:** bubblezone (better mouse UX)
4. **Week 4-5:** Huh (forms and dialogs)
5. **Week 5-6:** Harmonica + teatest (polish and testing)

---

## Official Examples to Study

All at: https://github.com/charmbracelet/bubbletea/tree/main/examples

- `file-picker/` - File system navigation
- `progress-download/` - Progress bars for operations
- `views/` - Multi-pane view management

---

## Community Resources

- **Discord:** https://charm.sh/discord
- **Discussions:** https://github.com/charmbracelet/bubbletea/discussions
- **Examples:** Search GitHub for "topic:bubbletea"

---

For full documentation, code examples, and detailed integration guides, see:
- `docs/ECOSYSTEM_RESEARCH_2025.md` (comprehensive report)
- Official Bubbletea docs: https://github.com/charmbracelet/bubbletea
