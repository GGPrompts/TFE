package main

// Module: ghost_text.go
// Purpose: Haiku-powered ghost text suggestions for the command prompt
// Responsibilities:
// - Debounced API requests (300ms after last keystroke)
// - Anthropic API integration for command suggestions
// - Building context (cwd, file list, selection) for the prompt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ghostTextDebounceDelay is the delay after the last keystroke before requesting ghost text
const ghostTextDebounceDelay = 300 * time.Millisecond

// ghostTextMinInput is the minimum input length before requesting ghost text
const ghostTextMinInput = 2

// requestGhostText creates a tea.Cmd that waits for the debounce delay, then calls
// the Anthropic API to get a command suggestion. The seq parameter is used to
// discard stale responses (if the user typed more while the request was in flight).
func requestGhostText(input string, cwd string, files []string, selectedFile string, seq int) tea.Cmd {
	return func() tea.Msg {
		// Debounce: wait before making the request
		time.Sleep(ghostTextDebounceDelay)

		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			return ghostTextMsg{seq: seq, err: fmt.Errorf("ANTHROPIC_API_KEY not set")}
		}

		// Build file list context (limit to avoid huge prompts)
		fileList := strings.Join(files, ", ")
		if len(fileList) > 1500 {
			// Truncate at a rune boundary to avoid splitting multi-byte UTF-8
			runes := []rune(fileList)
			if len(runes) > 500 {
				fileList = string(runes[:500]) + "..."
			}
		}

		systemPrompt := `You are a shell command autocomplete assistant embedded in a terminal file explorer (TFE).
Your job is to suggest a SINGLE shell command completion based on the user's partial input.

Rules:
- Only suggest valid shell commands (bash/zsh compatible)
- Focus on file operations: mv, cp, rm, mkdir, find, grep, ls, cat, chmod, chown, tar, zip, git, etc.
- Complete the user's partial input into a full command
- Use the provided working directory and file list for accurate path completions
- Output ONLY the completed command text (no explanation, no quotes, no markdown)
- If the input already looks like a complete command, suggest a reasonable extension or return it as-is
- Do NOT include the partial input prefix - return the FULL command
- Keep suggestions concise and practical
- If you cannot suggest anything meaningful, respond with just the original input`

		selectedContext := ""
		if selectedFile != "" {
			selectedContext = fmt.Sprintf("\nCurrently selected file: %s", selectedFile)
		}

		userPrompt := fmt.Sprintf(`Working directory: %s
Files in current directory: %s%s

User's partial input: %s

Suggest the completed command:`, cwd, fileList, selectedContext, input)

		// Build API request
		reqBody := map[string]interface{}{
			"model":      "claude-haiku-4-5-20250514",
			"max_tokens": 150,
			"system":     systemPrompt,
			"messages": []map[string]string{
				{"role": "user", "content": userPrompt},
			},
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			return ghostTextMsg{seq: seq, err: err}
		}

		req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(jsonBody))
		if err != nil {
			return ghostTextMsg{seq: seq, err: err}
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return ghostTextMsg{seq: seq, err: err}
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return ghostTextMsg{seq: seq, err: err}
		}

		if resp.StatusCode != 200 {
			return ghostTextMsg{seq: seq, err: fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))}
		}

		// Parse response
		var apiResp struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		}

		if err := json.Unmarshal(body, &apiResp); err != nil {
			return ghostTextMsg{seq: seq, err: err}
		}

		if len(apiResp.Content) == 0 {
			return ghostTextMsg{seq: seq, err: fmt.Errorf("empty response")}
		}

		suggestion := strings.TrimSpace(apiResp.Content[0].Text)

		// Clean up: remove backticks or code fences that the model might add
		suggestion = strings.TrimPrefix(suggestion, "```")
		suggestion = strings.TrimSuffix(suggestion, "```")
		suggestion = strings.TrimPrefix(suggestion, "`")
		suggestion = strings.TrimSuffix(suggestion, "`")
		suggestion = strings.TrimSpace(suggestion)

		// If suggestion equals input exactly, no ghost text needed
		if suggestion == input {
			return ghostTextMsg{seq: seq, suggestion: ""}
		}

		return ghostTextMsg{seq: seq, suggestion: suggestion}
	}
}

// getGhostTextSuffix returns only the part of the ghost text that extends
// beyond the current input. If the suggestion starts with the input, return
// the suffix; otherwise return the full suggestion.
func getGhostTextSuffix(input, suggestion string) string {
	if suggestion == "" {
		return ""
	}

	// If suggestion starts with the current input, show only the completion part
	if strings.HasPrefix(suggestion, input) {
		return suggestion[len(input):]
	}

	// Otherwise the model suggested a different command entirely;
	// show the whole thing as ghost text (user can tab to accept)
	return " [" + suggestion + "]"
}

// buildFileListForGhostText returns a slice of filenames in the current directory
// suitable for sending to the API as context.
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

// triggerGhostText increments the sequence counter and returns a Cmd to request ghost text.
// Should be called after each keystroke that modifies commandInput.
func (m *model) triggerGhostText() tea.Cmd {
	// Don't request if input is too short
	if len(m.commandInput) < ghostTextMinInput {
		m.ghostText = ""
		m.ghostTextLoading = false
		return nil
	}

	// Don't request if no API key
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		return nil
	}

	m.ghostTextSeq++
	m.ghostTextLoading = true

	selectedName := ""
	if f := m.getCurrentFile(); f != nil {
		selectedName = f.name
	}

	return requestGhostText(
		m.commandInput,
		m.currentPath,
		m.buildFileListForGhostText(),
		selectedName,
		m.ghostTextSeq,
	)
}

// clearGhostText removes any ghost text suggestion and cancels pending requests.
func (m *model) clearGhostText() {
	m.ghostText = ""
	m.ghostTextLoading = false
	m.ghostTextSeq++ // Invalidate any in-flight requests
}

// acceptGhostText applies the ghost text suggestion to the command input.
// Returns true if ghost text was accepted, false if there was nothing to accept.
func (m *model) acceptGhostText() bool {
	if m.ghostText == "" {
		return false
	}

	suffix := getGhostTextSuffix(m.commandInput, m.ghostText)
	if suffix == "" || strings.HasPrefix(suffix, " [") {
		// Full replacement suggestion
		if strings.HasPrefix(suffix, " [") {
			// Extract the actual suggestion from " [suggestion]"
			m.commandInput = m.ghostText
		} else {
			return false
		}
	} else {
		// Append the suffix
		m.commandInput += suffix
	}

	m.commandCursorPos = len(m.commandInput)
	m.ghostText = ""
	m.ghostTextLoading = false
	return true
}
