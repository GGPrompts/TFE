package main

// Module: render_preview.go
// Purpose: Core file preview rendering
// Responsibilities:
// - Rendering preview pane content with line numbers and scrollbar
// - Handling different file types (markdown, text, graphics protocol)
// - Managing scroll position and visible range

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

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
