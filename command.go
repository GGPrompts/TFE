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
// 3. Wait for user to press a key before returning
//
// SECURITY: Uses command allowlist to prevent arbitrary command execution.
// For more flexibility, use the ! prefix to bypass restrictions (e.g., ":!your-command")
func runCommand(command, dir string) tea.Cmd {
	return func() tea.Msg {
		// Parse command into parts
		parts := strings.Fields(command)
		if len(parts) == 0 {
			return commandFinishedMsg{err: fmt.Errorf("empty command")}
		}

		// Allowlist of safe read-only and utility commands
		// These commands are deemed safe for general use in a file explorer
		safeCommands := map[string]bool{
			"ls":     true, // List directory
			"cat":    true, // Display file contents
			"grep":   true, // Search text
			"find":   true, // Find files
			"head":   true, // Display file start
			"tail":   true, // Display file end
			"wc":     true, // Count lines/words
			"file":   true, // Determine file type
			"git":    true, // Git operations (read: status, log, diff)
			"tree":   true, // Directory tree
			"du":     true, // Disk usage
			"df":     true, // Disk free space
			"pwd":    true, // Print working directory
			"date":   true, // Display date/time
			"whoami": true, // Display current user
			"echo":   true, // Print text
			"which":  true, // Locate command
			"stat":   true, // File statistics
			"diff":   true, // Compare files
			"sort":   true, // Sort lines
			"uniq":   true, // Filter duplicate lines
			"cut":    true, // Extract columns
			"awk":    true, // Text processing
			"sed":    true, // Stream editor
			"less":   true, // Pager
			"more":   true, // Pager
			"hexdump": true, // Hex viewer
			"strings": true, // Extract strings
		}

		executable := parts[0]
		if !safeCommands[executable] {
			// Build wrapper script that shows error and prompts
			errMsg := fmt.Sprintf("⚠️  Command not allowed: %s\n\nFor security, only safe read-only commands are allowed.\nUse the ! prefix to execute without restrictions:\n  Example: :!%s\n\nSafe commands: ls, cat, grep, find, git, tree, du, etc.\nSee HOTKEYS.md for full list.", executable, command)
			script := fmt.Sprintf(`
echo %s
echo ""
echo "Press any key to continue..."
read -n 1 -s -r
`, shellQuote(errMsg))

			c := exec.Command("bash", "-c", script)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr

			return tea.Sequence(
				tea.ClearScreen,
				tea.ExecProcess(c, func(err error) tea.Msg {
					return commandFinishedMsg{err: err}
				}),
			)()
		}

		// Execute command safely without shell interpretation
		// This prevents injection attacks like "ls; rm -rf ~"
		script := fmt.Sprintf(`
echo "$ %s"
cd %s || exit 1
exec %s "$@"
echo ""
echo "Press any key to continue..."
read -n 1 -s -r
`, shellQuote(command), shellQuote(dir), shellQuote(executable))

		// Pass arguments after executable as separate parameters
		args := []string{"-c", script, "--"}
		args = append(args, parts[1:]...)

		c := exec.Command("bash", args...)
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
// This is useful for launching long-running TUI apps like Claude Code
//
// SECURITY NOTE: This function intentionally allows ANY command when using ! prefix.
// This is a power-user feature for flexibility. The security boundary is:
// - Regular commands (:ls) → Allowlist enforced in runCommand()
// - Unrestricted commands (:!anything) → Full shell access via this function
// Users must understand that ! prefix gives full shell access.
func runCommandAndExit(command, dir string) tea.Cmd {
	return func() tea.Msg {
		// Build a shell script that changes to directory and runs command
		// Note: command is still inserted into shell script, but this is intentional
		// for the ! prefix feature. Users explicitly request unrestricted access.
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

	// Save to disk after adding
	m.saveCommandHistory()
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

// loadCommandHistory reads command history from disk
// Returns empty slice if file doesn't exist or on error
func loadCommandHistory() []string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return []string{}
	}

	historyPath := filepath.Join(homeDir, ".config", "tfe", "command_history.json")
	data, err := os.ReadFile(historyPath)
	if err != nil {
		return []string{} // File doesn't exist yet, start fresh
	}

	var history struct {
		Commands []string `json:"commands"`
	}

	if err := json.Unmarshal(data, &history); err != nil {
		return []string{}
	}

	return history.Commands
}

// saveCommandHistory writes command history to disk
// Creates the config directory if it doesn't exist
func (m *model) saveCommandHistory() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".config", "tfe")
	os.MkdirAll(configDir, 0755) // Create directory if it doesn't exist

	historyPath := filepath.Join(configDir, "command_history.json")

	history := struct {
		Commands []string `json:"commands"`
		MaxSize  int      `json:"maxSize"`
	}{
		Commands: m.commandHistory,
		MaxSize:  100,
	}

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(historyPath, data, 0644)
}
