# TFE Customization Prompts

This folder contains guides for customizing TFE (Terminal File Explorer) without needing a complex configuration system.

## Available Guides

### Core Customizations

- **add-tui-tool.prompty** - Add tools like ncdu, ranger, tig to the context menu
- **customize-toolbar.prompty** - Change emoji toolbar buttons and colors
- **add-file-icons.prompty** - Add icons for specific file types
- **change-colors.prompty** - Customize color schemes (Gruvbox, Dracula, Nord, etc.)
- **add-keyboard-shortcut.prompty** - Add or modify keyboard shortcuts

## How to Use

1. **Open TFE** in your TFE project directory
2. **Press F11** to enter Prompts mode
3. **Navigate to "TFE-Customization"** folder (shown at top of file list)
4. **Select a customization guide** to view step-by-step instructions
5. **Copy relevant code** with F5 when prompted
6. **Edit the appropriate file** and rebuild with `go build -o tfe`

## Philosophy

Instead of adding YAML config files and complexity, TFE uses its own **Prompts Library** feature to document customizations. This approach:

- ✅ Keeps the codebase simple (no config parsing)
- ✅ Teaches users how TFE works internally
- ✅ Encourages contribution and understanding
- ✅ Uses TFE's own features for documentation
- ✅ All customizations are just code edits

## Quick Reference

| What to Customize | File to Edit | Prompt to Use |
|-------------------|--------------|---------------|
| Add TUI tool | `context_menu.go` | add-tui-tool.prompty |
| Change toolbar | `view.go` | customize-toolbar.prompty |
| Add file icons | `file_operations.go` | add-file-icons.prompty |
| Change colors | `styles.go` | change-colors.prompty |
| Add shortcuts | `update_keyboard.go` | add-keyboard-shortcut.prompty |

## Contributing

Found a customization not covered here? Please add a new `.prompty` file and submit a PR!

## License

These guides are part of TFE and released under the MIT License.
