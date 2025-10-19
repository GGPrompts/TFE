# Rebuild and Restart TFE

You are rebuilding the TFE application and restarting it.

## Your Task

Use Desktop Commander and standard tools to:

1. **Kill existing TFE process (if running)**
   - Use `list_processes` to find TFE process
   - Use `kill_process` to terminate it gracefully

2. **Clean build artifacts**
   - Run `go clean`
   - Remove old binary if exists

3. **Build TFE**
   - Run `go build` in the project directory
   - Capture build output
   - Check for compilation errors or warnings

4. **Report build results**
   - ✅ If successful: Report build time and binary size
   - ❌ If failed: Show detailed error messages with file:line references
   - ⚠️ If warnings: List all warnings

5. **Start TFE (if build succeeded)**
   - Use `start_process("./tfe")` to launch in background
   - Monitor initial startup (first 2-3 seconds)
   - Check for immediate crashes or panics
   - Report PID and status

6. **Verify it's running**
   - Use `list_processes` to confirm process is alive
   - Check initial output for errors

## Error Handling

If build fails:
- Read the source files with errors
- Analyze the issues
- Suggest fixes
- Ask if I want you to apply them

If runtime crash:
- Capture the panic/error
- Show stack trace
- Identify the problematic code
- Suggest debugging steps

## Report Format

```
🔄 Rebuilding TFE...
  • Killing old process (PID 12345)... ✅
  • Cleaning... ✅
  • Building... ✅ (2.3s, 8.4 MB)
  • Starting... ✅ (PID 12389)
  • Monitoring... ✅ No errors detected

🎯 TFE is running cleanly on PID 12389
```

Execute this rebuild sequence now.
