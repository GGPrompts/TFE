package main

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

	// Calculate visible range based on scroll position
	totalLines := len(m.preview.content)
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

	// Calculate available width for content (pane width - line number width - scrollbar - border - padding)
	// Line number is 8 chars: "9999 │ " (5 for number, 1 for space, 1 for │, 1 for space)
	// Scrollbar takes 2 chars
	// Borders take up 2-4 additional characters depending on lipgloss rendering
	availableWidth := m.rightWidth - 17 // More conservative: line nums (8) + scrollbar (2) + borders (4) + padding (3)
	if m.viewMode == viewFullPreview {
		availableWidth = m.width - 17
	}
	if availableWidth < 20 {
		availableWidth = 20 // Minimum width
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

		// Content line - no need to pad since scrollbar is at fixed position
		line := m.preview.content[i]
		lineWidth := visualWidth(line)

		if lineWidth > availableWidth {
			// Truncate to fit and add "..."
			line = truncateToWidth(line, availableWidth-3) + "..."
		}

		s.WriteString(line)
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
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("39")).
		Width(m.width).
		Padding(0, 1)

	titleText := fmt.Sprintf("Preview: %s", m.preview.fileName)
	if m.preview.tooLarge || m.preview.isBinary {
		titleText += " [Cannot Preview]"
	}
	s.WriteString(previewTitleStyle.Render(titleText))
	s.WriteString("\n")

	// File info line
	infoText := fmt.Sprintf("Size: %s | Lines: %d | Scroll: %d-%d",
		formatFileSize(m.preview.fileSize),
		len(m.preview.content),
		m.preview.scrollPos+1,
		m.preview.scrollPos+m.height-4)
	s.WriteString(pathStyle.Render(infoText))
	s.WriteString("\n")

	// Content
	maxVisible := m.height - 4 // Reserve space for title, info, and help
	s.WriteString(m.renderPreview(maxVisible))

	// Help text
	s.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).PaddingLeft(2)
	s.WriteString(helpStyle.Render("↑/↓: scroll • PgUp/PgDown: page • E: edit • y/c: copy path • Esc: close • q: quit"))

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

	// Calculate max visible for both panes
	// title=1 + path+padding=2 + panes=maxVisible + status=1 + help=1 = m.height
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
			Foreground(lipgloss.Color("39")).
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
	leftBorderColor := "241"  // dim gray
	rightBorderColor := "241" // dim gray
	if m.focusedPane == leftPane {
		leftBorderColor = "39" // bright blue for focused pane
	} else {
		rightBorderColor = "39" // bright blue for focused pane
	}

	leftPaneStyle := lipgloss.NewStyle().
		Width(m.leftWidth).
		MaxWidth(m.leftWidth).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderRight(true).
		BorderForeground(lipgloss.Color(leftBorderColor))

	rightPaneStyle := lipgloss.NewStyle().
		Width(m.rightWidth).
		MaxWidth(m.rightWidth).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderLeft(true).
		BorderForeground(lipgloss.Color(rightBorderColor))

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
	if m.commandFocused {
		focusInfo = " • [COMMAND focused]"
	} else if m.focusedPane == leftPane {
		focusInfo = " • [LEFT focused]"
	} else {
		focusInfo = " • [RIGHT focused]"
	}
	statusText := fmt.Sprintf("%s%s • %s%s", itemsInfo, hiddenIndicator, m.displayMode.String(), focusInfo)
	s.WriteString(statusStyle.Render(statusText))

	// Command prompt (always visible at bottom)
	s.WriteString("\n")
	promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true).PaddingLeft(2)
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)

	s.WriteString(promptStyle.Render(m.currentPath + " $ "))
	s.WriteString(inputStyle.Render(m.commandInput))
	if m.commandFocused {
		s.WriteString(cursorStyle.Render("█"))
	}

	return s.String()
}
