package main

// Module: text_wrapping.go
// Purpose: Text wrapping and line counting utilities
// Responsibilities:
// - Wrapping text to fit within width constraints
// - Counting wrapped lines for scrolling calculations
// - Calculating prompt header heights for layout

import (
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
)

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
		boxContentWidth = m.rightWidth - 6 // Box content width in dual-pane (match full preview)
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
			contentLineCount := len(renderedLines)

			// For prompt files, return only content line count
			// The header is fixed and doesn't scroll, so it shouldn't be included
			// in scroll calculations
			return contentLineCount
		}
		// Fallback if glamour fails or times out
	}

	// For regular text, count wrapped lines
	totalLines := 0
	for _, line := range m.preview.content {
		wrapped := wrapLine(line, availableWidth)
		totalLines += len(wrapped)
	}

	// For prompt files, return only content line count
	// The header is fixed and doesn't scroll, so it shouldn't be included
	// in scroll calculations
	return totalLines
}

// visualWidth calculates the visual width of a string, accounting for tabs and ANSI codes
// This is important for consistent scrollbar alignment and box borders
// NOTE: This is the non-terminal-aware version - use m.visualWidthCompensated() for layout calculations
func visualWidth(s string) int {
	// Strip ANSI codes first
	stripped := ""
	inAnsi := false

	for _, ch := range s {
		// Detect start of ANSI escape sequence
		if ch == '\033' {
			inAnsi = true
			continue
		}

		// Skip characters inside ANSI sequences
		if inAnsi {
			// ANSI sequences end with a letter (A-Z, a-z)
			if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
				inAnsi = false
			}
			continue
		}

		// Keep visible characters
		stripped += string(ch)
	}

	return runewidth.StringWidth(stripped)
}

// visualWidthCompensated calculates visual width with terminal-specific emoji compensation
// Use this for layout calculations that need accurate emoji widths
// STRIPS ANSI escape codes before calculating width
func (m model) visualWidthCompensated(s string) int {
	// Use visualWidth() which strips ANSI codes, not runewidth.StringWidth()
	width := visualWidth(s)

	// Apply variation selector compensation for Windows Terminal ONLY
	// runewidth reports emoji+VS as 1 cell, but Windows Terminal renders as 2 cells
	// We ADD 1 per VS for Windows Terminal to match its wider rendering
	variationSelectorCount := strings.Count(s, "\uFE0F")
	if m.terminalType == terminalWindowsTerminal && variationSelectorCount > 0 {
		width += variationSelectorCount
	}

	// NOTE: xterm.js actually renders emojis as 2 cells (same as runewidth)
	// Do NOT compensate for xterm - trust runewidth's 2-cell calculation

	return width
}

// truncateToWidth truncates a string to fit within a target visual width
func truncateToWidth(s string, targetWidth int) string {
	width := 0
	result := ""
	inAnsi := false

	for _, ch := range s {
		// Handle ANSI escape sequences (don't count toward width)
		if ch == '\033' {
			inAnsi = true
			result += string(ch)
			continue
		}

		if inAnsi {
			result += string(ch)
			if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
				inAnsi = false
			}
			continue
		}

		// Calculate character width
		charWidth := 1
		if ch == '\t' {
			charWidth = 8 - (width % 8)
		} else {
			// Use runewidth to properly handle wide characters (emojis, CJK)
			charWidth = runewidth.RuneWidth(ch)
		}

		if width+charWidth > targetWidth {
			// Can't fit this character
			if targetWidth-width >= 3 {
				return result + "..."
			}
			return result
		}

		width += charWidth
		result += string(ch)
	}

	return result
}

// truncateToWidthCompensated truncates a string to fit within a target visual width
// with terminal-specific emoji width compensation (uses m.runeWidth for accurate widths)
func (m model) truncateToWidthCompensated(s string, targetWidth int) string {
	width := 0
	result := ""
	inAnsi := false

	for _, ch := range s {
		// Handle ANSI escape sequences (don't count toward width)
		if ch == '\033' {
			inAnsi = true
			result += string(ch)
			continue
		}

		if inAnsi {
			result += string(ch)
			if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
				inAnsi = false
			}
			continue
		}

		// Calculate character width using terminal-aware function
		charWidth := 1
		if ch == '\t' {
			charWidth = 8 - (width % 8)
		} else {
			// Use terminal-aware runeWidth to properly handle wide characters
			charWidth = m.runeWidth(ch)
		}

		if width+charWidth > targetWidth {
			// Can't fit this character
			if targetWidth-width >= 3 {
				return result + "..."
			}
			return result
		}

		width += charWidth
		result += string(ch)
	}

	return result
}

// padIconToWidth pads an icon emoji to a fixed width (2 cells) for consistent alignment
// Some terminals render certain emojis as 1 cell, so we pad them to 2 cells
func (m model) padIconToWidth(icon string) string {
	return m.padToVisualWidth(icon, 2)
}

// padToVisualWidth pads a string to a specific visual width using spaces
// This correctly handles emojis, wide characters, AND ANSI escape codes
// Terminal-aware: Variation selector compensation only for WezTerm
func (m model) padToVisualWidth(s string, targetWidth int) string {
	// Use visualWidthCompensated which already handles ANSI codes via m.visualWidthCompensated
	// which delegates to visualWidth for ANSI stripping
	calculatedWidth := m.visualWidthCompensated(s)

	if calculatedWidth >= targetWidth {
		return s
	}
	padding := targetWidth - calculatedWidth
	return s + strings.Repeat(" ", padding)
}

// getPromptHeaderHeight calculates how many lines the prompt header takes up
// This matches the logic in renderPromptPreview() to ensure consistent calculations
func (m model) getPromptHeaderHeight(boxContentWidth int) int {
	if !m.preview.isPrompt || m.preview.promptTemplate == nil {
		return 0
	}

	tmpl := m.preview.promptTemplate
	headerWrapWidth := boxContentWidth - 2 // Leave room for padding
	if headerWrapWidth < 20 {
		headerWrapWidth = 20
	}

	headerLineCount := 0

	// Prompt name (if available)
	if tmpl.name != "" {
		nameLine := "📝 " + tmpl.name
		if visualWidth(nameLine) > headerWrapWidth {
			wrapped := wrapLine(nameLine, headerWrapWidth)
			headerLineCount += len(wrapped)
		} else {
			headerLineCount++ // One line
		}
		headerLineCount++ // Blank line after name
	}

	// Description (if available)
	if tmpl.description != "" {
		if visualWidth(tmpl.description) > headerWrapWidth {
			wrapped := wrapLine(tmpl.description, headerWrapWidth)
			headerLineCount += len(wrapped)
		} else {
			headerLineCount++ // One line
		}
		headerLineCount++ // Blank line after description
	}

	// Source indicator - account for wrapping
	sourceIcon := ""
	sourceLabel := ""
	switch tmpl.source {
	case "global":
		sourceIcon = "🌐"
		sourceLabel = "Global Prompt (~/.prompts/)"
	case "command":
		sourceIcon = "⚙"
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
	sourceLine := sourceIcon + " " + sourceLabel
	if visualWidth(sourceLine) > headerWrapWidth {
		wrapped := wrapLine(sourceLine, headerWrapWidth)
		headerLineCount += len(wrapped)
	} else {
		headerLineCount++ // One line
	}

	// Variables line (if any) - account for wrapping
	if len(tmpl.variables) > 0 {
		// Build the plain variables line to calculate wrapping
		plainVarsLine := "Variables: " + strings.Join(tmpl.variables, ", ")
		if visualWidth(plainVarsLine) > headerWrapWidth {
			wrapped := wrapLine(plainVarsLine, headerWrapWidth)
			headerLineCount += len(wrapped)
		} else {
			headerLineCount++ // One line
		}
	}

	// Separator line - always one line since it's exactly headerWrapWidth characters
	// Each '─' has visual width 1, so total width equals headerWrapWidth
	headerLineCount++

	return headerLineCount
}
