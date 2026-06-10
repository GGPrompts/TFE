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
