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
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
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

// The jsonl*Style package-level vars are declared in styles.go and assigned in
// initStyles() (theme.go), so theme reloads refresh them without constructing
// a new lipgloss.Style per rendered message.

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
				if visualWidth(resultText) > maxLen {
					// truncateToWidth appends "..." when there is room,
					// keeping the total within maxLen columns
					resultText = truncateToWidth(resultText, maxLen)
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
			toolLines := renderToolUseSummary(block, width)
			lines = append(lines, toolLines...)

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

// renderToolUseSummary creates a summary of a tool_use block.
// Allows detail text to wrap to multiple lines for readability.
func renderToolUseSummary(block jsonlContentBlock, width int) []string {
	name := block.Name
	if name == "" {
		name = "unknown"
	}

	// Extract key parameter for context
	detail := extractToolDetail(block)

	prefix := jsonlToolNameStyle.Render("  " + name)
	if detail == "" {
		return []string{prefix}
	}

	// Allow detail to use full width, wrapping naturally
	detailIndent := "    " // Indent continuation lines under tool name
	maxDetail := width - visualWidth(detailIndent)
	if maxDetail < 20 {
		maxDetail = 20
	}

	// First line: tool name + start of detail
	firstLineMax := width - visualWidth(name) - 5
	if firstLineMax < 10 {
		firstLineMax = 10
	}

	if visualWidth(detail) <= firstLineMax {
		return []string{prefix + " " + jsonlToolInputStyle.Render(detail)}
	}

	// Detail wraps: first chunk on name line, rest indented.
	// Chunk by visual width (never mid-rune): byte-offset slicing splits
	// multibyte runes (CJK, emoji) into � pairs across consecutive lines.
	chunks := chunkByVisualWidth(detail, firstLineMax, maxDetail)
	var lines []string
	lines = append(lines, prefix+" "+jsonlToolInputStyle.Render(chunks[0]))
	for _, chunk := range chunks[1:] {
		lines = append(lines, jsonlToolInputStyle.Render(detailIndent+chunk))
	}
	return lines
}

// chunkByVisualWidth hard-splits s into chunks at visual-width boundaries:
// the first chunk is at most firstWidth columns, all following chunks at most
// restWidth columns. Splits between runes only — never inside a multibyte
// rune — so arbitrary UTF-8 (CJK, emoji) renders cleanly on every line.
func chunkByVisualWidth(s string, firstWidth, restWidth int) []string {
	if firstWidth < 1 {
		firstWidth = 1
	}
	if restWidth < 1 {
		restWidth = 1
	}

	var chunks []string
	var b strings.Builder
	colWidth := 0
	limit := firstWidth
	for _, r := range s {
		rw := runewidth.RuneWidth(r)
		if colWidth+rw > limit && b.Len() > 0 {
			chunks = append(chunks, b.String())
			b.Reset()
			colWidth = 0
			limit = restWidth
		}
		b.WriteRune(r)
		colWidth += rw
	}
	if b.Len() > 0 {
		chunks = append(chunks, b.String())
	}
	if len(chunks) == 0 {
		chunks = []string{""}
	}
	return chunks
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
			if visualWidth(prompt) > 80 {
				prompt = truncateToWidth(prompt, 80)
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

	// Fallback: show all short string values, sorted by key for stable output
	// (Go map iteration order is randomized, which causes flickering on re-render)
	keys := make([]string, 0, len(input))
	for k := range input {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		if s, ok := input[k].(string); ok && len(s) > 0 && len(s) < 200 {
			parts = append(parts, s)
		}
	}
	if len(parts) > 0 {
		return strings.Join(parts, " ")
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

	if fileSize > maxJSONLBytes && !m.preview.jsonlFullLoad {
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

	// Pre-render the conversation at the current width so the View path
	// (renderJSONLPreview / getWrappedLineCount) can read cached lines instead
	// of re-rendering the whole transcript every frame
	m.populateJSONLRenderCache()
}

// jsonlPreviewWidth returns the text width JSONL conversation lines render at:
// the preview box content width minus scrollbar (1) + space (1). Derived from
// previewBoxContentWidth() — never compute this ad hoc, or the rendered-line
// cache silently misses and every frame re-renders the full transcript.
func (m model) jsonlPreviewWidth() int {
	availableWidth := m.previewBoxContentWidth() - 2
	if availableWidth < 20 {
		availableWidth = 20
	}
	return availableWidth
}

// populateJSONLRenderCache renders all parsed JSONL messages at the current
// preview width and caches the resulting styled lines. Called where the model
// is mutable: loadJSONLPreview (including the F-key full-transcript reload),
// populatePreviewCache (window resize, pane toggles), and
// refreshPreviewCacheIfStale (post-dispatch width changes).
func (m *model) populateJSONLRenderCache() {
	width := m.jsonlPreviewWidth()
	m.preview.cachedJSONLRenderedLines = renderJSONLFromMessages(
		m.preview.cachedJSONLMessages, width, m.preview.cachedJSONLIsTailed, m.preview.fileSize)
	m.preview.cachedJSONLRenderedWidth = width
}

// jsonlRenderedLines returns the rendered conversation lines for the current
// width, preferring the cache. The fallback render only triggers when a View
// happens before Update refreshed the cache (value receiver — cannot store
// the result back).
func (m model) jsonlRenderedLines() []string {
	width := m.jsonlPreviewWidth()
	if m.preview.cachedJSONLRenderedWidth == width {
		return m.preview.cachedJSONLRenderedLines
	}
	return renderJSONLFromMessages(
		m.preview.cachedJSONLMessages, width, m.preview.cachedJSONLIsTailed, m.preview.fileSize)
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

	// Read pre-rendered lines from the cache (falls back to a fresh render
	// only if the cache width doesn't match the current layout)
	renderedLines := m.jsonlRenderedLines()
	if len(renderedLines) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(uiSubtleText()).
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
		suffix := " [jsonl]"
		if m.preview.cachedJSONLIsTailed {
			suffix += " | F: load full"
		}
		scrollIndicator := m.renderScrollIndicator(end, totalLines, targetLines, suffix)

		for linesRendered < targetLines {
			writeLine("\033[0m")
		}
		if linesRendered > 0 {
			s.WriteString("\n")
		}
		s.WriteString(scrollIndicator)
		linesRendered++
	} else {
		for linesRendered < maxVisible {
			writeLine("\033[0m")
		}
	}

	return s.String()
}
