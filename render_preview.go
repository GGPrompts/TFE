package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderPreview renders the preview pane content
func (m model) renderPreview(maxVisible int) string {
	var s strings.Builder

	if !m.preview.loaded {
		s.WriteString("No file loaded")
		return s.String()
	}

	// Calculate visible range based on scroll position
	start := m.preview.scrollPos
	end := start + maxVisible
	if end > len(m.preview.content) {
		end = len(m.preview.content)
	}
	if start > len(m.preview.content) {
		start = 0
		end = maxVisible
		if end > len(m.preview.content) {
			end = len(m.preview.content)
		}
	}

	// Calculate available width for content (pane width - line number width - border - padding)
	// Line number is 8 chars: "9999 │ " (5 for number, 1 for space, 1 for │, 1 for space)
	// Borders take up 2-4 additional characters depending on lipgloss rendering
	availableWidth := m.rightWidth - 15 // More conservative: line nums (8) + borders (4) + padding (3)
	if m.viewMode == viewFullPreview {
		availableWidth = m.width - 15
	}
	if availableWidth < 20 {
		availableWidth = 20 // Minimum width
	}

	// Render lines with line numbers
	for i := start; i < end; i++ {
		// Use consistent 5-character width for line numbers (up to 9999 lines)
		lineNum := fmt.Sprintf("%5d │ ", i+1)
		lineNumStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		s.WriteString(lineNumStyle.Render(lineNum))

		// Truncate line to prevent wrapping
		line := m.preview.content[i]
		if len(line) > availableWidth {
			line = line[:availableWidth-3] + "..."
		}
		s.WriteString(line)
		s.WriteString("\n")
	}

	return s.String()
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
	statusText := fmt.Sprintf("%s%s • %s", itemsInfo, hiddenIndicator, m.displayMode.String())
	s.WriteString(statusStyle.Render(statusText))

	// Help text
	s.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).PaddingLeft(2)
	focusInfo := ""
	if m.focusedPane == leftPane {
		focusInfo = "[LEFT focused]"
	} else {
		focusInfo = "[RIGHT focused]"
	}
	s.WriteString(helpStyle.Render(fmt.Sprintf("%s • ↑/↓: scroll %s • Tab: switch pane • Space: exit • E: edit • y/c: copy", focusInfo,
		map[bool]string{true: "list", false: "preview"}[m.focusedPane == leftPane])))

	return s.String()
}
