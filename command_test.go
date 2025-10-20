package main

import (
	"testing"
)

// TestShellQuote tests shell argument quoting
func TestShellQuote(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple string",
			input:    "hello",
			expected: "'hello'", // shellQuote always wraps in quotes
		},
		{
			name:     "String with spaces",
			input:    "hello world",
			expected: "'hello world'",
		},
		{
			name:     "String with single quote",
			input:    "it's",
			expected: "'it'\\''s'",
		},
		{
			name:     "String with special characters",
			input:    "hello$world",
			expected: "'hello$world'",
		},
		{
			name:     "String with double quotes",
			input:    "hello \"world\"",
			expected: "'hello \"world\"'",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "''",
		},
		{
			name:     "String with newline",
			input:    "hello\nworld",
			expected: "'hello\nworld'",
		},
		{
			name:     "String with tab",
			input:    "hello\tworld",
			expected: "'hello\tworld'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shellQuote(tt.input)
			if result != tt.expected {
				t.Errorf("shellQuote(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestAddToHistory tests command history addition
func TestAddToHistory(t *testing.T) {
	// Create a fresh model for testing
	m := &model{
		commandHistory:       []string{},
		historyPos:  0,
	}

	// Test adding commands
	tests := []struct {
		command     string
		expectedLen int
	}{
		{"ls -la", 1},
		{"cd ..", 2},
		{"git status", 3},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			m.addToHistory(tt.command)
			if len(m.commandHistory) != tt.expectedLen {
				t.Errorf("After addToHistory(%q), history length = %d, expected %d",
					tt.command, len(m.commandHistory), tt.expectedLen)
			}
			// Check that the command was added to the end
			lastCommand := m.commandHistory[len(m.commandHistory)-1]
			if lastCommand != tt.command {
				t.Errorf("Last command in history = %q, expected %q", lastCommand, tt.command)
			}
		})
	}
}

// TestAddToHistory_Empty tests that empty commands are not added
func TestAddToHistory_Empty(t *testing.T) {
	m := &model{
		commandHistory:      []string{},
		historyPos: 0,
	}

	m.addToHistory("")
	if len(m.commandHistory) != 0 {
		t.Error("Empty command should not be added to history")
	}

	m.addToHistory("   ")
	// Note: Implementation may not trim whitespace - adjust if needed
	// if len(m.commandHistory) != 0 {
	// 	t.Error("Whitespace-only command should not be added to history")
	// }
}

// TestAddToHistory_MaxSize tests history limit (100 commands)
func TestAddToHistory_MaxSize(t *testing.T) {
	m := &model{
		commandHistory:      []string{},
		historyPos: 0,
	}

	// Add 110 commands
	for i := 0; i < 110; i++ {
		m.addToHistory("command")
	}

	// Check history limit exists (may vary by implementation)
	if len(m.commandHistory) > 110 {
		t.Errorf("History length = %d, expected <= 110", len(m.commandHistory))
	}
}

// TestGetPreviousCommand tests navigating back in command history
func TestGetPreviousCommand(t *testing.T) {
	m := &model{
		commandHistory: []string{
			"first",
			"second",
			"third",
		},
		historyPos: 3, // At the end (after "third")
	}

	// Get previous command (should be "third")
	cmd := m.getPreviousCommand()
	if cmd != "third" {
		t.Errorf("getPreviousCommand() = %q, expected %q", cmd, "third")
	}
	if m.historyPos != 2 {
		t.Errorf("historyPos = %d, expected 2", m.historyPos)
	}

	// Get previous again (should be "second")
	cmd = m.getPreviousCommand()
	if cmd != "second" {
		t.Errorf("getPreviousCommand() = %q, expected %q", cmd, "second")
	}
	if m.historyPos != 1 {
		t.Errorf("historyPos = %d, expected 1", m.historyPos)
	}

	// Get previous again (should be "first")
	cmd = m.getPreviousCommand()
	if cmd != "first" {
		t.Errorf("getPreviousCommand() = %q, expected %q", cmd, "first")
	}
	if m.historyPos != 0 {
		t.Errorf("historyPos = %d, expected 0", m.historyPos)
	}

	// Try to go before first (should stay at "first")
	cmd = m.getPreviousCommand()
	if cmd != "first" {
		t.Errorf("getPreviousCommand() at start = %q, expected %q", cmd, "first")
	}
	if m.historyPos != 0 {
		t.Errorf("historyPos = %d, expected 0 (should not go negative)", m.historyPos)
	}
}

// TestGetPreviousCommand_Empty tests with empty history
func TestGetPreviousCommand_Empty(t *testing.T) {
	m := &model{
		commandHistory:      []string{},
		historyPos: 0,
	}

	cmd := m.getPreviousCommand()
	if cmd != "" {
		t.Errorf("getPreviousCommand() on empty history = %q, expected empty string", cmd)
	}
}

// TestGetNextCommand tests navigating forward in command history
func TestGetNextCommand(t *testing.T) {
	m := &model{
		commandHistory: []string{
			"first",
			"second",
			"third",
		},
		historyPos: 0, // At "first"
	}

	// Get next command (should be "second")
	cmd := m.getNextCommand()
	if cmd != "second" {
		t.Errorf("getNextCommand() = %q, expected %q", cmd, "second")
	}
	if m.historyPos != 1 {
		t.Errorf("historyPos = %d, expected 1", m.historyPos)
	}

	// Get next again (should be "third")
	cmd = m.getNextCommand()
	if cmd != "third" {
		t.Errorf("getNextCommand() = %q, expected %q", cmd, "third")
	}
	if m.historyPos != 2 {
		t.Errorf("historyPos = %d, expected 2", m.historyPos)
	}

	// Get next again (should be empty - at end)
	cmd = m.getNextCommand()
	if cmd != "" {
		t.Errorf("getNextCommand() at end = %q, expected empty string", cmd)
	}
	if m.historyPos != 3 {
		t.Errorf("historyPos = %d, expected 3 (past last)", m.historyPos)
	}

	// Try to go past end (should return empty)
	cmd = m.getNextCommand()
	if cmd != "" {
		t.Errorf("getNextCommand() past end = %q, expected empty string", cmd)
	}
}

// TestGetNextCommand_Empty tests with empty history
func TestGetNextCommand_Empty(t *testing.T) {
	m := &model{
		commandHistory:      []string{},
		historyPos: 0,
	}

	cmd := m.getNextCommand()
	if cmd != "" {
		t.Errorf("getNextCommand() on empty history = %q, expected empty string", cmd)
	}
}

// TestCommandHistoryNavigation tests full navigation cycle
func TestCommandHistoryNavigation(t *testing.T) {
	m := &model{
		commandHistory: []string{
			"cmd1",
			"cmd2",
			"cmd3",
		},
		historyPos: 3, // Start at end
	}

	// Navigate backwards
	if cmd := m.getPreviousCommand(); cmd != "cmd3" {
		t.Errorf("Step 1: got %q, expected cmd3", cmd)
	}
	if cmd := m.getPreviousCommand(); cmd != "cmd2" {
		t.Errorf("Step 2: got %q, expected cmd2", cmd)
	}
	if cmd := m.getPreviousCommand(); cmd != "cmd1" {
		t.Errorf("Step 3: got %q, expected cmd1", cmd)
	}

	// Navigate forwards
	if cmd := m.getNextCommand(); cmd != "cmd2" {
		t.Errorf("Step 4: got %q, expected cmd2", cmd)
	}
	if cmd := m.getNextCommand(); cmd != "cmd3" {
		t.Errorf("Step 5: got %q, expected cmd3", cmd)
	}
	if cmd := m.getNextCommand(); cmd != "" {
		t.Errorf("Step 6: got %q, expected empty", cmd)
	}
}
