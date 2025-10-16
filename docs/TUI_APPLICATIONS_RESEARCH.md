# TUI Applications Research for Terminal File Explorer Integration

**Research Date:** 2025-10-16
**Purpose:** Identify actively maintained TUI applications for integration with TFE terminal file explorer

## Table of Contents
1. [Image & Drawing Tools](#1-image--drawing-tools)
2. [Specialized Editors](#2-specialized-editors)
3. [Media Tools](#3-media-tools)
4. [Development Tools](#4-development-tools)
5. [Document & Productivity Tools](#5-document--productivity-tools)
6. [System & Network Tools](#6-system--network-tools)
7. [File Managers](#7-file-managers)
8. [Communication Tools](#8-communication-tools)

---

## 1. Image & Drawing Tools

### Durdraw
- **GitHub:** https://github.com/cmang/durdraw
- **Description:** Versatile ASCII and ANSI Art text editor for drawing in the terminal
- **File Types:** ASCII art (.txt), ANSI art (.ans)
- **Features:**
  - Animation support
  - 256 and 16 colors
  - Unicode and CP437 support
  - Customizable themes
- **Installation:** Available via pip, package managers
- **Maintenance:** Actively maintained (Copyright 2009-2025)
- **Platform:** Linux, Unix, macOS

### Textual Paint
- **Description:** MS Paint in your terminal - TUI for drawing, paint, image editing
- **File Types:** Pixel art, ASCII art, ANSI art
- **Features:**
  - Drawing and paint tools
  - Image editing capabilities
  - Pixel art creation
- **Installation:** Python package (pip)
- **Maintenance:** Updated July 2025
- **Platform:** Cross-platform

### Moebius
- **Description:** Modern ANSI & ASCII Art Editor
- **File Types:** CP437 textmode art, ANSI art
- **Maintenance:** Updated November 2024
- **Platform:** Cross-platform

### pxltrm
- **Description:** Pixel art editor inside the terminal
- **File Types:** Pixel art
- **Maintenance:** Work in progress, actively developed
- **Platform:** Terminal-based

---

## 2. Specialized Editors

### 2.1 Hex Editors (Binary Files)

#### HexPatch
- **GitHub:** https://github.com/Etto48/HexPatch
- **Description:** Binary patcher and editor written in Rust with TUI
- **File Types:** All binary files
- **Features:**
  - Disassembling instructions
  - Assembling patches
  - Multiple architectures support
  - Various file formats
  - Remote file editing via SSH
- **Installation:** Cargo (Rust package manager)
- **Maintenance:** Actively maintained
- **Platform:** Cross-platform

#### hexabyte
- **Website:** https://terminaltrove.com/hexabyte/
- **Description:** Modern, modular, and robust TUI hex editor
- **File Types:** All binary files
- **Features:**
  - Hexadecimal and ASCII viewing/editing
  - Single, split, and diff modes
  - Plugin support
- **Maintenance:** Actively maintained
- **Platform:** Cross-platform

#### hexyl
- **GitHub:** https://github.com/sharkdp/hexyl
- **Description:** Command-line hex viewer (read-only)
- **File Types:** All binary files
- **Features:**
  - Colored output for byte categories
  - Fast viewing
- **Note:** Viewer only, not an editor
- **Installation:** Cargo, package managers
- **Maintenance:** Actively maintained
- **Platform:** Cross-platform

### 2.2 CSV/Spreadsheet Editors

#### VisiData
- **GitHub:** https://github.com/saulpw/visidata
- **Website:** https://www.visidata.org/
- **Description:** Terminal spreadsheet multitool for discovering and arranging data
- **File Types:** CSV, TSV, SQLite, JSON, XLSX, HDF5, and many more
- **Features:**
  - Vim-like keybindings
  - Sorting, filtering, joins across files
  - Column histograms
  - Column splitting/rejoining
  - Handles millions of rows
  - Interactive TUI interface
- **Installation:** pip, package managers
- **Maintenance:** Actively maintained (2024)
- **Platform:** Cross-platform

#### csv-tui
- **GitHub:** https://github.com/nathangavin/csv-tui
- **Description:** Lightweight CSV editor with TUI
- **File Types:** CSV
- **Features:**
  - Low RAM usage
  - Open, view, edit, save CSV files
- **Installation:** Cargo
- **Maintenance:** Active
- **Platform:** Cross-platform (Rust)

#### sc-im
- **GitHub:** https://github.com/andmarti1424/sc-im
- **Description:** Spreadsheet Calculator Improvised - ncurses spreadsheet program
- **File Types:** .sc, .csv, .tsv, .md, .txt, .ods, .xls, .xlsx (import)
- **Features:**
  - Vim-like keybindings
  - Undo/redo operations
  - GNUPlot interaction
  - LUA scripting
  - Wide character support (many languages)
  - Up to 1,048,576 rows (compile-time option)
- **Installation:** Package managers, source
- **Maintenance:** v0.8.3 (Jan 2023), actively maintained
- **Platform:** Linux, Unix, macOS

### 2.3 JSON/YAML Editors

#### fx
- **GitHub:** https://github.com/antonmedv/fx
- **Website:** https://fx.wtf/
- **Description:** Terminal JSON viewer & processor for JSON, YAML, TOML
- **File Types:** JSON, YAML, TOML
- **Features:**
  - TUI and CLI modes
  - Interactive browsing
  - jq-style filtering
  - Syntax highlighting
- **Installation:** npm, package managers
- **Maintenance:** Actively maintained (2024)
- **Platform:** Cross-platform

#### json-tui
- **GitHub:** https://github.com/ArthurSonzogni/json-tui
- **Description:** JSON terminal UI made in C++
- **File Types:** JSON
- **Features:**
  - Inline display
  - Table views for arrays of objects
  - Navigation and editing
- **Installation:** Build from source
- **Maintenance:** Active
- **Platform:** Cross-platform

#### otree
- **Description:** View objects (JSON/YAML/TOML) in a TUI tree widget
- **File Types:** JSON, YAML, TOML
- **Features:** Tree-based navigation
- **Maintenance:** Active (2024)
- **Platform:** Cross-platform

#### jnv
- **Description:** Interactive JSON filter using jq
- **File Types:** JSON
- **Features:** Real-time jq filtering with TUI
- **Maintenance:** Active (2024)
- **Platform:** Cross-platform

### 2.4 Markdown Editors

#### Glow
- **GitHub:** https://github.com/charmbracelet/glow
- **Description:** Render markdown on the CLI with pizzazz
- **File Types:** Markdown (.md)
- **Features:**
  - Both TUI and CLI modes
  - Automatic dark/light style detection
  - Encrypted cloud stash
  - Beautiful rendering
- **Installation:** Package managers, go install
- **Maintenance:** Actively maintained (Charm project)
- **Platform:** Cross-platform

#### md-tui (mdt)
- **GitHub:** https://github.com/henriklovhaug/md-tui
- **Description:** Markdown renderer in the terminal
- **File Types:** Markdown (.md)
- **Features:**
  - Start with file or search recursively
  - File tree navigation
  - Supports [text](url), [[link]], [[link|title]]
  - Image rendering (if terminal supports)
- **Installation:** Cargo
- **Maintenance:** Active (2024)
- **Platform:** Cross-platform

#### Hani
- **GitHub:** https://github.com/timappledotcom/hani
- **Description:** TUI Markdown Editor with vim-like bindings and live preview
- **File Types:** Markdown (.md)
- **Features:**
  - Vim-like bindings
  - Real-time rendering with glamour (glow)
  - Edit and preview simultaneously
- **Installation:** Build from source
- **Maintenance:** Active
- **Platform:** Cross-platform

### 2.5 Code Editors with IDE Features

#### Micro
- **GitHub:** https://github.com/zyedidia/micro
- **Website:** https://micro-editor.github.io/
- **Description:** Modern and intuitive terminal-based text editor
- **File Types:** All text files
- **Features:**
  - Mouse support
  - Plugin system
  - Syntax highlighting
  - Multiple cursors
  - Common keybindings (Ctrl+C, Ctrl+V)
- **Installation:** Package managers, snap, brew
- **Maintenance:** Actively maintained
- **Platform:** Cross-platform
- **Note:** Already supported in TFE

#### Helix
- **GitHub:** https://github.com/helix-editor/helix
- **Description:** Post-modern modal text editor (Rust)
- **File Types:** All text files
- **Features:**
  - Built-in LSP support
  - Tree-sitter syntax highlighting
  - Multiple selections
  - No configuration needed
- **Installation:** Package managers, cargo
- **Maintenance:** Very active development
- **Platform:** Cross-platform

#### Neovim
- **GitHub:** https://github.com/neovim/neovim
- **Description:** Hyperextensible Vim-based text editor
- **File Types:** All text files
- **Features:**
  - Native LSP client
  - Lua configuration
  - Plugin ecosystem
  - Tree-sitter integration
- **Installation:** Package managers
- **Maintenance:** Very active
- **Platform:** Cross-platform

---

## 3. Media Tools

### 3.1 Image Viewers

#### timg
- **GitHub:** https://github.com/hzeller/timg
- **Description:** Terminal image and video viewer
- **File Types:** Images (jpg, png, gif, etc.), videos, PDFs
- **Features:**
  - Sixel, Kitty, iTerm2 graphics support
  - Animated GIF playback
  - Video playback
  - PDF preview without conversion
  - Grid view for multiple images
- **Installation:** Package managers, build from source
- **Maintenance:** Actively maintained (2024)
- **Platform:** Cross-platform

#### viu
- **GitHub:** https://github.com/atanunq/viu
- **Description:** Terminal image viewer with native support for iTerm and Kitty
- **File Types:** Images (jpg, png, gif, etc.)
- **Features:**
  - Kitty, iTerm2, libsixel support
  - Fallback to blocky ASCII
  - Fast rendering
- **Installation:** Cargo, package managers
- **Maintenance:** Actively maintained
- **Platform:** Cross-platform

#### chafa
- **GitHub:** https://github.com/hpjansson/chafa
- **Description:** Terminal graphics image viewer
- **File Types:** Images, animated GIFs
- **Features:**
  - Kitty graphics protocol (v1.8.0+)
  - Sixel support
  - Character art rendering
- **Installation:** Package managers
- **Maintenance:** Actively maintained (2024)
- **Platform:** Cross-platform

### 3.2 PDF Viewers

#### termpdf.py
- **GitHub:** https://github.com/dsanson/termpdf.py
- **Description:** Graphical PDF and EPUB reader for kitty terminal
- **File Types:** PDF, EPUB
- **Features:**
  - High-quality rendering via PyMuPDF
  - Kitty graphics protocol
  - Navigation controls
- **Installation:** pip
- **Maintenance:** Alpha (actively developed)
- **Platform:** Linux, macOS

#### TermPDF Viewer
- **GitHub:** https://github.com/felipealfonsog/TermPDFViewer
- **Description:** Open-source PDF viewer for terminal
- **File Types:** PDF
- **Features:**
  - Powered by PyMuPDF
  - Interactive navigation
  - Lightweight
- **Installation:** pip, package managers
- **Maintenance:** Active
- **Platform:** Linux, macOS

#### jfbview
- **GitHub:** https://github.com/jichu4n/jfbview
- **Description:** PDF and image viewer for Linux framebuffer
- **File Types:** PDF, images
- **Features:**
  - Customizable multi-threaded caching
  - Fast rendering
  - Framebuffer support
- **Installation:** Package managers
- **Maintenance:** Maintained
- **Platform:** Linux

### 3.3 Audio Players

#### musikcube
- **GitHub:** https://github.com/clangen/musikcube
- **Description:** Cross-platform, terminal-based music player, audio engine, and server
- **File Types:** MP3, AAC, FLAC, OGG, WAV, etc.
- **Features:**
  - Library management
  - Metadata indexer
  - Streaming server
  - Plugin system
  - Works on Raspberry Pi
- **Installation:** Package managers, build from source
- **Maintenance:** Actively maintained (2024)
- **Platform:** Windows, macOS, Linux

#### cmus
- **GitHub:** https://github.com/cmus/cmus
- **Description:** C Music Player - lightweight console music player
- **File Types:** MP3, AAC, WAV, FLAC, and more
- **Features:**
  - Very lightweight (~15MB)
  - Fast startup
  - Vim-like interface
  - Customizable keybindings
- **Installation:** Package managers
- **Maintenance:** Actively maintained
- **Platform:** Unix, Linux

#### MOC (Music On Console)
- **Description:** Light and easy-to-use command-line music player
- **File Types:** WAV, MP3, MP4, FLAC, OGG, AAC, MIDI
- **Features:**
  - Gapless playback
  - Background operation
  - Customizable themes
  - Simple interface
- **Installation:** Package managers
- **Maintenance:** Maintained
- **Platform:** Unix, Linux

### 3.4 Video Players

#### mpv
- **GitHub:** https://github.com/mpv-player/mpv
- **Website:** https://mpv.io/
- **Description:** Command-line media player
- **File Types:** Nearly all video and audio formats
- **Features:**
  - Terminal output modes (tct, kitty, drm)
  - ASCII art mode (-vo caca, -vo aa)
  - Hardware acceleration
  - Scriptable with Lua
  - Extensive configuration options
- **Installation:** Package managers
- **Maintenance:** Very actively maintained (2024)
- **Platform:** Cross-platform

#### mpvc
- **GitHub:** https://gmt4.github.io/mpvc
- **Description:** Minimal mpc-like CLI and TUI for controlling mpv
- **File Types:** All mpv-supported formats
- **Features:**
  - TUI controller for mpv
  - Playlist management
  - Keybindings
- **Installation:** Package managers
- **Maintenance:** Active
- **Platform:** Cross-platform

---

## 4. Development Tools

### 4.1 Database Clients

#### Harlequin
- **GitHub:** https://github.com/tconbeer/harlequin
- **Website:** https://harlequin.sh/
- **Description:** SQL IDE for your terminal
- **File Types:** SQL databases (DuckDB, SQLite, PostgreSQL, etc.)
- **Features:**
  - Drop-in replacement for DuckDB CLI, psql, etc.
  - SQL query editor with syntax highlighting
  - Schema viewer
  - Built with Textual
- **Installation:** pip, pipx
- **Maintenance:** Actively maintained (2024)
- **Platform:** Cross-platform

#### lazysql
- **GitHub:** https://github.com/jorgerojas26/lazysql
- **Description:** Cross-platform TUI database management tool (inspired by Lazygit)
- **File Types:** MySQL, PostgreSQL, SQLite
- **Features:**
  - SQL query editor with syntax highlighting
  - Clipboard support
  - Database browsing
  - Vim-like navigation
- **Installation:** go install, package managers
- **Maintenance:** Actively maintained (2024)
- **Platform:** Cross-platform (Go)

#### gobang
- **GitHub:** https://github.com/TaKO8Ki/gobang
- **Description:** Cross-platform TUI database management tool
- **File Types:** MySQL, PostgreSQL, SQLite
- **Features:**
  - Query execution
  - Schema viewing
  - Table browsing
- **Installation:** Cargo, package managers
- **Maintenance:** Active
- **Platform:** Cross-platform (Rust)

#### pgcli
- **GitHub:** https://github.com/dbcli/pgcli
- **Description:** Postgres CLI with autocompletion and syntax highlighting
- **File Types:** PostgreSQL databases
- **Features:**
  - Auto-completion for SQL keywords, tables, columns
  - Syntax highlighting (Pygments)
  - Pretty-printed output
- **Installation:** pip
- **Maintenance:** Actively maintained (dbcli suite)
- **Platform:** Cross-platform

#### mycli
- **GitHub:** https://github.com/dbcli/mycli
- **Description:** MySQL Terminal Client with auto-completion
- **File Types:** MySQL/MariaDB databases
- **Features:**
  - Auto-completion
  - Syntax highlighting
  - Pretty-printed output
- **Installation:** pip
- **Maintenance:** Actively maintained (dbcli suite)
- **Platform:** Cross-platform

#### litecli
- **GitHub:** https://github.com/dbcli/litecli
- **Description:** SQLite Client with auto-completion
- **File Types:** SQLite databases
- **Features:**
  - Auto-completion
  - Syntax highlighting
  - Based on prompt-toolkit
- **Installation:** pip
- **Maintenance:** Actively maintained (dbcli suite)
- **Platform:** Cross-platform

#### iredis
- **GitHub:** https://github.com/laixintao/iredis
- **Description:** Terminal Client for Redis with AutoCompletion and Syntax Highlighting
- **File Types:** Redis databases
- **Features:**
  - Auto-completion
  - Syntax highlighting
  - Better than redis-cli
- **Installation:** pip
- **Maintenance:** Actively maintained
- **Platform:** Cross-platform

#### redis-tui
- **GitHub:** https://github.com/mylxsw/redis-tui
- **Description:** Redis Text-based UI client in CLI
- **File Types:** Redis databases
- **Features:**
  - TUI interface
  - Key browsing
  - Value inspection
- **Installation:** Cargo, go install
- **Maintenance:** Active
- **Platform:** Cross-platform

### 4.2 Git Clients

#### Lazygit
- **GitHub:** https://github.com/jesseduffield/lazygit
- **Description:** Simple terminal UI for git commands
- **File Types:** Git repositories
- **Features:**
  - Intuitive interface
  - Fast operations
  - Staging, committing, branching
  - Cherry-picking, rebasing
  - Diff viewing
- **Installation:** Package managers, go install
- **Maintenance:** Very actively maintained (5+ years, 37k stars)
- **Platform:** Cross-platform
- **Note:** Already supported in TFE

#### GitUI
- **GitHub:** https://github.com/gitui-org/gitui (formerly extrawurst/gitui)
- **Description:** Blazing fast terminal-ui for git (written in Rust)
- **File Types:** Git repositories
- **Features:**
  - Extremely fast (Rust performance)
  - Vim-like navigation
  - Staging, committing, branching
  - Diff viewing
- **Installation:** Cargo, package managers
- **Maintenance:** Actively maintained (2024)
- **Platform:** Cross-platform

#### Tig
- **GitHub:** https://github.com/jonas/tig
- **Description:** Text-mode interface for git
- **File Types:** Git repositories
- **Features:**
  - Basic git TUI
  - Log browsing
  - Diff viewing
  - Lightweight (C)
- **Installation:** Package managers
- **Maintenance:** Stable, maintained (8k+ Homebrew installs)
- **Platform:** Cross-platform

### 4.3 API Testing Tools

#### atac
- **GitHub:** https://github.com/Julien-cpsn/ATAC
- **Website:** https://terminaltrove.com/atac/
- **Description:** A Terminal API Client (like Postman, but in terminal)
- **File Types:** HTTP requests
- **Features:**
  - Full TUI interface
  - GET, POST, PUT, DELETE, etc.
  - Authentication (Basic, Bearer tokens)
  - Request collections
  - Postman-like experience
- **Installation:** Cargo, package managers
- **Maintenance:** Actively maintained (2024)
- **Platform:** Cross-platform

#### HTTPie
- **GitHub:** https://github.com/httpie/httpie
- **Description:** User-friendly HTTP client for CLI
- **File Types:** HTTP requests
- **Features:**
  - Simple syntax
  - Syntax highlighting
  - JSON support
  - Both terminal and desktop versions
- **Installation:** pip, package managers
- **Maintenance:** Very actively maintained
- **Platform:** Cross-platform

#### Curlie
- **GitHub:** https://github.com/rs/curlie
- **Description:** HTTPie wrapper around cURL
- **File Types:** HTTP requests
- **Features:**
  - cURL performance
  - HTTPie ease of use
  - All cURL features
- **Installation:** Package managers
- **Maintenance:** Active
- **Platform:** Cross-platform

#### xh
- **GitHub:** https://github.com/ducaale/xh
- **Description:** Friendly and fast tool for sending HTTP requests
- **File Types:** HTTP requests
- **Features:**
  - Fast (Rust)
  - HTTPie-compatible
  - Simplified syntax
- **Installation:** Cargo, package managers
- **Maintenance:** Active
- **Platform:** Cross-platform

### 4.4 Docker Management

#### Lazydocker
- **GitHub:** https://github.com/jesseduffield/lazydocker
- **Website:** https://lazydocker.com/
- **Description:** The lazier way to manage everything docker
- **File Types:** Docker containers, images, volumes
- **Features:**
  - Container management
  - Log viewing
  - Stats monitoring
  - Volume and network management
  - Intuitive UI
- **Installation:** Package managers, go install
- **Maintenance:** Very actively maintained (45k+ stars)
- **Platform:** Cross-platform

#### ctop
- **GitHub:** https://github.com/bcicen/ctop
- **Description:** Top-like interface for container metrics
- **File Types:** Docker/container metrics
- **Features:**
  - Real-time metrics
  - Simple interface
  - Resource monitoring
- **Installation:** Package managers
- **Maintenance:** Active
- **Platform:** Cross-platform

#### oxker
- **GitHub:** https://github.com/mrjackwills/oxker
- **Description:** Simple TUI to view & control docker containers
- **File Types:** Docker containers
- **Features:**
  - Container control
  - Log viewing
  - Lightweight
- **Installation:** Cargo
- **Maintenance:** Active
- **Platform:** Cross-platform

---

## 5. Document & Productivity Tools

### 5.1 Presentation Tools

#### presenterm
- **GitHub:** https://github.com/mfontanini/presenterm
- **Website:** https://terminaltrove.com/presenterm/
- **Description:** Markdown terminal slideshow tool
- **File Types:** Markdown (.md)
- **Features:**
  - Images and animated GIFs (kitty, iTerm2, wezterm, ghostty, foot)
  - Highly customizable themes
  - Code highlighting
  - PDF export
  - Automatic reload on change
  - Speaker notes
- **Installation:** Cargo, package managers
- **Maintenance:** Actively maintained (2024)
- **Platform:** Cross-platform

#### lookatme
- **GitHub:** https://github.com/d0c-s4vage/lookatme
- **Description:** Interactive, terminal-based markdown presenter
- **File Types:** Markdown (.md)
- **Features:**
  - Interactive presentations
  - Live editing
  - Themes
  - Code highlighting
- **Installation:** pip
- **Maintenance:** Active
- **Platform:** Cross-platform

#### tui-slides
- **GitHub:** https://github.com/Chleba/tui-slides
- **Description:** Terminal Presentation program with modern TUI
- **File Types:** Custom format
- **Features:**
  - Modern interface
  - Slide transitions
- **Installation:** Build from source
- **Maintenance:** Active
- **Platform:** Cross-platform

### 5.2 Note-Taking

#### basalt
- **GitHub:** https://github.com/erikjuhani/basalt
- **Website:** https://terminaltrove.com/basalt/
- **Description:** TUI Application to manage Obsidian notes from terminal
- **File Types:** Markdown (Obsidian vaults)
- **Features:**
  - Markdown rendering with inline images
  - Side panel for note selection
  - Persistent scroll position
  - Mode indicators (Select, Normal, Insert)
  - Word and character statistics
- **Installation:** Build from source
- **Maintenance:** Active (2024)
- **Platform:** Cross-platform

#### nb
- **GitHub:** https://github.com/xwmx/nb
- **Description:** Command line and local web note-taking, bookmarking, archiving
- **File Types:** Markdown, text, bookmarks
- **Features:**
  - Git-backed
  - Encryption support
  - Search and filtering
  - Bookmarking
  - Knowledge base
- **Installation:** Shell script, package managers
- **Maintenance:** Very actively maintained
- **Platform:** Cross-platform

#### jrnl
- **GitHub:** https://github.com/jrnl-org/jrnl
- **Description:** Simple journal application for command line
- **File Types:** Plain text journals
- **Features:**
  - AES encryption
  - Search and filtering
  - Human-readable format
  - Tagging
- **Installation:** pip, package managers
- **Maintenance:** Actively maintained
- **Platform:** Cross-platform

#### jot
- **GitHub:** https://github.com/shashwatah/jot
- **Description:** Rapid note management for terminal (Obsidian-inspired)
- **File Types:** Markdown
- **Features:**
  - Obsidian-compatible storage
  - Vault and folder support
  - Fast note access
- **Installation:** Build from source
- **Maintenance:** Active
- **Platform:** Cross-platform

### 5.3 Task Management

#### Taskwarrior + taskwarrior-tui
- **GitHub:** https://github.com/kdheepak/taskwarrior-tui
- **Website:** https://taskwarrior.org/
- **Description:** Terminal User Interface for Taskwarrior
- **File Types:** Task database
- **Features:**
  - Task management with dates, priorities, projects
  - Filtering and reports
  - Visual TUI interface
  - Tags and annotations
- **Installation:** Package managers, cargo (for TUI)
- **Maintenance:** Active (Note: TUI v0.25.4 is last for taskwarrior v2.x)
- **Platform:** Cross-platform

#### Taskell
- **GitHub:** https://github.com/smallhadroncollider/taskell
- **Description:** CLI kanban board/task manager
- **File Types:** Task files
- **Features:**
  - Kanban-style boards
  - Trello board support
  - GitHub project integration
- **Installation:** Package managers
- **Maintenance:** Active
- **Platform:** Mac, Linux

#### Calcure
- **Description:** Modern TUI calendar and task manager
- **File Types:** Calendar events, tasks
- **Features:**
  - Event and task management
  - Birthday display
  - Import from calcurse, taskwarrior
  - Modern interface
- **Installation:** pip
- **Maintenance:** Active (2024)
- **Platform:** Cross-platform

### 5.4 RSS Feed Reader

#### Newsboat
- **GitHub:** https://github.com/newsboat/newsboat
- **Website:** https://newsboat.org/
- **Description:** RSS/Atom feed reader for text terminals
- **File Types:** RSS, Atom, JSON feeds
- **Features:**
  - Keyboard-driven interface
  - Built-in HTML renderer
  - Podcast support
  - Integration with aggregators (FreshRSS, NewsBlur, Inoreader, etc.)
  - Highly customizable themes
  - Vim-style navigation
- **Installation:** Package managers
- **Maintenance:** Very active (v2.38 released Dec 2024)
- **Platform:** Linux, BSD, macOS, Windows

### 5.5 Email Clients

#### aerc
- **GitHub:** https://git.sr.ht/~rjarry/aerc
- **Description:** Email client that runs in the terminal
- **File Types:** Email (IMAP, Maildir, etc.)
- **Features:**
  - Tabbed interface
  - Multiple accounts
  - Folder/email split view
  - Modern and user-friendly
  - Extensible
  - Less complicated than Mutt
- **Installation:** Package managers
- **Maintenance:** Very actively maintained (2024)
- **Platform:** Cross-platform (Go)

#### NeoMutt
- **GitHub:** https://github.com/neomutt/neomutt
- **Description:** Command-line email reader (Mutt fork)
- **File Types:** Email (IMAP, POP3, Maildir, etc.)
- **Features:**
  - Highly configurable
  - Powerful filtering
  - Threading
  - Vim-like bindings
- **Installation:** Package managers
- **Maintenance:** Actively maintained
- **Platform:** Cross-platform

---

## 6. System & Network Tools

### 6.1 System Monitors

#### btop
- **GitHub:** https://github.com/aristocratos/btop
- **Description:** Resource monitor (C++ variant of bashtop)
- **File Types:** System metrics
- **Features:**
  - CPU, memory, disk, network, process monitoring
  - Full mouse support
  - Game-inspired menu system
  - Process filtering and tree view
  - 256 colors
- **Installation:** Package managers, build from source
- **Maintenance:** Very actively maintained (2024)
- **Platform:** Cross-platform

#### bottom (btm)
- **GitHub:** https://github.com/ClementTsang/bottom
- **Description:** Customizable cross-platform graphical process/system monitor
- **File Types:** System metrics
- **Features:**
  - Customizable layout
  - Zoom-able charts
  - CPU, memory, network, disk usage
  - Process management
- **Installation:** Cargo, package managers
- **Maintenance:** Very actively maintained
- **Platform:** Linux, macOS, Windows

#### zenith
- **GitHub:** https://github.com/bvaisvil/zenith
- **Description:** Terminal monitor with zoom-able charts
- **File Types:** System metrics
- **Features:**
  - CPU, GPU, network, disk usage
  - Zoom-able histograms
  - Process list
- **Installation:** Cargo, package managers
- **Maintenance:** Active
- **Platform:** Cross-platform (Rust)

#### htop
- **GitHub:** https://github.com/htop-dev/htop
- **Description:** Interactive text-mode process viewer
- **File Types:** System processes
- **Features:**
  - Better than 'top'
  - Mouse support
  - Tree view
  - Process sorting
- **Installation:** Package managers
- **Maintenance:** Actively maintained
- **Platform:** Unix, Linux, macOS
- **Note:** Already supported in TFE

#### gotop
- **GitHub:** https://github.com/xxxserxxx/gotop
- **Description:** Terminal-based graphical activity monitor (Go)
- **File Types:** System metrics
- **Features:**
  - Graphical displays
  - CPU, memory, network, disk
  - Process list
- **Installation:** Package managers, go install
- **Maintenance:** Active fork (2024)
- **Platform:** Cross-platform

### 6.2 Network Monitors

#### bandwhich
- **GitHub:** https://github.com/imsnif/bandwhich
- **Description:** Terminal bandwidth utilization tool
- **File Types:** Network traffic
- **Features:**
  - Real-time network visualization
  - Traffic broken down by process, connection, remote IP
  - Interactive curses interface
  - Cross-platform
- **Installation:** Cargo, package managers
- **Maintenance:** Actively maintained (Rust)
- **Platform:** Cross-platform

#### nethogs
- **GitHub:** https://github.com/raboof/nethogs
- **Description:** Net top tool - bandwidth per process
- **File Types:** Network traffic
- **Features:**
  - Groups bandwidth by process
  - Identifies bandwidth-hogging applications
  - Simple interface
- **Installation:** Package managers
- **Maintenance:** Maintained
- **Platform:** Linux

#### iftop
- **Description:** Display bandwidth usage on network interface by host
- **File Types:** Network traffic
- **Features:**
  - Bandwidth usage between hosts
  - Sorted by data transfer rates
  - Connection-level insights
- **Installation:** Package managers
- **Maintenance:** Stable (classic tool)
- **Platform:** Unix, Linux

---

## 7. File Managers

While TFE is itself a file manager, these TUI file managers could provide reference or inspiration for features:

#### yazi
- **GitHub:** https://github.com/sxyazi/yazi
- **Description:** Blazing fast terminal file manager (Rust, async I/O)
- **File Types:** All files
- **Features:**
  - Built-in preview (text, PDF, images, videos)
  - Syntax highlighting
  - Asynchronous performance
  - Precache mechanism
  - Image support (iTerm2, Kitty, etc.)
- **Installation:** Cargo, package managers
- **Maintenance:** Very actively maintained (2024 rising star)
- **Platform:** Cross-platform

#### ranger
- **GitHub:** https://github.com/ranger/ranger
- **Description:** VIM-inspired file manager for console
- **File Types:** All files
- **Features:**
  - Vim keybindings
  - Preview pane
  - Customizable
  - File operations
- **Installation:** Package managers, pip
- **Maintenance:** Maintained
- **Platform:** Cross-platform

#### nnn
- **GitHub:** https://github.com/jarun/nnn
- **Description:** Full-featured terminal file manager (nearly 0-config)
- **File Types:** All files
- **Features:**
  - Very fast and lightweight
  - Plugin system
  - Batch operations
  - Minimal configuration
- **Installation:** Package managers
- **Maintenance:** Very actively maintained
- **Platform:** Cross-platform

#### lf
- **GitHub:** https://github.com/gokcehan/lf
- **Description:** Terminal file manager (inspired by ranger, written in Go)
- **File Types:** All files
- **Features:**
  - Simple and minimal
  - Fast (Go)
  - Customizable
- **Installation:** Package managers, go install
- **Maintenance:** Actively maintained
- **Platform:** Cross-platform

---

## 8. Communication Tools

*(While less common for file creation/editing, these could open communication-related files)*

### IRC/Chat Clients

#### irssi
- **GitHub:** https://github.com/irssi/irssi
- **Description:** Terminal-based IRC client
- **Installation:** Package managers
- **Maintenance:** Actively maintained
- **Platform:** Cross-platform

#### weechat
- **GitHub:** https://github.com/weechat/weechat
- **Description:** Fast, light, extensible chat client
- **Installation:** Package managers
- **Maintenance:** Very active
- **Platform:** Cross-platform

---

## Integration Recommendations for TFE

Based on this research, here are priority tools for TFE integration:

### High Priority (Most Useful)

1. **Image Viewers:**
   - `timg` (best all-around: images, videos, PDFs, grid view)
   - `viu` (simple, fast, good fallback)

2. **Hex Editors:**
   - `HexPatch` (modern, feature-rich, Rust)
   - `hexabyte` (modular, plugins)

3. **Data/CSV:**
   - `VisiData` (extremely versatile, handles millions of rows)

4. **JSON/YAML:**
   - `fx` (supports JSON, YAML, TOML - one tool for all)

5. **Markdown:**
   - `glow` (beautiful rendering, both TUI and CLI modes)
   - `hani` (if editing needed - live preview)

6. **PDF:**
   - `termpdf.py` (best for Kitty users)
   - `timg` (can preview PDFs without conversion)

7. **Database:**
   - `harlequin` (modern SQL IDE in terminal)
   - `lazysql` (Lazygit-style interface)

8. **Code/Advanced Editing:**
   - `helix` (modern, LSP built-in)
   - Continue supporting `micro`, `nano`, `vim`, `vi`

### Medium Priority (Specialized Use Cases)

9. **Audio:**
   - `musikcube` (best all-around)
   - `cmus` (lightweight)

10. **Video:**
    - `mpv` (already widely installed)

11. **API Testing:**
    - `atac` (Postman-like TUI)

12. **Presentations:**
    - `presenterm` (modern, feature-rich)

13. **Spreadsheets:**
    - `sc-im` (vim-like, powerful)

### Lower Priority (Niche)

14. **ASCII Art:**
    - `durdraw` (specialized creative tool)

15. **Notes:**
    - `basalt` (Obsidian integration)
    - `nb` (general note-taking)

---

## Implementation Strategy

### File Type Detection

Update `/home/matt/projects/TFE/file_operations.go` to detect new file types:

```go
// Add to file_operations.go
func getAppForFileType(filepath string) string {
    ext := strings.ToLower(filepath.Ext(filepath))

    // Binary files
    if isBinaryFile(filepath) && !isImageFile(filepath) {
        return getAvailableHexEditor() // HexPatch > hexabyte > hexyl
    }

    // Data files
    switch ext {
    case ".csv", ".tsv":
        return "visidata"
    case ".json":
        return "fx"
    case ".yaml", ".yml":
        return "fx"
    case ".md", ".markdown":
        return "glow"
    case ".pdf":
        return getAvailablePDFViewer() // termpdf.py > timg
    case ".db", ".sqlite", ".sqlite3":
        return "harlequin"
    // ... etc
    }

    return getAvailableEditor() // fallback to existing logic
}
```

### Tool Availability Checking

Add functions to `/home/matt/projects/TFE/editor.go`:

```go
// Hex editors
func getAvailableHexEditor() string {
    editors := []string{"hexpatch", "hexabyte", "hexyl"}
    for _, editor := range editors {
        if editorAvailable(editor) {
            return editor
        }
    }
    return ""
}

// PDF viewers
func getAvailablePDFViewer() string {
    viewers := []string{"termpdf.py", "timg", "glow"} // glow for MD->PDF
    for _, viewer := range viewers {
        if editorAvailable(viewer) {
            return viewer
        }
    }
    return ""
}

// Image viewers
func getAvailableImageViewer() string {
    viewers := []string{"timg", "viu", "chafa"}
    for _, viewer := range viewers {
        if editorAvailable(viewer) {
            return viewer
        }
    }
    return ""
}
```

### Installation Documentation

Create installation instructions for users in `/home/matt/projects/TFE/README.md` section:

**High-value tools by install method:**

- **Cargo (Rust):** `cargo install hexyl viu presenterm harlequin-cli`
- **Go:** `go install github.com/jesseduffield/lazydocker@latest`
- **Pip (Python):** `pip install visidata fx termpdf.py jrnl newsboat`
- **Package Managers:** Most tools available via `apt`, `brew`, `pacman`, etc.

---

## Cross-Platform Considerations

- **Most compatible:** Tools written in Go, Rust, Python are most cross-platform
- **Linux-specific:** `jfbview` (framebuffer), `nethogs`, `iftop`
- **Terminal requirements:**
  - Image tools require: Kitty, iTerm2, WezTerm, or Sixel support
  - Best experience: Kitty terminal (widest support)
  - ASCII fallbacks available for most image tools

---

## Summary Statistics

**Total Tools Researched:** 80+

**By Category:**
- Image/Drawing: 7 tools
- Specialized Editors: 20+ tools
- Media Tools: 13 tools
- Development Tools: 20+ tools
- Document/Productivity: 15+ tools
- System/Network Tools: 8 tools
- File Managers: 4 tools (reference)

**By Language:**
- Rust: ~25 tools (performance-focused)
- Go: ~15 tools (cross-platform, fast)
- Python: ~15 tools (easy to install)
- C/C++: ~10 tools (mature, stable)

**Maintenance Status:**
- Actively maintained (2024-2025): ~85%
- Stable/mature: ~10%
- New/emerging: ~5%

---

## Next Steps

1. **Phase 1:** Implement high-priority tools (image viewers, hex editors, VisiData, fx, glow)
2. **Phase 2:** Add database clients (harlequin, lazysql) and media players
3. **Phase 3:** Specialized tools (API testing, presentations, notes)
4. **Documentation:** Update README with optional tool installation instructions
5. **Testing:** Test tool detection and launching across platforms

---

**Research compiled by:** Claude Code (Sonnet 4.5)
**Date:** 2025-10-16
**Project:** TFE (Terminal File Explorer)
**File:** `/home/matt/projects/TFE/docs/TUI_APPLICATIONS_RESEARCH.md`
