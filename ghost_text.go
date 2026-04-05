package main

// Module: ghost_text.go
// Purpose: AI-assisted command suggestions via claude CLI
// Responsibilities:
// - Handling ? prefix in command prompt to ask Haiku for command suggestions
// - Running claude -p with file context, suspending TFE to show the response
// - Capturing the suggested command and pre-filling it as ghost text

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ghostTextFinishedMsg is sent when the claude -p process completes
type ghostTextFinishedMsg struct {
	suggestion string // The suggested command to pre-fill
	err        error
}

// runGhostTextQuery runs claude -p with the user's question and file context.
// It suspends TFE (like runCommand), shows the response in the terminal,
// and captures the suggested command in a temp file for ghost text pre-fill.
func runGhostTextQuery(question string, cwd string, files []string, selectedFile string) tea.Cmd {
	return func() tea.Msg {
		// Build file context for the system prompt
		fileList := strings.Join(files, ", ")
		runes := []rune(fileList)
		if len(runes) > 500 {
			fileList = string(runes[:500]) + "..."
		}

		selectedContext := ""
		if selectedFile != "" {
			selectedContext = fmt.Sprintf("\nCurrently selected file: %s", selectedFile)
		}

		systemPrompt := fmt.Sprintf(`You suggest shell commands for file operations.

Working directory: %s
Files: %s%s

RULES:
- Output ONLY the shell command, nothing else
- No markdown, no backticks, no code fences, no formatting
- No explanations, no commentary, no follow-up text
- No "Here's the command:" or similar prefixes
- No "Let me know" or similar closers
- Just the raw command, one line, plain text
- If no command applies, output only: NO_COMMAND`, cwd, fileList, selectedContext)

		// Write system prompt to a temp file (avoids shell escaping issues)
		tmpFile, err := os.CreateTemp("", "tfe-ghost-*.txt")
		if err != nil {
			return ghostTextFinishedMsg{err: err}
		}
		systemPromptFile := tmpFile.Name()
		tmpFile.WriteString(systemPrompt)
		tmpFile.Close()

		// Temp file to capture the suggestion (last line of output)
		suggestionFile := systemPromptFile + ".suggestion"

		// Build the script: run claude -p, capture full output, display it, save last line
		outputFile := systemPromptFile + ".output"
		script := fmt.Sprintf(`
echo "? %s"
echo "---"
SYSTEM_PROMPT=$(cat %s)
claude -p %s --model haiku --system-prompt "$SYSTEM_PROMPT" --bare --tools "" --no-input 2>/dev/null > %s
cat %s
tail -1 %s > %s
echo ""
echo "Press any key to continue..."
read -n 1 -s -r
rm -f %s %s
`,
			shellQuote(question),
			shellQuote(systemPromptFile),
			shellQuote(question),
			shellQuote(outputFile),
			shellQuote(outputFile),
			shellQuote(outputFile),
			shellQuote(suggestionFile),
			shellQuote(systemPromptFile),
			shellQuote(outputFile),
		)

		c := exec.Command("bash", "-c", script)
		c.Dir = cwd
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		return tea.Sequence(
			tea.ClearScreen,
			tea.ExecProcess(c, func(err error) tea.Msg {
				// Read the captured suggestion
				suggestion := ""
				if data, readErr := os.ReadFile(suggestionFile); readErr == nil {
					suggestion = strings.TrimSpace(string(data))
					os.Remove(suggestionFile)
				}

				// Clean up suggestion
				if suggestion == "NO_COMMAND" || suggestion == "" {
					suggestion = ""
				}
				// Strip backticks/fences the model might add
				suggestion = strings.TrimPrefix(suggestion, "```")
				suggestion = strings.TrimSuffix(suggestion, "```")
				suggestion = strings.TrimPrefix(suggestion, "`")
				suggestion = strings.TrimSuffix(suggestion, "`")
				suggestion = strings.TrimSpace(suggestion)

				return ghostTextFinishedMsg{suggestion: suggestion, err: err}
			}),
		)()
	}
}

// buildFileListForGhostText returns a slice of filenames in the current directory
// suitable for sending as context.
func (m model) buildFileListForGhostText() []string {
	names := make([]string, 0, len(m.files))
	for _, f := range m.files {
		if f.name == ".." {
			continue
		}
		name := f.name
		if f.isDir {
			name += "/"
		}
		names = append(names, name)
	}
	// Cap at 100 files to keep prompt size reasonable
	if len(names) > 100 {
		names = names[:100]
	}
	return names
}

// getGhostTextSuffix returns only the part of the ghost text that extends
// beyond the current input. If the suggestion starts with the input, return
// the suffix; otherwise return the full suggestion.
func getGhostTextSuffix(input, suggestion string) string {
	if suggestion == "" {
		return ""
	}
	if strings.HasPrefix(suggestion, input) {
		return suggestion[len(input):]
	}
	return suggestion
}

// clearGhostText removes any ghost text suggestion.
func (m *model) clearGhostText() {
	m.ghostText = ""
	m.ghostTextLoading = false
}

// acceptGhostText applies the ghost text suggestion to the command input.
func (m *model) acceptGhostText() bool {
	if m.ghostText == "" {
		return false
	}
	m.commandInput = m.ghostText
	m.commandCursorPos = len(m.commandInput)
	m.ghostText = ""
	m.ghostTextLoading = false
	return true
}

// handleQuestionPrefix checks if the command starts with ? and runs the ghost text query.
// Returns true and a Cmd if handled, false otherwise.
func (m *model) handleQuestionPrefix(cmd string) (bool, tea.Cmd) {
	if !strings.HasPrefix(cmd, "?") {
		return false, nil
	}

	question := strings.TrimSpace(strings.TrimPrefix(cmd, "?"))
	if question == "" {
		m.setStatusMessage("Usage: ?<question> — ask AI for a command suggestion", false)
		return true, nil
	}

	selectedName := ""
	if f := m.getCurrentFile(); f != nil {
		selectedName = f.name
	}

	return true, runGhostTextQuery(
		question,
		m.currentPath,
		m.buildFileListForGhostText(),
		selectedName,
	)
}

