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
		// Markdown: no line numbers or scrollbar, but add left padding for readability
		// Subtract 2 for left padding (prevents code blocks from touching border)
		availableWidth = boxContentWidth - 2
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
		rendered, err := m.renderMarkdownWithTimeout(markdownContent, availableWidth, 5*time.Second)
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
				// Word is too long, force break it using visual width
				wrapped = append(wrapped, truncateToWidth(word, width))
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
				// Word is too long, force break it using visual width
				wrapped = append(wrapped, truncateToWidth(word, width))
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
		// Markdown: no line numbers or scrollbar, but add left padding for readability
		// Subtract 2 for left padding (prevents code blocks from touching border)
		availableWidth = boxContentWidth - 2
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

			// Reserve last line for scroll indicator in dual-pane mode
			targetLines := maxVisible
			if m.viewMode == viewDualPane && totalLines > 0 {
				targetLines = maxVisible - 1
			}

			if start >= totalLines {
				start = max(0, totalLines-targetLines)
			}

			end := start + targetLines
			if end > totalLines {
				// When end exceeds total lines, adjust start backwards to show a full page
				end = totalLines
				start = max(0, end-targetLines)
			}

			// Render visible lines without line numbers for markdown
			outputLines := 0

			writeLine := func(line string) {
				if outputLines > 0 {
					s.WriteString("\n")
				}
				s.WriteString(line)
				outputLines++
			}

			for i := start; i < end; i++ {
				line := renderedLines[i]

				// Add scrollbar indicator for markdown files (since they don't have line numbers)
				scrollbar := m.renderScrollbar(outputLines, targetLines, totalLines)

				// Add scrollbar + space + left padding for better readability
				paddedLine := scrollbar + " " + line
				// Truncate after padding to prevent exceeding box bounds (per CLAUDE.md guidance)
				// IMPORTANT: Use terminal-aware functions for emoji width compensation
				if m.visualWidthCompensated(paddedLine) > boxContentWidth {
					paddedLine = m.truncateToWidthCompensated(paddedLine, boxContentWidth)
				}
				writeLine(paddedLine + "\033[0m")
			}

			// Add scroll position indicator in dual-pane mode
			if m.viewMode == viewDualPane && totalLines > 0 {
				// Calculate scroll percentage based on how far through scrollable content we are
				// maxScrollPos is the last valid scroll position that shows the bottom of content
				maxScrollPos := totalLines - targetLines
				var scrollPercent int
				if maxScrollPos <= 0 {
					// Content fits in one screen
					scrollPercent = 100
				} else {
					scrollPercent = (m.preview.scrollPos * 100) / maxScrollPos
					if scrollPercent > 100 {
						scrollPercent = 100
					}
				}

				// Show the last visible line number (not the top line)
				// end is already correctly clamped, so use it directly
				lastVisibleLine := end
				scrollIndicator := fmt.Sprintf(" %d/%d (%d%%) ", lastVisibleLine, totalLines, scrollPercent)
				scrollStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("241")).
					Italic(true)

				// Pad with empty lines to reach target
				for outputLines < targetLines {
					writeLine("\033[0m")
				}

				// Add scroll indicator on last line
				if outputLines > 0 {
					s.WriteString("\n")
				}
				s.WriteString(scrollStyle.Render(scrollIndicator))
				outputLines++
			} else {
				// Pad with empty lines to reach exactly maxVisible lines
				// This ensures proper alignment with the file list pane
				for outputLines < maxVisible {
					writeLine("\033[0m")
				}
			}

			return s.String()
		}
		// If markdown flag is set but no rendered content, fall through to plain text rendering
		// This happens when Glamour fails or file is too large
	}

	// Wrap all lines first (use cache if available and width matches)
	// IMPORTANT: Skip wrapping for graphics protocol data (Kitty/iTerm2/Sixel)
	// These escape sequences must stay intact on their original lines
	var wrappedLines []string
	if m.preview.hasGraphicsProtocol {
		// Don't wrap graphics protocol data - use content as-is
		wrappedLines = m.preview.content
	} else if m.preview.cacheValid && len(m.preview.cachedWrappedLines) > 0 && m.preview.cachedWidth == availableWidth {
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

	// Reserve last line for scroll indicator in dual-pane mode
	targetLines := maxVisible
	if m.viewMode == viewDualPane && totalLines > 0 {
		targetLines = maxVisible - 1
	}

	if start >= totalLines {
		start = max(0, totalLines-targetLines)
	}

	end := start + targetLines
	if end > totalLines {
		// When end exceeds total lines, adjust start backwards to show a full page
		// This ensures we always show targetLines when possible
		end = totalLines
		start = max(0, end-targetLines)
	}

	// Render lines with line numbers and scrollbar
	linesRendered := 0
	writeLine := func(line string) {
		if linesRendered > 0 {
			s.WriteString("\n")
		}
		s.WriteString(line)
		linesRendered++
	}

	for i := start; i < end; i++ {
		// Line number (5 chars)
		lineNum := fmt.Sprintf("%5d ", i+1)
		lineNumStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		renderedLine := lineNumStyle.Render(lineNum)

		// Scrollbar right after line number (replaces the │ separator)
		scrollbar := m.renderScrollbar(i-start, maxVisible, totalLines)
		renderedLine += scrollbar

		// Space after scrollbar
		renderedLine += " "

		// Content line - ensure it doesn't exceed available width to prevent wrapping
		contentLine := wrappedLines[i]

		// IMPORTANT: Don't truncate graphics protocol data - it contains escape sequences
		// that must remain intact. Only truncate regular text content.
		if !m.preview.hasGraphicsProtocol {
			// Truncate to available width using ANSI-aware truncation
			// This prevents long lines with ANSI codes from wrapping outside the box
			if visualWidth(contentLine) > availableWidth {
				contentLine = truncateToWidth(contentLine, availableWidth)
			}
		}

		renderedLine += contentLine
		renderedLine += "\033[0m" // Reset ANSI codes to prevent bleed
		writeLine(renderedLine)
	}

	// Add scroll position indicator as the last line in dual-pane mode
	if m.viewMode == viewDualPane && totalLines > 0 {
		// Calculate scroll percentage based on how far through scrollable content we are
		maxScrollPos := totalLines - targetLines
		var scrollPercent int
		if maxScrollPos <= 0 {
			// Content fits in one screen
			scrollPercent = 100
		} else {
			scrollPercent = (m.preview.scrollPos * 100) / maxScrollPos
			if scrollPercent > 100 {
				scrollPercent = 100
			}
		}

		// Show the last visible line number (not the top line)
		// end is already correctly clamped, so use it directly
		lastVisibleLine := end
		scrollIndicator := fmt.Sprintf(" %d/%d (%d%%) ", lastVisibleLine, totalLines, scrollPercent)
		scrollStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)

		// Pad with empty lines to reach target
		for linesRendered < targetLines {
			writeLine("\033[0m")
		}

		// Add scroll indicator on last line
		if linesRendered > 0 {
			s.WriteString("\n")
		}
		s.WriteString(scrollStyle.Render(scrollIndicator))
		linesRendered++
	} else {
		// Pad with empty lines to reach exactly maxVisible lines
		// This ensures proper alignment with the file list pane in dual-pane mode
		for linesRendered < maxVisible {
			writeLine("\033[0m")
		}
	}

	return s.String()
}

// renderInlineVariables processes template text and renders variables inline
// In EDIT MODE: Shows variables without {{}} brackets, with values inline
// - Focused variable: highlighted background (235) + yellow foreground (220)
// - Filled variables: shown in blue (39), just the value
// - Unfilled variables: shown in dim gray (242), just the name
func (m model) renderInlineVariables(templateText string) string {
	if m.preview.promptTemplate == nil || len(m.preview.promptTemplate.variables) == 0 {
		return templateText
	}

	result := templateText

	// Process each variable in the template
	for i, varName := range m.preview.promptTemplate.variables {
		var replacement string
		isFocused := (i == m.focusedVariableIndex)

		// Check if this variable has a filled value
		filledValue, hasFilled := m.filledVariables[varName]

		if isFocused {
			// Focused variable - show with highlight (background 235, foreground 220)
			// NO brackets - just show the value or variable name
			if hasFilled && filledValue != "" {
				// Show the filled value with focus highlight
				// For multi-line content, apply highlighting to each line separately
				if strings.Contains(filledValue, "\n") {
					lines := strings.Split(filledValue, "\n")
					highlightedLines := make([]string, len(lines))
					for i, line := range lines {
						highlightedLines[i] = fmt.Sprintf("\033[48;5;235m\033[38;5;220m%s\033[0m", line)
					}
					replacement = strings.Join(highlightedLines, "\n")
				} else {
					replacement = fmt.Sprintf("\033[48;5;235m\033[38;5;220m%s\033[0m", filledValue)
				}
			} else {
				// Show variable name (no brackets) with focus highlight
				replacement = fmt.Sprintf("\033[48;5;235m\033[38;5;220m%s\033[0m", varName)
			}
		} else if hasFilled && filledValue != "" {
			// Filled but not focused - show value in blue (39), NO brackets
			// For multi-line content, apply blue color to each line separately
			if strings.Contains(filledValue, "\n") {
				lines := strings.Split(filledValue, "\n")
				highlightedLines := make([]string, len(lines))
				for i, line := range lines {
					highlightedLines[i] = fmt.Sprintf("\033[38;5;39m%s\033[0m", line)
				}
				replacement = strings.Join(highlightedLines, "\n")
			} else {
				replacement = fmt.Sprintf("\033[38;5;39m%s\033[0m", filledValue)
			}
		} else {
			// Unfilled and not focused - show variable name in dim gray (242), NO brackets
			replacement = fmt.Sprintf("\033[38;5;242m%s\033[0m", varName)
		}

		// Replace all case variations of this variable
		variations := []string{
			varName,
			strings.ToUpper(varName),
			strings.ToLower(varName),
			strings.Title(strings.ToLower(varName)),
		}

		for _, variant := range variations {
			varPattern := "{{" + variant + "}}"
			result = strings.ReplaceAll(result, varPattern, replacement)
		}
	}

	return result
}

// highlightVariablesBeforeEdit highlights {{variables}} with unique colors BEFORE edit mode
// File variables: blue (39), Date/Time: green (34), Custom: yellow (220)
// Keeps the {{}} brackets visible
func (m model) highlightVariablesBeforeEdit(templateText string) string {
	if m.preview.promptTemplate == nil || len(m.preview.promptTemplate.variables) == 0 {
		return templateText
	}

	result := templateText

	// Assign colors based on variable type
	for _, varName := range m.preview.promptTemplate.variables {
		varNameLower := strings.ToLower(varName)
		var color string

		// Determine color based on variable type
		if varNameLower == "date" || varNameLower == "time" {
			color = "34" // Green for auto-filled date/time
		} else if strings.Contains(varNameLower, "file") || strings.Contains(varNameLower, "path") {
			color = "39" // Blue for file-related
		} else {
			color = "220" // Yellow for custom variables
		}

		// Replace all case variations
		variations := []string{
			varName,
			strings.ToUpper(varName),
			strings.ToLower(varName),
			strings.Title(strings.ToLower(varName)),
		}

		for _, variant := range variations {
			pattern := "{{" + variant + "}}"
			colored := fmt.Sprintf("\033[38;5;%sm{{%s}}\033[0m", color, variant)
			result = strings.ReplaceAll(result, pattern, colored)
		}
	}

	return result
}

// renderPromptPreview renders a prompt file with metadata header
func (m model) renderPromptPreview(maxVisible int) string {
	var s strings.Builder
	tmpl := m.preview.promptTemplate

	// Calculate box content width early so we can use it for all header elements
	var boxContentWidth int
	if m.viewMode == viewFullPreview {
		boxContentWidth = m.width - 6
	} else {
		boxContentWidth = m.rightWidth - 2
	}

	// Calculate wrapping width for header elements (leave room for padding)
	headerWrapWidth := boxContentWidth - 2
	if headerWrapWidth < 20 {
		headerWrapWidth = 20 // Minimum width for readability
	}

	// Build header lines
	var headerLines []string

	// Prompt name (if available)
	if tmpl.name != "" {
		nameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
		nameLine := "📝 " + tmpl.name

		// Wrap name if too long (use visual width)
		if visualWidth(nameLine) > headerWrapWidth {
			wrapped := wrapLine(nameLine, headerWrapWidth)
			for _, line := range wrapped {
				headerLines = append(headerLines, nameStyle.Render(line))
			}
		} else {
			headerLines = append(headerLines, nameStyle.Render(nameLine))
		}
		headerLines = append(headerLines, "") // Blank line
	}

	// Description (if available) - wrap long descriptions
	if tmpl.description != "" {
		descStyle := lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("245"))

		// Wrap description if too long (use visual width, not byte length)
		if visualWidth(tmpl.description) > headerWrapWidth {
			wrapped := wrapLine(tmpl.description, headerWrapWidth)
			for _, line := range wrapped {
				headerLines = append(headerLines, descStyle.Render(line))
			}
		} else {
			headerLines = append(headerLines, descStyle.Render(tmpl.description))
		}
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

	// Wrap source line if needed (use visual width)
	sourceLine := sourceIcon + " " + sourceLabel
	if visualWidth(sourceLine) > headerWrapWidth {
		wrapped := wrapLine(sourceLine, headerWrapWidth)
		for _, line := range wrapped {
			headerLines = append(headerLines, sourceStyle.Render(line))
		}
	} else {
		headerLines = append(headerLines, sourceStyle.Render(sourceLine))
	}

	// Variables detected (if any) - wrap long lines to prevent layout issues
	if len(tmpl.variables) > 0 {
		// Count how many times each variable appears
		varCounts := countVariableOccurrences(tmpl.template)

		// Build variable display strings with colors and counts
		varDisplays := make([]string, 0, len(tmpl.variables))
		for _, varName := range tmpl.variables {
			// Check if variable has been filled
			filled := false
			if val, exists := m.filledVariables[varName]; exists && val != "" {
				filled = true
			}

			// Build display string: varname or varname (3)
			display := varName
			count := varCounts[varName]
			if count > 1 {
				display = fmt.Sprintf("%s (%d)", varName, count)
			}

			// Color: green if filled, gray if not
			if filled {
				filledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82")) // Green
				display = filledStyle.Render(display)
			} else {
				unfilledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242")) // Gray
				display = unfilledStyle.Render(display)
			}

			varDisplays = append(varDisplays, display)
		}

		// Build the variables line with proper label
		labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
		varsLine := labelStyle.Render("Variables: ") + strings.Join(varDisplays, labelStyle.Render(", "))

		// Wrap the variables line if it's too long (use visual width, not byte length)
		// Note: visualWidth doesn't account for ANSI codes, so this is approximate
		plainVarsLine := fmt.Sprintf("Variables: %s", strings.Join(tmpl.variables, ", "))
		if visualWidth(plainVarsLine) > headerWrapWidth {
			// For simplicity, just add the line (wrapping styled text is complex)
			// The visual width check prevents most overflow cases
			headerLines = append(headerLines, varsLine)
		} else {
			headerLines = append(headerLines, varsLine)
		}
	}

	// Separator line - use full header wrap width (each char is 1 visual width)
	separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	separatorLine := strings.Repeat("─", headerWrapWidth) // Full width separator
	headerLines = append(headerLines, separatorStyle.Render(separatorLine))
	// No blank line after separator - the separator itself provides visual separation

	// Calculate how many lines the header takes
	headerHeight := len(headerLines)

	// Calculate available width for content (prompts don't show line numbers, so use full width)
	availableWidth := boxContentWidth - 2 // Just padding

	if availableWidth < 20 {
		availableWidth = 20
	}

	// Determine content to display - always use preview content with inline variable highlighting
	var contentLines []string
	var isMarkdownPrompt bool

	// Use preview content (variables already substituted in prompt_parser.go)
	contentText := strings.Join(m.preview.content, "\n")

	// Check if this is a markdown file
	isMarkdownPrompt = m.preview.isMarkdown

	if isMarkdownPrompt && !m.promptEditMode {
		// Use cached Glamour rendering if available and valid (prevents lag on every frame)
		if m.preview.cachedRenderedContent != "" && m.preview.cachedWidth == availableWidth {
			// Use cached rendering
			contentLines = strings.Split(strings.TrimRight(m.preview.cachedRenderedContent, "\n"), "\n")
		} else {
			// Cache miss or invalid - render with Glamour (with timeout to prevent hangs)
			rendered, err := m.renderMarkdownWithTimeout(contentText, availableWidth, 5*time.Second)
			if err == nil {
				// Successfully rendered with Glamour
				contentLines = strings.Split(strings.TrimRight(rendered, "\n"), "\n")
			} else {
				// Glamour failed or timed out, fall back to plain text
				contentLines = m.preview.content
				isMarkdownPrompt = false
			}
		}
	} else {
		// Plain text or edit mode - use plain content
		if m.promptEditMode {
			// EDIT MODE: Show raw template with inline variable highlighting (no brackets)
			renderedTemplate := m.renderInlineVariables(tmpl.template)
			contentLines = strings.Split(renderedTemplate, "\n")
		} else if m.preview.isPrompt && m.showPromptsOnly {
			// BEFORE EDIT MODE: Show template with colored {{variables}} (keeps brackets)
			highlightedTemplate := m.highlightVariablesBeforeEdit(tmpl.template)
			contentLines = strings.Split(highlightedTemplate, "\n")
		} else {
			// Regular preview content (variables already substituted)
			contentLines = m.preview.content
		}
	}

	// Wrap content lines (markdown is already wrapped by Glamour, but ANSI-styled text needs wrapping)
	var wrappedLines []string
	if isMarkdownPrompt {
		// Glamour already wrapped - use as-is
		wrappedLines = contentLines
	} else {
		// Wrap all other content (plain text, edit mode, or colored variables)
		// wrapLine() uses visualWidth() which correctly handles ANSI codes
		for _, line := range contentLines {
			wrapped := wrapLine(line, availableWidth)
			wrappedLines = append(wrappedLines, wrapped...)
		}
	}

	// Calculate available height for content (no separate input fields section)
	contentHeight := maxVisible - headerHeight
	if contentHeight < 5 {
		contentHeight = 5
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
		// When end exceeds total lines, adjust start backwards to show a full page
		end = totalLines
		start = max(0, end-contentHeight)
	}

	// Collect rendered lines so we can join without trailing newline
	renderedLines := make([]string, 0, maxVisible)

	// Render header - truncate each line to fit within box width
	for _, line := range headerLines {
		if len(renderedLines) >= maxVisible {
			break
		}
		// Truncate to box content width to prevent terminal wrapping
		// (lipgloss styles and emojis can exceed expected width)
		// IMPORTANT: Use terminal-aware functions for emoji width compensation
		if m.visualWidthCompensated(line) > boxContentWidth {
			line = m.truncateToWidthCompensated(line, boxContentWidth)
		}
		renderedLines = append(renderedLines, line+"\033[0m")
	}

	// Reserve last line for scroll indicator in dual-pane mode
	targetLines := maxVisible
	if m.viewMode == viewDualPane && totalLines > 0 {
		targetLines = maxVisible - 1
	}

	// Render content (no line numbers for prompts, but show scrollbar)
	contentLineIndex := 0
	for i := start; i < end && len(renderedLines) < targetLines; i++ {
		line := wrappedLines[i]

		// Add scrollbar indicator for prompts (since they don't have line numbers)
		scrollbar := m.renderScrollbar(contentLineIndex, targetLines-len(headerLines), totalLines)

		// Add scrollbar + space + content
		if isMarkdownPrompt {
			line = scrollbar + " " + line // scrollbar + space + content
		} else {
			line = scrollbar + " " + line // scrollbar + space + content
		}

		// Truncate to box content width to prevent terminal wrapping
		// (even though content was wrapped, markdown padding or ANSI codes could push it over)
		// IMPORTANT: Use terminal-aware functions for emoji width compensation
		if m.visualWidthCompensated(line) > boxContentWidth {
			line = m.truncateToWidthCompensated(line, boxContentWidth)
		}
		renderedLines = append(renderedLines, line+"\033[0m")
		contentLineIndex++
	}

	// Add scroll position indicator in dual-pane mode
	if m.viewMode == viewDualPane && totalLines > 0 {
		// Calculate scroll percentage based on how far through scrollable content we are
		contentLinesAvailable := targetLines - len(headerLines)
		maxScrollPos := totalLines - contentLinesAvailable
		var scrollPercent int
		if maxScrollPos <= 0 {
			// Content fits in one screen
			scrollPercent = 100
		} else {
			scrollPercent = (m.preview.scrollPos * 100) / maxScrollPos
			if scrollPercent > 100 {
				scrollPercent = 100
			}
		}

		// Show the last visible line number (not the top line)
		// Use end instead of totalLines since end is already correctly clamped
		lastVisibleLine := min(start+contentLinesAvailable, end)
		scrollIndicator := fmt.Sprintf(" %d/%d (%d%%) ", lastVisibleLine, totalLines, scrollPercent)
		scrollStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)

		// Pad with empty lines to reach target
		for len(renderedLines) < targetLines {
			renderedLines = append(renderedLines, "\033[0m")
		}

		// Add scroll indicator on last line
		renderedLines = append(renderedLines, scrollStyle.Render(scrollIndicator))
	} else {
		// Pad with empty lines to reach exactly maxVisible lines
		for len(renderedLines) < maxVisible {
			renderedLines = append(renderedLines, "\033[0m")
		}
	}

	for i, line := range renderedLines {
		if i > 0 {
			s.WriteString("\n")
		}
		s.WriteString(line)
	}

	return s.String()
}

// renderScrollbar renders a scrollbar indicator for the current line
// Now renders in place of the separator between line numbers and content
func (m model) renderScrollbar(lineIndex, visibleLines, totalLines int) string {
	// Hide scrollbar in text selection mode (prevents it being copied)
	if !m.previewMouseEnabled {
		return " " // Return space to maintain spacing
	}

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

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// renderFullPreview renders the full-screen preview mode
func (m model) renderFullPreview() string {
	var s strings.Builder

	// Only show title bar and info line when mouse is enabled (not in text selection mode)
	headerLines := 0
	if m.previewMouseEnabled {
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

		// File info line with scroll position percentage
		var infoText string

		// Calculate scroll percentage
		totalLines := m.getWrappedLineCount()
		// Calculate how many lines will be visible (need to calculate early for percentage)
		maxVisible := m.height - 4 - 2 // headerLines = 2 when mouse enabled
		contentHeight := maxVisible - 2

		var scrollPercent int
		var lastVisibleLine int
		if totalLines > 0 {
			// Calculate percentage based on how far through scrollable content we are
			maxScrollPos := totalLines - contentHeight
			if maxScrollPos <= 0 {
				// Content fits in one screen
				scrollPercent = 100
				lastVisibleLine = totalLines
			} else {
				scrollPercent = (m.preview.scrollPos * 100) / maxScrollPos
				if scrollPercent > 100 {
					scrollPercent = 100
				}
				// Show the last visible line number (not the top line)
				lastVisibleLine = min(m.preview.scrollPos+contentHeight, totalLines)
			}
		}

		if m.preview.isMarkdown {
			// Show scroll position for markdown too
			if totalLines > 0 {
				infoText = fmt.Sprintf("Size: %s | Markdown Rendered | Line %d/%d (%d%%)",
					formatFileSize(m.preview.fileSize),
					lastVisibleLine,
					totalLines,
					scrollPercent)
			} else {
				infoText = fmt.Sprintf("Size: %s | Markdown Rendered",
					formatFileSize(m.preview.fileSize))
			}
		} else {
			// Show scroll position for regular text
			if totalLines > 0 {
				infoText = fmt.Sprintf("Size: %s | Lines: %d (wrapped) | Line %d/%d (%d%%)",
					formatFileSize(m.preview.fileSize),
					len(m.preview.content),
					lastVisibleLine,
					totalLines,
					scrollPercent)
			} else {
				infoText = fmt.Sprintf("Size: %s | Lines: %d (wrapped)",
					formatFileSize(m.preview.fileSize),
					len(m.preview.content))
			}
		}
		s.WriteString(pathStyle.Render(infoText))
		s.WriteString("\033[0m") // Reset ANSI codes
		s.WriteString("\n")

		headerLines = 2 // title + info line
	}

	// Content with border
	// Reserve space based on whether header is shown
	maxVisible := m.height - 4 - headerLines // Reserve space for header (if shown), help, and borders
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

	// Show different F5 text if viewing a prompt
	f5Text := "copy path"
	if m.preview.isPrompt {
		f5Text = "copy rendered prompt"
	}

	// Mouse toggle indicator - show what 'm' key does
	var modeText, helpText string
	if m.previewMouseEnabled {
		modeText = "🖱 text select"  // Press m to enable text selection
	} else {
		modeText = "⌨ mouse scroll"  // Press m to enable mouse scrolling
	}

	// Build help text
	if m.preview.isBinary && isImageFile(m.preview.filePath) {
		helpText = fmt.Sprintf("F1: help • V: view image • m: %s • F4: edit • Esc: close", modeText)
	} else {
		helpText = fmt.Sprintf("F1: help • ↑/↓: scroll • m: %s • F4: edit • F5: %s • Esc: close", modeText, f5Text)
	}
	s.WriteString(helpStyle.Render(helpText))
	s.WriteString("\033[0m") // Reset ANSI codes

	// Show search input if search is active
	if m.preview.searchActive {
		s.WriteString("\n")
		searchStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("33")). // Blue background
			Foreground(lipgloss.Color("0")).  // Black text
			Bold(true).
			Padding(0, 1)

		matchCount := len(m.preview.searchMatches)
		var searchText string
		if matchCount > 0 {
			currentPos := m.preview.currentMatch + 1
			searchText = fmt.Sprintf("🔍 Search: %s█ (%d/%d matches)", m.preview.searchQuery, currentPos, matchCount)
		} else if m.preview.searchQuery == "" {
			searchText = "🔍 Search: █ (type to search, n/Shift+N: navigate, Esc: exit)"
		} else {
			searchText = fmt.Sprintf("🔍 Search: %s█ (no matches)", m.preview.searchQuery)
		}

		s.WriteString(searchStyle.Render(searchText))
		s.WriteString("\033[0m") // Reset ANSI codes
	} else if m.statusMessage != "" && (m.promptEditMode || m.filePickerMode || time.Since(m.statusTime) < 3*time.Second) {
		// Show status message if present (auto-dismiss after 3s, except in edit mode or file picker mode) and search not active
		s.WriteString("\n")
		msgStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("28")). // Green
			Foreground(lipgloss.Color("15")). // White for better contrast
			Bold(true).
			Padding(0, 1)

		if m.statusIsError {
			msgStyle = msgStyle.Background(lipgloss.Color("196")) // Red
		}

		s.WriteString(msgStyle.Render(m.statusMessage))
		s.WriteString("\033[0m") // Reset ANSI codes
	}

	return s.String()
}

// renderDualPane renders the split-pane layout using Lipgloss layout utilities
func (m model) renderDualPane() string {
	var s strings.Builder

	// Check if we should show GitHub link (first 5 seconds) or menu bar
	showGitHub := time.Since(m.startupTime) < 5*time.Second

	if showGitHub {
		// Title with mode indicator (first 5 seconds) + terminal type for debugging
		titleText := fmt.Sprintf("(T)erminal (F)ile (E)xplorer [Dual-Pane] (%s)", m.terminalType.String())
		if m.commandFocused {
			titleText += " [Command Mode]"
		}
		if m.filePickerMode {
			if m.filePickerCopySource != "" {
				titleText += " [📋 Copy Mode - Select Destination]"
			} else {
				titleText += " [📁 File Picker]"
			}
		}

		// Right side: Update notification or GitHub link
		var rightLink string
		var displayText string

		if m.updateAvailable {
			// Show update available with clickable link
			displayText = fmt.Sprintf("🎉 Update Available: %s (click for details)", m.updateVersion)
			// Use special marker URL so we can detect clicks in mouse handler
			rightLink = fmt.Sprintf("\033]8;;update-available\033\\%s\033]8;;\033\\", displayText)
		} else {
			// Show GitHub link
			githubURL := "https://github.com/GGPrompts/TFE"
			displayText = githubURL
			rightLink = fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", githubURL, githubURL)
		}

		// Calculate spacing to right-align
		availableWidth := m.width - len(titleText) - len(displayText) - 2
		if availableWidth < 1 {
			availableWidth = 1
		}
		spacing := strings.Repeat(" ", availableWidth)

		// Render title on left, link/update on right
		title := titleStyle.Render(titleText) + spacing + titleStyle.Render(rightLink)
		s.WriteString(title)
		s.WriteString("\033[0m") // Reset ANSI codes
		s.WriteString("\n")
	} else {
		// Show menu bar after 5 seconds
		menuBar := m.renderMenuBar()
		s.WriteString(menuBar)
		s.WriteString("\n")
	}

	// Toolbar buttons
	// Home button - highlight with gray background when in home directory
	homeDir, _ := os.UserHomeDir()
	// Home button
	homeIcon := "🏠"
	if homeDir != "" && m.currentPath == homeDir {
		// Active: gray background (in home directory)
		homeButtonStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true).
			Background(lipgloss.Color("237"))
		s.WriteString(homeButtonStyle.Render("[" + homeIcon + "]"))
	} else {
		// Inactive: normal styling
		homeButtonStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)
		s.WriteString(homeButtonStyle.Render("[" + homeIcon + "]"))
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

	// View mode toggle button (cycles List → Detail → Tree)
	// Show different emoji based on current display mode
	viewIcon := "📊" // Detail view (default)
	switch m.displayMode {
	case modeList:
		viewIcon = "📄" // Document icon for simple list view
	case modeDetail:
		viewIcon = "📊" // Bar chart icon for detailed columns
	case modeTree:
		viewIcon = "🌲" // Tree icon for hierarchical view
	}
	viewButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)
	s.WriteString(viewButtonStyle.Render("[" + viewIcon + "]"))
	s.WriteString(" ")

	// Pane toggle button (toggles single ↔ dual-pane)
	paneIcon := "⬜"
	if m.viewMode == viewDualPane {
		paneIcon = "⬌"
	}
	paneButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)
	s.WriteString(paneButtonStyle.Render("[" + paneIcon + "]"))
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

	// Context-aware search button (in-file search when viewing, directory filter when browsing)
	// Highlight when search is active (either in-file or directory filter)
	searchIcon := "🔍"
	if m.preview.searchActive || m.searchMode {
		// Active: gray background
		activeSearchStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true).
			Background(lipgloss.Color("237"))
		s.WriteString(activeSearchStyle.Render("[" + searchIcon + "]"))
	} else {
		// Inactive: normal styling
		searchButtonStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)
		s.WriteString(searchButtonStyle.Render("[" + searchIcon + "]"))
	}
	s.WriteString(" ")

	// Prompts filter toggle button
	promptIcon := "📝"
	if m.showPromptsOnly {
		// Active: gray background (like command mode)
		activeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true).
			Background(lipgloss.Color("237"))
		s.WriteString(activeStyle.Render("[" + promptIcon + "]"))
	} else {
		// Inactive: normal styling
		promptButtonStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)
		s.WriteString(promptButtonStyle.Render("[" + promptIcon + "]"))
	}
	s.WriteString(" ")

	// Git repositories toggle button
	gitIcon := "🔀"
	if m.showGitReposOnly {
		// Active: gray background (like other active toggles)
		activeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true).
			Background(lipgloss.Color("237"))
		s.WriteString(activeStyle.Render("[" + gitIcon + "]"))
	} else {
		// Inactive: normal styling
		gitButtonStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)
		s.WriteString(gitButtonStyle.Render("[" + gitIcon + "]"))
	}
	s.WriteString(" ")

	// Trash/Recycle bin button
	trashIcon := "🗑"
	if m.showTrashOnly {
		trashIcon = "♻" // Recycle icon when viewing trash
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

	// Show helper text based on focus state
	helperStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	if !m.commandFocused && m.commandInput == "" {
		// Not focused - show how to enter command mode
		s.WriteString(helperStyle.Render(": to focus"))
	} else if m.commandFocused && m.commandInput == "" {
		// Focused but no input - show ! prefix hint and cursor
		s.WriteString(helperStyle.Render("! prefix to run & exit"))
		cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
		s.WriteString(cursorStyle.Render("█"))
	} else {
		// Has input - show the command with cursor at correct position
		if m.commandFocused {
			// Render text before cursor, cursor, text after cursor
			beforeCursor := m.commandInput[:m.commandCursorPos]
			afterCursor := m.commandInput[m.commandCursorPos:]

			// Handle ! prefix coloring
			if strings.HasPrefix(beforeCursor, "!") {
				prefixStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
				s.WriteString(prefixStyle.Render("!"))
				s.WriteString(inputStyle.Render(beforeCursor[1:]))
			} else {
				s.WriteString(inputStyle.Render(beforeCursor))
			}

			// Render cursor
			cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
			s.WriteString(cursorStyle.Render("█"))

			// Render text after cursor
			s.WriteString(inputStyle.Render(afterCursor))
		} else {
			// Not focused - just show the text
			if strings.HasPrefix(m.commandInput, "!") {
				prefixStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
				s.WriteString(prefixStyle.Render("!"))
				s.WriteString(inputStyle.Render(m.commandInput[1:]))
			} else {
				s.WriteString(inputStyle.Render(m.commandInput))
			}
		}
	}
	// Explicitly reset styling after cursor to prevent ANSI code leakage
	s.WriteString("\033[0m")
	s.WriteString("\n")

	// Blank line separator between command prompt and panes
	s.WriteString("\n")

	// Calculate max visible for both panes
	// Layout: title(1) + toolbar(1) + command(1) + blank(1) + panes(maxVisible) + blank_after(1) + status(2) + optional(1)
	// Total: 4 + maxVisible + 1 + 2 + (0-1) = maxVisible + 7-8
	// Use worst case (8) to ensure panes never overflow
	headerLines := 4  // title + toolbar + command + blank separator
	footerLines := 4  // blank after panes + 2 status lines + optional message/search
	maxVisible := m.height - headerLines - footerLines
	if maxVisible < 5 {
		maxVisible = 5 // Minimum pane height
	}

	// Content area is maxVisible - 2 (accounting for top/bottom borders)
	contentHeight := maxVisible - 2

	// Render panes based on display mode
	var panes string

	if m.displayMode == modeDetail {
		// VERTICAL SPLIT for detail view - gives full width to detail columns
		// ACCORDION: Focused pane gets 2/3 height, unfocused gets 1/3
		var topHeight, bottomHeight int
		if m.focusedPane == leftPane {
			// Top pane (detail view) is focused
			topHeight = (maxVisible * 2) / 3
			bottomHeight = maxVisible - topHeight
		} else {
			// Bottom pane (preview) is focused
			bottomHeight = (maxVisible * 2) / 3
			topHeight = maxVisible - bottomHeight
		}

		topContentHeight := topHeight - 2    // Account for borders
		bottomContentHeight := bottomHeight - 2

		// Render top pane (detail view with full width)
		topContent := m.renderDetailView(topContentHeight)

		// Render bottom pane (preview with full width)
		var bottomContent string
		if m.preview.loaded {
			bottomContent = m.renderPreview(bottomContentHeight)
		} else {
			emptyStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Italic(true)
			bottomContent = emptyStyle.Render("No preview available\n\nSelect a file to preview") + "\033[0m"
		}

		// Border colors based on focus
		topBorderColor := lipgloss.AdaptiveColor{Light: "#999999", Dark: "#585858"}
		bottomBorderColor := lipgloss.AdaptiveColor{Light: "#999999", Dark: "#585858"}
		if m.focusedPane == leftPane {
			topBorderColor = lipgloss.AdaptiveColor{Light: "#0087d7", Dark: "#00d7ff"}
		} else {
			bottomBorderColor = lipgloss.AdaptiveColor{Light: "#0087d7", Dark: "#00d7ff"}
		}

		// Create boxes with full width
		topPaneStyle := lipgloss.NewStyle().
			Width(m.width - 6).           // Full width minus margins
			Height(topContentHeight).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(topBorderColor)

		bottomPaneStyle := lipgloss.NewStyle().
			Width(m.width - 6).           // Full width minus margins
			Height(bottomContentHeight).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(bottomBorderColor)

		topPaneRendered := topPaneStyle.Render(topContent)
		bottomPaneRendered := bottomPaneStyle.Render(bottomContent)

		// Stack vertically
		panes = lipgloss.JoinVertical(lipgloss.Left, topPaneRendered, bottomPaneRendered)

	} else {
		// List/Tree view - use VERTICAL split on narrow terminals, HORIZONTAL on wide terminals

		if m.isNarrowTerminal() {
			// VERTICAL SPLIT for narrow terminals (phones) - same as detail view
			// ACCORDION: Focused pane gets 2/3 height, unfocused gets 1/3
			var topHeight, bottomHeight int
			if m.focusedPane == leftPane {
				// Top pane (file list) is focused
				topHeight = (maxVisible * 2) / 3
				bottomHeight = maxVisible - topHeight
			} else {
				// Bottom pane (preview) is focused
				bottomHeight = (maxVisible * 2) / 3
				topHeight = maxVisible - bottomHeight
			}

			topContentHeight := topHeight - 2    // Account for borders
			bottomContentHeight := bottomHeight - 2

			// Render top pane (file list with full width)
			var topContent string
			switch m.displayMode {
			case modeList:
				topContent = m.renderListView(topContentHeight)
			case modeTree:
				topContent = m.renderTreeView(topContentHeight)
			default:
				topContent = m.renderListView(topContentHeight)
			}

			// Render bottom pane (preview with full width)
			var bottomContent string
			if m.preview.loaded {
				bottomContent = m.renderPreview(bottomContentHeight)
			} else {
				emptyStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("241")).
					Italic(true)
				bottomContent = emptyStyle.Render("No preview available\n\nSelect a file to preview") + "\033[0m"
			}

			// Border colors based on focus
			topBorderColor := lipgloss.AdaptiveColor{Light: "#999999", Dark: "#585858"}
			bottomBorderColor := lipgloss.AdaptiveColor{Light: "#999999", Dark: "#585858"}
			if m.focusedPane == leftPane {
				topBorderColor = lipgloss.AdaptiveColor{Light: "#0087d7", Dark: "#00d7ff"}
			} else {
				bottomBorderColor = lipgloss.AdaptiveColor{Light: "#0087d7", Dark: "#00d7ff"}
			}

			// Create boxes with full width
			topPaneStyle := lipgloss.NewStyle().
				Width(m.width - 6).           // Full width minus margins
				Height(topContentHeight).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(topBorderColor)

			bottomPaneStyle := lipgloss.NewStyle().
				Width(m.width - 6).           // Full width minus margins
				Height(bottomContentHeight).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(bottomBorderColor)

			topPaneRendered := topPaneStyle.Render(topContent)
			bottomPaneRendered := bottomPaneStyle.Render(bottomContent)

			// Stack vertically
			panes = lipgloss.JoinVertical(lipgloss.Left, topPaneRendered, bottomPaneRendered)

		} else {
			// HORIZONTAL SPLIT for wide terminals - accordion style
			// Get left pane content - use contentHeight so content fits within the box
			var leftContent string
			switch m.displayMode {
			case modeList:
				leftContent = m.renderListView(contentHeight)
			case modeTree:
				leftContent = m.renderTreeView(contentHeight)
			default:
				leftContent = m.renderListView(contentHeight)
			}

			// Get right pane content (preview)
			var rightContent string
			if m.preview.loaded {
				rightContent = m.renderPreview(contentHeight)
			} else {
				emptyStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("241")).
					Italic(true)
				rightContent = emptyStyle.Render("No preview available\n\nSelect a file to preview") + "\033[0m"
			}

			// Border colors based on focus (accordion style)
			leftBorderColor := lipgloss.AdaptiveColor{Light: "#999999", Dark: "#585858"}
			rightBorderColor := lipgloss.AdaptiveColor{Light: "#999999", Dark: "#585858"}
			if m.focusedPane == leftPane {
				leftBorderColor = lipgloss.AdaptiveColor{Light: "#0087d7", Dark: "#00d7ff"}
			} else {
				rightBorderColor = lipgloss.AdaptiveColor{Light: "#0087d7", Dark: "#00d7ff"}
			}

			// Use exact Width and Height to ensure panes stay perfectly aligned
			leftPaneStyle := lipgloss.NewStyle().
				Width(m.leftWidth - 2).      // Content width (borders added by Lipgloss)
				Height(contentHeight).        // Exact content height (borders added by Lipgloss)
				Border(lipgloss.RoundedBorder()).
				BorderForeground(leftBorderColor)

			rightPaneStyle := lipgloss.NewStyle().
				Width(m.rightWidth - 2).     // Content width (borders added by Lipgloss)
				Height(contentHeight).        // Exact content height (borders added by Lipgloss)
				Border(lipgloss.RoundedBorder()).
				BorderForeground(rightBorderColor)

			// Apply styles to content
			leftPaneRendered := leftPaneStyle.Render(leftContent)
			rightPaneRendered := rightPaneStyle.Render(rightContent)

			// Join panes horizontally
			panes = lipgloss.JoinHorizontal(lipgloss.Top, leftPaneRendered, rightPaneRendered)
		}
	}

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

	gitReposIndicator := ""
	if m.showGitReposOnly {
		gitReposIndicator = " • 🔀 git repos only"
	}

	// Show focused pane info in status bar
	focusInfo := ""
	if m.focusedPane == leftPane {
		focusInfo = " • [LEFT focused]"
	} else {
		focusInfo = " • [RIGHT focused]"
	}

	// Help hint - show "/" search hint only when not already searching
	helpHint := " • F1: help"
	if !m.searchMode && m.searchQuery == "" {
		helpHint += " • /: search"
	}

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

			// For symlinks, truncate long paths to show the important trailing part
			if currentFile.isSymlink && currentFile.symlinkTarget != "" {
				// Calculate available space: terminal width minus other info
				// "Selected: filename (, size, date)"
				baseInfoLen := len("Selected: ") + len(currentFile.name) + len(", ") +
					len(formatFileSize(currentFile.size)) + len(", ") +
					len(formatModTime(currentFile.modTime)) + len(" ()") + 10 // padding

				availableForTarget := m.width - baseInfoLen
				if availableForTarget < 30 {
					availableForTarget = 30 // Minimum to show something useful
				}

				fullTarget := "Link → " + currentFile.symlinkTarget
				if len(fullTarget) > availableForTarget {
					// Show trailing end: "...filename" instead of "Link → /very/long/pa..."
					fileType = "..." + fullTarget[len(fullTarget)-(availableForTarget-3):]
				}
			}

			selectedInfo = fmt.Sprintf("Selected: %s (%s, %s, %s)",
				currentFile.name,
				fileType,
				formatFileSize(currentFile.size),
				formatModTime(currentFile.modTime))
		}
	}

	// Split status into two lines to prevent truncation
	// Line 1: Counts, indicators, view mode, focus, help
	statusLine1 := fmt.Sprintf("%s%s%s%s%s • %s%s%s", itemsInfo, hiddenIndicator, favoritesIndicator, promptsIndicator, gitReposIndicator, m.displayMode.String(), focusInfo, helpHint)
	s.WriteString(statusStyle.Render(statusLine1))
	s.WriteString("\033[0m") // Reset ANSI codes
	s.WriteString("\n")

	// Line 2: Selected file info
	statusLine2 := selectedInfo
	s.WriteString(statusStyle.Render(statusLine2))
	s.WriteString("\033[0m") // Reset ANSI codes

	// Show status message if present (auto-dismiss after 3s, except in edit mode or file picker mode)
	if m.statusMessage != "" && (m.promptEditMode || m.filePickerMode || time.Since(m.statusTime) < 3*time.Second) {
		s.WriteString("\n")
		msgStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("28")). // Green
			Foreground(lipgloss.Color("15")). // White for better contrast
			Bold(true).
			Padding(0, 1)

		if m.statusIsError {
			msgStyle = msgStyle.Background(lipgloss.Color("196")) // Red
		}

		s.WriteString(msgStyle.Render(m.statusMessage))
		s.WriteString("\033[0m") // Reset ANSI codes
	} else if m.searchMode || m.searchQuery != "" {
		// Show search status
		s.WriteString("\n")
		searchStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("33")). // Blue background
			Foreground(lipgloss.Color("255")). // Bright white for high contrast
			Bold(true).
			Padding(0, 1)

		// Calculate match count (exclude parent directory "..")
		matchCount := len(m.filteredIndices)
		if matchCount > 0 {
			matchCount-- // Exclude ".." which is always included
		}

		var searchStatus string
		if m.searchMode {
			// Active search mode with cursor
			searchStatus = fmt.Sprintf("Search: %s█ (%d matches)", m.searchQuery, matchCount)
		} else {
			// Search accepted (filter active but not in input mode)
			searchStatus = fmt.Sprintf("Filtered: %s (%d matches)", m.searchQuery, matchCount)
		}

		s.WriteString(searchStyle.Render(searchStatus))
		s.WriteString("\033[0m") // Reset ANSI codes
	}

	return s.String()
}
