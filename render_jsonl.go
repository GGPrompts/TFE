package main

// Module: render_jsonl.go
// Purpose: JSONL conversation file preview rendering
// Responsibilities:
// - Parsing Claude Code .jsonl conversation files
// - Color-coding user/assistant/tool_use/system messages
// - Extracting readable summaries from tool_use blocks
// - Rendering thinking blocks as dimmed/collapsed text

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// jsonlMessage represents a single line from a Claude Code JSONL file.
// Only the fields needed for display are parsed.
type jsonlMessage struct {
	Type    string          `json:"type"`    // "user", "assistant", "system", "file-history-snapshot"
	Subtype string          `json:"subtype"` // e.g. "stop_hook_summary"
	Message json.RawMessage `json:"message"` // The API message object
	UUID    string          `json:"uuid"`
}

// jsonlAPIMessage is the inner message object with role and content.
type jsonlAPIMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"` // string or []contentBlock
}

// jsonlContentBlock represents a content block inside a message.
type jsonlContentBlock struct {
	Type  string          `json:"type"` // "text", "tool_use", "tool_result", "thinking"
	Text  string          `json:"text"`
	Name  string          `json:"name"`  // tool_use: tool name
	Input json.RawMessage `json:"input"` // tool_use: input parameters
}

// Styles for JSONL rendering
var (
	jsonlUserStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)
	jsonlAssistantStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("78"))
	jsonlToolNameStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")).
				Bold(true)
	jsonlToolInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243"))
	jsonlThinkingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Italic(true)
	jsonlSystemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245"))
	jsonlSeparatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("236"))
	jsonlToolResultStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243"))
)

// renderJSONLPreview renders a Claude Code .jsonl conversation file
// with color-coded messages. Returns formatted lines for the preview pane.
func renderJSONLLines(content []string, availableWidth int) []string {
	var lines []string

	for _, rawLine := range content {
		rawLine = strings.TrimSpace(rawLine)
		if rawLine == "" {
			continue
		}

		var msg jsonlMessage
		if err := json.Unmarshal([]byte(rawLine), &msg); err != nil {
			continue // skip unparseable lines
		}

		rendered := renderJSONLEntry(msg, availableWidth)
		if len(rendered) > 0 {
			lines = append(lines, rendered...)
		}
	}

	return lines
}

// renderJSONLEntry renders a single JSONL entry into display lines.
func renderJSONLEntry(msg jsonlMessage, width int) []string {
	switch msg.Type {
	case "user":
		return renderJSONLUserMessage(msg, width)
	case "assistant":
		return renderJSONLAssistantMessage(msg, width)
	case "system":
		return renderJSONLSystemMessage(msg, width)
	default:
		// Skip file-history-snapshot and other internal types
		return nil
	}
}

// renderJSONLUserMessage renders a user message (either text or tool results).
func renderJSONLUserMessage(msg jsonlMessage, width int) []string {
	if len(msg.Message) == 0 {
		return nil
	}

	var apiMsg jsonlAPIMessage
	if err := json.Unmarshal(msg.Message, &apiMsg); err != nil {
		return nil
	}

	var lines []string

	// Content can be a string or array of content blocks
	var contentStr string
	if err := json.Unmarshal(apiMsg.Content, &contentStr); err == nil {
		// Simple string content — this is user's actual message
		if contentStr == "" {
			return nil
		}
		sep := jsonlSeparatorStyle.Render(strings.Repeat("─", min(width, 40)))
		lines = append(lines, sep)
		header := jsonlUserStyle.Render("USER")
		lines = append(lines, header)

		for _, textLine := range wrapLine(contentStr, width) {
			lines = append(lines, jsonlUserStyle.Render(textLine))
		}
		lines = append(lines, "")
		return lines
	}

	// Array of content blocks — check for user text vs tool results
	var blocks []jsonlContentBlock
	if err := json.Unmarshal(apiMsg.Content, &blocks); err != nil {
		return nil
	}

	for _, block := range blocks {
		switch block.Type {
		case "text":
			if block.Text == "" {
				continue
			}
			sep := jsonlSeparatorStyle.Render(strings.Repeat("─", min(width, 40)))
			lines = append(lines, sep)
			header := jsonlUserStyle.Render("USER")
			lines = append(lines, header)
			for _, textLine := range wrapLine(block.Text, width) {
				lines = append(lines, jsonlUserStyle.Render(textLine))
			}
			lines = append(lines, "")

		case "tool_result":
			// Show tool result as a compact summary
			resultText := extractToolResultText(block)
			if resultText != "" {
				maxLen := width - 4
				if len(resultText) > maxLen {
					resultText = resultText[:maxLen] + "..."
				}
				lines = append(lines, jsonlToolResultStyle.Render("  "+resultText))
			}
		}
	}

	return lines
}

// renderJSONLAssistantMessage renders an assistant message with text, tool use, and thinking.
func renderJSONLAssistantMessage(msg jsonlMessage, width int) []string {
	if len(msg.Message) == 0 {
		return nil
	}

	var apiMsg jsonlAPIMessage
	if err := json.Unmarshal(msg.Message, &apiMsg); err != nil {
		return nil
	}

	var blocks []jsonlContentBlock
	if err := json.Unmarshal(apiMsg.Content, &blocks); err != nil {
		return nil
	}

	var lines []string
	hasContent := false

	for _, block := range blocks {
		switch block.Type {
		case "text":
			if block.Text == "" {
				continue
			}
			if !hasContent {
				header := jsonlAssistantStyle.Render("ASSISTANT")
				lines = append(lines, header)
				hasContent = true
			}
			for _, textLine := range wrapLine(block.Text, width) {
				lines = append(lines, jsonlAssistantStyle.Render(textLine))
			}
			lines = append(lines, "")

		case "tool_use":
			if !hasContent {
				header := jsonlAssistantStyle.Render("ASSISTANT")
				lines = append(lines, header)
				hasContent = true
			}
			toolLine := renderToolUseSummary(block, width)
			lines = append(lines, toolLine)

		case "thinking":
			if block.Text == "" {
				continue
			}
			// Show first line of thinking, truncated
			firstLine := strings.SplitN(block.Text, "\n", 2)[0]
			if visualWidth(firstLine) > width-12 {
				firstLine = truncateToWidth(firstLine, width-12)
			}
			lines = append(lines, jsonlThinkingStyle.Render("  thinking: "+firstLine))
		}
	}

	return lines
}

// renderJSONLSystemMessage renders system messages (hook summaries, etc).
func renderJSONLSystemMessage(msg jsonlMessage, width int) []string {
	if msg.Subtype == "" {
		return nil // Skip generic system messages
	}
	line := jsonlSystemStyle.Render(fmt.Sprintf("  [system: %s]", msg.Subtype))
	return []string{line}
}

// renderToolUseSummary creates a one-line summary of a tool_use block.
func renderToolUseSummary(block jsonlContentBlock, width int) string {
	name := block.Name
	if name == "" {
		name = "unknown"
	}

	// Extract key parameter for context
	detail := extractToolDetail(block)

	prefix := jsonlToolNameStyle.Render("  " + name)
	if detail != "" {
		maxDetail := width - visualWidth(name) - 5
		if maxDetail > 10 {
			if len(detail) > maxDetail {
				detail = detail[:maxDetail] + "..."
			}
			return prefix + " " + jsonlToolInputStyle.Render(detail)
		}
	}
	return prefix
}

// extractToolDetail extracts a readable summary from tool_use input parameters.
func extractToolDetail(block jsonlContentBlock) string {
	if len(block.Input) == 0 {
		return ""
	}

	var input map[string]interface{}
	if err := json.Unmarshal(block.Input, &input); err != nil {
		return ""
	}

	// Prioritize showing the most informative parameter based on tool name
	switch block.Name {
	case "Read":
		if fp, ok := input["file_path"].(string); ok {
			return fp
		}
	case "Write":
		if fp, ok := input["file_path"].(string); ok {
			return fp
		}
	case "Edit":
		if fp, ok := input["file_path"].(string); ok {
			return fp
		}
	case "Bash":
		if cmd, ok := input["command"].(string); ok {
			return cmd
		}
		if desc, ok := input["description"].(string); ok {
			return desc
		}
	case "Grep":
		if pat, ok := input["pattern"].(string); ok {
			detail := pat
			if path, ok := input["path"].(string); ok {
				detail += " in " + path
			}
			return detail
		}
	case "Glob":
		if pat, ok := input["pattern"].(string); ok {
			return pat
		}
	case "Agent":
		if desc, ok := input["description"].(string); ok {
			return desc
		}
		if prompt, ok := input["prompt"].(string); ok {
			if len(prompt) > 80 {
				prompt = prompt[:80]
			}
			return prompt
		}
	case "WebSearch", "WebFetch":
		if q, ok := input["query"].(string); ok {
			return q
		}
		if u, ok := input["url"].(string); ok {
			return u
		}
	}

	// Fallback: show first string value found
	for _, v := range input {
		if s, ok := v.(string); ok && len(s) > 0 && len(s) < 200 {
			return s
		}
	}
	return ""
}

// extractToolResultText extracts readable text from a tool_result content block.
func extractToolResultText(block jsonlContentBlock) string {
	// tool_result content can be a string or array
	if block.Text != "" {
		firstLine := strings.SplitN(block.Text, "\n", 2)[0]
		return firstLine
	}
	return ""
}

// isJSONLFile returns true if the file path looks like a Claude Code JSONL conversation file.
func isJSONLFile(path string) bool {
	return strings.HasSuffix(path, ".jsonl")
}

// loadJSONLPreview reads a JSONL file for preview, using tail-reading for large files.
// Parses JSON messages once (expensive) and caches them. Rendering happens lazily.
func (m *model) loadJSONLPreview(path string, fileSize int64) {
	const maxJSONLBytes = 512 * 1024 // 512KB tail for large files

	f, err := os.Open(path)
	if err != nil {
		m.preview.content = []string{fmt.Sprintf("Error reading file: %v", err)}
		m.preview.loaded = true
		return
	}
	defer f.Close()

	var data []byte
	isTailed := false

	if fileSize > maxJSONLBytes {
		offset := fileSize - maxJSONLBytes
		if _, err := f.Seek(offset, io.SeekStart); err == nil {
			data, err = io.ReadAll(f)
			if err != nil {
				m.preview.content = []string{fmt.Sprintf("Error reading file: %v", err)}
				m.preview.loaded = true
				return
			}
			// Skip the first partial line
			if idx := strings.IndexByte(string(data), '\n'); idx >= 0 {
				data = data[idx+1:]
			}
			isTailed = true
		}
	}

	if data == nil {
		data, err = io.ReadAll(f)
		if err != nil {
			m.preview.content = []string{fmt.Sprintf("Error reading file: %v", err)}
			m.preview.loaded = true
			return
		}
	}

	// Parse JSON messages once (expensive part)
	rawLines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	messages := make([]jsonlMessage, 0, len(rawLines))
	for _, line := range rawLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var msg jsonlMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	m.preview.isJSONL = true
	m.preview.cachedJSONLMessages = messages
	m.preview.cachedJSONLIsTailed = isTailed
	m.preview.fileSize = fileSize
	m.preview.loaded = true
}

// renderJSONLFromMessages renders parsed JSONL messages at the given width.
func renderJSONLFromMessages(messages []jsonlMessage, width int, isTailed bool, fileSize int64) []string {
	var rendered []string
	if isTailed {
		header := jsonlSystemStyle.Render(fmt.Sprintf("... (showing tail of %s file)",
			formatFileSize(fileSize)))
		rendered = append(rendered, header, "")
	}

	for _, msg := range messages {
		lines := renderJSONLEntry(msg, width)
		rendered = append(rendered, lines...)
	}
	return rendered
}

// renderJSONLPreview renders a JSONL conversation in the preview pane with
// scrolling, scrollbar, and color-coded messages.
func (m model) renderJSONLPreview(maxVisible int) string {
	var s strings.Builder

	var boxContentWidth int
	if m.viewMode == viewFullPreview {
		boxContentWidth = m.width - 6
	} else {
		boxContentWidth = m.rightWidth - 2
	}
	availableWidth := boxContentWidth - 2 // scrollbar + space
	if availableWidth < 20 {
		availableWidth = 20
	}

	// Render from cached parsed messages (JSON parsing already done)
	renderedLines := renderJSONLFromMessages(m.preview.cachedJSONLMessages, availableWidth, m.preview.cachedJSONLIsTailed, m.preview.fileSize)
	if len(renderedLines) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)
		s.WriteString(emptyStyle.Render("Empty conversation"))
		for i := 1; i < maxVisible; i++ {
			s.WriteString("\n\033[0m")
		}
		return s.String()
	}

	// Calculate visible range
	totalLines := len(renderedLines)
	start := m.preview.scrollPos
	if start < 0 {
		start = 0
	}

	targetLines := maxVisible
	if m.viewMode == viewDualPane && totalLines > 0 {
		targetLines = maxVisible - 1
	}

	if start >= totalLines {
		start = max(0, totalLines-targetLines)
	}

	end := start + targetLines
	if end > totalLines {
		end = totalLines
		start = max(0, end-targetLines)
	}

	// Render visible lines with scrollbar
	linesRendered := 0
	writeLine := func(line string) {
		if linesRendered > 0 {
			s.WriteString("\n")
		}
		s.WriteString(line)
		linesRendered++
	}

	for i := start; i < end; i++ {
		scrollbar := m.renderScrollbar(i-start, maxVisible, totalLines)
		renderedLine := scrollbar + " " + renderedLines[i] + "\033[0m"
		writeLine(renderedLine)
	}

	// Scroll indicator for dual-pane
	if m.viewMode == viewDualPane && totalLines > 0 {
		maxScrollPos := totalLines - targetLines
		var scrollPercent int
		if maxScrollPos <= 0 {
			scrollPercent = 100
		} else {
			scrollPercent = (m.preview.scrollPos * 100) / maxScrollPos
			if scrollPercent > 100 {
				scrollPercent = 100
			}
		}
		scrollIndicator := fmt.Sprintf(" %d/%d (%d%%) [jsonl]", end, totalLines, scrollPercent)
		scrollStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)

		for linesRendered < targetLines {
			writeLine("\033[0m")
		}
		if linesRendered > 0 {
			s.WriteString("\n")
		}
		s.WriteString(scrollStyle.Render(scrollIndicator))
		linesRendered++
	} else {
		for linesRendered < maxVisible {
			writeLine("\033[0m")
		}
	}

	return s.String()
}
