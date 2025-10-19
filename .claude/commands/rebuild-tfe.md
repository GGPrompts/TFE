# Pull Changes and Rebuild TFE

You are pulling the latest changes and rebuilding the TFE application binary.

## Your Task

Use standard tools to:

1. **Pull remote changes**
   - Run `git pull` to fetch and merge latest changes
   - Report what was updated (files changed, commits pulled)

2. **Kill existing TFE process (if running)**
   - Use `list_processes` to find any running TFE process
   - Use `kill_process` to terminate it gracefully
   - Skip if no TFE process found

3. **Clean build artifacts**
   - Run `go clean`
   - Remove old binary if exists

4. **Build TFE**
   - Run `go build` in the project directory
   - Capture build output with timing
   - Check for compilation errors or warnings

5. **Install binary to PATH**
   - Ensure `~/.local/bin` directory exists
   - Copy `./tfe` to `~/.local/bin/tfe`
   - Make it executable
   - This allows the `tfe` command/alias to work from any directory

6. **Report build results**
   - ✅ If successful: Report build time, binary size, and install location
   - ❌ If failed: Show detailed error messages with file:line references
   - ⚠️ If warnings: List all warnings

**IMPORTANT:** Do NOT launch TFE as a background process. TFE is a full-screen TUI application that must be run by the user in their own terminal. Just build the binary and report success.

## Error Handling

If build fails:
- Read the source files with errors
- Analyze the issues
- Suggest fixes
- Ask if I want you to apply them

## Report Format

```
🔄 Pulling changes and rebuilding TFE...
  • Pulling from remote... ✅ (3 files changed, 2 commits)
  • Killing old process... ✅ (or: no TFE process found)
  • Cleaning... ✅
  • Building... ✅ (2.3s, 8.4 MB)
  • Installing to ~/.local/bin... ✅

🎯 TFE installed successfully!
   Run from anywhere: tfe
```

Execute this rebuild sequence now.
