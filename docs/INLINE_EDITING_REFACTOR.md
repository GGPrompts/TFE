# Complete Inline Variable Editing Refactor

## Context

We're replacing the 3-section prompt layout (header / content / fillable fields panel) with **inline variable editing** where users press Tab to cycle through `{{variables}}` directly in the content and edit them in place.

## What's Been Completed ✅

### 1. Types and State (types.go, model.go)
- ✅ Removed old system: `promptInputFields []promptInputField`, `inputFieldsActive bool`, `focusedInputField int`
- ✅ Removed old types: `promptInputField` struct, `inputFieldType` enum, helper methods
- ✅ Added new system:
  ```go
  promptEditMode         bool              // Whether prompt edit mode is active (Tab to activate)
  focusedVariableIndex   int               // Index of currently focused variable in template
  filledVariables        map[string]string // Map of variable name -> filled value
  ```
- ✅ Initialized in model.go:
  ```go
  promptEditMode:       false,
  focusedVariableIndex: 0,
  filledVariables:      make(map[string]string),
  ```

### 2. Rendering (render_preview.go)
- ✅ Simplified `renderPromptPreview()` - removed 3-section layout
- ✅ Removed `renderInputFields()` function (was ~160 lines)
- ✅ Added `renderInlineVariables()` helper:
  - Substitutes `{{varName}}` with filled values
  - Highlights focused variable with background color (235) and foreground (220)
  - Shows unfilled variables in placeholder style (39)
  - Shows filled variables in subtle blue (39)
- ✅ Updated content rendering logic to call `renderInlineVariables()` when `m.promptEditMode` is true

### 3. File Loading (file_operations.go)
- ✅ Removed calls to `createInputFields()`
- ✅ Added initialization:
  ```go
  m.filledVariables = make(map[string]string)
  m.promptEditMode = false
  m.focusedVariableIndex = 0
  ```

## What Needs To Be Done ❌

### 1. Remove Dead Code (prompt_parser.go)

**Lines 258-377** contain old functions that reference deleted types. Delete these functions:
- `detectFieldType()` - used inputFieldType enum
- `getFieldColor()` - used inputFieldType enum
- `getFilledVariables()` - used promptInputField struct
- `createInputFields()` - used promptInputField struct

These are no longer needed because we're using `m.filledVariables` map directly.

### 2. Fix helpers.go

**Line 190** references `m.inputFieldsActive` which no longer exists. This is in the `getCurrentFile()` function.

**Find the context** around line 190 and update the logic:
- If it was checking `inputFieldsActive` to determine behavior, replace with `promptEditMode`
- Or remove the check entirely if it's no longer relevant

### 3. Remove Old Keyboard Handling (update_keyboard.go)

**Search for all references to:**
- `inputFieldsActive`
- `promptInputFields`
- `focusedInputField`

**Lines to remove** (approximately):
- Lines 44-172: Old input field keyboard handling block (Tab, Shift+Tab, typing, backspace, Ctrl+U, F3 file picker, F5 copy)
- Line 370: `getFilledVariables(m.promptInputFields, &m)` - replace with `m.filledVariables`
- Lines 936-1000: File picker restore logic for input fields
- Lines 1429-1437: Old Tab navigation for input fields
- Lines 1798-1808: Prompts mode toggle that creates input fields

### 4. Implement New Inline Editing Keyboard Logic (update_keyboard.go)

Add this new logic when viewing a prompt file in prompts mode (F11):

#### A. Tab - Navigate to Next Variable
```go
// In the main keyboard handler, when preview is focused and it's a prompt:
case "tab":
    if m.preview.isPrompt && m.preview.promptTemplate != nil && m.showPromptsOnly {
        if !m.promptEditMode {
            // First Tab press - enter edit mode
            m.promptEditMode = true
            m.focusedVariableIndex = 0
            // Auto-fill defaults for DATE/TIME
            m.autofillDefaults()
        } else {
            // Already in edit mode - navigate to next variable
            if len(m.preview.promptTemplate.variables) > 0 {
                m.focusedVariableIndex++
                if m.focusedVariableIndex >= len(m.preview.promptTemplate.variables) {
                    m.focusedVariableIndex = 0 // Wrap around
                }
            }
        }
        return m, nil
    }
```

#### B. Shift+Tab - Navigate to Previous Variable
```go
case "shift+tab":
    if m.promptEditMode && m.preview.isPrompt && m.preview.promptTemplate != nil {
        if len(m.preview.promptTemplate.variables) > 0 {
            m.focusedVariableIndex--
            if m.focusedVariableIndex < 0 {
                m.focusedVariableIndex = len(m.preview.promptTemplate.variables) - 1 // Wrap around
            }
        }
        return m, nil
    }
```

#### C. Typing - Edit Focused Variable
```go
// In the default case (typing characters):
if m.promptEditMode && m.preview.isPrompt && m.preview.promptTemplate != nil {
    if m.focusedVariableIndex >= 0 && m.focusedVariableIndex < len(m.preview.promptTemplate.variables) {
        varName := m.preview.promptTemplate.variables[m.focusedVariableIndex]

        // Get or create value
        currentValue := m.filledVariables[varName]

        // Append typed character
        text := string(msg.Runes)
        if len(text) > 0 && !isSpecialKey(msg.String()) {
            m.filledVariables[varName] = currentValue + text
            return m, nil
        }
    }
}
```

#### D. Backspace - Delete Character
```go
case "backspace":
    if m.promptEditMode && m.preview.isPrompt && m.preview.promptTemplate != nil {
        if m.focusedVariableIndex >= 0 && m.focusedVariableIndex < len(m.preview.promptTemplate.variables) {
            varName := m.preview.promptTemplate.variables[m.focusedVariableIndex]
            currentValue := m.filledVariables[varName]
            if len(currentValue) > 0 {
                m.filledVariables[varName] = currentValue[:len(currentValue)-1]
            }
            return m, nil
        }
    }
```

#### E. Ctrl+U - Clear Variable
```go
case "ctrl+u":
    if m.promptEditMode && m.preview.isPrompt && m.preview.promptTemplate != nil {
        if m.focusedVariableIndex >= 0 && m.focusedVariableIndex < len(m.preview.promptTemplate.variables) {
            varName := m.preview.promptTemplate.variables[m.focusedVariableIndex]
            m.filledVariables[varName] = ""
            return m, nil
        }
    }
```

#### F. Esc - Exit Edit Mode
```go
case "esc":
    if m.promptEditMode {
        m.promptEditMode = false
        m.setStatusMessage("Exited edit mode", false)
        return m, nil
    }
```

#### G. F5 - Copy with Substitutions
Find the existing F5 handler and update it to use `m.filledVariables` instead of `getFilledVariables()`:

```go
case "f5":
    if m.preview.isPrompt && m.preview.promptTemplate != nil {
        // Get variables - use filled variables with defaults from context
        vars := getContextVariables(&m)

        // Override with user-filled values
        for varName, value := range m.filledVariables {
            if value != "" {
                vars[varName] = value
            }
        }

        // Substitute and copy
        result := substituteVariables(m.preview.promptTemplate.template, vars)
        err := copyToClipboard(result)
        if err == nil {
            m.setStatusMessage("✓ Copied filled prompt to clipboard", false)
        } else {
            m.setStatusMessage("✗ Failed to copy: "+err.Error(), true)
        }
        return m, nil
    }
```

### 5. Add Helper Function (helpers.go or render_preview.go)

Add `autofillDefaults()` method to auto-populate DATE/TIME when entering edit mode:

```go
// autofillDefaults populates DATE and TIME variables with current values
func (m *model) autofillDefaults() {
    if m.preview.promptTemplate == nil {
        return
    }

    contextVars := getContextVariables(m)

    for _, varName := range m.preview.promptTemplate.variables {
        varNameLower := strings.ToLower(varName)

        // Auto-fill DATE and TIME from context
        if varNameLower == "date" || varNameLower == "time" {
            if value, exists := contextVars[varName]; exists {
                m.filledVariables[varName] = value
            }
        }
    }
}
```

### 6. Testing Checklist

After implementing the above:

1. **Build and fix any remaining compile errors**
   ```bash
   go build
   ```

2. **Test basic navigation:**
   - Open a .prompty file
   - Press F11 to enable prompts mode
   - Press Tab - should enter edit mode and highlight first variable
   - Press Tab again - should move to next variable
   - Press Shift+Tab - should move to previous variable

3. **Test editing:**
   - Type characters - should fill the focused variable
   - Backspace - should delete characters
   - Ctrl+U - should clear the variable

4. **Test copy:**
   - Fill in some variables
   - Press F5 - should copy the substituted content to clipboard

5. **Test edge cases:**
   - Esc exits edit mode
   - Tab wraps around from last to first variable
   - DATE and TIME auto-fill on first Tab press
   - Works in both dual-pane and full preview modes
   - Works in narrow (Termux) terminals

## Architecture Notes

### Old System (3-Section):
```
┌─────────────────────────┐
│ 📝 Prompt Name          │ <- Header (1 line)
├─────────────────────────┤
│ Content with {{vars}}   │ <- Content (cramped)
│ highlighted             │
├─────────────────────────┤
│ 📝 Fillable Fields      │ <- Dedicated panel
│ ▶ file: ...             │    (ate 50% of space)
│   project: ...          │
└─────────────────────────┘
```

### New System (Inline):
```
┌─────────────────────────┐
│ 📝 Prompt Name          │ <- Header (1 line)
├─────────────────────────┤
│ **File:** {{file}}      │ <- Tab here to edit
│           ^^^^^^ highlighted when focused
│                         │
│ **Project:** {{project}}│ <- Tab again to move here
│                         │
│ All content visible!    │ <- Much more reading space
└─────────────────────────┘
```

## Key Design Decisions

1. **Tab activates edit mode** - First Tab enters edit mode on first variable
2. **Esc exits edit mode** - Returns to normal preview scrolling
3. **Auto-fill DATE/TIME** - Convenience feature, still editable
4. **Highlight focused variable** - Background 235, foreground 220
5. **Map-based storage** - `m.filledVariables[varName] = value` is simpler than array of structs
6. **No F3 file picker** - Can add later if needed, but simpler UX without it for now

## Benefits

- ✅ **50% more vertical space** for reading prompts (no dedicated fields panel)
- ✅ **~200 fewer lines of code** (removed renderInputFields, createInputFields, field types)
- ✅ **Simpler UX** - WYSIWYG editing right in the content
- ✅ **Better for Termux** - More space on small screens
- ✅ **Cleaner architecture** - No complex 3-section layout calculations

## Current Status

**Compiles:** ❌ (3 errors remaining)
**Functional:** ❌ (needs keyboard implementation)
**Estimated Time:** 45-60 minutes to complete

Good luck! 🚀
