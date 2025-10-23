---
description: Add debug event logging to TFE for real-time monitoring
---

# Setup TFE Event Logging

You are adding debug event logging to TFE so I can monitor what's happening in the application in real-time.

## Your Task

Add a debug logging system to TFE that logs events to `/tmp/tfe-events.jsonl`.

### Step 1: Add Debug Logger to Model

Edit `types.go` to add a debug log file handle:

```go
type model struct {
    // ... existing fields ...
    debugLog *os.File  // Add this field
}
```

### Step 2: Create Debug Helper

Create or edit `helpers.go` to add logging functions:

```go
func (m *model) enableDebugLogging() error {
    var err error
    m.debugLog, err = os.OpenFile("/tmp/tfe-events.jsonl",
        os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    return err
}

func (m *model) logEvent(event string, data interface{}) {
    if m.debugLog == nil {
        return
    }

    entry := map[string]interface{}{
        "time":  time.Now().Format("15:04:05.000"),
        "event": event,
        "data":  data,
    }

    json.NewEncoder(m.debugLog).Encode(entry)
}

func (m *model) closeDebugLog() {
    if m.debugLog != nil {
        m.debugLog.Close()
    }
}
```

### Step 3: Initialize Logger

Edit `model.go` in the `initialModel()` function:

```go
func initialModel() model {
    m := model{
        // ... existing initialization ...
    }

    // Enable debug logging if TFE_DEBUG env var is set
    if os.Getenv("TFE_DEBUG") == "1" {
        m.enableDebugLogging()
    }

    return m
}
```

### Step 4: Add Logging Calls

Add strategic logging calls throughout the codebase:

**In file_operations.go:**
```go
func (m *model) loadPreview(path string) {
    m.logEvent("preview_load_start", map[string]string{"path": path})

    // ... existing code ...

    m.logEvent("preview_loaded", map[string]interface{}{
        "path": path,
        "size": len(content),
        "lines": strings.Count(content, "\n"),
    })
}

func (m *model) loadFiles(path string) ([]fileItem, error) {
    m.logEvent("load_directory", map[string]string{"path": path})

    // ... existing code ...

    m.logEvent("directory_loaded", map[string]interface{}{
        "path": path,
        "file_count": len(files),
    })

    return files, nil
}
```

**In update_keyboard.go:**
```go
func handleKeyEvent(msg tea.KeyMsg, m model) (model, tea.Cmd) {
    m.logEvent("keypress", map[string]string{
        "key": msg.String(),
        "mode": m.viewMode.String(),
    })

    // ... existing code ...
}
```

**In render_preview.go:**
```go
func (m model) renderPreview() string {
    if m.cursor >= 0 && m.cursor < len(m.files) {
        file := m.files[m.cursor]
        m.logEvent("render_preview", map[string]string{
            "file": file.name,
            "type": file.fileType,
        })
    }

    // ... existing code ...
}
```

### Step 5: Add Cleanup

Edit `main.go` to close the log on exit:

```go
func main() {
    // ... existing code ...

    if m, ok := finalModel.(model); ok {
        m.closeDebugLog()
    }
}
```

### Step 6: Test It

After implementing:
1. Rebuild TFE: `go build`
2. Run with debug enabled: `TFE_DEBUG=1 ./tfe`
3. Test that events are logged: `tail -f /tmp/tfe-events.jsonl`

## Expected Output

When I run TFE with logging enabled, `/tmp/tfe-events.jsonl` should contain:

```json
{"time":"19:45:12.123","event":"load_directory","data":{"path":"/home/matt/projects"}}
{"time":"19:45:12.156","event":"directory_loaded","data":{"path":"/home/matt/projects","file_count":15}}
{"time":"19:45:15.234","event":"keypress","data":{"key":"down","mode":"dual"}}
{"time":"19:45:16.012","event":"preview_load_start","data":{"path":"README.md"}}
{"time":"19:45:16.045","event":"preview_loaded","data":{"path":"README.md","size":1234,"lines":42}}
```

## After Implementation

Once logging is set up, I can monitor TFE in real-time by:
1. Tailing the log file: `tail -f /tmp/tfe-events.jsonl`
2. Seeing exactly what you're doing in the app
3. Automatically reading files you preview
4. Detecting issues before you notice them
5. Suggesting improvements based on your usage patterns

Implement this logging system now. Show me the changes you make and confirm when it's ready to test.
