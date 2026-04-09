package main

// Module: render_preview.go
// Purpose: Core file preview rendering
// Responsibilities:
// - Rendering preview pane content with line numbers and scrollbar
// - Handling different file types (markdown, text, graphics protocol)
// - Rendering git diff previews in changes mode
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

	// If in changes mode with diff preview enabled, render diff instead of file content
	if m.showChangesOnly && m.showDiffPreview {
		return m.renderDiffPreview(maxVisible)
	}

	// JSONL conversation file rendering (uses pre-cached rendered lines)
	if m.preview.isJSONL {
		return m.renderJSONLPreview(maxVisible)
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
					Foreground(uiSubtleText()).
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
		lineNumStyle := lipgloss.NewStyle().Foreground(uiSubtleText())
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
			Foreground(uiSubtleText()).
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

// renderDiffPreview renders a colorized git diff in the preview pane.
// Used when in changes mode with showDiffPreview enabled.
func (m model) renderDiffPreview(maxVisible int) string {
	var s strings.Builder

	// Calculate widths (same logic as normal preview)
	var boxContentWidth int
	if m.viewMode == viewFullPreview {
		boxContentWidth = m.width - 6
	} else {
		boxContentWidth = m.rightWidth - 2
	}

	// Diff lines: scrollbar (1) + space (1) = 2 chars overhead (no line numbers for diff)
	availableWidth := boxContentWidth - 2
	if availableWidth < 20 {
		availableWidth = 20
	}

	// Get the current file's git status code by matching preview path against changedFiles
	var gitStatusCode string
	for _, cf := range m.changedFiles {
		if cf.path == m.preview.filePath {
			gitStatusCode = extractGitStatusCode(cf.name)
			break
		}
	}

	// Get diff content
	diffOutput, err := m.getFileDiff(m.preview.filePath, gitStatusCode)
	if err != nil {
		// Show error message with fallback hint
		emptyStyle := lipgloss.NewStyle().
			Foreground(uiSubtleText()).
			Italic(true)
		s.WriteString(emptyStyle.Render(fmt.Sprintf("No diff available: %s", err.Error())))
		s.WriteString("\n")
		s.WriteString(emptyStyle.Render("Press 'd' to switch to file view"))

		// Pad remaining lines
		linesWritten := 2
		for linesWritten < maxVisible {
			s.WriteString("\n\033[0m")
			linesWritten++
		}
		return s.String()
	}

	// Split diff into lines and wrap
	rawLines := strings.Split(strings.TrimRight(diffOutput, "\n"), "\n")
	var wrappedLines []string
	// Track which style applies to each wrapped line (inherit from source line)
	var lineStyles []int // 0=normal, 1=added, 2=removed, 3=hunk, 4=meta
	for _, line := range rawLines {
		style := classifyDiffLine(line)
		wrapped := wrapLine(line, availableWidth)
		for range wrapped {
			wrappedLines = append(wrappedLines, "")
			lineStyles = append(lineStyles, style)
		}
		// Replace empty placeholders with actual wrapped content
		startIdx := len(wrappedLines) - len(wrapped)
		for j, w := range wrapped {
			wrappedLines[startIdx+j] = w
		}
	}

	// Calculate visible range based on scroll position
	totalLines := len(wrappedLines)
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

	// Render visible diff lines with scrollbar (no line numbers for diff)
	linesRendered := 0
	writeLine := func(line string) {
		if linesRendered > 0 {
			s.WriteString("\n")
		}
		s.WriteString(line)
		linesRendered++
	}

	for i := start; i < end; i++ {
		// Scrollbar
		scrollbar := m.renderScrollbar(i-start, maxVisible, totalLines)

		// Space after scrollbar
		renderedLine := scrollbar + " "

		// Colorize the diff line based on its type
		contentLine := wrappedLines[i]
		if visualWidth(contentLine) > availableWidth {
			contentLine = truncateToWidth(contentLine, availableWidth)
		}

		switch lineStyles[i] {
		case 1: // Added
			contentLine = diffAddedStyle.Render(contentLine)
		case 2: // Removed
			contentLine = diffRemovedStyle.Render(contentLine)
		case 3: // Hunk header
			contentLine = diffHunkHeaderStyle.Render(contentLine)
		case 4: // Meta/header
			contentLine = diffMetaStyle.Render(contentLine)
		}

		renderedLine += contentLine
		renderedLine += "\033[0m"
		writeLine(renderedLine)
	}

	// Add scroll indicator in dual-pane mode
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

		lastVisibleLine := end
		scrollIndicator := fmt.Sprintf(" %d/%d (%d%%) [diff]", lastVisibleLine, totalLines, scrollPercent)
		scrollStyle := lipgloss.NewStyle().
			Foreground(uiSubtleText()).
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

// classifyDiffLine returns the style category for a diff line:
// 0=normal, 1=added, 2=removed, 3=hunk header, 4=meta/header
func classifyDiffLine(line string) int {
	if strings.HasPrefix(line, "@@") {
		return 3 // Hunk header
	}
	if strings.HasPrefix(line, "diff ") ||
		strings.HasPrefix(line, "index ") ||
		strings.HasPrefix(line, "--- ") ||
		strings.HasPrefix(line, "+++ ") ||
		strings.HasPrefix(line, "new file") ||
		strings.HasPrefix(line, "deleted file") ||
		strings.HasPrefix(line, "old mode") ||
		strings.HasPrefix(line, "new mode") ||
		strings.HasPrefix(line, "similarity index") ||
		strings.HasPrefix(line, "rename from") ||
		strings.HasPrefix(line, "rename to") {
		return 4 // Meta/header
	}
	if strings.HasPrefix(line, "+") {
		return 1 // Added
	}
	if strings.HasPrefix(line, "-") {
		return 2 // Removed
	}
	return 0 // Context/normal
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

	scrollbarStyle := lipgloss.NewStyle().Foreground(uiMutedText())
	scrollbarThumbStyle := lipgloss.NewStyle().Foreground(currentTheme.Title.adaptiveColor())

	// Determine what to render for this line
	if lineIndex >= thumbStart && lineIndex < thumbStart+thumbSize {
		// This line is part of the scrollbar thumb (bright blue)
		return scrollbarThumbStyle.Render("│")
	} else {
		// This line is part of the scrollbar track (dim gray)
		return scrollbarStyle.Render("│")
	}
}
