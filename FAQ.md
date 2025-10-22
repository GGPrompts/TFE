# TFE Frequently Asked Questions

## Installation & Setup

### Q: TFE won't start, shows "command not found"
**A:** Ensure Go is installed and `~/go/bin` is in your PATH:
```bash
echo 'export PATH=$PATH:~/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### Q: Permission denied errors when accessing directories
**A:** TFE respects file system permissions. Use `chmod` to grant access or run from directories you own. By default, TFE restricts navigation to your home directory and the directory where you started it for security.

### Q: How do I install TFE?
**A:** Install with Go:
```bash
go install github.com/GGPrompts/TFE@latest
```
Or clone and build from source:
```bash
git clone https://github.com/GGPrompts/TFE
cd TFE
go build
```

## Terminal Compatibility

### Q: Emoji buttons have weird spacing (CellBlocks, ttyd, wetty)
**A:** Web-based terminals using xterm.js have emoji rendering issues. This affects ~5% of users.
**Workaround:** Use native terminals (Termux, WSL, iTerm2, GNOME Terminal).
**Future:** v1.1 will add `--ascii-mode` flag.

### Q: Mouse clicks don't work
**A:** Ensure your terminal supports mouse events. Most modern terminals do. If using tmux, add `set -g mouse on` to `.tmux.conf`.

### Q: Colors look wrong or don't display
**A:** TFE requires 256-color terminal support. Check your `TERM` environment variable:
```bash
echo $TERM  # Should be xterm-256color or similar
```

## Feature Questions

### Q: How do I copy files?
**A:** Right-click → "📋 Copy to..." or use context menu (F2).

### Q: Clipboard copy (F5) doesn't work
**A:** Install clipboard utility:
- Linux: `sudo apt install xclip`
- macOS: Built-in (pbcopy)
- Termux: `pkg install termux-api`

### Q: How do I delete files?
**A:** Press Delete or Backspace to move files to trash. Files aren't permanently deleted - view trash with F7 and restore if needed.

### Q: Can I permanently delete files without trash?
**A:** Yes, use Shift+Delete for permanent deletion. Use with caution!

### Q: How do I search for files?
**A:**
- Press `/` for incremental search within current directory
- Press Ctrl+P for fuzzy search across all subdirectories
- Press Ctrl+F for find in files (content search)

## Performance

### Q: TFE is slow with large directories (10,000+ files)
**A:** Use tree view (press 3) for better performance. Detail view renders all metadata upfront. List view (press 1) is also faster for very large directories.

### Q: Preview is slow for large markdown files
**A:** Glamour rendering has a 2-second timeout. Files >1MB show size warning without preview. Press F4 to open in external editor for large files.

### Q: Tree view is slow to expand
**A:** Tree view loads subdirectories on-demand. First expansion of large directories may take a moment. Subsequent expansions are cached.

## Termux / Mobile

### Q: Text is too small on phone
**A:** Adjust Termux font size: Long-press → Style → Font size

### Q: Scrolling doesn't work on phone
**A:** Use arrow keys or vim keys (j/k). Mouse wheel scrolling works if Termux touch mode is enabled.

### Q: How do I use mouse on Termux?
**A:** Touch anywhere to click. Touch and drag to scroll. Double-tap to open files/folders.

### Q: Keyboard is hard to use on mobile
**A:** Termux has a special key row above the keyboard. Map common keys there for easier access. You can also use external Bluetooth keyboards.

## Command Execution

### Q: Why does TFE say "command not allowed"?
**A:** For security, TFE only allows safe read-only commands (ls, cat, grep, git, etc.) by default. Use the `!` prefix for unrestricted access:
```
:ls          # Restricted (safe commands only)
:!rm file    # Unrestricted (full shell access)
```

### Q: What commands are safe to run?
**A:** Safe commands include: ls, cat, grep, find, git, tree, du, df, wc, file, head, tail, diff, sort, and other read-only utilities. See HOTKEYS.md for the full list.

### Q: How do I run commands in the current directory?
**A:** Press `:` (colon) to open the command prompt. Commands run in the currently displayed directory.

## Prompts & Templates

### Q: What are prompt files?
**A:** Prompt files (.yaml, .md, .txt in .prompts directories) are templates with fillable variables like {{FILE}}, {{DIRECTORY}}, {{DATE}}. TFE auto-fills these when you copy the content.

### Q: How do I create a prompt template?
**A:** Create a file in `~/.prompts/` or `.prompts/` with variables in double braces:
```markdown
Analyze this file: {{FILE}}
Located in: {{DIRECTORY}}
Today's date: {{DATE}}
```

### Q: Where should I store prompts?
**A:**
- Global prompts: `~/.prompts/` (available everywhere)
- Project prompts: `.prompts/` in your project root
- Both locations are automatically recognized

## Favorites & Navigation

### Q: How do I bookmark directories?
**A:** Navigate to a directory and press F6 to toggle favorite. Press Ctrl+H to filter and show only favorites.

### Q: How do I get to home directory quickly?
**A:** Press H or click the 🏠 icon in the header.

### Q: What are all the navigation shortcuts?
**A:** See HOTKEYS.md for the complete list. Common ones:
- Arrow keys: Navigate
- Enter: Open directory/file
- Backspace: Go to parent
- H: Home directory
- F6: Toggle favorite
- Ctrl+H: Filter favorites

## Troubleshooting

### Q: TFE crashed, how do I report bugs?
**A:** Open an issue at https://github.com/GGPrompts/TFE/issues with:
1. Your OS and terminal emulator
2. TFE version (`tfe --version`)
3. Steps to reproduce the issue
4. Error message if available

### Q: Preview shows "Binary file" for text files
**A:** Some files with special characters may be detected as binary. Press F4 to open in an external editor instead.

### Q: File modifications don't update automatically
**A:** Press F1 to refresh the current directory view manually.

### Q: Can I customize the colors or theme?
**A:** Currently themes are not customizable. This is planned for a future release.

## Advanced Usage

### Q: Can I open TFE in a specific directory?
**A:** Yes, pass the directory as an argument:
```bash
tfe /path/to/directory
```

### Q: How do I integrate with other TUI tools?
**A:** Press Ctrl+G for lazygit, Ctrl+T for htop, or use Tools menu. You can also run any TUI tool via the command prompt with `!`:
```
:!lazygit
```

### Q: Can I use TFE over SSH?
**A:** Yes! TFE works perfectly over SSH. Ensure your terminal supports mouse events and 256 colors for the best experience.

### Q: Does TFE work on Windows?
**A:** TFE works on Windows via WSL (Windows Subsystem for Linux). Native Windows support is experimental - some features may not work correctly.

## Getting Help

**Q: I have a question not answered here**
**A:**
- Check HOTKEYS.md for keyboard shortcuts
- Read CLAUDE.md for architecture details
- Open a GitHub discussion: https://github.com/GGPrompts/TFE/discussions
- Report bugs: https://github.com/GGPrompts/TFE/issues

**Q: I want to contribute**
**A:** See CONTRIBUTING.md for development setup and guidelines!
