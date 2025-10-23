---
description: Pull latest changes and rebuild TFE and TUIClassics projects
---

# Sync TFE and TUIClassics from GitHub

You are synchronizing both TFE and TUIClassics projects from their GitHub remotes, rebuilding them, and updating the installed binaries.

## Your Task

Use standard tools to:

1. **Sync TFE Project**
   - Navigate to `~/TFE`
   - Run `git pull` to fetch and merge latest changes
   - Report what was updated (files changed, commits pulled)
   - Run `go build -o tfe .` to rebuild
   - Copy binary to `~/bin/tfe` (the alias points here)
   - Report build time and binary size

2. **Sync TUIClassics Project**
   - Navigate to `~/TUIClassics`
   - Run `git pull` to fetch and merge latest changes
   - Report what was updated
   - Run `make clean && make all` to rebuild all games
   - Build snake separately: `go build -o bin/snake ./cmd/snake`
   - Report which games were built successfully

3. **Update Installed Binaries**
   - TFE binary is automatically copied to `~/bin/tfe` (step 1)
   - Games remain in `~/TUIClassics/bin/` (accessible via the classics launcher)
   - Optional: Ask user if they want to install games to `~/bin/` for direct access

4. **Verify Installation**
   - Check that `~/bin/tfe` exists and is executable
   - List all built games in `~/TUIClassics/bin/`
   - Report final status

## Context

This command is used when the user:
- Makes changes on their PC and pushes to GitHub
- Wants to sync those changes to their phone/Termux environment
- Needs both TFE and TUIClassics updated at once

## Error Handling

If git pull fails:
- Check if there are local uncommitted changes
- Suggest stashing or committing them first
- Show git status for diagnosis

If build fails:
- Show detailed compilation errors
- Check if dependencies are missing
- Suggest fixes or ask if user wants you to apply them

## Report Format

```
🔄 Syncing projects from GitHub...

📂 TFE
  • Pulling from remote... ✅ (5 files changed, 3 commits)
  • Building... ✅ (1.8s, 16 MB)
  • Installing to ~/bin/tfe... ✅

🎮 TUIClassics
  • Pulling from remote... ✅ (12 files changed, 2 commits)
  • Cleaning build artifacts... ✅
  • Building minesweeper... ✅
  • Building solitaire... ✅
  • Building 2048... ✅
  • Building snake... ✅
  • Building classics launcher... ✅

🚀 All projects synced successfully!

Available commands:
  tfe          - Terminal File Explorer
  classics     - Game launcher (balatro, hero, snake, 2048, minesweeper, solitaire)
```

Execute this sync sequence now.
