package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// getWrappedLineCount calculates the total number of wrapped lines for the current preview
func (m model) getWrappedLineCount() int {
	if !m.preview.loaded {
		return 0
	}

	// Calculate available width
	availableWidth := m.rightWidth - 17
	if m.viewMode == viewFullPreview {
		availableWidth = m.width - 17
	}
	if availableWidth < 20 {
		availableWidth = 20
	}

	// For markdown, we need to render it to count lines
	if m.preview.isMarkdown {
		markdownContent := strings.Join(m.preview.content, "\n")
		renderer, err := glamour.NewTermRenderer(
			glamour.WithStandardStyle("dark"),
			glamour.WithWordWrap(availableWidth),
		)
		if err == nil {
			rendered, err := renderer.Render(markdownContent)
			if err == nil {
				renderedLines := strings.Split(strings.TrimRight(rendered, "\n"), "\n")
				return len(renderedLines)
			}
		}
		// Fallback if glamour fails
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

	// Calculate available width for content
	availableWidth := m.rightWidth - 17 // line nums (8) + scrollbar (2) + borders (4) + padding (3)
	if m.viewMode == viewFullPreview {
		availableWidth = m.width - 17
	}
	if availableWidth < 20 {
		availableWidth = 20 // Minimum width
	}

	// If markdown, render with Glamour
	if m.preview.isMarkdown {
		// Join content back to original markdown
		markdownContent := strings.Join(m.preview.content, "\n")

		// Create Glamour renderer with appropriate width
		// Use dark theme with standard styling
		renderer, err := glamour.NewTermRenderer(
			glamour.WithStandardStyle("dark"),
			glamour.WithWordWrap(availableWidth),
		)

		if err == nil {
			rendered, err := renderer.Render(markdownContent)
			if err == nil {
				// Split rendered markdown into lines for scrolling
				renderedLines := strings.Split(strings.TrimRight(rendered, "\n"), "\n")

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
					s.WriteString("\n")
				}

				return s.String()
			}
		}
		// If Glamour rendering fails, fall through to regular rendering
	}

	// Wrap all lines first
	var wrappedLines []string
	for _, line := range m.preview.content {
		wrapped := wrapLine(line, availableWidth)
		wrappedLines = append(wrappedLines, wrapped...)
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
		s.WriteString("\n")
	}

	return s.String()
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
	if m.preview.isMarkdown {
		titleText += " [Markdown]"
	}
	s.WriteString(previewTitleStyle.Render(titleText))
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
	s.WriteString("\n")

	// Content
	maxVisible := m.height - 4 // Reserve space for title, info, and help
	s.WriteString(m.renderPreview(maxVisible))

	// Help text
	s.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).PaddingLeft(2)
	s.WriteString(helpStyle.Render("↑/↓: scroll • PgUp/PgDown: page • F4: edit • F5: copy path • Esc: close • F10: quit"))

	return s.String()
}

// renderDualPane renders the split-pane layout using Lipgloss layout utilities
func (m model) renderDualPane() string {
	var s strings.Builder

	// Title
	s.WriteString(titleStyle.Render("TFE - Terminal File Explorer [Dual-Pane]"))
	s.WriteString("\n")

	// Current path
	s.WriteString(pathStyle.Render(m.currentPath))
	s.WriteString("\n")

	// Command prompt (left-aligned on its own line)
	promptPrefix := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true).Render("$ ")
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)

	s.WriteString(promptPrefix)
	s.WriteString(inputStyle.Render(m.commandInput))
	// Always show cursor (MC-style: command prompt is always active)
	s.WriteString(cursorStyle.Render("█"))
	s.WriteString("\n")

	// Separator line between command prompt and panes
	s.WriteString("\n")

	// Calculate max visible for both panes
	// title=1 + path=1 + command=1 + separator=1 + panes=maxVisible + status=1 = m.height
	// Therefore: maxVisible = m.height - 5
	maxVisible := m.height - 5

	// Get left pane content
	var leftContent string
	switch m.displayMode {
	case modeList:
		leftContent = m.renderListView(maxVisible)
	case modeGrid:
		leftContent = m.renderGridView(maxVisible)
	case modeDetail:
		leftContent = m.renderDetailView(maxVisible)
	case modeTree:
		leftContent = m.renderTreeView(maxVisible)
	default:
		leftContent = m.renderListView(maxVisible)
	}

	// Get right pane content
	rightContent := ""
	if m.preview.loaded {
		previewTitleText := fmt.Sprintf("Preview: %s", m.preview.fileName)
		previewTitle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#0087d7", Dark: "#5fd7ff"}).
			Render(previewTitleText)
		separatorLine := strings.Repeat("─", len(previewTitleText))
		rightContent = previewTitle + "\n" + separatorLine + "\n"
		rightContent += m.renderPreview(maxVisible - 2)
	} else {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)
		rightContent = emptyStyle.Render("No preview available\n\nSelect a file to preview")
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

	leftPaneStyle := lipgloss.NewStyle().
		Width(m.leftWidth).
		MaxWidth(m.leftWidth).
		Height(maxVisible).
		MaxHeight(maxVisible).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderRight(true).
		BorderForeground(leftBorderColor)

	rightPaneStyle := lipgloss.NewStyle().
		Width(m.rightWidth).
		MaxWidth(m.rightWidth).
		Height(maxVisible).
		MaxHeight(maxVisible).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderLeft(true).
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
	// Show focused pane info in status bar
	focusInfo := ""
	if m.focusedPane == leftPane {
		focusInfo = " • [LEFT focused]"
	} else {
		focusInfo = " • [RIGHT focused]"
	}
	// Help hint
	helpHint := " • F1: help"
	statusText := fmt.Sprintf("%s%s • %s%s%s", itemsInfo, hiddenIndicator, m.displayMode.String(), focusInfo, helpHint)
	s.WriteString(statusStyle.Render(statusText))

	return s.String()
}
