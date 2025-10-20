# Next Session: Add Paste Support to Prompt Fillable Fields

## Problem Statement

**Current Issue:** When using TFE's prompts library (F11), fillable fields like `{{CODE}}`, `{{DESCRIPTION}}`, etc. don't support pasting content. Users must type everything manually, which is impractical for:
- Large code snippets
- Multi-line descriptions
- Pre-written content from other sources

**Impact:** The prompts feature is less useful than it could be, especially for code review and documentation prompts that expect substantial pasted content.

## Current Behavior

**What works:**
- Typing individual characters into fillable fields ✅
- Backspace to delete characters ✅
- Ctrl+U to clear entire field ✅
- Tab/Shift+Tab to navigate between fields ✅

**What doesn't work:**
- Pasting content (Ctrl+V, Shift+Insert, right-click paste) ❌
- Terminal bracketed paste mode not handled ❌

## Desired Behavior

Users should be able to **paste multi-line content** into fillable fields:

**Example workflow:**
1. Copy code from another file (100+ lines)
2. Press F11, select "Code Review" prompt
3. Navigate to `{{CODE}}` field
4. Paste content (Ctrl+V or Shift+Insert)
5. Field populates with entire pasted content
6. Continue filling other fields
7. Press F5 to copy rendered prompt

## Technical Context

### Relevant Files
- `update_keyboard.go` - Lines 42-112 handle input field editing
- `types.go` - `promptInputField` struct definition
- `render_preview.go` - Renders fillable fields UI

### Current Input Handling

```go
// update_keyboard.go:96-112
default:
    // Handle regular character input
    keyStr := msg.String()

    // Filter out function keys and special keys
    // ...

    // Add to focused field
    if m.focusedInputField >= 0 && m.focusedInputField < len(m.promptInputFields) {
        m.promptInputFields[m.focusedInputField].value += keyStr
    }
    return m, nil
```

**This only handles single characters, not paste events!**

## Implementation Approach

### Step 1: Detect Paste Events

Terminal paste comes in two forms:

**1. Bracketed Paste Mode** (modern terminals):
```
ESC [ 2 0 0 ~ <pasted content> ESC [ 2 0 1 ~
```

**2. Bubbletea Paste Detection**:
- Check `msg.Paste` type or use `msg.Runes` for multi-character input
- Similar to how command prompt handles paste (see `update_keyboard.go:1196-1227`)

### Step 2: Handle Multi-line Content

Allow `\n` characters in field values:
```go
// Currently field.value is a string, so newlines should work
// But need to handle rendering multi-line content in UI
```

### Step 3: Update Rendering

**In `render_preview.go`**, display multi-line fields properly:
- Show first line + "... (X more lines)" for long fields
- Or show first 3-5 lines with scrollbar indicator
- Update character count to show line count too

### Step 4: Example Implementation

```go
// In update_keyboard.go, around line 96-112
default:
    if m.focusedInputField >= 0 && m.focusedInputField < len(m.promptInputFields) {
        field := &m.promptInputFields[m.focusedInputField]

        // Use msg.Runes to get raw input (handles paste properly)
        text := string(msg.Runes)

        // Check if this is a paste event (multiple characters at once)
        if len(msg.Runes) > 1 {
            // Paste detected - add entire content
            field.value += text

            // Show paste feedback
            lineCount := strings.Count(text, "\n") + 1
            charCount := len(text)
            m.setStatusMessage(fmt.Sprintf("✓ Pasted %d chars (%d lines)", charCount, lineCount), false)
        } else {
            // Single character typed
            field.value += text
        }
    }
    return m, nil
```

## Testing Plan

**Test cases:**
1. ✅ Paste single line of text
2. ✅ Paste multi-line content (10+ lines)
3. ✅ Paste very large content (100+ lines, 5k+ chars)
4. ✅ Paste content with special characters (quotes, backslashes)
5. ✅ Paste content with Unicode/emoji
6. ✅ Tab to next field after pasting
7. ✅ Ctrl+U clears pasted content
8. ✅ Rendered prompt shows pasted content correctly

**Terminals to test:**
- Windows Terminal (primary)
- Termux (mobile)
- Linux terminals (if available)

## Success Criteria

**After implementation:**
- [ ] Users can paste content into any fillable field
- [ ] Multi-line paste works correctly
- [ ] Large pastes don't freeze UI
- [ ] Character/line count updates correctly
- [ ] Pasted content renders in final prompt
- [ ] Paste feedback shows in status bar
- [ ] Works across all supported terminals

## Priority

**Medium-High** - This significantly improves usability of the prompts feature, which is one of TFE's unique selling points!

## Estimated Time

**1-2 hours** - The paste detection logic is already implemented for command prompt, just needs to be adapted for fillable fields.

## Related Code

**Command prompt paste handling** (reference implementation):
```go
// update_keyboard.go:1196-1227
case "enter":
    // ... command execution

default:
    // Use msg.Runes to avoid brackets from msg.String() on paste events
    text := string(msg.Runes)
    if len(text) > 0 {
        // Check if all characters are printable
        isPrintable := true
        for _, r := range msg.Runes {
            if r < 32 || r > 126 {
                isPrintable = false
                break
            }
        }

        if isPrintable {
            m.commandInput += text
            m.historyPos = len(m.commandHistory)
        }
    }
```

**This same pattern should work for fillable fields!**

## Notes

- Consider max field size limit (10k chars?) to prevent memory issues
- May want to show "Paste too large" warning for >100k chars
- Multi-line rendering in UI might need scrolling or truncation
- Check if ANSI codes need to be stripped from pasted content

---

**Good luck!** This will make the prompts feature much more practical for real-world use! 🚀
