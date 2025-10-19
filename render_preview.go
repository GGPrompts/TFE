package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// getWrappedLineCount calculates the total number of wrapped lines for the current preview
func (m model) getWrappedLineCount() int {
	if !m.preview.loaded {
		return 0
	}

	// Calculate available width based on file type and view mode
	var availableWidth int
	var boxContentWidth int

	if m.viewMode == viewFullPreview {
		boxContentWidth = m.width - 6 // Box content width in full preview
	} else {
		boxContentWidth = m.rightWidth - 2 // Box content width in dual-pane
	}

	if m.preview.isMarkdown {
		// Markdown: no line numbers or scrollbar, content uses full box width
		availableWidth = boxContentWidth
	} else {
		// Regular text: subtract line nums (6) + scrollbar (1) + space (1) = 8 chars
		availableWidth = boxContentWidth - 8
	}

	if availableWidth < 20 {
		availableWidth = 20
	}

	// Use cached line count if available and width matches
	if m.preview.cacheValid && m.preview.cachedLineCount > 0 && m.preview.cachedWidth == availableWidth {
		return m.preview.cachedLineCount
	}

	// For markdown, we need to render it to count lines
	if m.preview.isMarkdown {
		markdownContent := strings.Join(m.preview.content, "\n")
		// Use cached rendering with timeout to prevent hangs
		// Note: renderMarkdownWithTimeout is in file_operations.go
		rendered, err := renderMarkdownWithTimeout(markdownContent, availableWidth, 5*time.Second)
		if err == nil {
			renderedLines := strings.Split(strings.TrimRight(rendered, "\n"), "\n")
			return len(renderedLines)
		}
		// Fallback if glamour fails or times out
	}

	// For regular text, count wrapped lines
	totalLines := 0
	for _, line := range m.preview.content {
		wrapped := wrapLine(line, availableWidth)
		totalLines += len(wrapped)
	}
	return totalLines
}

// wrapLine wraps a line of text to fit within the specified width
func wrapLine(line string, width int) []string {
	if width <= 0 {
		return []string{line}
	}

	// Handle empty lines
	if len(line) == 0 {
		return []string{""}
	}

	var wrapped []string
	currentLine := ""
	currentWidth := 0

	words := strings.Fields(line)
	if len(words) == 0 {
		// Line is only whitespace
		return []string{line}
	}

	for i, word := range words {
		wordWidth := visualWidth(word)
		spaceWidth := 1

		// Check if this word fits on the current line
		if currentWidth == 0 {
			// First word on line
			if wordWidth <= width {
				currentLine = word
				currentWidth = wordWidth
			} else {
				// Word is too long, force break it
				wrapped = append(wrapped, word[:width])
				currentLine = ""
				currentWidth = 0
			}
		} else if currentWidth+spaceWidth+wordWidth <= width {
			// Word fits on current line
			currentLine += " " + word
			currentWidth += spaceWidth + wordWidth
		} else {
			// Word doesn't fit, start new line
			wrapped = append(wrapped, currentLine)
			if wordWidth <= width {
				currentLine = word
				currentWidth = wordWidth
			} else {
				// Word is too long, force break it
				wrapped = append(wrapped, word[:width])
				currentLine = ""
				currentWidth = 0
			}
		}

		// If this is the last word, add the current line
		if i == len(words)-1 && currentLine != "" {
			wrapped = append(wrapped, currentLine)
		}
	}

	if len(wrapped) == 0 {
		return []string{line}
	}

	return wrapped
}

// renderPreview renders the preview pane content with scrollbar
func (m model) renderPreview(maxVisible int) string {
	var s strings.Builder

	if !m.preview.loaded {
		s.WriteString("No file loaded")
		return s.String()
	}

	// If this is a prompt file, show metadata header
	if m.preview.isPrompt && m.preview.promptTemplate != nil {
		return m.renderPromptPreview(maxVisible)
	}

	// Calculate available width for content based on file type and view mode
	var availableWidth int
	var boxContentWidth int // Width of the box content area

	if m.viewMode == viewFullPreview {
		boxContentWidth = m.width - 6 // Box content width in full preview
	} else {
		boxContentWidth = m.rightWidth - 2 // Box content width in dual-pane (accounting for borders)
	}

	if m.preview.isMarkdown {
		// Markdown: no line numbers or scrollbar, content uses full box width
		availableWidth = boxContentWidth
	} else {
		// Regular text: subtract line nums (6) + scrollbar (1) + space (1) = 8 chars
		availableWidth = boxContentWidth - 8
	}

	if availableWidth < 20 {
		availableWidth = 20 // Minimum width
	}

	// If markdown, render with Glamour
	if m.preview.isMarkdown && m.preview.cachedRenderedContent != "" {
		// Use cached Glamour-rendered content (no line numbers)
		if m.preview.cachedRenderedContent != "" {
			renderedLines := strings.Split(strings.TrimRight(m.preview.cachedRenderedContent, "\n"), "\n")

			// Calculate visible range based on scroll position
			totalLines := len(renderedLines)
			start := m.preview.scrollPos

			if start < 0 {
				start = 0
			}
			if start >= totalLines {
				start = max(0, totalLines-maxVisible)
			}

			end := start + maxVisible
			if end > totalLines {
				end = totalLines
			}

			// Render visible lines without line numbers for markdown
			for i := start; i < end; i++ {
				s.WriteString(renderedLines[i])
				s.WriteString("\033[0m") // Reset ANSI codes to prevent bleed
				s.WriteString("\n")
			}

			return strings.TrimRight(s.String(), "\n")
		}
		// If markdown flag is set but no rendered content, fall through to plain text rendering
		// This happens when Glamour fails or file is too large
	}

	// Wrap all lines first (use cache if available and width matches)
	var wrappedLines []string
	if m.preview.cacheValid && len(m.preview.cachedWrappedLines) > 0 && m.preview.cachedWidth == availableWidth {
		// Use cached wrapped lines
		wrappedLines = m.preview.cachedWrappedLines
	} else {
		// Wrap lines (will be slow for large files without cache)
		for _, line := range m.preview.content {
			wrapped := wrapLine(line, availableWidth)
			wrappedLines = append(wrappedLines, wrapped...)
		}
	}

	// Calculate visible range based on scroll position
	totalLines := len(wrappedLines)
	start := m.preview.scrollPos

	// Ensure start is within bounds
	if start < 0 {
		start = 0
	}
	if start >= totalLines {
		start = max(0, totalLines-maxVisible)
	}

	end := start + maxVisible
	if end > totalLines {
		end = totalLines
	}

	// Render lines with line numbers and scrollbar
	for i := start; i < end; i++ {
		// Line number (5 chars)
		lineNum := fmt.Sprintf("%5d ", i+1)
		lineNumStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		s.WriteString(lineNumStyle.Render(lineNum))

		// Scrollbar right after line number (replaces the │ separator)
		scrollbar := m.renderScrollbar(i-start, maxVisible, totalLines)
		s.WriteString(scrollbar)

		// Space after scrollbar
		s.WriteString(" ")

		// Content line
		s.WriteString(wrappedLines[i])
		s.WriteString("\033[0m") // Reset ANSI codes to prevent bleed
		s.WriteString("\n")
	}

	return strings.TrimRight(s.String(), "\n")
}

// highlightPromptVariables highlights {{variables}} in template text with assigned colors
func (m model) highlightPromptVariables(templateText string) string {
	if !m.inputFieldsActive || len(m.promptInputFields) == 0 {
		return templateText
	}

	// Build a map of variable names to colors
	colorMap := make(map[string]string)
	for _, field := range m.promptInputFields {
		colorMap[field.name] = field.color
	}

	// Auto-filled variables (DATE, TIME) should already be in fields with green color
	// But check template variables in case they exist but weren't added as fields
	if m.preview.promptTemplate != nil {
		for _, varName := range m.preview.promptTemplate.variables {
			// Only add if not already in colorMap (from fields)
			if _, exists := colorMap[varName]; !exists {
				varLower := strings.ToLower(varName)
				if varLower == "date" || varLower == "time" {
					colorMap[varName] = "34" // Green
				}
			}
		}
	}

	result := templateText

	// Replace each {{variable}} with colored version
	// Try all case variations for each variable
	for varName, color := range colorMap {
		variations := []string{
			varName,
			strings.ToUpper(varName),
			strings.ToLower(varName),
			strings.Title(strings.ToLower(varName)),
		}

		for _, variant := range variations {
			pattern := "{{" + variant + "}}"
			// ANSI color: \033[38;5;<color>m for foreground
			colored := fmt.Sprintf("\033[38;5;%sm{{%s}}\033[0m", color, variant)
			result = strings.ReplaceAll(result, pattern, colored)
		}
	}

	return result
}

// renderInputFields renders the input fields section below the preview
func (m model) renderInputFields(availableWidth, availableHeight int) string {
	if !m.inputFieldsActive || len(m.promptInputFields) == 0 {
		return ""
	}

	var s strings.Builder

	// Title for input fields section
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	s.WriteString(titleStyle.Render("📝 Fillable Fields"))
	s.WriteString("\n")

	// Help text
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Italic(true)
	s.WriteString(helpStyle.Render("Tab: Navigate • Type: Edit • F3: File Picker • F5: Copy • 🕐 Auto-filled"))
	s.WriteString("\n\n")

	// Calculate how many fields we can show (reserve 3 lines: title + help + blank)
	headerLines := 3
	linesPerField := 2 // Label line + input line
	maxFields := (availableHeight - headerLines) / linesPerField
	if maxFields < 1 {
		maxFields = 1
	}

	// Render each field
	fieldsShown := 0
	for i, field := range m.promptInputFields {
		if fieldsShown >= maxFields {
			// Show "... X more fields" message
			remainingCount := len(m.promptInputFields) - fieldsShown
			moreStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Italic(true)
			s.WriteString(moreStyle.Render(fmt.Sprintf("... %d more field(s) below", remainingCount)))
			break
		}

		// Field label (no indicator - it goes on the value line)
		labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(field.color))
		fieldTypeIndicator := ""

		// Check if this is an auto-filled field (DATE/TIME)
		fieldNameLower := strings.ToLower(field.name)
		isAutoFilled := fieldNameLower == "date" || fieldNameLower == "time"

		if isAutoFilled {
			fieldTypeIndicator = " 🕐" // Clock icon for auto-filled time/date
		} else {
			switch field.fieldType {
			case fieldTypeFile:
				fieldTypeIndicator = " 📁"
			case fieldTypeLong:
				fieldTypeIndicator = " 📝"
			}
		}

		s.WriteString("  ")
		s.WriteString(labelStyle.Render(field.name + fieldTypeIndicator + ":"))
		s.WriteString("\n")

		// Field value with focus indicator
		focusIndicator := "  "
		if i == m.focusedInputField {
			focusIndicator = "▶ "
		}

		displayValue := field.getDisplayValue()
		charCount := field.getCharCountDisplay()

		// Build the input display
		inputStyle := lipgloss.NewStyle()
		if i == m.focusedInputField {
			// Focused field - highlighted background
			inputStyle = inputStyle.Background(lipgloss.Color("235"))
		}

		// Show [...] prefix for truncated long content
		valueDisplay := displayValue
		if field.hasContent() && len(field.value) > len(displayValue) {
			// Truncated - add prefix
			valueDisplay = "[...]" + displayValue + charCount
		} else if charCount != "" {
			valueDisplay = displayValue + charCount
		}

		// Dim style if showing default (not user-entered)
		if !field.hasContent() {
			inputStyle = inputStyle.Foreground(lipgloss.Color("242"))
			valueDisplay = displayValue + " (default)"
		}

		s.WriteString(focusIndicator)
		s.WriteString(inputStyle.Render(valueDisplay))
		s.WriteString("\n")

		fieldsShown++
	}

	return s.String()
}

// renderPromptPreview renders a prompt file with metadata header
func (m model) renderPromptPreview(maxVisible int) string {
	var s strings.Builder
	tmpl := m.preview.promptTemplate

	// Build header lines
	var headerLines []string

	// Prompt name (if available)
	if tmpl.name != "" {
		nameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
		headerLines = append(headerLines, nameStyle.Render("📝 "+tmpl.name))
		headerLines = append(headerLines, "") // Blank line
	}

	// Description (if available)
	if tmpl.description != "" {
		descStyle := lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("245"))
		headerLines = append(headerLines, descStyle.Render(tmpl.description))
		headerLines = append(headerLines, "") // Blank line
	}

	// Source indicator
	sourceStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	sourceIcon := ""
	sourceLabel := ""
	switch tmpl.source {
	case "global":
		sourceIcon = "🌐"
		sourceLabel = "Global Prompt (~/.prompts/)"
	case "command":
		sourceIcon = "⚙️"
		sourceLabel = "Project Command (.claude/commands/)"
	case "agent":
		sourceIcon = "🤖"
		sourceLabel = "Project Agent (.claude/agents/)"
	case "skill":
		sourceIcon = "🎯"
		sourceLabel = "Project Skill (.claude/skills/)"
	case "local":
		sourceIcon = "📁"
		sourceLabel = "Local Prompt"
	}
	headerLines = append(headerLines, sourceStyle.Render(sourceIcon+" "+sourceLabel))

	// Variables detected (if any)
	if len(tmpl.variables) > 0 {
		varsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
		varsLine := fmt.Sprintf("Variables: %s", strings.Join(tmpl.variables, ", "))
		headerLines = append(headerLines, varsStyle.Render(varsLine))
	}

	// Separator line
	separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	headerLines = append(headerLines, separatorStyle.Render("─────────────────────────────────────"))
	headerLines = append(headerLines, "") // Blank line before content

	// Calculate how many lines the header takes
	headerHeight := len(headerLines)

	// Calculate available width
	var availableWidth int
	var boxContentWidth int

	if m.viewMode == viewFullPreview {
		boxContentWidth = m.width - 6
	} else {
		boxContentWidth = m.rightWidth - 2
	}

	// Prompts don't show line numbers, so use full width
	availableWidth = boxContentWidth - 2 // Just padding

	if availableWidth < 20 {
		availableWidth = 20
	}

	// Determine content to display based on input fields state
	var contentLines []string
	var isMarkdownPrompt bool

	if m.inputFieldsActive {
		// Show template with highlighted variables (no Glamour - colors would conflict)
		highlightedTemplate := m.highlightPromptVariables(tmpl.template)
		contentLines = strings.Split(highlightedTemplate, "\n")
		isMarkdownPrompt = false // Don't apply Glamour when showing variable highlights
	} else {
		// Show substituted content - apply Glamour if markdown
		contentText := strings.Join(m.preview.content, "\n")

		// Check if this is a markdown file
		isMarkdownPrompt = m.preview.isMarkdown

		if isMarkdownPrompt {
			// Render with Glamour for beautiful formatting (with timeout to prevent hangs)
			rendered, err := renderMarkdownWithTimeout(contentText, availableWidth, 5*time.Second)
			if err == nil {
				// Successfully rendered with Glamour
				contentLines = strings.Split(strings.TrimRight(rendered, "\n"), "\n")
			} else {
				// Glamour failed or timed out, fall back to plain text
				contentLines = m.preview.content
				isMarkdownPrompt = false
			}
		} else {
			// Not markdown, use plain text
			contentLines = m.preview.content
		}
	}

	// Wrap content lines (only if not markdown - Glamour already wraps)
	var wrappedLines []string
	if isMarkdownPrompt {
		// Glamour already wrapped the text, use as-is
		wrappedLines = contentLines
	} else {
		// Plain text - wrap manually
		for _, line := range contentLines {
			wrapped := wrapLine(line, availableWidth)
			wrappedLines = append(wrappedLines, wrapped...)
		}
	}

	// Calculate available height for content and input fields
	var contentHeight int
	var inputFieldsSection string

	if m.inputFieldsActive {
		// Reserve space for input fields (approximately 1/3 of available space)
		inputFieldsHeight := maxVisible / 3
		if inputFieldsHeight < 8 {
			inputFieldsHeight = 8 // Minimum for at least 2 fields
		}
		contentHeight = maxVisible - headerHeight - inputFieldsHeight
		if contentHeight < 5 {
			contentHeight = 5
		}

		// Render input fields section
		inputFieldsSection = m.renderInputFields(availableWidth, inputFieldsHeight)
	} else {
		// No input fields - use all available space for content
		contentHeight = maxVisible - headerHeight
		if contentHeight < 5 {
			contentHeight = 5
		}
	}

	// Calculate visible range for content
	totalLines := len(wrappedLines)
	start := m.preview.scrollPos

	if start < 0 {
		start = 0
	}
	if start >= totalLines {
		start = max(0, totalLines-contentHeight)
	}

	end := start + contentHeight
	if end > totalLines {
		end = totalLines
	}

	// Render header
	for _, line := range headerLines {
		s.WriteString(line)
		s.WriteString("\033[0m") // Reset ANSI codes
		s.WriteString("\n")
	}

	// Render content (no line numbers for prompts)
	for i := start; i < end; i++ {
		s.WriteString(wrappedLines[i])
		s.WriteString("\033[0m") // Reset ANSI codes
		s.WriteString("\n")
	}

	// Render input fields section if active
	if m.inputFieldsActive && inputFieldsSection != "" {
		s.WriteString("\n")
		separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		s.WriteString(separatorStyle.Render("─────────────────────────────────────"))
		s.WriteString("\n")
		s.WriteString(inputFieldsSection)
	}

	return strings.TrimRight(s.String(), "\n")
}

// renderScrollbar renders a scrollbar indicator for the current line
// Now renders in place of the separator between line numbers and content
func (m model) renderScrollbar(lineIndex, visibleLines, totalLines int) string {
	// Calculate scrollbar position
	// The scrollbar thumb should represent the visible portion of the content
	scrollbarHeight := visibleLines
	thumbSize := max(1, (visibleLines*scrollbarHeight)/totalLines)
	thumbStart := (m.preview.scrollPos * scrollbarHeight) / totalLines

	scrollbarStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	scrollbarThumbStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))

	// Determine what to render for this line
	if lineIndex >= thumbStart && lineIndex < thumbStart+thumbSize {
		// This line is part of the scrollbar thumb (bright blue)
		return scrollbarThumbStyle.Render("│")
	} else {
		// This line is part of the scrollbar track (dim gray)
		return scrollbarStyle.Render("│")
	}
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// renderFullPreview renders the full-screen preview mode
func (m model) renderFullPreview() string {
	var s strings.Builder

	// Title bar with file name
	previewTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"}).
		Background(lipgloss.AdaptiveColor{Light: "#0087d7", Dark: "#00d7ff"}).
		Width(m.width).
		Padding(0, 1)

	titleText := fmt.Sprintf("Preview: %s", m.preview.fileName)
	if m.preview.tooLarge || m.preview.isBinary {
		titleText += " [Cannot Preview]"
	}
	if m.preview.isPrompt {
		titleText += " [Prompt Template]"
	} else if m.preview.isMarkdown {
		titleText += " [Markdown]"
	}
	s.WriteString(previewTitleStyle.Render(titleText))
	s.WriteString("\033[0m") // Reset ANSI codes
	s.WriteString("\n")

	// File info line - update based on whether we're showing markdown or wrapped text
	var infoText string
	if m.preview.isMarkdown {
		infoText = fmt.Sprintf("Size: %s | Markdown Rendered",
			formatFileSize(m.preview.fileSize))
	} else {
		infoText = fmt.Sprintf("Size: %s | Lines: %d (wrapped) | Scroll: %d",
			formatFileSize(m.preview.fileSize),
			len(m.preview.content),
			m.preview.scrollPos+1)
	}
	s.WriteString(pathStyle.Render(infoText))
	s.WriteString("\033[0m") // Reset ANSI codes
	s.WriteString("\n")

	// Content with border
	maxVisible := m.height - 6 // Reserve space for title, info, help (total box height INCLUDING borders)
	contentHeight := maxVisible - 2 // Content area accounting for borders
	previewContent := m.renderPreview(contentHeight)

	// Wrap preview in bordered box with fixed dimensions
	// Content is constrained to contentHeight lines to fit within the box
	// When mouse is disabled (for text selection), remove border for cleaner copying
	previewBoxStyle := lipgloss.NewStyle().
		Width(m.width - 6).       // Leave margin for borders
		Height(contentHeight)     // Content area height (borders added by Lipgloss)

	if m.previewMouseEnabled {
		// Mouse enabled: show decorative border
		previewBoxStyle = previewBoxStyle.
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.AdaptiveColor{
				Light: "#00af87", // Teal for light
				Dark:  "#5faf87",  // Light teal for dark
			})
	} else {
		// Mouse disabled (text selection mode): no border for cleaner copying
		previewBoxStyle = previewBoxStyle.Padding(0, 1) // Just add side padding
	}

	s.WriteString(previewBoxStyle.Render(previewContent))

	// Help text
	s.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).PaddingLeft(2)

	// Show different F5 text if viewing a prompt (with or without fillable fields)
	f5Text := "copy path"
	if m.preview.isPrompt || (m.inputFieldsActive && len(m.promptInputFields) > 0) {
		f5Text = "copy rendered prompt"
	}

	// Mouse toggle indicator and help text
	var mouseStatus, helpText string
	if m.previewMouseEnabled {
		mouseStatus = "ON"
		helpText = fmt.Sprintf("↑/↓: scroll • m: toggle border/mouse (🖱️  %s) • F4: edit • F5: %s • Esc: close", mouseStatus, f5Text)
	} else {
		mouseStatus = "OFF"
		helpText = fmt.Sprintf("↑/↓: scroll • m: toggle border/mouse (⌨️  %s) • F4: edit • F5: %s • Esc: close", mouseStatus, f5Text)
	}
	s.WriteString(helpStyle.Render(helpText))
	s.WriteString("\033[0m") // Reset ANSI codes

	return s.String()
}

// renderDualPane renders the split-pane layout using Lipgloss layout utilities
func (m model) renderDualPane() string {
	var s strings.Builder

	// Title with mode indicator and GitHub link
	titleText := "(T)erminal (F)ile (E)xplorer [Dual-Pane]"
	if m.commandFocused {
		titleText += " [Command Mode]"
	}

	// Create GitHub link (OSC 8 hyperlink format)
	githubURL := "https://github.com/GGPrompts/TFE"
	githubLink := fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", githubURL, githubURL)

	// Calculate spacing to right-align GitHub link
	githubText := githubURL // Display text
	availableWidth := m.width - len(titleText) - len(githubText) - 2
	if availableWidth < 1 {
		availableWidth = 1
	}
	spacing := strings.Repeat(" ", availableWidth)

	// Render title on left, GitHub link on right
	title := titleStyle.Render(titleText) + spacing + titleStyle.Render(githubLink)
	s.WriteString(title)
	s.WriteString("\033[0m") // Reset ANSI codes
	s.WriteString("\n")

	// Toolbar buttons
	// Home button - highlight with gray background when in home directory
	homeDir, _ := os.UserHomeDir()
	if homeDir != "" && m.currentPath == homeDir {
		// Active: gray background (in home directory)
		homeButtonStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true).
			Background(lipgloss.Color("237"))
		s.WriteString(homeButtonStyle.Render("[🏠]"))
	} else {
		// Inactive: normal styling
		homeButtonStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)
		s.WriteString(homeButtonStyle.Render("[🏠]"))
	}
	s.WriteString(" ")

	// Favorites filter toggle button
	starIcon := "⭐"
	if m.showFavoritesOnly {
		starIcon = "✨" // Different icon when filter is active
	}
	favButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)
	s.WriteString(favButtonStyle.Render("[" + starIcon + "]"))
	s.WriteString(" ")

	// Command mode toggle button with green >_ and blue brackets
	if m.commandFocused {
		// Active: gray background
		bracketStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true).Background(lipgloss.Color("237"))
		termStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true).Background(lipgloss.Color("237"))
		s.WriteString(bracketStyle.Render("["))
		s.WriteString(termStyle.Render(">_"))
		s.WriteString(bracketStyle.Render("]"))
	} else {
		// Inactive: normal styling
		bracketStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
		termStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
		s.WriteString(bracketStyle.Render("["))
		s.WriteString(termStyle.Render(">_"))
		s.WriteString(bracketStyle.Render("]"))
	}
	s.WriteString(" ")

	// Fuzzy search button
	searchButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)
	s.WriteString(searchButtonStyle.Render("[🔍]"))
	s.WriteString(" ")

	// Prompts filter toggle button
	if m.showPromptsOnly {
		// Active: gray background (like command mode)
		activeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true).
			Background(lipgloss.Color("237"))
		s.WriteString(activeStyle.Render("[📝]"))
	} else {
		// Inactive: normal styling
		promptButtonStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)
		s.WriteString(promptButtonStyle.Render("[📝]"))
	}
	s.WriteString(" ")

	// Trash/Recycle bin button
	trashIcon := "🗑️"
	if m.showTrashOnly {
		trashIcon = "♻️" // Recycle icon when viewing trash
	}
	trashButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)
	s.WriteString(trashButtonStyle.Render("[" + trashIcon + "]"))

	s.WriteString("\033[0m") // Reset ANSI codes
	s.WriteString("\n")

	// Command prompt with path (terminal-style)
	promptPrefix := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true).Render("$ ")
	pathPromptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	s.WriteString(promptPrefix)
	s.WriteString(pathPromptStyle.Render(getDisplayPath(m.currentPath)))
	s.WriteString(" ")

	// Show helper text when not focused and empty, otherwise show input
	if !m.commandFocused && m.commandInput == "" {
		helperStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
		s.WriteString(helperStyle.Render(": to focus"))
	} else {
		s.WriteString(inputStyle.Render(m.commandInput))
	}

	// Show cursor only when command mode is active
	if m.commandFocused {
		cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
		s.WriteString(cursorStyle.Render("█"))
	}
	// Explicitly reset styling after cursor to prevent ANSI code leakage
	s.WriteString("\033[0m")
	s.WriteString("\n")

	// Blank line separator between command prompt and panes
	s.WriteString("\n")

	// Calculate max visible for both panes
	// title=1 + toolbar=1 + command=1 + separator=1 + panes=maxVisible + separator=1 + status=2 = m.height
	// Therefore: maxVisible = m.height - 7 (total pane height INCLUDING borders)
	maxVisible := m.height - 7

	// Content area is maxVisible - 2 (accounting for top/bottom borders)
	contentHeight := maxVisible - 2

	// Get left pane content - use contentHeight so content fits within the box
	var leftContent string
	switch m.displayMode {
	case modeList:
		leftContent = m.renderListView(contentHeight)
	case modeDetail:
		leftContent = m.renderDetailView(contentHeight)
	case modeTree:
		leftContent = m.renderTreeView(contentHeight)
	default:
		leftContent = m.renderDetailView(contentHeight) // Default to detail view
	}

	// Get right pane content (just the preview, no title)
	rightContent := ""
	if m.preview.loaded {
		rightContent = m.renderPreview(contentHeight)
	} else {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)
		rightContent = emptyStyle.Render("No preview available\n\nSelect a file to preview") + "\033[0m"
	}

	// Create styled boxes for left and right panes using Lipgloss
	// Highlight the focused pane with a brighter border color
	leftBorderColor := lipgloss.AdaptiveColor{Light: "#999999", Dark: "#585858"}  // dim gray
	rightBorderColor := lipgloss.AdaptiveColor{Light: "#999999", Dark: "#585858"} // dim gray
	if m.focusedPane == leftPane {
		leftBorderColor = lipgloss.AdaptiveColor{Light: "#0087d7", Dark: "#00d7ff"} // bright blue for focused pane
	} else {
		rightBorderColor = lipgloss.AdaptiveColor{Light: "#0087d7", Dark: "#00d7ff"} // bright blue for focused pane
	}

	// Use fixed Width/Height for consistent borders
	// Content is constrained to contentHeight lines to fit within the box
	leftPaneStyle := lipgloss.NewStyle().
		Width(m.leftWidth - 2).   // -2 for left/right borders
		Height(contentHeight).    // Content area height (borders added by Lipgloss)
		Border(lipgloss.RoundedBorder()).
		BorderForeground(leftBorderColor)

	rightPaneStyle := lipgloss.NewStyle().
		Width(m.rightWidth - 2).  // -2 for left/right borders
		Height(contentHeight).    // Content area height (borders added by Lipgloss)
		Border(lipgloss.RoundedBorder()).
		BorderForeground(rightBorderColor)

	// Apply styles to content
	leftPaneRendered := leftPaneStyle.Render(leftContent)
	rightPaneRendered := rightPaneStyle.Render(rightContent)

	// Join panes horizontally
	panes := lipgloss.JoinHorizontal(lipgloss.Top, leftPaneRendered, rightPaneRendered)
	s.WriteString(panes)
	s.WriteString("\n")

	// Status bar (full width)
	// File counts
	dirCount, fileCount := 0, 0
	for _, f := range m.files {
		if f.name == ".." {
			continue
		}
		if f.isDir {
			dirCount++
		} else {
			fileCount++
		}
	}

	itemsInfo := fmt.Sprintf("%d folders, %d files", dirCount, fileCount)
	hiddenIndicator := ""
	if m.showHidden {
		hiddenIndicator = " • hidden"
	}

	favoritesIndicator := ""
	if m.showFavoritesOnly {
		favoritesIndicator = " • ⭐ favorites only"
	}

	promptsIndicator := ""
	if m.showPromptsOnly {
		promptsIndicator = " • 📝 prompts only"
	}

	// Show focused pane info in status bar
	focusInfo := ""
	if m.focusedPane == leftPane {
		focusInfo = " • [LEFT focused]"
	} else {
		focusInfo = " • [RIGHT focused]"
	}

	// Help hint
	helpHint := " • F1: help"

	// Selected file info
	var selectedInfo string
	if currentFile := m.getCurrentFile(); currentFile != nil {
		if currentFile.isDir {
			// Special handling for ".." to show parent directory name
			if currentFile.name == ".." {
				parentPath := filepath.Dir(m.currentPath)
				parentName := filepath.Base(parentPath)
				if parentName == "/" || parentName == "." {
					parentName = "root"
				}
				selectedInfo = fmt.Sprintf("Selected: .. (go up to %s)", parentName)
			} else {
				selectedInfo = fmt.Sprintf("Selected: %s (folder)", currentFile.name)
			}
		} else {
			fileType := getFileType(*currentFile)
			selectedInfo = fmt.Sprintf("Selected: %s (%s, %s, %s)",
				currentFile.name,
				fileType,
				formatFileSize(currentFile.size),
				formatModTime(currentFile.modTime))
		}
	}

	// Split status into two lines to prevent truncation
	// Line 1: Counts, indicators, view mode, focus, help
	statusLine1 := fmt.Sprintf("%s%s%s%s • %s%s%s", itemsInfo, hiddenIndicator, favoritesIndicator, promptsIndicator, m.displayMode.String(), focusInfo, helpHint)
	s.WriteString(statusStyle.Render(statusLine1))
	s.WriteString("\033[0m") // Reset ANSI codes
	s.WriteString("\n")

	// Line 2: Selected file info
	statusLine2 := selectedInfo
	s.WriteString(statusStyle.Render(statusLine2))
	s.WriteString("\033[0m") // Reset ANSI codes

	return s.String()
}
