package main

// Module: render_prompts.go
// Purpose: Prompt template rendering and variable highlighting
// Responsibilities:
// - Rendering prompt files with metadata headers
// - Inline variable highlighting for edit mode
// - Variable highlighting before edit mode

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// renderPromptPreview renders a prompt file with metadata header
func (m model) renderPromptPreview(maxVisible int) string {
	var s strings.Builder
	tmpl := m.preview.promptTemplate

	// Calculate box content width early so we can use it for all header elements
	var boxContentWidth int
	if m.viewMode == viewFullPreview {
		boxContentWidth = m.width - 6
	} else {
		// In dual-pane mode, use consistent width calculation
		// Subtract 6 to match full preview mode (accounts for borders and padding)
		boxContentWidth = m.rightWidth - 6
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
		// Check if wrapping is needed based on plain text (without ANSI codes)
		plainVarsLine := fmt.Sprintf("Variables: %s", strings.Join(tmpl.variables, ", "))

		if visualWidth(plainVarsLine) > headerWrapWidth {
			// Variables line needs wrapping - wrap the plain text first
			wrappedPlainLines := wrapLine(plainVarsLine, headerWrapWidth)

			// For each wrapped line, apply the appropriate styling
			labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
			for _, plainLine := range wrappedPlainLines {
				// Simple approach: style the entire line uniformly for wrapped content
				// This avoids complex ANSI code handling across line breaks
				styledLine := labelStyle.Render(plainLine)
				headerLines = append(headerLines, styledLine)
			}
		} else {
			// Variables line fits in one line - use original styling with colors
			labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
			varsLine := labelStyle.Render("Variables: ") + strings.Join(varDisplays, labelStyle.Render(", "))
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
