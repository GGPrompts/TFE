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
		availableWidth = m.width - 10 // line nums (6) + scrollbar (2) + padding (2)
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
		renderer, err := glamour.NewTermRenderer(
			glamour.WithStandardStyle("auto"),
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
		availableWidth = m.width - 10 // line nums (6) + scrollbar (2) + padding (2)
	}
	if availableWidth < 20 {
		availableWidth = 20 // Minimum width
	}

	// If markdown, render with Glamour
	if m.preview.isMarkdown && m.preview.cachedRenderedContent != "" {
		// Use cached Glamour-rendered content (no line numbers)
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

		return s.String()
	}
	// If markdown flag is set but no rendered content, fall through to plain text rendering
	// This happens when Glamour fails or file is too large

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

	// Content
	maxVisible := m.height - 4 // Reserve space for title, info, and help
	s.WriteString(m.renderPreview(maxVisible))

	// Help text
	s.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).PaddingLeft(2)
	s.WriteString(helpStyle.Render("↑/↓: scroll • PgUp/PgDown: page • F4: edit • F5: copy path • Esc: close • F10: quit"))
	s.WriteString("\033[0m") // Reset ANSI codes

	return s.String()
}

// renderDualPane renders the split-pane layout using Lipgloss layout utilities
func (m model) renderDualPane() string {
	var s strings.Builder

	// Title with mode indicator
	titleText := "TFE - Terminal File Explorer [Dual-Pane]"
	if m.commandFocused {
		titleText += " [Command Mode]"
	}
	s.WriteString(titleStyle.Render(titleText))
	s.WriteString("\033[0m") // Reset ANSI codes
	s.WriteString("\n")

	// Toolbar buttons
	homeButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)

	// Home button
	s.WriteString(homeButtonStyle.Render("[🏠]"))
	s.WriteString(" ")

	// Favorites filter toggle button
	starIcon := "⭐"
	if m.showFavoritesOnly {
		starIcon = "✨" // Different icon when filter is active
	}
	s.WriteString(homeButtonStyle.Render("[" + starIcon + "]"))
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

	s.WriteString("\033[0m") // Reset ANSI codes
	s.WriteString("\n")

	// Command prompt with path (terminal-style)
	promptPrefix := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true).Render("$ ")
	pathPromptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	s.WriteString(promptPrefix)
	s.WriteString(pathPromptStyle.Render(getDisplayPath(m.currentPath)))
	s.WriteString(" ")
	s.WriteString(inputStyle.Render(m.commandInput))

	// Show cursor only when command mode is active
	if m.commandFocused {
		cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
		s.WriteString(cursorStyle.Render("█"))
	}
	// Explicitly reset styling after cursor to prevent ANSI code leakage
	s.WriteString("\033[0m")
	s.WriteString("\n")

	// Separator line between command prompt and panes
	s.WriteString("\n")

	// Calculate max visible for both panes
	// title=1 + path=1 + command=1 + separator=1 + panes=maxVisible + status=2 = m.height
	// Therefore: maxVisible = m.height - 6
	maxVisible := m.height - 6

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
		rightContent = previewTitle + "\033[0m\n" + separatorLine + "\033[0m\n"
		rightContent += m.renderPreview(maxVisible - 2)
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

	favoritesIndicator := ""
	if m.showFavoritesOnly {
		favoritesIndicator = " • ⭐ favorites only"
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
			selectedInfo = fmt.Sprintf("Selected: %s (folder)", currentFile.name)
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
	statusLine1 := fmt.Sprintf("%s%s%s • %s%s%s", itemsInfo, hiddenIndicator, favoritesIndicator, m.displayMode.String(), focusInfo, helpHint)
	s.WriteString(statusStyle.Render(statusLine1))
	s.WriteString("\033[0m") // Reset ANSI codes
	s.WriteString("\n")

	// Line 2: Selected file info
	statusLine2 := selectedInfo
	s.WriteString(statusStyle.Render(statusLine2))
	s.WriteString("\033[0m") // Reset ANSI codes

	return s.String()
}
