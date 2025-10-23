# Termux-Specific Slash Commands

This directory contains slash commands designed specifically for use on Termux (Android terminal emulator).

## Commands

### `/sync-projects`
**Purpose**: Pull latest changes and rebuild TFE and TUIClassics projects from GitHub

**Use case**: When you make changes on your PC, push to GitHub, and want to sync those changes to your Termux environment. This command automatically:
- Pulls from GitHub for both TFE and TUIClassics
- Rebuilds both projects
- Updates installed binaries
- Reports build status

### `/prompt-engineer`
**Purpose**: Design AI prompts with best practices and copy to clipboard using `termux-clipboard-set`

**Use case**: Interactive prompt engineering assistant that:
- Guides you through creating effective AI prompts
- Follows prompt engineering best practices
- Copies the final prompt to clipboard using Termux's clipboard API
- Supports reusable templates with variables

## Installation

To use these commands on Termux:

1. Copy the desired command files to your `.claude/commands/` directory:
   ```bash
   cp examples/.prompts/termux/*.md ~/.claude/commands/
   ```

2. Make sure you have the required Termux packages:
   ```bash
   pkg install termux-api
   ```

3. Restart Claude Code or reload commands

## Notes

- These commands use Termux-specific tools like `termux-clipboard-set`
- They are NOT intended for use on PC/desktop environments
- The main codebase includes Termux detection to handle platform differences
- PC users should use the standard slash commands in `.claude/commands/`
