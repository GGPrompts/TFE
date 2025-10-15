package main

// Module: command.go
// Purpose: Command prompt functionality for executing shell commands
// Responsibilities:
// - Execute shell commands in the current directory context
// - Handle command completion and results
// - Manage command history (optional)

import (
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

// commandFinishedMsg is sent when a command execution completes
type commandFinishedMsg struct {
	err error
}

// runCommand executes a shell command in the specified directory
// It suspends the TFE UI, runs the command, then resumes
// Uses tea.ClearScreen to prevent phantom text issues
func runCommand(command, dir string) tea.Cmd {
	return func() tea.Msg {
		// Create the command using the system shell
		c := exec.Command("sh", "-c", command)
		c.Dir = dir
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		// Execute the command and restore the terminal
		return tea.Sequence(
			tea.ClearScreen,
			tea.ExecProcess(c, func(err error) tea.Msg {
				return commandFinishedMsg{err: err}
			}),
		)()
	}
}

// addToHistory adds a command to the history, avoiding duplicates
func (m *model) addToHistory(command string) {
	if command == "" {
		return
	}

	// Remove duplicate if it exists
	for i, cmd := range m.commandHistory {
		if cmd == command {
			m.commandHistory = append(m.commandHistory[:i], m.commandHistory[i+1:]...)
			break
		}
	}

	// Add to end of history
	m.commandHistory = append(m.commandHistory, command)

	// Limit history to 100 commands
	if len(m.commandHistory) > 100 {
		m.commandHistory = m.commandHistory[1:]
	}

	// Reset history position
	m.historyPos = len(m.commandHistory)
}

// getPreviousCommand navigates backward in command history
func (m *model) getPreviousCommand() string {
	if len(m.commandHistory) == 0 {
		return m.commandInput
	}

	if m.historyPos > 0 {
		m.historyPos--
	}

	if m.historyPos < len(m.commandHistory) {
		return m.commandHistory[m.historyPos]
	}

	return ""
}

// getNextCommand navigates forward in command history
func (m *model) getNextCommand() string {
	if len(m.commandHistory) == 0 {
		return ""
	}

	if m.historyPos < len(m.commandHistory)-1 {
		m.historyPos++
		return m.commandHistory[m.historyPos]
	}

	// At the end of history, return empty string
	m.historyPos = len(m.commandHistory)
	return ""
}
