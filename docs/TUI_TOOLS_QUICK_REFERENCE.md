# TUI Tools Quick Reference

**Quick lookup table for TFE integration** | **Date:** 2025-10-16

## By File Type

| File Type | Recommended Tool | Alternative 1 | Alternative 2 | Install Method |
|-----------|------------------|---------------|---------------|----------------|
| **Binary files** | HexPatch | hexabyte | hexyl (view-only) | cargo |
| **CSV/TSV** | VisiData | csv-tui | sc-im | pip / cargo / pkg |
| **JSON** | fx | json-tui | jnv | npm / build / cargo |
| **YAML** | fx | - | - | npm |
| **Markdown** | glow | md-tui | hani | pkg / cargo |
| **Images** | timg | viu | chafa | pkg / cargo |
| **PDF** | termpdf.py | timg | jfbview | pip / pkg |
| **Audio** | musikcube | cmus | MOC | pkg |
| **Video** | mpv | mpvc (controller) | - | pkg |
| **SQL DB** | harlequin | lazysql | gobang | pip / go / cargo |
| **PostgreSQL** | pgcli | harlequin | - | pip |
| **MySQL** | mycli | lazysql | - | pip / go |
| **SQLite** | litecli | harlequin | - | pip |
| **Redis** | iredis | redis-tui | - | pip / cargo |
| **Git repos** | lazygit | gitui | tig | pkg / cargo |
| **Spreadsheets** | sc-im | VisiData | - | pkg / pip |
| **Presentations** | presenterm | lookatme | - | cargo / pip |
| **Notes** | basalt (Obsidian) | nb | jrnl | build / shell / pip |
| **ASCII art** | durdraw | textual-paint | - | pip |
| **HTTP/API** | atac | httpie | curlie | cargo / pip / pkg |

## By Category

### Essential Tools (High Priority)

| Tool | Purpose | Language | GitHub Stars | Last Updated |
|------|---------|----------|--------------|--------------|
| **timg** | Image/video/PDF viewer | C++ | - | 2024 active |
| **viu** | Image viewer | Rust | 2.5k+ | 2024 active |
| **HexPatch** | Hex editor | Rust | - | Active |
| **VisiData** | Data viewer/editor | Python | 7k+ | 2024 active |
| **fx** | JSON/YAML/TOML viewer | Go | 18k+ | 2024 active |
| **glow** | Markdown renderer | Go | 15k+ | Active (Charm) |
| **harlequin** | SQL IDE | Python | 3k+ | 2024 active |
| **lazysql** | DB management | Go | 1k+ | 2024 active |
| **helix** | Code editor (LSP) | Rust | 30k+ | Very active |

### System Tools

| Tool | Purpose | Language | Note |
|------|---------|----------|------|
| **btop** | System monitor | C++ | Best htop alternative |
| **bottom** | System monitor | Rust | Customizable |
| **bandwhich** | Network monitor | Rust | Per-process bandwidth |
| **lazydocker** | Docker TUI | Go | 45k+ stars |

### Development Tools

| Tool | Purpose | Databases | Note |
|------|---------|-----------|------|
| **lazygit** | Git client | - | Already in TFE |
| **gitui** | Git client | - | Fast (Rust) |
| **atac** | API testing | - | Postman-like |
| **pgcli** | PostgreSQL | PostgreSQL | dbcli suite |
| **mycli** | MySQL | MySQL | dbcli suite |
| **litecli** | SQLite | SQLite | dbcli suite |

### Productivity Tools

| Tool | Purpose | File Format | Note |
|------|---------|-------------|------|
| **presenterm** | Presentations | Markdown | PDF export |
| **newsboat** | RSS reader | RSS/Atom | v2.38 (Dec 2024) |
| **aerc** | Email client | Email | Modern, tabbed |
| **taskwarrior-tui** | Task manager | Tasks | TUI for taskwarrior |
| **nb** | Note-taking | Markdown | Git-backed |

## Installation Commands

### Rust (Cargo) - Single Command Install

```bash
cargo install hexyl viu presenterm harlequin-cli gitui atac bandwhich bottom zenith
```

### Python (pip/pipx) - Single Command Install

```bash
pip install visidata termpdf.py jrnl newsboat harlequin pgcli mycli litecli iredis
```

### Go - Individual Installs

```bash
go install github.com/jesseduffield/lazygit@latest
go install github.com/jesseduffield/lazydocker@latest
go install github.com/jorgerojas26/lazysql@latest
```

### Package Managers

**Debian/Ubuntu:**
```bash
sudo apt install micro helix neovim mpv cmus htop btop newsboat mutt sc-im
```

**Homebrew (macOS/Linux):**
```bash
brew install micro helix mpv musikcube htop btop glow fx newsboat aerc
```

**Arch Linux:**
```bash
sudo pacman -S micro helix mpv cmus htop btop glow newsboat neomutt
```

## Terminal Requirements

### For Best Experience (Images/Media)

**Terminals with full graphics support:**
- Kitty (✓ Best support)
- iTerm2 (macOS)
- WezTerm
- Ghostty
- Foot
- Sixel-capable terminals

**Fallback for basic terminals:**
- Most tools have ASCII/block character fallbacks
- Will work in any terminal, just lower quality

## Detection Logic for TFE

### Priority Order by File Type

```go
// Example detection logic
func getToolForFile(filepath string) string {
    ext := strings.ToLower(path.Ext(filepath))

    switch ext {
    case ".csv", ".tsv":
        return firstAvailable([]string{"visidata", "sc-im"})
    case ".json":
        return firstAvailable([]string{"fx", "json-tui"})
    case ".yaml", ".yml":
        return "fx"
    case ".md", ".markdown":
        return firstAvailable([]string{"glow", "md-tui"})
    case ".pdf":
        return firstAvailable([]string{"termpdf.py", "timg"})
    case ".db", ".sqlite", ".sqlite3":
        return firstAvailable([]string{"harlequin", "litecli"})
    // Images
    case ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".webp":
        return firstAvailable([]string{"timg", "viu", "chafa"})
    // Audio
    case ".mp3", ".flac", ".ogg", ".wav", ".m4a":
        return firstAvailable([]string{"musikcube", "cmus", "moc"})
    // Video
    case ".mp4", ".mkv", ".avi", ".mov", ".webm":
        return "mpv"
    default:
        // Binary detection
        if isBinary(filepath) {
            return firstAvailable([]string{"hexpatch", "hexabyte", "hexyl"})
        }
        // Text files - existing editor logic
        return getAvailableEditor()
    }
}
```

## Special Purpose Tools

### ASCII Art Creation
- **durdraw** - Full-featured ANSI/ASCII editor
- **textual-paint** - MS Paint-like TUI

### Database Schema Viewers
- **harlequin** - Best overall SQL IDE
- **lazysql** - Lazygit-style interface
- **gobang** - Rust, cross-platform

### Specialized Viewers
- **termpdf.py** - PDF viewer (Kitty)
- **timg** - Universal viewer (images, videos, PDFs)
- **chafa** - Image to character art

## File Manager References

While TFE is a file manager, these provide inspiration:

| Tool | Language | Special Feature |
|------|----------|-----------------|
| **yazi** | Rust | Async I/O, fastest |
| **ranger** | Python | Vim-inspired, mature |
| **nnn** | C | Minimal, 0-config |
| **lf** | Go | Simple, fast |

## Platform Notes

### Linux-Specific Tools
- `jfbview` (framebuffer)
- `nethogs` (network)
- `iftop` (network)

### Cross-Platform (Best Compatibility)
- Rust tools (via cargo)
- Go tools (single binary)
- Python tools (via pip)

### macOS Considerations
- Use Homebrew for most tools
- iTerm2 has great image support
- Some Linux-specific tools unavailable

### Windows/WSL
- Most tools work in WSL2
- Native Windows support varies
- Use wslview for browser opening (already in TFE)

## Integration Phases

### Phase 1: Essential Viewers (Week 1)
- [ ] Image: timg, viu
- [ ] Hex: HexPatch, hexabyte
- [ ] Data: VisiData
- [ ] JSON/YAML: fx
- [ ] Markdown: glow

### Phase 2: Databases & Development (Week 2)
- [ ] SQL: harlequin
- [ ] Git: gitui (alternative to lazygit)
- [ ] API: atac
- [ ] Database CLIs: pgcli, mycli, litecli

### Phase 3: Media & Productivity (Week 3)
- [ ] Audio: musikcube, cmus
- [ ] Video: mpv (likely already installed)
- [ ] PDF: termpdf.py
- [ ] Presentations: presenterm
- [ ] Spreadsheet: sc-im

### Phase 4: Specialized Tools (Week 4)
- [ ] Docker: lazydocker
- [ ] System monitor: btop
- [ ] Network: bandwhich
- [ ] Notes: basalt, nb
- [ ] Tasks: taskwarrior-tui
- [ ] RSS: newsboat
- [ ] Email: aerc

## Testing Checklist

- [ ] Tool detection works on fresh system
- [ ] Fallback to next available tool
- [ ] Graceful degradation when no tool available
- [ ] Tool availability checking is fast
- [ ] Cross-platform testing (Linux, macOS, WSL)
- [ ] Terminal compatibility (Kitty, Alacritty, etc.)

## Resources

- **Full Research Doc:** `/home/matt/projects/TFE/docs/TUI_APPLICATIONS_RESEARCH.md` (1249 lines)
- **Terminal Trove:** https://terminaltrove.com/ (curated TUI tools)
- **Awesome TUIs:** https://github.com/rothgar/awesome-tuis
- **Charm Tools:** https://charm.sh/ (glow, vhs, etc.)

---

**Total Tools Researched:** 80+
**Actively Maintained (2024-2025):** ~85%
**Primary Languages:** Rust (25), Go (15), Python (15), C/C++ (10)
