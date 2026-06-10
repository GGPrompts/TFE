package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestCommandInput_StaleCursorInsert is a regression test for tfe-r3s:
// several code paths cleared m.commandInput without resetting
// m.commandCursorPos, so typing a character afterwards sliced an empty
// string with a stale cursor (`""[:2]`) and panicked. The keyboard handler
// now clamps the cursor before any command-input slicing.
func TestCommandInput_StaleCursorInsert(t *testing.T) {
	m := model{
		commandFocused:   true,
		commandInput:     "", // cleared by e.g. Enter / toolbar click / right-click
		commandCursorPos: 2,  // stale cursor left over from previous input ("ls")
	}

	// Typing a character must not panic and must insert at a clamped position
	newModel, _ := m.handleKeyEvent(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	result := newModel.(model)

	if result.commandInput != "a" {
		t.Errorf("commandInput = %q, expected %q", result.commandInput, "a")
	}
	if result.commandCursorPos != 1 {
		t.Errorf("commandCursorPos = %d, expected 1", result.commandCursorPos)
	}
}

// TestCommandInput_StaleCursorBackspace verifies backspace with a stale
// cursor position (beyond the input length) does not panic (tfe-r3s).
func TestCommandInput_StaleCursorBackspace(t *testing.T) {
	m := model{
		commandFocused:   true,
		commandInput:     "x",
		commandCursorPos: 5, // stale: beyond len(commandInput)
	}

	newModel, _ := m.handleKeyEvent(tea.KeyMsg{Type: tea.KeyBackspace})
	result := newModel.(model)

	if result.commandInput != "" {
		t.Errorf("commandInput = %q, expected empty string", result.commandInput)
	}
	if result.commandCursorPos != 0 {
		t.Errorf("commandCursorPos = %d, expected 0", result.commandCursorPos)
	}
}

// TestCommandInput_NegativeCursorClamped verifies a negative cursor position
// is clamped to 0 before any slicing (tfe-r3s defense in depth).
func TestCommandInput_NegativeCursorClamped(t *testing.T) {
	m := model{
		commandFocused:   true,
		commandInput:     "ab",
		commandCursorPos: -1,
	}

	newModel, _ := m.handleKeyEvent(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	result := newModel.(model)

	if result.commandInput != "cab" {
		t.Errorf("commandInput = %q, expected %q", result.commandInput, "cab")
	}
	if result.commandCursorPos != 1 {
		t.Errorf("commandCursorPos = %d, expected 1", result.commandCursorPos)
	}
}

// TestPromptEditMode_JKInsertIntoVariable is a regression test for tfe-36e:
// the prompt edit mode block bound "j"/"k" to preview scrolling, so those
// letters never reached the default rune-insertion handler and values like
// "json" or "kitchen" could not be typed into template variables. In edit
// mode, plain letters must insert; arrows still scroll.
func TestPromptEditMode_JKInsertIntoVariable(t *testing.T) {
	newEditModel := func() model {
		return model{
			promptEditMode:       true,
			focusedVariableIndex: 0,
			filledVariables:      make(map[string]string),
			preview: previewModel{
				isPrompt: true,
				promptTemplate: &promptTemplate{
					name:      "test",
					variables: []string{"TOPIC"},
					template:  "Write about {{TOPIC}}",
				},
				scrollPos: 5,
			},
		}
	}

	for _, r := range []rune{'j', 'k'} {
		m := newEditModel()
		newModel, _ := m.handleKeyEvent(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		result := newModel.(model)

		if got := result.filledVariables["TOPIC"]; got != string(r) {
			t.Errorf("typing %q: filledVariables[TOPIC] = %q, expected %q", r, got, string(r))
		}
		if result.preview.scrollPos != 5 {
			t.Errorf("typing %q: scrollPos = %d, expected 5 (must not scroll in edit mode)", r, result.preview.scrollPos)
		}
	}

	// Arrow keys must still scroll while editing
	m := newEditModel()
	newModel, _ := m.handleKeyEvent(tea.KeyMsg{Type: tea.KeyUp})
	result := newModel.(model)
	if result.preview.scrollPos != 4 {
		t.Errorf("up arrow: scrollPos = %d, expected 4", result.preview.scrollPos)
	}
	if got := result.filledVariables["TOPIC"]; got != "" {
		t.Errorf("up arrow: filledVariables[TOPIC] = %q, expected empty", got)
	}
}

// TestCommandInput_EnterResetsCursor verifies that executing a command via
// Enter resets the cursor position along with the input (tfe-r3s).
func TestCommandInput_EnterResetsCursor(t *testing.T) {
	m := model{
		commandFocused:      true,
		commandInput:        "ls",
		commandCursorPos:    2,
		commandHistoryByDir: make(map[string][]string),
	}

	newModel, _ := m.handleKeyEvent(tea.KeyMsg{Type: tea.KeyEnter})
	result := newModel.(model)

	if result.commandInput != "" {
		t.Errorf("commandInput = %q, expected empty string after Enter", result.commandInput)
	}
	if result.commandCursorPos != 0 {
		t.Errorf("commandCursorPos = %d, expected 0 after Enter", result.commandCursorPos)
	}
	if result.commandFocused {
		t.Error("commandFocused should be false after executing a command")
	}
}
