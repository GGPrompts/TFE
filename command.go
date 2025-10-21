package main

// Module: command.go
// Purpose: Command prompt functionality for executing shell commands
// Responsibilities:
// - Execute shell commands in the current directory context
// - Handle command completion and results
// - Manage command history (optional)

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// commandFinishedMsg is sent when a command execution completes
type commandFinishedMsg struct {
	err error
}

// runCommand executes a shell command in the specified directory
// It suspends the TFE UI, runs the command, then resumes
// Similar to Midnight Commander's "pause after run" feature:
// 1. Echo the command that was typed
// 2. Execute the command and show output
// 3. Wait for user to press a key before returning
func runCommand(command, dir string) tea.Cmd {
	return func() tea.Msg {
		// Build a shell script that:
		// 1. Echoes the command being run
		// 2. Executes the command
		// 3. Prompts user to press any key to continue
		// Note: Use bash instead of sh for better read support
		script := fmt.Sprintf(`
echo "$ %s"
cd %s || exit 1
%s
echo ""
echo "Press any key to continue..."
read -n 1 -s -r
`, shellQuote(command), shellQuote(dir), command)

		// Create the command using bash for better compatibility with read -n
		c := exec.Command("bash", "-c", script)
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

// runCommandAndExit executes a shell command and exits TFE
// Used when command is prefixed with ! (e.g., ":!claude --yolo")
// This is useful for launching long-running TUI apps like Claude Code
func runCommandAndExit(command, dir string) tea.Cmd {
	return func() tea.Msg {
		// Build a shell script that changes to directory and runs command
		script := fmt.Sprintf(`
cd %s || exit 1
exec %s
`, shellQuote(dir), command)

		// Create the command
		c := exec.Command("bash", "-c", script)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		// Execute command and exit TFE immediately
		// The command will take over the terminal
		return tea.Sequence(
			tea.ClearScreen,
			tea.ExecProcess(c, func(err error) tea.Msg {
				// After command exits, quit TFE
				return tea.Quit()
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

// runScript executes a script file safely without command injection
// Similar to runCommand but for executing script files directly
func runScript(scriptPath string) tea.Cmd {
	return func() tea.Msg {
		// Create a wrapper script that:
		// 1. Shows the script being executed
		// 2. Runs the script
		// 3. Pauses for user input
		wrapperScript := `
echo "$ bash $0"
echo ""
bash "$0"
exitCode=$?
echo ""
echo "Exit code: $exitCode"
echo "Press any key to continue..."
read -n 1 -s -r
exit $exitCode
`
		// Execute bash with the wrapper script and pass scriptPath as $0
		c := exec.Command("bash", "-c", wrapperScript, scriptPath)
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

// shellQuote quotes a string for safe use in shell commands
// Simple version that escapes single quotes
func shellQuote(s string) string {
	// Replace single quotes with '\'' (end quote, escaped quote, start quote)
	s = strings.ReplaceAll(s, "'", "'\\''")
	return "'" + s + "'"
}
