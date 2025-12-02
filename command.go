package main

// Module: command.go
// Purpose: Command prompt functionality for executing shell commands
// Responsibilities:
// - Execute shell commands in the current directory context
// - Handle command completion and results
// - Manage command history (optional)

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
// 3. Show exit code and wait for user to press a key before returning
//
// For long-running TUI apps (like claude, lazygit), use ! prefix to exit TFE: :!command
func runCommand(command, dir string) tea.Cmd {
	return func() tea.Msg {
		// Execute command - runs in current directory
		// Commands are executed in a safe wrapper that shows output and waits for keypress
		script := fmt.Sprintf(`
echo "$ %s"
cd %s || exit 1
%s
exitCode=$?
echo ""
echo "Exit code: $exitCode"
echo "Press any key to continue..."
read -n 1 -s -r
exit $exitCode
`, shellQuote(command), shellQuote(dir), command)

		c := exec.Command("bash", "-c", script)
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

// runCommandAndExit executes a shell command and exits TFE
// Used when command is prefixed with ! (e.g., ":!claude --yolo")
// This is useful for launching long-running TUI apps that need to take over the terminal
// The command will run and when it exits, TFE will quit (not resume)
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

// addToHistory adds a command to the current directory's history, avoiding duplicates
// Commands are stored per-directory for context-specific recall
func (m *model) addToHistory(command string) {
	if command == "" {
		return
	}

	// Get current directory history
	dirHistory := m.commandHistoryByDir[m.currentPath]

	// Remove duplicate if it exists in directory history
	for i, cmd := range dirHistory {
		if cmd == command {
			dirHistory = append(dirHistory[:i], dirHistory[i+1:]...)
			break
		}
	}

	// Add to end of directory history
	dirHistory = append(dirHistory, command)

	// Limit directory history to 50 commands
	if len(dirHistory) > 50 {
		dirHistory = dirHistory[1:]
	}

	// Save back to map
	m.commandHistoryByDir[m.currentPath] = dirHistory

	// Rebuild combined history (directory + global)
	m.rebuildCombinedHistory()

	// Reset history position
	m.historyPos = len(m.commandHistory)

	// Save to disk after adding
	m.saveCommandHistory()
}

// rebuildCombinedHistory creates a combined history list from current directory + global
// Directory-specific commands appear first (most relevant), then global commands
// Duplicates are removed (directory version takes precedence)
func (m *model) rebuildCombinedHistory() {
	seen := make(map[string]bool)
	combined := []string{}

	// Add directory-specific commands first (most relevant)
	if dirHistory, exists := m.commandHistoryByDir[m.currentPath]; exists {
		for _, cmd := range dirHistory {
			if !seen[cmd] {
				combined = append(combined, cmd)
				seen[cmd] = true
			}
		}
	}

	// Add global commands (if not already in directory history)
	for _, cmd := range m.commandHistoryGlobal {
		if !seen[cmd] {
			combined = append(combined, cmd)
			seen[cmd] = true
		}
	}

	m.commandHistory = combined
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

// termuxNewSession launches a command in a new Termux terminal session
// This is useful when TFE is launched from a widget (no parent shell)
// For Quick CD: command is "exec bash" to get an interactive shell in the target dir
// For Claude: command is "claude" to launch Claude Code
// Requires allow-external-apps = true in ~/.termux/termux.properties
func termuxNewSession(command string, workDir string) tea.Cmd {
	return func() tea.Msg {
		amCmd := termuxNewSessionCmd(command, workDir)

		c := exec.Command("bash", "-c", amCmd)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		err := c.Run()
		if err != nil {
			// Return error message that can be displayed
			return commandFinishedMsg{err: fmt.Errorf("failed to start Termux session: %v (ensure allow-external-apps=true in ~/.termux/termux.properties)", err)}
		}

		// After launching the new session, quit TFE
		return tea.Quit()
	}
}

// shellQuote quotes a string for safe use in shell commands
// Simple version that escapes single quotes
func shellQuote(s string) string {
	// Replace single quotes with '\'' (end quote, escaped quote, start quote)
	s = strings.ReplaceAll(s, "'", "'\\''")
	return "'" + s + "'"
}

// loadCommandHistory reads command history from disk
// Returns directory-specific map and global slice
// Handles backwards compatibility with old format
func loadCommandHistory() (map[string][]string, []string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return make(map[string][]string), []string{}
	}

	historyPath := filepath.Join(homeDir, ".config", "tfe", "command_history.json")
	data, err := os.ReadFile(historyPath)
	if err != nil {
		return make(map[string][]string), []string{} // File doesn't exist yet, start fresh
	}

	// Try new format first
	var newFormat struct {
		Version     int                 `json:"version"`
		Directories map[string][]string `json:"directories"`
		Global      []string            `json:"global"`
	}

	if err := json.Unmarshal(data, &newFormat); err == nil && newFormat.Version == 2 {
		// New format loaded successfully
		if newFormat.Directories == nil {
			newFormat.Directories = make(map[string][]string)
		}
		if newFormat.Global == nil {
			newFormat.Global = []string{}
		}
		return newFormat.Directories, newFormat.Global
	}

	// Try old format for backwards compatibility
	var oldFormat struct {
		Commands []string `json:"commands"`
	}

	if err := json.Unmarshal(data, &oldFormat); err == nil && len(oldFormat.Commands) > 0 {
		// Old format - migrate to global history
		return make(map[string][]string), oldFormat.Commands
	}

	// Failed to parse either format
	return make(map[string][]string), []string{}
}

// saveCommandHistory writes command history to disk
// Creates the config directory if it doesn't exist
// Saves in version 2 format with per-directory and global history
func (m *model) saveCommandHistory() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".config", "tfe")
	os.MkdirAll(configDir, 0755) // Create directory if it doesn't exist

	historyPath := filepath.Join(configDir, "command_history.json")

	history := struct {
		Version     int                 `json:"version"`
		MaxSize     int                 `json:"maxSize"`
		Directories map[string][]string `json:"directories"`
		Global      []string            `json:"global"`
	}{
		Version:     2,
		MaxSize:     100,
		Directories: m.commandHistoryByDir,
		Global:      m.commandHistoryGlobal,
	}

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(historyPath, data, 0644)
}
